package repository

import (
	"context"

	"gorm.io/gorm"

	"iot/internal/model"
)

type DeviceRepository struct {
	db *gorm.DB
}

func NewDeviceRepository(db *gorm.DB) *DeviceRepository {
	return &DeviceRepository{db: db}
}

func (r *DeviceRepository) Create(ctx context.Context, device *model.Device) error {
	return r.db.WithContext(ctx).Create(device).Error
}

func (r *DeviceRepository) GetByID(ctx context.Context, id uint) (*model.Device, error) {
	var device model.Device
	if err := r.db.WithContext(ctx).First(&device, id).Error; err != nil {
		return nil, err
	}
	return &device, nil
}

func (r *DeviceRepository) GetByVin(ctx context.Context, vin string) (*model.Device, error) {
	var device model.Device
	if err := r.db.WithContext(ctx).Where("vin = ?", vin).First(&device).Error; err != nil {
		return nil, err
	}
	return &device, nil
}

func (r *DeviceRepository) List(ctx context.Context, offset, limit int) ([]model.Device, int64, error) {
	var devices []model.Device
	var total int64

	db := r.db.WithContext(ctx).Model(&model.Device{})
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Offset(offset).Limit(limit).Find(&devices).Error; err != nil {
		return nil, 0, err
	}

	return devices, total, nil
}

func (r *DeviceRepository) Update(ctx context.Context, device *model.Device) error {
	return r.db.WithContext(ctx).Save(device).Error
}

func (r *DeviceRepository) UpdateFields(ctx context.Context, id uint, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&model.Device{}).Where("id = ?", id).Updates(updates).Error
}

func (r *DeviceRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.Device{}, id).Error
}
