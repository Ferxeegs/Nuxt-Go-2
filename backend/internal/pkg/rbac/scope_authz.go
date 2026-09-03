package rbac

import (
	"gorm.io/gorm"

	"go-cp-boilerplate/internal/domain"
)

// HasPermissionInScope reports whether the user may exercise perm inside scopeID.
// SuperAdmin (any assignment role named SuperAdmin) always passes.
// Otherwise the permission must come from a role assigned on the global scope
// or on the target scope / its ancestors.
func HasPermissionInScope(db *gorm.DB, userID string, perm string, scopeID uint) (bool, error) {
	var superCount int64
	if err := db.Table("role_assignments").
		Joins("JOIN roles ON roles.id = role_assignments.role_id").
		Where("role_assignments.user_id = ? AND roles.name = ?", userID, domain.RoleSuperAdmin).
		Count(&superCount).Error; err != nil {
		return false, err
	}
	if superCount > 0 {
		return true, nil
	}

	var globalID uint
	if err := db.Model(&domain.Scope{}).Select("id").Where("code = ?", domain.ScopeCodeGlobal).Scan(&globalID).Error; err != nil {
		return false, err
	}

	scopeIDs, err := ancestorScopeIDs(db, scopeID)
	if err != nil {
		return false, err
	}
	if globalID != 0 {
		scopeIDs = append(scopeIDs, globalID)
	}

	var has int64
	err = db.Table("role_assignments").
		Joins("JOIN role_permissions ON role_permissions.role_id = role_assignments.role_id").
		Joins("JOIN permissions ON permissions.id = role_permissions.permission_id").
		Where("role_assignments.user_id = ? AND role_assignments.scope_id IN ? AND permissions.name = ?", userID, scopeIDs, perm).
		Count(&has).Error
	return has > 0, err
}

func ancestorScopeIDs(db *gorm.DB, scopeID uint) ([]uint, error) {
	ids := []uint{scopeID}
	current := scopeID
	seen := map[uint]struct{}{scopeID: {}}
	for i := 0; i < 32; i++ {
		var parentID *uint
		if err := db.Model(&domain.Scope{}).Select("parent_id").Where("id = ?", current).Scan(&parentID).Error; err != nil {
			return nil, err
		}
		if parentID == nil || *parentID == 0 {
			break
		}
		if _, ok := seen[*parentID]; ok {
			break
		}
		seen[*parentID] = struct{}{}
		ids = append(ids, *parentID)
		current = *parentID
	}
	return ids, nil
}
