package activitylog

import (
	"gorm.io/gorm"

	"go-cp-boilerplate/internal/domain"
)

type Repository interface {
	List(page, perPage int, filter Filter) ([]domain.ActivityLog, int64, error)
	FindByID(id uint) (*domain.ActivityLog, error)
	Delete(id uint) error
	ListModules() ([]string, error)
	ListActions() ([]string, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) applyFilter(query *gorm.DB, filter Filter) *gorm.DB {
	if filter.Module != "" {
		query = query.Where("module = ?", filter.Module)
	}
	if filter.Action != "" {
		query = query.Where("action = ?", filter.Action)
	}
	if filter.UserID != "" {
		query = query.Where("user_id = ?", filter.UserID)
	}
	if filter.Q != "" {
		like := "%" + filter.Q + "%"
		query = query.Where(
			"action LIKE ? OR module LIKE ? OR description LIKE ? OR ip_address LIKE ? OR COALESCE(entity_id, '') LIKE ?",
			like, like, like, like, like,
		)
	}
	return query
}

func (r *repository) List(page, perPage int, filter Filter) ([]domain.ActivityLog, int64, error) {
	query := r.applyFilter(r.db.Model(&domain.ActivityLog{}), filter)

	var total int64
	if err := query.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []domain.ActivityLog
	offset := (page - 1) * perPage
	err := query.Session(&gorm.Session{}).
		Preload("User").
		Order("id DESC").
		Offset(offset).
		Limit(perPage).
		Find(&items).Error
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *repository) FindByID(id uint) (*domain.ActivityLog, error) {
	var item domain.ActivityLog
	if err := r.db.Preload("User").First(&item, id).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *repository) Delete(id uint) error {
	return r.db.Delete(&domain.ActivityLog{}, id).Error
}

func (r *repository) ListModules() ([]string, error) {
	var modules []string
	err := r.db.Model(&domain.ActivityLog{}).
		Distinct("module").
		Where("module <> ''").
		Order("module ASC").
		Pluck("module", &modules).Error
	return modules, err
}

func (r *repository) ListActions() ([]string, error) {
	var actions []string
	err := r.db.Model(&domain.ActivityLog{}).
		Distinct("action").
		Where("action <> ''").
		Order("action ASC").
		Pluck("action", &actions).Error
	return actions, err
}
