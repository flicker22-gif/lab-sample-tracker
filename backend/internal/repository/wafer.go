package repository

import (
	"context"

	"sample-tracker/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// WaferRepository 晶圆槽位与 bin map 版本
type WaferRepository struct{ db *gorm.DB }

// ListByLot 批次下的晶圆（含当前 map 版本）
func (r *WaferRepository) ListByLot(ctx context.Context, lotID uint) ([]model.Wafer, error) {
	var wafers []model.Wafer
	err := r.db.WithContext(ctx).
		Preload("CurrentMap").
		Where("lot_id = ?", lotID).
		Order("slot_no").
		Find(&wafers).Error
	return wafers, err
}

// GetWaferForUpdate 取晶圆并加行锁（SELECT ... FOR UPDATE）。
// 上传 map 计算下一版本号前必须持有该锁，避免并发上传产生相同版本。
func (r *WaferRepository) GetWaferForUpdate(ctx context.Context, id uint) (*model.Wafer, error) {
	var w model.Wafer
	err := r.db.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		First(&w, id).Error
	if err != nil {
		return nil, err
	}
	return &w, nil
}

// MaxMapVersion 晶圆当前最大 map 版本（无记录返回 0）
func (r *WaferRepository) MaxMapVersion(ctx context.Context, waferID uint) (int, error) {
	var v int
	err := r.db.WithContext(ctx).Model(&model.WaferBinMap{}).
		Where("wafer_id = ?", waferID).
		Select("COALESCE(MAX(version),0)").Scan(&v).Error
	return v, err
}

func (r *WaferRepository) CreateMap(ctx context.Context, m *model.WaferBinMap) error {
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *WaferRepository) UpdateMapFilePath(ctx context.Context, id uint, path string) error {
	return r.db.WithContext(ctx).Model(&model.WaferBinMap{}).Where("id = ?", id).
		Update("file_path", path).Error
}

func (r *WaferRepository) SetWaferCurrentMap(ctx context.Context, waferID uint, mapID uint) error {
	return r.db.WithContext(ctx).Model(&model.Wafer{}).Where("id = ?", waferID).
		Update("current_map_id", mapID).Error
}

// GetMapByID map 详情（含操作员、晶圆）
func (r *WaferRepository) GetMapByID(ctx context.Context, id uint) (*model.WaferBinMap, error) {
	var m model.WaferBinMap
	err := r.db.WithContext(ctx).
		Preload("Operator").Preload("Wafer").
		First(&m, id).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

// ListMapVersions 晶圆的 map 版本列表（不含网格大字段）
func (r *WaferRepository) ListMapVersions(ctx context.Context, waferID uint) ([]model.WaferBinMap, error) {
	var ms []model.WaferBinMap
	err := r.db.WithContext(ctx).
		Preload("Operator").
		Select("id", "wafer_id", "version", "file_name", "file_size", "rows", "cols",
			"total_dies", "pass_dies", "yield", "operator_id", "remark", "created_at").
		Where("wafer_id = ?", waferID).
		Order("version DESC").
		Find(&ms).Error
	return ms, err
}

// GetMapFileInfo 仅取原始文件路径与文件名（下载用）
func (r *WaferRepository) GetMapFileInfo(ctx context.Context, id uint) (*model.WaferBinMap, error) {
	var m model.WaferBinMap
	err := r.db.WithContext(ctx).
		Select("id", "file_path", "file_name").
		First(&m, id).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}
