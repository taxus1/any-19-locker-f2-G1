package model

import (
	"time"

	"gorm.io/gorm"
)

// LockerCell 状态常量
const (
	CellStatusFree  = "FREE"    // 空闲，可存件
	CellStatusBusy  = "BUSY"    // 已占用
	CellStatusOff   = "OFFLINE" // 已停用，不可存件
)

// Parcel 状态常量
const (
	ParcelStatusStored = "STORED" // 已存储，在柜
	ParcelStatusPicked = "PICKED" // 已取件
)

// Size 模型尺寸档（格口和包裹共用）
const (
	SizeSmall  = "SMALL"
	SizeMedium = "MEDIUM"
	SizeLarge  = "LARGE"
)

// LockerCell 餐柜格口模型
type LockerCell struct {
	ID          uint           `gorm:"primaryKey;column:id" json:"id"`
	CellNo      string         `gorm:"column:cell_no;size:32;not null;uniqueIndex" json:"cellNo"`
	Size        string         `gorm:"column:size;size:16;not null" json:"size"`
	Status      string         `gorm:"column:status;size:16;not null;default:FREE" json:"status"`
	CreateBy    string         `gorm:"column:create_by;size:64" json:"-"`
	CreateTime  time.Time      `gorm:"column:create_time" json:"-"`
	UpdateBy    string         `gorm:"column:update_by;size:64" json:"-"`
	UpdateTime  time.Time      `gorm:"column:update_time" json:"-"`
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

// TableName 返回表名
func (LockerCell) TableName() string { return "t_locker_cell" }

// BeforeCreate GORM 钩子：创建时填充审计字段
func (c *LockerCell) BeforeCreate(tx *gorm.DB) error {
	op := OperatorFrom(tx.Statement.Context)
	now := time.Now()
	c.CreateBy, c.CreateTime = op, now
	c.UpdateBy, c.UpdateTime = op, now
	return nil
}

// BeforeUpdate GORM 钩子：更新时填充更新字段
func (c *LockerCell) BeforeUpdate(tx *gorm.DB) error {
	c.UpdateBy, c.UpdateTime = OperatorFrom(tx.Statement.Context), time.Now()
	return nil
}

// Parcel 包裹模型
type Parcel struct {
	ID          uint           `gorm:"primaryKey;column:id" json:"id"`
	ParcelNo    string         `gorm:"column:parcel_no;size:32;not null;uniqueIndex" json:"parcelNo"`
	CellID      uint           `gorm:"column:cell_id;not null" json:"cellId"`
	Phone       string         `gorm:"column:phone;size:32;not null" json:"phone"`
	Size        string         `gorm:"column:size;size:16;not null" json:"size"`
	Status      string         `gorm:"column:status;size:16;not null;default:STORED" json:"status"`
	StoredAt    *time.Time     `gorm:"column:stored_at" json:"storedAt,omitempty"`
	PickedAt    *time.Time     `gorm:"column:picked_at" json:"pickedAt,omitempty"`
	CreateBy    string         `gorm:"column:create_by;size:64" json:"-"`
	CreateTime  time.Time      `gorm:"column:create_time" json:"-"`
	UpdateBy    string         `gorm:"column:update_by;size:64" json:"-"`
	UpdateTime  time.Time      `gorm:"column:update_time" json:"-"`
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

// TableName 返回表名
func (Parcel) TableName() string { return "t_parcel" }

// BeforeCreate GORM 钩子：创建时填充审计字段
func (p *Parcel) BeforeCreate(tx *gorm.DB) error {
	op := OperatorFrom(tx.Statement.Context)
	now := time.Now()
	p.CreateBy, p.CreateTime = op, now
	p.UpdateBy, p.UpdateTime = op, now
	return nil
}

// BeforeUpdate GORM 钩子：更新时填充更新字段
func (p *Parcel) BeforeUpdate(tx *gorm.DB) error {
	p.UpdateBy, p.UpdateTime = OperatorFrom(tx.Statement.Context), time.Now()
	return nil
}

// SizeLevel 将尺寸档转换为数值级别用于比较
func SizeLevel(size string) int {
	switch size {
	case SizeSmall:
		return 1
	case SizeMedium:
		return 2
	case SizeLarge:
		return 3
	default:
		return 0
	}
}

// 包裹尺寸是否小于等于格口尺寸
func ParcelFitCell(parcelSize, cellSize string) bool {
	return SizeLevel(parcelSize) <= SizeLevel(cellSize)
}

// Now 返回当前时间（供 service 层使用）
func Now() time.Time {
	return time.Now()
}