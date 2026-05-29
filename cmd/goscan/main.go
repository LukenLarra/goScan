package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/tu-usuario/goscan/internal/report"
	"github.com/tu-usuario/goscan/internal/scan"
)

var (
	host    string
	ports   string
	timeout int
	workers int
	output  string
	banner  bool
	all     bool
	showVer bool
)

var versionStr = "dev"

var rootCmd = &cobra.Command{
	Use:   "goscan",
	Short: "Port scanner concurrente escrito en Go",
	Long:  "goScan escanea puertos TCP usando un worker pool de goroutines.",
	RunE:  run,
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVarP(&host, "host", "H", "", "Host o IP a escanear (obligatorio)")
	rootCmd.Flags().StringVarP(&ports, "ports", "p", "1-1024", "Puertos: '80', '80,443', '1-1024'")
	rootCmd.Flags().IntVarP(&timeout, "timeout", "t", 1000, "Timeout por puerto en milisegundos")
	rootCmd.Flags().IntVarP(&workers, "workers", "w", 200, "Número de goroutines concurrentes")
	rootCmd.Flags().StringVarP(&output, "output", "o", "", "Archivo de salida (.json, .csv, .txt)")
	rootCmd.Flags().BoolVarP(&banner, "banner", "b", false, "Intentar banner grabbing")
	rootCmd.Flags().BoolVarP(&all, "all", "a", false, "Mostrar puertos cerrados también")
	rootCmd.Flags().BoolVarP(&showVer, "version", "v", false, "Mostrar versión")
	rootCmd.MarkFlagRequired("host")
}

func run(cmd *cobra.Command, args []string) error {
	if showVer {
		fmt.Printf("goscan %s\n", versionStr)
		return nil
	}

	portList, err := scan.ParsePorts(ports)
	if err != nil {
		return fmt.Errorf("error parseando puertos: %w", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	config := scan.Config{
		Host:       host,
		Ports:      portList,
		Timeout:    time.Duration(timeout) * time.Millisecond,
		Workers:    workers,
		GrabBanner: banner,
	}

	scanner := scan.New(config)

	fmt.Printf("Escaneando %s (%d puertos, %d workers)...\n\n",
		host, len(portList), workers)

	start := time.Now()
	results, err := scanner.Run(ctx)
	elapsed := time.Since(start)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Scan interrumpido: %v\n", err)
	}

	report.PrintTable(os.Stdout, results, all)

	if output != "" {
		reporter := report.ForFile(output)
		if err := reporter.Write(results); err != nil {
			return fmt.Errorf("error escribiendo archivo: %w", err)
		}
		fmt.Printf("\nResultados guardados en: %s\n", output)
	}

	report.Summary(os.Stdout, results, host, elapsed)

	return nil
}
