package service

import (
	"errors"
	"time"

	"sample-tracker/internal/dto"
	"sample-tracker/internal/model"

	"gorm.io/gorm"
)

type TransferService struct{ db *gorm.DB }

func NewTransferService(db *gorm.DB) *TransferService { return &TransferService{db: db} }

// Create 登记一次流转：校验位置、记录轨迹、更新样品当前位置与状态（同一事务）
func (s *TransferService) Create(req dto.CreateTransferRequest) (*dto.TransferVO, error) {
	if req.FromLocID != nil && *req.FromLocID == req.ToLocID {
		return nil, ErrSameLocation
	}

	var vo dto.TransferVO
	err := s.db.Transaction(func(tx *gorm.DB) error {
		var sample model.Sample
		if err := tx.First(&sample, req.SampleID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrNotFound
			}
			return err
		}
		if sample.Status == model.StatusDiscarded {
			return errors.New("样品已销毁，不能继续流转")
		}

		var toLoc model.Location
		if err := tx.First(&toLoc, req.ToLocID).Error; err != nil {
			return ErrNotFound
		}
		var operator model.User
		if err := tx.First(&operator, req.OperatorID).Error; err != nil {
			return ErrNotFound
		}

		fromID := req.FromLocID
		if fromID == nil {
			fromID = sample.CurrentLocID // 默认从当前位置移出
		}
		if fromID != nil && *fromID == req.ToLocID {
			return ErrSameLocation
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
		if err := tx.Create(&transfer).Error; err != nil {
			return err
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
		if err := tx.Model(&sample).Updates(updates).Error; err != nil {
			return err
		}

		// 组装返回
		if fromID != nil {
			var fromLoc model.Location
			if err := tx.First(&fromLoc, *fromID).Error; err == nil {
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
