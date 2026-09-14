package repository

import (
	"context"

	"sample-tracker/internal/model"

	"gorm.io/gorm"
)

// MetaRepository 基础数据：操作员与位置
type MetaRepository struct{ db *gorm.DB }

func (r *MetaRepository) ListUsers(ctx context.Context) ([]model.User, error) {
	var users []model.User
	err := r.db.WithContext(ctx).Order("id").Find(&users).Error
	return users, err
}

func (r *MetaRepository) GetUser(ctx context.Context, id uint) (*model.User, error) {
	var u model.User
	if err := r.db.WithContext(ctx).First(&u, id).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *MetaRepository) ListLocations(ctx context.Context, locType string) ([]model.Location, error) {
	q := r.db.WithContext(ctx).Model(&model.Location{})
	if locType != "" {
		q = q.Where("type = ?", locType)
	}
	var locs []model.Location
	err := q.Order("type, id").Find(&locs).Error
	return locs, err
}

func (r *MetaRepository) GetLocation(ctx context.Context, id uint) (*model.Location, error) {
	var loc model.Location
	if err := r.db.WithContext(ctx).First(&loc, id).Error; err != nil {
		return nil, err
	}
	return &loc, nil
}
