package store

import (
	"encoding/json"
	"time"

	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"github.com/gracew/widget-proxy/generated"
	"github.com/gracew/widget-proxy/metrics"
	"github.com/gracew/widget-proxy/model"
	"github.com/pkg/errors"
)

type Store struct {
	DB *pg.DB
}

func (s Store) CreateSchema() error {
	for _, model := range []interface{}{(*generated.Object)(nil)} {
		err := s.DB.CreateTable(model, &orm.CreateTableOptions{
			IfNotExists: true,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (s Store) CreateObject(req []byte, userID string) (*generated.Object, error) {
	start := time.Now()
	res, err := s.createObject(req, userID)
	end := time.Now()
	metrics.DatabaseSummary.WithLabelValues(model.OperationTypeCreate.String()).Observe(end.Sub(start).Seconds())
	return res, err
}

func (s Store) createObject(req []byte, userID string) (*generated.Object, error) {
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

func (s Store) GetObject(objectID string) (*generated.Object, error) {
	start := time.Now()
	res, err := s.getObject(objectID)
	end := time.Now()
	metrics.DatabaseSummary.WithLabelValues(model.OperationTypeRead.String()).Observe(end.Sub(start).Seconds())
	return res, err
}

func (s Store) getObject(objectID string) (*generated.Object, error) {
	object := &generated.Object{ID: objectID}
	err := s.DB.Select(object)
	if err != nil {
		return nil, err
	}

	return object, nil
}

func (s Store) ListObjects(pageSize int) ([]generated.Object, error) {
	start := time.Now()
	res, err := s.listObjects(pageSize)
	end := time.Now()
	metrics.DatabaseSummary.WithLabelValues(model.OperationTypeList.String()).Observe(end.Sub(start).Seconds())
	return res, err
}

func (s Store) listObjects(pageSize int) ([]generated.Object, error) {
	var models []generated.Object
	// TODO(gracew): specify a sort order so that paging actually works lol
	err := s.DB.Model(&models).Limit(pageSize).Select()
	if err != nil {
		return nil, err
	}

	return models, nil
}
