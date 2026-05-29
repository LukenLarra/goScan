package scan

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestScanPort(t *testing.T) {
	mock := MockDialer{OpenPorts: map[int]bool{80: true, 443: true}}

	result := ScanPort(mock, "localhost", 80, time.Second)
	if !result.Open {
		t.Error("Puerto 80 debería estar abierto")
	}

	result = ScanPort(mock, "localhost", 8080, time.Second)
	if result.Open {
		t.Error("Puerto 8080 debería estar cerrado")
	}
}

func TestScannerConcurrent(t *testing.T) {
	openPorts := map[int]bool{
		10: true, 20: true, 30: true, 40: true, 50: true,
		60: true, 70: true, 80: true, 90: true, 100: true,
	}

	var ports []int
	for p := 1; p <= 100; p++ {
		ports = append(ports, p)
	}

	config := Config{
		Host:    "localhost",
		Ports:   ports,
		Timeout: 500 * time.Millisecond,
		Workers: 20,
	}

	mock := MockDialer{OpenPorts: openPorts}
	scanner := NewWithDialer(config, mock)

	results, err := scanner.Run(context.Background())
	if err != nil {
		t.Fatalf("Run inesperadamente retornó error: %v", err)
	}

	if len(results) != 100 {
		t.Errorf("esperaba 100 resultados, got %d", len(results))
	}

	open := countOpen(results)
	if open != 10 {
		t.Errorf("esperaba 10 puertos abiertos, got %d", open)
	}
}

func BenchmarkScanner(b *testing.B) {
	ports := make([]int, 1000)
	for i := range ports {
		ports[i] = i + 1
	}
	mock := MockDialer{OpenPorts: map[int]bool{80: true}}

	for _, workers := range []int{1, 10, 50, 200, 500} {
		b.Run(fmt.Sprintf("workers-%d", workers), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				config := Config{Host: "localhost", Ports: ports,
					Timeout: time.Millisecond, Workers: workers}
				NewWithDialer(config, mock).Run(context.Background())
			}
		})
	}
}
