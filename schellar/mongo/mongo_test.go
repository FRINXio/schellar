package mongo

import (
	"testing"

	"github.com/frinx/schellar/ifc"
	"github.com/frinx/schellar/it"
)

func initIntegration(t *testing.T) ifc.DB {
	db := InitDB()
	it.ExpectTableSize(db, 0, "before test", t)
	return db
}

func TestAllIntegration(t *testing.T) {
	it.AllIntegration(t, initIntegration)
}
