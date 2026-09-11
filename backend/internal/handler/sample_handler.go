package handler

import (
	"errors"
	"net/http"
	"strconv"

	"sample-tracker/internal/dto"
	"sample-tracker/internal/service"

	"github.com/gin-gonic/gin"
)

type SampleHandler struct {
	svc      *service.SampleService
	transfer *service.TransferService
	result   *service.ResultService
}

func NewSampleHandler(s *service.SampleService, t *service.TransferService, r *service.ResultService) *SampleHandler {
	return &SampleHandler{svc: s, transfer: t, result: r}
}

// Create POST /api/v1/samples
func (h *SampleHandler) Create(c *gin.Context) {
	var req dto.CreateSampleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}
	sample, err := h.svc.Create(req)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			fail(c, http.StatusBadRequest, "接收人或存放位置不存在")
			return
		}
		fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, ok(gin.H{"id": sample.ID, "code": sample.Code}))
}

// List GET /api/v1/samples
func (h *SampleHandler) List(c *gin.Context) {
	var q dto.SampleQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		fail(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}
	page, err := h.svc.List(q)
	if err != nil {
		fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, ok(page))
}

// Detail GET /api/v1/samples/:id
func (h *SampleHandler) Detail(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		fail(c, http.StatusBadRequest, "无效的样品 ID")
		return
	}
	detail, err := h.svc.Get(id)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			fail(c, http.StatusNotFound, "样品不存在")
			return
		}
		fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, ok(detail))
}

// CreateTransfer POST /api/v1/samples/:id/transfers
func (h *SampleHandler) CreateTransfer(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		fail(c, http.StatusBadRequest, "无效的样品 ID")
		return
	}
	var req dto.CreateTransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}
	req.SampleID = id
	vo, err := h.transfer.Create(req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrNotFound):
			fail(c, http.StatusBadRequest, "样品、位置或操作人不存在")
		case errors.Is(err, service.ErrSameLocation):
			fail(c, http.StatusBadRequest, "目标位置与来源位置不能相同")
		default:
			fail(c, http.StatusInternalServerError, err.Error())
		}
		return
	}
	c.JSON(http.StatusOK, ok(vo))
}

// CreateResult POST /api/v1/samples/:id/results
func (h *SampleHandler) CreateResult(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		fail(c, http.StatusBadRequest, "无效的样品 ID")
		return
	}
	var req dto.CreateResultRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}
	req.SampleID = id
	if err := h.result.Create(req); err != nil {
		if errors.Is(err, service.ErrNotFound) {
			fail(c, http.StatusBadRequest, "样品或检测人不存在")
			return
		}
		fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, ok(gin.H{"sample_id": id}))
}

func parseID(c *gin.Context) (uint, error) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}
