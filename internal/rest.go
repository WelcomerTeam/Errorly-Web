package errorly

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/hashicorp/go-uuid"
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
func CreateUserToken(u *User) string {
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
func (er *Errorly) AuthenticateSession(session *sessions.Session) (auth bool, user *User) {
	token, ok := session.Values["token"].(string)
	if !ok {
		return false, nil
	}

	uid, _, valid := ParseUserToken(token)
	if !valid {
		return false, nil
	}

	user = &User{}
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

func (er *Errorly) createEndpoints() {
	router := er.Router

	router.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		session, _ := er.Store.Get(r, sessionName)
		defer session.Save(r, w)

		session.Values = make(map[interface{}]interface{})
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	}, "GET")

	router.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		session, _ := er.Store.Get(r, sessionName)
		defer session.Save(r, w)

		csrfString, err := uuid.GenerateUUID()
		if err != nil {
			http.Error(w, "Internal server error: "+err.Error(), 500)
			return
		}

		session.Values["oauth_csrf"] = csrfString

		url := er.Configuration.OAuth.AuthCodeURL(csrfString)
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	}, "GET")

	router.HandleFunc("/oauth2/callback", func(w http.ResponseWriter, r *http.Request) {
		session, _ := er.Store.Get(r, sessionName)
		defer session.Save(r, w)

		_csrfString := r.URL.Query().Get("state")
		csrfString, ok := session.Values["oauth_csrf"].(string)
		if !ok {
			http.Error(w, "Missing CSRF state", http.StatusInternalServerError)
			return
		}

		if _csrfString != csrfString {
			http.Error(w, "Mismatched CSRF states", http.StatusUnauthorized)
		}

		delete(session.Values, "oauth_csrf")

		code := r.URL.Query().Get("code")
		token, err := er.Configuration.OAuth.Exchange(er.ctx, code)
		if err != nil {
			http.Error(w, "Failed to exchange code: "+err.Error(), http.StatusInternalServerError)
		}

		client := er.Configuration.OAuth.Client(er.ctx, token)
		resp, err := client.Get(discordUsersMe)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		discordUserResponse := &DiscordUser{}
		err = json.Unmarshal(body, &discordUserResponse)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		user := &User{}
		err = er.Postgres.Model(user).Where("hook_id = ?", discordUserResponse.ID).Select()
		if err != nil {
			user = &User{
				ID:       er.IDGen.GenerateID(),
				UserType: discordUser,
				HookID:   discordUserResponse.ID.Int64(),

				Name:       discordUserResponse.Username,
				Avatar:     "https://cdn.discordapp.com/avatars/" + discordUserResponse.ID.String() + "/" + discordUserResponse.Avatar + ".png",
				CreatedAt:  time.Now().UTC(),
				ProjectIDs: make([]int64, 0),
			}

			token := CreateUserToken(user)
			user.Token = token

			_, err = er.Postgres.Model(user).WherePK().Insert()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			session.Values["token"] = token
		} else {
			token := CreateUserToken(user)
			user.Avatar = "https://cdn.discordapp.com/avatars/" + discordUserResponse.ID.String() + "/" + discordUserResponse.Avatar + ".png"
			user.Name = discordUserResponse.Username
			user.Token = token

			_, err = er.Postgres.Model(user).WherePK().Update()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			session.Values["token"] = token
		}

		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	}, "GET")

	router.HandleFunc("/api/me", func(w http.ResponseWriter, r *http.Request) {
		session, _ := er.Store.Get(r, sessionName)
		defer session.Save(r, w)

		auth, user := er.AuthenticateSession(session)

		var projects []PartialProject
		if auth {
			project := &Project{}

			sanitizedProjectIDs := make([]int64, 0)
			projects = make([]PartialProject, 0, len(user.ProjectIDs))

			for _, projectID := range user.ProjectIDs {
				err := er.Postgres.Model(project).Where("id = ?", projectID).Select()
				if err == nil {
					projects = append(projects, PartialProject{
						ID:             project.ID,
						Name:           project.Settings.DisplayName,
						Description:    project.Settings.Description,
						Archived:       project.Settings.Archived,
						Private:        project.Settings.Private,
						OpenIssues:     project.OpenIssues,
						ActiveIssues:   project.ActiveIssues,
						ResolvedIssues: project.ResolvedIssues,
					})
					sanitizedProjectIDs = append(sanitizedProjectIDs, projectID)
				}
			}
			if !reflect.DeepEqual(sanitizedProjectIDs, user.ProjectIDs) {
				user.ProjectIDs = sanitizedProjectIDs
				_, err := er.Postgres.Model(project).WherePK().Update()
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			}
		} else {
			projects = make([]PartialProject, 0)
		}

		resp, err := json.Marshal(BaseResponse{
			Success: true,
			Data: APIMe{
				Authenticated: auth,
				User:          user,
				Projects:      projects,
			}})
		if err != nil {
			resp, _ := json.Marshal(BaseResponse{
				Success: false,
				Error:   err.Error(),
			})
			http.Error(w, string(resp), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(resp)
	}, "GET")

	router.HandleFunc("/api/dictionary", func(w http.ResponseWriter, r *http.Request) {
		session, _ := er.Store.Get(r, sessionName)
		defer session.Save(r, w)

		// const dictionaryPath = "pages.json"
		// const dictionaryOutputPath = "dictionary.json"

		page, err := generatePageDictionary(dictionaryPath, dictionaryOutputPath)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		resp, err := json.Marshal(BaseResponse{
			Success: true,
			Data:    page})
		if err != nil {
			resp, _ := json.Marshal(BaseResponse{
				Success: false,
				Error:   err.Error(),
			})
			http.Error(w, string(resp), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(resp)
	}, "GET")

	// router.HandleFunc("", func(w http.ResponseWriter, r *http.Request) {
	//     session, _ := er.Store.Get(r, sessionName)
	//     defer session.Save(r, w)
	// }, "GET")

	return
}
