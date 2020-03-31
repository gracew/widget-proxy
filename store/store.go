package store

import (
	"github.com/gracew/widget-proxy/generated"
)

type Store interface {
	CreateSchema() error
	CreateObject(obj *generated.Object) (*generated.Object, error)
	GetObject(objectID string) (*generated.Object, error)
	ListObjects(pageSize int) ([]generated.Object, error)
	UpdateObject(ob *generated.Object, action string) error
	DeleteObject(objectID string) error
}
