package report

import (
	"fmt"
	"io"
	"os"
	"sort"
	"text/tabwriter"
	"time"

	"github.com/fatih/color"
	"github.com/tu-usuario/goscan/internal/scan"
)

var (
	green  = color.New(color.FgGreen, color.Bold).SprintFunc()
	red    = color.New(color.FgRed).SprintFunc()
	cyan   = color.New(color.FgCyan).SprintFunc()
	yellow = color.New(color.FgYellow).SprintFunc()
)

func PrintTable(w io.Writer, results []scan.Result, showAll bool) {
	sort.Slice(results, func(i, j int) bool {
		return results[i].Port < results[j].Port
	})

	tw := tabwriter.NewWriter(w, 0, 0, 3, ' ', 0)
	fmt.Fprintln(tw, cyan("PORT\tSTATE\tSERVICE\tLATENCY"))
	fmt.Fprintln(tw, "────\t─────\t───────\t───────")

	for _, r := range results {
		if !r.Open && !showAll {
			continue
		}
		state := green("open")
		if !r.Open {
			state = red("closed")
		}
		service := r.Service
		if service == "" {
			service = yellow("unknown")
		}
		fmt.Fprintf(tw, "%d\t%s\t%s\t%v\n",
			r.Port, state, service, r.Latency.Round(time.Millisecond))
	}
	tw.Flush()
}

type TextReporter struct {
	Path string
}

func (r *TextReporter) Write(results []scan.Result) error {
	f, err := os.Create(r.Path)
	if err != nil {
		return err
	}
	defer f.Close()

	PrintTable(f, results, true)
	return nil
}
