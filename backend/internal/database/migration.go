package database

import (
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"go-cp-boilerplate/internal/domain"
)

// AutoMigrate migrates schema in an order that respects foreign key
// dependencies (roles/permissions before users).
// isProduction controls whether legacy integer user-id tables may be dropped
// for a fresh UUID cutover (never auto-dropped in production).
func AutoMigrate(db *gorm.DB, isProduction bool) error {
	if err := dropRemovedContentTables(db); err != nil {
		return err
	}
	if err := normalizeUserAccountColumns(db); err != nil {
		return err
	}
	if err := ensureUUIDUserSchema(db, isProduction); err != nil {
		return err
	}
	if err := db.AutoMigrate(
		&domain.Role{},
		&domain.Permission{},
		&domain.RolePermission{},
		&domain.User{},
		&domain.RefreshToken{},
		&domain.WebSetting{},
		&domain.MediaAsset{},
		&domain.ActivityLog{},
		&domain.Scope{},
		&domain.RoleAssignment{},
	); err != nil {
		return err
	}
	return migrateDropUserRoleID(db)
}

// migrateDropUserRoleID backfills users.role_id into role_assignments on the
// global scope, then drops the redundant column.
func migrateDropUserRoleID(db *gorm.DB) error {
	if !db.Migrator().HasTable("users") {
		return nil
	}
	hasColumn := db.Migrator().HasColumn(&domain.User{}, "role_id")
	// GORM may already have dropped the field from the model; check DB directly.
	if !hasColumn {
		dataType, err := mysqlColumnDataType(db, "users", "role_id")
		if err != nil {
			return err
		}
		if dataType == "" {
			return nil
		}
	}

	log.Info().Msg("migrating users.role_id into role_assignments on global scope")

	if err := ensureGlobalScopeRow(db); err != nil {
		return fmt.Errorf("ensure global scope: %w", err)
	}

	var global domain.Scope
	if err := db.Where("code = ?", domain.ScopeCodeGlobal).First(&global).Error; err != nil {
		return fmt.Errorf("load global scope: %w", err)
	}

	type legacyUser struct {
		ID     string
		RoleID uint
	}
	var legacy []legacyUser
	if err := db.Raw("SELECT id, role_id FROM users WHERE role_id IS NOT NULL AND role_id <> 0").Scan(&legacy).Error; err != nil {
		return fmt.Errorf("read legacy role_id: %w", err)
	}

	for _, u := range legacy {
		var count int64
		if err := db.Model(&domain.RoleAssignment{}).
			Where("user_id = ? AND role_id = ? AND scope_id = ?", u.ID, u.RoleID, global.ID).
			Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			continue
		}
		ra := domain.RoleAssignment{
			UserID:  u.ID,
			RoleID:  u.RoleID,
			ScopeID: global.ID,
		}
		if err := db.Create(&ra).Error; err != nil {
			return fmt.Errorf("backfill assignment for user %s: %w", u.ID, err)
		}
	}

	// Drop FK first — MySQL rejects DROP COLUMN while fk_users_role exists.
	_ = db.Exec("ALTER TABLE users DROP FOREIGN KEY fk_users_role").Error

	if err := db.Exec("ALTER TABLE users DROP COLUMN role_id").Error; err != nil {
		// Column may already be gone after a partial run.
		if !strings.Contains(strings.ToLower(err.Error()), "check that it exists") &&
			!strings.Contains(strings.ToLower(err.Error()), "unknown column") {
			return fmt.Errorf("drop users.role_id: %w", err)
		}
	}
	log.Info().Msg("dropped users.role_id")
	return nil
}

func ensureGlobalScopeRow(db *gorm.DB) error {
	var existing domain.Scope
	err := db.Where("code = ?", domain.ScopeCodeGlobal).First(&existing).Error
	if err == nil {
		return nil
	}
	if err != gorm.ErrRecordNotFound {
		return err
	}
	return db.Create(&domain.Scope{
		Code:        domain.ScopeCodeGlobal,
		Name:        "Global",
		Description: "Root organizational scope",
		IsActive:    true,
	}).Error
}

// dropRemovedContentTables removes company-profile content tables that are
// no longer part of the schema. GORM AutoMigrate never drops tables, so this
// must run explicitly when features are retired.
func dropRemovedContentTables(db *gorm.DB) error {
	tables := []string{
		"registrations",
		"registration_periods",
		"articles",
		"news",
		"gallery_items",
		"services",
		"contact_messages",
		"landing_contents",
		// Singular leftovers from earlier naming experiments.
		"gallery_item",
		"service",
		"contact_message",
		"landing_content",
		"registration",
	}
	for _, table := range tables {
		if !db.Migrator().HasTable(table) {
			continue
		}
		log.Info().Str("table", table).Msg("dropping removed content table")
		if err := db.Migrator().DropTable(table); err != nil {
			return fmt.Errorf("drop %s: %w", table, err)
		}
	}
	return nil
}

// ensureUUIDUserSchema drops tables still keyed by integer user ids so
// AutoMigrate can recreate them with char(36) UUIDs. Production must migrate
// fresh manually (drop DB / tables) — we refuse to auto-drop there.
func ensureUUIDUserSchema(db *gorm.DB, isProduction bool) error {
	if !db.Migrator().HasTable("users") {
		return nil
	}
	dataType, err := mysqlColumnDataType(db, "users", "id")
	if err != nil {
		return err
	}
	if dataType == "" || isCharUUIDType(dataType) {
		return nil
	}
	if isProduction {
		return fmt.Errorf("users.id is %s; drop and recreate the database for UUID cutover", dataType)
	}

	log.Warn().Str("users_id_type", dataType).Msg("dropping legacy integer user-id tables for UUID cutover")
	for _, table := range []string{"refresh_tokens", "activity_logs", "media_assets", "users"} {
		if !db.Migrator().HasTable(table) {
			continue
		}
		if err := db.Migrator().DropTable(table); err != nil {
			return fmt.Errorf("drop %s for UUID cutover: %w", table, err)
		}
	}
	return nil
}

func isCharUUIDType(dataType string) bool {
	switch strings.ToLower(dataType) {
	case "char", "varchar", "uuid", "binary":
		return true
	default:
		return false
	}
}

func mysqlColumnDataType(db *gorm.DB, table, column string) (string, error) {
	var dataType string
	err := db.Raw(`
		SELECT DATA_TYPE
		FROM information_schema.COLUMNS
		WHERE TABLE_SCHEMA = DATABASE()
		  AND TABLE_NAME = ?
		  AND COLUMN_NAME = ?
		LIMIT 1
	`, table, column).Scan(&dataType).Error
	return dataType, err
}

// normalizeUserAccountColumns backfills legacy rows so profile metadata
// (is_active, last_login_at) matches what the admin users table shows.
func normalizeUserAccountColumns(db *gorm.DB) error {
	if !db.Migrator().HasTable("users") {
		return nil
	}
	_ = db.Exec("UPDATE users SET is_active = 1 WHERE is_active IS NULL").Error
	return nil
}
