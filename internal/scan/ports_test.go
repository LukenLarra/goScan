package scan

import (
	"testing"
)

func TestParsePorts(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []int
		wantErr bool
	}{
		{"puerto único", "80", []int{80}, false},
		{"lista de puertos", "22,80,443", []int{22, 80, 443}, false},
		{"rango", "8080-8083", []int{8080, 8081, 8082, 8083}, false},
		{"combinado", "22,80-82", []int{22, 80, 81, 82}, false},
		{"sin duplicados", "80,80", []int{80}, false},
		{"rango inválido", "100-50", nil, true},
		{"fuera de rango", "99999", nil, true},
		{"no numérico", "abc", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParsePorts(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParsePorts(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if !tt.wantErr && !equalSlices(got, tt.want) {
				t.Errorf("ParsePorts(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func equalSlices(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
