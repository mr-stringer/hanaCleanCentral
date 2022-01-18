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
		{"Good01", args{"hanaserver", "trace.trc"}, "ALTER SYSTEM REMOVE TRACES('hanaserver', 'trace.trc')"},
		{"Good02", args{"localhost", "nameserver_host.00000.000.trc"}, "ALTER SYSTEM REMOVE TRACES('localhost', 'nameserver_host.00000.000.trc')"},
		{"Good03", args{"hasd1453", "trace.trc"}, "ALTER SYSTEM REMOVE TRACES('hasd1453', 'trace.trc')"},
		{"Good04", args{"long.server.name.example.int", "trace.trc"}, "ALTER SYSTEM REMOVE TRACES('long.server.name.example.int', 'trace.trc')"},
		{"Good05", args{"long.server.name.example.int", "compileserver_alert_host_20210312144914.gz"}, "ALTER SYSTEM REMOVE TRACES('long.server.name.example.int', 'compileserver_alert_host_20210312144914.gz')"},
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

func TestGetCheckTracePresent(t *testing.T) {
	type args struct {
		filename string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"Good01", args{"TRACEFILE01"}, "SELECT COUNT(FILE_NAME) AS TRACE FROM \"SYS\".\"M_TRACEFILES\" WHERE FILE_NAME = 'TRACEFILE01'"},
		{"Good02", args{"nameserver_host.00000.000.trc"}, "SELECT COUNT(FILE_NAME) AS TRACE FROM \"SYS\".\"M_TRACEFILES\" WHERE FILE_NAME = 'nameserver_host.00000.000.trc'"},
		{"Good02", args{"compileserver_alert_host_20210312144914.gz"}, "SELECT COUNT(FILE_NAME) AS TRACE FROM \"SYS\".\"M_TRACEFILES\" WHERE FILE_NAME = 'compileserver_alert_host_20210312144914.gz'"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetCheckTracePresent(tt.args.filename); got != tt.want {
				t.Errorf("GetCheckTracePresent() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetLatestFullBackupID(t *testing.T) {
	type args struct {
		days uint
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"tc1", args{7}, "SELECT BACKUP_ID FROM \"SYS\".\"M_BACKUP_CATALOG\" WHERE STATE_NAME = 'successful' AND ENTRY_TYPE_NAME = 'complete data backup' AND SYS_END_TIME < (SELECT ADD_DAYS(NOW(),-7) FROM DUMMY) ORDER BY SYS_END_TIME DESC LIMIT 1"},
		{"tc2", args{14}, "SELECT BACKUP_ID FROM \"SYS\".\"M_BACKUP_CATALOG\" WHERE STATE_NAME = 'successful' AND ENTRY_TYPE_NAME = 'complete data backup' AND SYS_END_TIME < (SELECT ADD_DAYS(NOW(),-14) FROM DUMMY) ORDER BY SYS_END_TIME DESC LIMIT 1"},
		{"tc3", args{21}, "SELECT BACKUP_ID FROM \"SYS\".\"M_BACKUP_CATALOG\" WHERE STATE_NAME = 'successful' AND ENTRY_TYPE_NAME = 'complete data backup' AND SYS_END_TIME < (SELECT ADD_DAYS(NOW(),-21) FROM DUMMY) ORDER BY SYS_END_TIME DESC LIMIT 1"},
		{"tc4", args{28}, "SELECT BACKUP_ID FROM \"SYS\".\"M_BACKUP_CATALOG\" WHERE STATE_NAME = 'successful' AND ENTRY_TYPE_NAME = 'complete data backup' AND SYS_END_TIME < (SELECT ADD_DAYS(NOW(),-28) FROM DUMMY) ORDER BY SYS_END_TIME DESC LIMIT 1"},
		{"tc5", args{60}, "SELECT BACKUP_ID FROM \"SYS\".\"M_BACKUP_CATALOG\" WHERE STATE_NAME = 'successful' AND ENTRY_TYPE_NAME = 'complete data backup' AND SYS_END_TIME < (SELECT ADD_DAYS(NOW(),-60) FROM DUMMY) ORDER BY SYS_END_TIME DESC LIMIT 1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetLatestFullBackupID(tt.args.days); got != tt.want {
				t.Errorf("GetLatestFullBackupID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetBackupFileData(t *testing.T) {
	type args struct {
		backupid string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"tc1", args{"1234567890"}, "SELECT B.ENTRY_TYPE_NAME AS ENTRY, COUNT(B.BACKUP_ID) AS COUNT, SUM(F.BACKUP_SIZE) AS BYTES FROM \"SYS\".\"M_BACKUP_CATALOG\" AS B LEFT JOIN \"SYS\".\"M_BACKUP_CATALOG_FILES\" AS F ON B.BACKUP_ID = F.BACKUP_ID WHERE B.BACKUP_ID < 1234567890 GROUP BY B.ENTRY_TYPE_NAME"},
		{"tc2", args{"1234567890"}, "SELECT B.ENTRY_TYPE_NAME AS ENTRY, COUNT(B.BACKUP_ID) AS COUNT, SUM(F.BACKUP_SIZE) AS BYTES FROM \"SYS\".\"M_BACKUP_CATALOG\" AS B LEFT JOIN \"SYS\".\"M_BACKUP_CATALOG_FILES\" AS F ON B.BACKUP_ID = F.BACKUP_ID WHERE B.BACKUP_ID < 1234567890 GROUP BY B.ENTRY_TYPE_NAME"},
		{"tc3", args{"5555555555"}, "SELECT B.ENTRY_TYPE_NAME AS ENTRY, COUNT(B.BACKUP_ID) AS COUNT, SUM(F.BACKUP_SIZE) AS BYTES FROM \"SYS\".\"M_BACKUP_CATALOG\" AS B LEFT JOIN \"SYS\".\"M_BACKUP_CATALOG_FILES\" AS F ON B.BACKUP_ID = F.BACKUP_ID WHERE B.BACKUP_ID < 5555555555 GROUP BY B.ENTRY_TYPE_NAME"},
		{"tc4", args{"1346798520"}, "SELECT B.ENTRY_TYPE_NAME AS ENTRY, COUNT(B.BACKUP_ID) AS COUNT, SUM(F.BACKUP_SIZE) AS BYTES FROM \"SYS\".\"M_BACKUP_CATALOG\" AS B LEFT JOIN \"SYS\".\"M_BACKUP_CATALOG_FILES\" AS F ON B.BACKUP_ID = F.BACKUP_ID WHERE B.BACKUP_ID < 1346798520 GROUP BY B.ENTRY_TYPE_NAME"},
		{"tc5", args{"9632587410"}, "SELECT B.ENTRY_TYPE_NAME AS ENTRY, COUNT(B.BACKUP_ID) AS COUNT, SUM(F.BACKUP_SIZE) AS BYTES FROM \"SYS\".\"M_BACKUP_CATALOG\" AS B LEFT JOIN \"SYS\".\"M_BACKUP_CATALOG_FILES\" AS F ON B.BACKUP_ID = F.BACKUP_ID WHERE B.BACKUP_ID < 9632587410 GROUP BY B.ENTRY_TYPE_NAME"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetBackupFileData(tt.args.backupid); got != tt.want {
				t.Errorf("GetBackupFileData() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetBackupDelete(t *testing.T) {
	type args struct {
		backupid string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"tc1", args{"1234567890"}, "BACKUP CATALOG DELETE ALL BEFORE BACKUP_ID 1234567890"},
		{"tc1", args{"1234567890"}, "BACKUP CATALOG DELETE ALL BEFORE BACKUP_ID 1234567890"},
		{"tc2", args{"1234567890"}, "BACKUP CATALOG DELETE ALL BEFORE BACKUP_ID 1234567890"},
		{"tc3", args{"5555555555"}, "BACKUP CATALOG DELETE ALL BEFORE BACKUP_ID 5555555555"},
		{"tc4", args{"1346798520"}, "BACKUP CATALOG DELETE ALL BEFORE BACKUP_ID 1346798520"},
		{"tc5", args{"9632587410"}, "BACKUP CATALOG DELETE ALL BEFORE BACKUP_ID 9632587410"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetBackupDelete(tt.args.backupid); got != tt.want {
				t.Errorf("GetBackupDelete() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetBackupDeleteComplete(t *testing.T) {
	type args struct {
		backupid string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"tc1", args{"1234567890"}, "BACKUP CATALOG DELETE ALL BEFORE BACKUP_ID 1234567890 COMPLETE"},
		{"tc1", args{"1234567890"}, "BACKUP CATALOG DELETE ALL BEFORE BACKUP_ID 1234567890 COMPLETE"},
		{"tc2", args{"1234567890"}, "BACKUP CATALOG DELETE ALL BEFORE BACKUP_ID 1234567890 COMPLETE"},
		{"tc3", args{"5555555555"}, "BACKUP CATALOG DELETE ALL BEFORE BACKUP_ID 5555555555 COMPLETE"},
		{"tc4", args{"1346798520"}, "BACKUP CATALOG DELETE ALL BEFORE BACKUP_ID 1346798520 COMPLETE"},
		{"tc5", args{"9632587410"}, "BACKUP CATALOG DELETE ALL BEFORE BACKUP_ID 9632587410 COMPLETE"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetBackupDeleteComplete(tt.args.backupid); got != tt.want {
				t.Errorf("GetBackupDeleteComplete() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetAlertDelete(t *testing.T) {
	type args struct {
		days uint
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"tc1", args{7}, "DELETE FROM \"_SYS_STATISTICS\".\"STATISTICS_ALERTS_BASE\" WHERE ALERT_TIMESTAMP < ADD_DAYS(NOW(), -7)"},
		{"tc2", args{14}, "DELETE FROM \"_SYS_STATISTICS\".\"STATISTICS_ALERTS_BASE\" WHERE ALERT_TIMESTAMP < ADD_DAYS(NOW(), -14)"},
		{"tc3", args{21}, "DELETE FROM \"_SYS_STATISTICS\".\"STATISTICS_ALERTS_BASE\" WHERE ALERT_TIMESTAMP < ADD_DAYS(NOW(), -21)"},
		{"tc4", args{30}, "DELETE FROM \"_SYS_STATISTICS\".\"STATISTICS_ALERTS_BASE\" WHERE ALERT_TIMESTAMP < ADD_DAYS(NOW(), -30)"},
		{"tc5", args{90}, "DELETE FROM \"_SYS_STATISTICS\".\"STATISTICS_ALERTS_BASE\" WHERE ALERT_TIMESTAMP < ADD_DAYS(NOW(), -90)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetAlertDelete(tt.args.days); got != tt.want {
				t.Errorf("GetAlertDelete() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetAuditCount(t *testing.T) {
	type args struct {
		days uint
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"tc1", args{7}, "SELECT COUNT(TIMESTAMP) AS COUNT FROM \"SYS\".\"AUDIT_LOG\" WHERE TIMESTAMP < (SELECT ADD_DAYS(NOW(), -7) FROM DUMMY)"},
		{"tc2", args{14}, "SELECT COUNT(TIMESTAMP) AS COUNT FROM \"SYS\".\"AUDIT_LOG\" WHERE TIMESTAMP < (SELECT ADD_DAYS(NOW(), -14) FROM DUMMY)"},
		{"tc3", args{21}, "SELECT COUNT(TIMESTAMP) AS COUNT FROM \"SYS\".\"AUDIT_LOG\" WHERE TIMESTAMP < (SELECT ADD_DAYS(NOW(), -21) FROM DUMMY)"},
		{"tc4", args{30}, "SELECT COUNT(TIMESTAMP) AS COUNT FROM \"SYS\".\"AUDIT_LOG\" WHERE TIMESTAMP < (SELECT ADD_DAYS(NOW(), -30) FROM DUMMY)"},
		{"tc5", args{90}, "SELECT COUNT(TIMESTAMP) AS COUNT FROM \"SYS\".\"AUDIT_LOG\" WHERE TIMESTAMP < (SELECT ADD_DAYS(NOW(), -90) FROM DUMMY)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetAuditCount(tt.args.days); got != tt.want {
				t.Errorf("GetAuditCount() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetDatetime(t *testing.T) {
	type args struct {
		days uint
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"tc1", args{7}, "SELECT ADD_DAYS(NOW(), -7) AS NOW FROM DUMMY"},
		{"tc2", args{14}, "SELECT ADD_DAYS(NOW(), -14) AS NOW FROM DUMMY"},
		{"tc3", args{21}, "SELECT ADD_DAYS(NOW(), -21) AS NOW FROM DUMMY"},
		{"tc4", args{30}, "SELECT ADD_DAYS(NOW(), -30) AS NOW FROM DUMMY"},
		{"tc5", args{90}, "SELECT ADD_DAYS(NOW(), -90) AS NOW FROM DUMMY"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetDatetime(tt.args.days); got != tt.want {
				t.Errorf("GetDatetime() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetTruncateAuditLog(t *testing.T) {
	type args struct {
		datetime string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"tc1", args{"2020-01-01 00:00:00"}, "ALTER SYSTEM CLEAR AUDIT LOG UNTIL '2020-01-01 00:00:00'"},
		{"tc2", args{"2020-06-06 12:00:00"}, "ALTER SYSTEM CLEAR AUDIT LOG UNTIL '2020-06-06 12:00:00'"},
		{"tc3", args{"2021-12-18 23:45:12"}, "ALTER SYSTEM CLEAR AUDIT LOG UNTIL '2021-12-18 23:45:12'"},
		{"tc4", args{"2022-01-18 14:56:22"}, "ALTER SYSTEM CLEAR AUDIT LOG UNTIL '2022-01-18 14:56:22'"},
		{"tc5", args{"2019-06-26 13:13:13"}, "ALTER SYSTEM CLEAR AUDIT LOG UNTIL '2019-06-26 13:13:13'"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetTruncateAuditLog(tt.args.datetime); got != tt.want {
				t.Errorf("GetTruncateAuditLog() = %v, want %v", got, tt.want)
			}
		})
	}
}
