package errorly

import (
	"encoding/csv"
	"errors"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/TheRockettek/Errorly-Web/structs"
	"github.com/derekstavis/go-qs"
	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/hashicorp/go-uuid"
	"golang.org/x/xerrors"
)

// Issues per page.
const pageLimit = 25 // TODO: Move to webapp

// func parseJSONForm(r *http.Request) (vars map[string]string, err error) {
// 	err = json.NewDecoder(r.Body).Decode(&vars)
// 	return
// }

func passResponse(rw http.ResponseWriter, data interface{}, success bool, status int) {
	var resp []byte

	var err error

	if success {
		resp, err = json.Marshal(structs.BaseResponse{
			Success: true,
			Data:    data,
		})
	} else {
		resp, err = json.Marshal(structs.BaseResponse{
			Success: false,
			Error:   data.(string),
		})
	}

	if err != nil {
		resp, _ := json.Marshal(structs.BaseResponse{
			Success: false,
			Error:   err.Error(),
		})

		http.Error(rw, string(resp), http.StatusInternalServerError)

		return
	}

	if success {
		rw.WriteHeader(status)
		_, err = rw.Write(resp)

		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
		}
	} else {
		http.Error(rw, string(resp), status)
	}
}

func verifyProjectVisibility(er *Errorly, rw http.ResponseWriter, vars map[string]string,
	user *structs.User, auth bool, basic bool) (project *structs.Project, viewable bool, elevated bool, ok bool) {
	// Retrieve project_id from /project/{project_id}.
	_projectID, ok := vars["project_id"]

	if !ok {
		passResponse(rw, "Missing Project ID", false, http.StatusBadRequest)

		return
	}

	// Check projectID is a valid number.
	projectID, err := strconv.ParseInt(_projectID, 10, 64)
	if err != nil {
		ok = false
		passResponse(rw, "Could not find this project", false, http.StatusBadRequest)

		return
	}

	// Now we have a possibly valid ID we will first get the project.
	project = &structs.Project{
		Integrations: make([]*structs.User, 0),
		Webhooks:     make([]*structs.Webhook, 0),
		InviteCodes:  make([]*structs.InviteCode, 0),
	}

	query := er.Postgres.Model(project).
		Where("project.id = ?", projectID).
		Relation("Integrations")

	// If we do not want a basic config, we will also pass webhooks and invite codes.
	if !basic {
		query = query.
			Relation("Webhooks").
			Relation("InviteCodes")
	}

	err = query.Select()
	if err != nil {
		if errors.Is(err, pg.ErrNoRows) {
			// Invalid project ID
			ok = false
			passResponse(rw, "Could not find this project", false, http.StatusBadRequest)

			return
		}

		// Unexpected error
		ok = false
		passResponse(rw, err.Error(), false, http.StatusInternalServerError)

		return
	}

	viewable = true
	elevated = false

	// If the project is private, check if the user is able to view
	if project.Settings.Private {
		viewable = false
	}

	// Check if the user is able to view
	// (if there is a user logged in)
	if auth {
		for _, uid := range project.Settings.ContributorIDs {
			if uid == user.ID {
				viewable = true
				elevated = true
			}
		}

		if user.ID == project.CreatedByID {
			elevated = true
			viewable = true
		}
	}

	return project, viewable, elevated, true
}

func parseSorting(s string) string {
	if strings.ToUpper(s) == "DESC" {
		return "DESC"
	}

	return "ASC"
}

func fetchProjectIssues(er *Errorly, projectID int64, limit int, page int,
	query string, userID int64) (issues []structs.IssueEntry, totalissues int, err error) {
	// sort:created_by-desc
	_issues := make([]structs.IssueEntry, 0, limit)

	initialQuery := er.Postgres.Model(&_issues).
		Where("issue_entry.project_id = ?", projectID).
		Order("starred DESC")

	parts := []string{}
	query = strings.ReplaceAll(query, `'`, `"`)

	if len(query) > 0 {
		r := csv.NewReader(strings.NewReader(query))
		r.Comma = ' '

		parts, err = r.Read()
		if err != nil {
			return
		}
	}

	// // fetchStarred := false
	// fuzzyEntries := make([]string, 0)

	for _, part := range parts {
		subpart := strings.Split(part, ":")
		// if len(subpart) < 2 {
		// 	fuzzyEntries = append(fuzzyEntries, subpart[0])
		// } else {
		finger, thumb := subpart[0], subpart[1]
		switch finger {
		// new query's
		// case "has":
		// 	// assignee
		// 	// traceback
		// 	// messages
		// 	// star
		// case "no":
		// 	// assignee
		// 	// traceback
		// 	// messages
		// 	// star
		case "sort":
			thumbvalues := strings.Split(thumb, "-")
			if len(thumbvalues) >= 1 {
				if len(thumbvalues) == 1 {
					thumbvalues = append(thumbvalues, "DESC")
				}

				switch thumb := strings.ToLower(thumbvalues[0]); thumb {
				case "starred", "type", "occurrences", "assignee_id", "error", "function",
					"checkpoint", "last_modified", "created_at", "comment_count":
					initialQuery = initialQuery.Order(thumb + " " + parseSorting(thumbvalues[1]))
				}
			}
		case "is":
			switch strings.ToLower(thumb) {
			case "active":
				// initialQuery = initialQuery.Where("type = ?", structs.EntryActive)
				initialQuery = initialQuery.Where("type is NULL", structs.EntryActive)
			case "open":
				initialQuery = initialQuery.Where("type = ?", structs.EntryOpen)
			case "invalid":
				initialQuery = initialQuery.Where("type = ?", structs.EntryInvalid)
			case "resolved":
				initialQuery = initialQuery.Where("type = ?", structs.EntryResolved)
			case "starred":
				initialQuery = initialQuery.Where("starred = ?", true)
			}
		case "author", "from":
			switch strings.ToLower(thumb) {
			case "@me":
				if userID != 0 {
					initialQuery = initialQuery.Where("created_by_id = ?", userID)
				}
			case "no":
				initialQuery = initialQuery.WhereGroup(func(q *orm.Query) (*orm.Query, error) {
					q = q.WhereOr("created_by_id = ?", 0).WhereOr("created_by_id IS NULL")

					return q, nil
				})
			default:
				id, err := strconv.Atoi(thumb)
				if err == nil {
					initialQuery = initialQuery.Where("created_by_id = ?", id)
				}
			}
		case "assigned", "assignee":
			switch strings.ToLower(thumb) {
			case "@me":
				if userID != 0 {
					initialQuery = initialQuery.Where("assignee_id = ?", userID)
				}
			case "no":
				initialQuery = initialQuery.WhereGroup(func(q *orm.Query) (*orm.Query, error) {
					q = q.WhereOr("assignee_id = ?", 0).WhereOr("assignee_id IS NULL")

					return q, nil
				})
			default:
				id, err := strconv.Atoi(thumb)
				if err == nil {
					initialQuery = initialQuery.Where("assignee_id = ?", id)
				}
			}
		}
	}

	count, err := initialQuery.Limit(limit).Offset(int(math.Max(0, float64(limit*page)))).SelectAndCount()
	if err != nil {
		return
	}

	return _issues, count, nil
}

func isContributor(project *structs.Project, id int64) bool {
	if id == project.CreatedByID {
		return true
	}

	for _, b := range project.Integrations {
		if b.ID == id {
			return true
		}
	}

	for _, b := range project.Settings.ContributorIDs {
		if b == id {
			return true
		}
	}

	return false
}

// SaveSession should be used as a defer when handling requests.
func (er *Errorly) SaveSession(s *sessions.Session, r *http.Request, rw http.ResponseWriter) {
	if err := s.Save(r, rw); err != nil {
		er.Logger.Error().Err(err).Msg("Failed to save session")
	}
}

// LoginHandler handles CSRF and AuthCode redirection.
func LoginHandler(er *Errorly) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		session, _ := er.Store.Get(r, sessionName)
		defer er.SaveSession(session, r, rw)

		// Create a simple CSRF string to verify clients and 500 if we
		// cannot generate one.
		csrfString, err := uuid.GenerateUUID()
		if err != nil {
			http.Error(rw, "Internal server error: "+err.Error(), http.StatusInternalServerError)

			return
		}

		// Store the CSRF in the session then redirect the user to the
		// OAuth page.
		session.Values["oauth_csrf"] = csrfString

		url := er.Configuration.OAuth.AuthCodeURL(csrfString)
		http.Redirect(rw, r, url, http.StatusTemporaryRedirect)
	}
}

// LogoutHandler handles clearing a user session.
func LogoutHandler(er *Errorly) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		session, _ := er.Store.Get(r, sessionName)
		defer er.SaveSession(session, r, rw)

		session.Values = make(map[interface{}]interface{})

		http.Redirect(rw, r, "/", http.StatusTemporaryRedirect)
	}
}

// OAuthCallbackHandler handles authenticating discord OAuth and creating
// a user profile if necessary.
func OAuthCallbackHandler(er *Errorly) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		session, _ := er.Store.Get(r, sessionName)
		defer er.SaveSession(session, r, rw)

		urlQuery := r.URL.Query()

		// Validate the CSRF in the session and in the HTTP request.
		// If there is no CSRF in the session it is likely our fault :)
		_csrfString := urlQuery.Get("state")
		csrfString, ok := session.Values["oauth_csrf"].(string)

		if !ok {
			// http.Error(rw, "Missing CSRF state", http.StatusInternalServerError)
			http.Redirect(rw, r, "/login", http.StatusTemporaryRedirect)

			return
		}

		if _csrfString != csrfString {
			// http.Error(rw, "Mismatched CSRF states", http.StatusUnauthorized)
			http.Redirect(rw, r, "/login", http.StatusTemporaryRedirect)

			return
		}

		// Just to be sure, remove the CSRF after we have compared the CSRF
		delete(session.Values, "oauth_csrf")

		// Create an OAuth exchange with the code we were given.
		code := urlQuery.Get("code")

		token, err := er.Configuration.OAuth.Exchange(er.ctx, code)
		if err != nil {
			// http.Error(rw, "Failed to exchange code: "+err.Error(), http.StatusInternalServerError)
			http.Redirect(rw, r, "/login", http.StatusTemporaryRedirect)
		}

		// Create a client with our exchanged token and retrieve a user.
		client := er.Configuration.OAuth.Client(er.ctx, token)

		resp, err := client.Get(discordUsersMe) // nolint:noctx
		if err != nil {
			// http.Error(rw, err.Error(), http.StatusInternalServerError)
			http.Redirect(rw, r, "/login", http.StatusTemporaryRedirect)

			return
		}

		body, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()

		if err != nil {
			// http.Error(rw, err.Error(), http.StatusInternalServerError)
			http.Redirect(rw, r, "/login", http.StatusTemporaryRedirect)

			return
		}

		discordUserResponse := &DiscordUser{}

		err = json.Unmarshal(body, &discordUserResponse)
		if err != nil {
			// http.Error(rw, err.Error(), http.StatusInternalServerError)
			http.Redirect(rw, r, "/login", http.StatusTemporaryRedirect)

			return
		}

		// Lookup a user with the hook_id of the discord ID (used with
		// external login methods instead of using a discord ID as
		// an errorly user ID).
		user := &structs.User{}

		err = er.Postgres.Model(user).
			Where("hook_id = ?", discordUserResponse.ID).
			Select()
		if err != nil {
			if errors.Is(err, pg.ErrNoRows) {
				// The user could not be found so create a new user and insert.
				user = &structs.User{
					ID:       er.IDGen.GenerateID(),
					UserType: structs.DiscordUser,
					HookID:   discordUserResponse.ID.Int64(),

					Name: discordUserResponse.Username,
					Avatar: "https://cdn.discordapp.com/avatars/" +
						discordUserResponse.ID.String() + "/" +
						discordUserResponse.Avatar + ".png?size=32",
					CreatedAt:  time.Now().UTC(),
					ProjectIDs: make([]int64, 0),
				}

				uid, rand := CreateUserToken(user)
				user.Token = rand

				_, err = er.Postgres.Model(user).
					WherePK().
					Insert()
				if err != nil {
					http.Error(rw, err.Error(), http.StatusInternalServerError)
				}

				session.Values["token"] = uid + "." + rand
			} else {
				http.Error(rw, err.Error(), http.StatusInternalServerError)

				return
			}
		} else {
			// When we have a valid account, we will create a new user
			// token and update the username and avatar if necessary.
			uid, rand := CreateUserToken(user)

			user.Avatar = "https://cdn.discordapp.com/avatars/" +
				discordUserResponse.ID.String() + "/" +
				discordUserResponse.Avatar + ".png?size=32"
			user.Name = discordUserResponse.Username
			user.Token = rand

			_, err = er.Postgres.Model(user).
				WherePK().
				Update()
			if err != nil {
				http.Error(rw, err.Error(), http.StatusInternalServerError)
			}

			session.Values["token"] = uid + "." + rand
		}

		// Once the user has logged in, send them back to the home page.
		http.Redirect(rw, r, "/", http.StatusTemporaryRedirect)
	}
}

// APIMeHandler handles the /api/me request which returns the user
// object and a list of partial project objects.
func APIMeHandler(er *Errorly) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		session, _ := er.Store.Get(r, sessionName)
		defer er.SaveSession(session, r, rw)

		// Authenticate the user
		auth, user := er.AuthenticateSession(session)

		var projects []structs.PartialProject

		if auth {
			// If the user has authenticated, create partial projects
			project := &structs.Project{}
			sanitizedProjectIDs := make([]int64, 0)
			projects = make([]structs.PartialProject, 0, len(user.ProjectIDs))

			for _, projectID := range user.ProjectIDs {
				err := er.Postgres.Model(project).
					Where("id = ?", projectID).
					Select() // Todo: Convert to a single request
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

				_, err := er.Postgres.Model(project).
					WherePK().
					Update()
				if err != nil {
					http.Error(rw, err.Error(), http.StatusInternalServerError)

					return
				}
			}
		} else {
			// If you are not logged in, you cannot have any projects B)
			user = nil
			projects = make([]structs.PartialProject, 0)
		}

		// Reverse projects list
		for i, j := 0, len(projects)-1; i < j; i, j = i+1, j-1 {
			projects[i], projects[j] = projects[j], projects[i]
		}

		passResponse(rw, structs.APIMe{
			Authenticated: auth,
			User:          user,
			Projects:      projects,
		}, true, http.StatusOK)
	}
}

// APIProjectCreateHandler creates a new project and
// returns the project made.
func APIProjectCreateHandler(er *Errorly) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		session, _ := er.Store.Get(r, sessionName)
		defer er.SaveSession(session, r, rw)

		if err := r.ParseForm(); err != nil {
			er.Logger.Error().Err(err).Msg("Failed to parse form")
			passResponse(rw, "Failed to parse form", false, http.StatusBadRequest)

			return
		}

		// Authenticate the user
		auth, user := er.AuthenticateSession(session)
		if !auth {
			passResponse(rw, "You must be logged in to do this", false, http.StatusForbidden)

			return
		}

		projectName := r.PostFormValue("display_name")
		if len(projectName) < 3 {
			passResponse(rw, "Invalid name was passed", false, http.StatusBadRequest)

			return
		}

		projectURL := r.PostFormValue("url")
		if projectURL != "" {
			_, err := url.Parse(projectURL)
			if err != nil {
				passResponse(rw, "Invalid URL was passed", false, http.StatusBadRequest)
			}
		}

		projectPrivate, err := strconv.ParseBool(r.PostFormValue("private"))
		if err != nil {
			projectPrivate = false
		}

		projectLimited, err := strconv.ParseBool(r.PostFormValue("limited"))
		if err != nil {
			projectLimited = false
		}

		userProjects := make([]structs.Project, 0)

		err = er.Postgres.Model(&userProjects).
			Where("created_by_id = ?", user.ID).
			Select()
		if err != nil {
			passResponse(rw, err.Error(), false, http.StatusInternalServerError)

			return
		}

		for _, userProject := range userProjects {
			if userProject.Settings.DisplayName == projectName {
				passResponse(rw, "You cannot have multiple projects with the same name", false, http.StatusBadRequest)

				return
			}
		}

		project := structs.Project{
			ID: er.IDGen.GenerateID(),

			CreatedAt:   time.Now().UTC(),
			CreatedByID: user.ID,

			Integrations: make([]*structs.User, 0),
			Webhooks:     make([]*structs.Webhook, 0),

			Settings: structs.ProjectSettings{
				DisplayName: projectName,

				Description: r.PostFormValue("description"),
				URL:         projectURL,

				Archived: false,
				Private:  projectPrivate,

				Limited: projectLimited,
			},
		}

		_, err = er.Postgres.Model(&project).Insert()
		if err != nil {
			er.Logger.Error().Err(err).Msg("Failed to insert project")
			passResponse(rw, err.Error(), false, http.StatusInternalServerError)

			return
		}

		user.ProjectIDs = append(user.ProjectIDs, project.ID)

		_, err = er.Postgres.Model(user).
			WherePK().
			Update()
		if err != nil {
			er.Logger.Error().Err(err).Msg("Failed to update user projects")
			passResponse(rw, err.Error(), false, http.StatusInternalServerError)

			return
		}

		passResponse(rw, structs.PartialProject{
			ID:             project.ID,
			Name:           project.Settings.DisplayName,
			Description:    project.Settings.Description,
			Archived:       project.Settings.Archived,
			Private:        project.Settings.Private,
			OpenIssues:     project.OpenIssues,
			ActiveIssues:   project.ActiveIssues,
			ResolvedIssues: project.ResolvedIssues,
		}, true, http.StatusOK)
	}
}

// APIProjectHandler returns the initial project information
// and the first page of the project issues.
func APIProjectHandler(er *Errorly) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		session, _ := er.Store.Get(r, sessionName)
		defer er.SaveSession(session, r, rw)

		vars := mux.Vars(r)

		// Authenticate the user
		auth, user := er.AuthenticateSession(session)

		// Retrieve project and user permissions
		project, viewable, elevated, ok := verifyProjectVisibility(er, rw, vars, user, auth, false)
		if !ok {
			// If ok is False, an error has already been provided to the ResponseWriter so we should just return
			return
		}

		if !elevated {
			project.Integrations = make([]*structs.User, 0)
			project.Webhooks = make([]*structs.Webhook, 0)
			project.InviteCodes = make([]*structs.InviteCode, 0)
		}

		if !viewable {
			// No permission to view project. We will treat like the project
			// does not exist.
			passResponse(rw, "Could not find this project", false, http.StatusBadRequest)

			return
		}

		passResponse(rw, structs.APIProject{
			Project:  project,
			Elevated: elevated,
			// Contributors: contributors,
		}, true, http.StatusOK)
	}
}

// APIProjectLazyHandler returns a list of partial users based on the passed user ids query.
func APIProjectLazyHandler(er *Errorly) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		session, _ := er.Store.Get(r, sessionName)
		defer er.SaveSession(session, r, rw)

		contributorIDs := make([]int64, 0)

		query, err := qs.Unmarshal(r.URL.Query().Get("q"))
		if err != nil {
			passResponse(rw, "Invalid query passed", false, http.StatusBadRequest)

			return
		}

		for _, v := range query {
			if id, ok := v.(string); ok {
				if _id, err := strconv.Atoi(id); err == nil {
					contributorIDs = append(contributorIDs, int64(_id))
				} else {
					er.Logger.Error().Err(err).Msgf("Failed to convert ID '%v' to int64", v)
				}
			} else {
				er.Logger.Error().Msgf("Failed to convert '%v' to string", v)
			}
		}

		vars := mux.Vars(r)

		// Authenticate the user
		auth, user := er.AuthenticateSession(session)

		// Retrieve project and user permissions
		project, viewable, _, ok := verifyProjectVisibility(er, rw, vars, user, auth, true)
		if !ok {
			// If ok is False, an error has already been provided to the ResponseWriter so we should just return
			return
		}

		if !viewable {
			// No permission to view project. We will treat like the project
			// does not exist.
			passResponse(rw, "Could not find this project", false, http.StatusBadRequest)

			return
		}

		contributors := make(map[int64]structs.PartialUser)

		for _, contributorID := range contributorIDs {
			_, ok := contributors[contributorID]
			if ok {
				// Do not fetch the user if we already have fetched them
				continue
			}

			if isContributor(project, contributorID) {
				contributor := structs.User{}
				err := er.Postgres.Model(&contributor).
					Where("id = ?", contributorID).
					Select() // Todo: Convert to single request

				if err != nil {
					er.Logger.Error().Err(err).Msg("Failed to retrieve user contributor")
				} else {
					contributors[contributor.ID] = structs.PartialUser{
						ID:          contributor.ID,
						Name:        contributor.Name,
						Avatar:      contributor.Avatar,
						Integration: contributor.Integration,
					}
				}
			}
		}

		passResponse(rw, structs.APIProjectLazy{
			Users: contributors,
			IDs:   contributorIDs,
		}, true, http.StatusOK)
	}
}

// APIProjectExecutorHandler handles executing jobs.
func APIProjectExecutorHandler(er *Errorly) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		session, _ := er.Store.Get(r, sessionName)
		defer er.SaveSession(session, r, rw)

		vars := mux.Vars(r)

		if err := r.ParseForm(); err != nil {
			er.Logger.Error().Err(err).Msg("Failed to parse form")
			passResponse(rw, "Failed to parse form", false, http.StatusBadRequest)

			return
		}

		// Check if an action has been passed
		action, err := structs.ParseActionType(r.FormValue("action"))
		if err != nil {
			passResponse(rw, "Action argument is not valid", false, http.StatusBadRequest)

			return
		}

		_issueIDs, err := qs.Unmarshal(r.FormValue("issues"))
		if err != nil {
			passResponse(rw, "IssueIDs argument is not valid", false, http.StatusBadRequest)

			return
		}

		issueIDs := make([]int64, 0, len(_issueIDs))

		for _, _id := range _issueIDs {
			if _id, ok := _id.(string); ok {
				id, err := strconv.ParseInt(_id, 10, 64)
				if err == nil {
					issueIDs = append(issueIDs, id)
				}
			}
		}

		// Star
		var starring bool
		// Assign
		var assigning bool

		var assigneeID int64
		// LockComments
		var locking bool
		// MarkStatus
		var markType structs.EntryType

		switch action {
		case structs.ActionStar:
			starring, err = strconv.ParseBool(r.FormValue("starring"))
			if err != nil {
				passResponse(rw, "Starring argument is not valid", false, http.StatusBadRequest)

				return
			}
		case structs.ActionAssign:
			assigning, err = strconv.ParseBool(r.FormValue("assigning"))
			if err != nil {
				passResponse(rw, "Assigning argument is not valid", false, http.StatusBadRequest)

				return
			}

			assigneeID, err = strconv.ParseInt(r.FormValue("assignee_id"), 10, 64)
			if err != nil {
				passResponse(rw, "AssigneeID argument is not valid", false, http.StatusBadRequest)

				return
			}
		case structs.ActionLockComments:
			locking, err = strconv.ParseBool(r.FormValue("locking"))
			if err != nil {
				passResponse(rw, "Locking argument is not valid", false, http.StatusBadRequest)

				return
			}
		case structs.ActionMarkStatus:
			markType, err = structs.ParseEntryType(r.FormValue("mark_type"))
			if err != nil {
				passResponse(rw, "MarkType argument is not valid", false, http.StatusBadRequest)

				return
			}
		default:
			passResponse(rw, "Action argument is not valid", false, http.StatusBadRequest)

			return
		}

		// Authenticate the user
		auth, user := er.AuthenticateSession(session)
		if !auth {
			passResponse(rw, "You must be logged in to do this", false, http.StatusForbidden)

			return
		}

		// Retrieve project and user permissions
		project, viewable, elevated, ok := verifyProjectVisibility(er, rw, vars, user, auth, true)
		if !ok {
			// If ok is False, an error has already been provided to the ResponseWriter so we should just return
			return
		}

		if !viewable {
			// No permission to view project. We will treat like the project
			// does not exist.
			passResponse(rw, "Could not find this project", false, http.StatusBadRequest)

			return
		}

		if !elevated {
			// No permission to execute on project. We will simply tell them
			// they cannot do this.
			passResponse(rw, "Guests to a project cannot do this", false, http.StatusForbidden)

			return
		}

		unavailable := make([]int64, 0)
		issues := make([]structs.IssueEntry, 0)

		now := time.Now().UTC()

		for _, issueID := range issueIDs {
			// Fetch the request
			issue := structs.IssueEntry{}
			err = er.Postgres.Model(&issue).
				Where("project_id = ?", project.ID).
				Where("id = ?", issueID).
				Select() // TODO: convert this to a single request

			if err != nil {
				if errors.Is(err, pg.ErrNoRows) {
					// Invalid issue ID
					unavailable = append(unavailable, issueID)

					continue
				}

				passResponse(rw, err.Error(), false, http.StatusInternalServerError)

				return
			}

			// Handle action passed
			switch action {
			case structs.ActionStar:
				change := false

				if issue.Starred {
					project.StarredIssues--

					change = true
				}

				issue.Starred = starring
				if issue.Starred {
					project.StarredIssues++

					change = true
				}

				if change {
					// Update starred issue counter on project
					_, err = er.Postgres.Model(project).
						WherePK().
						Update()
					if err != nil {
						passResponse(rw, err.Error(), false, http.StatusInternalServerError)

						return
					}
				}

			case structs.ActionAssign:
				if assigning {
					issue.AssigneeID = assigneeID
				} else if issue.AssigneeID == assigneeID {
					issue.AssigneeID = 0
				}
			case structs.ActionLockComments:
				issue.CommentsLocked = locking

				// Create comment
				comment := structs.Comment{
					ID:             er.IDGen.GenerateID(),
					IssueID:        issue.ID,
					CreatedAt:      now,
					CreatedByID:    user.ID,
					Type:           structs.CommentsLocked,
					CommentsOpened: &locking,
				}

				_, err = er.Postgres.Model(&comment).Insert()
				if err != nil {
					passResponse(rw, err.Error(), false, http.StatusInternalServerError)

					return
				}
				issue.CommentCount++

			case structs.ActionMarkStatus:
				switch issue.Type {
				case structs.EntryActive:
					project.ActiveIssues--
				case structs.EntryOpen:
					project.OpenIssues--
				case structs.EntryResolved:
					project.ResolvedIssues--
				case structs.EntryInvalid:
				}

				issue.Type = markType

				switch issue.Type {
				case structs.EntryActive:
					project.ActiveIssues++
				case structs.EntryOpen:
					project.OpenIssues++
				case structs.EntryResolved:
					project.ResolvedIssues++
				case structs.EntryInvalid:
				}

				// Update issues cache counter on project
				_, err = er.Postgres.Model(project).
					WherePK().
					Update()
				if err != nil {
					passResponse(rw, err.Error(), false, http.StatusInternalServerError)

					return
				}

				// Create comment
				comment := structs.Comment{
					ID:          er.IDGen.GenerateID(),
					IssueID:     issue.ID,
					CreatedAt:   now,
					CreatedByID: user.ID,
					Type:        structs.IssueMarked,
					IssueMarked: &markType,
				}

				_, err = er.Postgres.Model(&comment).Insert()
				if err != nil {
					passResponse(rw, err.Error(), false, http.StatusInternalServerError)

					return
				}
				issue.CommentCount++
			}

			issue.LastModified = now

			_, err = er.Postgres.Model(&issue).
				WherePK().
				Update()
			if err != nil {
				passResponse(rw, err.Error(), false, http.StatusInternalServerError)

				return
			}

			issues = append(issues, issue)
		}

		passResponse(rw, structs.APIProjectExecutor{
			Issues:      issues,
			Unavailable: unavailable,
			Project:     project,
		}, true, http.StatusOK)
	}
}

// APIProjectContributorsHandler returns a list of partial users from all contributor IDs.
func APIProjectContributorsHandler(er *Errorly) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		session, _ := er.Store.Get(r, sessionName)
		defer er.SaveSession(session, r, rw)

		vars := mux.Vars(r)

		// Authenticate the user
		auth, user := er.AuthenticateSession(session)

		// Retrieve project and user permissions
		project, viewable, _, ok := verifyProjectVisibility(er, rw, vars, user, auth, true)
		if !ok {
			// If ok is False, an error has already been provided to the ResponseWriter so we should just return
			return
		}

		if !viewable {
			// No permission to view project. We will treat like the project
			// does not exist.
			passResponse(rw, "Could not find this project", false, http.StatusBadRequest)

			return
		}

		contributorIDs := make([]int64, 0, len(project.Settings.ContributorIDs))

		for _, contributorID := range project.Settings.ContributorIDs {
			if isContributor(project, contributorID) {
				contributorIDs = append(contributorIDs, contributorID)
			}
		}

		_contributors := []structs.User{}

		if len(contributorIDs) > 0 {
			err := er.Postgres.Model(&_contributors).
				WhereIn("id IN (?)", contributorIDs).
				Select()
			if err != nil {
				passResponse(rw, err.Error(), false, http.StatusInternalServerError)

				return
			}
		}

		contributors := make(map[int64]structs.PartialUser)
		for _, contributor := range _contributors {
			contributors[contributor.ID] = structs.PartialUser{
				ID:          contributor.ID,
				Name:        contributor.Name,
				Avatar:      contributor.Avatar,
				Integration: contributor.Integration,
			}
		}

		// Add owner to contributors if not in it already
		if _, ok := contributors[project.CreatedByID]; !ok {
			contributor := structs.User{}

			err := er.Postgres.Model(&contributor).
				Where("id = ?", project.CreatedByID).
				Select()
			if err != nil {
				er.Logger.Error().Err(err).Msg("Failed to retrieve user contributor")
			} else {
				contributors[contributor.ID] = structs.PartialUser{
					ID:          contributor.ID,
					Name:        contributor.Name,
					Avatar:      contributor.Avatar,
					Integration: contributor.Integration,
				}
			}
		}

		passResponse(rw, structs.APIProjectLazy{
			Users: contributors,
			IDs:   contributorIDs,
		}, true, http.StatusOK)
	}
}

// APIProjectContributorsRemoveHandler handles removing a contributor.
func APIProjectContributorsRemoveHandler(er *Errorly) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		session, _ := er.Store.Get(r, sessionName)
		defer er.SaveSession(session, r, rw)

		vars := mux.Vars(r)

		contributorID, ok := vars["contributor"]
		if !ok {
			passResponse(rw, "Missing ContributorID", false, http.StatusBadRequest)

			return
		}

		if err := r.ParseForm(); err != nil {
			er.Logger.Error().Err(err).Msg("Failed to parse form")
			passResponse(rw, "Failed to parse form", false, http.StatusBadRequest)

			return
		}

		userID, err := strconv.ParseInt(contributorID, 10, 64)
		if err != nil {
			passResponse(rw, "ContributorID argument is not valid", false, http.StatusBadRequest)

			return
		}

		// Authenticate the user
		auth, user := er.AuthenticateSession(session)

		// Retrieve project and user permissions
		project, viewable, _, ok := verifyProjectVisibility(er, rw, vars, user, auth, true)
		if !ok {
			// If ok is False, an error has already been provided to the ResponseWriter so we should just return
			return
		}

		if !viewable {
			// No permission to view project. We will treat like the project
			// does not exist.
			passResponse(rw, "Could not find this project", false, http.StatusBadRequest)

			return
		}

		if userID == project.CreatedByID {
			passResponse(rw, "You cannot remove this user", false, http.StatusForbidden)

			return
		}

		contributorIDs := make([]int64, 0, len(project.Settings.ContributorIDs))

		for _, contributorID := range project.Settings.ContributorIDs {
			if contributorID != userID && isContributor(project, contributorID) {
				contributorIDs = append(contributorIDs, contributorID)
			}
		}

		project.Settings.ContributorIDs = contributorIDs

		_, err = er.Postgres.Model(project).
			WherePK().
			Update()
		if err != nil {
			passResponse(rw, err.Error(), false, http.StatusInternalServerError)

			return
		}

		// Remove project from old contributors project list
		contributorUser := structs.User{}
		err = er.Postgres.Model(&contributorUser).
			Where("id = ?", userID).
			Select()
		if err != nil {
			er.Logger.Error().Err(err).Msg("Failed to retrieve user contributor")
		} else {
			contributorUserProjectList := make([]int64, 0)
			for _, projectID := range contributorUser.ProjectIDs {
				if projectID != project.ID {
					contributorUserProjectList = append(contributorUserProjectList, projectID)
				}
			}

			contributorUser.ProjectIDs = contributorUserProjectList

			_, err = er.Postgres.Model(&contributorUser).
				WherePK().
				Update()
			if err != nil {
				passResponse(rw, err.Error(), false, http.StatusInternalServerError)

				return
			}
		}

		_contributors := []structs.User{}

		if len(contributorIDs) > 0 {
			err := er.Postgres.Model(&_contributors).
				WhereIn("id IN (?)", contributorIDs).
				Select()
			if err != nil {
				passResponse(rw, err.Error(), false, http.StatusInternalServerError)

				return
			}
		}

		contributors := make(map[int64]structs.PartialUser)
		for _, contributor := range _contributors {
			contributors[contributor.ID] = structs.PartialUser{
				ID:          contributor.ID,
				Name:        contributor.Name,
				Avatar:      contributor.Avatar,
				Integration: contributor.Integration,
			}
		}

		// Add owner to contributors if not in it already
		if _, ok := contributors[project.CreatedByID]; !ok {
			contributor := structs.User{}

			err := er.Postgres.Model(&contributor).
				Where("id = ?", project.CreatedByID).
				Select()
			if err != nil {
				er.Logger.Error().Err(err).Msg("Failed to retrieve user contributor")
			} else {
				contributors[contributor.ID] = structs.PartialUser{
					ID:          contributor.ID,
					Name:        contributor.Name,
					Avatar:      contributor.Avatar,
					Integration: contributor.Integration,
				}
			}
		}

		passResponse(rw, structs.APIProjectLazy{
			Users: contributors,
			IDs:   contributorIDs,
		}, true, http.StatusOK)
	}
}

// APIProjectUpdateHandler handles updating the project settings.
func APIProjectUpdateHandler(er *Errorly) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		session, _ := er.Store.Get(r, sessionName)
		defer er.SaveSession(session, r, rw)

		vars := mux.Vars(r)

		if err := r.ParseForm(); err != nil {
			er.Logger.Error().Err(err).Msg("Failed to parse form")
			passResponse(rw, "Failed to parse form", false, http.StatusBadRequest)

			return
		}

		// Authenticate the user
		auth, user := er.AuthenticateSession(session)
		if !auth {
			passResponse(rw, "You must be logged in to do this", false, http.StatusForbidden)

			return
		}

		// Retrieve project and user permissions
		project, viewable, elevated, ok := verifyProjectVisibility(er, rw, vars, user, auth, true)
		if !ok {
			// If ok is False, an error has already been provided to the ResponseWriter so we should just return
			return
		}

		if !elevated {
			passResponse(rw, "You do not have permission to do this", false, http.StatusForbidden)

			return
		}

		if !viewable {
			// No permission to view project. We will treat like the project
			// does not exist.
			passResponse(rw, "Could not find this project", false, http.StatusBadRequest)

			return
		}

		if _displayName := r.FormValue("display_name"); _displayName != "" {
			_displayName = strings.TrimSpace(_displayName)
			if len(_displayName) > 3 && _displayName != project.Settings.DisplayName {
				// ensure there is no other project with the same display name
				userProjects := make([]structs.Project, 0)

				err := er.Postgres.Model(&userProjects).
					Where("created_by_id = ?", user.ID).
					Select()
				if err != nil {
					passResponse(rw, err.Error(), false, http.StatusInternalServerError)

					return
				}

				for _, userProject := range userProjects {
					if userProject.Settings.DisplayName == _displayName {
						passResponse(rw, "You cannot have multiple projects with the same name",
							false, http.StatusBadRequest)

						return
					}
				}

				project.Settings.DisplayName = _displayName
			}
		}

		if _description := r.FormValue("description"); _description != "" {
			project.Settings.Description = _description
		}

		if _url := r.FormValue("url"); _url != "" {
			_, err := url.ParseRequestURI(_url)
			if err == nil {
				project.Settings.URL = _url
			}
		}

		if _archived, err := strconv.ParseBool(r.FormValue("archived")); err == nil {
			project.Settings.Archived = _archived
		}

		if _private, err := strconv.ParseBool(r.FormValue("private")); err == nil {
			project.Settings.Private = _private
		}

		if _limited, err := strconv.ParseBool(r.FormValue("limited")); err == nil {
			project.Settings.Limited = _limited
		}

		_contributorIDs := []int64{}
		if err := json.UnmarshalFromString(r.FormValue("contributors"), &_contributorIDs); err != nil {
			project.Settings.ContributorIDs = _contributorIDs
		}

		_, err := er.Postgres.Model(project).
			WherePK().
			Update()
		if err != nil {
			passResponse(rw, err.Error(), false, http.StatusInternalServerError)

			return
		}

		passResponse(rw, structs.APIProjectUpdate{
			Settings: project.Settings,
		}, true, http.StatusOK)
	}
}

// APIProjectDeleteHandler handles deleting a project. Only the project creator can do this.
func APIProjectDeleteHandler(er *Errorly) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		session, _ := er.Store.Get(r, sessionName)
		defer er.SaveSession(session, r, rw)

		vars := mux.Vars(r)

		if err := r.ParseForm(); err != nil {
			er.Logger.Error().Err(err).Msg("Failed to parse form")
			passResponse(rw, "Failed to parse form", false, http.StatusBadRequest)

			return
		}

		// Authenticate the user
		auth, user := er.AuthenticateSession(session)
		if !auth {
			passResponse(rw, "You must be logged in to do this", false, http.StatusForbidden)

			return
		}

		// Retrieve project and user permissions
		project, viewable, elevated, ok := verifyProjectVisibility(er, rw, vars, user, auth, true)
		if !ok {
			// If ok is False, an error has already been provided to the ResponseWriter so we should just return
			return
		}

		if !elevated {
			passResponse(rw, "You do not have permission to do this", false, http.StatusForbidden)

			return
		}

		if !viewable {
			// No permission to view project. We will treat like the project
			// does not exist.
			passResponse(rw, "Could not find this project", false, http.StatusBadRequest)

			return
		}

		// Only the user who created the project is allowed to delete it
		if project.CreatedByID != user.ID {
			passResponse(rw, "You must be the project owner to delete it", false, http.StatusUnauthorized)

			return
		}

		if project.Settings.DisplayName != r.FormValue("confirm") {
			passResponse(rw, "Invalid confirm passed", false, http.StatusBadRequest)

			return
		}

		results, err := er.Postgres.Model(project).
			WherePK().
			Delete()
		if err != nil {
			passResponse(rw, err.Error(), false, http.StatusInternalServerError)

			return
		}

		println("Removed", results.RowsAffected(), "project entries")

		results, err = er.Postgres.Model(&structs.User{}).
			Where("project_id = ?", project.ID).
			Delete()
		if err != nil {
			passResponse(rw, err.Error(), false, http.StatusInternalServerError)

			return
		}

		println("Removed", results.RowsAffected(), "user entries")

		results, err = er.Postgres.Model(&structs.Webhook{}).
			Where("project_id = ?", project.ID).
			Delete()
		if err != nil {
			passResponse(rw, err.Error(), false, http.StatusInternalServerError)

			return
		}

		println("Removed", results.RowsAffected(), "webhook entries")

		issues := make([]structs.IssueEntry, 0)

		err = er.Postgres.Model(&issues).
			Where("project_id = ?", project.ID).
			Select()
		if err != nil {
			passResponse(rw, err.Error(), false, http.StatusInternalServerError)

			return
		}

		results, err = er.Postgres.Model(&structs.IssueEntry{}).
			Where("project_id = ?", project.ID).
			Delete()
		if err != nil {
			passResponse(rw, err.Error(), false, http.StatusInternalServerError)

			return
		}

		println("Removed", results.RowsAffected(), "issue entries")

		affected := 0

		for _, issue := range issues {
			results, err = er.Postgres.Model(&structs.Comment{}).
				Where("issue_id = ?", issue.ID).
				Delete()
			if err != nil {
				passResponse(rw, err.Error(), false, http.StatusInternalServerError)

				return
			}

			affected += results.RowsAffected()
		}

		println("Removed", affected, "comment entries")

		// Remove project from user's project list
		_projectIDs := make([]int64, 0, len(user.ProjectIDs))

		for _, projectID := range user.ProjectIDs {
			if projectID != project.ID {
				_projectIDs = append(_projectIDs, projectID)
			}
		}

		user.ProjectIDs = _projectIDs

		_, err = er.Postgres.Model(user).
			WherePK().
			Update()
		if err != nil {
			passResponse(rw, err.Error(), false, http.StatusInternalServerError)

			return
		}

		passResponse(rw, "Project was deleted", true, http.StatusOK)
	}
}

// APIProjectTransferHandler handles transferring a project to another user.
func APIProjectTransferHandler(er *Errorly) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		session, _ := er.Store.Get(r, sessionName)
		defer er.SaveSession(session, r, rw)

		vars := mux.Vars(r)

		if err := r.ParseForm(); err != nil {
			er.Logger.Error().Err(err).Msg("Failed to parse form")
			passResponse(rw, "Failed to parse form", false, http.StatusBadRequest)

			return
		}

		userID, err := strconv.ParseInt(r.FormValue("contributor_id"), 10, 64)
		if err != nil {
			passResponse(rw, "ContributorID argument is not valid", false, http.StatusBadRequest)

			return
		}

		// Authenticate the user
		auth, user := er.AuthenticateSession(session)

		// Retrieve project and user permissions
		project, viewable, _, ok := verifyProjectVisibility(er, rw, vars, user, auth, true)
		if !ok {
			// If ok is False, an error has already been provided to the ResponseWriter so we should just return
			return
		}

		if !viewable {
			// No permission to view project. We will treat like the project
			// does not exist.
			passResponse(rw, "Could not find this project", false, http.StatusBadRequest)

			return
		}

		if user.ID != project.CreatedByID {
			passResponse(rw, "You need to own the project to transfer", false, http.StatusForbidden)

			return
		}

		newUser := structs.User{}

		err = er.Postgres.Model(&newUser).
			Where("id = ?", userID).
			Select()
		if err != nil {
			if xerrors.Is(err, pg.ErrNoRows) {
				passResponse(rw, "Cannot find user you are trying to transfer to", false, http.StatusBadRequest)

				return
			}

			passResponse(rw, err.Error(), false, http.StatusInternalServerError)

			return
		}

		if newUser.Name != r.FormValue("confirm") {
			passResponse(rw, "Invalid confirm passed", false, http.StatusBadRequest)

			return
		}

		// If we transfer, we will set the created by id to the transferred user
		// then add the previous owner to the contributors list.

		contributorIDs := make([]int64, 0, len(project.Settings.ContributorIDs))
		isPreviousContributor := false

		for _, contributorID := range project.Settings.ContributorIDs {
			if isContributor(project, contributorID) && contributorID != userID {
				contributorIDs = append(contributorIDs, contributorID)
			}

			if contributorID == userID {
				isPreviousContributor = true
			}
		}

		if !isPreviousContributor {
			// The user it is transferred to has to have already been on the contributors list
			passResponse(rw, "You cannot transfer to this user", false, http.StatusBadRequest)

			return
		}

		contributorIDs = append(contributorIDs, project.CreatedByID)

		project.CreatedByID = userID
		project.Settings.ContributorIDs = contributorIDs

		_, err = er.Postgres.Model(project).
			WherePK().
			Update()
		if err != nil {
			passResponse(rw, err.Error(), false, http.StatusInternalServerError)

			return
		}

		passResponse(rw, "Project owner transferred", true, http.StatusOK)
	}
}

// APIProjectIssueHandler returns paginated results.
func APIProjectIssueHandler(er *Errorly) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		session, _ := er.Store.Get(r, sessionName)
		defer er.SaveSession(session, r, rw)

		vars := mux.Vars(r)

		urlQuery := r.URL.Query()

		// Authenticate the user
		auth, user := er.AuthenticateSession(session)

		project, viewable, _, ok := verifyProjectVisibility(er, rw, vars, user, auth, true)
		if !ok {
			// If ok is False, an error has already been provided to the ResponseWriter so we should just return
			return
		}

		if !viewable {
			// No permission to view project. We will treat like the project
			// does not exist.
			passResponse(rw, "Could not find this project", false, http.StatusBadRequest)

			return
		}

		_issueID := urlQuery.Get("id")
		if _issueID != "" {
			// If an ID is specified we will instead return just the issue
			issueID, err := strconv.ParseInt(_issueID, 10, 64)
			if err != nil {
				passResponse(rw, "ID argument is not valid", false, http.StatusBadRequest)

				return
			}

			issue := &structs.IssueEntry{}

			err = er.Postgres.Model(issue).
				Where("issue_entry.project_id = ?", project.ID).
				Where("issue_entry.id = ?", issueID).
				Select()
			if err != nil {
				if errors.Is(err, pg.ErrNoRows) {
					// Invalid issue ID
					passResponse(rw, "Could not find this issue", false, http.StatusBadRequest)

					return
				}

				passResponse(rw, err.Error(), false, http.StatusInternalServerError)

				return
			}

			passResponse(rw, structs.APIProjectIssues{
				Issue: issue,
			}, true, http.StatusOK)

			return
		}

		// Get query from search
		query := urlQuery.Get("q")
		if query == "" {
			query = ""
		}

		// We should get a limit argument here but at the moment
		// we will hardcode 25 per page.
		_pageLimit := pageLimit

		// Retrieve page argument from URL
		_page := r.FormValue("page")
		if _page == "" {
			passResponse(rw, "Page argument is missing", false, http.StatusBadRequest)

			return
		}

		// Check page is a valid number. We will use the first page
		// argument provided as multiple could be passed.
		page, err := strconv.Atoi(_page)
		if err != nil {
			passResponse(rw, "Page argument is not valid", false, http.StatusBadRequest)

			return
		}

		var issues []structs.IssueEntry

		var totalissues int

		if auth {
			issues, totalissues, err = fetchProjectIssues(er, project.ID, _pageLimit, page, query, user.ID)
		} else {
			issues, totalissues, err = fetchProjectIssues(er, project.ID, _pageLimit, page, query, 0)
		}

		if err != nil {
			passResponse(rw, err.Error(), false, http.StatusInternalServerError)

			return
		}

		passResponse(rw, structs.APIProjectIssues{
			Page:        page,
			TotalIssues: totalissues,
			Issues:      issues,
		}, true, http.StatusOK)
	}
}

// APIProjectIssueCreateHandler creates a new issue or increments an already made issue.
func APIProjectIssueCreateHandler(er *Errorly) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		session, _ := er.Store.Get(r, sessionName)
		defer er.SaveSession(session, r, rw)

		vars := mux.Vars(r)

		if err := r.ParseForm(); err != nil {
			er.Logger.Error().Err(err).Msg("Failed to parse form")
			passResponse(rw, "Failed to parse form", false, http.StatusBadRequest)

			return
		}

		// Authenticate the user
		auth, user := er.AuthenticateSession(session)
		if !auth {
			passResponse(rw, "You must be logged in to do this", false, http.StatusForbidden)

			return
		}

		// Retrieve project and user permissions
		project, viewable, elevated, ok := verifyProjectVisibility(er, rw, vars, user, auth, true)
		if !ok {
			// If ok is False, an error has already been provided to the ResponseWriter so we should just return
			return
		}

		if !elevated {
			passResponse(rw, "You do not have permission to do this", false, http.StatusForbidden)

			return
		}

		if !viewable {
			// No permission to view project. We will treat like the project
			// does not exist.
			passResponse(rw, "Could not find this project", false, http.StatusBadRequest)

			return
		}

		if project.Settings.Archived {
			// If the project is Archived, new issues cannot be made
			passResponse(rw, "This project is archived", false, http.StatusForbidden)

			return
		}

		// Retrieve issue error
		issueError := r.FormValue("error")
		if issueError == "" {
			passResponse(rw, "Error is missing", false, http.StatusBadRequest)

			return
		}

		// Retrieve issue function
		issueFunction := r.FormValue("function")
		if issueFunction == "" {
			passResponse(rw, "Function is missing", false, http.StatusBadRequest)

			return
		}

		now := time.Now().UTC()
		issue := &structs.IssueEntry{}
		newIssue := false

		err := er.Postgres.Model(issue).
			Where("error = ?", issueError).
			Where("function = ?", issueFunction).
			Select()
		if err != nil {
			if !errors.Is(err, pg.ErrNoRows) {
				// Unexpected error
				passResponse(rw, err.Error(), false, http.StatusInternalServerError)

				return
			}

			assigneeID, err := strconv.ParseInt(r.FormValue("assigned"), 10, 64)
			if err != nil {
				assigneeID = 0
			}

			isContributor := false
			if assigneeID == project.CreatedByID {
				isContributor = true
			} else {
				for _, contributorID := range project.Settings.ContributorIDs {
					if contributorID == assigneeID {
						isContributor = true
					}
				}
			}

			if !isContributor {
				assigneeID = 0
			}

			commentsLocked, err := strconv.ParseBool(r.FormValue("lock_comments"))
			if err != nil {
				commentsLocked = false
			}

			newIssue = true
			issue = &structs.IssueEntry{
				ID:             er.IDGen.GenerateID(),
				ProjectID:      project.ID,
				Starred:        false,
				Type:           structs.EntryOpen,
				Occurrences:    1,
				AssigneeID:     assigneeID,
				Error:          issueError,
				Function:       issueFunction,
				Checkpoint:     r.FormValue("checkpoint"),
				Description:    r.FormValue("description"),
				Traceback:      r.FormValue("traceback"),
				LastModified:   now,
				CreatedAt:      now,
				CreatedByID:    user.ID,
				CommentCount:   0,
				CommentsLocked: commentsLocked,
			}

			_, err = er.Postgres.Model(issue).Insert()
			if err != nil {
				passResponse(rw, err.Error(), false, http.StatusInternalServerError)

				return
			}

			switch issue.Type {
			case structs.EntryActive:
				project.ActiveIssues++
			case structs.EntryOpen:
				project.OpenIssues++
			case structs.EntryResolved:
				project.ResolvedIssues++
			case structs.EntryInvalid:
			}

			// Update issues cache counter on project
			_, err = er.Postgres.Model(project).
				WherePK().
				Update()
			if err != nil {
				passResponse(rw, err.Error(), false, http.StatusInternalServerError)

				return
			}
		} else {
			// An error with this function and error already exists, increment it again
			issue.Occurrences++
			issue.LastModified = now

			// We will overwrite the assignee and lock comments if the creator is the same person
			if user.ID == issue.CreatedByID {
				assigneeID, err := strconv.ParseInt(r.FormValue("assigned"), 10, 64)
				if err != nil {
					assigneeID = 0
				}

				isContributor := false
				if assigneeID == project.CreatedByID {
					isContributor = true
				} else {
					for _, contributorID := range project.Settings.ContributorIDs {
						if contributorID == assigneeID {
							isContributor = true
						}
					}
				}
				if isContributor {
					issue.AssigneeID = assigneeID
				}

				commentsLocked, err := strconv.ParseBool(r.FormValue("lock_comments"))
				if err == nil {
					issue.CommentsLocked = commentsLocked
				}
			}

			_, err = er.Postgres.Model(issue).
				WherePK().
				Update()
			if err != nil {
				// Error updating query
				passResponse(rw, err.Error(), false, http.StatusInternalServerError)

				return
			}
		}

		passResponse(rw, structs.APIProjectIssueCreate{
			New:   newIssue,
			Issue: issue,
		}, true, http.StatusOK)
	}
}

// APIProjectFetchIssueHandler returns an issue from a project.
func APIProjectFetchIssueHandler(er *Errorly) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		session, _ := er.Store.Get(r, sessionName)
		defer er.SaveSession(session, r, rw)

		vars := mux.Vars(r)

		_issueID, ok := vars["issue_id"]
		if !ok {
			passResponse(rw, "Missing Issue ID", false, http.StatusBadRequest)

			return
		}

		// Authenticate the user
		auth, user := er.AuthenticateSession(session)

		project, viewable, _, ok := verifyProjectVisibility(er, rw, vars, user, auth, true)
		if !ok {
			// If ok is False, an error has already been provided to the ResponseWriter so we should just return
			return
		}

		if !viewable {
			// No permission to view project. We will treat like the project
			// does not exist.
			passResponse(rw, "Could not find this project", false, http.StatusBadRequest)

			return
		}

		// If an ID is specified we will instead return just the issue
		issueID, err := strconv.ParseInt(_issueID, 10, 64)
		if err != nil {
			passResponse(rw, "ID argument is not valid", false, http.StatusBadRequest)

			return
		}

		issue := &structs.IssueEntry{}

		err = er.Postgres.Model(issue).
			Where("issue_entry.project_id = ?", project.ID).
			Where("issue_entry.id = ?", issueID).
			Select()
		if err != nil {
			if errors.Is(err, pg.ErrNoRows) {
				// Invalid issue ID
				passResponse(rw, "Could not find this issue", false, http.StatusBadRequest)

				return
			}

			passResponse(rw, err.Error(), false, http.StatusInternalServerError)

			return
		}

		passResponse(rw, structs.APIProjectIssues{
			Issue: issue,
		}, true, http.StatusOK)
	}
}

// APIProjectIssueDeleteHandler handles deleting an issue.
func APIProjectIssueDeleteHandler(er *Errorly) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		session, _ := er.Store.Get(r, sessionName)
		defer er.SaveSession(session, r, rw)

		vars := mux.Vars(r)

		_issueID, ok := vars["issue_id"]
		if !ok {
			passResponse(rw, "Missing Issue ID", false, http.StatusBadRequest)

			return
		}

		// Authenticate the user
		auth, user := er.AuthenticateSession(session)

		project, viewable, elevated, ok := verifyProjectVisibility(er, rw, vars, user, auth, true)
		if !ok {
			// If ok is False, an error has already been provided to the ResponseWriter so we should just return
			return
		}

		if !viewable {
			// No permission to view project. We will treat like the project
			// does not exist.
			passResponse(rw, "Could not find this project", false, http.StatusBadRequest)

			return
		}

		// If an ID is specified we will instead return just the issue
		issueID, err := strconv.ParseInt(_issueID, 10, 64)
		if err != nil {
			passResponse(rw, "ID argument is not valid", false, http.StatusBadRequest)

			return
		}

		issue := &structs.IssueEntry{}

		err = er.Postgres.Model(issue).
			Where("issue_entry.project_id = ?", project.ID).
			Where("issue_entry.id = ?", issueID).
			Select()
		if err != nil {
			if errors.Is(err, pg.ErrNoRows) {
				// Invalid issue ID
				passResponse(rw, "Could not find this issue", false, http.StatusBadRequest)

				return
			}

			passResponse(rw, err.Error(), false, http.StatusInternalServerError)

			return
		}

		if !elevated || issue.CreatedByID != user.ID {
			passResponse(rw, "You must be elevated or the issue creator to delete it", false, http.StatusForbidden)

			return
		}

		_, err = er.Postgres.Model(issue).
			WherePK().
			Delete()
		if err != nil {
			passResponse(rw, err.Error(), false, http.StatusInternalServerError)

			return
		}

		switch issue.Type {
		case structs.EntryActive:
			project.ActiveIssues--
		case structs.EntryOpen:
			project.OpenIssues--
		case structs.EntryResolved:
			project.ResolvedIssues--
		case structs.EntryInvalid:
		}

		// Update issues cache counter on project
		_, err = er.Postgres.Model(project).
			WherePK().
			Update()
		if err != nil {
			passResponse(rw, err.Error(), false, http.StatusInternalServerError)

			return
		}

		_, err = er.Postgres.Model(&structs.Comment{}).
			Where("issue_id = ?", issue.ID).
			Delete()
		if err != nil {
			passResponse(rw, err.Error(), false, http.StatusInternalServerError)

			return
		}

		passResponse(rw, "Issue was deleted", true, http.StatusOK)
	}
}

// APIProjectIssueCommentHandler returns a list of comments from an issue and returns paginated results.
func APIProjectIssueCommentHandler(er *Errorly) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		session, _ := er.Store.Get(r, sessionName)
		defer er.SaveSession(session, r, rw)

		vars := mux.Vars(r)

		_issueID, ok := vars["issue_id"]
		if !ok {
			passResponse(rw, "Missing Issue ID", false, http.StatusBadRequest)

			return
		}

		// Authenticate the user
		auth, user := er.AuthenticateSession(session)

		project, viewable, _, ok := verifyProjectVisibility(er, rw, vars, user, auth, true)
		if !ok {
			// If ok is False, an error has already been provided to the ResponseWriter so we should just return
			return
		}

		if !viewable {
			// No permission to view project. We will treat like the project
			// does not exist.
			passResponse(rw, "Could not find this project", false, http.StatusBadRequest)

			return
		}

		// If an ID is specified we will instead return just the issue
		issueID, err := strconv.ParseInt(_issueID, 10, 64)
		if err != nil {
			passResponse(rw, "ID argument is not valid", false, http.StatusBadRequest)

			return
		}

		issue := &structs.IssueEntry{}

		err = er.Postgres.Model(issue).
			Where("issue_entry.project_id = ?", project.ID).
			Where("issue_entry.id = ?", issueID).
			Select()
		if err != nil {
			if errors.Is(err, pg.ErrNoRows) {
				// Invalid issue ID
				passResponse(rw, "Could not find this issue", false, http.StatusBadRequest)

				return
			}

			passResponse(rw, err.Error(), false, http.StatusInternalServerError)

			return
		}

		// We will use the same page limit for the issues per query limit
		_issueLimit := pageLimit

		// Retrieve page argument from URL
		_page := r.URL.Query().Get("page")
		if _page == "" {
			_page = "0"
		}

		// Check page is a valid number. We will use the first page
		// argument provided as multiple could be passed.
		page, err := strconv.Atoi(_page)
		if err != nil {
			passResponse(rw, "Page argument is not valid", false, http.StatusBadRequest)

			return
		}

		comments := make([]structs.Comment, 0, _issueLimit)

		err = er.Postgres.Model(&comments).
			Limit(_issueLimit).
			Where("issue_id = ?", issue.ID).
			Offset(int(math.Max(0, float64(_issueLimit*page)))).
			Select()
		if err != nil {
			passResponse(rw, err.Error(), false, http.StatusInternalServerError)

			return
		}

		passResponse(rw, structs.APIProjectIssueComments{
			Page:     page,
			Comments: comments,
			End:      len(comments) < _issueLimit,
		}, true, http.StatusOK)
	}
}

// APIProjectIssueCommentCreateHandler handles the creation of issue comments.
func APIProjectIssueCommentCreateHandler(er *Errorly) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		session, _ := er.Store.Get(r, sessionName)
		defer er.SaveSession(session, r, rw)

		vars := mux.Vars(r)

		_issueID, ok := vars["issue_id"]
		if !ok {
			passResponse(rw, "Missing Issue ID", false, http.StatusBadRequest)

			return
		}

		if err := r.ParseForm(); err != nil {
			er.Logger.Error().Err(err).Msg("Failed to parse form")
			passResponse(rw, "Failed to parse form", false, http.StatusBadRequest)

			return
		}

		// Authenticate the user
		auth, user := er.AuthenticateSession(session)
		if !auth {
			passResponse(rw, "You must be logged in to do this", false, http.StatusForbidden)

			return
		}

		project, viewable, elevated, ok := verifyProjectVisibility(er, rw, vars, user, auth, true)
		if !ok {
			// If ok is False, an error has already been provided to the ResponseWriter so we should just return
			return
		}

		if !viewable {
			// No permission to view project. We will treat like the project
			// does not exist.
			passResponse(rw, "Could not find this project", false, http.StatusBadRequest)

			return
		}

		if project.Settings.Limited && !elevated {
			passResponse(rw, "You need to be a contributor of this project to create issues", false, http.StatusForbidden)

			return
		}

		if project.Settings.Archived {
			passResponse(rw, "This project is archived", false, http.StatusForbidden)

			return
		}

		// If an ID is specified we will instead return just the issue
		issueID, err := strconv.ParseInt(_issueID, 10, 64)
		if err != nil {
			passResponse(rw, "ID argument is not valid", false, http.StatusBadRequest)

			return
		}

		issue := &structs.IssueEntry{}

		err = er.Postgres.Model(issue).
			Where("issue_entry.project_id = ?", project.ID).
			Where("issue_entry.id = ?", issueID).
			Select()
		if err != nil {
			if errors.Is(err, pg.ErrNoRows) {
				// Invalid issue ID
				passResponse(rw, "Could not find this issue", false, http.StatusBadRequest)

				return
			}

			passResponse(rw, err.Error(), false, http.StatusInternalServerError)

			return
		}

		if issue.CommentsLocked {
			passResponse(rw, "Comments are locked for this issue", false, http.StatusForbidden)

			return
		}

		content := strings.TrimSpace(r.PostFormValue("content"))
		if len(content) == 0 {
			passResponse(rw, "Invalid message content was passed", false, http.StatusBadRequest)

			return
		}

		now := time.Now().UTC()
		comment := &structs.Comment{
			ID:          er.IDGen.GenerateID(),
			IssueID:     issue.ID,
			CreatedAt:   now,
			CreatedByID: user.ID,
			Type:        structs.Message,
			Content:     &content,
		}

		_, err = er.Postgres.Model(comment).Insert()
		if err != nil {
			passResponse(rw, err.Error(), false, http.StatusInternalServerError)

			return
		}

		passResponse(rw, comment, true, http.StatusOK)
	}
}

// APIProjectInviteGetHandler handles retrieving an invite.
func APIProjectInviteGetHandler(er *Errorly) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		session, _ := er.Store.Get(r, sessionName)
		defer er.SaveSession(session, r, rw)

		vars := mux.Vars(r)

		inviteCode, ok := vars["join_code"]
		if !ok {
			passResponse(rw, "No invite code supplied", false, http.StatusBadRequest)

			return
		}

		// Authenticate the user
		auth, user := er.AuthenticateSession(session)
		if !auth {
			passResponse(rw, "You must be logged in to do this", false, http.StatusForbidden)

			return
		}

		project, _, _, ok := verifyProjectVisibility(er, rw, vars, user, auth, true)
		if !ok {
			// If ok is False, an error has already been provided to the ResponseWriter so we should just return
			return
		}

		invite := &structs.InviteCode{}

		err := er.Postgres.Model(invite).
			Where("code = ?", inviteCode).
			Where("project_id = ?", project.ID).
			Select()
		if err != nil {
			if errors.Is(err, pg.ErrNoRows) {
				passResponse(rw, "Invalid invite code passed", false, http.StatusBadRequest)

				return
			}

			// Unexpected error
			passResponse(rw, err.Error(), false, http.StatusInternalServerError)

			return
		}

		passResponse(rw, structs.APIProjectInviteGet{
			ValidInvite: true,
			ProjectName: project.Settings.DisplayName,
		}, true, http.StatusOK)
	}
}

// APIProjectInviteUseHandler handles a user joining an invite.
func APIProjectInviteUseHandler(er *Errorly) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		session, _ := er.Store.Get(r, sessionName)
		defer er.SaveSession(session, r, rw)

		vars := mux.Vars(r)

		inviteCode, ok := vars["join_code"]
		if !ok {
			passResponse(rw, "No invite code supplied", false, http.StatusBadRequest)
			return
		}

		// Authenticate the user
		auth, user := er.AuthenticateSession(session)
		if !auth {
			passResponse(rw, "You must be logged in to do this", false, http.StatusForbidden)
			return
		}

		project, _, _, ok := verifyProjectVisibility(er, rw, vars, user, auth, true)
		if !ok {
			// If ok is False, an error has already been provided to the ResponseWriter so we should just return
			return
		}

		if project.CreatedByID == user.ID {
			passResponse(rw, "You are already in this project", false, http.StatusBadRequest)

			return
		}

		for _, contributor := range project.Settings.ContributorIDs {
			if contributor == user.ID {
				passResponse(rw, "You are already in this project", false, http.StatusBadRequest)

				return
			}
		}

		invite := &structs.InviteCode{}

		err := er.Postgres.Model(invite).
			Where("project_id = ?", project.ID).
			Where("code = ?", inviteCode).
			Select()
		if err != nil {
			if errors.Is(err, pg.ErrNoRows) {
				passResponse(rw, "Invalid invite code passed", false, http.StatusBadRequest)

				return
			}

			// Unexpected error
			passResponse(rw, err.Error(), false, http.StatusInternalServerError)

			return
		}

		// Check if the invite has enough uses or has expired, if so delete it.
		if (invite.MaxUses > 0 && invite.Uses >= invite.MaxUses) ||
			(!invite.ExpiresBy.IsZero() && time.Now().After(invite.ExpiresBy)) {
			_, err = er.Postgres.Model(invite).WherePK().Delete()
			if err != nil && !errors.Is(err, pg.ErrNoRows) {
				// Unexpected error
				passResponse(rw, err.Error(), false, http.StatusInternalServerError)

				return
			}

			passResponse(rw, "Invalid invite code passed", false, http.StatusBadRequest)

			return
		}

		// Increment invite uses
		invite.Uses++

		// Add user to contributors
		project.Settings.ContributorIDs = append(project.Settings.ContributorIDs, user.ID)

		// Add project to users project list
		user.ProjectIDs = append(user.ProjectIDs, project.ID)

		_, err = er.Postgres.Model(project).WherePK().Update()
		if err != nil && !errors.Is(err, pg.ErrNoRows) {
			// Unexpected error
			passResponse(rw, err.Error(), false, http.StatusInternalServerError)

			return
		}

		_, err = er.Postgres.Model(invite).WherePK().Update()
		if err != nil && !errors.Is(err, pg.ErrNoRows) {
			// Unexpected error
			passResponse(rw, err.Error(), false, http.StatusInternalServerError)

			return
		}

		_, err = er.Postgres.Model(user).WherePK().Update()
		if err != nil && !errors.Is(err, pg.ErrNoRows) {
			// Unexpected error
			passResponse(rw, err.Error(), false, http.StatusInternalServerError)

			return
		}

		passResponse(rw, "OK", true, http.StatusOK)
	}
}

// APIProjectInviteDeleteHandler handles deleting an invite.
func APIProjectInviteDeleteHandler(er *Errorly) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		session, _ := er.Store.Get(r, sessionName)
		defer er.SaveSession(session, r, rw)

		vars := mux.Vars(r)

		inviteCode, ok := vars["join_code"]
		if !ok {
			passResponse(rw, "No invite code supplied", false, http.StatusBadRequest)

			return
		}

		// Authenticate the user
		auth, user := er.AuthenticateSession(session)
		if !auth {
			passResponse(rw, "You must be logged in to do this", false, http.StatusForbidden)

			return
		}

		project, viewable, elevated, ok := verifyProjectVisibility(er, rw, vars, user, auth, true)
		if !ok {
			// If ok is False, an error has already been provided to the ResponseWriter so we should just return
			return
		}

		if !viewable {
			// No permission to view project. We will treat like the project
			// does not exist.
			passResponse(rw, "Could not find this project", false, http.StatusBadRequest)

			return
		}

		if !elevated {
			// No permission to execute on project. We will simply tell them
			// they cannot do this.
			passResponse(rw, "Guests to a project cannot do this", false, http.StatusForbidden)

			return
		}

		invite := &structs.InviteCode{}

		res, err := er.Postgres.Model(invite).
			Where("code = ?", inviteCode).
			Where("project_id = ?", project.ID).
			Delete()
		if err != nil {
			if errors.Is(err, pg.ErrNoRows) {
				passResponse(rw, "Invalid invite code passed", false, http.StatusBadRequest)

				return
			}

			// Unexpected error
			passResponse(rw, err.Error(), false, http.StatusInternalServerError)

			return
		}

		if res.RowsAffected() == 0 {
			passResponse(rw, "Invalid invite code passed", false, http.StatusBadRequest)

			return
		}

		passResponse(rw, "OK", true, http.StatusOK)
	}
}

// APIProjectInviteCreateHandler handles creating an invite.
func APIProjectInviteCreateHandler(er *Errorly) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		session, _ := er.Store.Get(r, sessionName)
		defer er.SaveSession(session, r, rw)

		vars := mux.Vars(r)

		// Authenticate the user
		auth, user := er.AuthenticateSession(session)
		if !auth {
			passResponse(rw, "You must be logged in to do this", false, http.StatusForbidden)

			return
		}

		project, viewable, elevated, ok := verifyProjectVisibility(er, rw, vars, user, auth, true)
		if !ok {
			// If ok is False, an error has already been provided to the ResponseWriter so we should just return
			return
		}

		if !viewable {
			// No permission to view project. We will treat like the project
			// does not exist.
			passResponse(rw, "Could not find this project", false, http.StatusBadRequest)

			return
		}

		if !elevated {
			// No permission to execute on project. We will simply tell them
			// they cannot do this.
			passResponse(rw, "Guests to a project cannot do this", false, http.StatusForbidden)

			return
		}

		expiration, err := strconv.ParseInt(r.FormValue("expiration"), 10, 64)
		if err != nil {
			passResponse(rw, "Expiration argument is not valid", false, http.StatusBadRequest)

			return
		}

		uses, err := strconv.ParseInt(r.FormValue("uses"), 10, 64)
		if err != nil {
			passResponse(rw, "Expiration argument is not valid", false, http.StatusBadRequest)

			return
		}

		if expiration < 0 || uses < 0 {
			passResponse(rw, "Invalid argument for either expiration or uses", false, http.StatusBadRequest)

			return
		}

		now := time.Now()

		var expiresBy time.Time

		if expiration > 0 {
			expiresBy = now.Add(time.Duration(expiration) * time.Minute)
		} else {
			expiresBy = time.Time{}
		}

		id := er.IDGen.GenerateID()

		invite := structs.InviteCode{
			ID:   id,
			Code: idFromUInt64(uint64(id)),

			CreatedAt:   now,
			CreatedByID: user.ID,

			Uses:    0,
			MaxUses: uses,

			ProjectID: project.ID,
			ExpiresBy: expiresBy,
		}

		_, err = er.Postgres.Model(&invite).Insert()
		if err != nil {
			er.Logger.Error().Err(err).Msg("Failed to insert invite")
			passResponse(rw, err.Error(), false, http.StatusInternalServerError)

			return
		}

		passResponse(rw, invite, true, http.StatusOK)
	}
}

// func Handler(er *Errorly) http.HandlerFunc {
// 	return
// }
