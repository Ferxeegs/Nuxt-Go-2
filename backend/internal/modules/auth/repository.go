package auth

import (
	"time"

	"gorm.io/gorm"

	"go-cp-boilerplate/internal/domain"
)

type Repository interface {
	FindUserByEmail(email string) (*domain.User, error)
	FindUserByID(id string) (*domain.User, error)
	FindPermissionsByRoleID(roleID uint) ([]string, error)
	FindUnionPermissions(userID string) ([]string, error)
	FindAssignmentsByUserID(userID string) ([]domain.RoleAssignment, error)
	FindAssignment(userID string, roleID, scopeID uint) (*domain.RoleAssignment, error)
	UpdateUserFields(userID string, fields map[string]interface{}) error
	EmailTakenByOther(email string, userID string) (bool, error)
	UpdateLastLogin(userID string, at time.Time) error

	CreateRefreshToken(rt *domain.RefreshToken) error
	FindRefreshTokenByHash(hash string) (*domain.RefreshToken, error)
	RevokeRefreshToken(id uint) error
	RevokeAllUserRefreshTokens(userID string) error

	CreateActivityLog(log *domain.ActivityLog) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) FindUserByEmail(email string) (*domain.User, error) {
	var user domain.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *repository) FindUserByID(id string) (*domain.User, error) {
	var user domain.User
	if err := r.db.Where("id = ?", id).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *repository) FindPermissionsByRoleID(roleID uint) ([]string, error) {
	var names []string
	err := r.db.Table("permissions").
		Joins("JOIN role_permissions ON role_permissions.permission_id = permissions.id").
		Where("role_permissions.role_id = ?", roleID).
		Pluck("permissions.name", &names).Error
	return names, err
}

func (r *repository) FindUnionPermissions(userID string) ([]string, error) {
	var names []string
	err := r.db.Table("permissions").
		Joins("JOIN role_permissions ON role_permissions.permission_id = permissions.id").
		Joins("JOIN role_assignments ON role_assignments.role_id = role_permissions.role_id").
		Where("role_assignments.user_id = ?", userID).
		Pluck("permissions.name", &names).Error
	if err != nil {
		return nil, err
	}

	seen := make(map[string]struct{}, len(names))
	out := make([]string, 0, len(names))
	for _, name := range names {
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		out = append(out, name)
	}
	return out, nil
}

func (r *repository) FindAssignmentsByUserID(userID string) ([]domain.RoleAssignment, error) {
	var items []domain.RoleAssignment
	err := r.db.Preload("Role").Preload("Scope").Where("user_id = ?", userID).Find(&items).Error
	return items, err
}

func (r *repository) FindAssignment(userID string, roleID, scopeID uint) (*domain.RoleAssignment, error) {
	var item domain.RoleAssignment
	err := r.db.Preload("Role").Preload("Scope").
		Where("user_id = ? AND role_id = ? AND scope_id = ?", userID, roleID, scopeID).
		First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *repository) UpdateLastLogin(userID string, at time.Time) error {
	return r.db.Model(&domain.User{}).Where("id = ?", userID).Update("last_login_at", at).Error
}

func (r *repository) UpdateUserFields(userID string, fields map[string]interface{}) error {
	if len(fields) == 0 {
		return nil
	}
	return r.db.Model(&domain.User{}).Where("id = ?", userID).Updates(fields).Error
}

func (r *repository) EmailTakenByOther(email string, userID string) (bool, error) {
	var count int64
	err := r.db.Model(&domain.User{}).Where("email = ? AND id <> ?", email, userID).Count(&count).Error
	return count > 0, err
}

func (r *repository) CreateRefreshToken(rt *domain.RefreshToken) error {
	return r.db.Create(rt).Error
}

func (r *repository) FindRefreshTokenByHash(hash string) (*domain.RefreshToken, error) {
	var rt domain.RefreshToken
	if err := r.db.Where("token_hash = ?", hash).First(&rt).Error; err != nil {
		return nil, err
	}
	return &rt, nil
}

func (r *repository) RevokeRefreshToken(id uint) error {
	return r.db.Model(&domain.RefreshToken{}).Where("id = ?", id).Update("revoked", true).Error
}

func (r *repository) RevokeAllUserRefreshTokens(userID string) error {
	return r.db.Model(&domain.RefreshToken{}).Where("user_id = ? AND revoked = false", userID).Update("revoked", true).Error
}

func (r *repository) CreateActivityLog(log *domain.ActivityLog) error {
	return r.db.Create(log).Error
}
