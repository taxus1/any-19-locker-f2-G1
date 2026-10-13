package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"somepro/internal/model"
)

// LockerCellRepository 格口仓储
type LockerCellRepository struct {
	db *gorm.DB
}

func NewLockerCellRepository(db *gorm.DB) *LockerCellRepository {
	return &LockerCellRepository{db: db}
}

// Create 创建格口
func (r *LockerCellRepository) Create(ctx context.Context, cell *model.LockerCell) error {
	return r.db.WithContext(ctx).Create(cell).Error
}

// FindByCellNo 通过格口编号查询（包含已删除的）
func (r *LockerCellRepository) FindByCellNo(ctx context.Context, cellNo string) (*model.LockerCell, error) {
	var cell model.LockerCell
	if err := r.db.WithContext(ctx).Where("cell_no = ?", cellNo).First(&cell).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &cell, nil
}

// FindByID 通过 ID 查询（自动排除已删除）
func (r *LockerCellRepository) FindByID(ctx context.Context, id uint64) (*model.LockerCell, error) {
	var cell model.LockerCell
	if err := r.db.WithContext(ctx).First(&cell, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &cell, nil
}

// FindAllFree 可用格口列表（空闲且未停用）
func (r *LockerCellRepository) FindAllFree(ctx context.Context) ([]model.LockerCell, error) {
	var cells []model.LockerCell
	if err := r.db.WithContext(ctx).Where("status = 'FREE'").Find(&cells).Error; err != nil {
		return nil, err
	}
	return cells, nil
}

// Disable 停用格口
func (r *LockerCellRepository) Disable(ctx context.Context, cellID uint64) error {
	return r.db.WithContext(ctx).Model(&model.LockerCell{}).
		Where("id = ?", cellID).
		Update("status", "LOCKED").Error
}

// LockerParcelRepository 包裹仓储
type LockerParcelRepository struct {
	db *gorm.DB
}

func NewLockerParcelRepository(db *gorm.DB) *LockerParcelRepository {
	return &LockerParcelRepository{db: db}
}

// Create 创建包裹
func (r *LockerParcelRepository) Create(ctx context.Context, parcel *model.LockerParcel) error {
	return r.db.WithContext(ctx).Create(parcel).Error
}

// FindByParcelNo 通过包裹单号查询
func (r *LockerParcelRepository) FindByParcelNo(ctx context.Context, parcelNo string) (*model.LockerParcel, error) {
	var parcel model.LockerParcel
	if err := r.db.WithContext(ctx).Where("parcel_no = ?", parcelNo).First(&parcel).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &parcel, nil
}

// FindByCellID 通过格口ID查询当前存放的包裹
func (r *LockerParcelRepository) FindByCellID(ctx context.Context, cellID uint64) (*model.LockerParcel, error) {
	var parcel model.LockerParcel
	if err := r.db.WithContext(ctx).Where("cell_id = ? AND status = 'STORED'", cellID).First(&parcel).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &parcel, nil
}

// FindByPhone 分页查询用户的包裹
func (r *LockerParcelRepository) FindByPhone(ctx context.Context, phone string, pageNum, pageSize int) ([]model.LockerParcel, int64, error) {
	if pageNum < 1 {
		pageNum = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	var total int64
	if err := r.db.WithContext(ctx).
		Model(&model.LockerParcel{}).
		Where("phone = ? AND status = 'STORED'", phone).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var parcels []model.LockerParcel
	if err := r.db.WithContext(ctx).
		Model(&model.LockerParcel{}).
		Where("phone = ? AND status = 'STORED'", phone).
		Order("id DESC").
		Offset((pageNum - 1) * pageSize).
		Limit(pageSize).
		Find(&parcels).Error; err != nil {
		return nil, 0, err
	}

	return parcels, total, nil
}

// PickUp 取件（更新状态为已取出）
func (r *LockerParcelRepository) PickUp(ctx context.Context, parcelID uint64) error {
	return r.db.WithContext(ctx).Model(&model.LockerParcel{}).
		Where("id = ?", parcelID).
		Updates(map[string]interface{}{
			"status":    "PICKED",
			"picked_at": gorm.Expr("NOW()"),
		}).Error
}

// FindFreeCell 取一个空闲格口（用于并发场景）
func (r *LockerParcelRepository) FindFreeCell(ctx context.Context, cellID uint64) (*model.LockerCell, error) {
	var cell model.LockerCell
	if err := r.db.WithContext(ctx).Where("id = ? AND status = 'FREE'", cellID).First(&cell).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("格口已被占用或不存在")
		}
		return nil, err
	}
	return &cell, nil
}