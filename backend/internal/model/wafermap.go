package model

import (
	"database/sql/driver"
	"time"

	"gorm.io/gorm"
)

// Wafer 晶圆（属于一个批次，槽位 1..N）
type Wafer struct {
	ID     uint     `gorm:"primaryKey" json:"id"`
	LotID  uint     `gorm:"not null;index" json:"lot_id"`
	Lot    *WaferLot `gorm:"foreignKey:LotID" json:"lot,omitempty"`
	SlotNo int      `gorm:"not null" json:"slot_no"` // 槽位号
	// Scribe 晶圆标识/刻号（数据库列名 wafer_id）。
	// 注意：Go 字段名不能叫 WaferID——GORM 解析 WaferBinMap.Wafer 关系时
	// 会按 foreignKey:WaferID 在 Wafer 上找到同名字段，把它误建成指向
	// wafer_bin_maps 的 bigint 外键列（历史迁移 0002 已修复该缺陷）。
	Scribe       string         `gorm:"column:wafer_id;size:64" json:"wafer_id"`
	Product      string         `gorm:"size:128" json:"product"`
	CurrentMapID *uint          `gorm:"index" json:"current_map_id"` // 当前生效的 map 版本
	CurrentMap   *WaferBinMap   `gorm:"foreignKey:CurrentMapID" json:"current_map,omitempty"`
	Remark       string         `json:"remark"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// BinDef bin 定义
type BinDef struct {
	Number int    `json:"number"`
	Name   string `json:"name"`
	Color  string `json:"color"` // #RRGGBB，空则前端自动配色
	Pass   bool   `json:"pass"`  // 是否合格 bin
}

// WaferBinMap 单片晶圆的一次 bin map 上传（保留历史版本）
type WaferBinMap struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	WaferID  uint   `gorm:"not null;index" json:"wafer_id"`
	Wafer    *Wafer `gorm:"foreignKey:WaferID;references:ID" json:"wafer,omitempty"`
	Version  int    `gorm:"not null;default:1" json:"version"`
	FileName string `gorm:"size:256" json:"file_name"`
	FilePath string `gorm:"size:512" json:"-"` // 原始文件落盘路径
	FileSize int64  `json:"file_size"`
	// MapData 网格数据：rows 行 × cols 列，-1 表示无晶粒，>=0 为 bin 号
	MapData Int2DArray `gorm:"type:jsonb;not null" json:"map_data"`
	Rows    int        `gorm:"not null" json:"rows"`
	Cols    int        `gorm:"not null" json:"cols"`
	Notch   string     `gorm:"size:8;not null;default:down" json:"notch"` // up/down/left/right
	BinDefs BinDefList `gorm:"type:jsonb;not null;default:'[]'" json:"bin_defs"`
	// 汇总
	TotalDies  int            `gorm:"not null;default:0" json:"total_dies"`
	PassDies   int            `gorm:"not null;default:0" json:"pass_dies"`
	BinSummary BinSummaryList `gorm:"type:jsonb;not null;default:'[]'" json:"bin_summary"`
	Yield      float64        `gorm:"not null;default:0" json:"yield"` // 良率 0~1
	OperatorID uint           `gorm:"index" json:"operator_id"`
	Operator   *User          `gorm:"foreignKey:OperatorID" json:"operator,omitempty"`
	Remark     string         `json:"remark"`
	CreatedAt  time.Time      `json:"created_at"`
}

// TableName 显式指定表名
func (WaferBinMap) TableName() string { return "wafer_bin_maps" }

// BinSummaryItem 单个 bin 的数量
type BinSummaryItem struct {
	Bin   int    `json:"bin"`
	Name  string `json:"name"`
	Count int    `json:"count"`
	Color string `json:"color"`
	Pass  bool   `json:"pass"`
}

// ---- JSONB 自定义类型 ----

// Int2DArray [][]int 的 JSONB 映射
type Int2DArray [][]int

func (a Int2DArray) Value() (driver.Value, error) {
	return jsonMarshal(a)
}
func (a *Int2DArray) Scan(src interface{}) error {
	return jsonScan(src, a)
}
func (Int2DArray) GormDataType() string { return "jsonb" }

// BinDefList []BinDef 的 JSONB 映射
type BinDefList []BinDef

func (a BinDefList) Value() (driver.Value, error) { return jsonMarshal(a) }
func (a *BinDefList) Scan(src interface{}) error  { return jsonScan(src, a) }
func (BinDefList) GormDataType() string           { return "jsonb" }

// BinSummaryList []BinSummaryItem 的 JSONB 映射
type BinSummaryList []BinSummaryItem

func (a BinSummaryList) Value() (driver.Value, error) { return jsonMarshal(a) }
func (a *BinSummaryList) Scan(src interface{}) error  { return jsonScan(src, a) }
func (BinSummaryList) GormDataType() string           { return "jsonb" }
