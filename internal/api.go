package errorly

import (
	"github.com/TheRockettek/Errorly-Web/structs"
	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
)

// startEpoch for IDs (12/10/2020 13:01:14)
const epoch = 1602507674941

// createSchema creates database schema
func createSchema(db *pg.DB) (err error) {
	models := []interface{}{
		(*structs.User)(nil),
		(*structs.Project)(nil),
		(*structs.Integration)(nil),
		(*structs.Webhook)(nil),
		(*structs.IssueEntry)(nil),
		(*structs.Comment)(nil),
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
	// 	Occurrences: 1,

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
