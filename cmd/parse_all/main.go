package main

import (
	"encoding/json"
	"fmt"
	"log"
	"maillist-scrub/pkg/parser"
	"os"
	"path/filepath"
)

type ListConfig struct {
	ID         string `json:"id"`
	ArchiveURL string `json:"archive_url"`
}

type Config struct {
	Lists []ListConfig `json:"lists"`
}

func main() {
	// Read the config file
	configData, err := os.ReadFile("data/config.json")
	if err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}

	var config Config
	if err := json.Unmarshal(configData, &config); err != nil {
		log.Fatalf("Failed to parse config: %v", err)
	}

	fmt.Printf("Found %d mailing lists in config\n", len(config.Lists))

	// Process each mailing list
	for _, list := range config.Lists {
		fmt.Printf("\nProcessing mailing list: %s\n", list.ID)

		// Set up input directory
		inputDir := filepath.Join("data", "lists", list.ID, "extracted")
		outputFile := filepath.Join("data", "lists", list.ID, "messages.json")

		// Create parser instance
		p := parser.New(inputDir)

		// Parse all files
		messages, err := p.ParseAll()
		if err != nil {
			log.Printf("Failed to parse files for %s: %v", list.ID, err)
			continue
		}

		// Convert messages to JSON
		jsonData, err := json.MarshalIndent(messages, "", "  ")
		if err != nil {
			log.Printf("Failed to convert messages to JSON for %s: %v", list.ID, err)
			continue
		}

		// Write to output file
		if err := os.WriteFile(outputFile, jsonData, 0644); err != nil {
			log.Printf("Failed to write output file for %s: %v", list.ID, err)
			continue
		}

		// Print summary for this list
		fmt.Printf("\nParsing complete for %s!\n", list.ID)
		fmt.Printf("Total messages parsed: %d\n", len(messages))
		fmt.Printf("Output written to: %s\n", outputFile)
	}

	fmt.Printf("\nAll mailing lists processed!\n")
}
