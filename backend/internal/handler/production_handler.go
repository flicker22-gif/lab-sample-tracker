package handler

import (
	"errors"
	"net/http"
	"strconv"

	"sample-tracker/internal/dto"
	"sample-tracker/internal/service"

	"github.com/gin-gonic/gin"
)

type ProductionHandler struct {
	svc *service.ProductionService
}

func NewProductionHandler(svc *service.ProductionService) *ProductionHandler {
	return &ProductionHandler{svc: svc}
}

// ---- 工艺 / 机台 ----

// ListProcesses GET /api/v1/processes
func (h *ProductionHandler) ListProcesses(c *gin.Context) {
	ps, err := h.svc.ListProcesses()
	if err != nil {
		fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, ok(ps))
}

// ListMachines GET /api/v1/machines?process_id=
func (h *ProductionHandler) ListMachines(c *gin.Context) {
	pid, _ := strconv.ParseUint(c.Query("process_id"), 10, 64)
	ms, err := h.svc.ListMachines(uint(pid))
	if err != nil {
		fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, ok(ms))
}

// UpdateMachineStatus PUT /api/v1/machines/:id/status
func (h *ProductionHandler) UpdateMachineStatus(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		fail(c, http.StatusBadRequest, "无效的机台 ID")
		return
	}
	var req dto.MachineStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}
	if err := h.svc.UpdateMachineStatus(id, req); err != nil {
		h.writeServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, ok(gin.H{"id": id}))
}

// ListMachineLogs GET /api/v1/machines/:id/status-logs
func (h *ProductionHandler) ListMachineLogs(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		fail(c, http.StatusBadRequest, "无效的机台 ID")
		return
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	logs, err := h.svc.ListStatusLogs(id, limit)
	if err != nil {
		fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, ok(logs))
}

// ---- 批次 ----

// CreateLot POST /api/v1/lots
func (h *ProductionHandler) CreateLot(c *gin.Context) {
	var req dto.CreateLotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}
	if req.WaferCount < 0 {
		fail(c, http.StatusBadRequest, "晶圆数量不能为负")
		return
	}
	lot, err := h.svc.CreateLot(req)
	if err != nil {
		h.writeServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, ok(gin.H{"id": lot.ID, "lot_no": lot.LotNo}))
}

// ListLots GET /api/v1/lots
func (h *ProductionHandler) ListLots(c *gin.Context) {
	var q dto.LotQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		fail(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}
	page, err := h.svc.ListLots(q)
	if err != nil {
		fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, ok(page))
}

// GetLot GET /api/v1/lots/:id
func (h *ProductionHandler) GetLot(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		fail(c, http.StatusBadRequest, "无效的批次 ID")
		return
	}
	detail, err := h.svc.GetLot(id)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			fail(c, http.StatusNotFound, "批次不存在")
			return
		}
		fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, ok(detail))
}

// StartProcess POST /api/v1/lots/:id/start
func (h *ProductionHandler) StartProcess(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		fail(c, http.StatusBadRequest, "无效的批次 ID")
		return
	}
	var req dto.StartProcessRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}
	record, err := h.svc.StartProcessing(id, req)
	if err != nil {
		h.writeServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, ok(record))
}

// EndProcess POST /api/v1/lots/:id/end
func (h *ProductionHandler) EndProcess(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		fail(c, http.StatusBadRequest, "无效的批次 ID")
		return
	}
	var req dto.EndProcessRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}
	if err := h.svc.EndProcessing(id, req); err != nil {
		h.writeServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, ok(gin.H{"lot_id": id}))
}

// CompleteLot POST /api/v1/lots/:id/complete
func (h *ProductionHandler) CompleteLot(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		fail(c, http.StatusBadRequest, "无效的批次 ID")
		return
	}
	if err := h.svc.CompleteLot(id); err != nil {
		h.writeServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, ok(gin.H{"lot_id": id}))
}

func (h *ProductionHandler) writeServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrNotFound):
		fail(c, http.StatusBadRequest, "相关记录不存在")
	case errors.Is(err, service.ErrLotNotActive):
		fail(c, http.StatusBadRequest, err.Error())
	case errors.Is(err, service.ErrLotBusy):
		fail(c, http.StatusConflict, err.Error())
	case errors.Is(err, service.ErrMachineNotIdle):
		fail(c, http.StatusConflict, err.Error())
	case errors.Is(err, service.ErrMachineRunning):
		fail(c, http.StatusConflict, err.Error())
	case errors.Is(err, service.ErrNoOpenRecord):
		fail(c, http.StatusBadRequest, err.Error())
	case errors.Is(err, service.ErrInvalidTransition):
		fail(c, http.StatusBadRequest, err.Error())
	default:
		fail(c, http.StatusInternalServerError, err.Error())
	}
}

func parseUintParam(c *gin.Context, name string) (uint, error) {
	v, err := strconv.ParseUint(c.Param(name), 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(v), nil
}
