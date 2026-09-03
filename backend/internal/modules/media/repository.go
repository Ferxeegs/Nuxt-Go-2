package media

import (
	"gorm.io/gorm"

	"go-cp-boilerplate/internal/domain"
)

type Repository interface {
	Create(m *domain.MediaAsset) error
	FindByID(id uint) (*domain.MediaAsset, error)
	Delete(id uint) error
	List(page, perPage int, module string) ([]domain.MediaAsset, int64, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(m *domain.MediaAsset) error {
	return r.db.Create(m).Error
}

func (r *repository) FindByID(id uint) (*domain.MediaAsset, error) {
	var m domain.MediaAsset
	if err := r.db.First(&m, id).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *repository) Delete(id uint) error {
	return r.db.Delete(&domain.MediaAsset{}, id).Error
}

func (r *repository) List(page, perPage int, module string) ([]domain.MediaAsset, int64, error) {
	query := r.db.Model(&domain.MediaAsset{})
	if module != "" {
		query = query.Where("related_module = ?", module)
	}

	var total int64
	if err := query.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []domain.MediaAsset
	offset := (page - 1) * perPage
	if err := query.Session(&gorm.Session{}).Order("id DESC").Offset(offset).Limit(perPage).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}
