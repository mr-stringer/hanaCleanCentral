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

	defer close(lc)
	defer close(quit)

	go Logger(AppConfig{"file", true, false, false}, lc, quit)

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
		{"Good01", &DbConfig{"", "", 30015, "", "", true, 60, true, 60, true, true, 60, true, true, 60, true, db1, CleanResults{}}, args{lc}, "2.00.45.00", false},
		{"Good02", &DbConfig{"", "", 30015, "", "", true, 60, true, 60, true, true, 60, true, true, 60, true, db1, CleanResults{}}, args{lc}, "2.00.43.33", false},
		{"Good03", &DbConfig{"", "", 30015, "", "", true, 60, true, 60, true, true, 60, true, true, 60, true, db1, CleanResults{}}, args{lc}, "3.00.00.10", false},
		{"Good04", &DbConfig{"", "", 30015, "", "", true, 60, true, 60, true, true, 60, true, true, 60, true, db1, CleanResults{}}, args{lc}, "1.00.112.3", false},
		{"DbError", &DbConfig{"", "", 30015, "", "", true, 60, true, 60, true, true, 60, true, true, 60, true, db1, CleanResults{}}, args{lc}, "", true},
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

	defer close(lc)
	defer close(quit)

	go Logger(AppConfig{"file", true, false, false}, lc, quit)

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
		{"Good01", &DbConfig{"", "", 30015, "", "", true, 60, true, 60, true, true, 60, true, true, 60, true, db1, CleanResults{}}, args{lc, 60, false}, false},
		{"Good02", &DbConfig{"", "", 30015, "", "", true, 60, true, 60, true, true, 60, true, true, 60, true, db1, CleanResults{}}, args{lc, 14, false}, false},
		{"Good03", &DbConfig{"", "", 30015, "", "", true, 60, true, 60, true, true, 60, true, true, 60, true, db1, CleanResults{}}, args{lc, 7, false}, false},
		{"TraceQueryFails", &DbConfig{"", "", 30015, "", "", true, 60, true, 60, true, true, 60, true, true, 60, true, db1, CleanResults{}}, args{lc, 60, false}, true},
		{"TraceQueryUnscannable", &DbConfig{"", "", 30015, "", "", true, 60, true, 60, true, true, 60, true, true, 60, true, db1, CleanResults{}}, args{lc, 60, false}, true},
		{"ClearTraceFails", &DbConfig{"", "", 30015, "", "", true, 60, true, 60, true, true, 60, true, true, 60, true, db1, CleanResults{}}, args{lc, 60, false}, false},
		{"MultiTraceGood", &DbConfig{"", "", 30015, "", "", true, 60, true, 60, true, true, 60, true, true, 60, true, db1, CleanResults{}}, args{lc, 60, false}, false},
		{"MultiTraceCantDeleteFirst", &DbConfig{"", "", 30015, "", "", true, 60, true, 60, true, true, 60, true, true, 60, true, db1, CleanResults{}}, args{lc, 60, false}, false},
		{"NothingToDelete", &DbConfig{"", "", 30015, "", "", true, 60, true, 60, true, true, 60, true, true, 60, true, db1, CleanResults{}}, args{lc, 60, false}, false},
		{"RemovalRowEmpty", &DbConfig{"", "", 30015, "", "", true, 60, true, 60, true, true, 60, true, true, 60, true, db1, CleanResults{}}, args{lc, 60, false}, false},
		{"RemovalRowError", &DbConfig{"", "", 30015, "", "", true, 60, true, 60, true, true, 60, true, true, 60, true, db1, CleanResults{}}, args{lc, 60, false}, false},
		{"DryRun", &DbConfig{"", "", 30015, "", "", true, 60, true, 60, true, true, 60, true, true, 60, true, db1, CleanResults{}}, args{lc, 60, true}, false},
	}
	for _, tt := range tests {

		/*Set up per case mocking*/
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
			t.Errorf("Couldn't find DB mocking for test \"%s\"\n", tt.name)
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

	defer close(lc)
	defer close(quit)

	go Logger(AppConfig{"file", true, false, false}, lc, quit)

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
		{"GoodClean", &DbConfig{"", "", 30015, "", "", true, 60, true, 60, true, true, 60, true, true, 60, true, db1, CleanResults{}}, args{lc, 60, false, false}, false},
		{"GoodDelete", &DbConfig{"", "", 30015, "", "", true, 60, true, 60, true, true, 60, true, true, 60, true, db1, CleanResults{}}, args{lc, 60, true, false}, false},
		{"QueryBackupIdFailed", &DbConfig{"", "", 30015, "", "", true, 60, true, 60, true, true, 60, true, true, 60, true, db1, CleanResults{}}, args{lc, 60, false, false}, true},
		{"QueryBackupIdNoRows", &DbConfig{"", "", 30015, "", "", true, 60, true, 60, true, true, 60, true, true, 60, true, db1, CleanResults{}}, args{lc, 60, false, false}, false},
		{"QueryFileDataFailed", &DbConfig{"", "", 30015, "", "", true, 60, true, 60, true, true, 60, true, true, 60, true, db1, CleanResults{}}, args{lc, 60, false, false}, true},
		{"NothingToDelete", &DbConfig{"", "", 30015, "", "", true, 60, true, 60, true, true, 60, true, true, 60, true, db1, CleanResults{}}, args{lc, 60, false, false}, false},
		{"CleanFailed", &DbConfig{"", "", 30015, "", "", true, 60, true, 60, true, true, 60, true, true, 60, true, db1, CleanResults{}}, args{lc, 60, false, false}, true},
		{"DeleteFailed", &DbConfig{"", "", 30015, "", "", true, 60, true, 60, true, true, 60, true, true, 60, true, db1, CleanResults{}}, args{lc, 60, true, false}, true},
	}
	for _, tt := range tests {

		/*Set up per case mocking*/
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
			t.Errorf("Couldn't find DB mocking for test \"%s\"\n", tt.name)
		}

		t.Run(tt.name, func(t *testing.T) {
			if err := tt.dbc.CleanBackupFunc(tt.args.lc, tt.args.CleanDaysOlder, tt.args.delete, tt.args.dryrun); (err != nil) != tt.wantErr {
				t.Errorf("DbConfig.CleanBackupFunc() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDbConfig_CleanAlertFunc(t *testing.T) {
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

	defer close(lc)
	defer close(quit)

	go Logger(AppConfig{"file", true, false, false}, lc, quit)

	/*args*/
	type args struct {
		lc             chan<- LogMessage
		CleanDaysOlder uint
		dryrun         bool
	}

	/*tests*/
	tests := []struct {
		name    string
		dbc     *DbConfig
		args    args
		wantErr bool
	}{
		{"Good01", &DbConfig{"", "", 30015, "", "", true, 60, true, 60, true, true, 60, true, true, 60, true, db1, CleanResults{}}, args{lc, 14, false}, false},
		{"DryRun", &DbConfig{"", "", 30015, "", "", true, 60, true, 60, true, true, 60, true, true, 60, true, db1, CleanResults{}}, args{lc, 14, true}, false},
		{"CountAlertsNoRows", &DbConfig{"", "", 30015, "", "", true, 60, true, 60, true, true, 60, true, true, 60, true, db1, CleanResults{}}, args{lc, 14, false}, true},
		{"CountAlertsDbError", &DbConfig{"", "", 30015, "", "", true, 60, true, 60, true, true, 60, true, true, 60, true, db1, CleanResults{}}, args{lc, 14, false}, true},
		{"NothingToDo", &DbConfig{"", "", 30015, "", "", true, 60, true, 60, true, true, 60, true, true, 60, true, db1, CleanResults{}}, args{lc, 14, false}, false},
		{"CleanAlertsDbError", &DbConfig{"", "", 30015, "", "", true, 60, true, 60, true, true, 60, true, true, 60, true, db1, CleanResults{}}, args{lc, 14, false}, true},
	}
	for _, tt := range tests {
		/*Set up per case mocking*/
		switch {
		case tt.name == "Good01":
			rows1 := sqlmock.NewRows([]string{"COUNT"}).AddRow("250")
			mock.ExpectQuery(GetAlertCount(tt.args.CleanDaysOlder)).WillReturnRows(rows1)
			mock.ExpectExec(GetAlertDelete(tt.args.CleanDaysOlder)).WillReturnResult(sqlmock.NewResult(1, 1))
		case tt.name == "DryRun":
			rows1 := sqlmock.NewRows([]string{"COUNT"}).AddRow("250")
			mock.ExpectQuery(GetAlertCount(tt.args.CleanDaysOlder)).WillReturnRows(rows1)
		case tt.name == "CountAlertsNoRows":
			rows1 := sqlmock.NewRows([]string{"COUNT"})
			mock.ExpectQuery(GetAlertCount(tt.args.CleanDaysOlder)).WillReturnRows(rows1)
		case tt.name == "CountAlertsDbError":
			mock.ExpectQuery(GetAlertCount(tt.args.CleanDaysOlder)).WillReturnError(fmt.Errorf("some DB error"))
		case tt.name == "NothingToDo":
			rows1 := sqlmock.NewRows([]string{"COUNT"}).AddRow("0")
			mock.ExpectQuery(GetAlertCount(tt.args.CleanDaysOlder)).WillReturnRows(rows1)
		case tt.name == "CleanAlertsDbError":
			rows1 := sqlmock.NewRows([]string{"COUNT"}).AddRow("250")
			mock.ExpectQuery(GetAlertCount(tt.args.CleanDaysOlder)).WillReturnRows(rows1)
			mock.ExpectExec(GetAlertDelete(tt.args.CleanDaysOlder)).WillReturnError(fmt.Errorf("some DB error"))
		default:
			t.Errorf("Couldn't find DB mocking for test \"%s\"\n", tt.name)
		}

		t.Run(tt.name, func(t *testing.T) {
			if err := tt.dbc.CleanAlertFunc(tt.args.lc, tt.args.CleanDaysOlder, tt.args.dryrun); (err != nil) != tt.wantErr {
				t.Errorf("DbConfig.CleanAlertFunc() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
	quit <- true
}

func TestDbConfig_CleanLogFunc(t *testing.T) {
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

	defer close(lc)
	defer close(quit)

	go Logger(AppConfig{"file", true, false, false}, lc, quit)

	/*Args*/
	type args struct {
		lc     chan<- LogMessage
		dryrun bool
	}

	/*Tests*/
	tests := []struct {
		name    string
		dbc     *DbConfig
		args    args
		wantErr bool
	}{
		{"Good01", &DbConfig{"", "", 30015, "", "", true, 60, true, 60, true, true, 60, true, true, 60, true, db1, CleanResults{}}, args{lc, false}, false},
		{"DryRun", &DbConfig{"", "", 30015, "", "", true, 60, true, 60, true, true, 60, true, true, 60, true, db1, CleanResults{}}, args{lc, true}, false},
		{"GetSegmentsNoRows", &DbConfig{"", "", 30015, "", "", true, 60, true, 60, true, true, 60, true, true, 60, true, db1, CleanResults{}}, args{lc, false}, true},
		{"GetSegmentsDbError", &DbConfig{"", "", 30015, "", "", true, 60, true, 60, true, true, 60, true, true, 60, true, db1, CleanResults{}}, args{lc, false}, true},
		{"ReclaimDbError", &DbConfig{"", "", 30015, "", "", true, 60, true, 60, true, true, 60, true, true, 60, true, db1, CleanResults{}}, args{lc, false}, true},
	}
	for _, tt := range tests {
		/*Set up per case mocking*/
		switch {
		case tt.name == "Good01":
			rows1 := sqlmock.NewRows([]string{"COUNT", "BYTES"}).AddRow("10", "2048000")
			mock.ExpectQuery(QUERY_GetFeeLogSegments).WillReturnRows(rows1)
			mock.ExpectExec(QUERY_ReclaimLog).WillReturnResult(sqlmock.NewResult(1, 1))
		case tt.name == "DryRun":
			rows1 := sqlmock.NewRows([]string{"COUNT", "BYTES"}).AddRow("10", "2048000")
			mock.ExpectQuery(QUERY_GetFeeLogSegments).WillReturnRows(rows1)
		case tt.name == "GetSegmentsNoRows":
			rows1 := sqlmock.NewRows([]string{"COUNT", "BYTES"})
			mock.ExpectQuery(QUERY_GetFeeLogSegments).WillReturnRows(rows1)
		case tt.name == "GetSegmentsDbError":
			mock.ExpectQuery(QUERY_GetFeeLogSegments).WillReturnError(fmt.Errorf("some DB error"))
		case tt.name == "ReclaimDbError":
			rows1 := sqlmock.NewRows([]string{"COUNT", "BYTES"}).AddRow("10", "2048000")
			mock.ExpectQuery(QUERY_GetFeeLogSegments).WillReturnRows(rows1)
			mock.ExpectExec(QUERY_ReclaimLog).WillReturnError(fmt.Errorf("some DB error"))
		default:
			t.Errorf("Couldn't find DB mocking for test \"%s\"\n", tt.name)
		}
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.dbc.CleanLogFunc(tt.args.lc, tt.args.dryrun); (err != nil) != tt.wantErr {
				t.Errorf("DbConfig.CleanLogFunc() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
	quit <- true
}

func TestDbConfig_CleanAuditFunc(t *testing.T) {
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

	defer close(lc)
	defer close(quit)

	go Logger(AppConfig{"file", true, false, false}, lc, quit)

	/*args*/
	type args struct {
		lc             chan<- LogMessage
		CleanDaysOlder uint
		dryrun         bool
	}

	/*tests*/
	tests := []struct {
		name    string
		dbc     *DbConfig
		args    args
		wantErr bool
	}{
		{"Good", &DbConfig{"", "", 30015, "", "", true, 60, true, 60, true, true, 60, true, true, 60, true, db1, CleanResults{}}, args{lc, 60, false}, false},
		{"DryRun", &DbConfig{"", "", 30015, "", "", true, 60, true, 60, true, true, 60, true, true, 60, true, db1, CleanResults{}}, args{lc, 60, true}, false},
		{"GoodOneRecordToDelete", &DbConfig{"", "", 30015, "", "", true, 60, true, 60, true, true, 60, true, true, 60, true, db1, CleanResults{}}, args{lc, 60, false}, false},
		{"GoodNothingToDelete", &DbConfig{"", "", 30015, "", "", true, 60, true, 60, true, true, 60, true, true, 60, true, db1, CleanResults{}}, args{lc, 60, false}, false},
		{"CountEventsNoRows", &DbConfig{"", "", 30015, "", "", true, 60, true, 60, true, true, 60, true, true, 60, true, db1, CleanResults{}}, args{lc, 60, false}, true},
		{"CountEventsDbError", &DbConfig{"", "", 30015, "", "", true, 60, true, 60, true, true, 60, true, true, 60, true, db1, CleanResults{}}, args{lc, 60, false}, true},
		{"GetDateNoRows", &DbConfig{"", "", 30015, "", "", true, 60, true, 60, true, true, 60, true, true, 60, true, db1, CleanResults{}}, args{lc, 60, false}, true},
		{"GetDateDbError", &DbConfig{"", "", 30015, "", "", true, 60, true, 60, true, true, 60, true, true, 60, true, db1, CleanResults{}}, args{lc, 60, false}, true},
		{"GetDateWrongFormat", &DbConfig{"", "", 30015, "", "", true, 60, true, 60, true, true, 60, true, true, 60, true, db1, CleanResults{}}, args{lc, 60, false}, true},
		{"TruncateFailed", &DbConfig{"", "", 30015, "", "", true, 60, true, 60, true, true, 60, true, true, 60, true, db1, CleanResults{}}, args{lc, 60, false}, true},
	}
	for _, tt := range tests {
		/*Set up per case mocking*/
		switch {
		case tt.name == "Good":
			rows1 := sqlmock.NewRows([]string{"COUNT"}).AddRow("10")
			rows2 := sqlmock.NewRows([]string{"NOW"}).AddRow("2022-01-01 10:00:00.431000000")
			mock.ExpectQuery(GetAuditCount(tt.args.CleanDaysOlder)).WillReturnRows(rows1)
			mock.ExpectQuery(GetDatetime(tt.args.CleanDaysOlder)).WillReturnRows(rows2)
			mock.ExpectExec(GetTruncateAuditLog("2022-01-01 10:00:00")).WillReturnResult(sqlmock.NewResult(0, 0))
		case tt.name == "DryRun":
			rows1 := sqlmock.NewRows([]string{"COUNT"}).AddRow("10")
			rows2 := sqlmock.NewRows([]string{"NOW"}).AddRow("2022-01-01 10:00:00.431000000")
			mock.ExpectQuery(GetAuditCount(tt.args.CleanDaysOlder)).WillReturnRows(rows1)
			mock.ExpectQuery(GetDatetime(tt.args.CleanDaysOlder)).WillReturnRows(rows2)
		case tt.name == "GoodOneRecordToDelete":
			rows1 := sqlmock.NewRows([]string{"COUNT"}).AddRow("1")
			rows2 := sqlmock.NewRows([]string{"NOW"}).AddRow("2022-01-01 10:00:00.431000000")
			mock.ExpectQuery(GetAuditCount(tt.args.CleanDaysOlder)).WillReturnRows(rows1)
			mock.ExpectQuery(GetDatetime(tt.args.CleanDaysOlder)).WillReturnRows(rows2)
			mock.ExpectExec(GetTruncateAuditLog("2022-01-01 10:00:00")).WillReturnResult(sqlmock.NewResult(0, 0))
		case tt.name == "GoodNothingToDelete":
			rows1 := sqlmock.NewRows([]string{"COUNT"}).AddRow("0")
			mock.ExpectQuery(GetAuditCount(tt.args.CleanDaysOlder)).WillReturnRows(rows1)
		case tt.name == "CountEventsNoRows":
			rows1 := sqlmock.NewRows([]string{"COUNT"})
			mock.ExpectQuery(GetAuditCount(tt.args.CleanDaysOlder)).WillReturnRows(rows1)
		case tt.name == "CountEventsDbError":
			mock.ExpectQuery(GetAuditCount(tt.args.CleanDaysOlder)).WillReturnError(fmt.Errorf("some db error"))
		case tt.name == "GetDateNoRows":
			rows1 := sqlmock.NewRows([]string{"COUNT"}).AddRow("10")
			rows2 := sqlmock.NewRows([]string{"NOW"})
			mock.ExpectQuery(GetAuditCount(tt.args.CleanDaysOlder)).WillReturnRows(rows1)
			mock.ExpectQuery(GetDatetime(tt.args.CleanDaysOlder)).WillReturnRows(rows2)
		case tt.name == "GetDateDbError":
			rows1 := sqlmock.NewRows([]string{"COUNT"}).AddRow("10")
			mock.ExpectQuery(GetAuditCount(tt.args.CleanDaysOlder)).WillReturnRows(rows1)
			mock.ExpectQuery(GetDatetime(tt.args.CleanDaysOlder)).WillReturnError(fmt.Errorf("some db error"))
		case tt.name == "GetDateWrongFormat":
			rows1 := sqlmock.NewRows([]string{"COUNT"}).AddRow("10")
			rows2 := sqlmock.NewRows([]string{"NOW"}).AddRow("2022-01-01 10:00:00.431000000.0000")
			mock.ExpectQuery(GetAuditCount(tt.args.CleanDaysOlder)).WillReturnRows(rows1)
			mock.ExpectQuery(GetDatetime(tt.args.CleanDaysOlder)).WillReturnRows(rows2)
		case tt.name == "TruncateFailed":
			rows1 := sqlmock.NewRows([]string{"COUNT"}).AddRow("10")
			rows2 := sqlmock.NewRows([]string{"NOW"}).AddRow("2022-01-01 10:00:00.431000000")
			mock.ExpectQuery(GetAuditCount(tt.args.CleanDaysOlder)).WillReturnRows(rows1)
			mock.ExpectQuery(GetDatetime(tt.args.CleanDaysOlder)).WillReturnRows(rows2)
			mock.ExpectExec(GetTruncateAuditLog("2022-01-01 10:00:00")).WillReturnError(fmt.Errorf("some db error"))
		default:
			t.Errorf("Couldn't find DB mocking for test \"%s\"\n", tt.name)
		}
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.dbc.CleanAuditFunc(tt.args.lc, tt.args.CleanDaysOlder, tt.args.dryrun); (err != nil) != tt.wantErr {
				t.Errorf("DbConfig.CleanAuditFunc() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDbConfig_CleanDataVolumeFunc(t *testing.T) {
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

	defer close(lc)
	defer close(quit)

	go Logger(AppConfig{"file", true, false, false}, lc, quit)

	/*args*/
	type args struct {
		lc     chan<- LogMessage
		dryrun bool
	}

	/*tests*/
	tests := []struct {
		name    string
		dbc     *DbConfig
		args    args
		wantErr bool
	}{
		{"GoodClean", &DbConfig{"", "", 30015, "", "", true, 60, true, 60, true, true, 60, true, true, 60, true, db1, CleanResults{}}, args{lc, false}, false},
		{"GoodCleanPostCheckFails", &DbConfig{"", "", 30015, "", "", true, 60, true, 60, true, true, 60, true, true, 60, true, db1, CleanResults{}}, args{lc, false}, false},
		{"GoodNoCleanNeeded", &DbConfig{"", "", 30015, "", "", true, 60, true, 60, true, true, 60, true, true, 60, true, db1, CleanResults{}}, args{lc, false}, false},
		{"DataVolumeQueryFail", &DbConfig{"", "", 30015, "", "", true, 60, true, 60, true, true, 60, true, true, 60, true, db1, CleanResults{}}, args{lc, false}, true},
		{"NoDataVolumes", &DbConfig{"", "", 30015, "", "", true, 60, true, 60, true, true, 60, true, true, 60, true, db1, CleanResults{}}, args{lc, false}, false},
		{"ScanError", &DbConfig{"", "", 30015, "", "", true, 60, true, 60, true, true, 60, true, true, 60, true, db1, CleanResults{}}, args{lc, false}, true},
		{"SingleCleanFailed", &DbConfig{"", "", 30015, "", "", true, 60, true, 60, true, true, 60, true, true, 60, true, db1, CleanResults{}}, args{lc, false}, true},
		{"GoodCleanDryRun", &DbConfig{"", "", 30015, "", "", true, 60, true, 60, true, true, 60, true, true, 60, true, db1, CleanResults{}}, args{lc, true}, false},
		{"GoodCleanTwoVolumes", &DbConfig{"", "", 30015, "", "", true, 60, true, 60, true, true, 60, true, true, 60, true, db1, CleanResults{}}, args{lc, false}, false},
		{"CleanTwoVolumesOneFails", &DbConfig{"", "", 30015, "", "", true, 60, true, 60, true, true, 60, true, true, 60, true, db1, CleanResults{}}, args{lc, false}, true},
	}
	for _, tt := range tests {
		/*Set up per case mocking*/
		switch {
		case tt.name == "GoodClean":
			rows1 := sqlmock.NewRows([]string{"HOST", "PORT", "USED_SIZE", "TOTAL_SIZE"}).AddRow("testhana", "30040", "1000000", "3000000")
			rows2 := sqlmock.NewRows([]string{"TOTAL_SIZE"}).AddRow("1500000")
			mock.ExpectQuery(QUERY_GetDataVolume).WillReturnRows(rows1)
			mock.ExpectExec(GetCleanDataVolume("testhana", 30040)).WillReturnResult(sqlmock.NewResult(0, 0))
			mock.ExpectQuery(GetSpecificDataVolume("testhana", 30040)).WillReturnRows(rows2)
		case tt.name == "GoodCleanPostCheckFails":
			rows1 := sqlmock.NewRows([]string{"HOST", "PORT", "USED_SIZE", "TOTAL_SIZE"}).AddRow("testhana", "30040", "1000000", "3000000")
			mock.ExpectQuery(QUERY_GetDataVolume).WillReturnRows(rows1)
			mock.ExpectExec(GetCleanDataVolume("testhana", 30040)).WillReturnResult(sqlmock.NewResult(0, 0))
			mock.ExpectQuery(GetSpecificDataVolume("testhana", 30040)).WillReturnError(fmt.Errorf("some db error"))
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
			rows2 := sqlmock.NewRows([]string{"TOTAL_SIZE"}).AddRow("2000000")
			rows3 := sqlmock.NewRows([]string{"TOTAL_SIZE"}).AddRow("3000000")
			mock.ExpectQuery(QUERY_GetDataVolume).WillReturnRows(rows1)
			mock.ExpectExec(GetCleanDataVolume("testhana", 30040)).WillReturnResult(sqlmock.NewResult(0, 0))
			mock.ExpectQuery(GetSpecificDataVolume("testhana", 30040)).WillReturnRows(rows2)
			mock.ExpectExec(GetCleanDataVolume("testhana", 30044)).WillReturnResult(sqlmock.NewResult(0, 0))
			mock.ExpectQuery(GetSpecificDataVolume("testhana", 30044)).WillReturnRows(rows3)

		case tt.name == "CleanTwoVolumesOneFails":
			rows1 := sqlmock.NewRows([]string{"HOST", "PORT", "USED_SIZE", "TOTAL_SIZE"}).AddRow("testhana", "30040", "1000000", "3000000").AddRow("testhana", "30044", "2000000", "6000000")
			rows2 := sqlmock.NewRows([]string{"TOTAL_SIZE"}).AddRow("2000000")
			mock.ExpectQuery(QUERY_GetDataVolume).WillReturnRows(rows1)
			mock.ExpectExec(GetCleanDataVolume("testhana", 30040)).WillReturnResult(sqlmock.NewResult(0, 0))
			mock.ExpectQuery(GetSpecificDataVolume("testhana", 30040)).WillReturnRows(rows2)
			mock.ExpectExec(GetCleanDataVolume("testhana", 30044)).WillReturnError(fmt.Errorf("some db error"))
		default:
			t.Errorf("Couldn't find DB mocking for test \"%s\"\n", tt.name)
		}
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.dbc.CleanDataVolumeFunc(tt.args.lc, tt.args.dryrun); (err != nil) != tt.wantErr {
				t.Errorf("DbConfig.CleanDataVolumeFunc() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDbConfig_CheckPrivileges(t *testing.T) {
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

	defer close(lc)
	defer close(quit)

	go Logger(AppConfig{"file", true, false, false}, lc, quit)

	type args struct {
		lc chan<- LogMessage
	}
	tests := []struct {
		name    string
		dbc     *DbConfig
		args    args
		wantErr bool
	}{
		{"NothingMissing", &DbConfig{"TST", "test-hostname", 30015, "hccadmin", "", true, 60, true, 60, true, true, 60, true, true, 60, true, db1, CleanResults{}}, args{lc}, false},
		{"NoMonitoring", &DbConfig{"TST", "test-hostname", 30015, "hccadmin", "", true, 60, true, 60, true, true, 60, true, true, 60, true, db1, CleanResults{}}, args{lc}, true},
		{"NoTraceAdmin", &DbConfig{"TST", "test-hostname", 30015, "hccadmin", "", true, 60, true, 60, true, true, 60, true, true, 60, true, db1, CleanResults{}}, args{lc}, true},
		{"NoBackupAdmin", &DbConfig{"TST", "test-hostname", 30015, "hccadmin", "", true, 60, true, 60, true, true, 60, true, true, 60, true, db1, CleanResults{}}, args{lc}, true},
		{"NoLogAdmin", &DbConfig{"TST", "test-hostname", 30015, "hccadmin", "", true, 60, true, 60, true, true, 60, true, true, 60, true, db1, CleanResults{}}, args{lc}, true},
		{"NoAuditOperator", &DbConfig{"TST", "test-hostname", 30015, "hccadmin", "", true, 60, true, 60, true, true, 60, true, true, 60, true, db1, CleanResults{}}, args{lc}, true},
		{"NoResourceAdmin", &DbConfig{"TST", "test-hostname", 30015, "hccadmin", "", true, 60, true, 60, true, true, 60, true, true, 60, true, db1, CleanResults{}}, args{lc}, true},
		{"NoSelectAlerts", &DbConfig{"TST", "test-hostname", 30015, "hccadmin", "", true, 60, true, 60, true, true, 60, true, true, 60, true, db1, CleanResults{}}, args{lc}, true},
		{"NoDeleteAlerts", &DbConfig{"TST", "test-hostname", 30015, "hccadmin", "", true, 60, true, 60, true, true, 60, true, true, 60, true, db1, CleanResults{}}, args{lc}, true},
		{"NoRows", &DbConfig{"TST", "test-hostname", 30015, "hccadmin", "", true, 60, true, 60, true, true, 60, true, true, 60, true, db1, CleanResults{}}, args{lc}, true},
		{"DbError", &DbConfig{"TST", "test-hostname", 30015, "hccadmin", "", true, 60, true, 60, true, true, 60, true, true, 60, true, db1, CleanResults{}}, args{lc}, true},
		{"WrongBool", &DbConfig{"TST", "test-hostname", 30015, "hccadmin", "", true, 60, true, 60, true, true, 60, true, true, 60, true, db1, CleanResults{}}, args{lc}, true},
		{"MissingPriv", &DbConfig{"TST", "test-hostname", 30015, "hccadmin", "", true, 60, true, 60, true, true, 60, true, true, 60, true, db1, CleanResults{}}, args{lc}, true},
	}
	for _, tt := range tests {
		/*Set up per case mocking*/
		switch {
		case tt.name == "NothingMissing":
			rows1 := mock.NewRows([]string{"ROLE", "RESULT"})
			rows1.AddRow("MONITORING", "TRUE")
			rows1.AddRow("TRACE_ADMIN", "TRUE")
			rows1.AddRow("BACKUP_ADMIN", "TRUE")
			rows1.AddRow("LOG_ADMIN", "TRUE")
			rows1.AddRow("AUDIT_OPERATOR", "TRUE")
			rows1.AddRow("RESOURCE_ADMIN", "TRUE")
			rows1.AddRow("SELECT_STATISTICS_ALERTS_BASE", "TRUE")
			rows1.AddRow("DELETE_STATISTICS_ALERTS_BASE", "TRUE")
			mock.ExpectQuery(GetPrivCheck(tt.dbc.Username)).WillReturnRows(rows1)
		case tt.name == "NoMonitoring":
			rows1 := mock.NewRows([]string{"ROLE", "RESULT"})
			rows1.AddRow("MONITORING", "FALSE")
			rows1.AddRow("TRACE_ADMIN", "TRUE")
			rows1.AddRow("BACKUP_ADMIN", "TRUE")
			rows1.AddRow("LOG_ADMIN", "TRUE")
			rows1.AddRow("AUDIT_OPERATOR", "TRUE")
			rows1.AddRow("RESOURCE_ADMIN", "TRUE")
			rows1.AddRow("SELECT_STATISTICS_ALERTS_BASE", "TRUE")
			rows1.AddRow("DELETE_STATISTICS_ALERTS_BASE", "TRUE")
			mock.ExpectQuery(GetPrivCheck(tt.dbc.Username)).WillReturnRows(rows1)
		case tt.name == "NoTraceAdmin":
			rows1 := mock.NewRows([]string{"ROLE", "RESULT"})
			rows1.AddRow("MONITORING", "TRUE")
			rows1.AddRow("TRACE_ADMIN", "FALSE")
			rows1.AddRow("BACKUP_ADMIN", "TRUE")
			rows1.AddRow("LOG_ADMIN", "TRUE")
			rows1.AddRow("AUDIT_OPERATOR", "TRUE")
			rows1.AddRow("RESOURCE_ADMIN", "TRUE")
			rows1.AddRow("SELECT_STATISTICS_ALERTS_BASE", "TRUE")
			rows1.AddRow("DELETE_STATISTICS_ALERTS_BASE", "TRUE")
			mock.ExpectQuery(GetPrivCheck(tt.dbc.Username)).WillReturnRows(rows1)
		case tt.name == "NoBackupAdmin":
			rows1 := mock.NewRows([]string{"ROLE", "RESULT"})
			rows1.AddRow("MONITORING", "TRUE")
			rows1.AddRow("TRACE_ADMIN", "TRUE")
			rows1.AddRow("BACKUP_ADMIN", "FALSE")
			rows1.AddRow("LOG_ADMIN", "TRUE")
			rows1.AddRow("AUDIT_OPERATOR", "TRUE")
			rows1.AddRow("RESOURCE_ADMIN", "TRUE")
			rows1.AddRow("SELECT_STATISTICS_ALERTS_BASE", "TRUE")
			rows1.AddRow("DELETE_STATISTICS_ALERTS_BASE", "TRUE")
			mock.ExpectQuery(GetPrivCheck(tt.dbc.Username)).WillReturnRows(rows1)
		case tt.name == "NoLogAdmin":
			rows1 := mock.NewRows([]string{"ROLE", "RESULT"})
			rows1.AddRow("MONITORING", "TRUE")
			rows1.AddRow("TRACE_ADMIN", "TRUE")
			rows1.AddRow("BACKUP_ADMIN", "TRUE")
			rows1.AddRow("LOG_ADMIN", "FALSE")
			rows1.AddRow("AUDIT_OPERATOR", "TRUE")
			rows1.AddRow("RESOURCE_ADMIN", "TRUE")
			rows1.AddRow("SELECT_STATISTICS_ALERTS_BASE", "TRUE")
			rows1.AddRow("DELETE_STATISTICS_ALERTS_BASE", "TRUE")
			mock.ExpectQuery(GetPrivCheck(tt.dbc.Username)).WillReturnRows(rows1)
		case tt.name == "NoAuditOperator":
			rows1 := mock.NewRows([]string{"ROLE", "RESULT"})
			rows1.AddRow("MONITORING", "TRUE")
			rows1.AddRow("TRACE_ADMIN", "TRUE")
			rows1.AddRow("BACKUP_ADMIN", "TRUE")
			rows1.AddRow("LOG_ADMIN", "TRUE")
			rows1.AddRow("AUDIT_OPERATOR", "FALSE")
			rows1.AddRow("RESOURCE_ADMIN", "TRUE")
			rows1.AddRow("SELECT_STATISTICS_ALERTS_BASE", "TRUE")
			rows1.AddRow("DELETE_STATISTICS_ALERTS_BASE", "TRUE")
			mock.ExpectQuery(GetPrivCheck(tt.dbc.Username)).WillReturnRows(rows1)
		case tt.name == "NoResourceAdmin":
			rows1 := mock.NewRows([]string{"ROLE", "RESULT"})
			rows1.AddRow("MONITORING", "TRUE")
			rows1.AddRow("TRACE_ADMIN", "TRUE")
			rows1.AddRow("BACKUP_ADMIN", "TRUE")
			rows1.AddRow("LOG_ADMIN", "TRUE")
			rows1.AddRow("AUDIT_OPERATOR", "TRUE")
			rows1.AddRow("RESOURCE_ADMIN", "FALSE")
			rows1.AddRow("SELECT_STATISTICS_ALERTS_BASE", "TRUE")
			rows1.AddRow("DELETE_STATISTICS_ALERTS_BASE", "TRUE")
			mock.ExpectQuery(GetPrivCheck(tt.dbc.Username)).WillReturnRows(rows1)
		case tt.name == "NoSelectAlerts":
			rows1 := mock.NewRows([]string{"ROLE", "RESULT"})
			rows1.AddRow("MONITORING", "TRUE")
			rows1.AddRow("TRACE_ADMIN", "TRUE")
			rows1.AddRow("BACKUP_ADMIN", "TRUE")
			rows1.AddRow("LOG_ADMIN", "TRUE")
			rows1.AddRow("AUDIT_OPERATOR", "TRUE")
			rows1.AddRow("RESOURCE_ADMIN", "TRUE")
			rows1.AddRow("SELECT_STATISTICS_ALERTS_BASE", "FALSE")
			rows1.AddRow("DELETE_STATISTICS_ALERTS_BASE", "TRUE")
			mock.ExpectQuery(GetPrivCheck(tt.dbc.Username)).WillReturnRows(rows1)
		case tt.name == "NoDeleteAlerts":
			rows1 := mock.NewRows([]string{"ROLE", "RESULT"})
			rows1.AddRow("MONITORING", "TRUE")
			rows1.AddRow("TRACE_ADMIN", "TRUE")
			rows1.AddRow("BACKUP_ADMIN", "TRUE")
			rows1.AddRow("LOG_ADMIN", "TRUE")
			rows1.AddRow("AUDIT_OPERATOR", "TRUE")
			rows1.AddRow("RESOURCE_ADMIN", "TRUE")
			rows1.AddRow("SELECT_STATISTICS_ALERTS_BASE", "TRUE")
			rows1.AddRow("DELETE_STATISTICS_ALERTS_BASE", "FALSE")
			mock.ExpectQuery(GetPrivCheck(tt.dbc.Username)).WillReturnRows(rows1)
		case tt.name == "NoRows":
			mock.ExpectQuery(GetPrivCheck(tt.dbc.Username)).WillReturnError(sql.ErrNoRows)
		case tt.name == "DbError":
			mock.ExpectQuery(GetPrivCheck(tt.dbc.Username)).WillReturnError(fmt.Errorf("some db error"))
		case tt.name == "WrongBool":
			rows1 := mock.NewRows([]string{"ROLE", "RESULT"})
			rows1.AddRow("MONITORING", "TRUEZ")
			rows1.AddRow("TRACE_ADMIN", "TRUE")
			rows1.AddRow("BACKUP_ADMIN", "TRUE")
			rows1.AddRow("LOG_ADMIN", "TRUE")
			rows1.AddRow("AUDIT_OPERATOR", "TRUE")
			rows1.AddRow("RESOURCE_ADMIN", "TRUE")
			rows1.AddRow("SELECT_STATISTICS_ALERTS_BASE", "TRUE")
			rows1.AddRow("DELETE_STATISTICS_ALERTS_BASE", "FALSE")
			mock.ExpectQuery(GetPrivCheck(tt.dbc.Username)).WillReturnRows(rows1)
		case tt.name == "MissingPriv":
			rows1 := mock.NewRows([]string{"ROLE", "RESULT"})
			rows1.AddRow("MONITORING", "TRUE")
			rows1.AddRow("BACKUP_ADMIN", "TRUE")
			rows1.AddRow("LOG_ADMIN", "TRUE")
			rows1.AddRow("AUDIT_OPERATOR", "TRUE")
			rows1.AddRow("RESOURCE_ADMIN", "TRUE")
			rows1.AddRow("SELECT_STATISTICS_ALERTS_BASE", "TRUE")
			rows1.AddRow("DELETE_STATISTICS_ALERTS_BASE", "FALSE")
			mock.ExpectQuery(GetPrivCheck(tt.dbc.Username)).WillReturnRows(rows1)
		default:
			t.Errorf("Couldn't find DB mocking for test \"%s\"\n", tt.name)
		}
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.dbc.CheckPrivileges(tt.args.lc); (err != nil) != tt.wantErr {
				t.Errorf("DbConfig.CheckPrivileges() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDbConfig_CheckDataClean(t *testing.T) {
	/*Test Setup*/
	/*Mock DB*/
	db1, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Errorf("an error '%s' was not expected when opening mock database connection", err)
	}
	defer db1.Close()
	type args struct {
		host string
		port uint
	}
	tests := []struct {
		name    string
		dbc     *DbConfig
		args    args
		want    uint64
		wantErr bool
	}{
		{"Good01", &DbConfig{"TST", "test-hostname", 30015, "hccadmin", "", true, 60, true, 60, true, true, 60, true, true, 60, true, db1, CleanResults{}}, args{"localhost", 30040}, 30000, false},
		{"DbError", &DbConfig{"TST", "test-hostname", 30015, "hccadmin", "", true, 60, true, 60, true, true, 60, true, true, 60, true, db1, CleanResults{}}, args{"localhost", 30040}, 0, true},
		{"NoRows", &DbConfig{"TST", "test-hostname", 30015, "hccadmin", "", true, 60, true, 60, true, true, 60, true, true, 60, true, db1, CleanResults{}}, args{"localhost", 30040}, 0, true},
	}
	for _, tt := range tests {
		/*per case mocking*/
		switch {
		case tt.name == "Good01":
			row := mock.NewRows([]string{"TOTAL_BYTES"}).AddRow("30000")
			mock.ExpectQuery(GetSpecificDataVolume(tt.args.host, tt.args.port)).WillReturnRows(row)
		case tt.name == "DbError":
			mock.ExpectQuery(GetSpecificDataVolume(tt.args.host, tt.args.port)).WillReturnError(fmt.Errorf("some db error"))
		case tt.name == "NoRows":
			mock.ExpectQuery(GetSpecificDataVolume(tt.args.host, tt.args.port)).WillReturnError(sql.ErrNoRows)
		}
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.dbc.CheckDataClean(tt.args.host, tt.args.port)
			if (err != nil) != tt.wantErr {
				t.Errorf("DbConfig.CheckDataClean() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("DbConfig.CheckDataClean() = %v, want %v", got, tt.want)
			}
		})
	}
}
