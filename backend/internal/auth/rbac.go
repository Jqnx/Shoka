package auth

import (
	"slices"
	"strings"
)

// Role/permission matrix, mirroring the canonical better-auth access-control
// definition in frontend/src/lib/auth/permissions.ts. The frontend hides
// controls with it; this is what actually enforces them. Keep both in sync —
// rbac_test.go locks the statement list down.
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
	ActionCreate    = "create"
	ActionRead      = "read"
	ActionUpdate    = "update"
	ActionDelete    = "delete"
	ActionCover     = "cover"
	ActionThumbnail = "thumbnail"
	ActionFetch     = "fetch"
	ActionApply     = "apply"
	ActionWrite     = "write"
	ActionScan      = "scan"
	ActionRun       = "run"
	ActionList      = "list"
	ActionSetRole   = "set-role"
	ActionBan       = "ban"
	ActionRevoke    = "revoke"
)

// statements is every permission the system knows about, independent of who
// holds it.
var statements = map[string][]string{
	ResourceArchive:     {ActionUpdate, ActionDelete, ActionCover, ActionThumbnail},
	ResourceMetadata:    {ActionFetch, ActionApply, ActionWrite},
	ResourceTaxonomy:    {ActionCreate, ActionUpdate, ActionDelete},
	ResourceLibrary:     {ActionRead, ActionCreate, ActionUpdate, ActionDelete, ActionScan},
	ResourceJob:         {ActionRead},
	ResourceMaintenance: {ActionRead, ActionRun},
	// better-auth's own admin statements. Only the frontend serves these
	// routes, but they belong to the same matrix.
	ResourceUser:    {ActionCreate, ActionList, ActionSetRole, ActionBan, "impersonate", ActionDelete, "set-password", "get", ActionUpdate},
	ResourceSession: {ActionList, ActionRevoke, ActionDelete},
}

// roles maps a role to the actions it holds per resource. `user` is absent on
// purpose: reading the library is the baseline of being authenticated, and
// per-user data is scoped by user id rather than by role, so it grants nothing.
var roles = map[string]map[string][]string{
	RoleAdmin: statements,
}

// Can reports whether role may perform action on resource. An empty, NULL-
// sourced, or unrecognised role is treated as `user` and grants nothing. A
// comma-separated role string grants the union of its parts, matching
// better-auth's own hasPermission.
func Can(role, resource, action string) bool {
	for name := range strings.SplitSeq(role, ",") {
		granted, ok := roles[strings.TrimSpace(name)]
		if !ok {
			continue
		}

		if slices.Contains(granted[resource], action) {
			return true
		}
	}

	return false
}
