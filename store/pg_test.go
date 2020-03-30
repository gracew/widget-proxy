package store

import (
	"encoding/json"
	"testing"

	"github.com/go-pg/pg"
	"github.com/stretchr/testify/assert"
)

type TestInput struct {
	Test string `json:"test"`
}

func TestCreateGetDeleteObject(t *testing.T) {
	db := pg.Connect(&pg.Options{User: "postgres", Addr: "localhost:5433"})
	defer db.Close()
	s := PgStore{DB: db}

	err := s.CreateSchema()
	assert.NoError(t, err)

	req := TestInput{Test: "test"}
	marshaled, err := json.Marshal(req)
	assert.NoError(t, err)

	createRes, err := s.CreateObject(marshaled, "userID")
	assert.NoError(t, err)
	assert.Equal(t, req.Test, createRes.Test)
	assert.Equal(t, "userID", createRes.CreatedBy)

	getRes, err := s.GetObject(createRes.ID)
	assert.NoError(t, err)
	assert.Equal(t, createRes, getRes)

	err = s.DeleteObject(createRes.ID)
	assert.NoError(t, err)

	_, err = s.GetObject(createRes.ID)
	assert.Error(t, err)
}

func TestListObjects(t *testing.T) {
	db := pg.Connect(&pg.Options{User: "postgres", Addr: "localhost:5433"})
	defer db.Close()
	s := PgStore{DB: db}

	err := s.CreateSchema()
	assert.NoError(t, err)

	req1 := TestInput{Test: "test"}
	marshaled, err := json.Marshal(req1)
	assert.NoError(t, err)
	_, err = s.CreateObject(marshaled, "userID")
	assert.NoError(t, err)

	req2 := TestInput{Test: "test"}
	marshaled, err = json.Marshal(req2)
	assert.NoError(t, err)
	_, err = s.CreateObject(marshaled, "userID")
	assert.NoError(t, err)

	res, err := s.ListObjects(100)
	assert.NoError(t, err)
	assert.Greater(t, len(res), 0)
}
