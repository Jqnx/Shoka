package auth

import (
	"slices"
	"testing"
)

// routePermissions is every permission referenced by a route in
// internal/api/routes.go. If a route gains a permission that isn't declared in
// the statement set, this fails rather than silently denying everyone.
var routePermissions = []struct{ resource, action string }{
	{ResourceArchive, ActionUpdate},
	{ResourceArchive, ActionDelete},
	{ResourceArchive, ActionCover},
	{ResourceArchive, ActionThumbnail},
	{ResourceMetadata, ActionFetch},
	{ResourceMetadata, ActionApply},
	{ResourceMetadata, ActionWrite},
	{ResourceTaxonomy, ActionCreate},
	{ResourceTaxonomy, ActionUpdate},
	{ResourceTaxonomy, ActionDelete},
	{ResourceLibrary, ActionRead},
	{ResourceLibrary, ActionCreate},
	{ResourceLibrary, ActionUpdate},
	{ResourceLibrary, ActionDelete},
	{ResourceLibrary, ActionScan},
	{ResourceJob, ActionRead},
	{ResourceMaintenance, ActionRead},
	{ResourceMaintenance, ActionRun},
}

func TestCan(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		role     string
		resource string
		action   string
		want     bool
	}{
		{"admin updates archives", RoleAdmin, ResourceArchive, ActionUpdate, true},
		{"admin runs maintenance", RoleAdmin, ResourceMaintenance, ActionRun, true},
		{"admin lists users", RoleAdmin, ResourceUser, ActionList, true},
		{"user cannot update archives", RoleUser, ResourceArchive, ActionUpdate, false},
		{"user cannot read libraries", RoleUser, ResourceLibrary, ActionRead, false},
		{"empty role behaves as user", "", ResourceArchive, ActionUpdate, false},
		{"unknown role behaves as user", "librarian", ResourceArchive, ActionUpdate, false},
		{"comma-separated role grants the union", "user,admin", ResourceArchive, ActionUpdate, true},
		{"comma-separated role tolerates spaces", " user , admin ", ResourceJob, ActionRead, true},
		{"comma-separated user-only role grants nothing", "user,librarian", ResourceJob, ActionRead, false},
		{"unknown resource is denied", RoleAdmin, "collection", ActionRead, false},
		{"unknown action is denied", RoleAdmin, ResourceArchive, "publish", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := Can(tt.role, tt.resource, tt.action); got != tt.want {
				t.Errorf("Can(%q, %q, %q) = %v, want %v", tt.role, tt.resource, tt.action, got, tt.want)
			}
		})
	}
}

func TestCanAdminGrantsEveryStatement(t *testing.T) {
	t.Parallel()

	for resource, actions := range statements {
		for _, action := range actions {
			if !Can(RoleAdmin, resource, action) {
				t.Errorf("admin should be granted %s:%s", resource, action)
			}

			if Can(RoleUser, resource, action) {
				t.Errorf("user should not be granted %s:%s", resource, action)
			}
		}
	}
}

// Anti-drift lock against the TypeScript matrix in
// frontend/src/lib/auth/permissions.ts.
func TestRoutePermissionsAreDeclared(t *testing.T) {
	t.Parallel()

	for _, p := range routePermissions {
		if !slices.Contains(statements[p.resource], p.action) {
			t.Errorf("route permission %s:%s is not in the statement set", p.resource, p.action)
		}
	}
}

func TestEveryStatementIsGranted(t *testing.T) {
	t.Parallel()

	for resource, actions := range statements {
		for _, action := range actions {
			granted := false
			for role := range roles {
				if Can(role, resource, action) {
					granted = true
					break
				}
			}

			if !granted {
				t.Errorf("no role grants %s:%s", resource, action)
			}
		}
	}
}
