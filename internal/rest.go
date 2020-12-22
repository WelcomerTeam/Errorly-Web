package errorly

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"net/http"
	"strings"
	"time"

	"github.com/TheRockettek/Errorly-Web/structs"
	"github.com/bwmarrin/snowflake"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

const (
	sessionName            = "session"
	discordUsersMe         = "https://discord.com/api/users/@me"
	discordRefreshDuration = time.Hour
)

// NewMethodRouter creates a new method router.
func NewMethodRouter() *MethodRouter {
	return &MethodRouter{mux.NewRouter()}
}

// MethodRouter beepboop.
type MethodRouter struct {
	*mux.Router
}

// HandleFunc registers a route that handles both paths and methods.
func (mr *MethodRouter) HandleFunc(path string, f func(http.ResponseWriter, *http.Request),
	methods ...string) *mux.Route {
	if len(methods) == 0 {
		methods = []string{"GET"}
	}

	return mr.NewRoute().Path(path).Methods(methods...).HandlerFunc(f)
}

// DiscordUser is the structure of a /users/@me request.
type DiscordUser struct {
	ID            snowflake.ID `json:"id" msgpack:"id"`
	Username      string       `json:"username" msgpack:"username"`
	Discriminator string       `json:"discriminator" msgpack:"discriminator"`
	Avatar        string       `json:"avatar" msgpack:"avatar"`
	// MFAEnabled    bool         `json:"mfa_enabled,omitempty" msgpack:"mfa_enabled,omitempty"`
	// Locale        string       `json:"locale,omitempty" msgpack:"locale,omitempty"`
	Verified bool `json:"verified,omitempty" msgpack:"verified,omitempty"`
	// Email    string `json:"email,omitempty" msgpack:"email,omitempty"`
	// Flags         int          `json:"flags" msgpack:"flags"`
	// PremiumType   int          `json:"premium_type" msgpack:"premium_type"`
}

// CreateUserToken creates a user token for a user.
func CreateUserToken(u *structs.User) string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	res := make([]byte, 8)
	binary.BigEndian.PutUint64(res, uint64(u.ID))

	return base64.URLEncoding.EncodeToString(res) + "." + base64.URLEncoding.EncodeToString(b)
}

// ParseUserToken returns the user id and random.
func ParseUserToken(token string) (uid int64, random []byte, valid bool) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return 0, nil, false
	}

	_uid, err := base64.URLEncoding.DecodeString(parts[0])
	if err != nil {
		return 0, nil, false
	}

	uid = int64(binary.BigEndian.Uint64(_uid))

	random, err = base64.URLEncoding.DecodeString(parts[1])
	if err != nil {
		return 0, nil, false
	}

	return uid, random, true
}

// AuthenticateSession verifies the session is valid.
func (er *Errorly) AuthenticateSession(session *sessions.Session) (auth bool, user *structs.User) {
	token, ok := session.Values["token"].(string)
	if !ok {
		return false, nil
	}

	uid, _, valid := ParseUserToken(token)
	if !valid {
		return false, nil
	}

	_user := &structs.User{}

	err := er.Postgres.Model(_user).Where("id = ?", uid).Select()
	if err != nil {
		er.Logger.Error().Err(err).Msg("Failed to fetch user")

		return false, nil
	}

	if _user.Token != token {
		return false, nil
	}

	return true, _user
}

func createEndpoints(er *Errorly) (router *MethodRouter) {
	router = NewMethodRouter()

	router.HandleFunc("/login", LoginHandler(er), "GET")
	router.HandleFunc("/logout", LogoutHandler(er), "GET")
	router.HandleFunc("/oauth2/callback", OAuthCallbackHandler(er), "GET")

	router.HandleFunc("/api/me", APIMeHandler(er), "GET")

	// Projects:
	router.HandleFunc("/api/projects", APIProjectCreateHandler(er), "POST")
	// Creates a project
	router.HandleFunc("/api/project/{project_id}", APIProjectHandler(er), "GET")
	// Project information and page 1 of issues
	router.HandleFunc("/api/project/{project_id}/lazy", APIProjectLazyHandler(er), "GET")
	// Returns partial user objects from provided user arguments
	router.HandleFunc("/api/project/{project_id}/execute", APIProjectExecutorHandler(er), "POST")
	// Execute task (star, assign, unassign etc.)
	router.HandleFunc("/api/project/{project_id}/contributors", APIProjectContributorsHandler(er), "GET")
	// Returns partial user objects of all contributors
	router.HandleFunc("/api/project/{project_id}", APIProjectUpdateHandler(er), "POST")
	// Update project settings
	router.HandleFunc("/api/project/{project_id}/delete", APIProjectDeleteHandler(er), "POST")
	// Deletes the project, only the project owner can do this

	router.HandleFunc("/api/project/{project_id}/transfer", APIProjectTransferHandler(er), "POST")
	// Transfers project to another user
	router.HandleFunc("/api/project/{project_id}/contributor/{contributor}", APIProjectContributorsRemoveHandler(er), "DELETE")
	// Removes a contributor

	// Issues:
	router.HandleFunc("/api/project/{project_id}/issues", APIProjectIssueHandler(er), "GET")
	// Returns issued based off of a query
	router.HandleFunc("/api/project/{project_id}/issues", APIProjectIssueCreateHandler(er), "POST")
	// Create issue
	router.HandleFunc("/api/project/{project_id}/issue/{issue_id}", APIProjectFetchIssueHandler(er), "GET")
	// Fetches issue. alias for /api/project/{project_id}/issues?issue=?
	router.HandleFunc("/api/project/{project_id}/issue/{issue_id}/delete", APIProjectIssueDeleteHandler(er), "POST")
	// Deletes issue, elevated or issue creator can do this.

	// Comments:
	router.HandleFunc("/api/project/{project_id}/issue/{issue_id}/comments", APIProjectIssueCommentHandler(er), "GET")
	//  Lists issue comments
	router.HandleFunc("/api/project/{project_id}/issue/{issue_id}/comments", APIProjectIssueCommentCreateHandler(er), "POST")
	// Create issue comment
	// PATCH /api/project/{project_id}/issue/{issue_id}/comments - Updates issue comment
	// DELETE /api/project/{project_id}/issue/{issue_id}/comments - Deletes issue comment

	// Invites:
	// GET /api/project/{project_id}/join/{join_code} - Get invite code
	// POST /api/project/{project_id}/join/{join_code} - Join invite code
	// POST /api/project/{project_id}/join/{join_code}/delete - Remove invite code
	// POST /api/project/{project_id}/join - Create invite code

	// Webhooks:
	// POST /api/project/{project_id}/webhook - Creates a webhook
	// PATCH /api/project/{project_id}/webhook/{webhook_id} - Updates a webhook
	// DELETE /api/project/{project_id}/webhook/{webhook_id} - Deletes a webhook
	// POST /api/project/{project_id}/webhook/{webhook_id}/test - Tests webhook

	// Integrations:
	// POST /api/project/{project_id}/integration - Creates an integration
	// PATCH /api/project/{project_id}/integration - Updates an integration
	// DELETE /api/project/{project_id}/integration - Deletes an integration
	// POST /api/project/{project_id}/integration/{integration_id}/regenerate - Creates a new integration token

	return router
}
