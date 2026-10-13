package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"somepro/internal/common"
	"somepro/internal/service"
)

// LockerHandler 快递柜处理器
type LockerHandler struct {
	svc *service.LockerService
}

func NewLockerHandler(svc *service.LockerService) *LockerHandler {
	return &LockerHandler{svc: svc}
}

// RegisterCell 登记格口
// POST /api/locker/cell
// 参数: cellNo, size
func (h *LockerHandler) RegisterCell(c *gin.Context) {
	cellNo := c.Query("cellNo")
	if cellNo == "" {
		cellNo = c.PostForm("cellNo")
	}
	size := c.Query("size")
	if size == "" {
		size = c.PostForm("size")
	}

	if cellNo == "" {
		c.JSON(http.StatusOK, common.Fail("格口编号不能为空"))
		return
	}
	if size == "" {
		c.JSON(http.StatusOK, common.Fail("尺寸不能为空"))
		return
	}

	cell, err := h.svc.RegisterCell(c.Request.Context(), cellNo, size)
	if err != nil {
		c.JSON(http.StatusOK, common.Fail(err.Error()))
		return
	}
	c.JSON(http.StatusOK, common.OK(cell))
}

// DisableCell 停用格口
// POST /api/locker/cell/disable
// 参数: cellNo
func (h *LockerHandler) DisableCell(c *gin.Context) {
	cellNo := c.Query("cellNo")
	if cellNo == "" {
		cellNo = c.PostForm("cellNo")
	}

	if cellNo == "" {
		c.JSON(http.StatusOK, common.Fail("格口编号不能为空"))
		return
	}

	if err := h.svc.DisableCell(c.Request.Context(), cellNo); err != nil {
		c.JSON(http.StatusOK, common.Fail(err.Error()))
		return
	}
	c.JSON(http.StatusOK, common.OK(nil))
}

// StoreParcel 存件
// POST /api/locker/parcel
// 参数: parcelNo, cellNo, phone, size
func (h *LockerHandler) StoreParcel(c *gin.Context) {
	parcelNo := c.Query("parcelNo")
	if parcelNo == "" {
		parcelNo = c.PostForm("parcelNo")
	}
	cellNo := c.Query("cellNo")
	if cellNo == "" {
		cellNo = c.PostForm("cellNo")
	}
	phone := c.Query("phone")
	if phone == "" {
		phone = c.PostForm("phone")
	}
	size := c.Query("size")
	if size == "" {
		size = c.PostForm("size")
	}

	if parcelNo == "" {
		c.JSON(http.StatusOK, common.Fail("包裹单号不能为空"))
		return
	}
	if cellNo == "" {
		c.JSON(http.StatusOK, common.Fail("格口编号不能为空"))
		return
	}
	if phone == "" {
		c.JSON(http.StatusOK, common.Fail("收货手机号不能为空"))
		return
	}
	if size == "" {
		c.JSON(http.StatusOK, common.Fail("包裹尺寸不能为空"))
		return
	}

	parcel, err := h.svc.StoreParcel(c.Request.Context(), parcelNo, cellNo, phone, size)
	if err != nil {
		c.JSON(http.StatusOK, common.Fail(err.Error()))
		return
	}
	c.JSON(http.StatusOK, common.OK(parcel))
}

// PickUpParcel 取件
// POST /api/locker/parcel/pickup
// 参数: parcelNo
func (h *LockerHandler) PickUpParcel(c *gin.Context) {
	parcelNo := c.Query("parcelNo")
	if parcelNo == "" {
		parcelNo = c.PostForm("parcelNo")
	}

	if parcelNo == "" {
		c.JSON(http.StatusOK, common.Fail("包裹单号不能为空"))
		return
	}

	if err := h.svc.PickUpParcel(c.Request.Context(), parcelNo); err != nil {
		c.JSON(http.StatusOK, common.Fail(err.Error()))
		return
	}
	c.JSON(http.StatusOK, common.OK(nil))
}

// ListParcelsByPhone 查询用户包裹
// GET /api/locker/parcel/list
// 参数: phone, pageNum, pageSize
func (h *LockerHandler) ListParcelsByPhone(c *gin.Context) {
	phone := c.Query("phone")
	pageNum, _ := strconv.Atoi(c.DefaultQuery("pageNum", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	if phone == "" {
		c.JSON(http.StatusOK, common.Fail("手机号不能为空"))
		return
	}

	data, err := h.svc.ListParcelsByPhone(c.Request.Context(), phone, pageNum, pageSize)
	if err != nil {
		c.JSON(http.StatusOK, common.Fail(err.Error()))
		return
	}
	c.JSON(http.StatusOK, common.OK(data))
}