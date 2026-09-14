package repository

import (
	"context"

	"sample-tracker/internal/model"

	"gorm.io/gorm"
)

// SampleFilter 样品分页查询条件（Offset/Limit 已由 service 归一化）
type SampleFilter struct {
	Keyword string
	Status  string
	Offset  int
	Limit   int
}

// SampleRepository 样品、流转记录、检测结果与样品编号序列
type SampleRepository struct{ db *gorm.DB }

// NextSampleSeq 样品编号按天序列原子递增并返回新值。
// INSERT ... ON CONFLICT DO UPDATE ... RETURNING 单语句完成，
// 并发收样不重号（冲突行锁串行化，且 RETURNING 返回本事务递增值）。
func (r *SampleRepository) NextSampleSeq(ctx context.Context, day string) (int, error) {
	var v int
	err := r.db.WithContext(ctx).Raw(
		`INSERT INTO daily_seqs (day, value, updated_at) VALUES (?, 1, now())
		 ON CONFLICT (day) DO UPDATE SET value = daily_seqs.value + 1, updated_at = now()
		 RETURNING value`, day).Scan(&v).Error
	return v, err
}

func (r *SampleRepository) CreateSample(ctx context.Context, sample *model.Sample) error {
	return r.db.WithContext(ctx).Create(sample).Error
}

func (r *SampleRepository) UpdateSampleFields(ctx context.Context, id uint, fields map[string]any) error {
	return r.db.WithContext(ctx).Model(&model.Sample{}).Where("id = ?", id).Updates(fields).Error
}

// GetSampleByID 样品（含收样人、当前位置）
func (r *SampleRepository) GetSampleByID(ctx context.Context, id uint) (*model.Sample, error) {
	var sample model.Sample
	err := r.db.WithContext(ctx).
		Preload("Receiver").Preload("CurrentLoc").
		First(&sample, id).Error
	if err != nil {
		return nil, err
	}
	return &sample, nil
}

// ListSamples 分页查询，返回当前页数据与总数
func (r *SampleRepository) ListSamples(ctx context.Context, f SampleFilter) ([]model.Sample, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.Sample{})
	if f.Keyword != "" {
		kw := "%" + f.Keyword + "%"
		q = q.Where("code ILIKE ? OR name ILIKE ?", kw, kw)
	}
	if f.Status != "" {
		q = q.Where("status = ?", f.Status)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var samples []model.Sample
	err := q.Preload("Receiver").Preload("CurrentLoc").
		Order("received_at DESC, id DESC").
		Offset(f.Offset).Limit(f.Limit).
		Find(&samples).Error
	return samples, total, err
}

func (r *SampleRepository) CreateTransfer(ctx context.Context, t *model.Transfer) error {
	return r.db.WithContext(ctx).Create(t).Error
}

// ListTransfersBySample 样品全部流转记录（含位置与操作人）
func (r *SampleRepository) ListTransfersBySample(ctx context.Context, sampleID uint) ([]model.Transfer, error) {
	var transfers []model.Transfer
	err := r.db.WithContext(ctx).
		Where("sample_id = ?", sampleID).
		Preload("FromLoc").Preload("ToLoc").Preload("Operator").
		Order("occurred_at ASC, id ASC").
		Find(&transfers).Error
	return transfers, err
}

func (r *SampleRepository) CreateTestResult(ctx context.Context, res *model.TestResult) error {
	return r.db.WithContext(ctx).Create(res).Error
}

// ListResultsBySample 样品全部检测结果（含检测人）
func (r *SampleRepository) ListResultsBySample(ctx context.Context, sampleID uint) ([]model.TestResult, error) {
	var results []model.TestResult
	err := r.db.WithContext(ctx).
		Where("sample_id = ?", sampleID).
		Preload("Analyst").
		Order("created_at DESC, id DESC").
		Find(&results).Error
	return results, err
}
