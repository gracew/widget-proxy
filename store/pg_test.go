package store

import (
	"os"
	"testing"

	"github.com/go-pg/pg"
	"github.com/google/uuid"
	"github.com/gracew/widget-proxy/generated"
	"github.com/gracew/widget-proxy/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type PgTestSuite struct {
	suite.Suite
	s PgStore
}

var db *pg.DB

func (suite *PgTestSuite) SetupTest() {
	port := os.Getenv("PG_PORT")
	if port == "" {
		port = "5432"
	}
	db = pg.Connect(&pg.Options{User: "postgres", Password: "postgres", Addr: "localhost:" + port})
	suite.s = PgStore{
		DB: db,
		API: model.API{
			Operations: &model.OperationDefinition{
				Update: &model.UpdateDefinition{
					Actions: []model.ActionDefinition{
						model.ActionDefinition{Name: "action", Fields: []string{"Test"}},
					},
				},
			},
		},
	}
	err := suite.s.CreateSchema()
	assert.NoError(suite.T(), err)
}

func (suite *PgTestSuite) TearDownTest() {
	db.Close()
}

func (suite *PgTestSuite) TestCreateGet() {
	obj := &generated.Object{Test: "test", CreatedBy: "userID"}
	createRes, err := suite.s.CreateObject(obj)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), obj.Test, createRes.Test)
	assert.Equal(suite.T(), "userID", createRes.CreatedBy)
	assert.NotEmpty(suite.T(), obj.ID)
	assert.NotEmpty(suite.T(), obj.CreatedAt)

	getRes, err := suite.s.GetObject(createRes.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), createRes, getRes)
}

func (suite *PgTestSuite) TestGetUnknownID() {
	res, err := suite.s.GetObject(uuid.New().String())
	assert.NoError(suite.T(), err)
	assert.Nil(suite.T(), res)
}

func (suite *PgTestSuite) TestList() {
	obj1 := &generated.Object{Test: "test", CreatedBy: "userID"}
	_, err := suite.s.CreateObject(obj1)
	assert.NoError(suite.T(), err)

	obj2 := &generated.Object{Test: "test", CreatedBy: "userID"}
	_, err = suite.s.CreateObject(obj2)
	assert.NoError(suite.T(), err)

	res, err := suite.s.ListObjects(100)
	assert.NoError(suite.T(), err)
	ids := []string{}
	for _, o := range res {
		ids = append(ids, o.ID)
	}
	assert.Contains(suite.T(), ids, obj1.ID)
	assert.Contains(suite.T(), ids, obj2.ID)
}

func (suite *PgTestSuite) TestUpdate() {
	obj := &generated.Object{Test: "test", CreatedBy: "userID"}
	createRes, err := suite.s.CreateObject(obj)
	assert.NoError(suite.T(), err)

	update := &generated.Object{ID: obj.ID, Test: "test2", CreatedBy: "userID2"}
	err = suite.s.UpdateObject(update, "action")
	assert.NoError(suite.T(), err)

	getRes, err := suite.s.GetObject(createRes.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), update.Test, getRes.Test)
	assert.Equal(suite.T(), createRes.CreatedAt, getRes.CreatedAt)
	// CreatedBy is unchanged since it's not an action field
	assert.Equal(suite.T(), createRes.CreatedBy, getRes.CreatedBy)
}

func (suite *PgTestSuite) TestDelete() {
	obj := &generated.Object{Test: "test", CreatedBy: "userID"}
	createRes, err := suite.s.CreateObject(obj)
	assert.NoError(suite.T(), err)

	err = suite.s.DeleteObject(createRes.ID)
	assert.NoError(suite.T(), err)

	nilRes, err := suite.s.GetObject(createRes.ID)
	assert.NoError(suite.T(), err)
	assert.Nil(suite.T(), nilRes)
}

func TestPgTestSuite(t *testing.T) {
	suite.Run(t, new(PgTestSuite))
}
