package players

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	zomboidv1 "github.com/hordehost/zomboid-operator/api/v1"
	"golang.org/x/crypto/bcrypt"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type sqliteRequest struct {
	Transaction []sqliteStatement `json:"transaction"`
}

type sqliteStatement struct {
	Query     string                 `json:"query,omitempty"`
	Statement string                 `json:"statement,omitempty"`
	Values    map[string]interface{} `json:"values,omitempty"`
}

// GetAllowlist queries the ws4sqlite service to retrieve the current allowlist
func GetAllowlist(hostname string, port int, serverName string) ([]zomboidv1.AllowlistUser, error) {
	sqliteUrl := fmt.Sprintf("http://%s:%d/%s", hostname, port, serverName)

	request := sqliteRequest{
		Transaction: []sqliteStatement{
			{
				Query: "SELECT * FROM whitelist ORDER BY id ASC",
			},
		},
	}

	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := http.Post(sqliteUrl, "application/json", bytes.NewReader(requestBody))
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
				Username       string      `json:"username"`
				SteamID        string      `json:"steamid"`
				OwnerID        *string     `json:"ownerid"`
				AccessLevel    string      `json:"accesslevel"`
				DisplayName    *string     `json:"displayName"`
				Banned         interface{} `json:"banned"`
				Password       string      `json:"password"`
				LastConnection string      `json:"lastConnection"`
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
				Username:       user.Username,
				SteamID:        &user.SteamID,
				OwnerID:        user.OwnerID,
				AccessLevel:    user.AccessLevel,
				DisplayName:    user.DisplayName,
				Banned:         fmt.Sprint(user.Banned) == "true",
				LastConnection: &user.LastConnection,
			})
		}
	}

	return allowlist, nil
}

func SetPassword(ctx context.Context, hostname string, port int, serverName string, username string, password string) error {
	sqliteUrl := fmt.Sprintf("http://%s:%d/%s", hostname, port, serverName)

	bcryptHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	request := sqliteRequest{
		Transaction: []sqliteStatement{
			{
				Statement: "UPDATE whitelist SET password = :password WHERE username = :username",
				Values: map[string]interface{}{
					"password": string(bcryptHash),
					"username": username,
				},
			},
		},
	}

	requestBody, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	logger := log.FromContext(ctx)
	logger.Info("setting password for %s to %s", username, bcryptHash)

	resp, err := http.Post(sqliteUrl, "application/json", bytes.NewReader(requestBody))
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("received non-2xx response: %d, failed to read body: %w", resp.StatusCode, err)
		}
		return fmt.Errorf("received non-2xx response: %d, body: %s", resp.StatusCode, string(body))
	}

	return nil
}

// GetAllowlistCount queries the ws4sqlite service to get the count of allowlisted players
func GetAllowlistCount(hostname string, port int, serverName string) (int, error) {
	sqliteUrl := fmt.Sprintf("http://%s:%d/%s", hostname, port, serverName)

	request := sqliteRequest{
		Transaction: []sqliteStatement{
			{
				Query: "SELECT COUNT(*) FROM whitelist",
			},
		},
	}

	requestBody, err := json.Marshal(request)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := http.Post(sqliteUrl, "application/json", bytes.NewReader(requestBody))
	if err != nil {
		return 0, fmt.Errorf("failed to make HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return 0, fmt.Errorf("received non-2xx response: %d, failed to read body: %w", resp.StatusCode, err)
		}
		return 0, fmt.Errorf("received non-2xx response: %d, body: %s", resp.StatusCode, string(body))
	}

	var response struct {
		Results []struct {
			ResultSet []struct {
				Count int `json:"COUNT(*)"`
			} `json:"resultSet"`
		} `json:"results"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return 0, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(response.Results) == 0 || len(response.Results[0].ResultSet) == 0 {
		return 0, fmt.Errorf("invalid response format")
	}

	return response.Results[0].ResultSet[0].Count, nil
}
