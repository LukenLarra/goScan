package scan

import (
	"fmt"
	"strconv"
	"strings"
)

func ParsePorts(input string) ([]int, error) {
	var ports []int
	seen := make(map[int]bool)

	parts := strings.Split(input, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.Contains(part, "-") {
			rangeParts := strings.SplitN(part, "-", 2)
			start, err := strconv.Atoi(rangeParts[0])
			if err != nil {
				return nil, fmt.Errorf("rango inválido: %s", part)
			}
			end, err := strconv.Atoi(rangeParts[1])
			if err != nil {
				return nil, fmt.Errorf("rango inválido: %s", part)
			}
			if start > end {
				return nil, fmt.Errorf("rango inválido: %d > %d", start, end)
			}
			for p := start; p <= end; p++ {
				if err := validatePort(p); err != nil {
					return nil, err
				}
				if !seen[p] {
					ports = append(ports, p)
					seen[p] = true
				}
			}
		} else {
			p, err := strconv.Atoi(part)
			if err != nil {
				return nil, fmt.Errorf("puerto inválido: %s", part)
			}
			if err := validatePort(p); err != nil {
				return nil, err
			}
			if !seen[p] {
				ports = append(ports, p)
				seen[p] = true
			}
		}
	}

	if len(ports) == 0 {
		return nil, fmt.Errorf("no se especificaron puertos válidos")
	}
	return ports, nil
}

func validatePort(p int) error {
	if p < 1 || p > 65535 {
		return fmt.Errorf("puerto fuera de rango (1-65535): %d", p)
	}
	return nil
}
