package scan

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/schollz/progressbar/v3"
	"github.com/tu-usuario/goscan/pkg/banner"
)

type Dialer interface {
	Dial(network, address string, timeout time.Duration) (net.Conn, error)
}

type NetDialer struct{}

func (d NetDialer) Dial(network, address string, timeout time.Duration) (net.Conn, error) {
	return net.DialTimeout(network, address, timeout)
}

func ScanPort(dialer Dialer, host string, port int, timeout time.Duration) Result {
	address := fmt.Sprintf("%s:%d", host, port)
	start := time.Now()

	conn, err := dialer.Dial("tcp", address, timeout)
	latency := time.Since(start)

	if err != nil {
		return Result{
			Host:    host,
			Port:    port,
			Open:    false,
			Latency: latency,
			Error:   err.Error(),
		}
	}
	conn.Close()

	return Result{
		Host:    host,
		Port:    port,
		Open:    true,
		Latency: latency,
	}
}

func ScanPortWithBanner(dialer Dialer, host string, port int, timeout time.Duration) Result {
	address := fmt.Sprintf("%s:%d", host, port)
	start := time.Now()

	conn, err := dialer.Dial("tcp", address, timeout)
	latency := time.Since(start)

	if err != nil {
		return Result{
			Host:    host,
			Port:    port,
			Open:    false,
			Latency: latency,
			Error:   err.Error(),
		}
	}
	defer conn.Close()

	result := Result{
		Host:    host,
		Port:    port,
		Open:    true,
		Latency: latency,
	}

	bannerData := banner.Grab(conn, timeout/2)
	if len(bannerData) > 0 {
		result.Service = banner.Identify(bannerData)
		result.Banner = string(bannerData[:min(len(bannerData), 80)])
	}

	return result
}

type Config struct {
	Host       string
	Ports      []int
	Timeout    time.Duration
	Workers    int
	GrabBanner bool
}

type Scanner struct {
	config Config
	dialer Dialer
}

func New(config Config) *Scanner {
	return &Scanner{config: config, dialer: NetDialer{}}
}

func NewWithDialer(config Config, dialer Dialer) *Scanner {
	return &Scanner{config: config, dialer: dialer}
}

func (s *Scanner) Run(ctx context.Context) ([]Result, error) {
	jobs := make(chan int, len(s.config.Ports))
	results := make(chan Result, len(s.config.Ports))

	bar := progressbar.NewOptions(len(s.config.Ports),
		progressbar.OptionSetDescription("Escaneando"),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "█",
			SaucerPadding: "░",
			BarStart:      "│",
			BarEnd:        "│",
		}),
		progressbar.OptionShowCount(),
		progressbar.OptionShowIts(),
		progressbar.OptionSetItsString("ports/s"),
	)

	var wg sync.WaitGroup
	for i := 0; i < s.config.Workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for port := range jobs {
				select {
				case <-ctx.Done():
					return
				default:
				}

				var result Result
				if s.config.GrabBanner {
					result = ScanPortWithBanner(s.dialer, s.config.Host, port, s.config.Timeout)
				} else {
					result = ScanPort(s.dialer, s.config.Host, port, s.config.Timeout)
				}
				results <- result
			}
		}()
	}

	go func() {
		for _, port := range s.config.Ports {
			select {
			case <-ctx.Done():
				close(jobs)
				return
			case jobs <- port:
			}
		}
		close(jobs)
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	var all []Result
	for result := range results {
		bar.Add(1)
		all = append(all, result)
	}
	fmt.Println()

	if ctx.Err() != nil {
		return all, ctx.Err()
	}

	return all, nil
}

func countOpen(results []Result) int {
	n := 0
	for _, r := range results {
		if r.Open {
			n++
		}
	}
	return n
}
