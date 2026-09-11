package service

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"sort"

	"sample-tracker/internal/binmap"
	"sample-tracker/internal/dto"
	"sample-tracker/internal/model"

	"gorm.io/gorm"
)

type WaferService struct {
	db         *gorm.DB
	storageDir string
}

func NewWaferService(db *gorm.DB) *WaferService {
	dir := os.Getenv("MAP_STORAGE_DIR")
	if dir == "" {
		dir = "./data/maps"
	}
	_ = os.MkdirAll(dir, 0o755)
	return &WaferService{db: db, storageDir: dir}
}

const maxMapFileSize = 16 << 20 // 16MB

// ListByLot 批次下的晶圆（含当前 map 汇总）
func (s *WaferService) ListByLot(lotID uint) ([]dto.WaferVO, error) {
	var wafers []model.Wafer
	if err := s.db.Preload("CurrentMap").
		Where("lot_id = ?", lotID).
		Order("slot_no").Find(&wafers).Error; err != nil {
		return nil, err
	}
	out := make([]dto.WaferVO, 0, len(wafers))
	for _, w := range wafers {
		vo := dto.WaferVO{
			ID: w.ID, LotID: w.LotID, SlotNo: w.SlotNo, WaferID: w.WaferID,
			Product: w.Product, CurrentMapID: w.CurrentMapID, Remark: w.Remark,
		}
		if m := w.CurrentMap; m != nil {
			vo.HasMap = true
			vo.Version = m.Version
			vo.TotalDies = m.TotalDies
			vo.PassDies = m.PassDies
			vo.Yield = m.Yield
			vo.Rows = m.Rows
			vo.Cols = m.Cols
			vo.UploadedAt = m.CreatedAt
		}
		out = append(out, vo)
	}
	return out, nil
}

// UploadMap 上传单片晶圆的 bin map
func (s *WaferService) UploadMap(waferID uint, header *multipart.FileHeader, operatorID uint, remark string) (*model.WaferBinMap, error) {
	if header.Size > maxMapFileSize {
		return nil, errors.New("文件过大，最大支持 16MB")
	}
	f, err := header.Open()
	if err != nil {
		return nil, err
	}
	defer f.Close()

	buf := make([]byte, header.Size)
	if _, err := io.ReadFull(f, buf); err != nil && err != io.ErrUnexpectedEOF {
		return nil, err
	}

	parsed, err := binmap.Parse(string(buf))
	if err != nil {
		return nil, fmt.Errorf("bin map 解析失败: %w", err)
	}

	var operator model.User
	if err := s.db.First(&operator, operatorID).Error; err != nil {
		return nil, ErrNotFound
	}

	// 汇总
	counts := map[int]int{}
	total := 0
	for _, r := range parsed.Data {
		for _, b := range r {
			if b >= 0 {
				counts[b]++
				total++
			}
		}
	}
	passByBin := map[int]bool{}
	for _, d := range parsed.BinDefs {
		passByBin[d.Number] = d.Pass
	}
	passDies := 0
	summary := make([]model.BinSummaryItem, 0, len(parsed.BinDefs))
	bins := make([]int, 0, len(parsed.BinDefs))
	for _, d := range parsed.BinDefs {
		bins = append(bins, d.Number)
	}
	sort.Ints(bins)
	defMap := map[int]model.BinDef{}
	for _, d := range parsed.BinDefs {
		defMap[d.Number] = d
	}
	for _, b := range bins {
		c := counts[b]
		if passByBin[b] {
			passDies += c
		}
		d := defMap[b]
		summary = append(summary, model.BinSummaryItem{
			Bin: b, Name: d.Name, Count: c, Color: d.Color, Pass: d.Pass,
		})
	}
	yield := 0.0
	if total > 0 {
		yield = float64(passDies) / float64(total)
	}

	var saved *model.WaferBinMap
	err = s.db.Transaction(func(tx *gorm.DB) error {
		var wafer model.Wafer
		if err := tx.First(&wafer, waferID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrNotFound
			}
			return err
		}

		var lastVersion int
		if err := tx.Model(&model.WaferBinMap{}).
			Where("wafer_id = ?", waferID).
			Select("COALESCE(MAX(version),0)").Scan(&lastVersion).Error; err != nil {
			return err
		}

		m := &model.WaferBinMap{
			WaferID:    waferID,
			Version:    lastVersion + 1,
			FileName:   filepath.Base(header.Filename),
			FileSize:   header.Size,
			MapData:    model.Int2DArray(parsed.Data),
			Rows:       parsed.Rows,
			Cols:       parsed.Cols,
			Notch:      parsed.Notch,
			BinDefs:    model.BinDefList(parsed.BinDefs),
			TotalDies:  total,
			PassDies:   passDies,
			BinSummary: model.BinSummaryList(summary),
			Yield:      yield,
			OperatorID: operatorID,
			Remark:     remark,
		}
		if err := tx.Create(m).Error; err != nil {
			return err
		}

		// 原始文件落盘：wafer_<id>/v<版本>_<原名>
		dir := filepath.Join(s.storageDir, fmt.Sprintf("wafer_%d", waferID))
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
		name := fmt.Sprintf("v%d_%s", m.Version, safeName(m.FileName))
		full := filepath.Join(dir, name)
		if err := os.WriteFile(full, buf, 0o644); err != nil {
			return err
		}
		m.FilePath = full
		if err := tx.Model(m).Update("file_path", full).Error; err != nil {
			return err
		}

		if err := tx.Model(&wafer).Update("current_map_id", m.ID).Error; err != nil {
			return err
		}
		saved = m
		return nil
	})
	if err != nil {
		return nil, err
	}
	return saved, nil
}

// GetMap map 详情（含网格）
func (s *WaferService) GetMap(id uint) (*model.WaferBinMap, error) {
	var m model.WaferBinMap
	if err := s.db.Preload("Operator").Preload("Wafer").First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &m, nil
}

// ListVersions 晶圆的上传版本列表（不含网格大字段）
func (s *WaferService) ListVersions(waferID uint) ([]dto.MapVersionVO, error) {
	var ms []model.WaferBinMap
	if err := s.db.Preload("Operator").
		Select("id", "wafer_id", "version", "file_name", "file_size", "rows", "cols",
			"total_dies", "pass_dies", "yield", "operator_id", "remark", "created_at").
		Where("wafer_id = ?", waferID).
		Order("version DESC").Find(&ms).Error; err != nil {
		return nil, err
	}
	out := make([]dto.MapVersionVO, 0, len(ms))
	for _, m := range ms {
		name := ""
		if m.Operator != nil {
			name = m.Operator.FullName
		}
		out = append(out, dto.MapVersionVO{
			ID: m.ID, Version: m.Version, FileName: m.FileName, FileSize: m.FileSize,
			Rows: m.Rows, Cols: m.Cols, TotalDies: m.TotalDies, PassDies: m.PassDies,
			Yield: m.Yield, OperatorName: name, Remark: m.Remark, CreatedAt: m.CreatedAt,
		})
	}
	return out, nil
}

// GetMapFile 取原始文件路径用于下载
func (s *WaferService) GetMapFile(id uint) (path, name string, err error) {
	var m model.WaferBinMap
	if err := s.db.Select("file_path", "file_name").First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", "", ErrNotFound
		}
		return "", "", err
	}
	if m.FilePath == "" {
		return "", "", errors.New("原始文件不存在")
	}
	if _, err := os.Stat(m.FilePath); err != nil {
		return "", "", errors.New("原始文件已丢失")
	}
	return m.FilePath, m.FileName, nil
}

func safeName(n string) string {
	name := filepath.Base(n)
	if name == "." || name == "/" || name == "" {
		return "map.txt"
	}
	return name
}
