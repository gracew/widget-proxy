package store

//go:generate $GOPATH/bin/mockgen -source=$GOFILE -destination=$PWD/mocks/$GOFILE -package=mocks

import (
	"github.com/gracew/widget-proxy/generated"
)

type Store interface {
	CreateSchema() error
	CreateObject(obj *generated.Object) (*generated.Object, error)
	GetObject(objectID string) (*generated.Object, error)
	ListObjects(pageSize int, filter *Filter) ([]generated.Object, error)
	UpdateObject(ob *generated.Object, action string) (*generated.Object, error)
	DeleteObject(objectID string) error
}

type Filter struct {
	Field string
	Value interface{}
}
