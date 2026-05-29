package report

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"

	"github.com/tu-usuario/goscan/internal/scan"
)

type CSVReporter struct {
	Path string
}

func (r *CSVReporter) Write(results []scan.Result) error {
	f, err := os.Create(r.Path)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	w.Write([]string{"host", "port", "open", "service", "latency_ms", "banner"})
	for _, res := range results {
		w.Write([]string{
			res.Host,
			strconv.Itoa(res.Port),
			strconv.FormatBool(res.Open),
			res.Service,
			fmt.Sprintf("%d", res.Latency.Milliseconds()),
			res.Banner,
		})
	}
	return w.Error()
}
