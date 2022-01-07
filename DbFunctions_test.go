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
		ac            AppConfig
		name          string
		hdb           *sql.DB
		TrncDaysOlder uint
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"Good01", args{AppConfig{"file", false, false}, "db@sid", db1, 28}, false},
		{"Good02", args{AppConfig{"file", false, false}, "db@sid", db1, 7}, false},
		{"Good03", args{AppConfig{"file", false, false}, "db@sid", db1, 365}, false},
		{"Set to zero", args{AppConfig{"file", false, false}, "db@sid", db1, 0}, false},
		{"Bad01 Trace Query Fails", args{AppConfig{"file", false, false}, "db@sid", db1, 28}, true},
		{"Bad01 Trace Clear Fails", args{AppConfig{"file", false, false}, "db@sid", db1, 28}, false},
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
			if err := TruncateTraceFiles(tt.args.ac, tt.args.name, tt.args.hdb, tt.args.TrncDaysOlder); (err != nil) != tt.wantErr {
				t.Errorf("TruncateTraceFiles() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTruncateBackupCatalog(t *testing.T) {
	/*Mock DB*/
	db1, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Errorf("an error '%s' was not expected when opening mock database connection", err)
	}
	defer db1.Close()

	type args struct {
		ac            AppConfig
		name          string
		hdb           *sql.DB
		TrncDaysOlder uint
		delete        bool
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"Good01", args{AppConfig{"file", false, false}, "db@sid", db1, 28, false}, false},
		{"Good02 Complete", args{AppConfig{"file", false, false}, "db@sid", db1, 28, true}, false},
		{"Get ID Failed", args{AppConfig{"file", false, false}, "db@sid", db1, 28, true}, true},
		{"Get Backup File Data Failed", args{AppConfig{"file", false, false}, "db@sid", db1, 28, true}, true},
		{"Get Backup Delete Complete Failed", args{AppConfig{"file", false, false}, "db@sid", db1, 28, true}, true},
		{"Get Backup Delete Failed", args{AppConfig{"file", false, false}, "db@sid", db1, 28, false}, true},
	}
	for _, tt := range tests {

		if tt.name == "Good01" {
			var backupID string = "12345678890"
			rows1 := sqlmock.NewRows([]string{"BACKUP_ID"}).AddRow(backupID)
			rows2 := sqlmock.NewRows([]string{"ENTRY", "COUNT", "BYTES"}).AddRow("complete data backup", 10, 100000000).AddRow("log backup", 100, 100000000)
			mock.ExpectQuery(GetLatestFullBackupID(tt.args.TrncDaysOlder)).WillReturnRows(rows1)
			mock.ExpectQuery(GetBackupFileData(backupID)).WillReturnRows(rows2)
			mock.ExpectExec(GetBackupDelete(backupID)).WillReturnResult(sqlmock.NewResult(1, 1))
		} else if tt.name == "Good02 Complete" {
			var backupID string = "12345678890"
			rows1 := sqlmock.NewRows([]string{"BACKUP_ID"}).AddRow(backupID)
			rows2 := sqlmock.NewRows([]string{"ENTRY", "COUNT", "BYTES"}).AddRow("complete data backup", 10, 100000000).AddRow("log backup", 100, 100000000)
			mock.ExpectQuery(GetLatestFullBackupID(tt.args.TrncDaysOlder)).WillReturnRows(rows1)
			mock.ExpectQuery(GetBackupFileData(backupID)).WillReturnRows(rows2)
			mock.ExpectExec(GetBackupDeleteComplete(backupID)).WillReturnResult(sqlmock.NewResult(1, 1))
		} else if tt.name == "Get ID Failed" {
			mock.ExpectQuery(GetLatestFullBackupID(tt.args.TrncDaysOlder)).WillReturnError(fmt.Errorf("Some DB error"))
		} else if tt.name == "Get Backup File Data Failed" {
			var backupID string = "12345678890"
			rows1 := sqlmock.NewRows([]string{"BACKUP_ID"}).AddRow(backupID)
			mock.ExpectQuery(GetLatestFullBackupID(tt.args.TrncDaysOlder)).WillReturnRows(rows1)
			mock.ExpectQuery(GetBackupFileData(backupID)).WillReturnError(fmt.Errorf("Some DB error"))
		} else if tt.name == "Get Backup Delete Complete Failed" {
			var backupID string = "12345678890"
			rows1 := sqlmock.NewRows([]string{"BACKUP_ID"}).AddRow(backupID)
			rows2 := sqlmock.NewRows([]string{"ENTRY", "COUNT", "BYTES"}).AddRow("complete data backup", 10, 100000000).AddRow("log backup", 100, 100000000)
			mock.ExpectQuery(GetLatestFullBackupID(tt.args.TrncDaysOlder)).WillReturnRows(rows1)
			mock.ExpectQuery(GetBackupFileData(backupID)).WillReturnRows(rows2)
			mock.ExpectExec(GetBackupDeleteComplete(backupID)).WillReturnError(fmt.Errorf("Some DB error"))
		} else if tt.name == "Get Backup Delete Failed" {
			var backupID string = "12345678890"
			rows1 := sqlmock.NewRows([]string{"BACKUP_ID"}).AddRow(backupID)
			rows2 := sqlmock.NewRows([]string{"ENTRY", "COUNT", "BYTES"}).AddRow("complete data backup", 10, 100000000).AddRow("log backup", 100, 100000000)
			mock.ExpectQuery(GetLatestFullBackupID(tt.args.TrncDaysOlder)).WillReturnRows(rows1)
			mock.ExpectQuery(GetBackupFileData(backupID)).WillReturnRows(rows2)
			mock.ExpectExec(GetBackupDelete(backupID)).WillReturnError(fmt.Errorf("Some DB error"))
		} else {
			return
		}

		t.Run(tt.name, func(t *testing.T) {
			if err := TruncateBackupCatalog(tt.args.ac, tt.args.name, tt.args.hdb, tt.args.TrncDaysOlder, tt.args.delete); (err != nil) != tt.wantErr {
				t.Errorf("TruncateBackupCatalog() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
