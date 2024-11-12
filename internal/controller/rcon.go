package controller

import (
	"context"
	"fmt"
	"time"

	"github.com/gorcon/rcon"
	zomboidv1 "github.com/hordehost/zomboid-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

func (r *ZomboidServerReconciler) connectRCON(ctx context.Context, zomboidServer *zomboidv1.ZomboidServer) (*rcon.Conn, func(), error) {
	password, err := r.getRCONPassword(ctx, zomboidServer)
	if err != nil {
		return nil, nil, err
	}

	hostname, port, cleanup, err := r.getServiceEndpoint(ctx,
		zomboidServer.Name+"-rcon",
		zomboidServer.Namespace,
		27015,
	)
	if err != nil {
		cleanup()
		return nil, nil, err
	}

	address := fmt.Sprintf("%s:%d", hostname, port)
	conn, err := rcon.Dial(
		address, password,
		rcon.SetDialTimeout(5*time.Second),
		rcon.SetDeadline(5*time.Second),
	)
	if err != nil {
		cleanup()
		return nil, nil, fmt.Errorf("failed to connect to RCON: %w", err)
	}

	return conn, func() {
		cleanup()
		conn.Close()
	}, nil
}

func (r *ZomboidServerReconciler) getRCONPassword(ctx context.Context, zomboidServer *zomboidv1.ZomboidServer) (string, error) {
	secret := &corev1.Secret{}
	if err := r.Get(ctx, types.NamespacedName{
		Name:      zomboidServer.Spec.Administrator.Password.LocalObjectReference.Name,
		Namespace: zomboidServer.Namespace,
	}, secret); err != nil {
		return "", fmt.Errorf("failed to get RCON secret: %w", err)
	}

	password := string(secret.Data[zomboidServer.Spec.Administrator.Password.Key])
	if password == "" {
		return "", fmt.Errorf(
			"RCON password not found in secret %s",
			zomboidServer.Spec.Administrator.Password.LocalObjectReference.Name,
		)
	}

	return password, nil
}
