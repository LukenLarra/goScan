package report

import (
	"strings"

	"github.com/tu-usuario/goscan/internal/scan"
)

type Reporter interface {
	Write(results []scan.Result) error
}

func ForFile(path string) Reporter {
	switch {
	case strings.HasSuffix(path, ".json"):
		return &JSONReporter{Path: path}
	case strings.HasSuffix(path, ".csv"):
		return &CSVReporter{Path: path}
	default:
		return &TextReporter{Path: path}
	}
}
