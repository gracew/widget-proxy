package store

import (
	"time"

	"github.com/gracew/widget-proxy/generated"
	"github.com/gracew/widget-proxy/metrics"
	"github.com/gracew/widget-proxy/model"
)

type InstrumentedStore struct {
	Store
	Delegate Store
}

func (s InstrumentedStore) CreateSchema() error {
	return s.Delegate.CreateSchema()
}

func (s InstrumentedStore) CreateObject(req []byte, userID string) (*generated.Object, error) {
	start := time.Now()
	res, err := s.Delegate.CreateObject(req, userID)
	end := time.Now()
	metrics.DatabaseSummary.WithLabelValues(model.OperationTypeCreate.String()).Observe(end.Sub(start).Seconds())
	return res, err
}

func (s InstrumentedStore) GetObject(objectID string) (*generated.Object, error) {
	start := time.Now()
	res, err := s.Delegate.GetObject(objectID)
	end := time.Now()
	metrics.DatabaseSummary.WithLabelValues(model.OperationTypeRead.String()).Observe(end.Sub(start).Seconds())
	return res, err
}

func (s InstrumentedStore) ListObjects(pageSize int) ([]generated.Object, error) {
	start := time.Now()
	res, err := s.Delegate.ListObjects(pageSize)
	end := time.Now()
	metrics.DatabaseSummary.WithLabelValues(model.OperationTypeList.String()).Observe(end.Sub(start).Seconds())
	return res, err
}
