package players

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	zomboidv1 "github.com/hordehost/zomboid-operator/api/v1"
)

// GetAllowlist queries the ws4sqlite service to retrieve the current allowlist
func GetAllowlist(hostname string, port int, serverName string) ([]zomboidv1.AllowlistUser, error) {
	sqliteUrl := fmt.Sprintf("http://%s:%d/%s", hostname, port, serverName)
	queries, err := json.Marshal(map[string]interface{}{
		"transaction": []map[string]string{
			{"query": "SELECT * FROM whitelist"},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal queries: %w", err)
	}

	resp, err := http.Post(sqliteUrl, "application/json", bytes.NewReader(queries))
	if err != nil {
		return nil, fmt.Errorf("failed to make HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("received non-2xx response: %d, failed to read body: %w", resp.StatusCode, err)
		}
		return nil, fmt.Errorf("received non-2xx response: %d, body: %s", resp.StatusCode, string(body))
	}

	var response struct {
		Results []struct {
			ResultSet []struct {
				ID             int         `json:"id"`
				World          string      `json:"world"`
				Username       string      `json:"username"`
				Password       string      `json:"password"`
				Admin          int         `json:"admin"`
				Moderator      int         `json:"moderator"`
				Banned         interface{} `json:"banned"`
				Priority       *int        `json:"priority"`
				LastConnection string      `json:"lastConnection"`
				SteamID        string      `json:"steamid"`
				OwnerID        *string     `json:"ownerid"`
				AccessLevel    string      `json:"accesslevel"`
				DisplayName    *string     `json:"displayName"`
			} `json:"resultSet"`
		} `json:"results"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	var allowlist []zomboidv1.AllowlistUser
	if len(response.Results) > 0 {
		for _, user := range response.Results[0].ResultSet {
			allowlist = append(allowlist, zomboidv1.AllowlistUser{
				ID:             user.ID,
				World:          user.World,
				Username:       user.Username,
				SteamID:        &user.SteamID,
				OwnerID:        user.OwnerID,
				AccessLevel:    user.AccessLevel,
				DisplayName:    user.DisplayName,
				Admin:          user.Admin == 1,
				Moderator:      user.Moderator == 1,
				Banned:         fmt.Sprint(user.Banned) == "true",
				Priority:       user.Priority,
				LastConnection: &user.LastConnection,
			})
		}
	}

	return allowlist, nil
}
