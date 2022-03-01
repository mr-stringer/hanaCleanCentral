package main

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestDbConfig_HanaVersionFunc(t *testing.T) {
	/*Test Setup*/
	/*Mock DB*/
	db1, mock, err := sqlmock.New()
	if err != nil {
		t.Errorf("an error '%s' was not expected when opening mock database connection", err)
	}
	defer db1.Close()

	/*Logger*/
	lc := make(chan LogMessage)
	quit := make(chan bool)
	go Logger(AppConfig{"file", true, false}, lc, quit)

	/*Types*/
	type args struct {
		lc chan<- LogMessage
	}

	/*Tests*/
	tests := []struct {
		name    string
		dbc     *DbConfig
		args    args
		want    string
		wantErr bool
	}{
		{"Good01", &DbConfig{"", "", 30015, "", "", true, 60, true, 60, true, true, 60, true, true, 60, true, db1}, args{lc}, "2.00.45.00", false},
		{"Good02", &DbConfig{"", "", 30015, "", "", true, 60, true, 60, true, true, 60, true, true, 60, true, db1}, args{lc}, "2.00.43.33", false},
		{"Good03", &DbConfig{"", "", 30015, "", "", true, 60, true, 60, true, true, 60, true, true, 60, true, db1}, args{lc}, "3.00.00.10", false},
		{"Good04", &DbConfig{"", "", 30015, "", "", true, 60, true, 60, true, true, 60, true, true, 60, true, db1}, args{lc}, "1.00.112.3", false},
		{"DbError", &DbConfig{"", "", 30015, "", "", true, 60, true, 60, true, true, 60, true, true, 60, true, db1}, args{lc}, "", true},
	}
	for _, tt := range tests {
		/*Set up per case mocking*/
		rows := sqlmock.NewRows([]string{"VERSION"}).AddRow(tt.want)
		if tt.wantErr {
			mock.ExpectQuery(QUERY_GetVersion).WillReturnError(fmt.Errorf("DB Error"))
		} else {
			mock.ExpectQuery(QUERY_GetVersion).WillReturnRows(rows)
		}

		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.dbc.HanaVersionFunc(tt.args.lc)
			if (err != nil) != tt.wantErr {
				t.Errorf("DbConfig.HanaVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("DbConfig.HanaVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDbConfig_CleanTraceFilesFunc(t *testing.T) {
	/*Test Setup*/
	/*Mock DB*/
	db1, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Errorf("an error '%s' was not expected when opening mock database connection", err)
	}
	defer db1.Close()

	/*Logger*/
	lc := make(chan LogMessage)
	quit := make(chan bool)
	go Logger(AppConfig{"file", true, false}, lc, quit)

	/*Args*/
	type args struct {
		lc             chan<- LogMessage
		CleanDaysOlder uint
		dryrun         bool
	}

	/*Tests*/
	tests := []struct {
		name    string
		dbc     *DbConfig
		args    args
		wantErr bool
	}{
		{"Good01", &DbConfig{"", "", 30015, "", "", true, 60, true, 60, true, true, 60, true, true, 60, true, db1}, args{lc, 60, false}, false},
		{"Good02", &DbConfig{"", "", 30015, "", "", true, 60, true, 60, true, true, 60, true, true, 60, true, db1}, args{lc, 14, false}, false},
		{"Good03", &DbConfig{"", "", 30015, "", "", true, 60, true, 60, true, true, 60, true, true, 60, true, db1}, args{lc, 7, false}, false},
		{"SetToZero", &DbConfig{"", "", 30015, "", "", true, 60, true, 60, true, true, 60, true, true, 60, true, db1}, args{lc, 0, false}, false},
		{"TraceQueryFails", &DbConfig{"", "", 30015, "", "", true, 60, true, 60, true, true, 60, true, true, 60, true, db1}, args{lc, 60, false}, true},
		{"TraceQueryUnscannable", &DbConfig{"", "", 30015, "", "", true, 60, true, 60, true, true, 60, true, true, 60, true, db1}, args{lc, 60, false}, true},
		{"ClearTraceFails", &DbConfig{"", "", 30015, "", "", true, 60, true, 60, true, true, 60, true, true, 60, true, db1}, args{lc, 60, false}, false},
		{"MultiTraceGood", &DbConfig{"", "", 30015, "", "", true, 60, true, 60, true, true, 60, true, true, 60, true, db1}, args{lc, 60, false}, false},
		{"MultiTraceCantDeleteFirst", &DbConfig{"", "", 30015, "", "", true, 60, true, 60, true, true, 60, true, true, 60, true, db1}, args{lc, 60, false}, false},
		{"NothingToDelete", &DbConfig{"", "", 30015, "", "", true, 60, true, 60, true, true, 60, true, true, 60, true, db1}, args{lc, 60, false}, false},
		{"RemovalRowEmpty", &DbConfig{"", "", 30015, "", "", true, 60, true, 60, true, true, 60, true, true, 60, true, db1}, args{lc, 60, false}, false},
		{"RemovalRowError", &DbConfig{"", "", 30015, "", "", true, 60, true, 60, true, true, 60, true, true, 60, true, db1}, args{lc, 60, false}, false},
		{"DryRun", &DbConfig{"", "", 30015, "", "", true, 60, true, 60, true, true, 60, true, true, 60, true, db1}, args{lc, 60, true}, false},
	}
	for _, tt := range tests {

		/*Mock switch*/
		switch {
		case tt.name[0:4] == "Good":
			rows1 := sqlmock.NewRows([]string{"HOST", "FILE_NAME", "FILE_SIZE", "FILE_MTIME"}).AddRow("hanaserver", "trace.trc", "6400000", "2020-03-14 23:13:35.000000000")
			rows2 := sqlmock.NewRows([]string{"TRACE"}).AddRow("0")
			mock.ExpectQuery(GetTraceFileQuery(tt.args.CleanDaysOlder)).WillReturnRows(rows1)
			mock.ExpectExec(GetRemoveTrace("hanaserver", "trace.trc")).WillReturnResult(sqlmock.NewResult(1, 1))
			mock.ExpectQuery(GetCheckTracePresent("trace.trc")).WillReturnRows(rows2)
		case tt.name == "SetToZero":
			//nothing to mock
		case tt.name == "TraceQueryFails":
			mock.ExpectQuery(GetTraceFileQuery(tt.args.CleanDaysOlder)).WillReturnError(fmt.Errorf("Some DB error"))
		case tt.name == "TraceQueryUnscannable":
			rows := sqlmock.NewRows([]string{"HOST", "FILE_NAME", "FILE_SIZE", "FILE_MTIME"}).AddRow("hanaserver", "trace.trc", "BAD_DATA", "2020-03-14 23:13:35.000000000")
			mock.ExpectQuery(GetTraceFileQuery(tt.args.CleanDaysOlder)).WillReturnRows(rows)
		case tt.name == "ClearTraceFails":
			rows1 := sqlmock.NewRows([]string{"HOST", "FILE_NAME", "FILE_SIZE", "FILE_MTIME"}).AddRow("hanaserver", "trace.trc", "6400000", "2020-03-14 23:13:35.000000000")
			mock.ExpectQuery(GetTraceFileQuery(tt.args.CleanDaysOlder)).WillReturnRows(rows1)
			mock.ExpectExec(GetRemoveTrace("hanaserver", "trace.trc")).WillReturnError(fmt.Errorf("Some DB error"))
		case tt.name == "MultiTraceGood":
			rows1 := sqlmock.NewRows([]string{"HOST", "FILE_NAME", "FILE_SIZE", "FILE_MTIME"}).AddRow("hanaserver", "trace.trc", "6400000", "2020-03-14 23:13:35.000000000").AddRow("hanaserver", "trace2.trc", "6400000", "2020-03-14 23:13:35.000000000")
			rows2 := sqlmock.NewRows([]string{"TRACE"}).AddRow("0")
			mock.ExpectQuery(GetTraceFileQuery(tt.args.CleanDaysOlder)).WillReturnRows(rows1)
			mock.ExpectExec(GetRemoveTrace("hanaserver", "trace.trc")).WillReturnResult(sqlmock.NewResult(1, 1))
			mock.ExpectQuery(GetCheckTracePresent("trace.trc")).WillReturnRows(rows2)
			mock.ExpectExec(GetRemoveTrace("hanaserver", "trace2.trc")).WillReturnResult(sqlmock.NewResult(1, 1))
			mock.ExpectQuery(GetCheckTracePresent("trace2.trc")).WillReturnRows(rows2)
		case tt.name == "MultiTraceCantDeleteFirst":
			rows1 := sqlmock.NewRows([]string{"HOST", "FILE_NAME", "FILE_SIZE", "FILE_MTIME"}).AddRow("hanaserver", "trace.trc", "6400000", "2020-03-14 23:13:35.000000000").AddRow("hanaserver", "trace2.trc", "6400000", "2020-03-14 23:13:35.000000000")
			rows2 := sqlmock.NewRows([]string{"TRACE"}).AddRow("1")
			rows3 := sqlmock.NewRows([]string{"TRACE"}).AddRow("0")
			mock.ExpectQuery(GetTraceFileQuery(tt.args.CleanDaysOlder)).WillReturnRows(rows1)
			mock.ExpectExec(GetRemoveTrace("hanaserver", "trace.trc")).WillReturnResult(sqlmock.NewResult(1, 1))
			mock.ExpectQuery(GetCheckTracePresent("trace.trc")).WillReturnRows(rows2)
			mock.ExpectExec(GetRemoveTrace("hanaserver", "trace2.trc")).WillReturnResult(sqlmock.NewResult(1, 1))
			mock.ExpectQuery(GetCheckTracePresent("trace2.trc")).WillReturnRows(rows3)
		case tt.name == "NothingToDelete":
			rows1 := sqlmock.NewRows([]string{"HOST", "FILE_NAME", "FILE_SIZE", "FILE_MTIME"})
			mock.ExpectQuery(GetTraceFileQuery(tt.args.CleanDaysOlder)).WillReturnRows(rows1)
		case tt.name == "RemovalRowEmpty":
			rows1 := sqlmock.NewRows([]string{"HOST", "FILE_NAME", "FILE_SIZE", "FILE_MTIME"}).AddRow("hanaserver", "traceNoRows.trc", "6400000", "2020-03-14 23:13:35.000000000")
			rows2 := sqlmock.NewRows([]string{"TRACE"})
			mock.ExpectQuery(GetTraceFileQuery(tt.args.CleanDaysOlder)).WillReturnRows(rows1)
			mock.ExpectExec(GetRemoveTrace("hanaserver", "traceNoRows.trc")).WillReturnResult(sqlmock.NewResult(1, 1))
			mock.ExpectQuery(GetCheckTracePresent("traceNoRows.trc")).WillReturnRows(rows2)
		case tt.name == "RemovalRowError":
			rows1 := sqlmock.NewRows([]string{"HOST", "FILE_NAME", "FILE_SIZE", "FILE_MTIME"}).AddRow("hanaserver", "traceNoRows.trc", "6400000", "2020-03-14 23:13:35.000000000")
			mock.ExpectQuery(GetTraceFileQuery(tt.args.CleanDaysOlder)).WillReturnRows(rows1)
			mock.ExpectExec(GetRemoveTrace("hanaserver", "traceNoRows.trc")).WillReturnResult(sqlmock.NewResult(1, 1))
			mock.ExpectQuery(GetCheckTracePresent("traceNoRows.trc")).WillReturnError(fmt.Errorf("some DB error"))
		case tt.name == "DryRun":
			rows1 := sqlmock.NewRows([]string{"HOST", "FILE_NAME", "FILE_SIZE", "FILE_MTIME"}).AddRow("hanaserver", "traceNoRows.trc", "6400000", "2020-03-14 23:13:35.000000000")
			mock.ExpectQuery(GetTraceFileQuery(tt.args.CleanDaysOlder)).WillReturnRows(rows1)
		default:
			continue
		}

		t.Run(tt.name, func(t *testing.T) {
			if err := tt.dbc.CleanTraceFilesFunc(tt.args.lc, tt.args.CleanDaysOlder, tt.args.dryrun); (err != nil) != tt.wantErr {
				t.Errorf("DbConfig.CleanTraceFiles() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
	quit <- true
}

func TestDbConfig_CleanBackupFunc(t *testing.T) {
	/*Test Setup*/
	/*Mock DB*/
	db1, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Errorf("an error '%s' was not expected when opening mock database connection", err)
	}
	defer db1.Close()

	/*Logger*/
	lc := make(chan LogMessage)
	quit := make(chan bool)
	go Logger(AppConfig{"file", true, false}, lc, quit)

	/*args*/
	type args struct {
		lc             chan<- LogMessage
		CleanDaysOlder uint
		delete         bool
		dryrun         bool
	}

	/*Tests*/
	tests := []struct {
		name    string
		dbc     *DbConfig
		args    args
		wantErr bool
	}{
		{"GoodClean", &DbConfig{"", "", 30015, "", "", true, 60, true, 60, true, true, 60, true, true, 60, true, db1}, args{lc, 60, false, false}, false},
		{"GoodDelete", &DbConfig{"", "", 30015, "", "", true, 60, true, 60, true, true, 60, true, true, 60, true, db1}, args{lc, 60, true, false}, false},
		{"QueryBackupIdFailed", &DbConfig{"", "", 30015, "", "", true, 60, true, 60, true, true, 60, true, true, 60, true, db1}, args{lc, 60, false, false}, true},
		{"QueryBackupIdNoRows", &DbConfig{"", "", 30015, "", "", true, 60, true, 60, true, true, 60, true, true, 60, true, db1}, args{lc, 60, false, false}, false},
		{"QueryFileDataFailed", &DbConfig{"", "", 30015, "", "", true, 60, true, 60, true, true, 60, true, true, 60, true, db1}, args{lc, 60, false, false}, true},
		{"NothingToDelete", &DbConfig{"", "", 30015, "", "", true, 60, true, 60, true, true, 60, true, true, 60, true, db1}, args{lc, 60, false, false}, false},
		{"CleanFailed", &DbConfig{"", "", 30015, "", "", true, 60, true, 60, true, true, 60, true, true, 60, true, db1}, args{lc, 60, false, false}, true},
		{"DeleteFailed", &DbConfig{"", "", 30015, "", "", true, 60, true, 60, true, true, 60, true, true, 60, true, db1}, args{lc, 60, true, false}, true},
	}
	for _, tt := range tests {

		/*Mock Switch*/
		switch {
		case tt.name == "GoodClean":
			var backupID string = "12345678890"
			rows1 := sqlmock.NewRows([]string{"BACKUP_ID"}).AddRow(backupID)
			rows2 := sqlmock.NewRows([]string{"ENTRY", "COUNT", "BYTES"}).AddRow("complete data backup", 10, 100000000).AddRow("log backup", 100, 100000000)
			mock.ExpectQuery(GetLatestFullBackupID(tt.args.CleanDaysOlder)).WillReturnRows(rows1)
			mock.ExpectQuery(GetBackupFileData(backupID)).WillReturnRows(rows2)
			mock.ExpectExec(GetBackupDelete(backupID)).WillReturnResult(sqlmock.NewResult(1, 1))
		case tt.name == "GoodDelete":
			var backupID string = "12345678890"
			rows1 := sqlmock.NewRows([]string{"BACKUP_ID"}).AddRow(backupID)
			rows2 := sqlmock.NewRows([]string{"ENTRY", "COUNT", "BYTES"}).AddRow("complete data backup", 10, 100000000).AddRow("log backup", 100, 100000000)
			mock.ExpectQuery(GetLatestFullBackupID(tt.args.CleanDaysOlder)).WillReturnRows(rows1)
			mock.ExpectQuery(GetBackupFileData(backupID)).WillReturnRows(rows2)
			mock.ExpectExec(GetBackupDeleteComplete(backupID)).WillReturnResult(sqlmock.NewResult(1, 1))
		case tt.name == "QueryBackupIdFailed":
			mock.ExpectQuery(GetLatestFullBackupID(tt.args.CleanDaysOlder)).WillReturnError(fmt.Errorf("some DB error"))
		case tt.name == "QueryBackupIdNoRows":
			mock.ExpectQuery(GetLatestFullBackupID(tt.args.CleanDaysOlder)).WillReturnError(sql.ErrNoRows)
		case tt.name == "QueryFileDataFailed":
			var backupID string = "12345678890"
			rows1 := sqlmock.NewRows([]string{"BACKUP_ID"}).AddRow(backupID)
			mock.ExpectQuery(GetLatestFullBackupID(tt.args.CleanDaysOlder)).WillReturnRows(rows1)
			mock.ExpectQuery(GetBackupFileData(backupID)).WillReturnError(fmt.Errorf("Some DB error"))
		case tt.name == "NothingToDelete":
			var backupID string = "12345678890"
			rows1 := sqlmock.NewRows([]string{"BACKUP_ID"}).AddRow(backupID)
			rows2 := sqlmock.NewRows([]string{"ENTRY", "COUNT", "BYTES"})
			mock.ExpectQuery(GetLatestFullBackupID(tt.args.CleanDaysOlder)).WillReturnRows(rows1)
			mock.ExpectQuery(GetBackupFileData(backupID)).WillReturnRows(rows2)
		case tt.name == "CleanFailed":
			var backupID string = "12345678890"
			rows1 := sqlmock.NewRows([]string{"BACKUP_ID"}).AddRow(backupID)
			rows2 := sqlmock.NewRows([]string{"ENTRY", "COUNT", "BYTES"}).AddRow("complete data backup", 10, 100000000).AddRow("log backup", 100, 100000000)
			mock.ExpectQuery(GetLatestFullBackupID(tt.args.CleanDaysOlder)).WillReturnRows(rows1)
			mock.ExpectQuery(GetBackupFileData(backupID)).WillReturnRows(rows2)
			mock.ExpectExec(GetBackupDelete(backupID)).WillReturnError(fmt.Errorf("Some DB error"))
		case tt.name == "DeleteFailed":
			var backupID string = "12345678890"
			rows1 := sqlmock.NewRows([]string{"BACKUP_ID"}).AddRow(backupID)
			rows2 := sqlmock.NewRows([]string{"ENTRY", "COUNT", "BYTES"}).AddRow("complete data backup", 10, 100000000).AddRow("log backup", 100, 100000000)
			mock.ExpectQuery(GetLatestFullBackupID(tt.args.CleanDaysOlder)).WillReturnRows(rows1)
			mock.ExpectQuery(GetBackupFileData(backupID)).WillReturnRows(rows2)
			mock.ExpectExec(GetBackupDeleteComplete(backupID)).WillReturnError(fmt.Errorf("Some DB error"))
		default:
			continue

		}

		t.Run(tt.name, func(t *testing.T) {
			if err := tt.dbc.CleanBackupFunc(tt.args.lc, tt.args.CleanDaysOlder, tt.args.delete, tt.args.dryrun); (err != nil) != tt.wantErr {
				t.Errorf("DbConfig.CleanBackupFunc() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

//	tests := []struct {
//		name    string
//		args    args
//		wantErr bool
//	}{
//		{"Get Backup Delete Complete Failed", args{lc, "db@sid", db1, 28, true, false}, true},
//		{"Get Backup Delete Failed", args{lc, "db@sid", db1, 28, false, false}, true},
//	}
//	for _, tt := range tests {
//
//			} else if tt.name == "Get Backup Delete Complete Failed" {
//			var backupID string = "12345678890"
//			rows1 := sqlmock.NewRows([]string{"BACKUP_ID"}).AddRow(backupID)
//			rows2 := sqlmock.NewRows([]string{"ENTRY", "COUNT", "BYTES"}).AddRow("complete data backup", 10, 100000000).AddRow("log backup", 100, 100000000)
//			mock.ExpectQuery(GetLatestFullBackupID(tt.args.TrncDaysOlder)).WillReturnRows(rows1)
//			mock.ExpectQuery(GetBackupFileData(backupID)).WillReturnRows(rows2)
//			mock.ExpectExec(GetBackupDeleteComplete(backupID)).WillReturnError(fmt.Errorf("Some DB error"))
//		} else if tt.name == "Get Backup Delete Failed" {
//			var backupID string = "12345678890"
//			rows1 := sqlmock.NewRows([]string{"BACKUP_ID"}).AddRow(backupID)
//			rows2 := sqlmock.NewRows([]string{"ENTRY", "COUNT", "BYTES"}).AddRow("complete data backup", 10, 100000000).AddRow("log backup", 100, 100000000)
//			mock.ExpectQuery(GetLatestFullBackupID(tt.args.TrncDaysOlder)).WillReturnRows(rows1)
//			mock.ExpectQuery(GetBackupFileData(backupID)).WillReturnRows(rows2)
//			mock.ExpectExec(GetBackupDelete(backupID)).WillReturnError(fmt.Errorf("Some DB error"))
//		} else {
//			return
//		}
//
//		t.Run(tt.name, func(t *testing.T) {
//			if err := TruncateBackupCatalog(tt.args.lc, tt.args.name, tt.args.hdb, tt.args.TrncDaysOlder, tt.args.delete, tt.args.dryrun); (err != nil) != tt.wantErr {
//				t.Errorf("TruncateBackupCatalog() error = %v, wantErr %v", err, tt.wantErr)
//			}
//		})
//	}
//	quit <- true
//}

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
			mock.ExpectExec(QUERY_ReclaimLog).WillReturnResult(sqlmock.NewResult(1, 1))
		case tt.name == "GetSegmentsDbError":
			rows1 := sqlmock.NewRows([]string{"COUNT", "BYTES"})
			mock.ExpectQuery(QUERY_GetFeeLogSegments).WillReturnRows(rows1)
		case tt.name == "GetSegmentsDbError":
			mock.ExpectQuery(QUERY_GetFeeLogSegments).WillReturnError(fmt.Errorf("some db Error"))
		case tt.name == "ReclaimLogDbError":
			rows1 := sqlmock.NewRows([]string{"COUNT", "BYTES"}).AddRow("10", "2048000")
			mock.ExpectQuery(QUERY_GetFeeLogSegments).WillReturnRows(rows1)
			mock.ExpectExec(QUERY_ReclaimLog).WillReturnError(fmt.Errorf("some DB error"))
		}

		t.Run(tt.name, func(t *testing.T) {
			if err := ReclaimLog(tt.args.lc, tt.args.name, tt.args.hdb, tt.args.dryrun); (err != nil) != tt.wantErr {
				t.Errorf("ReclaimLog() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	quit <- true
}

func TestTruncateAuditLog(t *testing.T) {

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
		lc         chan<- LogMessage
		name       string
		hdb        *sql.DB
		daysToKeep uint
		dryrun     bool
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"Good", args{lc, "db@sid", db1, 14, false}, false},
		{"GoodOneRecordToDelete", args{lc, "db@sid", db1, 14, false}, false},
		{"GoodNothingToDelete", args{lc, "db@sid", db1, 14, false}, false},
		{"CountEventsNoRows", args{lc, "db@sid", db1, 14, false}, true},
		{"CountEventsDbError", args{lc, "db@sid", db1, 14, false}, true},
		{"GetDateNoRows", args{lc, "db@sid", db1, 14, false}, true},
		{"GetDateDbError", args{lc, "db@sid", db1, 14, false}, true},
		{"GetDateWrongFormat", args{lc, "db@sid", db1, 14, false}, true},
		{"TruncateFailed", args{lc, "db@sid", db1, 14, false}, true},
	}

	for _, tt := range tests {

		switch {
		case tt.name == "Good":
			rows1 := sqlmock.NewRows([]string{"COUNT"}).AddRow("10")
			rows2 := sqlmock.NewRows([]string{"NOW"}).AddRow("2022-01-01 10:00:00.431000000")
			mock.ExpectQuery(GetAuditCount(tt.args.daysToKeep)).WillReturnRows(rows1)
			mock.ExpectQuery(GetDatetime(tt.args.daysToKeep)).WillReturnRows(rows2)
			mock.ExpectExec(GetTruncateAuditLog("2022-01-01 10:00:00")).WillReturnResult(sqlmock.NewResult(0, 0))
		case tt.name == "GoodOneRecordToDelete":
			rows1 := sqlmock.NewRows([]string{"COUNT"}).AddRow("1")
			rows2 := sqlmock.NewRows([]string{"NOW"}).AddRow("2022-01-01 10:00:00.431000000")
			mock.ExpectQuery(GetAuditCount(tt.args.daysToKeep)).WillReturnRows(rows1)
			mock.ExpectQuery(GetDatetime(tt.args.daysToKeep)).WillReturnRows(rows2)
			mock.ExpectExec(GetTruncateAuditLog("2022-01-01 10:00:00")).WillReturnResult(sqlmock.NewResult(0, 0))
		case tt.name == "GoodNothingToDelete":
			rows1 := sqlmock.NewRows([]string{"COUNT"}).AddRow("0")
			mock.ExpectQuery(GetAuditCount(tt.args.daysToKeep)).WillReturnRows(rows1)
		case tt.name == "CountEventsNoRows":
			rows1 := sqlmock.NewRows([]string{"COUNT"})
			mock.ExpectQuery(GetAuditCount(tt.args.daysToKeep)).WillReturnRows(rows1)
		case tt.name == "CountEventsDbError":
			mock.ExpectQuery(GetAuditCount(tt.args.daysToKeep)).WillReturnError(fmt.Errorf("some db error"))
		case tt.name == "GetDateNoRows":
			rows1 := sqlmock.NewRows([]string{"COUNT"}).AddRow("10")
			rows2 := sqlmock.NewRows([]string{"NOW"})
			mock.ExpectQuery(GetAuditCount(tt.args.daysToKeep)).WillReturnRows(rows1)
			mock.ExpectQuery(GetDatetime(tt.args.daysToKeep)).WillReturnRows(rows2)
		case tt.name == "GetDateDbError":
			rows1 := sqlmock.NewRows([]string{"COUNT"}).AddRow("10")
			mock.ExpectQuery(GetAuditCount(tt.args.daysToKeep)).WillReturnRows(rows1)
			mock.ExpectQuery(GetDatetime(tt.args.daysToKeep)).WillReturnError(fmt.Errorf("some db error"))
		case tt.name == "GetDateWrongFormat":
			rows1 := sqlmock.NewRows([]string{"COUNT"}).AddRow("10")
			rows2 := sqlmock.NewRows([]string{"NOW"}).AddRow("2022-01-01 10:00:00.431000000.0000")
			mock.ExpectQuery(GetAuditCount(tt.args.daysToKeep)).WillReturnRows(rows1)
			mock.ExpectQuery(GetDatetime(tt.args.daysToKeep)).WillReturnRows(rows2)
		case tt.name == "TruncateFailed":
			rows1 := sqlmock.NewRows([]string{"COUNT"}).AddRow("10")
			rows2 := sqlmock.NewRows([]string{"NOW"}).AddRow("2022-01-01 10:00:00.431000000")
			mock.ExpectQuery(GetAuditCount(tt.args.daysToKeep)).WillReturnRows(rows1)
			mock.ExpectQuery(GetDatetime(tt.args.daysToKeep)).WillReturnRows(rows2)
			mock.ExpectExec(GetTruncateAuditLog("2022-01-01 10:00:00")).WillReturnError(fmt.Errorf("some db error"))
		}

		t.Run(tt.name, func(t *testing.T) {
			if err := TruncateAuditLog(tt.args.lc, tt.args.name, tt.args.hdb, tt.args.daysToKeep, tt.args.dryrun); (err != nil) != tt.wantErr {
				t.Errorf("TruncateAuditLog() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	quit <- true
}

func TestCleanDataVolume(t *testing.T) {

	db1, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Errorf("an error '%s' was not expected when opening mock database connection", err)
	}
	defer db1.Close()

	lc := make(chan LogMessage)
	quit := make(chan bool)

	go Logger(AppConfig{"file", false, false}, lc, quit)

	defer close(lc)
	defer close(quit)

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
		{"GoodClean", args{lc, "ten_db", db1, false}, false},
		{"GoodNoCleanNeeded", args{lc, "ten_db", db1, false}, false},
		{"DataVolumeQueryFail", args{lc, "ten_db", db1, false}, true},
		{"NoDataVolumes", args{lc, "ten_db", db1, false}, false}, //no data volumes doesn't return an error, perhaps it should
		{"ScanError", args{lc, "ten_db", db1, false}, true},
		{"SingleCleanFailed", args{lc, "ten_db", db1, false}, true},
		{"GoodCleanDryRun", args{lc, "ten_db", db1, true}, false},
		{"GoodCleanTwoVolumes", args{lc, "ten_db", db1, false}, false},
		{"CleanTwoVolumesOneFails", args{lc, "ten_db", db1, true}, false},
	}

	for _, tt := range tests {
		switch {
		case tt.name == "GoodClean":
			rows1 := sqlmock.NewRows([]string{"HOST", "PORT", "USED_SIZE", "TOTAL_SIZE"}).AddRow("testhana", "30040", "1000000", "3000000")
			mock.ExpectQuery(QUERY_GetDataVolume).WillReturnRows(rows1)
			mock.ExpectExec(GetCleanDataVolume("testhana", 30040)).WillReturnResult(sqlmock.NewResult(0, 0))
		case tt.name == "GoodNoCleanNeeded":
			rows1 := sqlmock.NewRows([]string{"HOST", "PORT", "USED_SIZE", "TOTAL_SIZE"}).AddRow("testhana", "30040", "2000000", "3000000")
			mock.ExpectQuery(QUERY_GetDataVolume).WillReturnRows(rows1)
		case tt.name == "DataVolumeQueryFail":
			mock.ExpectQuery(QUERY_GetDataVolume).WillReturnError(fmt.Errorf("some db error"))
		case tt.name == "NoDataVolumes":
			rows1 := sqlmock.NewRows([]string{"HOST", "PORT", "USED_SIZE", "TOTAL_SIZE"})
			mock.ExpectQuery(QUERY_GetDataVolume).WillReturnRows(rows1)
		case tt.name == "ScanError":
			rows1 := sqlmock.NewRows([]string{"HOST", "PORT", "USED_SIZE", "TOTAL_SIZE"}).AddRow("testhana", "30040.12", "1000000", "3000000")
			mock.ExpectQuery(QUERY_GetDataVolume).WillReturnRows(rows1)
		case tt.name == "SingleCleanFailed":
			rows1 := sqlmock.NewRows([]string{"HOST", "PORT", "USED_SIZE", "TOTAL_SIZE"}).AddRow("testhana", "30040", "1000000", "3000000")
			mock.ExpectQuery(QUERY_GetDataVolume).WillReturnRows(rows1)
			mock.ExpectExec(GetCleanDataVolume("testhana", 30040)).WillReturnError(fmt.Errorf("some Db error"))
		case tt.name == "GoodCleanDryRun":
			rows1 := sqlmock.NewRows([]string{"HOST", "PORT", "USED_SIZE", "TOTAL_SIZE"}).AddRow("testhana", "30040", "1000000", "3000000")
			mock.ExpectQuery(QUERY_GetDataVolume).WillReturnRows(rows1)
		case tt.name == "GoodCleanTwoVolumes":
			rows1 := sqlmock.NewRows([]string{"HOST", "PORT", "USED_SIZE", "TOTAL_SIZE"}).AddRow("testhana", "30040", "1000000", "3000000").AddRow("testhana", "30044", "2000000", "6000000")
			mock.ExpectQuery(QUERY_GetDataVolume).WillReturnRows(rows1)
			mock.ExpectExec(GetCleanDataVolume("testhana", 30040)).WillReturnResult(sqlmock.NewResult(0, 0))
			mock.ExpectExec(GetCleanDataVolume("testhana", 30044)).WillReturnResult(sqlmock.NewResult(0, 0))
		case tt.name == "CleanTwoVolumesOneFails":
			rows1 := sqlmock.NewRows([]string{"HOST", "PORT", "USED_SIZE", "TOTAL_SIZE"}).AddRow("testhana", "30040", "1000000", "3000000").AddRow("testhana", "30044", "2000000", "6000000")
			mock.ExpectQuery(QUERY_GetDataVolume).WillReturnRows(rows1)
			mock.ExpectExec(GetCleanDataVolume("testhana", 30040)).WillReturnResult(sqlmock.NewResult(0, 0))
			mock.ExpectExec(GetCleanDataVolume("testhana", 30044)).WillReturnError(fmt.Errorf("some db error"))
		default:
			//nothing to do
			continue
		}
		t.Run(tt.name, func(t *testing.T) {
			if err := CleanDataVolume(tt.args.lc, tt.args.name, tt.args.hdb, tt.args.dryrun); (err != nil) != tt.wantErr {
				t.Errorf("CleanDataVolume() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
	quit <- true

}
