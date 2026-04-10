package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"iot/internal/model"
	"iot/internal/service"
	"iot/pkg/response"
)

type DeviceHandler struct {
	svc *service.DeviceService
}

func NewDeviceHandler(svc *service.DeviceService) *DeviceHandler {
	return &DeviceHandler{svc: svc}
}

// CreateDevice godoc
// @Summary 创建设备
// @Tags 设备管理
// @Accept json
// @Produce json
// @Param body body model.DeviceCreate true "设备信息"
// @Success 200 {object} response.Response
// @Router /api/v1/devices [post]
func (h *DeviceHandler) CreateDevice(c *gin.Context) {
	var req model.DeviceCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	device := model.Device{
		VIN:         req.VIN,
		CarPlate:    req.CarPlate,
		Power:       req.Power,
		Speed:       req.Speed,
		Status:      req.Status,
		Lat:         req.Lat,
		Lon:         req.Lon,
		Description: req.Description,
	}

	if err := h.svc.Create(c.Request.Context(), &device); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, device)
}

// GetDevice godoc
// @Summary 获取设备详情
// @Tags 设备管理
// @Produce json
// @Param id path int true "设备ID"
// @Success 200 {object} response.Response
// @Router /api/v1/devices/{id} [get]
func (h *DeviceHandler) GetDevice(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	device, err := h.svc.GetByID(c.Request.Context(), uint(id))
	if err != nil {
		response.NotFound(c, "device not found")
		return
	}

	response.Success(c, device)
}

// ListDevices godoc
// @Summary 获取设备列表
// @Tags 设备管理
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Success 200 {object} response.Response
// @Router /api/v1/devices [get]
func (h *DeviceHandler) ListDevices(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	devices, total, err := h.svc.List(c.Request.Context(), page, pageSize)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":      0,
		"message":   "success",
		"data":      devices,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// UpdateDevice godoc
// @Summary 更新设备
// @Tags 设备管理
// @Accept json
// @Produce json
// @Param id path int true "设备ID"
// @Param body body model.DeviceUpdate true "设备信息"
// @Success 200 {object} response.Response
// @Router /api/v1/devices/{id} [put]
func (h *DeviceHandler) UpdateDevice(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	var req model.DeviceUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	updates := make(map[string]interface{})
	if req.VIN != nil {
		updates["vin"] = *req.VIN
	}
	if req.CarPlate != nil {
		updates["car_plate"] = *req.CarPlate
	}
	if req.Power != nil {
		updates["power"] = *req.Power
	}
	if req.Speed != nil {
		updates["speed"] = *req.Speed
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.Lat != nil {
		updates["lat"] = *req.Lat
	}
	if req.Lon != nil {
		updates["lon"] = *req.Lon
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}

	if len(updates) == 0 {
		response.BadRequest(c, "no fields to update")
		return
	}

	device, err := h.svc.UpdateFields(c.Request.Context(), uint(id), updates)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, device)
}

// DeleteDevice godoc
// @Summary 删除设备
// @Tags 设备管理
// @Produce json
// @Param id path int true "设备ID"
// @Success 200 {object} response.Response
// @Router /api/v1/devices/{id} [delete]
func (h *DeviceHandler) DeleteDevice(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	if err := h.svc.Delete(c.Request.Context(), uint(id)); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, nil)
}
