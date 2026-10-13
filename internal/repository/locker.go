package repository

import (
	"context"

	"gorm.io/gorm"

	"somepro/internal/model"
)

// LockerRepository 快递柜仓储，操作 LockerCell 和 Parcel
type LockerRepository struct {
	DB *gorm.DB
}

func NewLockerRepository(db *gorm.DB) *LockerRepository {
	return &LockerRepository{DB: db}
}


// ===== LockerCell 相关操作 =====

// FindCellByNo 根据编号查找格口（不含已删除）
func (r *LockerRepository) FindCellByNo(ctx context.Context, cellNo string) (*model.LockerCell, error) {
	var cell model.LockerCell
	if err := r.DB.WithContext(ctx).Where("cell_no = ? AND deleted_at IS NULL", cellNo).First(&cell).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &cell, nil
}

// CreateCell 创建格口
func (r *LockerRepository) CreateCell(ctx context.Context, cell *model.LockerCell) error {
	return r.DB.WithContext(ctx).Create(cell).Error
}

// UpdateCell 更新格口
func (r *LockerRepository) UpdateCell(ctx context.Context, cell *model.LockerCell) error {
	return r.DB.WithContext(ctx).Save(cell).Error
}

// ===== Parcel 相关操作 =====

// FindParcelByNo 根据单号查找包裹（不含已删除）
func (r *LockerRepository) FindParcelByNo(ctx context.Context, parcelNo string) (*model.Parcel, error) {
	var parcel model.Parcel
	if err := r.DB.WithContext(ctx).Where("parcel_no = ? AND deleted_at IS NULL", parcelNo).First(&parcel).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &parcel, nil
}

// FindParcelByCellID 根据格口ID查找包裹（不含已删除）
func (r *LockerRepository) FindParcelByCellID(ctx context.Context, cellID uint) (*model.Parcel, error) {
	var parcel model.Parcel
	if err := r.DB.WithContext(ctx).Where("cell_id = ? AND deleted_at IS NULL", cellID).First(&parcel).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &parcel, nil
}

// CreateParcel 创建包裹
func (r *LockerRepository) CreateParcel(ctx context.Context, parcel *model.Parcel) error {
	return r.DB.WithContext(ctx).Create(parcel).Error
}

// UpdateParcel 更新包裹
func (r *LockerRepository) UpdateParcel(ctx context.Context, parcel *model.Parcel) error {
	return r.DB.WithContext(ctx).Save(parcel).Error
}

// UpdateCellInTx 事务中更新格口
func (r *LockerRepository) UpdateCellInTx(ctx context.Context, tx *gorm.DB, cell *model.LockerCell) error {
	return tx.WithContext(ctx).Save(cell).Error
}

// UpdateParcelInTx 事务中更新包裹
func (r *LockerRepository) UpdateParcelInTx(ctx context.Context, tx *gorm.DB, parcel *model.Parcel) error {
	return tx.WithContext(ctx).Save(parcel).Error
}

// FindParcelsByPhone 分页查询某手机号的在柜包裹
func (r *LockerRepository) FindParcelsByPhone(ctx context.Context, phone string, pageNum, pageSize int) ([]model.Parcel, int64, error) {
	if pageNum < 1 {
		pageNum = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	// 统计总数
	var total int64
	if err := r.DB.WithContext(ctx).
		Where("phone = ? AND status = 'STORED' AND deleted_at IS NULL", phone).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 查询分页数据
	var parcels []model.Parcel
	if err := r.DB.WithContext(ctx).
		Where("phone = ? AND status = 'STORED' AND deleted_at IS NULL", phone).
		Order("id DESC").
		Offset((pageNum - 1) * pageSize).
		Limit(pageSize).
		Find(&parcels).Error; err != nil {
		return nil, 0, err
	}

	return parcels, total, nil
}

// FindCellByID 根据ID查找格口
func (r *LockerRepository) FindCellByID(ctx context.Context, cellID uint) (*model.LockerCell, error) {
	var cell model.LockerCell
	if err := r.DB.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", cellID).First(&cell).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &cell, nil
}

// FindAllCells 获取所有格口（用于初始化检查）
func (r *LockerRepository) FindAllCells(ctx context.Context) ([]model.LockerCell, error) {
	var cells []model.LockerCell
	if err := r.DB.WithContext(ctx).Where("deleted_at IS NULL").Find(&cells).Error; err != nil {
		return nil, err
	}
	return cells, nil
}

// FindAllParcels 获取所有包裹（用于初始化检查）
func (r *LockerRepository) FindAllParcels(ctx context.Context) ([]model.Parcel, error) {
	var parcels []model.Parcel
	if err := r.DB.WithContext(ctx).Where("deleted_at IS NULL").Find(&parcels).Error; err != nil {
		return nil, err
	}
	return parcels, nil
}