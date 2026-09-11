package model

import (
	"time"

	"gorm.io/gorm"
)

// LocationType 位置类型
type LocationType string

const (
	LocDevice  LocationType = "device"  // 检测/处理设备
	LocFridge  LocationType = "fridge"  // 冰箱/冷库
	LocBench   LocationType = "bench"   // 实验台
	LocDiscard LocationType = "discard" // 销毁点
)

// SampleStatus 样品状态
type SampleStatus string

const (
	StatusReceived  SampleStatus = "received"   // 已收样
	StatusInTransit SampleStatus = "in_transit" // 流转中（兼容扩展）
	StatusTesting   SampleStatus = "testing"    // 检测中
	StatusStored    SampleStatus = "stored"     // 存储中
	StatusReported  SampleStatus = "reported"   // 已出报告
	StatusDiscarded SampleStatus = "discarded"  // 已销毁
)

// Sample 样品
type Sample struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	Code         string         `gorm:"size:32;uniqueIndex;not null" json:"code"` // 唯一编号，如 SP-20260910-0001
	Name         string         `gorm:"size:128;not null" json:"name"`            // 样品名称
	Category     string         `gorm:"size:64" json:"category"`                  // 样品类型/基质
	Source       string         `gorm:"size:128" json:"source"`                   // 送检单位/来源
	ReceiverID   uint           `gorm:"index" json:"receiver_id"`
	Receiver     *User          `gorm:"foreignKey:ReceiverID" json:"receiver,omitempty"`
	ReceivedAt   time.Time      `gorm:"not null" json:"received_at"`
	Status       SampleStatus   `gorm:"size:16;not null;default:received;index" json:"status"`
	CurrentLocID *uint          `gorm:"index" json:"current_location_id"`
	CurrentLoc   *Location      `gorm:"foreignKey:CurrentLocID" json:"current_location,omitempty"`
	Remark       string         `json:"remark"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// User 操作人员
type User struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Username  string         `gorm:"size:64;uniqueIndex;not null" json:"username"`
	FullName  string         `gorm:"size:64;not null" json:"full_name"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// Location 存放位置（设备 / 冰箱）
type Location struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Name        string         `gorm:"size:128;not null;uniqueIndex" json:"name"` // 如 HPLC-01、冰箱A2-3层
	Type        LocationType   `gorm:"size:16;not null;index" json:"type"`
	Building    string         `gorm:"size:64" json:"building"`
	Room        string         `gorm:"size:64" json:"room"`
	Temperature float64        `json:"temperature"` // 冰箱设定温度，设备可空
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// Transfer 流转记录：谁在什么时候把样品从哪移到哪
type Transfer struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	SampleID   uint      `gorm:"not null;index" json:"sample_id"`
	Sample     *Sample   `gorm:"foreignKey:SampleID" json:"sample,omitempty"`
	FromLocID  *uint     `gorm:"index" json:"from_location_id"` // 收样首条记录可为空
	FromLoc    *Location `gorm:"foreignKey:FromLocID" json:"from_location,omitempty"`
	ToLocID    uint      `gorm:"not null;index" json:"to_location_id"`
	ToLoc      *Location `gorm:"foreignKey:ToLocID" json:"to_location,omitempty"`
	OperatorID uint      `gorm:"not null;index" json:"operator_id"`
	Operator   *User     `gorm:"foreignKey:OperatorID" json:"operator,omitempty"`
	Action     string    `gorm:"size:32;not null" json:"action"` // receive收样/transfer移交/load上机/unload下机/store入库/discard销毁
	OccurredAt time.Time `gorm:"not null;index" json:"occurred_at"`
	Note       string    `json:"note"`
	CreatedAt  time.Time `json:"created_at"`
}

// DailySeq 样品编号按天自增序列（day=YYYYMMDD）
type DailySeq struct {
	Day       string    `gorm:"primaryKey;size:8" json:"day"`
	Value     int       `gorm:"not null;default:0" json:"value"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TestResult 检测项目与结果
type TestResult struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	SampleID    uint           `gorm:"not null;index" json:"sample_id"`
	Sample      *Sample        `gorm:"foreignKey:SampleID" json:"sample,omitempty"`
	ItemName    string         `gorm:"size:128;not null" json:"item_name"` // 检测项目
	Method      string         `gorm:"size:128" json:"method"`             // 检测方法/标准
	Instrument  string         `gorm:"size:128" json:"instrument"`         // 使用仪器
	ResultValue string         `gorm:"size:256" json:"result_value"`       // 结果值
	Unit        string         `gorm:"size:32" json:"unit"`                // 单位
	Conclusion  string         `gorm:"size:32" json:"conclusion"`          // qualified合格/unqualified不合格/na待定
	ReportNo    string         `gorm:"size:64;index" json:"report_no"`     // 报告编号
	AnalystID   uint           `gorm:"index" json:"analyst_id"`
	Analyst     *User          `gorm:"foreignKey:AnalystID" json:"analyst,omitempty"`
	TestedAt    *time.Time     `json:"tested_at"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}
