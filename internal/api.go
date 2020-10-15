package errorly

import (
	"sync"
	"time"

	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
	"golang.org/x/xerrors"
)

// startEpoch for IDs (12/10/2020 13:01:14)
const epoch = 1602507674941

// UserType signifies how they have authenticated into Errorly
type UserType uint8

const (
	discordUser UserType = iota
)

// WebhookType signifies how the payload should be sent.
type WebhookType uint8

const (
	// RegularPayload denotes the payload should be similar to a
	// REST response.
	regularPayload WebhookType = iota
	// DiscordWebhook denotes the payload should supply a payload
	// that would accept by a discord webhook.
	discordWebhook
)

// ContentType signifies the comment type
type ContentType uint8

const (
	// Message denotes the comment is a regular message.
	message ContentType = iota
	// IssueMarked denotes the issue status has changed. The flag is
	// available in data.
	issueMarked
	// CommentsLocked denotes comments have been locked or unlocked.
	// Marked by a boolean as data.
	commentsLocked
)

// EntryType signifies the entry status type
type EntryType uint8

const (
	entryActive EntryType = iota
	entryOpen
	entryInvalid
	entryResolved
)

func (eT EntryType) String() string {
	switch eT {
	case entryActive:
		return "entryActive"
	case entryOpen:
		return "entryOpen"
	case entryInvalid:
		return "entryInvalid"
	case entryResolved:
		return "entryResolved"
	}
	return ""
}

// ParseUserType converts a response string into a UserType value.
// Returns an error if the input string does not match known values.
func ParseUserType(entryTypeStr string) (EntryType, error) {
	switch entryTypeStr {
	case entryActive.String():
		return entryActive, nil
	case entryOpen.String():
		return entryOpen, nil
	case entryInvalid.String():
		return entryInvalid, nil
	case entryResolved.String():
		return entryResolved, nil
	}
	return entryActive, xerrors.Errorf("Unknown EntryType String: '%s', defaulting to entryActive", entryTypeStr)
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

	Type       EntryType `json:"type"`
	Occurences int       `json:"ocurrences"`
	Assignee   User      `json:"assignee" pg:"rel:has-one"`
	AssigneeID int64     `json:"assignee_id"`

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

// NewIDGenerator returns an IDGenerator
func NewIDGenerator(initialEpoch int64, shardID int64) *IDGenerator {
	return &IDGenerator{
		initialEpoch: initialEpoch,
		shardID:      shardID,
		sequence:     0,
	}
}

// IDGenerator contains the structure of the id generator.
// IDs are comprised of a similar structure to intagram where 41 bits
// contain the timestamp then a further 23 bits for a shard and sequence.
type IDGenerator struct {
	sync.Mutex
	initialEpoch int64
	shardID      int64
	sequence     int64
}

// GenerateID returns a new id that is int64
func (id *IDGenerator) GenerateID() int64 {
	id.Lock()
	defer id.Unlock()

	ms := time.Now().UTC().UnixNano() / int64(time.Millisecond)
	id.sequence++ // We only want 10 bits
	return ((ms - id.initialEpoch) << 23) | (id.shardID << 10) | (id.sequence % 1024)
}

// createSchema creates database schema
func createSchema(db *pg.DB) (err error) {
	models := []interface{}{
		(*User)(nil),
		(*Project)(nil),
		(*Integration)(nil),
		(*Webhook)(nil),
		(*IssueEntry)(nil),
		(*Comment)(nil),
	}

	for _, model := range models {
		// db.Model(model).DropTable(&orm.DropTableOptions{
		// 	IfExists: true,
		// 	Cascade:  true,
		// })
		err = db.Model(model).CreateTable(&orm.CreateTableOptions{
			IfNotExists: true,
		})
		if err != nil {
			return err
		}
	}

	// idGen := IDGenerator{initialEpoch: 1602507674941, shardID: 0, sequence: 0}

	// println("Create User")
	// user := &User{
	// 	ID:         idGen.GenerateID(),
	// 	Name:       "ImRock",
	// 	Avatar:     "",
	// 	UserType:   discordUser,
	// 	ProjectIDs: make([]int64, 0),
	// }
	// _, err = db.Model(user).Insert()
	// if err != nil {
	// 	return err
	// }

	// user := &User{}
	// err = db.Model(user).Where("id = ?", 739602763612161).Select()

	// println("Create Project")
	// project := &Project{
	// 	ID: idGen.GenerateID(),

	// 	CreatedAt:   time.Now().UTC(),
	// 	CreatedByID: user.ID,

	// 	Integrations: make([]Integration, 0),
	// 	Webhooks:     make([]Webhook, 0),

	// 	Settings: ProjectSettings{
	// 		DisplayName:    "Welcomer",
	// 		Archived:       false,
	// 		Limited:        false,
	// 		ContributorIDs: make([]int64, 0),
	// 	},

	// 	OpenIssues:     0,
	// 	ActiveIssues:   0,
	// 	ResolvedIssues: 0,
	// }
	// _, err = db.Model(project).Insert()
	// if err != nil {
	// 	return err
	// }

	// println("Append Project ID to user")
	// user.ProjectIDs = append(user.ProjectIDs, project.ID)
	// _, err = db.Model(user).WherePK().Update()
	// if err != nil {
	// 	return err
	// }

	// println("Create integration")
	// integration := &Integration{
	// 	ID:        idGen.GenerateID(),
	// 	ProjectID: project.ID,

	// 	Name: "Welcomer",

	// 	CreatedAt:   time.Now().UTC(),
	// 	CreatedByID: user.ID,
	// }
	// _, err = db.Model(integration).Insert()
	// if err != nil {
	// 	return err
	// }

	// println("Create issue")
	// issue := &IssueEntry{
	// 	ID:        idGen.GenerateID(),
	// 	ProjectID: project.ID,

	// 	Starred: false,

	// 	Type:       entryOpen,
	// 	Occurences: 1,

	// 	Error:       "GenericError",
	// 	Function:    "error()",
	// 	Checkpoint:  "folder/file.go:linename",
	// 	Description: "generic error message",

	// 	LastModified: time.Now().UTC(),

	// 	CreatedAt:   time.Now().UTC(),
	// 	CreatedByID: integration.ID,

	// 	CommentsLocked: false,
	// 	Comments:       make([]Comment, 0),
	// }
	// _, err = db.Model(issue).Insert()
	// if err != nil {
	// 	return err
	// }

	// println("Create comment")
	// comment := &Comment{
	// 	ID:      idGen.GenerateID(),
	// 	IssueID: issue.ID,

	// 	CreatedAt:   time.Now().UTC(),
	// 	CreatedByID: user.ID,

	// 	Type:    message,
	// 	Content: "fuck",
	// }
	// _, err = db.Model(comment).Insert()
	// if err != nil {
	// 	return err
	// }

	// println("Select project")
	// _project := &Project{}
	// err = db.Model(_project).Where("project.id = ?", project.ID).Relation("CreatedBy").Relation("Integrations").Relation("Webhooks").Select()
	// if err != nil {
	// 	return err
	// }

	// println("Marshal")
	// data, err := json.Marshal(_project)
	// if err != nil {
	// 	return err
	// }

	// println(string(data))

	return nil
}
