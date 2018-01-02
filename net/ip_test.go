package net

import (
	"reflect"
	"testing"
)

func TestIPv4(t *testing.T) {
	ips := []struct {
		in  string
		out IP
	}{
		{"127.0.1.2", IPv4(127, 0, 1, 2)},
	}

	for _, ip := range ips {
		if out := ParseIP(ip.in); !reflect.DeepEqual(out, ip.out) {
			t.Errorf("ParseIP(%q) = %v, want %v", ip.in, out, ip.out)
		}
	}
}
