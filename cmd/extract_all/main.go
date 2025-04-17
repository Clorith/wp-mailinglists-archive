package main

import (
	"encoding/json"
	"fmt"
	"log"
	"maillist-scrub/pkg/extractor"
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

		// Set up input and output directories
		inputDir := filepath.Join("data", "lists", list.ID, "archives")
		outputDir := filepath.Join("data", "lists", list.ID, "extracted")

		// Create extractor instance
		ext := extractor.New(inputDir, outputDir)

		// Extract all files
		successCount, failureCount, err := ext.ExtractAll()
		if err != nil {
			log.Printf("Failed to extract files for %s: %v", list.ID, err)
			continue
		}

		// Print summary for this list
		fmt.Printf("\nExtraction complete for %s!\n", list.ID)
		fmt.Printf("Successfully extracted: %d files\n", successCount)
		if failureCount > 0 {
			fmt.Printf("Failed to extract: %d files\n", failureCount)
		}
	}

	fmt.Printf("\nAll mailing lists processed!\n")
}
