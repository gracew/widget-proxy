package store

import (
	"testing"

	"github.com/go-pg/pg"
	"github.com/gracew/widget-proxy/generated"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type PgTestSuite struct {
	suite.Suite
	s PgStore
}

var db *pg.DB

func (suite *PgTestSuite) SetupTest() {
	db = pg.Connect(&pg.Options{User: "postgres", Addr: "localhost:5433"})
	suite.s = PgStore{DB: db}
	err := suite.s.CreateSchema()
	assert.NoError(suite.T(), err)
}

func (suite *PgTestSuite) TearDownTest() {
	db.Close()
}

func (suite *PgTestSuite) TestCreateGetDeleteObject() {
	err := suite.s.CreateSchema()
	assert.NoError(suite.T(), err)

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

	err = suite.s.DeleteObject(createRes.ID)
	assert.NoError(suite.T(), err)

	nilRes, err := suite.s.GetObject(createRes.ID)
	assert.NoError(suite.T(), err)
	assert.Nil(suite.T(), nilRes)
}

func (suite *PgTestSuite) TestListObjects() {
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

func TestPgTestSuite(t *testing.T) {
	suite.Run(t, new(PgTestSuite))
}
