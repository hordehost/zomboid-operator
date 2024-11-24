package metrics

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gorcon/rcon"
	"github.com/hordehost/zomboid-operator/internal/players"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type Server struct {
	playerCount    prometheus.Gauge
	allowlistCount prometheus.Gauge
}

func NewServer() *Server {
	return &Server{
		playerCount: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "zomboid_connected_players",
			Help: "Number of players currently connected to the game server",
		}),
		allowlistCount: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "zomboid_allowlist_players",
			Help: "Number of players registered in the server allowlist",
		}),
	}
}

func (s *Server) Run(ctx context.Context) error {
	go s.collectMetrics(ctx)

	http.Handle("/metrics", promhttp.Handler())
	return http.ListenAndServe(":9090", nil)
}

func (s *Server) collectMetrics(ctx context.Context) {
	logger := log.FromContext(ctx)
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := s.updateMetrics(ctx); err != nil {
				logger.Error(err, "Failed to update metrics")
			}
		}
	}
}

func (s *Server) updateMetrics(ctx context.Context) error {
	rconConn, err := rcon.Dial(
		"localhost:27015",
		os.Getenv("RCON_PASSWORD"),
		rcon.SetDialTimeout(5*time.Second),
		rcon.SetDeadline(5*time.Second),
	)
	if err != nil {
		return fmt.Errorf("failed to connect to RCON: %w", err)
	}
	defer rconConn.Close()

	connected, err := players.GetConnectedPlayers(ctx, rconConn)
	if err != nil {
		return fmt.Errorf("failed to query connected players: %w", err)
	}
	s.playerCount.Set(float64(len(connected)))

	count, err := players.GetAllowlistCount("localhost", 12321, os.Getenv("ZOMBOID_SERVER_NAME"))
	if err != nil {
		return fmt.Errorf("failed to query allowlist count: %w", err)
	}
	s.allowlistCount.Set(float64(count))

	return nil
}
