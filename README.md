# goScan

[![CI](https://github.com/tu-usuario/goscan/actions/workflows/ci.yml/badge.svg)](https://github.com/tu-usuario/goscan/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/badge/go-1.21+-blue.svg)](https://go.dev)

Port scanner concurrente en Go. Usa un worker pool de goroutines con canales
para escanear miles de puertos por segundo.

## Instalación

```bash
go install github.com/tu-usuario/goscan@latest
```

O descarga el binario desde [Releases](https://github.com/tu-usuario/goscan/releases).

## Uso

```bash
# Escaneo básico
goscan --host scanme.nmap.org --ports 1-1024

# Con detección de servicio
goscan --host 192.168.1.1 --ports 22,80,443,8080-8090 --banner

# Exportar resultados
goscan --host example.com --ports 1-65535 --workers 1000 --output results.json
```

## Flags

| Flag | Default | Descripción |
|------|---------|-------------|
| `--host` `-H` | (obligatorio) | Host o IP a escanear |
| `--ports` `-p` | `1-1024` | Puertos: `80`, `80,443`, `1-1024` |
| `--timeout` `-t` | `1000` | Timeout por puerto en ms |
| `--workers` `-w` | `200` | Goroutines concurrentes |
| `--output` `-o` | `""` | Archivo de salida (.json, .csv, .txt) |
| `--banner` `-b` | `false` | Intentar banner grabbing |
| `--all` `-a` | `false` | Mostrar puertos cerrados |
| `--version` `-v` | `false` | Mostrar versión |

## Benchmarks

| Workers | Puertos | Tiempo |
|---------|---------|--------|
| 1 | 1024 | ~18s |
| 50 | 1024 | ~1.2s |
| 200 | 1024 | ~0.4s |
| 500 | 1024 | ~0.3s |

## Desarrollo

```bash
go test ./...
go test -race ./...
go test -bench=. ./internal/scan/
go build -o goscan ./cmd/goscan
```

## Docker

```bash
docker build -t goscan .
docker run --rm goscan --host scanme.nmap.org --ports 80,443
```

## Aviso legal

Usar únicamente en hosts propios o en scanme.nmap.org (servidor oficial para pruebas).
Escanear hosts sin permiso puede ser ilegal.
