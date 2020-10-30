package errorly

import (
	"encoding/csv"
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
	"github.com/hashicorp/go-uuid"
)

// Issues per page
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
		rw.Write(resp)
	} else {
		http.Error(rw, string(resp), status)
	}
	return
}

func verifyProjectVisibility(er *Errorly, rw http.ResponseWriter, vars map[string]string, user *structs.User, auth bool) (project structs.Project, viewable bool, elevated bool, ok bool) {
	// Retrieve project_id from /project/{project_id}...
	_projectID, ok := vars["project_id"]

	if !ok {
		ok = false
		passResponse(rw, "Missing Project ID", false, http.StatusBadRequest)
		return
	}

	// Check projectID is a valid number
	projectID, err := strconv.ParseInt(_projectID, 10, 64)
	if err != nil {
		ok = false
		passResponse(rw, "Could not find this project", false, http.StatusBadRequest)
		return
	}
	// Now we have a possibly valid ID we will first get the project
	project = structs.Project{}
	err = er.Postgres.Model(&project).Where("project.id = ?", projectID).Relation("Webhooks").Relation("Integrations").Select()
	if err != nil {
		if err == pg.ErrNoRows {
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

func fetchProjectIssues(er *Errorly, projectID int64, limit int, page int, query string, userID int64) (issues []structs.IssueEntry, totalissues int, err error) {
	// sort:created_by-desc
	_issues := make([]structs.IssueEntry, 0, limit)

	initialQuery := er.Postgres.Model(&_issues).Where("issue_entry.project_id = ?", projectID)

	query = strings.ReplaceAll(query, `'`, `"`)
	r := csv.NewReader(strings.NewReader(query))
	r.Comma = ' '

	parts, err := r.Read()
	if err != nil {
		return
	}

	fetchStarred := false
	fuzzyEntries := make([]string, 0)
	for _, part := range parts {
		subpart := strings.Split(part, ":")
		if len(subpart) < 2 {
			fuzzyEntries = append(fuzzyEntries, subpart[0])
		} else {
			finger, thumb := subpart[0], subpart[1]
			switch finger {
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
					initialQuery = initialQuery.Where("type = ?", structs.EntryActive)
				case "open":
					initialQuery = initialQuery.Where("type = ?", structs.EntryOpen)
				case "invalid":
					initialQuery = initialQuery.Where("type = ?", structs.EntryInvalid)
				case "resolved":
					initialQuery = initialQuery.Where("type = ?", structs.EntryResolved)
				case "starred":
					fetchStarred = true
				}
			case "assignee":
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
				// new querys
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
			}
		}
	}

	// if fetching stars
	// get starres and page
	// else
	// get starres and page
	// get total issues

	totalissues = 0
	issues = make([]structs.IssueEntry, 0, limit)

	// Fetch issues
	count, err := initialQuery.Clone().Limit(limit).Where("starred = ?", true).Offset(int(math.Max(0, float64(limit*page)))).SelectAndCount()
	if err != nil {
		return
	}
	totalissues += count
	if len(fuzzyEntries) > 0 {
		// for _, issue := range _issues {
		// 	for _, fuzz := range fuzzyEntries {
		// 		if fuzzy.Match(fuzz, issue.Error) {
		// 			issues = append(issues, issue)
		// 			break
		// 		}
		// 		if fuzzy.Match(fuzz, issue.Description) {
		// 			issues = append(issues, issue)
		// 			break
		// 		}
		// 	}
		// }
		issues = append(issues, _issues...)
	} else {
		issues = append(issues, _issues...)
	}

	if !fetchStarred {
		// If we are not retrieving stars, we will retrieve any starred issues before normal issues
		// if there are 15 starred issues, we will retrieve at least 10 regular ones and if there are
		// 25+ starred issues, we do not retrieve any regular issues.
		if len(issues) < limit {
			count, err = initialQuery.Clone().Limit(limit - len(issues)).Where("starred is NULL").Offset(int(math.Max(0, float64((limit*page)-count)))).SelectAndCount()
			if err != nil {
				return
			}
			totalissues += count
			if len(fuzzyEntries) > 0 {
				// for _, issue := range _issues {
				// 	for _, fuzz := range fuzzyEntries {
				// 		if fuzzy.Match(fuzz, issue.Error) {
				// 			issues = append(issues, issue)
				// 			break
				// 		}
				// 		if fuzzy.Match(fuzz, issue.Description) {
				// 			issues = append(issues, issue)
				// 			break
				// 		}
				// 	}
				// }
				issues = append(issues, _issues...)
			} else {
				issues = append(issues, _issues...)
			}
		} else {
			count, err = initialQuery.Clone().Limit(limit - len(issues)).Where("starred is NULL").Offset(int(math.Max(0, float64((limit*page)-count)))).Count()
			if err != nil {
				return
			}
			totalissues += count
		}
	}

	// totalissues, err = initialQuery.Limit(limit).Offset(limit * page).SelectAndCount()

	// this is sort:starred-desc sort:type-desc sort:created_at-desc sort:occurrences-desc
	// totalissues, err = er.Postgres.Model(&issues).Where("issue_entry.project_id = ?", projectID).Relation("CreatedBy").Relation("Assignee").Order("starred DESC").Order("type DESC").Order("created_at DESC").Order("occurrences DESC").Limit(limit).Offset(limit * page).SelectAndCount()

	return
}

// LogoutHandler handles clearing a user session
func LogoutHandler(er *Errorly) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		session, _ := er.Store.Get(r, sessionName)
		defer session.Save(r, rw)

		session.Values = make(map[interface{}]interface{})
		http.Redirect(rw, r, "/", http.StatusTemporaryRedirect)
	}
}

// LoginHandler handles CSRF and AuthCode redirection
func LoginHandler(er *Errorly) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		session, _ := er.Store.Get(r, sessionName)
		defer session.Save(r, rw)

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

// OAuthCallbackHandler handles authenticating discord OAuth and creating
// a user profile if necessary
func OAuthCallbackHandler(er *Errorly) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		session, _ := er.Store.Get(r, sessionName)
		defer session.Save(r, rw)

		// Validate the CSRF in the session and in the HTTP request.
		// If there is no CSRF in the session it is likely our fault :)
		_csrfString := r.URL.Query().Get("state")
		csrfString, ok := session.Values["oauth_csrf"].(string)
		if !ok {
			http.Error(rw, "Missing CSRF state", http.StatusInternalServerError)
			return
		}

		if _csrfString != csrfString {
			http.Error(rw, "Mismatched CSRF states", http.StatusUnauthorized)
		}

		// Just to be sure, remove the CSRF after we have compared the CSRF
		delete(session.Values, "oauth_csrf")

		// Create an OAuth exchange with the code we were given.
		code := r.URL.Query().Get("code")
		token, err := er.Configuration.OAuth.Exchange(er.ctx, code)
		if err != nil {
			http.Error(rw, "Failed to exchange code: "+err.Error(), http.StatusInternalServerError)
		}

		// Create a client with our exchanged token and retrieve a user.
		client := er.Configuration.OAuth.Client(er.ctx, token)
		resp, err := client.Get(discordUsersMe)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		discordUserResponse := &DiscordUser{}
		err = json.Unmarshal(body, &discordUserResponse)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
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
					http.Error(rw, err.Error(), http.StatusInternalServerError)
				}
				session.Values["token"] = token
			} else {
				http.Error(rw, err.Error(), http.StatusInternalServerError)
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
				http.Error(rw, err.Error(), http.StatusInternalServerError)
			}
			session.Values["token"] = token
		}

		// Once the user has logged in, send them back to the home page.
		http.Redirect(rw, r, "/", http.StatusTemporaryRedirect)
	}
}

// APIMeHandler handles the /api/me request which returns the user
// object and a list of partial project objects
func APIMeHandler(er *Errorly) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		session, _ := er.Store.Get(r, sessionName)
		defer session.Save(r, rw)

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
					http.Error(rw, err.Error(), http.StatusInternalServerError)
					return
				}
			}
		} else {
			// If you are not logged in, you cannot have any projects B)
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
		defer session.Save(r, rw)

		// vars, err := parseJSONForm(r)
		// if err != nil {
		// 	er.Logger.Error().Err(err).Msg("Failed to parse json form")
		// 	rw.WriteHeader(http.StatusBadRequest)
		// 	return
		// }
		if err := r.ParseForm(); err != nil {
			er.Logger.Error().Err(err).Msg("Failed to parse form")
			rw.WriteHeader(http.StatusBadRequest)
			return
		}

		// Authenticate the user
		auth, user := er.AuthenticateSession(session)
		if !auth {
			passResponse(rw, "You must be signed into create a project", false, http.StatusForbidden)
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
		err = er.Postgres.Model(&userProjects).Where("created_by_id = ?", user.ID).Select()

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

			Integrations: make([]structs.User, 0),
			Webhooks:     make([]structs.Webhook, 0),

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
		_, err = er.Postgres.Model(&user).WherePK().Update()
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
		defer session.Save(r, rw)

		vars := mux.Vars(r)

		// Authenticate the user
		auth, user := er.AuthenticateSession(session)

		// Retrieve project and user permissions
		project, viewable, elevated, ok := verifyProjectVisibility(er, rw, vars, &user, auth)
		if !ok {
			// If ok is False, an error has already been provided to the ResponseWriter so we should just return
			return
		}

		if !elevated {
			project.Integrations = make([]structs.User, 0)
			project.Webhooks = make([]structs.Webhook, 0)
		}

		if !viewable {
			// No permission to view project. We will treat like the project
			// does not exist.
			passResponse(rw, "Could not find this project", false, http.StatusBadRequest)
			return
		}

		// contributors := make(map[int64]structs.PartialUser)
		// for _, contributorID := range project.Settings.ContributorIDs {
		// 	contributor := structs.User{}
		// 	err := er.Postgres.Model(&contributor).Where("id = ?", contributorID).Select()
		// 	if err != nil {
		// 		er.Logger.Error().Err(err).Msg("Failed to retrieve user contributor")
		// 	} else {
		// 		contributors[contributor.ID] = structs.PartialUser{
		// 			ID:     contributor.ID,
		// 			Name:   contributor.Name,
		// 			Avatar: contributor.Avatar,
		// 		}
		// 	}
		// }

		// contributors[project.CreatedBy.ID] = structs.PartialUser{
		// 	ID:     project.CreatedBy.ID,
		// 	Name:   project.CreatedBy.Name,
		// 	Avatar: project.CreatedBy.Avatar,
		// }

		passResponse(rw, structs.APIProject{
			Project: project,
			// Contributors: contributors,
		}, true, http.StatusOK)
	}
}

// This simply returns a true or false if the id is from a contributor or owner
func isContributor(project structs.Project, id int64) bool {
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

// APIProjectContributorsHandler returns a list of partial users from all contributor ids
func APIProjectContributorsHandler(er *Errorly) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		session, _ := er.Store.Get(r, sessionName)
		defer session.Save(r, rw)

		vars := mux.Vars(r)

		// Authenticate the user
		auth, user := er.AuthenticateSession(session)

		// Retrieve project and user permissions
		project, viewable, _, ok := verifyProjectVisibility(er, rw, vars, &user, auth)
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
		contributors := make(map[int64]structs.PartialUser)

		for _, contributorID := range project.Settings.ContributorIDs {
			_, ok := contributors[contributorID]
			if ok {
				// Do not fetch the user if we already have fetched them
				continue
			}

			if isContributor(project, contributorID) {
				contributor := structs.User{}
				err := er.Postgres.Model(&contributor).Where("id = ?", contributorID).Select()
				if err != nil {
					er.Logger.Error().Err(err).Msg("Failed to retrieve user contributor")
				} else {
					contributors[contributor.ID] = structs.PartialUser{
						ID:          contributor.ID,
						Name:        contributor.Name,
						Integration: contributor.Integration,
					}
				}
			}
		}

		// Add owner to contributors if not in it already
		if _, ok := contributors[project.CreatedByID]; !ok {
			contributor := structs.User{}
			err := er.Postgres.Model(&contributor).Where("id = ?", project.CreatedByID).Select()
			if err != nil {
				er.Logger.Error().Err(err).Msg("Failed to retrieve user contributor")
			} else {
				contributors[contributor.ID] = structs.PartialUser{
					ID:          contributor.ID,
					Name:        contributor.Name,
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

// APIProjectLazyHandler returns a list of partial users based on the passed user ids query
func APIProjectLazyHandler(er *Errorly) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		session, _ := er.Store.Get(r, sessionName)
		defer session.Save(r, rw)

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
		project, viewable, _, ok := verifyProjectVisibility(er, rw, vars, &user, auth)
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
				err := er.Postgres.Model(&contributor).Where("id = ?", contributorID).Select()
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

// APIProjectExecutorHandler handles executing jobs
func APIProjectExecutorHandler(er *Errorly) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		session, _ := er.Store.Get(r, sessionName)
		defer session.Save(r, rw)

		vars := mux.Vars(r)

		if err := r.ParseForm(); err != nil {
			er.Logger.Error().Err(err).Msg("Failed to parse form")
			rw.WriteHeader(http.StatusBadRequest)
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
			passResponse(rw, "You must be signed in to do this", false, http.StatusForbidden)
			return
		}

		// Retrieve project and user permissions
		project, viewable, elevated, ok := verifyProjectVisibility(er, rw, vars, &user, auth)
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
		}

		unavailable := make([]int64, 0)
		issues := make([]structs.IssueEntry, 0)

		now := time.Now().UTC()

		for _, issueID := range issueIDs {
			// Fetch the request
			issue := structs.IssueEntry{}
			err = er.Postgres.Model(&issue).Where("project_id = ?", project.ID).Where("id = ?", issueID).Select()
			if err != nil {
				if err == pg.ErrNoRows {
					// Invalid issue ID
					unavailable = append(unavailable, issueID)
					continue
					// passResponse(rw, "Could not find this issue", false, http.StatusBadRequest)
					// return
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
					_, err = er.Postgres.Model(&project).WherePK().Update()
					if err != nil {
						passResponse(rw, err.Error(), false, http.StatusInternalServerError)
						return
					}
				}

			case structs.ActionAssign:
				if assigning {
					issue.AssigneeID = assigneeID
				} else {
					if issue.AssigneeID == assigneeID {
						issue.AssigneeID = 0
					}
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
					CommentsOpened: locking,
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
				}

				issue.Type = markType

				switch issue.Type {
				case structs.EntryActive:
					project.ActiveIssues++
				case structs.EntryOpen:
					project.OpenIssues++
				case structs.EntryResolved:
					project.ResolvedIssues++
				}

				// Update issues cache counter on project
				_, err = er.Postgres.Model(&project).WherePK().Update()
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
					IssueMarked: markType,
				}
				_, err = er.Postgres.Model(&comment).Insert()
				if err != nil {
					passResponse(rw, err.Error(), false, http.StatusInternalServerError)
					return
				}
				issue.CommentCount++
			}

			issue.LastModified = now
			_, err = er.Postgres.Model(&issue).WherePK().Update()
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

// APIProjectIssueHandler returns paginated results
func APIProjectIssueHandler(er *Errorly) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		session, _ := er.Store.Get(r, sessionName)
		defer session.Save(r, rw)

		vars := mux.Vars(r)

		urlQuery := r.URL.Query()

		// Get query from search
		query := urlQuery.Get("q")
		if query == "" {
			query = "sort:created_by-asc sort:starred-asc"
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

		// Authenticate the user
		auth, user := er.AuthenticateSession(session)

		project, viewable, _, ok := verifyProjectVisibility(er, rw, vars, &user, auth)
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

		issues, totalissues, err := fetchProjectIssues(er, project.ID, _pageLimit, page, query, user.ID)
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

// APIProjectIssueCreateHandler creates a new issue or increments an already made issue
func APIProjectIssueCreateHandler(er *Errorly) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		session, _ := er.Store.Get(r, sessionName)
		defer session.Save(r, rw)

		vars := mux.Vars(r)

		if err := r.ParseForm(); err != nil {
			er.Logger.Error().Err(err).Msg("Failed to parse form")
			rw.WriteHeader(http.StatusBadRequest)
			return
		}

		// Authenticate the user
		auth, user := er.AuthenticateSession(session)
		if !auth {
			passResponse(rw, "You must be signed into create a project", false, http.StatusForbidden)
			return
		}

		// Retrieve project and user permissions
		project, viewable, elevated, ok := verifyProjectVisibility(er, rw, vars, &user, auth)
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
		issue := structs.IssueEntry{}
		newIssue := false
		err := er.Postgres.Model(&issue).Where("error = ?", issueError).Where("function = ?", issueFunction).Select()
		if err != nil {
			if err != pg.ErrNoRows {
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
			issue = structs.IssueEntry{
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

			_, err = er.Postgres.Model(&issue).Insert()
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

			_, err = er.Postgres.Model(&issue).WherePK().Update()
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

// func Handler(er *Errorly) http.HandlerFunc {
// 	return
// }
