package postgres

import (
	"database/sql"
	"io/ioutil"
	"testing"
)

var dbdsn string = "postgres://kishor:@localhost/test_kanban?sslmode=disable"

func newTestDB(t *testing.T) (*sql.DB, func()) {
	db, err := sql.Open("postgres", dbdsn)
	if err != nil {
		t.Fatal(err)
	}

	script, err := ioutil.ReadFile("../sql/setup.sql")
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(string(script))
	if err != nil {
		t.Fatal(err)
	}

	return db, func() {
		script, err := ioutil.ReadFile("../sql/setup_reverse.sql")
		if err != nil {
			t.Fatal(err)
		}
		_, err = db.Exec(string(script))
		if err != nil {
			t.Fatal(err)
		}

		db.Close()
	}
}
