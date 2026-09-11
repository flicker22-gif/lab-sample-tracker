package dto

import "time"

// CreateSampleRequest 新增样品
type CreateSampleRequest struct {
	Name       string `json:"name" binding:"required,max=128"`
	Category   string `json:"category" binding:"max=64"`
	Source     string `json:"source" binding:"max=128"`
	ReceiverID uint   `json:"receiver_id" binding:"required"`
	LocationID uint   `json:"location_id" binding:"required"` // 收样后首次存放位置
	Remark     string `json:"remark"`
}

// SampleQuery 样品列表查询
type SampleQuery struct {
	Keyword  string `form:"keyword"` // 编号/名称模糊
	Status   string `form:"status"`
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
}

type SampleVO struct {
	ID             uint      `json:"id"`
	Code           string    `json:"code"`
	Name           string    `json:"name"`
	Category       string    `json:"category"`
	Source         string    `json:"source"`
	Status         string    `json:"status"`
	Remark         string    `json:"remark"`
	ReceiverID     uint      `json:"receiver_id"`
	ReceiverName   string    `json:"receiver_name"`
	CurrentLocID   *uint     `json:"current_location_id"`
	CurrentLocName string    `json:"current_location_name"`
	CurrentLocType string    `json:"current_location_type"`
	ReceivedAt     time.Time `json:"received_at"`
	CreatedAt      time.Time `json:"created_at"`
}

type PageResult[T any] struct {
	List     []T   `json:"list"`
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
}

// CreateTransferRequest 登记一次流转（移交/上机/入库/销毁）
type CreateTransferRequest struct {
	SampleID   uint       `json:"sample_id" binding:"required"`
	FromLocID  *uint      `json:"from_location_id"` // 不传则取样品当前位置
	ToLocID    uint       `json:"to_location_id" binding:"required"`
	OperatorID uint       `json:"operator_id" binding:"required"`
	Action     string     `json:"action" binding:"required,oneof=transfer load unload store discard"`
	OccurredAt *time.Time `json:"occurred_at"` // 不传取当前时间
	Note       string     `json:"note"`
}

type TransferVO struct {
	ID           uint      `json:"id"`
	Action       string    `json:"action"`
	FromLocID    *uint     `json:"from_location_id"`
	FromLocName  string    `json:"from_location_name"`
	FromLocType  string    `json:"from_location_type"`
	ToLocID      uint      `json:"to_location_id"`
	ToLocName    string    `json:"to_location_name"`
	ToLocType    string    `json:"to_location_type"`
	OperatorID   uint      `json:"operator_id"`
	OperatorName string    `json:"operator_name"`
	OccurredAt   time.Time `json:"occurred_at"`
	Note         string    `json:"note"`
}

// CreateResultRequest 登记检测结果
type CreateResultRequest struct {
	SampleID    uint       `json:"sample_id" binding:"required"`
	ItemName    string     `json:"item_name" binding:"required,max=128"`
	Method      string     `json:"method"`
	Instrument  string     `json:"instrument"`
	ResultValue string     `json:"result_value"`
	Unit        string     `json:"unit"`
	Conclusion  string     `json:"conclusion" binding:"omitempty,oneof=qualified unqualified na"`
	ReportNo    string     `json:"report_no"`
	AnalystID   uint       `json:"analyst_id" binding:"required"`
	TestedAt    *time.Time `json:"tested_at"`
}

type ResultVO struct {
	ID          uint       `json:"id"`
	ItemName    string     `json:"item_name"`
	Method      string     `json:"method"`
	Instrument  string     `json:"instrument"`
	ResultValue string     `json:"result_value"`
	Unit        string     `json:"unit"`
	Conclusion  string     `json:"conclusion"`
	ReportNo    string     `json:"report_no"`
	AnalystID   uint       `json:"analyst_id"`
	AnalystName string     `json:"analyst_name"`
	TestedAt    *time.Time `json:"tested_at"`
	CreatedAt   time.Time  `json:"created_at"`
}

// SampleDetailVO 样品详情 + 完整流转轨迹 + 检测结果
type SampleDetailVO struct {
	SampleVO
	Transfers []TransferVO `json:"transfers"`
	Results   []ResultVO   `json:"results"`
}

type UserVO struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	FullName string `json:"full_name"`
}

type LocationVO struct {
	ID          uint    `json:"id"`
	Name        string  `json:"name"`
	Type        string  `json:"type"`
	Building    string  `json:"building"`
	Room        string  `json:"room"`
	Temperature float64 `json:"temperature"`
}
