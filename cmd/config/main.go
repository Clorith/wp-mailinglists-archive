package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type ListConfig struct {
	ID         string `json:"id"`
	ArchiveURL string `json:"archive_url"`
}

type Config struct {
	Lists []ListConfig `json:"lists"`
}

func main() {
	// Read the archive links file
	data, err := os.ReadFile("data/archive_links.txt")
	if err != nil {
		fmt.Printf("Error reading archive_links.txt: %v\n", err)
		os.Exit(1)
	}

	// Parse the file content
	lines := strings.Split(string(data), "\n")
	var config Config

	for _, line := range lines {
		// Skip comments and empty lines
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse the line
		parts := strings.Split(line, "|")
		if len(parts) != 2 {
			fmt.Printf("Warning: Invalid line format: %s\n", line)
			continue
		}

		listConfig := ListConfig{
			ID:         strings.TrimSpace(parts[0]),
			ArchiveURL: strings.TrimSpace(parts[1]),
		}

		config.Lists = append(config.Lists, listConfig)
	}

	// Create the output directory if it doesn't exist
	err = os.MkdirAll("data", 0755)
	if err != nil {
		fmt.Printf("Error creating data directory: %v\n", err)
		os.Exit(1)
	}

	// Write the config file
	configData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling config: %v\n", err)
		os.Exit(1)
	}

	err = os.WriteFile("data/config.json", configData, 0644)
	if err != nil {
		fmt.Printf("Error writing config.json: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully generated config.json with %d mailing lists\n", len(config.Lists))
}
