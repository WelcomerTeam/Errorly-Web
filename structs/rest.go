package structs

// BaseResponse is the structure of all REST requests
type BaseResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// APIMe is the structure of the /api/me endpoint
type APIMe struct {
	Authenticated bool             `json:"authenticated"`
	User          User             `json:"user"`
	Projects      []PartialProject `json:"projects"`
}

// PartialProject is a partial version of a project object
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

// PartialUser is the partial version of a user object
type PartialUser struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Avatar      string `json:"avatar"`
	Integration bool   `json:"integration"`
}

// APIProject is the structure of the GET /api/project/{id} endpoint
type APIProject struct {
	Project      Project               `json:"project"`
	Contributors map[int64]PartialUser `json:"contributors"`
}

// APIProjectLazy is the structure of the GET /api/project/{id}/lazy endpoint
type APIProjectLazy struct {
	Users map[int64]PartialUser `json:"users"`
	IDs   []int64               `json:"ids"`
}

// APIProjectIssues is the structure of the GET /api/project/{id}/issues endpoint
type APIProjectIssues struct {
	Page        int          `json:"page"`
	TotalIssues int          `json:"total_issues"`
	Issues      []IssueEntry `json:"issues"`
}
