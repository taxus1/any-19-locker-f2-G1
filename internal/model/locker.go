package model

import (
	"time"

	"gorm.io/gorm"
)

// LockerCell 快递柜格口
// status: FREE(空闲), LOCKED(停用)
type LockerCell struct {
	ID         uint64       `gorm:"primaryKey;column:id" json:"id"`
	CellNo     string       `gorm:"column:cell_no;size:32;not null;uniqueIndex" json:"cellNo"`
	Size       string       `gorm:"column:size;size:16;not null" json:"size"`
	Status     string       `gorm:"column:status;size:16;not null;default:FREE" json:"status"`
	CreateBy   string       `gorm:"column:create_by;size:64" json:"-"`
	CreateTime time.Time    `gorm:"column:create_time" json:"-"`
	UpdateBy   string       `gorm:"column:update_by;size:64" json:"-"`
	UpdateTime time.Time    `gorm:"column:update_time" json:"-"`
	DeletedAt  gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

func (LockerCell) TableName() string { return "t_locker_cell" }

func (c *LockerCell) BeforeCreate(tx *gorm.DB) error {
	op := OperatorFrom(tx.Statement.Context)
	now := time.Now()
	c.CreateBy, c.CreateTime = op, now
	c.UpdateBy, c.UpdateTime = op, now
	return nil
}

func (c *LockerCell) BeforeUpdate(tx *gorm.DB) error {
	c.UpdateBy, c.UpdateTime = OperatorFrom(tx.Statement.Context), time.Now()
	return nil
}

// ParcelSize 包裹尺寸档位
const (
	ParcelSizeSmall   = "SMALL"
	ParcelSizeMedium  = "MEDIUM"
	ParcelSizeLarge   = "LARGE"
)

// CellSize 限制：SMALL < MEDIUM < LARGE
func SizeValue(size string) int {
	switch size {
	case ParcelSizeSmall:
		return 1
	case ParcelSizeMedium:
		return 2
	case ParcelSizeLarge:
		return 3
	default:
		return 0
	}
}

// CanFit 判断包裹尺寸是否能放入格口尺寸
func CanFit(parcelSize, cellSize string) bool {
	return SizeValue(parcelSize) <= SizeValue(cellSize)
}

// LockerParcel 快递柜包裹
// status: STORED(已存入), PICKED(已取出)
type LockerParcel struct {
	ID         uint64       `gorm:"primaryKey;column:id" json:"id"`
	ParcelNo   string       `gorm:"column:parcel_no;size:32;not null;uniqueIndex" json:"parcelNo"`
	CellID     uint64       `gorm:"column:cell_id;not null" json:"-"`
	Phone      string       `gorm:"column:phone;size:32;not null" json:"phone"`
	Size       string       `gorm:"column:size;size:16;not null" json:"size"`
	Status     string       `gorm:"column:status;size:16;not null;default:STORED" json:"status"`
	StoredAt   *time.Time   `gorm:"column:stored_at" json:"storedAt"`
	PickedAt   *time.Time   `gorm:"column:picked_at" json:"pickedAt"`
	CreateBy   string       `gorm:"column:create_by;size:64" json:"-"`
	CreateTime time.Time    `gorm:"column:create_time" json:"-"`
	UpdateBy   string       `gorm:"column:update_by;size:64" json:"-"`
	UpdateTime time.Time    `gorm:"column:update_time" json:"-"`
	DeletedAt  gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

func (LockerParcel) TableName() string { return "t_parcel" }

func (p *LockerParcel) BeforeCreate(tx *gorm.DB) error {
	op := OperatorFrom(tx.Statement.Context)
	now := time.Now()
	p.CreateBy, p.CreateTime = op, now
	p.UpdateBy, p.UpdateTime = op, now
	return nil
}

func (p *LockerParcel) BeforeUpdate(tx *gorm.DB) error {
	p.UpdateBy, p.UpdateTime = OperatorFrom(tx.Statement.Context), time.Now()
	return nil
}