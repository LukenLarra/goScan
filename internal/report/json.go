package report

import (
	"encoding/json"
	"os"

	"github.com/tu-usuario/goscan/internal/scan"
)

type JSONReporter struct {
	Path string
}

func (r *JSONReporter) Write(results []scan.Result) error {
	f, err := os.Create(r.Path)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(results)
}
