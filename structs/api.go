package structs

import (
	"time"

	"golang.org/x/xerrors"
)

// UserType signifies how they have authenticated into Errorly
type UserType uint8

const (
	// DiscordUser is a user who has authenticated with discord
	// OAuth2
	DiscordUser UserType = iota
)

// WebhookType signifies how the payload should be sent.
type WebhookType uint8

const (
	// RegularPayload denotes the payload should be similar to a
	// REST response.
	RegularPayload WebhookType = iota
	// DiscordWebhook denotes the payload should supply a payload
	// that would accept by a discord webhook.
	DiscordWebhook
)

// ContentType signifies the comment type
type ContentType uint8

const (
	// Message denotes the comment is a regular message.
	Message ContentType = iota
	// IssueMarked denotes the issue status has changed. The flag is
	// available in data.
	IssueMarked
	// CommentsLocked denotes comments have been locked or unlocked.
	// Marked by a boolean as data.
	CommentsLocked
)

// EntryType signifies the entry status type
type EntryType uint8

const (
	// EntryActive means an issue has not been fixed and not assigned yet
	EntryActive EntryType = iota
	// EntryOpen means an issue has been assigned but not fixed yet
	EntryOpen
	// EntryInvalid means an issue is likely incorrect or false positive
	EntryInvalid
	// EntryResolved means an issue has been fixed
	EntryResolved
)

func (eT EntryType) String() string {
	switch eT {
	case EntryActive:
		return "EntryActive"
	case EntryOpen:
		return "EntryOpen"
	case EntryInvalid:
		return "EntryInvalid"
	case EntryResolved:
		return "EntryResolved"
	}
	return ""
}

// ParseEntryType converts a response string into a EntryType value.
// Returns an error if the input string does not match known values.
func ParseEntryType(entryTypeStr string) (EntryType, error) {

	switch entryTypeStr {
	case EntryActive.String():
		return EntryActive, nil
	case EntryOpen.String():
		return EntryOpen, nil
	case EntryInvalid.String():
		return EntryInvalid, nil
	case EntryResolved.String():
		return EntryResolved, nil
	}
	return EntryActive, xerrors.Errorf("Unknown EntryType String: '%s', defaulting to EntryActive", entryTypeStr)
}

// User contains the structure of a user
type User struct {
	ID       int64    `json:"id"`
	Name     string   `json:"name"`
	Avatar   string   `json:"avatar"`
	UserType UserType `json:"type"`
	HookID   int64    `json:"hook_id,omitempty"` // Used with usertype to reference an external ID (such as a discord id)

	CreatedAt time.Time `json:"created_at" pg:"default:now()"`

	ProjectIDs []int64 `json:"project_ids"`

	Token string `json:"-"`
}

// Project contains the structure of a project
type Project struct {
	ID int64 `json:"id"`

	CreatedAt   time.Time `json:"created_at" pg:"default:now()"`
	CreatedBy   *User     `json:"created_by" pg:"rel:has-one"`
	CreatedByID int64     `json:"created_by_id"`

	Integrations []Integration `json:"integrations" pg:"rel:has-many,join_fk:project_id"`
	Webhooks     []Webhook     `json:"webhooks" pg:"rel:has-many,join_fk:project_id"`
	Issues       []IssueEntry  `json:"issues" pg:"rel:has-many,join_fk:project_id"`

	Settings ProjectSettings `json:"settings"`

	// Cached values that change on an event in order to reduce lookup times
	OpenIssues     int `json:"open_issues"`
	ActiveIssues   int `json:"active_issues"`
	ResolvedIssues int `json:"resolved_issues"`
}

// ProjectSettings contains the structure of project settings
type ProjectSettings struct {
	DisplayName string `json:"display_name"` // Display Name
	Background  string `json:"background"`   // URI for background of image

	Description string `json:"description"`
	URL         string `json:"url"` // Link to a project appopriate URL. Will not show if left blank.

	Archived bool `json:"archived"` // When archived, no new issues can be made until unarchived by creator
	Private  bool `json:"private"`  // If a project is private, users can only view it if they have been added as a contributor

	Limited        bool    `json:"limited"`         // When enabled, only contributes can create errors
	ContributorIDs []int64 `json:"contributor_ids"` // Contributors for project
}

// Webhook contains the structure of a webhook integration
type Webhook struct {
	ID        int64 `json:"id"`
	ProjectID int64 `json:"project_id"`

	Active   bool  `json:"active"`   // Boolean if it is enabled
	Failures uint8 `json:"failures"` // If 4 failures sending webhook, will disable webhook

	CreatedAt   time.Time `json:"created_at" pg:"default:now()"`
	CreatedBy   *User     `json:"created_by" pg:"rel:has-one"`
	CreatedByID int64     `json:"created_by_id"`

	URL         string      `json:"url"`
	Type        WebhookType `json:"type"`
	JSONContent bool        `json:"json_content"` // When true, uses json else urlencoded
	Secret      string      `json:"secret"`       // Secret to send in the header to confirm origin
}

// Integration contains the structure of an integration
type Integration struct {
	ID        int64 `json:"id"`
	ProjectID int64 `json:"project_id"`

	Name string `json:"name"`

	CreatedAt   time.Time `json:"created_at" pg:"default:now()"`
	CreatedBy   *User     `json:"created_by" pg:"rel:has-one"`
	CreatedByID int64     `json:"created_by_id"`

	Token string `json:"token"`
}

// IssueEntry contains the structure of an issue entry
type IssueEntry struct {
	ID        int64 `json:"id"`
	ProjectID int64 `json:"project_id"`

	Starred bool `json:"starred"`

	Type        EntryType `json:"type"`
	Occurrences int       `json:"occurrences"`
	Assignee    User      `json:"assignee" pg:"rel:has-one"`
	AssigneeID  int64     `json:"assignee_id"`

	Error       string `json:"error"`
	Function    string `json:"function"`
	Checkpoint  string `json:"checkpoint"`
	Description string `json:"description"`
	Traceback   string `json:"traceback"`

	LastModified time.Time `json:"last_modified" pg:"default:now()"`

	CreatedAt   time.Time `json:"created_at" pg:"default:now()"`
	CreatedBy   *User     `json:"created_by" pg:"rel:has-one"`
	CreatedByID int64     `json:"creator_by_id"`

	CommentsLocked bool      `json:"comments_locked"`
	Comments       []Comment `json:"comment_ids" pg:"rel:has-many,join_fk:issue_id"`
}

// Comment contains the structure of an issue comment
type Comment struct {
	ID      int64 `json:"id"`
	IssueID int64 `json:"issue_id"`

	CreatedAt   time.Time `json:"created_at" pg:"default:now()"`
	CreatedBy   *User     `json:"created_by" pg:"rel:has-one"`
	CreatedByID int64     `json:"created_by_id"`

	Type           ContentType `json:"type"`
	Content        string      `json:"content,omitempty"`
	IssueMarked    EntryType   `json:"issue_marked,omitempty"`
	CommentsOpened bool        `json:"comments_opened,omitempty"`
}