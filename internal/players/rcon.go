package players

import (
	"context"
	"fmt"
	"strings"

	"github.com/gorcon/rcon"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// GetConnectedPlayers connects to an RCON server and retrieves the list of connected players
func GetConnectedPlayers(ctx context.Context, conn *rcon.Conn) ([]string, error) {
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

// AddUser adds a new user to the server
func AddUser(ctx context.Context, conn *rcon.Conn, username, password string) error {
	logger := log.FromContext(ctx)
	cmd := fmt.Sprintf("adduser \"%s\" \"%s\"", username, password)
	logger.Info("Executing RCON command", "command", fmt.Sprintf("adduser \"%s\" \"****\"", username))
	response, err := conn.Execute(cmd)
	if err == nil {
		logger.Info("RCON command response", "response", response)
	}
	return err
}

// SetAccessLevel sets the access level for a user
func SetAccessLevel(ctx context.Context, conn *rcon.Conn, username, accessLevel string) error {
	logger := log.FromContext(ctx)
	cmd := fmt.Sprintf("setaccesslevel \"%s\" \"%s\"", username, accessLevel)
	logger.Info("Executing RCON command", "command", cmd)
	response, err := conn.Execute(cmd)
	if err == nil {
		logger.Info("RCON command response", "response", response)
	}
	return err
}

// BanUser bans a user from the server
func BanUser(ctx context.Context, conn *rcon.Conn, username string) error {
	logger := log.FromContext(ctx)
	cmd := fmt.Sprintf("banuser \"%s\"", username)
	logger.Info("Executing RCON command", "command", cmd)
	response, err := conn.Execute(cmd)
	if err == nil {
		logger.Info("RCON command response", "response", response)
	}
	return err
}

// UnbanUser unbans a user from the server
func UnbanUser(ctx context.Context, conn *rcon.Conn, username string) error {
	logger := log.FromContext(ctx)
	cmd := fmt.Sprintf("unbanuser \"%s\"", username)
	logger.Info("Executing RCON command", "command", cmd)
	response, err := conn.Execute(cmd)
	if err == nil {
		logger.Info("RCON command response", "response", response)
	}
	return err
}

// RemoveUser removes a user from the server's whitelist
func RemoveUser(ctx context.Context, conn *rcon.Conn, username string) error {
	logger := log.FromContext(ctx)
	cmd := fmt.Sprintf("removeuserfromwhitelist \"%s\"", username)
	logger.Info("Executing RCON command", "command", cmd)
	response, err := conn.Execute(cmd)
	if err == nil {
		logger.Info("RCON command response", "response", response)
	}
	return err
}
