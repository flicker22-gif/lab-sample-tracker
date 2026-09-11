package dto

import "time"

// ---- 工艺 ----

type ProcessVO struct {
	ID      uint   `json:"id"`
	Code    string `json:"code"`
	Name    string `json:"name"`
	Seq     int    `json:"seq"`
	Enabled bool   `json:"enabled"`
}

// ---- 机台 ----

type MachineVO struct {
	ID           uint   `json:"id"`
	Code         string `json:"code"`
	Name         string `json:"name"`
	ProcessID    uint   `json:"process_id"`
	ProcessCode  string `json:"process_code"`
	ProcessName  string `json:"process_name"`
	Status       string `json:"status"`
	CurrentLotID *uint  `json:"current_lot_id"`
	CurrentLotNo string `json:"current_lot_no"`
	Remark       string `json:"remark"`
}

type MachineStatusRequest struct {
	Status     string `json:"status" binding:"required,oneof=idle running maintenance fault"`
	OperatorID uint   `json:"operator_id" binding:"required"`
	Reason     string `json:"reason"`
}

// ---- 晶圆批次 ----

type CreateLotRequest struct {
	LotNo      string `json:"lot_no" binding:"max=32"` // 不传自动生成 LOT-YYYYMMDD-NNN
	Product    string `json:"product" binding:"max=128"`
	WaferCount int    `json:"wafer_count"`
	Remark     string `json:"remark"`
}

type LotVO struct {
	ID                 uint      `json:"id"`
	LotNo              string    `json:"lot_no"`
	Product            string    `json:"product"`
	WaferCount         int       `json:"wafer_count"`
	Status             string    `json:"status"`
	CurrentProcessID   *uint     `json:"current_process_id"`
	CurrentProcessName string    `json:"current_process_name"`
	ReceivedAt         time.Time `json:"received_at"`
	Remark             string    `json:"remark"`
	CreatedAt          time.Time `json:"created_at"`
}

type LotQuery struct {
	Keyword  string `form:"keyword"`
	Status   string `form:"status"`
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
}

// StartProcessRequest 开始加工：选择机台（决定工艺）、操作员
type StartProcessRequest struct {
	MachineID  uint       `json:"machine_id" binding:"required"`
	OperatorID uint       `json:"operator_id" binding:"required"`
	StartTime  *time.Time `json:"start_time"` // 不传取当前时间
	Remark     string     `json:"remark"`
}

// EndProcessRequest 结束加工：填结果与产出
type EndProcessRequest struct {
	Result     string     `json:"result" binding:"required,oneof=ok ng rework"`
	WaferOut   int        `json:"wafer_out"`
	EndTime    *time.Time `json:"end_time"` // 不传取当前时间
	Remark     string     `json:"remark"`
	OperatorID uint       `json:"operator_id"` // 可选，记录结束操作人
}

type RecordVO struct {
	ID           uint       `json:"id"`
	LotID        uint       `json:"lot_id"`
	ProcessID    uint       `json:"process_id"`
	ProcessCode  string     `json:"process_code"`
	ProcessName  string     `json:"process_name"`
	MachineID    uint       `json:"machine_id"`
	MachineCode  string     `json:"machine_code"`
	MachineName  string     `json:"machine_name"`
	OperatorID   uint       `json:"operator_id"`
	OperatorName string     `json:"operator_name"`
	StartTime    time.Time  `json:"start_time"`
	EndTime      *time.Time `json:"end_time"`
	DurationSec  int        `json:"duration_sec"`
	Result       string     `json:"result"`
	WaferOut     int        `json:"wafer_out"`
	Remark       string     `json:"remark"`
}

type StatusLogVO struct {
	ID           uint      `json:"id"`
	MachineID    uint      `json:"machine_id"`
	FromStatus   string    `json:"from_status"`
	ToStatus     string    `json:"to_status"`
	OperatorName string    `json:"operator_name"`
	Reason       string    `json:"reason"`
	LotID        *uint     `json:"lot_id"`
	OccurredAt   time.Time `json:"occurred_at"`
}

type LotDetailVO struct {
	LotVO
	Records []RecordVO `json:"records"`
}
