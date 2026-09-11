package service

import (
	"errors"
	"fmt"
	"time"

	"sample-tracker/internal/dto"
	"sample-tracker/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrNotFound     = errors.New("记录不存在")
	ErrSameLocation = errors.New("目标位置与当前位置相同")
)

type SampleService struct{ db *gorm.DB }

func NewSampleService(db *gorm.DB) *SampleService { return &SampleService{db: db} }

// Create 新增样品 + 收样首条流转记录（同一事务）
func (s *SampleService) Create(req dto.CreateSampleRequest) (*model.Sample, error) {
	var loc model.Location
	if err := s.db.First(&loc, req.LocationID).Error; err != nil {
		return nil, ErrNotFound
	}
	var receiver model.User
	if err := s.db.First(&receiver, req.ReceiverID).Error; err != nil {
		return nil, ErrNotFound
	}

	now := time.Now()
	var sample *model.Sample

	err := s.db.Transaction(func(tx *gorm.DB) error {
		code, err := s.generateCode(tx, now)
		if err != nil {
			return err
		}
		sample = &model.Sample{
			Code:         code,
			Name:         req.Name,
			Category:     req.Category,
			Source:       req.Source,
			ReceiverID:   req.ReceiverID,
			ReceivedAt:   now,
			Status:       model.StatusReceived,
			CurrentLocID: &req.LocationID,
			Remark:       req.Remark,
		}
		if err := tx.Create(sample).Error; err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return fmt.Errorf("样品编号冲突，请重试")
			}
			return err
		}
		// 初始状态：存入冰箱→stored；上机→testing；其余保持 received
		status := model.StatusReceived
		switch loc.Type {
		case model.LocFridge:
			status = model.StatusStored
		case model.LocDevice:
			status = model.StatusTesting
		}
		if err := tx.Model(sample).Update("status", status).Error; err != nil {
			return err
		}
		sample.Status = status

		transfer := model.Transfer{
			SampleID:   sample.ID,
			FromLocID:  nil,
			ToLocID:    req.LocationID,
			OperatorID: req.ReceiverID,
			Action:     "receive",
			OccurredAt: now,
			Note:       "收样登记",
		}
		return tx.Create(&transfer).Error
	})
	if err != nil {
		return nil, err
	}
	return sample, nil
}

// generateCode 生成编号 SP-YYYYMMDD-NNNN。
// 通过 daily_seqs 表的 INSERT ... ON CONFLICT DO UPDATE 原子递增，并发收样不重号。
func (s *SampleService) generateCode(tx *gorm.DB, now time.Time) (string, error) {
	day := now.Format("20060102")
	seq := model.DailySeq{Day: day, Value: 1}
	if err := tx.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "day"}},
		DoUpdates: clause.Assignments(map[string]any{
			"value":      gorm.Expr("daily_seqs.value + 1"),
			"updated_at": gorm.Expr("now()"),
		}),
	}).Create(&seq).Error; err != nil {
		return "", err
	}
	var v int
	if err := tx.Model(&model.DailySeq{}).Select("value").
		Where("day = ?", day).Scan(&v).Error; err != nil {
		return "", err
	}
	return fmt.Sprintf("SP-%s-%04d", day, v), nil
}

// List 分页查询样品
func (s *SampleService) List(q dto.SampleQuery) (*dto.PageResult[dto.SampleVO], error) {
	page := q.Page
	if page < 1 {
		page = 1
	}
	size := q.PageSize
	if size < 1 || size > 200 {
		size = 20
	}

	tx := s.db.Model(&model.Sample{})
	if q.Keyword != "" {
		kw := "%" + q.Keyword + "%"
		tx = tx.Where("code ILIKE ? OR name ILIKE ?", kw, kw)
	}
	if q.Status != "" {
		tx = tx.Where("status = ?", q.Status)
	}

	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, err
	}

	var samples []model.Sample
	if err := tx.Preload("Receiver").Preload("CurrentLoc").
		Order("received_at DESC, id DESC").
		Offset((page - 1) * size).Limit(size).
		Find(&samples).Error; err != nil {
		return nil, err
	}

	list := make([]dto.SampleVO, 0, len(samples))
	for _, sp := range samples {
		list = append(list, toSampleVO(&sp))
	}
	return &dto.PageResult[dto.SampleVO]{List: list, Total: total, Page: page, PageSize: size}, nil
}

// Get 样品详情（含流转轨迹与检测结果）
func (s *SampleService) Get(id uint) (*dto.SampleDetailVO, error) {
	var sample model.Sample
	if err := s.db.Preload("Receiver").Preload("CurrentLoc").First(&sample, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	var transfers []model.Transfer
	if err := s.db.Where("sample_id = ?", id).
		Preload("FromLoc").Preload("ToLoc").Preload("Operator").
		Order("occurred_at ASC, id ASC").
		Find(&transfers).Error; err != nil {
		return nil, err
	}

	var results []model.TestResult
	if err := s.db.Where("sample_id = ?", id).
		Preload("Analyst").
		Order("created_at DESC, id DESC").
		Find(&results).Error; err != nil {
		return nil, err
	}

	vo := &dto.SampleDetailVO{SampleVO: toSampleVO(&sample)}
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
