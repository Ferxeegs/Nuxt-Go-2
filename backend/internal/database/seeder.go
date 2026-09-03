package database

import (
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"go-cp-boilerplate/internal/config"
	"go-cp-boilerplate/internal/constants"
	"go-cp-boilerplate/internal/domain"
	"go-cp-boilerplate/internal/pkg/hash"
)

// Seed is idempotent (checks existence before insert). Controlled by cfg.DBSeed
// — disable in production via DB_SEED=false in .env.
func Seed(db *gorm.DB, cfg *config.Config) error {
	if !cfg.DBSeed {
		log.Info().Msg("database seed skipped (DB_SEED=false)")
		return nil
	}

	roles, err := seedRoles(db)
	if err != nil {
		return err
	}

	permissions, err := seedPermissions(db)
	if err != nil {
		return err
	}

	if err := seedRolePermissions(db, roles, permissions); err != nil {
		return err
	}

	if err := seedGlobalScope(db); err != nil {
		return err
	}

	if err := seedSuperAdmin(db, cfg, roles[domain.RoleSuperAdmin]); err != nil {
		return err
	}

	if err := seedWebSettings(db); err != nil {
		return err
	}

	log.Info().Msg("database seed completed")
	return nil
}

func seedRoles(db *gorm.DB) (map[string]domain.Role, error) {
	names := []struct {
		name string
		desc string
	}{
		{domain.RoleSuperAdmin, "Full access to all resources"},
		{domain.RoleAdmin, "Manage users, settings, and media"},
		{domain.RoleEditor, "Upload media and view settings"},
		{domain.RoleOperator, "Day-to-day operational access (media and invoices)"},
	}

	result := map[string]domain.Role{}
	for _, r := range names {
		var role domain.Role
		err := db.Where("name = ?", r.name).First(&role).Error
		if err == gorm.ErrRecordNotFound {
			role = domain.Role{Name: r.name, Description: r.desc}
			if err := db.Create(&role).Error; err != nil {
				return nil, err
			}
		} else if err != nil {
			return nil, err
		}
		result[r.name] = role
	}
	return result, nil
}

func seedPermissions(db *gorm.DB) (map[string]domain.Permission, error) {
	result := map[string]domain.Permission{}
	for _, name := range constants.AllPermissions {
		var perm domain.Permission
		err := db.Where("name = ?", name).First(&perm).Error
		if err == gorm.ErrRecordNotFound {
			perm = domain.Permission{Name: name}
			if err := db.Create(&perm).Error; err != nil {
				return nil, err
			}
		} else if err != nil {
			return nil, err
		}
		result[name] = perm
	}
	return result, nil
}

func seedRolePermissions(db *gorm.DB, roles map[string]domain.Role, permissions map[string]domain.Permission) error {
	for roleName, permNames := range constants.RolePermissionMap {
		role, ok := roles[roleName]
		if !ok {
			continue
		}
		for _, permName := range permNames {
			perm, ok := permissions[permName]
			if !ok {
				continue
			}
			var existing domain.RolePermission
			err := db.Where("role_id = ? AND permission_id = ?", role.ID, perm.ID).First(&existing).Error
			if err == gorm.ErrRecordNotFound {
				rp := domain.RolePermission{RoleID: role.ID, PermissionID: perm.ID}
				if err := db.Create(&rp).Error; err != nil {
					return err
				}
			} else if err != nil {
				return err
			}
		}
	}
	return nil
}

func seedSuperAdmin(db *gorm.DB, cfg *config.Config, superAdminRole domain.Role) error {
	var global domain.Scope
	if err := db.Where("code = ?", domain.ScopeCodeGlobal).First(&global).Error; err != nil {
		return err
	}

	var existing domain.User
	err := db.Where("email = ?", cfg.SuperAdminEmail).First(&existing).Error
	if err == nil {
		return ensureSuperAdminAssignment(db, existing.ID, superAdminRole.ID, global.ID)
	}
	if err != gorm.ErrRecordNotFound {
		return err
	}

	hashed, err := hash.Hash(cfg.SuperAdminPassword, cfg.BcryptCost)
	if err != nil {
		return err
	}

	admin := domain.User{
		Name:     "Super Admin",
		Email:    cfg.SuperAdminEmail,
		Password: hashed,
		IsActive: true,
	}
	if err := db.Create(&admin).Error; err != nil {
		return err
	}

	if err := ensureSuperAdminAssignment(db, admin.ID, superAdminRole.ID, global.ID); err != nil {
		return err
	}

	log.Warn().Str("email", cfg.SuperAdminEmail).Msg("seeded default SuperAdmin account - change the password after first login")
	return nil
}

func ensureSuperAdminAssignment(db *gorm.DB, userID string, roleID, scopeID uint) error {
	var count int64
	if err := db.Model(&domain.RoleAssignment{}).
		Where("user_id = ? AND role_id = ? AND scope_id = ?", userID, roleID, scopeID).
		Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	return db.Create(&domain.RoleAssignment{
		UserID:  userID,
		RoleID:  roleID,
		ScopeID: scopeID,
	}).Error
}

func seedWebSettings(db *gorm.DB) error {
	obsoleteKeys := []string{
		"psb_open",
		"hero_title",
		"hero_subtitle",
		"hero_video",
		"about_title",
		"about_body",
		"about_visi",
		"about_misi",
		"cta_title",
		"cta_subtitle",
	}
	if err := db.Where("`key` IN ?", obsoleteKeys).Delete(&domain.WebSetting{}).Error; err != nil {
		return err
	}

	defaults := []domain.WebSetting{
		{Key: "site_name", Value: "My Company Profile", Type: "string"},
		{Key: "logo", Value: "", Type: "image"},
		{Key: "contact_email", Value: "info@example.com", Type: "string"},
		{Key: "contact_phone", Value: "", Type: "string"},
		{Key: "social_facebook", Value: "", Type: "string"},
		{Key: "footer_text", Value: "© 2026 My Company", Type: "string"},
	}
	for _, s := range defaults {
		var existing domain.WebSetting
		err := db.Where("`key` = ?", s.Key).First(&existing).Error
		if err == gorm.ErrRecordNotFound {
			if err := db.Create(&s).Error; err != nil {
				return err
			}
		} else if err != nil {
			return err
		}
	}
	return nil
}

func seedGlobalScope(db *gorm.DB) error {
	return ensureGlobalScopeRow(db)
}
