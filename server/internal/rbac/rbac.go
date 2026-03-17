package rbac

type Resource string
type Action string

const (
	ResourceSecret    Resource = "secret"
	ResourceProject   Resource = "project"
	ResourceAgent     Resource = "agent"
	ResourceScans     Resource = "scans"
	ResourceDynSecret Resource = "dynsecret"
)

const (
	ActionRead    Action = "read"
	ActionWrite   Action = "write"
	ActionAdmin   Action = "admin"
	ActionApprove Action = "approve"
)

// rolePerms defines what each role can do
var rolePerms = map[string]map[Resource][]Action{
	"owner": {
		ResourceSecret:    {ActionRead, ActionWrite, ActionAdmin, ActionApprove},
		ResourceProject:   {ActionRead, ActionWrite, ActionAdmin},
		ResourceAgent:     {ActionRead, ActionWrite, ActionAdmin},
		ResourceScans:     {ActionRead, ActionWrite},
		ResourceDynSecret: {ActionRead, ActionWrite},
	},
	"admin": {
		ResourceSecret:    {ActionRead, ActionWrite, ActionAdmin, ActionApprove},
		ResourceProject:   {ActionRead, ActionWrite, ActionAdmin},
		ResourceAgent:     {ActionRead, ActionWrite, ActionAdmin},
		ResourceScans:     {ActionRead, ActionWrite},
		ResourceDynSecret: {ActionRead, ActionWrite},
	},
	"member": {
		ResourceSecret:    {ActionRead, ActionWrite},
		ResourceProject:   {ActionRead},
		ResourceAgent:     {ActionRead, ActionWrite},
		ResourceScans:     {ActionRead, ActionWrite},
		ResourceDynSecret: {ActionRead},
	},
	"viewer": {
		ResourceSecret:    {ActionRead},
		ResourceProject:   {ActionRead},
		ResourceAgent:     {ActionRead},
		ResourceScans:     {ActionRead},
		ResourceDynSecret: {ActionRead},
	},
}

// Can checks if role has permission to perform action on resource.
func Can(role string, resource Resource, action Action) bool {
	perms, ok := rolePerms[role]
	if !ok {
		return false
	}
	actions, ok := perms[resource]
	if !ok {
		return false
	}
	for _, a := range actions {
		if a == action {
			return true
		}
	}
	return false
}

// RoleFromProjectMembership normalizes the role string.
// Returns "viewer" for unknown roles (safe default).
func RoleFromProjectMembership(role string) string {
	switch role {
	case "owner", "admin", "member", "viewer":
		return role
	default:
		return "viewer"
	}
}
