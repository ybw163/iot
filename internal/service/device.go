package service

import (
	"context"

	"go.uber.org/zap"

	"iot/internal/model"
	"iot/internal/repository"
	"iot/pkg/logger"
)

type DeviceService struct {
	repo *repository.DeviceRepository
}

func NewDeviceService(repo *repository.DeviceRepository) *DeviceService {
	return &DeviceService{repo: repo}
}

func (s *DeviceService) Create(ctx context.Context, device *model.Device) error {
	if err := s.repo.Create(ctx, device); err != nil {
		logger.Log.Error("create device failed", zap.Error(err))
		return err
	}
	return nil
}

func (s *DeviceService) GetByID(ctx context.Context, id uint) (*model.Device, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *DeviceService) List(ctx context.Context, page, pageSize int) ([]model.Device, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.List(ctx, offset, pageSize)
}

func (s *DeviceService) Update(ctx context.Context, device *model.Device) error {
	if err := s.repo.Update(ctx, device); err != nil {
		logger.Log.Error("update device failed", zap.Error(err))
		return err
	}
	return nil
}

func (s *DeviceService) UpdateFields(ctx context.Context, id uint, updates map[string]interface{}) (*model.Device, error) {
	if err := s.repo.UpdateFields(ctx, id, updates); err != nil {
		logger.Log.Error("update device fields failed", zap.Error(err))
		return nil, err
	}
	return s.repo.GetByID(ctx, id)
}

func (s *DeviceService) Delete(ctx context.Context, id uint) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		logger.Log.Error("delete device failed", zap.Error(err))
		return err
	}
	return nil
}
