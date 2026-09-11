package service

import (
	"errors"
	"time"

	"sample-tracker/internal/dto"
	"sample-tracker/internal/model"

	"gorm.io/gorm"
)

type ResultService struct{ db *gorm.DB }

func NewResultService(db *gorm.DB) *ResultService { return &ResultService{db: db} }

// Create 登记检测结果；有报告编号时样品状态置为"已出报告"
func (s *ResultService) Create(req dto.CreateResultRequest) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var sample model.Sample
		if err := tx.First(&sample, req.SampleID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrNotFound
			}
			return err
		}
		var analyst model.User
		if err := tx.First(&analyst, req.AnalystID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrNotFound
			}
			return err
		}

		testedAt := req.TestedAt
		if testedAt == nil {
			now := time.Now()
			testedAt = &now
		}
		result := model.TestResult{
			SampleID:    req.SampleID,
			ItemName:    req.ItemName,
			Method:      req.Method,
			Instrument:  req.Instrument,
			ResultValue: req.ResultValue,
			Unit:        req.Unit,
			Conclusion:  req.Conclusion,
			ReportNo:    req.ReportNo,
			AnalystID:   req.AnalystID,
			TestedAt:    testedAt,
		}
		if err := tx.Create(&result).Error; err != nil {
			return err
		}
		if req.ReportNo != "" && sample.Status != model.StatusDiscarded {
			return tx.Model(&sample).Update("status", model.StatusReported).Error
		}
		return nil
	})
}
