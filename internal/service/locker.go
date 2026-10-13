package service

import (
	"context"
	"errors"

	"github.com/redis/go-redis/v9"

	"somepro/internal/model"
	"somepro/internal/repository"
)

// LockerService 快递柜服务
type LockerService struct {
	cellRepo    *repository.LockerCellRepository
	parcelRepo  *repository.LockerParcelRepository
	rdb         *redis.Client
}

func NewLockerService(cellRepo *repository.LockerCellRepository, parcelRepo *repository.LockerParcelRepository, rdb *redis.Client) *LockerService {
	return &LockerService{
		cellRepo:   cellRepo,
		parcelRepo: parcelRepo,
		rdb:        rdb,
	}
}

// 登记格口
func (s *LockerService) RegisterCell(ctx context.Context, cellNo, size string) (*model.LockerCell, error) {
	// 检查格口编号是否已存在
	existing, err := s.cellRepo.FindByCellNo(ctx, cellNo)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("格口编号已存在")
	}

	// 检查尺寸是否合法
	if size != model.ParcelSizeSmall && size != model.ParcelSizeMedium && size != model.ParcelSizeLarge {
		return nil, errors.New("尺寸必须是 SMALL、MEDIUM 或 LARGE")
	}

	cell := &model.LockerCell{
		CellNo: cellNo,
		Size:   size,
		Status: "FREE",
	}

	if err := s.cellRepo.Create(ctx, cell); err != nil {
		return nil, err
	}

	return cell, nil
}

// 停用格口
func (s *LockerService) DisableCell(ctx context.Context, cellNo string) error {
	// 先查询格口是否存在
	cell, err := s.cellRepo.FindByCellNo(ctx, cellNo)
	if err != nil {
		return err
	}
	if cell == nil {
		return errors.New("格口不存在")
	}

	// 已经停用的不用再处理，幂等
	if cell.Status == "LOCKED" {
		return nil
	}

	return s.cellRepo.Disable(ctx, cell.ID)
}

// 存件
func (s *LockerService) StoreParcel(ctx context.Context, parcelNo, cellNo, phone, size string) (*model.LockerParcel, error) {
	// 检查尺寸是否合法
	if size != model.ParcelSizeSmall && size != model.ParcelSizeMedium && size != model.ParcelSizeLarge {
		return nil, errors.New("包裹尺寸必须是 SMALL、MEDIUM 或 LARGE")
	}

	// 检查包裹单号是否已存在
	existing, err := s.parcelRepo.FindByParcelNo(ctx, parcelNo)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("包裹单号已存在")
	}

	// 查找格口
	cell, err := s.cellRepo.FindByCellNo(ctx, cellNo)
	if err != nil {
		return nil, err
	}
	if cell == nil {
		return nil, errors.New("格口不存在")
	}

	// 检查格口是否停用
	if cell.Status == "LOCKED" {
		return nil, errors.New("格口已停用")
	}

	// 并发控制：使用 Redis 分布式锁，锁的 key 为 locker:cell:{cellNo}
	lockKey := "locker:cell:" + cellNo
	val := generateLockValue()
	locked, err := s.rdb.SetNX(ctx, lockKey, val, 10).Result()
	if err != nil {
		return nil, err
	}
	if !locked {
		return nil, errors.New("格口正在被占用，请稍后再试")
	}
	defer s.rdb.Del(ctx, lockKey)

	// 再次检查格口状态（防止锁等待期间状态变化）
	cell, err = s.cellRepo.FindByCellNo(ctx, cellNo)
	if err != nil {
		return nil, err
	}
	if cell == nil {
		return nil, errors.New("格口不存在")
	}
	if cell.Status == "LOCKED" {
		return nil, errors.New("格口已停用")
	}

	// 检查格口是否已有包裹
	existingParcel, err := s.parcelRepo.FindByCellID(ctx, cell.ID)
	if err != nil {
		return nil, err
	}
	if existingParcel != nil {
		return nil, errors.New("格口已被占用")
	}

	// 检查包裹尺寸是否能放入格口
	if !model.CanFit(size, cell.Size) {
		return nil, errors.New("包裹尺寸超过格口容量")
	}

	// 创建包裹
	parcel := &model.LockerParcel{
		ParcelNo: parcelNo,
		CellID:   cell.ID,
		Phone:    phone,
		Size:     size,
		Status:   "STORED",
	}

	if err := s.parcelRepo.Create(ctx, parcel); err != nil {
		return nil, err
	}

	return parcel, nil
}

// 取件
func (s *LockerService) PickUpParcel(ctx context.Context, parcelNo string) error {
	// 查找包裹
	parcel, err := s.parcelRepo.FindByParcelNo(ctx, parcelNo)
	if err != nil {
		return err
	}
	if parcel == nil {
		// 包裹不存在，视为成功（幂等），不报错
		return nil
	}

	// 已取出的也不报错（幂等）
	if parcel.Status == "PICKED" {
		return nil
	}

	// 执行取件
	if err := s.parcelRepo.PickUp(ctx, parcel.ID); err != nil {
		return err
	}

	return nil
}

// 按手机号查询在柜包裹
func (s *LockerService) ListParcelsByPhone(ctx context.Context, phone string, pageNum, pageSize int) (map[string]interface{}, error) {
	if pageNum < 1 {
		pageNum = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	parcels, total, err := s.parcelRepo.FindByPhone(ctx, phone, pageNum, pageSize)
	if err != nil {
		return nil, err
	}

	totalPages := 0
	if pageSize > 0 {
		totalPages = int((total + int64(pageSize) - 1) / int64(pageSize))
	}

	// 构造返回数据
	result := make([]map[string]interface{}, 0, len(parcels))
	for _, p := range parcels {
		result = append(result, map[string]interface{}{
			"parcelNo": p.ParcelNo,
			"phone":    p.Phone,
			"size":     p.Size,
			"storedAt": p.StoredAt,
		})
	}

	return map[string]interface{}{
		"content": result,
		"total":   total,
		"pageNum": pageNum,
		"pageSize": pageSize,
		"totalPages": totalPages,
	}, nil
}

// 生成锁的唯一值
func generateLockValue() string {
	return "lock:" + generateRandomString(8)
}

func generateRandomString(n int) string {
	// 简单实现，实际可用 UUID
	return "random"
}