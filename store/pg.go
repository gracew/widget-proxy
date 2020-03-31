package store

import (
	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"github.com/gracew/widget-proxy/generated"
	"github.com/pkg/errors"
)

// PgStore implements the Store interface using Postgres.
type PgStore struct {
	Store
	DB *pg.DB
}

// CreateSchema creates the object table if it does not exist.
func (s PgStore) CreateSchema() error {
	_, err := s.DB.Exec("CREATE EXTENSION IF NOT EXISTS pgcrypto")
	if err != nil {
		return errors.Wrap(err, "failed to create pgcrypto extension")
	}
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

// CreateObject inserts the object into the database.
func (s PgStore) CreateObject(obj *generated.Object) (*generated.Object, error) {
	err := s.DB.Insert(obj)
	if err != nil {
		return nil, errors.Wrap(err, "failed to insert into database")
	}

	return obj, nil
}

// GetObject gets an object by ID. It returns nil if the object is not found.
func (s PgStore) GetObject(objectID string) (*generated.Object, error) {
	object := &generated.Object{ID: objectID}
	err := s.DB.Select(object)
	if err != nil {
		if errors.Is(err, pg.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return object, nil
}

// ListObjects retrieves the specified number of objects, ordered by created_at DESC.
func (s PgStore) ListObjects(pageSize int) ([]generated.Object, error) {
	var models []generated.Object
	err := s.DB.Model(&models).Order("created_at DESC").Limit(pageSize).Select()
	if err != nil {
		return nil, err
	}

	return models, nil
}

// UpdateObject updates the specified object in the database.
func (s PgStore) UpdateObject(obj *generated.Object, action string) (*generated.Object, error) {
	err := s.DB.Update(obj)
	if err != nil {
		return nil, errors.Wrap(err, "failed to update object")
	}

	return obj, nil
}

// DeleteObject deletes the specified object from the database.
func (s PgStore) DeleteObject(objectID string) error {
	object := &generated.Object{ID: objectID}
	return s.DB.Delete(object)
}
