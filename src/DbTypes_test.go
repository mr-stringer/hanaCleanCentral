package main

import "testing"

func TestDataVolume_CleanNeeded(t *testing.T) {
	tests := []struct {
		name string
		d    DataVolume
		want bool
	}{
		{"tc01", DataVolume{"localhost", 30015, 250, 1000}, true},
		{"tc01", DataVolume{"localhost", 30015, 750, 1000}, false},
		{"tc01", DataVolume{"longhostname.longdomain.com", 30040, 1234567890, 1234567899}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.d.CleanNeeded(); got != tt.want {
				t.Errorf("DataVolume.CleanNeeded() = %v, want %v", got, tt.want)
			}
		})
	}
}
