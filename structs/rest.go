package structs

import "golang.org/x/xerrors"

// ActionType signifies the action type of a task.
type ActionType uint8

const (
	// ActionStar signifies the issue will be starred/un-starred.
	ActionStar ActionType = iota
	// ActionAssign signifies that someone is being assigned/unassigned.
	ActionAssign
	// ActionLockComments signifies the comments for an issue is about
	// to be locked/unlocked.
	ActionLockComments
	// ActionMarkStatus signifies the status of an issue is changing.
	ActionMarkStatus
)

// ParseActionType converts a response string into a ActionType value.
// Returns an error if the input string does not match known values.
func ParseActionType(actionTypeStr string) (ActionType, error) {
	switch actionTypeStr {
	case "star":
		return ActionStar, nil
	case "assign":
		return ActionAssign, nil
	case "lock_comments":
		return ActionLockComments, nil
	case "mark_status":
		return ActionMarkStatus, nil
	}

	return ActionStar, xerrors.Errorf("Unknown EntryType String: '%s', defaulting to EntryActive", actionTypeStr)
}

// BaseResponse is the structure of all REST requests.
type BaseResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// APIMe is the structure of the /api/me endpoint.
type APIMe struct {
	Authenticated bool             `json:"authenticated"`
	User          *User            `json:"user"`
	Projects      []PartialProject `json:"projects"`
}

// PartialProject is a partial version of a project object.
type PartialProject struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`

	Archived bool `json:"archived"`
	Private  bool `json:"private"`

	OpenIssues     int `json:"open_issues"`
	ActiveIssues   int `json:"active_issues"`
	ResolvedIssues int `json:"resolved_issues"`
}

// PartialUser is the partial version of a user object.
type PartialUser struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Avatar      string `json:"avatar,omitempty"`
	Integration bool   `json:"integration"`
}

// APIProject is the structure of the GET /api/project/{id} endpoint.
type APIProject struct {
	Project      *Project              `json:"project"`
	Elevated     bool                  `json:"elevated"`
	Contributors map[int64]PartialUser `json:"contributors"`
}

// APIProjectLazy is the structure of the GET /api/project/{id}/lazy endpoint and /api/project/{id}/contributors.
type APIProjectLazy struct {
	Users map[int64]PartialUser `json:"users"`
	IDs   []int64               `json:"ids,omitempty"`
}

// APIProjectExecutor is the structure of the POST /api/project/{id}/execute endpoint.
type APIProjectExecutor struct {
	Issues      []IssueEntry `json:"issues"`
	Unavailable []int64      `json:"unavailable"`
	Project     *Project     `json:"project"`
}

// APIProjectIssues is the structure of the GET /api/project/{id}/issues endpoint.
type APIProjectIssues struct {
	Page        int          `json:"page"`
	TotalIssues int          `json:"total_issues"`
	Issues      []IssueEntry `json:"issues,omitempty"`
	Issue       *IssueEntry  `json:"issue,omitempty"`
}

// APIProjectIssueCreate is the structure of the POST /api/project/{id}/issues endpoint.
type APIProjectIssueCreate struct {
	New   bool        `json:"new"`
	Issue *IssueEntry `json:"issue"`
}

// APIProjectIssueComments is the structure of the GET /api/project/{id}/issues/{issue_id}/comments endpoint.
type APIProjectIssueComments struct {
	Page     int       `json:"page"`
	Comments []Comment `json:"comments"`
	End      bool      `json:"end"`
}

// APIProjectUpdate is the structure of the POST /api/project/{id}.
type APIProjectUpdate struct {
	Settings ProjectSettings `json:"settings"`
}
