package main

import (
	"testing"
)

func TestGetTraceFileQuery(t *testing.T) {
	type args struct {
		days uint
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"Good02", args{2}, "SELECT HOST, FILE_NAME, FILE_SIZE, FILE_MTIME FROM \"SYS\".\"M_TRACEFILES\" WHERE FILE_MTIME < (SELECT ADD_DAYS(NOW(), -2) FROM DUMMY) AND RIGHT(FILE_NAME, 3) = 'trc' OR FILE_MTIME < (SELECT ADD_DAYS(NOW(), -2) FROM DUMMY) AND RIGHT(FILE_NAME, 2) = 'gz'"},
		{"Good01", args{1}, "SELECT HOST, FILE_NAME, FILE_SIZE, FILE_MTIME FROM \"SYS\".\"M_TRACEFILES\" WHERE FILE_MTIME < (SELECT ADD_DAYS(NOW(), -1) FROM DUMMY) AND RIGHT(FILE_NAME, 3) = 'trc' OR FILE_MTIME < (SELECT ADD_DAYS(NOW(), -1) FROM DUMMY) AND RIGHT(FILE_NAME, 2) = 'gz'"},
		{"Good03", args{3}, "SELECT HOST, FILE_NAME, FILE_SIZE, FILE_MTIME FROM \"SYS\".\"M_TRACEFILES\" WHERE FILE_MTIME < (SELECT ADD_DAYS(NOW(), -3) FROM DUMMY) AND RIGHT(FILE_NAME, 3) = 'trc' OR FILE_MTIME < (SELECT ADD_DAYS(NOW(), -3) FROM DUMMY) AND RIGHT(FILE_NAME, 2) = 'gz'"},
		{"Good04", args{7}, "SELECT HOST, FILE_NAME, FILE_SIZE, FILE_MTIME FROM \"SYS\".\"M_TRACEFILES\" WHERE FILE_MTIME < (SELECT ADD_DAYS(NOW(), -7) FROM DUMMY) AND RIGHT(FILE_NAME, 3) = 'trc' OR FILE_MTIME < (SELECT ADD_DAYS(NOW(), -7) FROM DUMMY) AND RIGHT(FILE_NAME, 2) = 'gz'"},
		{"Good06", args{21}, "SELECT HOST, FILE_NAME, FILE_SIZE, FILE_MTIME FROM \"SYS\".\"M_TRACEFILES\" WHERE FILE_MTIME < (SELECT ADD_DAYS(NOW(), -21) FROM DUMMY) AND RIGHT(FILE_NAME, 3) = 'trc' OR FILE_MTIME < (SELECT ADD_DAYS(NOW(), -21) FROM DUMMY) AND RIGHT(FILE_NAME, 2) = 'gz'"},
		{"Good07", args{28}, "SELECT HOST, FILE_NAME, FILE_SIZE, FILE_MTIME FROM \"SYS\".\"M_TRACEFILES\" WHERE FILE_MTIME < (SELECT ADD_DAYS(NOW(), -28) FROM DUMMY) AND RIGHT(FILE_NAME, 3) = 'trc' OR FILE_MTIME < (SELECT ADD_DAYS(NOW(), -28) FROM DUMMY) AND RIGHT(FILE_NAME, 2) = 'gz'"},
		{"Good08", args{30}, "SELECT HOST, FILE_NAME, FILE_SIZE, FILE_MTIME FROM \"SYS\".\"M_TRACEFILES\" WHERE FILE_MTIME < (SELECT ADD_DAYS(NOW(), -30) FROM DUMMY) AND RIGHT(FILE_NAME, 3) = 'trc' OR FILE_MTIME < (SELECT ADD_DAYS(NOW(), -30) FROM DUMMY) AND RIGHT(FILE_NAME, 2) = 'gz'"},
		{"Good05", args{14}, "SELECT HOST, FILE_NAME, FILE_SIZE, FILE_MTIME FROM \"SYS\".\"M_TRACEFILES\" WHERE FILE_MTIME < (SELECT ADD_DAYS(NOW(), -14) FROM DUMMY) AND RIGHT(FILE_NAME, 3) = 'trc' OR FILE_MTIME < (SELECT ADD_DAYS(NOW(), -14) FROM DUMMY) AND RIGHT(FILE_NAME, 2) = 'gz'"},
		{"Good09", args{999}, "SELECT HOST, FILE_NAME, FILE_SIZE, FILE_MTIME FROM \"SYS\".\"M_TRACEFILES\" WHERE FILE_MTIME < (SELECT ADD_DAYS(NOW(), -999) FROM DUMMY) AND RIGHT(FILE_NAME, 3) = 'trc' OR FILE_MTIME < (SELECT ADD_DAYS(NOW(), -999) FROM DUMMY) AND RIGHT(FILE_NAME, 2) = 'gz'"},
		{"Good10", args{365}, "SELECT HOST, FILE_NAME, FILE_SIZE, FILE_MTIME FROM \"SYS\".\"M_TRACEFILES\" WHERE FILE_MTIME < (SELECT ADD_DAYS(NOW(), -365) FROM DUMMY) AND RIGHT(FILE_NAME, 3) = 'trc' OR FILE_MTIME < (SELECT ADD_DAYS(NOW(), -365) FROM DUMMY) AND RIGHT(FILE_NAME, 2) = 'gz'"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetTraceFileQuery(tt.args.days); got != tt.want {
				t.Errorf("GetTraceFileQuery() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetRemoveTrace(t *testing.T) {
	type args struct {
		hostname string
		filename string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"Good01", args{"hanaserver", "trace.trc"}, "ALTER SYSTEM REMOVE TRACES('hanaserver', 'trace.trc'"},
		{"Good02", args{"localhost", "nameserver_host.00000.000.trc"}, "ALTER SYSTEM REMOVE TRACES('localhost', 'nameserver_host.00000.000.trc'"},
		{"Good03", args{"hasd1453", "trace.trc"}, "ALTER SYSTEM REMOVE TRACES('hasd1453', 'trace.trc'"},
		{"Good04", args{"long.server.name.example.int", "trace.trc"}, "ALTER SYSTEM REMOVE TRACES('long.server.name.example.int', 'trace.trc'"},
		{"Good05", args{"long.server.name.example.int", "compileserver_alert_host_20210312144914.gz"}, "ALTER SYSTEM REMOVE TRACES('long.server.name.example.int', 'compileserver_alert_host_20210312144914.gz'"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetRemoveTrace(tt.args.hostname, tt.args.filename); got != tt.want {
				t.Errorf("GetRemoveTrace() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetAlertCount(t *testing.T) {
	type args struct {
		days uint
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"Good01", args{1}, "SELECT COUNT(SNAPSHOT_ID) AS COUNT FROM \"_SYS_STATISTICS\".\"STATISTICS_ALERTS_BASE\" WHERE ALERT_TIMESTAMP < ADD_DAYS(NOW(), -1) LIMIT 1"},
		{"Good02", args{2}, "SELECT COUNT(SNAPSHOT_ID) AS COUNT FROM \"_SYS_STATISTICS\".\"STATISTICS_ALERTS_BASE\" WHERE ALERT_TIMESTAMP < ADD_DAYS(NOW(), -2) LIMIT 1"},
		{"Good03", args{4}, "SELECT COUNT(SNAPSHOT_ID) AS COUNT FROM \"_SYS_STATISTICS\".\"STATISTICS_ALERTS_BASE\" WHERE ALERT_TIMESTAMP < ADD_DAYS(NOW(), -4) LIMIT 1"},
		{"Good04", args{10}, "SELECT COUNT(SNAPSHOT_ID) AS COUNT FROM \"_SYS_STATISTICS\".\"STATISTICS_ALERTS_BASE\" WHERE ALERT_TIMESTAMP < ADD_DAYS(NOW(), -10) LIMIT 1"},
		{"Good05", args{28}, "SELECT COUNT(SNAPSHOT_ID) AS COUNT FROM \"_SYS_STATISTICS\".\"STATISTICS_ALERTS_BASE\" WHERE ALERT_TIMESTAMP < ADD_DAYS(NOW(), -28) LIMIT 1"},
		{"Good06", args{60}, "SELECT COUNT(SNAPSHOT_ID) AS COUNT FROM \"_SYS_STATISTICS\".\"STATISTICS_ALERTS_BASE\" WHERE ALERT_TIMESTAMP < ADD_DAYS(NOW(), -60) LIMIT 1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetAlertCount(tt.args.days); got != tt.want {
				t.Errorf("GetAlertCount() = %v, want %v", got, tt.want)
			}
		})
	}
}
