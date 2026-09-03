package scope

import (
	"gorm.io/gorm"

	"go-cp-boilerplate/internal/domain"
)

type Repository interface {
	List(page, perPage int, search string) ([]domain.Scope, int64, error)
	FindByID(id uint) (*domain.Scope, error)
	FindByCode(code string) (*domain.Scope, error)
	Create(m *domain.Scope) error
	Update(m *domain.Scope) error
	Delete(id uint) error
	CountChildren(id uint) (int64, error)
	CountAssignments(id uint) (int64, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) List(page, perPage int, search string) ([]domain.Scope, int64, error) {
	query := r.db.Model(&domain.Scope{}).Preload("Parent")
	if search != "" {
		like := "%" + search + "%"
		query = query.Where("code LIKE ? OR name LIKE ?", like, like)
	}

	var total int64
	if err := query.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []domain.Scope
	offset := (page - 1) * perPage
	if err := query.Session(&gorm.Session{}).Order("code ASC").Offset(offset).Limit(perPage).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *repository) FindByID(id uint) (*domain.Scope, error) {
	var m domain.Scope
	if err := r.db.Preload("Parent").First(&m, id).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *repository) FindByCode(code string) (*domain.Scope, error) {
	var m domain.Scope
	if err := r.db.Where("code = ?", code).First(&m).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *repository) Create(m *domain.Scope) error {
	return r.db.Create(m).Error
}

func (r *repository) Update(m *domain.Scope) error {
	return r.db.Save(m).Error
}

func (r *repository) Delete(id uint) error {
	return r.db.Delete(&domain.Scope{}, id).Error
}

func (r *repository) CountChildren(id uint) (int64, error) {
	var count int64
	err := r.db.Model(&domain.Scope{}).Where("parent_id = ?", id).Count(&count).Error
	return count, err
}

func (r *repository) CountAssignments(id uint) (int64, error) {
	var count int64
	err := r.db.Model(&domain.RoleAssignment{}).Where("scope_id = ?", id).Count(&count).Error
	return count, err
}
