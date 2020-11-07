package errorly

import (
	"time"

	idgenerator "github.com/TheRockettek/Errorly-Web/pkg/idgenerator"
	"github.com/TheRockettek/Errorly-Web/structs"
	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
)

// startEpoch for IDs (12/10/2020 13:01:14)
const epoch = 1602507674941

// type dbLogger struct{}

// func (d dbLogger) BeforeQuery(c context.Context, q *pg.QueryEvent) (context.Context, error) {
// 	return c, nil
// }

// func (d dbLogger) AfterQuery(c context.Context, q *pg.QueryEvent) error {
// 	b, _ := q.FormattedQuery()
// 	println(string(b))
// 	return nil
// }

// createSchema creates database schema
func createSchema(db *pg.DB) (err error) {
	// db.AddQueryHook(dbLogger{})

	refresh := false

	models := []interface{}{
		&structs.User{},
		&structs.Project{},
		&structs.Webhook{},
		&structs.IssueEntry{},
		&structs.Comment{},
	}

	for _, model := range models {
		if refresh {
			db.Model(model).DropTable(&orm.DropTableOptions{
				IfExists: true,
				Cascade:  true,
			})
		}

		err = db.Model(model).CreateTable(&orm.CreateTableOptions{
			IfNotExists: true,
		})
		if err != nil {
			panic(err)
		}
	}

	if refresh {
		idGen := idgenerator.NewIDGenerator(1602507674941, 0)

		println("Create User")
		user := &structs.User{
			ID:          idGen.GenerateID(),
			Name:        "ImRock",
			Avatar:      "https://cdn.discordapp.com/avatars/143090142360371200/a_70444022ea3e5d73dd00d59c5578b07e.gif?size=1024",
			UserType:    structs.DiscordUser,
			HookID:      143090142360371200,
			ProjectIDs:  make([]int64, 0),
			Integration: false,
		}
		_, err = db.Model(user).Insert()
		if err != nil {
			panic(err)
		}

		println("Create second user")
		user2 := &structs.User{
			ID:          idGen.GenerateID(),
			Name:        "biscuitcord",
			Avatar:      "https://cdn.discordapp.com/avatars/164297154276360192/4c8f9b0310948cce460613081d074a13.webp?size=1024",
			UserType:    structs.DiscordUser,
			HookID:      164297154276360192,
			ProjectIDs:  make([]int64, 0),
			Integration: false,
		}
		_, err = db.Model(user2).Insert()
		if err != nil {
			panic(err)
		}

		println("Create Project")
		project := &structs.Project{
			ID: idGen.GenerateID(),

			CreatedAt:   time.Now().UTC(),
			CreatedByID: user.ID,

			Integrations: make([]*structs.User, 0),
			Webhooks:     make([]*structs.Webhook, 0),

			Settings: structs.ProjectSettings{
				DisplayName: "Welcomer",
				URL:         "https://welcomer.gg",
				Archived:    false,
				Private:     false,
				Limited:     false,
			},

			StarredIssues:  0,
			OpenIssues:     0,
			ActiveIssues:   0,
			ResolvedIssues: 0,
		}
		_, err = db.Model(project).Insert()
		if err != nil {
			return err
		}

		println("Add project to user")
		user.ProjectIDs = append(user.ProjectIDs, project.ID)
		_, err = db.Model(user).WherePK().Update()
		if err != nil {
			panic(err)
		}

		println("Add project to user 2")
		user2.ProjectIDs = append(user2.ProjectIDs, project.ID)
		_, err = db.Model(user).WherePK().Update()
		if err != nil {
			panic(err)
		}

		println("Add second user to contributors")
		project.Settings.ContributorIDs = append(project.Settings.ContributorIDs, user2.ID)
		_, err = db.Model(project).WherePK().Update()
		if err != nil {
			panic(err)
		}

		println("Create Webhooks")
		webhook := &structs.Webhook{
			ID:          idGen.GenerateID(),
			ProjectID:   project.ID,
			Active:      false,
			Failures:    16,
			CreatedAt:   time.Now().UTC(),
			CreatedByID: user.ID,
			URL:         "https://welcomer.gg/webhook",
			Type:        structs.DiscordWebhook,
			JSONContent: true,
			Secret:      "",
		}
		_, err = db.Model(webhook).Insert()
		if err != nil {
			panic(err)
		}

		println("Create Integration")
		integration := &structs.User{
			ID:   idGen.GenerateID(),
			Name: "Welcomer",

			UserType: structs.IntegrationUser,

			CreatedAt: time.Now().UTC(),

			ProjectID:   project.ID,
			Integration: true,
			CreatedByID: user.ID,
		}
		_, err = db.Model(integration).Insert()
		if err != nil {
			panic(err)
		}

		println("Create user issue")
		now := time.Now().UTC()
		issue := &structs.IssueEntry{
			ID:        idGen.GenerateID(),
			ProjectID: project.ID,

			Starred: false,

			Type:        structs.EntryOpen,
			Occurrences: 1,
			AssigneeID:  0,

			Error:       "genericError",
			Function:    "createSchema(db *pg.DB)",
			Checkpoint:  "internal/api.go:147",
			Description: "",
			Traceback:   "",

			LastModified: now,

			CreatedAt:   now,
			CreatedByID: user.ID,

			CommentCount:   0,
			CommentsLocked: false,
		}
		_, err = db.Model(issue).Insert()
		if err != nil {
			panic(err)
		}

		println("Increment project issue counter")
		project.OpenIssues++
		_, err = db.Model(project).WherePK().Update()
		if err != nil {
			panic(err)
		}

		println("Create user issue 2")
		now = time.Now().UTC()
		issue2 := &structs.IssueEntry{
			ID:        idGen.GenerateID(),
			ProjectID: project.ID,

			Starred: false,

			Type:        structs.EntryOpen,
			Occurrences: 5,
			AssigneeID:  user.ID,

			Error:       "panic:",
			Function:    "main.main.func1",
			Checkpoint:  "main.go:11",
			Description: "",
			Traceback:   "stacktrace from panic: \ngoroutine 1 [running]:\nruntime/debug.Stack(0x1042ff18, 0x98b2, 0xf0ba0, 0x17d048)\n    /usr/local/go/src/runtime/debug/stack.go:24 +0xc0\nmain.main.func1()\n    /tmp/sandbox973508195/main.go:11 +0x60\npanic(0xf0ba0, 0x17d048)\n    /usr/local/go/src/runtime/panic.go:502 +0x2c0\nmain.main()\n    /tmp/sandbox973508195/main.go:16 +0x60",

			LastModified: now,

			CreatedAt:   now,
			CreatedByID: user2.ID,

			CommentCount:   0,
			CommentsLocked: false,
		}
		_, err = db.Model(issue2).Insert()
		if err != nil {
			panic(err)
		}

		println("Increment project issue counter")
		project.OpenIssues++
		_, err = db.Model(project).WherePK().Update()
		if err != nil {
			panic(err)
		}

		println("Create integration issue")
		now = time.Now().UTC()
		issue3 := &structs.IssueEntry{
			ID:        idGen.GenerateID(),
			ProjectID: project.ID,

			Starred: false,

			Type:        structs.EntryOpen,
			Occurrences: 1,
			AssigneeID:  user2.ID,

			Error:       "TypeError",
			Function:    "",
			Checkpoint:  "",
			Description: "can only concatenate str (not \"int\") to str",
			Traceback:   "Traceback (most recent call last):\n  File \"<stdin>\", line 1, in <module>\n  File \"<stdin>\", line 2, in a\nTypeError: can only concatenate str (not \"int\") to str",

			LastModified: now,

			CreatedAt:   now,
			CreatedByID: integration.ID,

			CommentCount:   0,
			CommentsLocked: false,
		}
		_, err = db.Model(issue3).Insert()
		if err != nil {
			panic(err)
		}

		println("Increment project issue counter")
		project.OpenIssues++
		_, err = db.Model(project).WherePK().Update()
		if err != nil {
			panic(err)
		}

		println("Create user issue comment")
		content := "Test :)"
		comment := &structs.Comment{
			ID:      idGen.GenerateID(),
			IssueID: issue.ID,

			CreatedAt:   time.Now().UTC(),
			CreatedByID: user2.ID,

			Type:    structs.Message,
			Content: &content,
		}
		_, err = db.Model(comment).Insert()
		if err != nil {
			panic(err)
		}
		issue.CommentCount++

		println("Create user issue comment2")
		open := structs.EntryOpen
		comment2 := &structs.Comment{
			ID:      idGen.GenerateID(),
			IssueID: issue.ID,

			CreatedAt:   time.Now().UTC(),
			CreatedByID: user2.ID,

			Type:        structs.IssueMarked,
			IssueMarked: &open,
		}
		_, err = db.Model(comment2).Insert()
		if err != nil {
			panic(err)
		}
		issue.CommentCount++

		println("Create user issue comment3")
		opened := true
		comment3 := &structs.Comment{
			ID:      idGen.GenerateID(),
			IssueID: issue.ID,

			CreatedAt:   time.Now().UTC(),
			CreatedByID: user.ID,

			Type:           structs.CommentsLocked,
			CommentsOpened: &opened,
		}
		_, err = db.Model(comment3).Insert()
		if err != nil {
			panic(err)
		}
		issue.CommentCount++

		println("Update issue comment count")
		_, err = db.Model(issue).WherePK().Update()
		if err != nil {
			panic(err)
		}

		println("Create user 2 issue comment")
		comment4 := &structs.Comment{
			ID:      idGen.GenerateID(),
			IssueID: issue2.ID,

			CreatedAt:   time.Now().UTC(),
			CreatedByID: user.ID,

			Type:           structs.CommentsLocked,
			CommentsOpened: &opened,
		}
		_, err = db.Model(comment4).Insert()
		if err != nil {
			panic(err)
		}

		println("Update issue2 comment count")
		issue2.CommentCount++
		_, err = db.Model(issue2).WherePK().Update()
		if err != nil {
			panic(err)
		}

		println("Star user issue")
		issue2.Starred = true
		_, err = db.Model(issue2).WherePK().Update()
		if err != nil {
			panic(err)
		}

		println("Update project stars")
		project.StarredIssues++
		_, err = db.Model(project).WherePK().Update()
		if err != nil {
			panic(err)
		}

		println("Close integration issue")
		issue3.Type = structs.EntryResolved
		issue3.LastModified = time.Now().UTC()
		_, err = db.Model(issue3).WherePK().Update()
		if err != nil {
			panic(err)
		}

		println("Update project issue counter")
		project.ResolvedIssues++
		project.OpenIssues--
		_, err = db.Model(project).WherePK().Update()
		if err != nil {
			panic(err)
		}

		println("Create close integration issue comment")
		resolved := structs.EntryResolved
		comment5 := &structs.Comment{
			ID:      idGen.GenerateID(),
			IssueID: issue3.ID,

			CreatedAt:   time.Now().UTC(),
			CreatedByID: user.ID,

			Type:        structs.IssueMarked,
			IssueMarked: &resolved,
		}
		_, err = db.Model(comment5).Insert()
		if err != nil {
			panic(err)
		}
	}

	return nil
}
