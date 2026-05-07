package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	configPath := os.Getenv("COLLECTOR_CONFIG")
	if configPath == "" {
		configPath = "/etc/otel/collector_config.yaml"
	}

	// Start OpenTelemetry Collector as child process
	cmd := exec.Command("otelcol-contrib", "--config", configPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start collector: %v\n", err)
		os.Exit(1)
	}

	// Health-check endpoint
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		if cmd.Process == nil || cmd.Process.Signal(syscall.Signal(0)) != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		// Check if collector is responding
		resp, err := http.Get("http://127.0.0.1:13133/healthz")
		if err != nil || resp.StatusCode != http.StatusOK {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	go func() {
		if err := http.ListenAndServe(":8082", nil); err != nil {
			fmt.Fprintf(os.Stderr, "Health server error: %v\n", err)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	fmt.Println("Observability sidecar started. Collector PID:", cmd.Process.Pid)
	fmt.Println("Health endpoint: http://localhost:8082/healthz")

	// Graceful shutdown on SIGINT/SIGTERM
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	fmt.Println("Shutting down observability sidecar...")
	if cmd.Process != nil {
		_ = cmd.Process.Signal(syscall.SIGTERM)
		// Wait for collector to exit with timeout
		done := make(chan error, 1)
		go func() { done <- cmd.Wait() }()
		select {
		case err := <-done:
			if err != nil {
				fmt.Fprintf(os.Stderr, "Collector exited with error: %v\n", err)
			}
		case <-ctx.Done():
			fmt.Fprintln(os.Stderr, "Collector did not exit in time, killing")
			_ = cmd.Process.Kill()
		}
	}
	os.Exit(0)
}
