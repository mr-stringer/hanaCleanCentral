# hanaCleanCentral

A centralised maintenance tool for the HANA database.

hanaCleanCentral is based on [hanacleaner](https://github.com/chriselswede/hanacleaner).  This project aims to perform all of the housekeeping tasks performed by hanacleaner but do so in a centralised manner.  Rather than being installed upon each HANA database, hanaCleanCentral should be installed on a central server where it will remotely housekeep many HANA instance.  hanaCleanCentral can perform maintainence on one remote database at a time or many in parallel.

Due to the way that tracefiles are deleted, we need to go into each system and tenant DB individually.  This has the advantage that each tenant can have individual settings, meaning that, for example, two tenant DBs on a single SID could have different tracefile retention periods and so on.

## Required Privs

|Application area |Type | Value |
|---|---|---|
|General|Role|`MANAGEMENT`|
|TraceFile management |Privilege|`TRACE ADMIN`|
|Backup catalog management|Privilege|`BACKUP OPERATOR`|
