// Package auth carries the caller identity through requests and enforces the
// role/permission matrix used by the API handlers.
package auth

import (
	"slices"
	"strings"
)

// Role/permission matrix, mirroring the canonical better-auth access-control
// definition in frontend/src/lib/auth/permissions.ts. The frontend hides
// controls with it; this is what actually enforces them. Both statement sets
// are locked together by rbac_test.go, which reads permissions.ts directly
// instead of trusting a hand-maintained copy.
const (
	RoleUser  = "user"
	RoleAdmin = "admin"
)

// Resources.
const (
	ResourceArchive     = "archive"
	ResourceMetadata    = "metadata"
	ResourceTaxonomy    = "taxonomy"
	ResourceLibrary     = "library"
	ResourceJob         = "job"
	ResourceMaintenance = "maintenance"
	ResourceUser        = "user"
	ResourceSession     = "session"
)

// Actions. Reused across resources where the verb is the same.
const (
	ActionCreate      = "create"
	ActionRead        = "read"
	ActionUpdate      = "update"
	ActionDelete      = "delete"
	ActionCover       = "cover"
	ActionThumbnail   = "thumbnail"
	ActionFetch       = "fetch"
	ActionApply       = "apply"
	ActionWrite       = "write"
	ActionScan        = "scan"
	ActionRun         = "run"
	ActionList        = "list"
	ActionSetRole     = "set-role"
	ActionBan         = "ban"
	ActionRevoke      = "revoke"
	ActionImpersonate = "impersonate"
	ActionSetPassword = "set-password"
	ActionGet         = "get"
)

// Permission is a resource/action pair. Routes reference the declared values
// below rather than loose strings, so a typo is a compile error instead of a
// permission nobody holds - which would silently 403 the route for everyone,
// admins included.
type Permission struct{ Resource, Action string }

// Permissions the Go routes gate on.
var (
	ArchiveUpdate    = Permission{ResourceArchive, ActionUpdate}
	ArchiveDelete    = Permission{ResourceArchive, ActionDelete}
	ArchiveCover     = Permission{ResourceArchive, ActionCover}
	ArchiveThumbnail = Permission{ResourceArchive, ActionThumbnail}

	MetadataFetch = Permission{ResourceMetadata, ActionFetch}
	MetadataApply = Permission{ResourceMetadata, ActionApply}
	MetadataWrite = Permission{ResourceMetadata, ActionWrite}

	TaxonomyCreate = Permission{ResourceTaxonomy, ActionCreate}
	TaxonomyUpdate = Permission{ResourceTaxonomy, ActionUpdate}
	TaxonomyDelete = Permission{ResourceTaxonomy, ActionDelete}

	LibraryRead   = Permission{ResourceLibrary, ActionRead}
	LibraryCreate = Permission{ResourceLibrary, ActionCreate}
	LibraryUpdate = Permission{ResourceLibrary, ActionUpdate}
	LibraryDelete = Permission{ResourceLibrary, ActionDelete}
	LibraryScan   = Permission{ResourceLibrary, ActionScan}

	JobRead = Permission{ResourceJob, ActionRead}

	MaintenanceRead = Permission{ResourceMaintenance, ActionRead}
	MaintenanceRun  = Permission{ResourceMaintenance, ActionRun}
)

// better-auth's own admin statements. Only the frontend serves these routes,
// but they belong to the same matrix.
var (
	UserCreate      = Permission{ResourceUser, ActionCreate}
	UserList        = Permission{ResourceUser, ActionList}
	UserSetRole     = Permission{ResourceUser, ActionSetRole}
	UserBan         = Permission{ResourceUser, ActionBan}
	UserImpersonate = Permission{ResourceUser, ActionImpersonate}
	UserDelete      = Permission{ResourceUser, ActionDelete}
	UserSetPassword = Permission{ResourceUser, ActionSetPassword}
	UserGet         = Permission{ResourceUser, ActionGet}
	UserUpdate      = Permission{ResourceUser, ActionUpdate}

	SessionList   = Permission{ResourceSession, ActionList}
	SessionRevoke = Permission{ResourceSession, ActionRevoke}
	SessionDelete = Permission{ResourceSession, ActionDelete}
)

// allPermissions is every permission the system knows about, independent of
// who holds it.
var allPermissions = []Permission{
	ArchiveUpdate,
	ArchiveDelete,
	ArchiveCover,
	ArchiveThumbnail,
	MetadataFetch,
	MetadataApply,
	MetadataWrite,
	TaxonomyCreate,
	TaxonomyUpdate,
	TaxonomyDelete,
	LibraryRead,
	LibraryCreate,
	LibraryUpdate,
	LibraryDelete,
	LibraryScan,
	JobRead,
	MaintenanceRead,
	MaintenanceRun,
	UserCreate,
	UserList,
	UserSetRole,
	UserBan,
	UserImpersonate,
	UserDelete,
	UserSetPassword,
	UserGet,
	UserUpdate,
	SessionList,
	SessionRevoke,
	SessionDelete,
}

// statements is every permission, grouped by resource. Derived from
// allPermissions so the two can never disagree.
var statements = permissionsByResource(allPermissions)

func permissionsByResource(permissions []Permission) map[string][]string {
	byResource := make(map[string][]string, len(permissions))

	for _, p := range permissions {
		byResource[p.Resource] = append(byResource[p.Resource], p.Action)
	}

	return byResource
}

// roles maps a role to the actions it holds per resource. `user` grants
// nothing - reading the library is the baseline of being authenticated, and
// per-user data is scoped by user id rather than by role - but it must be
// present: better-auth validates setRole/createUser against the keys of the
// map passed to the admin plugin.
var roles = map[string]map[string][]string{
	RoleAdmin: statements,
	RoleUser:  {},
}

// IsAdmin reports whether role contains the admin role. A comma-separated
// role string is split the same way Can splits it, so a session that satisfies
// requirePermission as an admin also bypasses the library access list.
func IsAdmin(role string) bool {
	for name := range strings.SplitSeq(role, ",") {
		if strings.TrimSpace(name) == RoleAdmin {
			return true
		}
	}

	return false
}

// Can reports whether role may perform permission. An empty, NULL-sourced, or
// unrecognised role is treated as `user` and grants nothing. A comma-separated
// role string grants the union of its parts, matching better-auth's own
// hasPermission.
func Can(role string, permission Permission) bool {
	for name := range strings.SplitSeq(role, ",") {
		granted, ok := roles[strings.TrimSpace(name)]
		if !ok {
			continue
		}

		if slices.Contains(granted[permission.Resource], permission.Action) {
			return true
		}
	}

	return false
}
