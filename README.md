# goScan

[![CI](https://github.com/tu-usuario/goscan/actions/workflows/ci.yml/badge.svg)](https://github.com/tu-usuario/goscan/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/badge/go-1.21+-blue.svg)](https://go.dev)

**goScan** es un escáner de puertos TCP concurrente escrito en Go. Escanea miles de puertos por segundo usando un worker pool de goroutines, canales y `context.Context`. Incluye detección de servicios por banner grabbing, exportación a múltiples formatos, barra de progreso, y está diseñado desde cero para ser portable, testeable y fácil de extender.

---

## Cómo funciona

goScan sigue el patrón clásico de **worker pool**:

1. **Dispatcher** — envía los puertos a escanear a un canal buffered (`jobs chan int`)
2. **Workers** — N goroutines (configurable con `--workers`) leen del canal y ejecutan `net.DialTimeout` concurrentemente
3. **Aggregator** — recolecta los resultados desde el canal de resultados (`results chan Result`) y actualiza una barra de progreso
4. **Output** — al finalizar, imprime una tabla coloreada y opcionalmente exporta a JSON, CSV o TXT

Cada worker respeta la cancelación del `context.Context`, lo que permite interrupción graceful con Ctrl+C.

```
┌────────────┐   jobs chan   ┌──────────┐   results chan   ┌────────────┐
│ Dispatcher │─────ports─────▶│ Workers  │─────results─────▶│ Aggregator │
│            │                │ (N goros)│                  │            │
│  range     │                │          │                  │  progress  │
│  ports[]   │                │ Dial(..) │                  │  bar +     │
│            │                │          │                  │  collect   │
└────────────┘                └──────────┘                  └─────┬──────┘
                                                                  │
                                                           ┌──────▼──────┐
                                                           │   Output    │
                                                           │ table,JSON, │
                                                           │ CSV,summary │
                                                           └─────────────┘
```

## Arquitectura

```
goscan/
├── cmd/goscan/main.go       # Punto de entrada, flags cobra, signal handling
├── internal/
│   ├── scan/                # Núcleo del escáner
│   │   ├── scanner.go       #   Dialer interface, ScanPort, Scanner + Run()
│   │   ├── ports.go         #   Parseo de rangos de puertos
│   │   ├── result.go        #   Struct Result
│   │   └── *_test.go        #   Tests con MockDialer (net.Pipe)
│   └── report/              # Salida y exportación
│       ├── reporter.go      #   Interfaz Reporter + factory ForFile
│       ├── text.go          #   Tabla coloreada en terminal
│       ├── json.go          #   Exportación JSON indentada
│       ├── csv.go           #   Exportación CSV
│       └── summary.go       #   Resumen final
└── pkg/
    └── banner/              # Identificación de servicios
        ├── grab.go          #   Banner grabbing con probe HTTP
        ├── identify.go      #   Firma por prefijo de bytes
        └── *_test.go        #   Tests con net.Pipe
```

## Conceptos de Go que demuestra

| Concepto | Implementación |
|---|---|
| **Goroutines** | Worker pool con N workers concurrentes |
| **Canales buffered** | `jobs chan int`, `results chan Result` con buffer = len(puertos) |
| **sync.WaitGroup** | Coordinación del cierre de workers y del canal de results |
| **context.Context** | Cancelación global + signal.NotifyContext para Ctrl+C |
| **Interfaces** | `Dialer` (mockeable en tests), `Reporter` (JSON/CSV/Text) |
| **Table-driven tests** | Tests de parseo, scanner y banner |
| **Race detector** | `go test -race` verifica ausencia de data races |
| **net.Pipe()** | Simulación de conexiones TCP sin red real en tests |
| **Multi-stage Docker** | Builder + scratch para binario estático mínimo |
| **Cross-compilation** | CI y GoReleaser para linux/windows/darwin, amd64/arm64 |

## Instalación

```bash
go install github.com/tu-usuario/goscan@latest
```

O descarga el binario desde [Releases](https://github.com/tu-usuario/goscan/releases).

## Uso

```bash
# Escaneo básico de los 1024 primeros puertos
goscan --host scanme.nmap.org --ports 1-1024

# Con detección de servicio (banner grabbing)
goscan --host scanme.nmap.org --ports 22,80,443 --banner

# Escaneo completo con exportación
goscan --host scanme.nmap.org --ports 1-65535 --workers 1000 --output results.json

# Mostrar puertos cerrados también
goscan --host scanme.nmap.org --ports 1-100 --all

# Ayuda completa
goscan --help
```

### Flags

| Flag | Default | Descripción |
|---|---|---|
| `--host` `-H` | (obligatorio) | Host o IP a escanear |
| `--ports` `-p` | `1-1024` | Puertos: `80`, `80,443`, `1-1024` |
| `--timeout` `-t` | `1000` | Timeout por puerto en milisegundos |
| `--workers` `-w` | `200` | Número de goroutines concurrentes |
| `--output` `-o` | `""` | Archivo de salida (.json, .csv, .txt) |
| `--banner` `-b` | `false` | Intentar banner grabbing y detectar servicio |
| `--all` `-a` | `false` | Mostrar puertos cerrados en la tabla |
| `--version` `-v` | `false` | Mostrar versión del binario |

## Benchmarks

Midiendo tiempo de escaneo de 1024 puertos contra `scanme.nmap.org`:

| Workers | Tiempo | Speedup |
|---|---|---|
| 1 (secuencial) | ~18s | 1x |
| 50 | ~1.2s | 15x |
| 200 (default) | ~0.4s | 45x |
| 500 | ~0.3s | 60x |

```bash
go test -bench=. -benchmem ./internal/scan/
```

## Desarrollo

```bash
# Compilar
go build -o goscan ./cmd/goscan

# Tests
go test ./...                     # todos los tests
go test -race ./...               # con race detector
go test -cover ./...              # cobertura
go test -bench=. ./internal/scan/ # benchmarks

# Linter estático
go vet ./...
```

## Docker

```bash
docker build -t goscan .
docker run --rm goscan --host scanme.nmap.org --ports 80,443
```

## Releases

Al crear un tag, GoReleaser genera binarios automáticamente:

```bash
git tag v1.0.0
git push origin v1.0.0
```

## Aviso legal

Usar únicamente en hosts propios o en [scanme.nmap.org](https://scanme.nmap.org) (servidor oficial de pruebas autorizado por Nmap). Escanear hosts sin permiso puede ser ilegal. El autor no se responsabiliza del mal uso de esta herramienta.
