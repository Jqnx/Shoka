package auth

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

// frontendPath is relative to this package, which is where `go test` runs.
const frontendPath = "../../../frontend/src/lib/auth/permissions.ts"

func TestCan(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		role       string
		permission Permission
		want       bool
	}{
		{"admin updates archives", RoleAdmin, ArchiveUpdate, true},
		{"admin runs maintenance", RoleAdmin, MaintenanceRun, true},
		{"admin lists users", RoleAdmin, UserList, true},
		{"user cannot update archives", RoleUser, ArchiveUpdate, false},
		{"user cannot read libraries", RoleUser, LibraryRead, false},
		{"empty role behaves as user", "", ArchiveUpdate, false},
		{"unknown role behaves as user", "librarian", ArchiveUpdate, false},
		{"comma-separated role grants the union", "user,admin", ArchiveUpdate, true},
		{"comma-separated role tolerates spaces", " user , admin ", JobRead, true},
		{"comma-separated user-only role grants nothing", "user,librarian", JobRead, false},
		{"unknown resource is denied", RoleAdmin, Permission{"collection", ActionRead}, false},
		{"unknown action is denied", RoleAdmin, Permission{ResourceArchive, "publish"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := Can(tt.role, tt.permission); got != tt.want {
				t.Errorf("Can(%q, %v) = %v, want %v", tt.role, tt.permission, got, tt.want)
			}
		})
	}
}

// `user` exists only because better-auth validates the role map's keys; it
// must never grant anything.
func TestUserRoleGrantsNothing(t *testing.T) {
	t.Parallel()

	for _, permission := range allPermissions {
		if Can(RoleUser, permission) {
			t.Errorf("user should not be granted %s:%s", permission.Resource, permission.Action)
		}
	}
}

// Resources the frontend takes from `...defaultStatements` (better-auth's
// admin plugin) instead of writing out. The Go side owns those values because
// the parser below can't resolve the import.
var inheritedResources = map[string]bool{
	ResourceUser:    true,
	ResourceSession: true,
}

// Locks the Go matrix to the canonical definition in permissions.ts. Only the
// resources that file declares itself are compared; see inheritedResources.
func TestStatementMatrixMatchesFrontend(t *testing.T) {
	t.Parallel()

	declared := frontendStatements(t)

	for resource, actions := range declared {
		if !sameActions(actions, statements[resource]) {
			t.Errorf("permissions.ts declares %s: %v, Go declares %v", resource, actions, statements[resource])
		}
	}

	for resource := range statements {
		if inheritedResources[resource] {
			continue
		}

		if _, ok := declared[resource]; !ok {
			t.Errorf("Go declares resource %q that permissions.ts doesn't", resource)
		}
	}
}

func frontendStatements(t *testing.T) map[string][]string {
	t.Helper()

	data, err := os.ReadFile(frontendPath)
	if err != nil {
		t.Fatalf("read %s: %v", frontendPath, err)
	}

	source := string(data)

	body, ok := between(source, "export const statement = {", "} as const;")
	if !ok {
		t.Fatalf("%s: couldn't find the `statement` literal", frontendPath)
	}

	entry := regexp.MustCompile(`(?m)^\s*(\w+):\s*\[([^\]]*)\],?\s*$`)
	quoted := regexp.MustCompile(`'([^']*)'`)

	parsed := map[string][]string{}
	for _, match := range entry.FindAllStringSubmatch(body, -1) {
		actions := []string{}
		for _, action := range quoted.FindAllStringSubmatch(match[2], -1) {
			actions = append(actions, action[1])
		}

		parsed[match[1]] = actions
	}

	if len(parsed) == 0 {
		t.Fatalf("%s: parsed no statement entries", frontendPath)
	}

	return parsed
}

func between(s, open, close string) (string, bool) {
	_, rest, found := strings.Cut(s, open)
	if !found {
		return "", false
	}

	body, _, found := strings.Cut(rest, close)

	return body, found
}

func sameActions(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	set := make(map[string]bool, len(a))
	for _, action := range a {
		set[action] = true
	}

	for _, action := range b {
		if !set[action] {
			return false
		}
	}

	return true
}
