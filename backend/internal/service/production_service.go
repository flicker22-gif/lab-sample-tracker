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
	ErrLotNotActive      = errors.New("批次已完工或报废，不能继续加工")
	ErrLotBusy           = errors.New("该批次已有进行中的加工，请先结束")
	ErrMachineNotIdle    = errors.New("机台当前不是空闲状态")
	ErrMachineRunning    = errors.New("机台加工中，请先在批次上结束当前加工")
	ErrNoOpenRecord      = errors.New("该批次没有进行中的加工记录")
	ErrInvalidTransition = errors.New("不允许手动切换为加工中，请通过开始加工操作")
)

type ProductionService struct{ db *gorm.DB }

func NewProductionService(db *gorm.DB) *ProductionService { return &ProductionService{db: db} }

// ---------- 工艺 / 机台 ----------

func (s *ProductionService) ListProcesses() ([]model.Process, error) {
	var ps []model.Process
	err := s.db.Where("enabled = ?", true).Order("seq, id").Find(&ps).Error
	return ps, err
}

func (s *ProductionService) ListMachines(processID uint) ([]dto.MachineVO, error) {
	tx := s.db.Model(&model.Machine{}).
		Preload("Process").
		Preload("CurrentLot")
	if processID > 0 {
		tx = tx.Where("process_id = ?", processID)
	}
	var ms []model.Machine
	if err := tx.Order("process_id, code").Find(&ms).Error; err != nil {
		return nil, err
	}
	out := make([]dto.MachineVO, 0, len(ms))
	for _, m := range ms {
		vo := dto.MachineVO{
			ID: m.ID, Code: m.Code, Name: m.Name, ProcessID: m.ProcessID,
			Status: string(m.Status), CurrentLotID: m.CurrentLotID, Remark: m.Remark,
		}
		if m.Process != nil {
			vo.ProcessCode = m.Process.Code
			vo.ProcessName = m.Process.Name
		}
		if m.CurrentLot != nil {
			vo.CurrentLotNo = m.CurrentLot.LotNo
		}
		out = append(out, vo)
	}
	return out, nil
}

// UpdateMachineStatus 手动切换机台状态（维护/故障/恢复空闲），并写状态变更记录
func (s *ProductionService) UpdateMachineStatus(machineID uint, req dto.MachineStatusRequest) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var m model.Machine
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&m, machineID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrNotFound
			}
			return err
		}
		to := model.MachineStatus(req.Status)
		if to == model.MachineRunning {
			return ErrInvalidTransition
		}
		if m.Status == model.MachineRunning {
			return ErrMachineRunning
		}
		if m.Status == to {
			return nil
		}
		var operator model.User
		if err := tx.First(&operator, req.OperatorID).Error; err != nil {
			return ErrNotFound
		}

		if err := tx.Model(&m).Update("status", to).Error; err != nil {
			return err
		}
		log := model.MachineStatusLog{
			MachineID:  m.ID,
			FromStatus: m.Status,
			ToStatus:   to,
			OperatorID: req.OperatorID,
			Reason:     req.Reason,
			OccurredAt: time.Now(),
		}
		return tx.Create(&log).Error
	})
}

func (s *ProductionService) ListStatusLogs(machineID uint, limit int) ([]model.MachineStatusLog, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	var logs []model.MachineStatusLog
	err := s.db.Where("machine_id = ?", machineID).
		Preload("Operator").
		Order("occurred_at DESC, id DESC").
		Limit(limit).
		Find(&logs).Error
	return logs, err
}

// ---------- 批次 ----------

// CreateLot 新建晶圆批次；批次号缺省时生成 LOT-YYYYMMDD-NNN（独立按天序列，并发安全）
func (s *ProductionService) CreateLot(req dto.CreateLotRequest) (*model.WaferLot, error) {
	now := time.Now()
	lot := &model.WaferLot{
		LotNo:      req.LotNo,
		Product:    req.Product,
		WaferCount: req.WaferCount,
		Status:     model.LotPending,
		ReceivedAt: now,
		Remark:     req.Remark,
	}
	err := s.db.Transaction(func(tx *gorm.DB) error {
		if lot.LotNo == "" {
			no, err := s.nextLotNo(tx, now)
			if err != nil {
				return err
			}
			lot.LotNo = no
		} else {
			var n int64
			if err := tx.Model(&model.WaferLot{}).Where("lot_no = ?", lot.LotNo).Count(&n).Error; err != nil {
				return err
			}
			if n > 0 {
				return errors.New("批次号已存在")
			}
		}
		return tx.Create(lot).Error
	})
	if err != nil {
		return nil, err
	}

	// 按晶圆数量创建槽位（失败不阻断批次创建，可后续补录）
	if lot.WaferCount > 0 {
		wafers := make([]model.Wafer, 0, lot.WaferCount)
		for slot := 1; slot <= lot.WaferCount; slot++ {
			wafers = append(wafers, model.Wafer{
				LotID: lot.ID, SlotNo: slot, Product: lot.Product,
			})
		}
		_ = s.db.Create(&wafers).Error
	}
	return lot, nil
}

func (s *ProductionService) nextLotNo(tx *gorm.DB, now time.Time) (string, error) {
	day := now.Format("20060102")
	seq := model.LotDailySeq{Day: day, Value: 1}
	if err := tx.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "day"}},
		DoUpdates: clause.Assignments(map[string]any{
			"value":      gorm.Expr("lot_daily_seqs.value + 1"),
			"updated_at": gorm.Expr("now()"),
		}),
	}).Create(&seq).Error; err != nil {
		return "", err
	}
	var v int
	if err := tx.Model(&model.LotDailySeq{}).Select("value").Where("day = ?", day).Scan(&v).Error; err != nil {
		return "", err
	}
	return fmt.Sprintf("LOT-%s-%03d", day, v), nil
}

func (s *ProductionService) ListLots(q dto.LotQuery) (*dto.PageResult[dto.LotVO], error) {
	page := q.Page
	if page < 1 {
		page = 1
	}
	size := q.PageSize
	if size < 1 || size > 200 {
		size = 20
	}
	tx := s.db.Model(&model.WaferLot{})
	if q.Keyword != "" {
		kw := "%" + q.Keyword + "%"
		tx = tx.Where("lot_no ILIKE ? OR product ILIKE ?", kw, kw)
	}
	if q.Status != "" {
		tx = tx.Where("status = ?", q.Status)
	}
	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, err
	}
	var lots []model.WaferLot
	if err := tx.Preload("CurrentProcess").
		Order("received_at DESC, id DESC").
		Offset((page - 1) * size).Limit(size).
		Find(&lots).Error; err != nil {
		return nil, err
	}
	list := make([]dto.LotVO, 0, len(lots))
	for i := range lots {
		list = append(list, toLotVO(&lots[i]))
	}
	return &dto.PageResult[dto.LotVO]{List: list, Total: total, Page: page, PageSize: size}, nil
}

func (s *ProductionService) GetLot(id uint) (*dto.LotDetailVO, error) {
	var lot model.WaferLot
	if err := s.db.Preload("CurrentProcess").First(&lot, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	var records []model.LotProcessRecord
	if err := s.db.Where("lot_id = ?", id).
		Preload("Process").Preload("Machine").Preload("Operator").
		Order("start_time DESC, id DESC").
		Find(&records).Error; err != nil {
		return nil, err
	}
	vo := &dto.LotDetailVO{LotVO: toLotVO(&lot)}
	vo.Records = make([]dto.RecordVO, 0, len(records))
	for i := range records {
		vo.Records = append(vo.Records, toRecordVO(&records[i]))
	}
	return vo, nil
}

// StartProcessing 批次在指定机台开始加工：写加工记录、机台置为加工中、批次进入该工艺
func (s *ProductionService) StartProcessing(lotID uint, req dto.StartProcessRequest) (*dto.RecordVO, error) {
	start := time.Now()
	if req.StartTime != nil {
		start = *req.StartTime
	}
	var result dto.RecordVO
	err := s.db.Transaction(func(tx *gorm.DB) error {
		var lot model.WaferLot
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&lot, lotID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrNotFound
			}
			return err
		}
		if lot.Status == model.LotCompleted || lot.Status == model.LotScrapped {
			return ErrLotNotActive
		}
		var openCount int64
		if err := tx.Model(&model.LotProcessRecord{}).
			Where("lot_id = ? AND end_time IS NULL", lotID).Count(&openCount).Error; err != nil {
			return err
		}
		if openCount > 0 {
			return ErrLotBusy
		}

		var machine model.Machine
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Preload("Process").First(&machine, req.MachineID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrNotFound
			}
			return err
		}
		if machine.Status != model.MachineIdle {
			return ErrMachineNotIdle
		}
		var operator model.User
		if err := tx.First(&operator, req.OperatorID).Error; err != nil {
			return ErrNotFound
		}

		record := model.LotProcessRecord{
			LotID:      lot.ID,
			ProcessID:  machine.ProcessID,
			MachineID:  machine.ID,
			OperatorID: req.OperatorID,
			StartTime:  start,
			Result:     model.ResultPending,
			Remark:     req.Remark,
		}
		if err := tx.Create(&record).Error; err != nil {
			return err
		}

		if err := tx.Model(&machine).Updates(map[string]any{
			"status":         model.MachineRunning,
			"current_lot_id": lot.ID,
		}).Error; err != nil {
			return err
		}
		if err := tx.Model(&lot).Updates(map[string]any{
			"status":             model.LotInProgress,
			"current_process_id": machine.ProcessID,
		}).Error; err != nil {
			return err
		}

		statusLog := model.MachineStatusLog{
			MachineID:  machine.ID,
			FromStatus: model.MachineIdle,
			ToStatus:   model.MachineRunning,
			OperatorID: req.OperatorID,
			LotID:      &lot.ID,
			Reason:     "开始加工",
			OccurredAt: start,
		}
		if err := tx.Create(&statusLog).Error; err != nil {
			return err
		}

		// 组装返回
		result = dto.RecordVO{
			ID: record.ID, LotID: lot.ID, ProcessID: machine.ProcessID,
			MachineID: machine.ID, OperatorID: req.OperatorID,
			StartTime: start, Result: string(model.ResultPending), Remark: req.Remark,
		}
		if machine.Process != nil {
			result.ProcessCode = machine.Process.Code
			result.ProcessName = machine.Process.Name
		}
		result.MachineCode = machine.Code
		result.MachineName = machine.Name
		result.OperatorName = operator.FullName
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// EndProcessing 结束批次当前加工：写结束时间/结果/产出，机台恢复空闲并记录状态
func (s *ProductionService) EndProcessing(lotID uint, req dto.EndProcessRequest) error {
	end := time.Now()
	if req.EndTime != nil {
		end = *req.EndTime
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		var record model.LotProcessRecord
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Preload("Machine").
			Where("lot_id = ? AND end_time IS NULL", lotID).
			Order("id DESC").First(&record).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrNoOpenRecord
			}
			return err
		}
		if end.Before(record.StartTime) {
			return errors.New("结束时间不能早于开始时间")
		}
		duration := int(end.Sub(record.StartTime).Seconds())

		updates := map[string]any{
			"end_time":     end,
			"duration_sec": duration,
			"result":       model.ProcessResult(req.Result),
			"wafer_out":    req.WaferOut,
		}
		if req.Remark != "" {
			updates["remark"] = req.Remark
		}
		if err := tx.Model(&record).Updates(updates).Error; err != nil {
			return err
		}

		// 机台恢复空闲
		if err := tx.Model(&model.Machine{}).Where("id = ?", record.MachineID).
			Updates(map[string]any{"status": model.MachineIdle, "current_lot_id": nil}).Error; err != nil {
			return err
		}
		operatorID := req.OperatorID
		if operatorID == 0 {
			operatorID = record.OperatorID
		}
		log := model.MachineStatusLog{
			MachineID:  record.MachineID,
			FromStatus: model.MachineRunning,
			ToStatus:   model.MachineIdle,
			OperatorID: operatorID,
			LotID:      &lotID,
			Reason:     "结束加工：" + resultText(model.ProcessResult(req.Result)),
			OccurredAt: end,
		}
		if err := tx.Create(&log).Error; err != nil {
			return err
		}
		return nil
	})
}

// CompleteLot 批次完工（要求无进行中的加工）
func (s *ProductionService) CompleteLot(lotID uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var lot model.WaferLot
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&lot, lotID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrNotFound
			}
			return err
		}
		if lot.Status == model.LotScrapped {
			return ErrLotNotActive
		}
		var openCount int64
		if err := tx.Model(&model.LotProcessRecord{}).
			Where("lot_id = ? AND end_time IS NULL", lotID).Count(&openCount).Error; err != nil {
			return err
		}
		if openCount > 0 {
			return ErrLotBusy
		}
		return tx.Model(&lot).Updates(map[string]any{
			"status":             model.LotCompleted,
			"current_process_id": nil,
		}).Error
	})
}

func resultText(r model.ProcessResult) string {
	switch r {
	case model.ResultOK:
		return "合格"
	case model.ResultNG:
		return "不合格"
	case model.ResultRework:
		return "返工"
	default:
		return string(r)
	}
}

func toLotVO(lot *model.WaferLot) dto.LotVO {
	vo := dto.LotVO{
		ID: lot.ID, LotNo: lot.LotNo, Product: lot.Product, WaferCount: lot.WaferCount,
		Status: string(lot.Status), CurrentProcessID: lot.CurrentProcessID,
		ReceivedAt: lot.ReceivedAt, Remark: lot.Remark, CreatedAt: lot.CreatedAt,
	}
	if lot.CurrentProcess != nil {
		vo.CurrentProcessName = lot.CurrentProcess.Name
	}
	return vo
}

func toRecordVO(r *model.LotProcessRecord) dto.RecordVO {
	vo := dto.RecordVO{
		ID: r.ID, LotID: r.LotID, ProcessID: r.ProcessID, MachineID: r.MachineID,
		OperatorID: r.OperatorID, StartTime: r.StartTime, EndTime: r.EndTime,
		DurationSec: r.DurationSec, Result: string(r.Result), WaferOut: r.WaferOut,
		Remark: r.Remark,
	}
	if r.Process != nil {
		vo.ProcessCode = r.Process.Code
		vo.ProcessName = r.Process.Name
	}
	if r.Machine != nil {
		vo.MachineCode = r.Machine.Code
		vo.MachineName = r.Machine.Name
	}
	if r.Operator != nil {
		vo.OperatorName = r.Operator.FullName
	}
	return vo
}
