package postgres

import (
	"testing"

	"github.com/frinx/schellar/ifc"
	"github.com/frinx/schellar/it"
)

func TestSqlParamsRange(t *testing.T) {
	actual := sqlParamsRange(3)
	expected := "($1,$2,$3)"
	if actual != expected {
		t.Fatalf("Unexpected: %v, should be %v", actual, expected)
	}
}

func initIntegration(t *testing.T) ifc.DB {
	db := InitDB()
	it.ExpectTableSize(db, 0, "before test", t)
	return db
}

func TestAllIntegration(t *testing.T) {
	it.AllIntegration(t, initIntegration)
}
