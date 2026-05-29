# goScan — Plan de implementación

Port scanner concurrente en Go, construido desde cero para demostrar dominio real del lenguaje: goroutines, canales, interfaces, CLI y tooling profesional.

**Repositorio:** `github.com/tu-usuario/goscan`  
**Requisitos:** Go 1.21+, conexión a internet  

---

## Estructura de directorios

```
goscan/
├── cmd/
│   └── goscan/
│       └── main.go              # Punto de entrada
├── internal/
│   ├── scan/
│   │   ├── scanner.go           # Dialer, ScanPort, Scanner, Run
│   │   ├── ports.go             # Parseado de rangos de puertos
│   │   ├── result.go            # Struct Result
│   │   ├── mock_test.go         # MockDialer (compartido entre tests)
│   │   ├── ports_test.go
│   │   └── scanner_test.go
│   └── report/
│       ├── reporter.go          # Interfaz Reporter + ForFile
│       ├── text.go              # Tabla en terminal (io.Writer)
│       ├── json.go              # Exportación JSON
│       ├── csv.go               # Exportación CSV
│       └── summary.go           # Resumen final (io.Writer)
├── pkg/
│   └── banner/
│       ├── grab.go              # Banner grabbing
│       ├── identify.go          # Identificación de servicios
│       └── banner_test.go
├── .github/
│   └── workflows/
│       └── ci.yml               # GitHub Actions
├── .goreleaser.yml              # Releases automáticos
├── Dockerfile                   # Multi-stage build
├── go.mod
├── go.sum
└── README.md
```

---

## Fase 1 — Scaffolding y CLI básica

**Objetivo:** proyecto compilable con flags definidos, parseado de puertos y signal handling.

### 1.1 Inicializar módulo y crear directorios

```bash
mkdir -p cmd/goscan internal/scan internal/report pkg/banner .github/workflows
go mod init github.com/tu-usuario/goscan
```

### 1.2 Dependencias

```bash
go get github.com/spf13/cobra@latest
go get github.com/schollz/progressbar/v3@latest
go get github.com/fatih/color@latest
```

### 1.3 `cmd/goscan/main.go`

Punto de entrada con flags, signal handling y validación de host.

```go
package main

import (
    "context"
    "fmt"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/spf13/cobra"
)

var (
    host    string
    ports   string
    timeout int
    workers int
    output  string
    banner  bool
    all     bool
    version bool
)

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
    rootCmd.Flags().BoolVarP(&version, "version", "v", false, "Mostrar versión")
    rootCmd.MarkFlagRequired("host")
}
```

La función `run` se completa incrementalmente en cada fase.

### 1.4 `internal/scan/ports.go`

Parsea tres formatos: `"80"`, `"80,443,8080"`, `"1-1024"` y combinaciones con/sin duplicados.

### 1.5 `internal/scan/ports_test.go`

Table-driven tests para `ParsePorts`.

**Verificación:**
```bash
go build ./...
go vet ./...
go test ./...
```

---

## Fase 2 — Scanner secuencial base

**Objetivo:** `Dialer` interface + `ScanPort` función + `Result` struct. Base testeable.

### 2.1 `internal/scan/result.go`

```go
package scan

import "time"

type Result struct {
    Host    string
    Port    int
    Open    bool
    Latency time.Duration
    Service string
    Banner  string
    Error   string
}
```

### 2.2 `internal/scan/scanner.go`

- `Dialer` interface abstrae `net.DialTimeout` (mockeable en tests)
- `NetDialer` implementación real
- `ScanPort` función secuencial (retorna `Result`)
- En Fase 3 se añade `Scanner` struct + `Run()` en el **mismo archivo**

### 2.3 `internal/scan/mock_test.go`

`MockDialer` compartido entre `ports_test.go` y `scanner_test.go`.

**Verificación:**
```bash
go test -race ./...
go run . --host scanme.nmap.org --ports 22,80,443
```

---

## Fase 3 — Worker pool concurrente

**Objetivo:** reemplazar loop secuencial con worker pool.

### 3.1 Añadir a `internal/scan/scanner.go`

- `Config` struct con Host, Ports, Timeout, Workers, GrabBanner
- `Scanner` struct con config + dialer
- `New(config)` y `NewWithDialer(config, dialer)`
- `Run(ctx)` con worker pool completo (jobs chan, results chan, WaitGroup)

### 3.2 Signal handling en `main.go`

Crear contexto cancelable con `signal.NotifyContext` para Ctrl+C.

### 3.3 `internal/scan/scanner_test.go`

Test concurrente con 100 puertos mockeados + benchmark.

**Verificación:**
```bash
go test -race ./...
go test -bench=. -benchmem ./internal/scan/
time go run . --host scanme.nmap.org --ports 1-1024 --workers 1
time go run . --host scanme.nmap.org --ports 1-1024 --workers 500
```

---

## Fase 4 — Banner grabbing

**Objetivo:** identificar servicios en puertos abiertos.

### 4.1 `pkg/banner/grab.go`

Lee hasta 256 bytes del socket; si el servicio es silencioso, envía probe HTTP.

### 4.2 `pkg/banner/identify.go`

Firma por prefijo de bytes → nombre de servicio (SSH, HTTP, FTP, TLS/SSL, etc.).

### 4.3 `pkg/banner/banner_test.go`

Test de `Identify` + `Grab` con `net.Pipe()`.

### 4.4 `internal/scan/scanner.go` — añadir `ScanPortWithBanner`

```go
func ScanPortWithBanner(dialer Dialer, host string, port int, timeout time.Duration) Result
```

Usa `banner.Grab` y `banner.Identify`. Se integra en el worker de `Run()`.

**Verificación:**
```bash
go run . --host scanme.nmap.org --ports 22,80,443 --banner
```

---

## Fase 5 — Output: tabla + exportación

**Objetivo:** UX completa con tabla en terminal coloreada, barra de progreso y exportación.

### 5.1 `internal/report/reporter.go`

Interfaz `Reporter` con método `Write([]scan.Result) error`.
Función `ForFile(path)` que retorna `JSONReporter`, `CSVReporter` o `TextReporter` según extensión.

### 5.2 `internal/report/text.go`

`PrintTable(w io.Writer, results []scan.Result, showAll bool)` — tabla alineada con `tabwriter` y colores.
`TextReporter` struct con `Write()` para exportación a archivo de texto.

### 5.3 `internal/report/json.go`

`JSONReporter` — escribe resultados como JSON indentado.

### 5.4 `internal/report/csv.go`

`CSVReporter` — escribe CSV con cabecera `host,port,open,service,latency_ms,banner`.

### 5.5 `internal/report/summary.go`

`Summary(w io.Writer, results []scan.Result, host string, elapsed time.Duration)`.

### 5.6 Progress bar en `Run()`

Integrar `progressbar/v3` en el aggregator del worker pool.

**Verificación:**
```bash
go run . --host scanme.nmap.org --ports 1-1024
go run . --host scanme.nmap.org --ports 1-1024 --output results.json
go run . --host scanme.nmap.org --ports 1-100 --all
```

---

## Fase 6 — Pulido: README, CI, Docker, releases

### 6.1 README.md

Badges (CI, Go version), demo GIF, instrucciones de instalación, ejemplos de uso, benchmarks, aviso legal.

### 6.2 `.github/workflows/ci.yml`

Push/PR a main: `go mod verify`, `go vet`, `go test -race -cover`, cross-build (linux/mac/windows).

### 6.3 `.goreleaser.yml`

Build multi-plataforma con `CGO_ENABLED=0`, `ldflags` para versión, archives tar.gz/zip, checksums.

### 6.4 Dockerfile

Multi-stage: builder `golang:1.21-alpine` → scratch con certificados CA.

---

## Comandos de referencia

```bash
# Uso
go run . --host scanme.nmap.org --ports 1-1024
go run . --host scanme.nmap.org --ports 22,80 --banner
go run . --host scanme.nmap.org --ports 1-1024 --output results.json

# Tests
go test ./...
go test -race ./...
go test -bench=. ./internal/scan/
go test -cover ./...

# Build
go build -o goscan ./cmd/goscan
GOOS=linux GOARCH=amd64 go build -o goscan-linux ./cmd/goscan

# Docker
docker build -t goscan .
docker run --rm goscan --host scanme.nmap.org --ports 80,443
```

---

## Conceptos de Go demostrados

| Concepto | Dónde aparece |
|----------|---------------|
| Goroutines | Worker pool en `scanner.go` |
| Canales (buffered) | `jobs chan int`, `results chan Result` |
| `sync.WaitGroup` | Coordinación del cierre del pool |
| `context.Context` | Cancelación y timeout global + signal handling |
| Interfaces | `Dialer`, `Reporter` |
| Table-driven tests | `ports_test.go`, `scanner_test.go` |
| Race detector | `go test -race` |
| `net.Pipe()` en tests | `banner_test.go`, `mock_test.go` |
| Multi-stage Docker | `Dockerfile` |
| Cross-compilation | CI y goreleaser |
| Signal handling | `main.go` con `signal.NotifyContext` |
