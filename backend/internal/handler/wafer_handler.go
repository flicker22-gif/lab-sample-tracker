package handler

import (
	"errors"
	"net/http"
	"strconv"

	"sample-tracker/internal/dto"
	"sample-tracker/internal/model"
	"sample-tracker/internal/service"

	"github.com/gin-gonic/gin"
)

type WaferHandler struct {
	svc *service.WaferService
}

func NewWaferHandler(svc *service.WaferService) *WaferHandler { return &WaferHandler{svc: svc} }

// ListByLot GET /api/v1/lots/:id/wafers
func (h *WaferHandler) ListByLot(c *gin.Context) {
	lotID, err := parseUintParam(c, "id")
	if err != nil {
		fail(c, http.StatusBadRequest, "无效的批次 ID")
		return
	}
	list, err := h.svc.ListByLot(lotID)
	if err != nil {
		fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, ok(list))
}

// UploadMap POST /api/v1/wafers/:id/maps  (multipart: file, operator_id, remark)
func (h *WaferHandler) UploadMap(c *gin.Context) {
	waferID, err := parseUintParam(c, "id")
	if err != nil {
		fail(c, http.StatusBadRequest, "无效的晶圆 ID")
		return
	}
	operatorID, err := strconv.ParseUint(c.PostForm("operator_id"), 10, 64)
	if err != nil || operatorID == 0 {
		fail(c, http.StatusBadRequest, "请选择上传操作员")
		return
	}
	file, err := c.FormFile("file")
	if err != nil {
		fail(c, http.StatusBadRequest, "请上传 bin map 文件（字段名 file）")
		return
	}
	remark := c.PostForm("remark")
	m, err := h.svc.UploadMap(waferID, file, uint(operatorID), remark)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrNotFound):
			fail(c, http.StatusBadRequest, "晶圆或操作员不存在")
		default:
			fail(c, http.StatusBadRequest, err.Error())
		}
		return
	}
	c.JSON(http.StatusOK, ok(gin.H{
		"id": m.ID, "wafer_id": m.WaferID, "version": m.Version,
		"total_dies": m.TotalDies, "pass_dies": m.PassDies, "yield": m.Yield,
	}))
}

// GetMap GET /api/v1/maps/:id
func (h *WaferHandler) GetMap(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		fail(c, http.StatusBadRequest, "无效的 map ID")
		return
	}
	m, err := h.svc.GetMap(id)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			fail(c, http.StatusNotFound, "map 不存在")
			return
		}
		fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, ok(toMapVO(m)))
}

// ListVersions GET /api/v1/wafers/:id/maps
func (h *WaferHandler) ListVersions(c *gin.Context) {
	waferID, err := parseUintParam(c, "id")
	if err != nil {
		fail(c, http.StatusBadRequest, "无效的晶圆 ID")
		return
	}
	list, err := h.svc.ListVersions(waferID)
	if err != nil {
		fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, ok(list))
}

// DownloadMap GET /api/v1/maps/:id/download
func (h *WaferHandler) DownloadMap(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		fail(c, http.StatusBadRequest, "无效的 map ID")
		return
	}
	path, name, err := h.svc.GetMapFile(id)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			fail(c, http.StatusNotFound, "文件不存在")
			return
		}
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	c.FileAttachment(path, name)
}

func toMapVO(m *model.WaferBinMap) dto.WaferMapVO {
	vo := dto.WaferMapVO{
		ID: m.ID, WaferID: m.WaferID, Version: m.Version,
		FileName: m.FileName, FileSize: m.FileSize,
		MapData: [][]int(m.MapData), Rows: m.Rows, Cols: m.Cols, Notch: m.Notch,
		BinDefs:    make([]dto.BinDefVO, 0, len(m.BinDefs)),
		BinSummary: make([]dto.BinSummaryVO, 0, len(m.BinSummary)),
		TotalDies:  m.TotalDies, PassDies: m.PassDies, Yield: m.Yield,
		OperatorID: m.OperatorID, Remark: m.Remark, CreatedAt: m.CreatedAt,
	}
	for _, d := range m.BinDefs {
		vo.BinDefs = append(vo.BinDefs, dto.BinDefVO{
			Number: d.Number, Name: d.Name, Color: d.Color, Pass: d.Pass,
		})
	}
	for _, s := range m.BinSummary {
		vo.BinSummary = append(vo.BinSummary, dto.BinSummaryVO{
			Bin: s.Bin, Name: s.Name, Count: s.Count, Color: s.Color, Pass: s.Pass,
		})
	}
	if m.Operator != nil {
		vo.OperatorName = m.Operator.FullName
	}
	return vo
}
