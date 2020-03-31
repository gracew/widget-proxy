package store

import (
	"github.com/gracew/widget-proxy/generated"
)

type Store interface {
	CreateSchema() error
	CreateObject(obj *generated.Object) (*generated.Object, error)
	GetObject(objectID string) (*generated.Object, error)
	ListObjects(pageSize int) ([]generated.Object, error)
	UpdateObject(objectID string, action string, req []byte) (*generated.Object, error)
	DeleteObject(objectID string) error
}
