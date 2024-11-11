package players

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gorcon/rcon"
)

// GetConnectedPlayers connects to an RCON server and retrieves the list of connected players
func GetConnectedPlayers(ctx context.Context, hostname string, port int, password string) ([]string, error) {
	address := fmt.Sprintf("%s:%d", hostname, port)

	conn, err := rcon.Dial(
		address, password,
		rcon.SetDialTimeout(5*time.Second),
		rcon.SetDeadline(5*time.Second),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RCON server: %w", err)
	}
	defer conn.Close()

	response, err := conn.Execute("players")
	if err != nil {
		return nil, fmt.Errorf("failed to execute players command: %w", err)
	}

	// If no players are connected, return empty slice
	if response == "" {
		return []string{}, nil
	}

	// Split response into lines and process each line
	lines := strings.Split(response, "\n")
	var players []string

	for i, line := range lines {
		// Skip empty lines and first line (player count)
		if line == "" || i == 0 {
			continue
		}

		// Remove leading "-" and trim whitespace
		if strings.HasPrefix(line, "-") {
			line = strings.TrimSpace(line[1:])
			players = append(players, line)
		}
	}

	return players, nil
}
