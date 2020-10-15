package errorly

import (
	"io/ioutil"
	"net/http"
	"time"

	"github.com/TheRockettek/Errorly-Web/pkg/dictionary"
	"github.com/TheRockettek/Errorly-Web/structs"
	"github.com/go-pg/pg/v10"
	"github.com/hashicorp/go-uuid"
)

// LogoutHandler handles clearing a user session
func LogoutHandler(er *Errorly) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, _ := er.Store.Get(r, sessionName)
		defer session.Save(r, w)

		session.Values = make(map[interface{}]interface{})
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	}
}

// LoginHandler handles CSRF and AuthCode redirection
func LoginHandler(er *Errorly) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, _ := er.Store.Get(r, sessionName)
		defer session.Save(r, w)

		// Create a simple CSRF string to verify clients and 500 if we
		// cannot generate one.
		csrfString, err := uuid.GenerateUUID()
		if err != nil {
			http.Error(w, "Internal server error: "+err.Error(), 500)
			return
		}

		// Store the CSRF in the session then redirect the user to the
		// OAuth page.
		session.Values["oauth_csrf"] = csrfString

		url := er.Configuration.OAuth.AuthCodeURL(csrfString)
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	}
}

// OAuthCallbackHandler handles authenticating discord OAuth and creating
// a user profile if necessary
func OAuthCallbackHandler(er *Errorly) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, _ := er.Store.Get(r, sessionName)
		defer session.Save(r, w)

		// Validate the CSRF in the session and in the HTTP request.
		// If there is no CSRF in the session it is likely our fault :)
		_csrfString := r.URL.Query().Get("state")
		csrfString, ok := session.Values["oauth_csrf"].(string)
		if !ok {
			http.Error(w, "Missing CSRF state", http.StatusInternalServerError)
			return
		}

		if _csrfString != csrfString {
			http.Error(w, "Mismatched CSRF states", http.StatusUnauthorized)
		}

		// Just to be sure, remove the CSRF after we have compared the CSRF
		delete(session.Values, "oauth_csrf")

		// Create an OAuth exchange with the code we were given.
		code := r.URL.Query().Get("code")
		token, err := er.Configuration.OAuth.Exchange(er.ctx, code)
		if err != nil {
			http.Error(w, "Failed to exchange code: "+err.Error(), http.StatusInternalServerError)
		}

		// Create a client with our exchanged token and retrieve a user.
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

		// Lookup a user with the hook_id of the discord ID (used with
		// external login methods instead of using a discord ID as
		// an errorly user ID).
		user := &structs.User{}
		err = er.Postgres.Model(user).Where("hook_id = ?", discordUserResponse.ID).Select()
		if err != nil {
			if err == pg.ErrNoRows {
				// The user could not be found so create a new user and insert.
				user = &structs.User{
					ID:       er.IDGen.GenerateID(),
					UserType: structs.DiscordUser,
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
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		} else {
			// When we have a valid account, we will create a new user
			// token and update the username and avatar if necessary.
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

		// Once the user has logged in, send them back to the home page.
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	}
}

// APIMeHandler handles the /api/me request which returns the user
// object and a list of partial project objects
func APIMeHandler(er *Errorly) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, _ := er.Store.Get(r, sessionName)
		defer session.Save(r, w)

		// Authenticate the user
		auth, user := er.AuthenticateSession(session)

		var projects []structs.PartialProject
		if auth {
			// If the user has authenticated, create partial projects
			project := &structs.Project{}
			sanitizedProjectIDs := make([]int64, 0)
			projects = make([]structs.PartialProject, 0, len(user.ProjectIDs))

			for _, projectID := range user.ProjectIDs {
				err := er.Postgres.Model(project).Where("id = ?", projectID).Select()
				if err == nil {
					projects = append(projects, structs.PartialProject{
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

			// We will compare the sanitizedProjectIDs vs the cached
			// project IDs. If this is not equal, this means the
			// project has been deleted and should be removed from
			// the local project cache. We should update their
			// project IDs and update
			if len(sanitizedProjectIDs) != len(user.ProjectIDs) {
				user.ProjectIDs = sanitizedProjectIDs
				_, err := er.Postgres.Model(project).WherePK().Update()
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			}
		} else {
			// If you are not logged in, you cannot have any projects B)
			projects = make([]structs.PartialProject, 0)
		}

		resp, err := json.Marshal(structs.BaseResponse{
			Success: true,
			Data: structs.APIMe{
				Authenticated: auth,
				User:          user,
				Projects:      projects,
			}})
		if err != nil {
			// If we ever are here we've fucked up
			resp, _ := json.Marshal(structs.BaseResponse{
				Success: false,
				Error:   err.Error(),
			})
			http.Error(w, string(resp), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(resp)
	}
}

// APIDictionaryHandler handles the generation of a page dictionary used
// by vue-router.
func APIDictionaryHandler(er *Errorly) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, _ := er.Store.Get(r, sessionName)
		defer session.Save(r, w)

		// This function will return a PageDictionary object
		page, err := dictionary.GeneratePageDictionary(dictionaryPath, dictionaryOutputPath)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		resp, err := json.Marshal(structs.BaseResponse{
			Success: true,
			Data:    page})
		if err != nil {
			resp, _ := json.Marshal(structs.BaseResponse{
				Success: false,
				Error:   err.Error(),
			})
			http.Error(w, string(resp), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(resp)
	}
}

// func Handler(er *Errorly) http.HandlerFunc {
// 	return
// }

// func Handler(er *Errorly) http.HandlerFunc {
// 	return
// }
