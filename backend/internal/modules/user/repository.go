package user

import (
	"gorm.io/gorm"

	"go-cp-boilerplate/internal/domain"
)

type Repository interface {
	List(page, perPage int, search string) ([]domain.User, int64, error)
	FindByID(id string) (*domain.User, error)
	FindByEmail(email string) (*domain.User, error)
	Create(u *domain.User) error
	Update(u *domain.User) error
	UpdatePassword(id, hashedPassword string) error
	RevokeAllRefreshTokens(userID string) error
	Delete(id string) error

	FindRoleByID(id uint) (*domain.Role, error)
	FindScopeByID(id uint) (*domain.Scope, error)
	ListAssignments(userID string) ([]domain.RoleAssignment, error)
	ReplaceAssignments(userID string, items []domain.RoleAssignment) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) List(page, perPage int, search string) ([]domain.User, int64, error) {
	query := r.db.Model(&domain.User{})
	if search != "" {
		like := "%" + search + "%"
		query = query.Where("name LIKE ? OR email LIKE ?", like, like)
	}

	var total int64
	if err := query.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var users []domain.User
	offset := (page - 1) * perPage
	if err := query.Session(&gorm.Session{}).Order("created_at DESC").Offset(offset).Limit(perPage).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (r *repository) FindByID(id string) (*domain.User, error) {
	var u domain.User
	if err := r.db.Where("id = ?", id).First(&u).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *repository) FindByEmail(email string) (*domain.User, error) {
	var u domain.User
	if err := r.db.Where("email = ?", email).First(&u).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *repository) Create(u *domain.User) error {
	return r.db.Create(u).Error
}

func (r *repository) Update(u *domain.User) error {
	return r.db.Save(u).Error
}

func (r *repository) UpdatePassword(id, hashedPassword string) error {
	return r.db.Model(&domain.User{}).Where("id = ?", id).Update("password", hashedPassword).Error
}

func (r *repository) RevokeAllRefreshTokens(userID string) error {
	return r.db.Model(&domain.RefreshToken{}).Where("user_id = ? AND revoked = false", userID).Update("revoked", true).Error
}

func (r *repository) Delete(id string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", id).Delete(&domain.RoleAssignment{}).Error; err != nil {
			return err
		}
		return tx.Where("id = ?", id).Delete(&domain.User{}).Error
	})
}

func (r *repository) FindRoleByID(id uint) (*domain.Role, error) {
	var role domain.Role
	if err := r.db.First(&role, id).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *repository) FindScopeByID(id uint) (*domain.Scope, error) {
	var scope domain.Scope
	if err := r.db.First(&scope, id).Error; err != nil {
		return nil, err
	}
	return &scope, nil
}

func (r *repository) ListAssignments(userID string) ([]domain.RoleAssignment, error) {
	var items []domain.RoleAssignment
	err := r.db.Preload("Role").Preload("Scope").Where("user_id = ?", userID).Find(&items).Error
	return items, err
}

func (r *repository) ReplaceAssignments(userID string, items []domain.RoleAssignment) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", userID).Delete(&domain.RoleAssignment{}).Error; err != nil {
			return err
		}
		if len(items) == 0 {
			return nil
		}
		return tx.Create(&items).Error
	})
}
