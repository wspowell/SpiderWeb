package db

import (
	"github.com/wspowell/spiderweb/examples/custom/resources"
)

var _ resources.Datastore = (*Database)(nil)

type Database struct{}

func NewDatabase() *Database {
	return &Database{}
}

func (self *Database) RetrieveValue() string {
	return "db_value"
}
