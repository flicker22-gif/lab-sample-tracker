package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"sample-tracker/internal/apperr"
	"sample-tracker/internal/dto"
	"sample-tracker/internal/model"
	"sample-tracker/internal/repository"

	"gorm.io/gorm"
)

// SampleService 样品登记与查询
type SampleService struct {
	repos *repository.Repositories
}

func NewSampleService(repos *repository.Repositories) *SampleService {
	return &SampleService{repos: repos}
}

// Create 新增样品 + 收样首条流转记录（同一事务）
func (s *SampleService) Create(ctx context.Context, req dto.CreateSampleRequest) (*model.Sample, error) {
	loc, err := s.repos.Meta.GetLocation(ctx, req.LocationID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperr.BadRequest("接收人或存放位置不存在")
		}
		return nil, apperr.Wrap(err, "查询存放位置")
	}
	if _, err := s.repos.Meta.GetUser(ctx, req.ReceiverID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperr.BadRequest("接收人或存放位置不存在")
		}
		return nil, apperr.Wrap(err, "查询收样人")
	}

	now := time.Now()
	var sample *model.Sample

	err = s.repos.WithTx(ctx, func(tx *repository.Repositories) error {
		// 编号 SP-YYYYMMDD-NNNN：按天序列原子递增，并发收样不重号
		seq, err := tx.Samples.NextSampleSeq(ctx, now.Format("20060102"))
		if err != nil {
			return apperr.Wrap(err, "生成样品编号")
		}
		sample = &model.Sample{
			Code:         fmt.Sprintf("SP-%s-%04d", now.Format("20060102"), seq),
			Name:         req.Name,
			Category:     req.Category,
			Source:       req.Source,
			ReceiverID:   req.ReceiverID,
			ReceivedAt:   now,
			Status:       model.StatusReceived,
			CurrentLocID: &req.LocationID,
			Remark:       req.Remark,
		}
		if err := tx.Samples.CreateSample(ctx, sample); err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return apperr.Conflict("样品编号冲突，请重试")
			}
			return apperr.Wrap(err, "创建样品")
		}
		// 初始状态：存入冰箱→stored；上机→testing；其余保持 received
		status := model.StatusReceived
		switch loc.Type {
		case model.LocFridge:
			status = model.StatusStored
		case model.LocDevice:
			status = model.StatusTesting
		}
		if status != model.StatusReceived {
			if err := tx.Samples.UpdateSampleFields(ctx, sample.ID, map[string]any{"status": status}); err != nil {
				return apperr.Wrap(err, "更新样品状态")
			}
			sample.Status = status
		}

		transfer := model.Transfer{
			SampleID:   sample.ID,
			FromLocID:  nil,
			ToLocID:    req.LocationID,
			OperatorID: req.ReceiverID,
			Action:     "receive",
			OccurredAt: now,
			Note:       "收样登记",
		}
		if err := tx.Samples.CreateTransfer(ctx, &transfer); err != nil {
			return apperr.Wrap(err, "写入收样流转记录")
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return sample, nil
}

// List 分页查询样品
func (s *SampleService) List(ctx context.Context, q dto.SampleQuery) (*dto.PageResult[dto.SampleVO], error) {
	page := q.Page
	if page < 1 {
		page = 1
	}
	size := q.PageSize
	if size < 1 || size > 200 {
		size = 20
	}

	samples, total, err := s.repos.Samples.ListSamples(ctx, repository.SampleFilter{
		Keyword: q.Keyword,
		Status:  q.Status,
		Offset:  (page - 1) * size,
		Limit:   size,
	})
	if err != nil {
		return nil, apperr.Wrap(err, "查询样品列表")
	}

	list := make([]dto.SampleVO, 0, len(samples))
	for i := range samples {
		list = append(list, toSampleVO(&samples[i]))
	}
	return &dto.PageResult[dto.SampleVO]{List: list, Total: total, Page: page, PageSize: size}, nil
}

// Get 样品详情（含流转轨迹与检测结果）
func (s *SampleService) Get(ctx context.Context, id uint) (*dto.SampleDetailVO, error) {
	sample, err := s.repos.Samples.GetSampleByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperr.NotFound("样品不存在")
		}
		return nil, apperr.Wrap(err, "查询样品")
	}

	transfers, err := s.repos.Samples.ListTransfersBySample(ctx, id)
	if err != nil {
		return nil, apperr.Wrap(err, "查询流转记录")
	}
	results, err := s.repos.Samples.ListResultsBySample(ctx, id)
	if err != nil {
		return nil, apperr.Wrap(err, "查询检测结果")
	}

	vo := &dto.SampleDetailVO{SampleVO: toSampleVO(sample)}
	vo.Transfers = make([]dto.TransferVO, 0, len(transfers))
	for i := range transfers {
		vo.Transfers = append(vo.Transfers, toTransferVO(&transfers[i]))
	}
	vo.Results = make([]dto.ResultVO, 0, len(results))
	for i := range results {
		vo.Results = append(vo.Results, toResultVO(&results[i]))
	}
	return vo, nil
}

func toSampleVO(sp *model.Sample) dto.SampleVO {
	vo := dto.SampleVO{
		ID:         sp.ID,
		Code:       sp.Code,
		Name:       sp.Name,
		Category:   sp.Category,
		Source:     sp.Source,
		Status:     string(sp.Status),
		Remark:     sp.Remark,
		ReceiverID: sp.ReceiverID,
		ReceivedAt: sp.ReceivedAt,
		CreatedAt:  sp.CreatedAt,
	}
	if sp.Receiver != nil {
		vo.ReceiverName = sp.Receiver.FullName
	}
	if sp.CurrentLoc != nil {
		vo.CurrentLocID = &sp.CurrentLoc.ID
		vo.CurrentLocName = sp.CurrentLoc.Name
		vo.CurrentLocType = string(sp.CurrentLoc.Type)
	}
	return vo
}

func toTransferVO(t *model.Transfer) dto.TransferVO {
	vo := dto.TransferVO{
		ID:         t.ID,
		Action:     t.Action,
		FromLocID:  t.FromLocID,
		ToLocID:    t.ToLocID,
		OperatorID: t.OperatorID,
		OccurredAt: t.OccurredAt,
		Note:       t.Note,
	}
	if t.FromLoc != nil {
		vo.FromLocName = t.FromLoc.Name
		vo.FromLocType = string(t.FromLoc.Type)
	}
	if t.ToLoc != nil {
		vo.ToLocName = t.ToLoc.Name
		vo.ToLocType = string(t.ToLoc.Type)
	}
	if t.Operator != nil {
		vo.OperatorName = t.Operator.FullName
	}
	return vo
}

func toResultVO(r *model.TestResult) dto.ResultVO {
	vo := dto.ResultVO{
		ID:          r.ID,
		ItemName:    r.ItemName,
		Method:      r.Method,
		Instrument:  r.Instrument,
		ResultValue: r.ResultValue,
		Unit:        r.Unit,
		Conclusion:  r.Conclusion,
		ReportNo:    r.ReportNo,
		AnalystID:   r.AnalystID,
		TestedAt:    r.TestedAt,
		CreatedAt:   r.CreatedAt,
	}
	if r.Analyst != nil {
		vo.AnalystName = r.Analyst.FullName
	}
	return vo
}
