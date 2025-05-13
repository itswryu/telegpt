// Package main provides a simple healthcheck utility for the TeleGPT application
package main

import (
	"fmt"
	"net/http"
	"os"
	"time"
)

func main() {
	// In a minimal scratch image, we can't rely on process checks
	// Instead, we'll try to connect to the Telegram API to see if our bot can connect

	// Create an HTTP client with a timeout
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// Try to reach the Telegram API (simple connectivity check)
	resp, err := client.Get("https://api.telegram.org")
	if err != nil {
		fmt.Println("Health check failed: couldn't connect to Telegram API:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	// Check if we got a successful response
	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		fmt.Printf("Health check failed: unexpected status code: %d\n", resp.StatusCode)
		os.Exit(1)
	}

	fmt.Println("TeleGPT health check passed")
	os.Exit(0)
}
