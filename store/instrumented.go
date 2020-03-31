package store

import (
	"time"

	"github.com/gracew/widget-proxy/generated"
	"github.com/gracew/widget-proxy/metrics"
)

// InstrumentedStore implements the Store interface and delegates to another Store instance, instrumenting key
// operations.
type InstrumentedStore struct {
	Store
	Delegate Store
}

// CreateSchema delegates to another Store instance. It does not record the duration of the operation.
func (s InstrumentedStore) CreateSchema() error {
	return s.Delegate.CreateSchema()
}

// CreateObject delegates to another Store instance and records the duration of the operation.
func (s InstrumentedStore) CreateObject(obj *generated.Object) (*generated.Object, error) {
	start := time.Now()
	res, err := s.Delegate.CreateObject(obj)
	end := time.Now()
	metrics.DatabaseSummary.WithLabelValues(metrics.CREATE).Observe(end.Sub(start).Seconds())
	return res, err
}

// GetObject delegates to another Store instance and records the duration of the operation.
func (s InstrumentedStore) GetObject(objectID string) (*generated.Object, error) {
	start := time.Now()
	res, err := s.Delegate.GetObject(objectID)
	end := time.Now()
	metrics.DatabaseSummary.WithLabelValues(metrics.READ).Observe(end.Sub(start).Seconds())
	return res, err
}

// ListObjects delegates to another Store instance and records the duration of the operation.
func (s InstrumentedStore) ListObjects(pageSize int) ([]generated.Object, error) {
	start := time.Now()
	res, err := s.Delegate.ListObjects(pageSize)
	end := time.Now()
	metrics.DatabaseSummary.WithLabelValues(metrics.LIST).Observe(end.Sub(start).Seconds())
	return res, err
}

// UpdateObject delegates to another Store instance and records the duration of the operation.
func (s InstrumentedStore) UpdateObject(obj *generated.Object, action string) error {
	start := time.Now()
	err := s.Delegate.UpdateObject(obj, action)
	end := time.Now()
	metrics.DatabaseSummary.WithLabelValues(action).Observe(end.Sub(start).Seconds())
	return err
}

// DeleteObject delegates to another Store instance and records the duration of the operation.
func (s InstrumentedStore) DeleteObject(objectID string) error {
	start := time.Now()
	err := s.Delegate.DeleteObject(objectID)
	end := time.Now()
	metrics.DatabaseSummary.WithLabelValues(metrics.DELETE).Observe(end.Sub(start).Seconds())
	return err
}
