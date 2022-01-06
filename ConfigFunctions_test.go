package main

import (
	"reflect"
	"testing"
)

func TestGetConfigFromFile(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		want    Config
		wantErr bool
	}{
		{"GoodFile01", args{"testfiles/configtest01.json"}, Config{[]DbConfig{{"systemdb@TST", "hanadb.mydomain.int", 30015, "sstringer", "ReallzyKoolPassw0rd", 14}}}, false},
		{"GoodFile02", args{"testfiles/configtest02.json"}, Config{[]DbConfig{{"Test", "hanadb.mydomain.int", 30015, "sstringer", "ReallzyKoolPassw0rd", 14}, {"Prod", "localhost", 30040, "sstringer", "ZeroPasswordsRGUD", 14}}}, false},
		{"InvalidJson", args{"testfiles/invalidJson.json"}, Config{}, true},
		{"InvalidPAth", args{"testfiles/NOFILE.json"}, Config{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetConfigFromFile(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetConfigFromFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetConfigFromFile() = %v, want %v", got, tt.want)
			}
		})
	}
}
