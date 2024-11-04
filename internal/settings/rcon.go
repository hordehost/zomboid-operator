package settings

import (
	"fmt"
	"strings"
	"time"

	"github.com/gorcon/rcon"
	zomboidv1 "github.com/hordehost/zomboid-operator/api/v1"
)

// GetServerOptions connects to an RCON server, executes the showoptions command,
// and returns the output. It handles connection and cleanup automatically.
func GetServerOptions(hostname string, port int, password string) (zomboidv1.ZomboidSettings, error) {
	address := fmt.Sprintf("%s:%d", hostname, port)

	conn, err := rcon.Dial(
		address, password,
		rcon.SetDialTimeout(5*time.Second),
		rcon.SetDeadline(5*time.Second),
	)
	if err != nil {
		return zomboidv1.ZomboidSettings{}, fmt.Errorf("failed to connect to RCON server: %w", err)
	}
	defer conn.Close()

	response, err := conn.Execute("showoptions")
	if err != nil {
		return zomboidv1.ZomboidSettings{}, fmt.Errorf("failed to execute showoptions command: %w", err)
	}

	return ParseRCONShowOptions(response), nil
}

// ParseRCONShowOptions parses the output of the RCON "showoptions" command into server settings
func ParseRCONShowOptions(output string) zomboidv1.ZomboidSettings {
	settings := zomboidv1.ZomboidSettings{}
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || !strings.Contains(line, "=") {
			continue
		}

		// Remove any "*" prefix that the RCON output includes
		line = strings.TrimPrefix(line, "* ")

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		if value != "" {
			ParseSettingValue(&settings, key, value)
		}
	}

	return settings
}
