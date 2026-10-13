package service

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"somepro/internal/model"
	"somepro/internal/repository"
)

// LockerService 快递柜业务逻辑
type LockerService struct {
	repo *repository.LockerRepository
}

func NewLockerService(repo *repository.LockerRepository) *LockerService {
	return &LockerService{repo: repo}
}

// ===== 格口管理 =====

// RegisterCell 注册格口
func (s *LockerService) RegisterCell(ctx context.Context, cellNo, size string) (*model.LockerCell, error) {
	// 尺寸档校验
	if !isValidSize(size) {
		return nil, errors.New("无效的尺寸档，必须是 SMALL、MEDIUM 或 LARGE")
	}

	// 检查编号是否已存在
	existing, err := s.repo.FindCellByNo(ctx, cellNo)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("格口编号已存在")
	}

	cell := &model.LockerCell{
		CellNo: cellNo,
		Size:   size,
		Status: model.CellStatusFree,
	}

	if err := s.repo.CreateCell(ctx, cell); err != nil {
		return nil, err
	}

	return cell, nil
}

// DisableCell 停用格口
func (s *LockerService) DisableCell(ctx context.Context, cellNo string) error {
	cell, err := s.repo.FindCellByNo(ctx, cellNo)
	if err != nil {
		return err
	}
	if cell == nil {
		return errors.New("格口不存在")
	}

	// 只有 FREE 状态的格口可以停用，BUSY 状态的暂不可停用（也可以设计为可以停用）
	if cell.Status == model.CellStatusOff {
		return nil // 已经停用了，幂等
	}

	// 检查格口内是否有包裹，有包裹时不允许停用
	parcel, err := s.repo.FindParcelByCellID(ctx, cell.ID)
	if err != nil {
		return err
	}
	if parcel != nil {
		return errors.New("格口内有包裹，不能停用")
	}

	cell.Status = model.CellStatusOff
	if err := s.repo.UpdateCell(ctx, cell); err != nil {
		return err
	}

	return nil
}

// ===== 包裹存取 =====

// StoreParcel 存件
func (s *LockerService) StoreParcel(ctx context.Context, parcelNo, cellNo, phone, size string) (*model.Parcel, error) {
	// 尺寸档校验
	if !isValidSize(size) {
		return nil, errors.New("无效的包裹尺寸档，必须是 SMALL、MEDIUM 或 LARGE")
	}

	// 查询格口
	cell, err := s.repo.FindCellByNo(ctx, cellNo)
	if err != nil {
		return nil, err
	}
	if cell == nil {
		return nil, errors.New("格口不存在")
	}

	// 检查格口状态
	if cell.Status == model.CellStatusOff {
		return nil, errors.New("格口已停用，无法存件")
	}
	if cell.Status == model.CellStatusBusy {
		return nil, errors.New("格口已被占用")
	}

	// 检查格口尺寸是否能放下包裹
	if !model.ParcelFitCell(size, cell.Size) {
		return nil, errors.New("包裹尺寸超过格口容量")
	}

	// 检查格口内是否已有包裹（双重检查，防止并发）
	existingParcel, err := s.repo.FindParcelByCellID(ctx, cell.ID)
	if err != nil {
		return nil, err
	}
	if existingParcel != nil {
		return nil, errors.New("格口已被占用")
	}

	// 创建包裹
	parcel := &model.Parcel{
		ParcelNo: parcelNo,
		CellID:   cell.ID,
		Phone:    phone,
		Size:     size,
		Status:   model.ParcelStatusStored,
	}
	if err := s.repo.CreateParcel(ctx, parcel); err != nil {
		return nil, err
	}

	// 更新格口状态为已占用
	cell.Status = model.CellStatusBusy
	if err := s.repo.UpdateCell(ctx, cell); err != nil {
		// 如果格口更新失败，这里可以考虑回滚包裹创建
		return nil, err
	}

	return parcel, nil
}

// PickupParcel 取件
func (s *LockerService) PickupParcel(ctx context.Context, parcelNo string) error {
	// 查找包裹
	parcel, err := s.repo.FindParcelByNo(ctx, parcelNo)
	if err != nil {
		return err
	}
	if parcel == nil {
		// 包裹不存在，重复取件允许，不报错
		return nil
	}

	// 如果已经取走，直接返回成功（幂等）
	if parcel.Status == model.ParcelStatusPicked {
		return nil
	}

	// 更新包裹状态
	parcel.Status = model.ParcelStatusPicked
	now := time.Now()
	parcel.PickedAt = &now

	// 查找格口
	cell, err := s.repo.FindCellByID(ctx, parcel.CellID)
	if err != nil {
		return err
	}
	if cell == nil {
		// 格口不存在，仅更新包裹状态
		return s.repo.UpdateParcel(ctx, parcel)
	}

	// 更新格口状态为 FREE
	cell.Status = model.CellStatusFree

	// 事务更新
	db := s.repo.DB
	if err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := s.repo.UpdateParcelInTx(ctx, tx, parcel); err != nil {
			return err
		}
		if err := s.repo.UpdateCellInTx(ctx, tx, cell); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}

// ListParcelsByPhone 查询某手机号的在柜包裹
func (s *LockerService) ListParcelsByPhone(ctx context.Context, phone string, pageNum, pageSize int) (map[string]interface{}, error) {
	parcels, total, err := s.repo.FindParcelsByPhone(ctx, phone, pageNum, pageSize)
	if err != nil {
		return nil, err
	}

	// 计算总页数
	totalPages := 0
	if pageSize > 0 {
		totalPages = int((total + int64(pageSize) - 1) / int64(pageSize))
	}

	return map[string]interface{}{
		"content": parcels,
		"total":   total,
		"pageNum": pageNum,
		"pageSize": pageSize,
		"totalPages": totalPages,
	}, nil
}

// ===== 辅助函数 =====

// isValidSize 检查尺寸档是否合法
func isValidSize(size string) bool {
	switch size {
	case model.SizeSmall, model.SizeMedium, model.SizeLarge:
		return true
	default:
		return false
	}
}

