# goScan

**goScan** es un escáner de puertos TCP concurrente escrito en Go. Escanea miles de puertos por segundo usando un worker pool de goroutines, canales y `context.Context`. Incluye detección de servicios por banner grabbing, exportación a múltiples formatos, barra de progreso, y está diseñado desde cero para ser portable, testeable y fácil de extender.

---

## Cómo funciona

goScan sigue el patrón clásico de **worker pool**:

1. **Dispatcher** — envía los puertos a escanear a un canal buffered (`jobs chan int`)
2. **Workers** — N goroutines (configurable con `--workers`) leen del canal y ejecutan `net.DialTimeout` concurrentemente
3. **Aggregator** — recolecta los resultados desde el canal de resultados (`results chan Result`) y actualiza una barra de progreso
4. **Output** — al finalizar, imprime una tabla coloreada y opcionalmente exporta a JSON, CSV o TXT

Cada worker respeta la cancelación del `context.Context`, lo que permite interrupción graceful con Ctrl+C.

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

## Aviso legal

Usar únicamente en hosts propios o en [scanme.nmap.org](https://scanme.nmap.org) (servidor oficial de pruebas autorizado por Nmap). Escanear hosts sin permiso puede ser ilegal. El autor no se responsabiliza del mal uso de esta herramienta.
