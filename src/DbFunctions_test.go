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

	/*Create the log channel*/
	lc := make(chan LogMessage)
	quit := make(chan bool)
	go Logger(AppConfig{"file", true, false}, lc, quit)

	type args struct {
		lc            chan LogMessage
		name          string
		hdb           *sql.DB
		TrncDaysOlder uint
		DryRun        bool
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"Good01", args{lc, "db@sid", db1, 28, false}, false},
		{"Good02", args{lc, "db@sid", db1, 7, false}, false},
		{"Good03", args{lc, "db@sid", db1, 365, false}, false},
		{"Set to zero", args{lc, "db@sid", db1, 0, false}, false},
		{"Bad01 Trace Query Fails", args{lc, "db@sid", db1, 28, false}, true},
		{"Bad02 Unscannable Data", args{lc, "db@sid", db1, 28, false}, true},
		{"Bad03 Trace Clear Fails", args{lc, "db@sid", db1, 28, false}, false},
		{"MultiTraceGood", args{lc, "db@sid", db1, 28, false}, false},
		{"MultiTraceCantDelete1", args{lc, "db@sid", db1, 28, false}, false},
		{"NothingToDelete", args{lc, "db@sid", db1, 28, false}, false},
		{"RemovalRowError", args{lc, "db@sid", db1, 28, false}, false},
	}
	for _, tt := range tests {

		/*Prep the DB*/
		if tt.name == "Bad01 Trace Query Fails" {
			mock.ExpectQuery(GetTraceFileQuery(tt.args.TrncDaysOlder)).WillReturnError(fmt.Errorf("Some DB error"))
		} else if tt.name == "Bad01 Trace Clear Fails" {
			rows := sqlmock.NewRows([]string{"HOST", "FILE_NAME", "FILE_SIZE", "FILE_MTIME"}).AddRow("hanaserver", "trace.trc", "6400000", "2020-03-14 23:13:35.000000000")
			mock.ExpectQuery(GetTraceFileQuery(tt.args.TrncDaysOlder)).WillReturnRows(rows)
			mock.ExpectExec(GetRemoveTrace("hanaserver", "trace.trc")).WillReturnError(fmt.Errorf("Some DB error"))
		} else if tt.name == "Bad02 Unscannable Data" {
			rows := sqlmock.NewRows([]string{"HOST", "FILE_NAME", "FILE_SIZE", "FILE_MTIME"}).AddRow("hanaserver", "trace.trc", "BAD_DATA", "2020-03-14 23:13:35.000000000")
			mock.ExpectQuery(GetTraceFileQuery(tt.args.TrncDaysOlder)).WillReturnRows(rows)
		} else if tt.name == "Bad03 Trace Clear Fails" {
			rows1 := sqlmock.NewRows([]string{"HOST", "FILE_NAME", "FILE_SIZE", "FILE_MTIME"}).AddRow("hanaserver", "trace.trc", "6400000", "2020-03-14 23:13:35.000000000")
			mock.ExpectQuery(GetTraceFileQuery(tt.args.TrncDaysOlder)).WillReturnRows(rows1)
			mock.ExpectExec(GetRemoveTrace("hanaserver", "trace.trc")).WillReturnError(fmt.Errorf("Some DB error"))
		} else if tt.name[0:4] == "Good" {
			rows1 := sqlmock.NewRows([]string{"HOST", "FILE_NAME", "FILE_SIZE", "FILE_MTIME"}).AddRow("hanaserver", "trace.trc", "6400000", "2020-03-14 23:13:35.000000000")
			rows2 := sqlmock.NewRows([]string{"TRACE"}).AddRow("0")
			mock.ExpectQuery(GetTraceFileQuery(tt.args.TrncDaysOlder)).WillReturnRows(rows1)
			mock.ExpectExec(GetRemoveTrace("hanaserver", "trace.trc")).WillReturnResult(sqlmock.NewResult(1, 1))
			mock.ExpectQuery(GetCheckTracePresent("trace.trc")).WillReturnRows(rows2)
		} else if tt.name == "MultiTraceGood" {
			rows1 := sqlmock.NewRows([]string{"HOST", "FILE_NAME", "FILE_SIZE", "FILE_MTIME"}).AddRow("hanaserver", "trace.trc", "6400000", "2020-03-14 23:13:35.000000000").AddRow("hanaserver", "trace2.trc", "6400000", "2020-03-14 23:13:35.000000000")
			rows2 := sqlmock.NewRows([]string{"TRACE"}).AddRow("0")
			mock.ExpectQuery(GetTraceFileQuery(tt.args.TrncDaysOlder)).WillReturnRows(rows1)
			mock.ExpectExec(GetRemoveTrace("hanaserver", "trace.trc")).WillReturnResult(sqlmock.NewResult(1, 1))
			mock.ExpectQuery(GetCheckTracePresent("trace.trc")).WillReturnRows(rows2)
			mock.ExpectExec(GetRemoveTrace("hanaserver", "trace2.trc")).WillReturnResult(sqlmock.NewResult(1, 1))
			mock.ExpectQuery(GetCheckTracePresent("trace2.trc")).WillReturnRows(rows2)
		} else if tt.name == "MultiTraceCantDelete1" {
			rows1 := sqlmock.NewRows([]string{"HOST", "FILE_NAME", "FILE_SIZE", "FILE_MTIME"}).AddRow("hanaserver", "trace.trc", "6400000", "2020-03-14 23:13:35.000000000").AddRow("hanaserver", "trace2.trc", "6400000", "2020-03-14 23:13:35.000000000")
			rows2 := sqlmock.NewRows([]string{"TRACE"}).AddRow("1")
			rows3 := sqlmock.NewRows([]string{"TRACE"}).AddRow("0")
			mock.ExpectQuery(GetTraceFileQuery(tt.args.TrncDaysOlder)).WillReturnRows(rows1)
			mock.ExpectExec(GetRemoveTrace("hanaserver", "trace.trc")).WillReturnResult(sqlmock.NewResult(1, 1))
			mock.ExpectQuery(GetCheckTracePresent("trace.trc")).WillReturnRows(rows2)
			mock.ExpectExec(GetRemoveTrace("hanaserver", "trace2.trc")).WillReturnResult(sqlmock.NewResult(1, 1))
			mock.ExpectQuery(GetCheckTracePresent("trace2.trc")).WillReturnRows(rows3)
		} else if tt.name == "NothingToDelete" {
			rows1 := sqlmock.NewRows([]string{"HOST", "FILE_NAME", "FILE_SIZE", "FILE_MTIME"})
			mock.ExpectQuery(GetTraceFileQuery(tt.args.TrncDaysOlder)).WillReturnRows(rows1)
		} else if tt.name == "RemovalNoRows" {
			rows1 := sqlmock.NewRows([]string{"HOST", "FILE_NAME", "FILE_SIZE", "FILE_MTIME"}).AddRow("hanaserver", "traceNoRows.trc", "6400000", "2020-03-14 23:13:35.000000000")
			rows2 := sqlmock.NewRows([]string{"TRACE"})
			mock.ExpectQuery(GetTraceFileQuery(tt.args.TrncDaysOlder)).WillReturnRows(rows1)
			mock.ExpectExec(GetRemoveTrace("hanaserver", "traceNoRows.trc")).WillReturnResult(sqlmock.NewResult(1, 1))
			mock.ExpectQuery(GetCheckTracePresent("traceNoRows.trc")).WillReturnRows(rows2)
		} else if tt.name == "RemovalRowError" {
			rows1 := sqlmock.NewRows([]string{"HOST", "FILE_NAME", "FILE_SIZE", "FILE_MTIME"}).AddRow("hanaserver", "traceNoRows.trc", "6400000", "2020-03-14 23:13:35.000000000")
			mock.ExpectQuery(GetTraceFileQuery(tt.args.TrncDaysOlder)).WillReturnRows(rows1)
			mock.ExpectExec(GetRemoveTrace("hanaserver", "traceNoRows.trc")).WillReturnResult(sqlmock.NewResult(1, 1))
			mock.ExpectQuery(GetCheckTracePresent("traceNoRows.trc")).WillReturnError(fmt.Errorf("Some DB error"))
		}

		t.Run(tt.name, func(t *testing.T) {
			if err := TruncateTraceFiles(tt.args.lc, tt.args.name, tt.args.hdb, tt.args.TrncDaysOlder, tt.args.DryRun); (err != nil) != tt.wantErr {
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

	lc := make(chan LogMessage)
	quit := make(chan bool)
	go Logger(AppConfig{"file", false, false}, lc, quit)

	type args struct {
		lc            chan LogMessage
		name          string
		hdb           *sql.DB
		TrncDaysOlder uint
		delete        bool
		dryrun        bool
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"Good01", args{lc, "db@sid", db1, 28, false, false}, false},
		{"Good02 Complete", args{lc, "db@sid", db1, 28, true, false}, false},
		{"Get ID Failed", args{lc, "db@sid", db1, 28, true, false}, true},
		{"Get ID No Rows", args{lc, "db@sid", db1, 28, true, false}, false},
		{"Nothing to Delete", args{lc, "db@sid", db1, 28, true, false}, false},
		{"Get Backup File Data Failed", args{lc, "db@sid", db1, 28, true, false}, true},
		{"Get Backup Delete Complete Failed", args{lc, "db@sid", db1, 28, true, false}, true},
		{"Get Backup Delete Failed", args{lc, "db@sid", db1, 28, false, false}, true},
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
		} else if tt.name == "Get ID No Rows" {
			mock.ExpectQuery(GetLatestFullBackupID(tt.args.TrncDaysOlder)).WillReturnError(sql.ErrNoRows)
		} else if tt.name == "Get Backup File Data Failed" {
			var backupID string = "12345678890"
			rows1 := sqlmock.NewRows([]string{"BACKUP_ID"}).AddRow(backupID)
			mock.ExpectQuery(GetLatestFullBackupID(tt.args.TrncDaysOlder)).WillReturnRows(rows1)
			mock.ExpectQuery(GetBackupFileData(backupID)).WillReturnError(fmt.Errorf("Some DB error"))
		} else if tt.name == "Nothing to Delete" {
			var backupID string = "12345678890"
			rows1 := sqlmock.NewRows([]string{"BACKUP_ID"}).AddRow(backupID)
			rows2 := sqlmock.NewRows([]string{"ENTRY", "COUNT", "BYTES"})
			mock.ExpectQuery(GetLatestFullBackupID(tt.args.TrncDaysOlder)).WillReturnRows(rows1)
			mock.ExpectQuery(GetBackupFileData(backupID)).WillReturnRows(rows2)

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
			if err := TruncateBackupCatalog(tt.args.lc, tt.args.name, tt.args.hdb, tt.args.TrncDaysOlder, tt.args.delete, tt.args.dryrun); (err != nil) != tt.wantErr {
				t.Errorf("TruncateBackupCatalog() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClearAlert(t *testing.T) {
	/*Mock DB*/
	db1, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Errorf("an error '%s' was not expected when opening mock database connection", err)
	}
	defer db1.Close()

	lc := make(chan LogMessage)
	quit := make(chan bool)
	go Logger(AppConfig{"file", false, false}, lc, quit)

	type args struct {
		lc              chan<- LogMessage
		name            string
		hdb             *sql.DB
		DeleteOlderDays uint
		dryrun          bool
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"Good01", args{lc, "db@sid", db1, 14, false}, false},
		{"GetAlertsNoRows", args{lc, "db@sid", db1, 14, false}, true},
		{"GetAlertsDbError", args{lc, "db@sid", db1, 14, false}, true},
		{"NothingToDo", args{lc, "db@sid", db1, 14, false}, false},
		{"AlertDeleteFailed", args{lc, "db@sid", db1, 14, false}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			/*Set up mocks for test table*/
			if tt.name == "Good01" {
				rows1 := sqlmock.NewRows([]string{"COUNT"}).AddRow("250")
				mock.ExpectQuery(GetAlertCount(tt.args.DeleteOlderDays)).WillReturnRows(rows1)
				mock.ExpectExec(GetAlertDelete(tt.args.DeleteOlderDays)).WillReturnResult(sqlmock.NewResult(1, 1))
			} else if tt.name == "GetAlertsNoRows" {
				rows1 := sqlmock.NewRows([]string{"COUNT"})
				mock.ExpectQuery(GetAlertCount(tt.args.DeleteOlderDays)).WillReturnRows(rows1)
			} else if tt.name == "GetAlertsDbError" {
				mock.ExpectQuery(GetAlertCount(tt.args.DeleteOlderDays)).WillReturnError(fmt.Errorf("Get Alerts Error"))
			} else if tt.name == "NothingToDo" {
				rows1 := sqlmock.NewRows([]string{"COUNT"}).AddRow("0")
				mock.ExpectQuery(GetAlertCount(tt.args.DeleteOlderDays)).WillReturnRows(rows1)
			} else if tt.name == "AlertDeleteFailed" {
				rows1 := sqlmock.NewRows([]string{"COUNT"}).AddRow("250")
				mock.ExpectQuery(GetAlertCount(tt.args.DeleteOlderDays)).WillReturnRows(rows1)
				mock.ExpectExec(GetAlertDelete(tt.args.DeleteOlderDays)).WillReturnError(fmt.Errorf("Delete Alerts Error"))
			}

			//	var backupID string = "12345678890"
			//rows1 := sqlmock.NewRows([]string{"BACKUP_ID"}).AddRow(backupID)
			//rows2 := sqlmock.NewRows([]string{"ENTRY", "COUNT", "BYTES"}).AddRow("complete data backup", 10, 100000000).AddRow("log backup", 100, 100000000)
			//mock.ExpectQuery(GetLatestFullBackupID(tt.args.TrncDaysOlder)).WillReturnRows(rows1)
			//mock.ExpectQuery(GetBackupFileData(backupID)).WillReturnRows(rows2)
			//mock.ExpectExec(GetBackupDeleteComplete(backupID)).WillReturnError(fmt.Errorf("Some DB error"))
			//
			if err := ClearAlert(tt.args.lc, tt.args.name, tt.args.hdb, tt.args.DeleteOlderDays, tt.args.dryrun); (err != nil) != tt.wantErr {
				t.Errorf("ClearAlert() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestReclaimLog(t *testing.T) {
	/*Mock DB*/
	db1, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Errorf("an error '%s' was not expected when opening mock database connection", err)
	}
	defer db1.Close()

	lc := make(chan LogMessage)
	quit := make(chan bool)

	defer close(lc)
	defer close(quit)

	go Logger(AppConfig{"file", false, false}, lc, quit)

	type args struct {
		lc     chan<- LogMessage
		name   string
		hdb    *sql.DB
		dryrun bool
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"Good", args{lc, "db@sid", db1, false}, false},
		{"GetSegmentsNoRows", args{lc, "db@sid", db1, false}, true},
		{"GetSegmentsDbError", args{lc, "db@sid", db1, false}, true},
		{"ReclaimLogDbError", args{lc, "db@sid", db1, false}, true},
	}
	for _, tt := range tests {

		switch {
		case tt.name == "Good":
			rows1 := sqlmock.NewRows([]string{"COUNT", "BYTES"}).AddRow("10", "2048000")
			mock.ExpectQuery(QUERY_GetFeeLogSegments).WillReturnRows(rows1)
			mock.ExpectExec(QUERY_RecalimLog).WillReturnResult(sqlmock.NewResult(1, 1))
		case tt.name == "GetSegmentsDbError":
			rows1 := sqlmock.NewRows([]string{"COUNT", "BYTES"})
			mock.ExpectQuery(QUERY_GetFeeLogSegments).WillReturnRows(rows1)
		case tt.name == "GetSegmentsDbError":
			mock.ExpectQuery(QUERY_GetFeeLogSegments).WillReturnError(fmt.Errorf("some db Error"))
		case tt.name == "ReclaimLogDbError":
			rows1 := sqlmock.NewRows([]string{"COUNT", "BYTES"}).AddRow("10", "2048000")
			mock.ExpectQuery(QUERY_GetFeeLogSegments).WillReturnRows(rows1)
			mock.ExpectExec(QUERY_RecalimLog).WillReturnError(fmt.Errorf("some DB error"))
		}

		t.Run(tt.name, func(t *testing.T) {
			if err := ReclaimLog(tt.args.lc, tt.args.name, tt.args.hdb, tt.args.dryrun); (err != nil) != tt.wantErr {
				t.Errorf("ReclaimLog() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	quit <- true
}
