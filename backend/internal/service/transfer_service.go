package service

import (
	"context"
	"errors"
	"time"

	"sample-tracker/internal/apperr"
	"sample-tracker/internal/dto"
	"sample-tracker/internal/model"
	"sample-tracker/internal/repository"

	"gorm.io/gorm"
)

// TransferService 样品流转登记
type TransferService struct {
	repos *repository.Repositories
}

func NewTransferService(repos *repository.Repositories) *TransferService {
	return &TransferService{repos: repos}
}

// Create 登记一次流转：校验位置、记录轨迹、更新样品当前位置与状态（同一事务）
func (s *TransferService) Create(ctx context.Context, req dto.CreateTransferRequest) (*dto.TransferVO, error) {
	if req.FromLocID != nil && *req.FromLocID == req.ToLocID {
		return nil, apperr.BadRequest("目标位置与来源位置不能相同")
	}

	var vo dto.TransferVO
	err := s.repos.WithTx(ctx, func(tx *repository.Repositories) error {
		sample, err := tx.Samples.GetSampleByID(ctx, req.SampleID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return apperr.BadRequest("样品、位置或操作人不存在")
			}
			return apperr.Wrap(err, "查询样品")
		}
		if sample.Status == model.StatusDiscarded {
			return apperr.BadRequest("样品已销毁，不能继续流转")
		}

		toLoc, err := tx.Meta.GetLocation(ctx, req.ToLocID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return apperr.BadRequest("样品、位置或操作人不存在")
			}
			return apperr.Wrap(err, "查询目标位置")
		}
		operator, err := tx.Meta.GetUser(ctx, req.OperatorID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return apperr.BadRequest("样品、位置或操作人不存在")
			}
			return apperr.Wrap(err, "查询操作人")
		}

		fromID := req.FromLocID
		if fromID == nil {
			fromID = sample.CurrentLocID // 默认从当前位置移出
		}
		if fromID != nil && *fromID == req.ToLocID {
			return apperr.BadRequest("目标位置与来源位置不能相同")
		}

		occurredAt := time.Now()
		if req.OccurredAt != nil {
			occurredAt = *req.OccurredAt
		}

		transfer := model.Transfer{
			SampleID:   sample.ID,
			FromLocID:  fromID,
			ToLocID:    req.ToLocID,
			OperatorID: req.OperatorID,
			Action:     req.Action,
			OccurredAt: occurredAt,
			Note:       req.Note,
		}
		if err := tx.Samples.CreateTransfer(ctx, &transfer); err != nil {
			return apperr.Wrap(err, "写入流转记录")
		}

		// 更新样品当前位置与状态
		updates := map[string]any{"current_location_id": req.ToLocID}
		switch {
		case toLoc.Type == model.LocDiscard || req.Action == "discard":
			updates["status"] = model.StatusDiscarded
		case toLoc.Type == model.LocFridge || req.Action == "store":
			updates["status"] = model.StatusStored
		case toLoc.Type == model.LocDevice || req.Action == "load":
			updates["status"] = model.StatusTesting
		case req.Action == "unload":
			updates["status"] = model.StatusStored
		default:
			updates["status"] = model.StatusReceived
		}
		if err := tx.Samples.UpdateSampleFields(ctx, sample.ID, updates); err != nil {
			return apperr.Wrap(err, "更新样品位置与状态")
		}

		// 组装返回
		if fromID != nil {
			if fromLoc, err := tx.Meta.GetLocation(ctx, *fromID); err == nil {
				vo.FromLocID = &fromLoc.ID
				vo.FromLocName = fromLoc.Name
				vo.FromLocType = string(fromLoc.Type)
			}
		}
		vo.ID = transfer.ID
		vo.Action = transfer.Action
		vo.ToLocID = toLoc.ID
		vo.ToLocName = toLoc.Name
		vo.ToLocType = string(toLoc.Type)
		vo.OperatorID = operator.ID
		vo.OperatorName = operator.FullName
		vo.OccurredAt = occurredAt
		vo.Note = req.Note
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &vo, nil
}
