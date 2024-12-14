package settings

import (
	"context"
	"fmt"
	"strings"

	"github.com/gorcon/rcon"
	zomboidv1 "github.com/zomboidhost/zomboid-operator/api/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// ReadServerOptions executes the showoptions command and updates the provided settings object
// with the current server settings.
func ReadServerOptions(ctx context.Context, conn *rcon.Conn, settings *zomboidv1.ZomboidSettings) error {
	response, err := conn.Execute("showoptions")
	if err != nil {
		return fmt.Errorf("failed to execute showoptions command: %w", err)
	}

	ParseRCONShowOptions(response, settings)
	return nil
}

// ParseRCONShowOptions parses the output of the RCON "showoptions" command into the provided settings object
func ParseRCONShowOptions(output string, settings *zomboidv1.ZomboidSettings) {
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
			ParseSettingValue(settings, key, value)
		}
	}
}

// ApplySettingsUpdates applies the given settings changes via RCON
func ApplySettingsUpdates(ctx context.Context, conn *rcon.Conn, updates [][2]string, settings *zomboidv1.ZomboidSettings) error {
	logger := log.FromContext(ctx)

	for _, update := range updates {
		settingName := update[0]
		settingValue := update[1]

		command := fmt.Sprintf("changeoption %s \"%s\"", settingName, settingValue)
		resultLine, err := conn.Execute(command)
		if err != nil {
			return fmt.Errorf("failed to execute RCON command %q: %w", command, err)
		}

		parts := strings.Split(resultLine, " : ")
		if len(parts) != 3 {
			return fmt.Errorf("unexpected RCON response format: %s", resultLine)
		}

		newValue := strings.TrimSpace(parts[2])
		// If the desired setting was "", that means we want to remove the setting from the server
		// and return to the default value.
		if settingValue != "" && newValue != settingValue {
			return fmt.Errorf("setting %s was not updated correctly. Expected %q but got %q", settingName, settingValue, newValue)
		}

		// Update the settings object with the confirmed value
		if settings != nil {
			ParseSettingValue(settings, settingName, newValue)
		}

		logger.Info("Applied setting change", "setting", settingName, "value", settingValue)
	}

	return nil
}

// RestartServer sends the quit command to the RCON server to restart it
func RestartServer(ctx context.Context, conn *rcon.Conn) error {
	_, err := conn.Execute("quit")
	return err
}
