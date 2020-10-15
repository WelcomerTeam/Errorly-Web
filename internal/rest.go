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

const dictionaryPath = "web/pages.json"
const dictionaryOutputPath = "web/dictionary.json"

const sessionName = "session"
const discordUsersMe = "https://discord.com/api/users/@me"
const discordRefreshDuration = time.Hour

// - User
// FetchUser(id)
// CreateUser(user)

// JoinProject(user, project)
// LeaveProject(user, project)

// - Project
// CreateProject(project)
// DeleteProject(project)

// AddIntegration(project, integration)
// RemoveIntegration(project, integration)
// RegenerateIntegrationToken(project, integration)

// CreateWebhook(project, webhook)
// DeleteWebhook(project, webhook)
// TestWebhook(project, webhook)

// UpdateProjectSettings(project)
// CreateProjectIssue(project, issue, force)

// FetchProject(project)
// FetchProjectIssues(project, limit, sorted)

// - Issue
// AssignToIssue(project, issue, assigned, deassigned)
// MarkIssue(project, issue, type)
// LockComments(project, issue, locked)

// UpdateIssueData(project, issue)
// StarIssue(project, issue, starred)

// FetchIssue(project, issue)
// FetchIssueComments(project, issue)

// - Comment
// CreateComment(project, issue, content)
// // assign, marked, lock are internal and ran in issue command

// FetchComment(project, issue, comment)

// NewMethodRouter creates a new method router
func NewMethodRouter() *MethodRouter {
	return &MethodRouter{mux.NewRouter()}
}

// MethodRouter beepboop
type MethodRouter struct {
	*mux.Router
}

// HandleFunc registers a route that handles both paths and methods
func (mr *MethodRouter) HandleFunc(path string, f func(http.ResponseWriter, *http.Request), methods ...string) *mux.Route {
	if len(methods) == 0 {
		methods = []string{"GET"}
	}
	return mr.NewRoute().Path(path).Methods(methods...).HandlerFunc(f)
}

// DiscordUser is the structure of a /users/@me request
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
	rand.Read(b)
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

// AuthenticateSession verifies the session is valid
func (er *Errorly) AuthenticateSession(session *sessions.Session) (auth bool, user *structs.User) {
	token, ok := session.Values["token"].(string)
	if !ok {
		return false, nil
	}

	uid, _, valid := ParseUserToken(token)
	if !valid {
		return false, nil
	}

	user = &structs.User{}
	err := er.Postgres.Model(user).Where("id = ?", uid).Select()
	if err != nil {
		er.Logger.Error().Err(err).Msg("Failed to fetch user")
		return false, nil
	}

	if user.Token != token {
		return false, nil
	}

	return true, user
}

func createEndpoints(er *Errorly) (router *MethodRouter) {
	router = NewMethodRouter()

	router.HandleFunc("/login", LoginHandler(er), "GET")
	router.HandleFunc("/oauth2/callback", OAuthCallbackHandler(er), "GET")
	router.HandleFunc("/logout", LogoutHandler(er), "GET")

	router.HandleFunc("/api/me", APIMeHandler(er), "GET")
	router.HandleFunc("/api/dictionary", APIDictionaryHandler(er), "GET")

	// router.HandleFunc("", func(w http.ResponseWriter, r *http.Request) {
	//     session, _ := er.Store.Get(r, sessionName)
	//     defer session.Save(r, w)
	// }, "GET")

	return
}
