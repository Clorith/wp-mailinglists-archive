package main

import (
	"encoding/json"
	"fmt"
	"log"
	"maillist-scrub/pkg/scraper"
	"os"
	"path/filepath"
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

		// Create output directory for this list
		outputDir := filepath.Join("data", "lists", list.ID, "archives")
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			log.Printf("Failed to create output directory for %s: %v", list.ID, err)
			continue
		}

		// Create a new scraper instance
		s := scraper.New(list.ArchiveURL, outputDir)

		// Scrape the page for archive links
		fmt.Printf("Scanning %s for mailing list archives...\n", list.ArchiveURL)
		archives, err := s.ScrapePage(list.ArchiveURL)
		if err != nil {
			log.Printf("Failed to scrape page for %s: %v", list.ID, err)
			continue
		}

		fmt.Printf("Found %d .txt.gz archive files for %s\n", len(archives), list.ID)
		if len(archives) == 0 {
			fmt.Printf("No mailing list archives found for %s. Make sure the URL points to a valid mailing list archive page.\n", list.ID)
			continue
		}

		// Download each archive
		var downloadCount, errorCount int
		for i, archive := range archives {
			filename := filepath.Base(archive)
			month := strings.TrimSuffix(filename, ".txt.gz")
			fmt.Printf("Downloading archive %d/%d: %s\n", i+1, len(archives), month)

			if err := s.DownloadFile(archive); err != nil {
				log.Printf("Failed to download %s: %v", month, err)
				errorCount++
				continue
			}
			downloadCount++
			fmt.Printf("Successfully downloaded: %s\n", month)
		}

		fmt.Printf("\nDownload complete for %s!\n", list.ID)
		fmt.Printf("Successfully downloaded: %d archives\n", downloadCount)
		if errorCount > 0 {
			fmt.Printf("Failed to download: %d archives\n", errorCount)
		}
	}

	fmt.Printf("\nAll mailing lists processed!\n")
}
