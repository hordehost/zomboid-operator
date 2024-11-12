package controller

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"golang.org/x/exp/rand"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func isRunningInCluster() bool {
	// If running in a pod, this env var will be set
	if os.Getenv("KUBERNETES_SERVICE_HOST") != "" {
		return true
	}

	// Alternatively, check if the serviceaccount token exists
	if _, err := os.Stat("/var/run/secrets/kubernetes.io/serviceaccount/token"); err == nil {
		return true
	}

	return false
}

func (r *ZomboidServerReconciler) getServiceEndpoint(ctx context.Context, name, namespace string, port int) (string, int, func(), error) {
	hostname := fmt.Sprintf("%s.%s.svc.cluster.local", name, namespace)
	var cleanup func()

	if !isRunningInCluster() {
		parts := strings.Split(hostname, ".")
		localPort, cleanupFn, err := SetupPortForwarder(ctx, r.Config, r.Client, parts[1], parts[0], port)
		if err != nil {
			return "", 0, cleanup, fmt.Errorf("failed to setup port forwarder: %w", err)
		}
		cleanup = cleanupFn
		hostname = "localhost"
		port = localPort
	}

	return hostname, port, cleanup, nil
}

func SetupPortForwarder(ctx context.Context, config *rest.Config, k8sClient client.Client, namespace string, serviceName string, targetPort int) (int, func(), error) {
	roundTripper, upgrader, err := spdy.RoundTripperFor(config)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to create round tripper: %w", err)
	}

	pod, err := getRandomPodFromService(ctx, k8sClient, namespace, serviceName)
	if err != nil {
		return 0, nil, err
	}

	path := fmt.Sprintf("/api/v1/namespaces/%s/pods/%s/portforward", namespace, pod.Name)
	hostIP := strings.TrimLeft(config.Host, "htps:/")
	serverURL := url.URL{Scheme: "https", Path: path, Host: hostIP}

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: roundTripper}, http.MethodPost, &serverURL)

	stopChan := make(chan struct{}, 1)
	readyChan := make(chan struct{})

	fw, err := portforward.NewOnAddresses(
		dialer,
		[]string{"localhost"}, []string{fmt.Sprintf("%d:%d", 0, targetPort)},
		stopChan, readyChan,
		nil, nil,
	)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to create port forwarder: %w", err)
	}

	go func() {
		err := fw.ForwardPorts()
		if err != nil {
			log.Log.Error(err, "Error forwarding ports")
		}
	}()

	select {
	case <-readyChan:
	case <-ctx.Done():
		return 0, nil, ctx.Err()
	}

	ports, err := fw.GetPorts()
	if err != nil {
		return 0, nil, fmt.Errorf("failed to get forwarded ports: %w", err)
	}

	cleanup := func() {
		close(stopChan)
	}

	return int(ports[0].Local), cleanup, nil
}

func getRandomPodFromService(ctx context.Context, k8sClient client.Client, namespace string, serviceName string) (*corev1.Pod, error) {
	service := &corev1.Service{}
	if err := k8sClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: serviceName}, service); err != nil {
		return nil, fmt.Errorf("failed to get service: %w", err)
	}

	pods := &corev1.PodList{}
	if err := k8sClient.List(ctx, pods, &client.ListOptions{
		Namespace:     namespace,
		LabelSelector: labels.SelectorFromSet(service.Spec.Selector),
	}); err != nil {
		return nil, fmt.Errorf("failed to list pods: %w", err)
	}

	if len(pods.Items) == 0 {
		return nil, fmt.Errorf("no pods found for service %s", service.Name)
	}

	return &pods.Items[rand.Intn(len(pods.Items))], nil
}
