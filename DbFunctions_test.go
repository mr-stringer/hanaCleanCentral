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

func TestTruncateTraceFiles(t *testing.T) {
	/*Mock DB*/
	db1, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Errorf("an error '%s' was not expected when opening mock database connection", err)
	}
	defer db1.Close()

	type args struct {
		hdb           *sql.DB
		TrncDaysOlder uint
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"Good01", args{db1, 28}, false},
		{"Good02", args{db1, 7}, false},
		{"Good03", args{db1, 365}, false},
		{"Set to zero", args{db1, 0}, false},
		{"Bad01 Trace Query Fails", args{db1, 28}, true},
		{"Bad01 Trace Clear Fails", args{db1, 28}, false},
	}
	for _, tt := range tests {

		/*Prep the DB*/
		if tt.name == "Bad01 Trace Query Fails" {
			mock.ExpectQuery(GetTraceFileQuery(tt.args.TrncDaysOlder)).WillReturnError(fmt.Errorf("Some DB error"))
		} else if tt.name == "Bad01 Trace Clear Fails" {
			rows := sqlmock.NewRows([]string{"HOST", "FILE_NAME", "FILE_SIZE", "FILE_MTIME"}).AddRow("hanaserver", "trace.trc", "6400000", "2020-03-14 23:13:35.000000000")
			mock.ExpectQuery(GetTraceFileQuery(tt.args.TrncDaysOlder)).WillReturnRows(rows)
			mock.ExpectExec(GetRemoveTrace("hanaserver", "trace.trc")).WillReturnError(fmt.Errorf("Some DB error"))
		} else if tt.name[0:4] == "Good" {
			rows := sqlmock.NewRows([]string{"HOST", "FILE_NAME", "FILE_SIZE", "FILE_MTIME"}).AddRow("hanaserver", "trace.trc", "6400000", "2020-03-14 23:13:35.000000000")
			mock.ExpectQuery(GetTraceFileQuery(tt.args.TrncDaysOlder)).WillReturnRows(rows)
			mock.ExpectExec(GetRemoveTrace("hanaserver", "trace.trc")).WillReturnResult(sqlmock.NewResult(1, 1))
		}

		t.Run(tt.name, func(t *testing.T) {
			if err := TruncateTraceFiles(tt.args.hdb, tt.args.TrncDaysOlder); (err != nil) != tt.wantErr {
				t.Errorf("TruncateTraceFiles() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
