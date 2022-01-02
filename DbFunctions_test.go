package main

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestHanaVersion(t *testing.T) {
	/*Mock DB*/
	db1, mock, err := sqlmock.New()
	if err != nil {
		t.Errorf("an error '%s' was not expected when opening mock database connection", err)
	}
	defer db1.Close()

	type args struct {
		hdb *sql.DB
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{"Good001", args{db1}, "2.00.45.00", false},
		{"Good002", args{db1}, "2.00.044.00.1571081837", false},
		{"Good003", args{db1}, "1.00.044.00.1571081837", false},
		{"Error001", args{db1}, "", true},
	}
	for _, tt := range tests {

		rows := sqlmock.NewRows([]string{"VERSION"}).AddRow(tt.want)

		if tt.wantErr {
			mock.ExpectQuery(QUERY_GetVersion).WillReturnError(fmt.Errorf("DB Error"))
		} else {
			mock.ExpectQuery(QUERY_GetVersion).WillReturnRows(rows)
		}

		t.Run(tt.name, func(t *testing.T) {
			got, err := HanaVersion(tt.args.hdb)
			if (err != nil) != tt.wantErr {
				t.Errorf("HanaVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("HanaVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}
