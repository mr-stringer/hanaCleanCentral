package main

/*This file contains helper types for handling the results of database queries*/

//Struct to hold information about tracefiles
type TraceFile struct {
	Hostname     string
	TraceFile    string
	SizeBytes    uint64
	LastModified string
}

//Struct to hold information about backup file
type BackupFiles struct {
	EntryType string
	FileCount uint
	Bytes     uint64
}

//Struct to hold information about data volumes
type DataVolume struct {
	Host           string
	Port           uint
	UsedSizeBytes  uint64
	TotalSizeBytes uint64
}

func (d DataVolume) CleanNeeded() bool {
	return float32(d.UsedSizeBytes)/float32(d.TotalSizeBytes)*100 < 50
}
