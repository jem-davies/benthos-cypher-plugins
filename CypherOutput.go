package cypher

import (
	"context"

	"github.com/benthosdev/benthos/v4/public/service"
	"github.com/gocarina/gocsv"
	"github.com/neo4j/neo4j-go-driver/neo4j"
)

type Neo4j struct {
	Database string
	Uri      string
	User     string
	Password string
	Driver   neo4j.Driver
	Session  neo4j.Session
}

type subjectObjectRelationCsv struct {
	Subject  string `csv:"Subject"` // struct tags are required for gocsv
	Relation string `csv:"Relation"`
	Object   string `csv:"Object"`
}

var getNeoDriver = neo4j.NewDriver

func init() {
	// Register our new output with benthos.
	configSpec := service.NewConfigSpec().
		Description("This output processor inserts Subject-Object-Relations into Neo4j.").
		Field(service.NewInterpolatedStringField("Database")).
		Field(service.NewInterpolatedStringField("Uri")).
		Field(service.NewInterpolatedStringField("User")).
		Field(service.NewInterpolatedStringField("Password"))

	constructor := func(conf *service.ParsedConfig, mgr *service.Resources) (out service.Output, maxInFlight int, err error) {
		database, _ := conf.FieldString("Database")
		uri, _ := conf.FieldString("Uri")
		user, _ := conf.FieldString("User")
		password, _ := conf.FieldString("Password")

		return &Neo4j{Database: database, Uri: uri, User: user, Password: password}, 0, nil
	}

	err := service.RegisterOutput("cypher", configSpec, constructor)
	if err != nil {
		panic(err)
	}
}

func (neo *Neo4j) Connect(ctx context.Context) error {

	driver, err := getNeoDriver(neo.Uri, neo4j.NoAuth(), func(c *neo4j.Config) { c.Encrypted = false })
	if err != nil {
		return err
	}

	neo.Driver = driver

	session, err := driver.NewSession(neo4j.SessionConfig{
		AccessMode:   neo4j.AccessModeWrite,
		Bookmarks:    []string{},
		DatabaseName: neo.Database,
	})
	if err != nil {
		return err
	}

	neo.Session = session

	return nil
}

func (neo *Neo4j) Write(ctx context.Context, msg *service.Message) error {
	content, err := msg.AsStructuredMut()
	if err != nil {
		return err
	}

	collateTriples := content.(map[string]interface{})["SOR"].(string)

	SORs := []*subjectObjectRelationCsv{}
	gocsv.UnmarshalString(collateTriples, &SORs)

	for _, SOR := range SORs {
		_, err = neo.gdb_create_relation(SOR.Subject, SOR.Object, SOR.Relation)
	}

	return nil
}

func (neo *Neo4j) Close(ctx context.Context) error {
	neo.Driver.Close()
	neo.Session.Close()
	return nil
}

func (neo *Neo4j) gdb_create_relation(subject_name string, target_name string, relation_type string) (any, error) {

	_, err := neo.Session.WriteTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		result, err := tx.Run("MATCH n,m WHERE n.name = '"+subject_name+"' AND m.name = '"+target_name+"' MERGE (n)-[l:"+relation_type+"]->(m)", nil)
		if err != nil {
			return nil, err
		}

		return result.Consume()
	})

	return nil, err
}
