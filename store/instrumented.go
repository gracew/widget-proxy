package store

import (
	"time"

	"github.com/gracew/widget-proxy/generated"
	"github.com/gracew/widget-proxy/metrics"
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
	metrics.DatabaseSummary.WithLabelValues(metrics.CREATE).Observe(end.Sub(start).Seconds())
	return res, err
}

func (s InstrumentedStore) GetObject(objectID string) (*generated.Object, error) {
	start := time.Now()
	res, err := s.Delegate.GetObject(objectID)
	end := time.Now()
	metrics.DatabaseSummary.WithLabelValues(metrics.READ).Observe(end.Sub(start).Seconds())
	return res, err
}

func (s InstrumentedStore) ListObjects(pageSize int) ([]generated.Object, error) {
	start := time.Now()
	res, err := s.Delegate.ListObjects(pageSize)
	end := time.Now()
	metrics.DatabaseSummary.WithLabelValues(metrics.LIST).Observe(end.Sub(start).Seconds())
	return res, err
}

func (s InstrumentedStore) UpdateObject(objectID string, action string, req []byte) (*generated.Object, error) {
	start := time.Now()
	res, err := s.Delegate.UpdateObject(objectID, action, req)
	end := time.Now()
	metrics.DatabaseSummary.WithLabelValues(action).Observe(end.Sub(start).Seconds())
	return res, err
}

func (s InstrumentedStore) DeleteObject(objectID string) error {
	start := time.Now()
	err := s.Delegate.DeleteObject(objectID)
	end := time.Now()
	metrics.DatabaseSummary.WithLabelValues(metrics.DELETE).Observe(end.Sub(start).Seconds())
	return err
}
