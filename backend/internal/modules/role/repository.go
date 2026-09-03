package role

import (
	"gorm.io/gorm"

	"go-cp-boilerplate/internal/domain"
)

type Repository interface {
	List() ([]domain.Role, error)
	FindByID(id uint) (*domain.Role, error)
	Create(r *domain.Role) error
	Update(r *domain.Role) error
	Delete(id uint) error
	ReplacePermissions(roleID uint, permissionIDs []uint) error
	AllPermissions() ([]domain.Permission, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) List() ([]domain.Role, error) {
	var roles []domain.Role
	err := r.db.Preload("Permissions").Order("id ASC").Find(&roles).Error
	return roles, err
}

func (r *repository) FindByID(id uint) (*domain.Role, error) {
	var role domain.Role
	if err := r.db.Preload("Permissions").First(&role, id).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *repository) Create(role *domain.Role) error {
	return r.db.Create(role).Error
}

func (r *repository) Update(role *domain.Role) error {
	return r.db.Save(role).Error
}

func (r *repository) Delete(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("role_id = ?", id).Delete(&domain.RolePermission{}).Error; err != nil {
			return err
		}
		return tx.Delete(&domain.Role{}, id).Error
	})
}

// ReplacePermissions overwrites a role's permission set atomically.
func (r *repository) ReplacePermissions(roleID uint, permissionIDs []uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("role_id = ?", roleID).Delete(&domain.RolePermission{}).Error; err != nil {
			return err
		}
		for _, pid := range permissionIDs {
			rp := domain.RolePermission{RoleID: roleID, PermissionID: pid}
			if err := tx.Create(&rp).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *repository) AllPermissions() ([]domain.Permission, error) {
	var perms []domain.Permission
	err := r.db.Order("name ASC").Find(&perms).Error
	return perms, err
}
