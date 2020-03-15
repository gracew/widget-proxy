package store

import (
	"github.com/gracew/widget-proxy/generated"
)

type Store interface {
	CreateSchema() error
	CreateObject(req []byte, userID string) (*generated.Object, error)
	GetObject(objectID string) (*generated.Object, error)
	ListObjects(pageSize int) ([]generated.Object, error)
}