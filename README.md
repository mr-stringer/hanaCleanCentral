# hanaCleanCentral

A centralised maintenance tool for routine the HANA database.

## Do not use, yet

This project is in early development.  There are no releases at this time.  Once released, binaries will be available, but of course, you are welcome to compile this yourself.

## Introduction

hanaCleanCentral is based on [hanacleaner](https://github.com/chriselswede/hanacleaner).  This project aims to perform a similar range of tasks that hanacleaner performs but centralised.  Rather than being installed on each HANA database, hanaCleanCentral will be installed on a central server where it will remotely housekeep many HANA instances.  hanaCleanCentral can be configured to perform maintenance on remote databases in a series or parallel.

## Required Privileges

The following list documents the required privileges for hanaCleanCentral.  This list is likely to grow as more features are developed.  It is strongly recommended to use a dedicated user for the hanaCleanCentral application.

|Application area |Type | Value |
|---|---|---|
|General|Role|`MANAGEMENT`|
|TraceFile management |Privilege|`TRACE ADMIN`|
|Backup catalog management|Privilege|`BACKUP OPERATOR`|

## Flags and Configuration

hanaCleanCentral is controlled with a combination of command-line flags and a configuration file.  Supported flags are:

* -f the location of the configuration file.  Required, defaults to config.json
* -v verbose.  When used, verbose logging is enabled, defaults to off
* -d dry run.  When used, only read-only queries will be executed.  This mode will make no changes to the target databases.

The -f flag specifies the configuration.  The configuration file represents an array of DbConfig structs described below.

```go-lang
type DbConfig struct {
    Name                       string // Friendly name of the DB.  <Tenant>@<SID> is a good option here
    Hostname                   string // Hostname or IP address of the primary HANA node
    Port                       uint   // Port of the HANA DB
    Username                   string // HANA DB user name to use
    Password                   string // Password for HANA DB user
    RemoveTraces               bool   // If true, trace file management will be enabled - Defaults to false
    TraceRetentionDays         uint   // Specifies the number of days of trace files to retain
    TruncateBackupCatalog      bool   // If true, backup catalog truncation will be enabled - Defaults to false
    BackupCatalogRetentionDays uint   // Specifies the number of days of entries to retain
    DeleteOldBackups           bool   // If true, truncated files will be physically removed, if false entries are removed from the database only - Defaults to false
    ClearAlerts                bool   // If true, old alerts are removed from the embedded statistics server - Defaults to false
    AlertsOlderDeleteDays      uint   // Specifies the number of days of alerts to retain
}```

An example of a configuration file  a single databases is provided below.

```JSON
{
    "Databases":[
        {
            "Name": "DB1@HAN",
            "Hostname": "hanasever.int.bybiz.net",
            "Port": 30041,
            "Username": "HccUser",
            "Password": "234F£$5gf£t345H$%",
            "RemoveTraces": true,
            "TraceRetentionDays": 60,
            "TruncateBackupCatalog": false,
            "BackupCatalogRetentionDays": 60,
            "DeleteOldBackups": true,
            "ClearAlerts" :true,
            "AlertsOlderDeleteDays": 60
        }
    ]
}
```
