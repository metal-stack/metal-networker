package net

import "testing"

func TestNetworkApplier_Compare(t *testing.T) {
	tests := []struct {
		name   string
		source string
		target string
		want   bool
	}{
		{
			name:   "simple test",
			source: "/etc/hostname",
			target: "/etc/passwd",
			want:   false,
		},
		{
			name:   "simple test",
			source: "/etc/hostname",
			target: "/etc/hostname",
			want:   true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			n := &NetworkApplier{}
			if got := n.compare(tt.source, tt.target); got != tt.want {
				t.Errorf("NetworkApplier.compare() = %v, want %v", got, tt.want)
			}
		})
	}
}
