package constants

// Permission keys, format "<resource>.<action>". Matched against the
// permission list embedded in the JWT / looked up via RBAC middleware.
const (
	PermUserView        = "user.view"
	PermUserCreate      = "user.create"
	PermUserUpdate      = "user.update"
	PermUserDelete      = "user.delete"
	PermUserImpersonate = "user.impersonate"

	PermRoleView   = "role.view"
	PermRoleCreate = "role.create"
	PermRoleUpdate = "role.update"
	PermRoleDelete = "role.delete"

	PermScopeView   = "scope.view"
	PermScopeCreate = "scope.create"
	PermScopeUpdate = "scope.update"
	PermScopeDelete = "scope.delete"

	PermSettingView   = "setting.view"
	PermSettingUpdate = "setting.update"

	PermMediaUpload = "media.upload"
	PermMediaDelete = "media.delete"

	PermActivityView   = "activity.view"
	PermActivityDelete = "activity.delete"

	PermInvoiceView   = "invoice.view"
	PermInvoiceCreate = "invoice.create"
	PermInvoiceUpdate = "invoice.update"
	PermInvoiceDelete = "invoice.delete"
)

// AllPermissions lists every permission seeded on first run.
var AllPermissions = []string{
	PermUserView, PermUserCreate, PermUserUpdate, PermUserDelete, PermUserImpersonate,
	PermRoleView, PermRoleCreate, PermRoleUpdate, PermRoleDelete,
	PermScopeView, PermScopeCreate, PermScopeUpdate, PermScopeDelete,
	PermSettingView, PermSettingUpdate,
	PermMediaUpload, PermMediaDelete,
	PermActivityView, PermActivityDelete,
	PermInvoiceView, PermInvoiceCreate, PermInvoiceUpdate, PermInvoiceDelete,
}

// RolePermissionMap defines which permissions each default role gets on seed.
var RolePermissionMap = map[string][]string{
	"SuperAdmin": AllPermissions,
	"Admin": {
		PermUserView, PermUserCreate, PermUserUpdate, PermUserDelete,
		PermRoleView, PermRoleCreate, PermRoleUpdate, PermRoleDelete,
		PermScopeView, PermScopeCreate, PermScopeUpdate, PermScopeDelete,
		PermSettingView, PermSettingUpdate,
		PermMediaUpload, PermMediaDelete,
		PermActivityView, PermActivityDelete,
		PermInvoiceView, PermInvoiceCreate, PermInvoiceUpdate, PermInvoiceDelete,
	},
	"Editor": {
		PermMediaUpload,
		PermSettingView,
	},
	"Operator": {
		PermMediaUpload,
		PermInvoiceView, PermInvoiceCreate, PermInvoiceUpdate,
	},
}
