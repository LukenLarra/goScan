package report

import (
	"fmt"
	"io"
	"time"

	"github.com/tu-usuario/goscan/internal/scan"
)

func Summary(w io.Writer, results []scan.Result, host string, elapsed time.Duration) {
	open, total := 0, len(results)
	var totalLatency time.Duration

	for _, r := range results {
		if r.Open {
			open++
			totalLatency += r.Latency
		}
	}

	var avgLatency time.Duration
	if open > 0 {
		avgLatency = totalLatency / time.Duration(open)
	}

	fmt.Fprintf(w, "\n── Resumen ──────────────────────────\n")
	fmt.Fprintf(w, "  Host:           %s\n", host)
	fmt.Fprintf(w, "  Puertos totales: %d\n", total)
	fmt.Fprintf(w, "  Puertos abiertos: %d\n", open)
	fmt.Fprintf(w, "  Latencia media:  %v\n", avgLatency.Round(time.Millisecond))
	fmt.Fprintf(w, "  Tiempo total:    %v\n", elapsed.Round(time.Millisecond))
	fmt.Fprintf(w, "─────────────────────────────────────\n")
}
