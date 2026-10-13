// SPDX-FileCopyrightText: (c) Mauve Mailorder Software GmbH & Co. KG, 2022. Licensed under [MIT](LICENSE) license.
//
// SPDX-License-Identifier: MIT

package main

import (
	"testing"
)

func TestParseTarget(t *testing.T) {
	tests := []struct {
		raw     string
		want    string
		wantErr bool
	}{
		{raw: "172.16.0.200", want: "172.16.0.200"},
		{raw: "ilo-01.example.com", want: "ilo-01.example.com"},
		{raw: "ilo-01.example.com:8443", want: "ilo-01.example.com:8443"},
		{raw: "::1", want: "[::1]"},
		{raw: "[2001:db8::1]:443", want: "[2001:db8::1]:443"},
		{raw: "", wantErr: true},
		{raw: "evil.example#", wantErr: true},
		{raw: "evil.example/foo", wantErr: true},
		{raw: "evil.example?x=1", wantErr: true},
		{raw: "user:pass@evil.example", wantErr: true},
		{raw: "ilo 01", wantErr: true},
		{raw: "ilo\n01", wantErr: true},
		{raw: "ilo:0", wantErr: true},
		{raw: "ilo:99999", wantErr: true},
		{raw: "ilo:abc", wantErr: true},
		{raw: "\xff", wantErr: true},
		{raw: "fe80::1%25eth0", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.raw, func(t *testing.T) {
			got, err := parseTarget(tt.raw)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got %q", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTargetFilter(t *testing.T) {
	f, err := parseTargetFilter("10.0.0.0/8, ilo-01.Example.com,192.168.1.10,2001:db8::/32")
	if err != nil {
		t.Fatal(err)
	}

	tests := map[string]bool{
		"10.1.2.3":               true,
		"10.1.2.3:443":           true,
		"11.1.2.3":               false,
		"ilo-01.example.com":     true,
		"ILO-01.example.com:443": true,
		"ilo-02.example.com":     false,
		"192.168.1.10":           true,
		"192.168.1.11":           false,
		"[2001:db8::1]":          true,
		"[2001:db9::1]:443":      false,
	}

	for addr, want := range tests {
		if got := f.allows(addr); got != want {
			t.Errorf("allows(%q) = %v, want %v", addr, got, want)
		}
	}

	var none *targetFilter
	if !none.allows("anything") {
		t.Error("nil filter should allow all targets")
	}

	if _, err := parseTargetFilter("10.0.0.0/33"); err == nil {
		t.Error("expected error for invalid prefix")
	}
}
