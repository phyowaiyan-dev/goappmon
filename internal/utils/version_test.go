package utils

import "testing"

func TestCompareSemanticVersion(t *testing.T) {
	tests := []struct {
		name string
		a    string
		b    string
		want int
	}{
		{name: "equal", a: "1.0.0", b: "1.0.0", want: 0},
		{name: "greater major", a: "2.0.0", b: "1.9.9", want: 1},
		{name: "greater minor", a: "1.2.0", b: "1.1.9", want: 1},
		{name: "greater patch", a: "1.2.3", b: "1.2.2", want: 1},
		{name: "less", a: "1.2.2", b: "1.2.3", want: -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CompareSemanticVersion(tt.a, tt.b)
			if err != nil {
				t.Fatalf("compare version: %v", err)
			}
			if got != tt.want {
				t.Fatalf("compare version = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestParseSemanticVersion(t *testing.T) {
	if _, err := ParseSemanticVersion("1.2"); err == nil {
		t.Fatal("expected invalid version format")
	}
	if _, err := ParseSemanticVersion("1.2.3"); err != nil {
		t.Fatalf("parse semantic version: %v", err)
	}
}
