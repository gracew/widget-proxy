package store

import (
	"encoding/json"

	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"github.com/gracew/widget-proxy/generated"
	"github.com/pkg/errors"
)

type PgStore struct {
	Store
	DB *pg.DB
}

func (s PgStore) CreateSchema() error {
	for _, model := range []interface{}{(*generated.Object)(nil)} {
		err := s.DB.CreateTable(model, &orm.CreateTableOptions{
			IfNotExists: true,
		})
		if err != nil {
			return errors.Wrap(err, "failed to initialize schema")
		}
	}
	return nil
}

func (s PgStore) CreateObject(req []byte, userID string) (*generated.Object, error) {
	var dbModel generated.Object
	err := json.Unmarshal(req, &dbModel)
	if err != nil {
		return nil, errors.Wrap(err, "could not unmarshal input as object")
	}

	dbModel.CreatedBy = userID

	err = s.DB.Insert(&dbModel)
	if err != nil {
		return nil, errors.Wrap(err, "failed to insert into database")
	}

	return &dbModel, nil
}

func (s PgStore) GetObject(objectID string) (*generated.Object, error) {
	object := &generated.Object{ID: objectID}
	err := s.DB.Select(object)
	if err != nil {
		return nil, err
	}

	return object, nil
}

func (s PgStore) ListObjects(pageSize int) ([]generated.Object, error) {
	var models []generated.Object
	// TODO(gracew): specify a sort order so that paging actually works lol
	err := s.DB.Model(&models).Order("created_at DESC").Limit(pageSize).Select()
	if err != nil {
		return nil, err
	}

	return models, nil
}

func (s PgStore) DeleteObject(objectID string) error {
	object := &generated.Object{ID: objectID}
	return s.DB.Delete(object)
}