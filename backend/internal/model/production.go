package model

import (
	"time"

	"gorm.io/gorm"
)

// MachineStatus 机台状态
type MachineStatus string

const (
	MachineIdle        MachineStatus = "idle"        // 空闲
	MachineRunning     MachineStatus = "running"     // 加工中
	MachineMaintenance MachineStatus = "maintenance" // 维护
	MachineFault       MachineStatus = "fault"       // 故障
)

// Process 工艺（一道工艺下有多台机台）
type Process struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Code      string         `gorm:"size:32;uniqueIndex;not null" json:"code"` // 工艺代码，如 ETCH
	Name      string         `gorm:"size:64;not null" json:"name"`             // 工艺名称，如 刻蚀
	Seq       int            `gorm:"default:0" json:"seq"`                     // 工序顺序
	Enabled   bool           `gorm:"not null;default:true" json:"enabled"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// Machine 机台（归属于一道工艺）
type Machine struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	Code         string         `gorm:"size:32;uniqueIndex;not null" json:"code"` // 机台编号，如 ETCH-01
	Name         string         `gorm:"size:64;not null" json:"name"`
	ProcessID    uint           `gorm:"not null;index" json:"process_id"`
	Process      *Process       `gorm:"foreignKey:ProcessID" json:"process,omitempty"`
	Status       MachineStatus  `gorm:"size:16;not null;default:idle;index" json:"status"`
	CurrentLotID *uint          `gorm:"index" json:"current_lot_id"` // 正在加工的批次
	CurrentLot   *WaferLot      `gorm:"foreignKey:CurrentLotID" json:"current_lot,omitempty"`
	Remark       string         `json:"remark"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// LotStatus 晶圆批次状态
type LotStatus string

const (
	LotPending    LotStatus = "pending"     // 待加工
	LotInProgress LotStatus = "in_progress" // 加工中
	LotCompleted  LotStatus = "completed"   // 已完工
	LotScrapped   LotStatus = "scrapped"    // 已报废
)

// WaferLot 晶圆批次
type WaferLot struct {
	ID               uint           `gorm:"primaryKey" json:"id"`
	LotNo            string         `gorm:"size:32;uniqueIndex;not null" json:"lot_no"` // 批次号，如 LOT-20260911-001
	Product          string         `gorm:"size:128" json:"product"`                    // 产品/型号
	WaferCount       int            `gorm:"not null;default:0" json:"wafer_count"`      // 晶圆数量
	Status           LotStatus      `gorm:"size:16;not null;default:pending;index" json:"status"`
	CurrentProcessID *uint          `gorm:"index" json:"current_process_id"` // 当前所在工艺
	CurrentProcess   *Process       `gorm:"foreignKey:CurrentProcessID" json:"current_process,omitempty"`
	ReceivedAt       time.Time      `gorm:"not null" json:"received_at"`
	Remark           string         `json:"remark"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
}

// LotDailySeq 批次号按天自增序列（day=YYYYMMDD）
type LotDailySeq struct {
	Day       string    `gorm:"primaryKey;size:8" json:"day"`
	Value     int       `gorm:"not null;default:0" json:"value"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ProcessResult 加工结果
type ProcessResult string

const (
	ResultPending ProcessResult = "pending" // 加工中，未出结果
	ResultOK      ProcessResult = "ok"      // 合格
	ResultNG      ProcessResult = "ng"      // 不合格
	ResultRework  ProcessResult = "rework"  // 返工
)

// LotProcessRecord 批次加工记录：某批次在哪台机台、加工起止时间、操作员
type LotProcessRecord struct {
	ID          uint          `gorm:"primaryKey" json:"id"`
	LotID       uint          `gorm:"not null;index" json:"lot_id"`
	Lot         *WaferLot     `gorm:"foreignKey:LotID" json:"lot,omitempty"`
	ProcessID   uint          `gorm:"not null;index" json:"process_id"`
	Process     *Process      `gorm:"foreignKey:ProcessID" json:"process,omitempty"`
	MachineID   uint          `gorm:"not null;index" json:"machine_id"`
	Machine     *Machine      `gorm:"foreignKey:MachineID" json:"machine,omitempty"`
	OperatorID  uint          `gorm:"not null;index" json:"operator_id"`
	Operator    *User         `gorm:"foreignKey:OperatorID" json:"operator,omitempty"`
	StartTime   time.Time     `gorm:"not null;index" json:"start_time"`
	EndTime     *time.Time    `gorm:"index" json:"end_time"` // nil 表示加工中
	DurationSec int           `gorm:"not null;default:0" json:"duration_sec"`
	Result      ProcessResult `gorm:"size:16;not null;default:pending;index" json:"result"`
	WaferOut    int           `gorm:"not null;default:0" json:"wafer_out"` // 完工产出片数
	Remark      string        `json:"remark"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}

// MachineStatusLog 机台状态变更记录
type MachineStatusLog struct {
	ID         uint          `gorm:"primaryKey" json:"id"`
	MachineID  uint          `gorm:"not null;index" json:"machine_id"`
	Machine    *Machine      `gorm:"foreignKey:MachineID" json:"machine,omitempty"`
	FromStatus MachineStatus `gorm:"size:16" json:"from_status"`
	ToStatus   MachineStatus `gorm:"size:16;not null" json:"to_status"`
	OperatorID uint          `gorm:"index" json:"operator_id"`
	Operator   *User         `gorm:"foreignKey:OperatorID" json:"operator,omitempty"`
	Reason     string        `json:"reason"`
	LotID      *uint         `gorm:"index" json:"lot_id"`
	OccurredAt time.Time     `gorm:"not null;index" json:"occurred_at"`
	CreatedAt  time.Time     `json:"created_at"`
}
