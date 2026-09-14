package repository

import (
	"context"

	"sample-tracker/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// LotFilter 批次分页查询条件（Offset/Limit 已由 service 归一化）
type LotFilter struct {
	Keyword string
	Status  string
	Offset  int
	Limit   int
}

// ProductionRepository 工艺、机台、晶圆批次、加工记录与机台状态日志
type ProductionRepository struct{ db *gorm.DB }

// ---------- 工艺 / 机台 ----------

func (r *ProductionRepository) ListProcesses(ctx context.Context) ([]model.Process, error) {
	var ps []model.Process
	err := r.db.WithContext(ctx).Where("enabled = ?", true).Order("seq, id").Find(&ps).Error
	return ps, err
}

func (r *ProductionRepository) ListMachines(ctx context.Context, processID uint) ([]model.Machine, error) {
	q := r.db.WithContext(ctx).Model(&model.Machine{}).
		Preload("Process").Preload("CurrentLot")
	if processID > 0 {
		q = q.Where("process_id = ?", processID)
	}
	var ms []model.Machine
	err := q.Order("process_id, code").Find(&ms).Error
	return ms, err
}

// GetMachineForUpdate 取机台并加行锁（SELECT ... FOR UPDATE），含工艺
func (r *ProductionRepository) GetMachineForUpdate(ctx context.Context, id uint) (*model.Machine, error) {
	var m model.Machine
	err := r.db.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Preload("Process").
		First(&m, id).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *ProductionRepository) UpdateMachineFields(ctx context.Context, id uint, fields map[string]any) error {
	return r.db.WithContext(ctx).Model(&model.Machine{}).Where("id = ?", id).Updates(fields).Error
}

func (r *ProductionRepository) CreateMachineStatusLog(ctx context.Context, log *model.MachineStatusLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

// ListMachineStatusLogs 机台状态变更记录（limit 已由 service 归一化）
func (r *ProductionRepository) ListMachineStatusLogs(ctx context.Context, machineID uint, limit int) ([]model.MachineStatusLog, error) {
	var logs []model.MachineStatusLog
	err := r.db.WithContext(ctx).
		Where("machine_id = ?", machineID).
		Preload("Operator").
		Order("occurred_at DESC, id DESC").
		Limit(limit).
		Find(&logs).Error
	return logs, err
}

// ---------- 批次 ----------

// NextLotSeq 批次号按天序列原子递增并返回新值（语义同 NextSampleSeq）
func (r *ProductionRepository) NextLotSeq(ctx context.Context, day string) (int, error) {
	var v int
	err := r.db.WithContext(ctx).Raw(
		`INSERT INTO lot_daily_seqs (day, value, updated_at) VALUES (?, 1, now())
		 ON CONFLICT (day) DO UPDATE SET value = lot_daily_seqs.value + 1, updated_at = now()
		 RETURNING value`, day).Scan(&v).Error
	return v, err
}

func (r *ProductionRepository) LotNoExists(ctx context.Context, lotNo string) (bool, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&model.WaferLot{}).Where("lot_no = ?", lotNo).Count(&n).Error
	return n > 0, err
}

func (r *ProductionRepository) CreateLot(ctx context.Context, lot *model.WaferLot) error {
	return r.db.WithContext(ctx).Create(lot).Error
}

// CreateWafers 批量创建晶圆槽位
func (r *ProductionRepository) CreateWafers(ctx context.Context, wafers []model.Wafer) error {
	if len(wafers) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Create(&wafers).Error
}

// GetLotByID 批次（含当前工艺）
func (r *ProductionRepository) GetLotByID(ctx context.Context, id uint) (*model.WaferLot, error) {
	var lot model.WaferLot
	err := r.db.WithContext(ctx).Preload("CurrentProcess").First(&lot, id).Error
	if err != nil {
		return nil, err
	}
	return &lot, nil
}

// GetLotForUpdate 取批次并加行锁（SELECT ... FOR UPDATE）
func (r *ProductionRepository) GetLotForUpdate(ctx context.Context, id uint) (*model.WaferLot, error) {
	var lot model.WaferLot
	err := r.db.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		First(&lot, id).Error
	if err != nil {
		return nil, err
	}
	return &lot, nil
}

func (r *ProductionRepository) UpdateLotFields(ctx context.Context, id uint, fields map[string]any) error {
	return r.db.WithContext(ctx).Model(&model.WaferLot{}).Where("id = ?", id).Updates(fields).Error
}

// ListLots 分页查询批次
func (r *ProductionRepository) ListLots(ctx context.Context, f LotFilter) ([]model.WaferLot, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.WaferLot{})
	if f.Keyword != "" {
		kw := "%" + f.Keyword + "%"
		q = q.Where("lot_no ILIKE ? OR product ILIKE ?", kw, kw)
	}
	if f.Status != "" {
		q = q.Where("status = ?", f.Status)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var lots []model.WaferLot
	err := q.Preload("CurrentProcess").
		Order("received_at DESC, id DESC").
		Offset(f.Offset).Limit(f.Limit).
		Find(&lots).Error
	return lots, total, err
}

// ---------- 加工记录 ----------

// CountOpenRecords 批次进行中的加工记录数（end_time 为空）
func (r *ProductionRepository) CountOpenRecords(ctx context.Context, lotID uint) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&model.LotProcessRecord{}).
		Where("lot_id = ? AND end_time IS NULL", lotID).Count(&n).Error
	return n, err
}

func (r *ProductionRepository) CreateProcessRecord(ctx context.Context, rec *model.LotProcessRecord) error {
	return r.db.WithContext(ctx).Create(rec).Error
}

// GetOpenRecordForUpdate 取批次当前进行中的加工记录并加行锁，含机台
func (r *ProductionRepository) GetOpenRecordForUpdate(ctx context.Context, lotID uint) (*model.LotProcessRecord, error) {
	var rec model.LotProcessRecord
	err := r.db.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Preload("Machine").
		Where("lot_id = ? AND end_time IS NULL", lotID).
		Order("id DESC").
		First(&rec).Error
	if err != nil {
		return nil, err
	}
	return &rec, nil
}

func (r *ProductionRepository) UpdateProcessRecordFields(ctx context.Context, id uint, fields map[string]any) error {
	return r.db.WithContext(ctx).Model(&model.LotProcessRecord{}).Where("id = ?", id).Updates(fields).Error
}

// ListRecordsByLot 批次全部加工记录（含工艺、机台、操作员）
func (r *ProductionRepository) ListRecordsByLot(ctx context.Context, lotID uint) ([]model.LotProcessRecord, error) {
	var records []model.LotProcessRecord
	err := r.db.WithContext(ctx).
		Where("lot_id = ?", lotID).
		Preload("Process").Preload("Machine").Preload("Operator").
		Order("start_time DESC, id DESC").
		Find(&records).Error
	return records, err
}
