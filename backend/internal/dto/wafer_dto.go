package dto

import "time"

type WaferVO struct {
	ID           uint      `json:"id"`
	LotID        uint      `json:"lot_id"`
	SlotNo       int       `json:"slot_no"`
	WaferID      string    `json:"wafer_id"`
	Product      string    `json:"product"`
	CurrentMapID *uint     `json:"current_map_id"`
	HasMap       bool      `json:"has_map"`
	Version      int       `json:"version"`
	Rows         int       `json:"rows"`
	Cols         int       `json:"cols"`
	TotalDies    int       `json:"total_dies"`
	PassDies     int       `json:"pass_dies"`
	Yield        float64   `json:"yield"`
	Remark       string    `json:"remark"`
	UploadedAt   time.Time `json:"uploaded_at"`
}

type BinDefVO struct {
	Number int    `json:"number"`
	Name   string `json:"name"`
	Color  string `json:"color"`
	Pass   bool   `json:"pass"`
}

type BinSummaryVO struct {
	Bin   int    `json:"bin"`
	Name  string `json:"name"`
	Count int    `json:"count"`
	Color string `json:"color"`
	Pass  bool   `json:"pass"`
}

type WaferMapVO struct {
	ID           uint           `json:"id"`
	WaferID      uint           `json:"wafer_id"`
	Version      int            `json:"version"`
	FileName     string         `json:"file_name"`
	FileSize     int64          `json:"file_size"`
	MapData      [][]int        `json:"map_data"`
	Rows         int            `json:"rows"`
	Cols         int            `json:"cols"`
	Notch        string         `json:"notch"`
	BinDefs      []BinDefVO     `json:"bin_defs"`
	TotalDies    int            `json:"total_dies"`
	PassDies     int            `json:"pass_dies"`
	BinSummary   []BinSummaryVO `json:"bin_summary"`
	Yield        float64        `json:"yield"`
	OperatorID   uint           `json:"operator_id"`
	OperatorName string         `json:"operator_name"`
	Remark       string         `json:"remark"`
	CreatedAt    time.Time      `json:"created_at"`
}

type MapVersionVO struct {
	ID           uint      `json:"id"`
	Version      int       `json:"version"`
	FileName     string    `json:"file_name"`
	FileSize     int64     `json:"file_size"`
	Rows         int       `json:"rows"`
	Cols         int       `json:"cols"`
	TotalDies    int       `json:"total_dies"`
	PassDies     int       `json:"pass_dies"`
	Yield        float64   `json:"yield"`
	OperatorName string    `json:"operator_name"`
	Remark       string    `json:"remark"`
	CreatedAt    time.Time `json:"created_at"`
}
