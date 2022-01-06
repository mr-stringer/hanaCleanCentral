package main

/*This file contains helper types for handling the results of database queries*/

type TraceFile struct {
	Hostname     string
	TraceFile    string
	SizeBytes    uint64
	LastModified string
}
