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
			n := &networkApplier{}
			if got := n.Compare(tt.source, tt.target); got != tt.want {
				t.Errorf("NetworkApplier.Compare() = %v, want %v", got, tt.want)
			}
		})
	}
}
