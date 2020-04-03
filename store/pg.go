package store

import (
	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"github.com/gracew/widget-proxy/generated"
	"github.com/gracew/widget-proxy/model"
	"github.com/pkg/errors"
)

// PgStore implements the Store interface using Postgres.
type PgStore struct {
	Store
	API model.API
	DB  *pg.DB
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
func (s PgStore) ListObjects(pageSize int, filter *Filter) ([]generated.Object, error) {
	if filter != nil && !s.validFilter(*filter) {
		return nil, errors.New("invalid filter field: " + filter.Field)
	}

	var models []generated.Object
	m := s.DB.Model(&models).Order("created_at DESC")
	if filter != nil {
		m.Where(underscore(filter.Field) + " = ?", filter.Value)
	}
	err := m.Limit(pageSize).Select()
	if err != nil {
		return nil, err
	}

	return models, nil
}

func (s PgStore) validFilter(filter Filter) bool {
	for _, f := range s.API.Operations.List.Filter {
		if f == filter.Field {
			return true
		}
	}
	return false
}

// UpdateObject updates the specified object in the database.
func (s PgStore) UpdateObject(obj *generated.Object, actionName string) (*generated.Object, error) {
	// update only the fields specified by the action
	action := s.findAction(actionName)
	if action == nil {
		return nil, errors.New("unknown action " + actionName)
	}

	m := s.DB.Model(obj)
	for _, f := range action.Fields {
		m.Column(underscore(f))
	}
	_, err := m.WherePK().Returning("*").Update()
	if err != nil {
		return nil, errors.Wrap(err, "failed to update object")
	}
	return obj, nil
}

func (s PgStore) findAction(actionName string) *model.ActionDefinition {
	if s.API.Operations == nil || s.API.Operations.Update == nil {
		return nil
	}
	for _, action := range s.API.Operations.Update.Actions {
		if action.Name == actionName {
			return &action
		}
	}
	return nil
}

// DeleteObject deletes the specified object from the database.
func (s PgStore) DeleteObject(objectID string) error {
	object := &generated.Object{ID: objectID}
	return s.DB.Delete(object)
}
