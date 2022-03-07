# hanaCleanCentral

[![Go](https://github.com/mr-stringer/hanaCleanCentral/actions/workflows/go.yml/badge.svg)](https://github.com/mr-stringer/hanaCleanCentral/actions/workflows/go.yml)

hanaCleanCentral is a centralised maintenance tool for the HANA database.  Hereafter referred to as HCC.

## Do not use, yet

This project is in development.  There are no releases at this time.  Once released, binaries will be available, but of course, you are welcome to compile this yourself.

## Introduction

HCC is based on [hanacleaner](https://github.com/chriselswede/hanacleaner).  This project aims to perform a similar range of tasks that hanacleaner performs but centralised.  Rather than being installed on each HANA database, HCC will be installed on a central server and remotely clean many HANA instances.  HCC can be configured to perform maintenance on remote databases in a series or parallel.

## What does HanaCleanCentral do?

HCC is currently capable of performing the following tasks:

* Trace file management - removing trace files that are no longer open are or older than the specified number of days.
* Backup catalog management - removing entries from the backup catalog that are older than the specified number of days.  HCC has the option to physically delete the files referred to in the catalog too.
* Alerts management - removing alerts from the alerts table older than the specified number of days.
* Log volume management - removing freed segments from the log volume
* Audit table management - removing audit entries older than the specified number of days.

## hanacleaner vs hanaCleanCentral

The following table describes the differences between hanacleaner and HCC.  You can use this table to help you decide which project is right for you.

| Area | hanacleaner | hanaCleanCentral | Description |
|---|---|---|---|
| Deployment | Local | Central | hanacleaner is deployed locally on each database server that is to be managed.  HCC can manage many databases from a single installation |
| Execution | Uses SQL & OS-level commands | SQL only | hanacleaner can perform more types of tasks as it executes SQL statements and OS level commands.  HCC performs SQL statements only. |
| Runtime dependencies | Relies on python runtime shipped with HANA. | Statically compiled, no specific dependencies | HCC can be compiled for and run on any major OS/architecture combination.
| Code testing | No unit testing | Significant unit testing coverage | Unit tests helps ensure that code executes as expected and allows developers to simulate situations that may be rare in the field |

## Required Privileges

The following list documents the required privileges for HCC.  This list is likely to grow as more features are developed.  It is strongly recommended to use a dedicated user for the HCC application.

|Application area |Type | Value |
|---|---|---|
|General|Role|`MANAGEMENT`|
|TraceFile management |Privilege|`TRACE ADMIN`|
|Backup catalog management|Privilege|`BACKUP OPERATOR`|
|Log management|Privilege|`LOG ADMIN`|

## Flags and Configuration

HCC is controlled with a combination of command-line flags and a configuration file.  Supported flags are:

* -f the location of the configuration file.  Required, defaults to config.json
* -v verbose.  When used, verbose logging is enabled, defaults to off
* -d dry run.  When used, only read-only queries will be executed.  This mode will make no changes to the target databases.
* -p print effective config, When used, the application configuration is printed to screen and the application quits.  Useful for understand the impact of the config inheritance.  Please note, passwords will not be printed for security purposes.

The -f flag specifies the configuration.  HCC expects the configuration file passed to it to be a JSON representation of the following struct:

```go
type Config struct {
  CleanTrace              bool // If true, trace file management will be enabled
  RetainTraceDays         uint // Specifies the number of days of trace files to retain
  CleanBackupCatalog      bool // If true, backup catalog truncation will be enabled
  RetainBackupCatalogDays uint // Specifies the number of days of entries to retain
  DeleteOldBackups        bool // If true, truncated files will be physically removed, if false entries are removed from the database only
  CleanAlerts             bool // If true, old alerts are removed from the embedded statistics server
  RetainAlertsDays        uint // Specifies the number of days of alerts to retain
  CleanLogVolume          bool // If true, free log segments will be removed from the file system
  CleanAudit              bool // If true, old audit records will be deleted
  RetainAuditDays         uint // Specifies the number of days of audit log to retain
  Databases               []DbConfig
}
```

The `Databases` field represents a slice (or array in JSON) of DbConfig structs.  The DbConfig struct is shown below:

```go
  Name                    string // Friendly name of the DB.  <Tenant>@<SID> is a good option here
  Hostname                string // Hostname or IP address of the primary HANA node
  Port                    uint   // Port of the HANA DB
  Username                string // HANA DB user name to use
  password                string // Password for HANA DB user
  CleanTrace              bool   // If true, trace file management will be enabled
  RetainTraceDays         uint   // Specifies the number of days of trace files to retain
  CleanBackupCatalog      bool   // If true, backup catalog truncation will be enabled
  RetainBackupCatalogDays uint   // Specifies the number of days of entries to retain
  DeleteOldBackups        bool   // If true, truncated files will be physically removed, if false entries are removed from the database only
  CleanAlerts             bool   // If true, old alerts are removed from the embedded statistics server
  RetainAlertsDays        uint   // Specifies the number of days of alerts to retain
  CleanLogVolume          bool   // If true, free log segments will be removed from the file system
  CleanAudit              bool   // If true, old audit records will be deleted
  RetainAuditDays         uint   // Specifies the number of days of audit log to retain
```

__Important notes about configuration!__

* All of the root level configuration parameters must be set
* Each database must be have the following fields set as a minimum:
  * Name
  * Hostname
  * Port
  * Username
* The Password field for each DB must be set either in the file or within the environment see [this section](#Reading-passwords-from-the-environment)
* Database level parameters that are not set will be inherited from the root level configuration

An example of a configuration file a single database is provided below.

```JSON
{
  "CleanTrace": true,
  "RetainTraceDays": 60,
  "CleanBackupCatalog": true,
  "RetainBackupCatalogDays" : 60,
  "DeleteOldBackups": true,
  "CleanAlerts": true,
  "RetainAlertsDays" : 60,
  "CleanLogVolume" : true,
  "CleanAudit": true,
  "RetainAuditDays": 60,
  "Databases":[
    {
      "Name": "systemdb_TST",
      "Hostname": "hanadb.mydomain.int",
      "Port": 30015,
      "Username": "hccuser",
      "Password": "AnUnsuitablePassw0rd"
    }
  ]
}
```

In the above configuration, all the database inherits all of the root level configuration.  Alternatively, database configurations can provide their own overrides by specifying fields that differ from the root config.  This is useful when working with many databases that share a common configuration with one or two exceptions.

## Reading passwords from the environment

If you don't want to source the database user passwords from the configuration, HCC can read passwords from an environment variable.  To do this, you should leave the password out of the configuration, and store the password in an environment variable which us database configuration name prefixed with `HCC_`.  For example, the following configuration would store the password in the environment variable `HCC_systemdb_TST`.

```JSON
  "Databases":[
    {
      "Name": "systemdb_TST",
      "Hostname": "hanadb.mydomain.int",
      "Port": 30015,
      "Username": "hccuser",
      }
  ]
  ```
