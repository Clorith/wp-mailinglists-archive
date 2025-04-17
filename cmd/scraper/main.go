package main

import (
	"flag"
	"fmt"
	"log"
	"maillist-scrub/pkg/scraper"
	"path/filepath"
	"strings"
)

func main() {
	// Parse command line flags
	url := flag.String("url", "", "URL of the mailing list archive to scrape")
	outputDir := flag.String("output", "downloads", "Directory to save downloaded files")
	flag.Parse()

	if *url == "" {
		log.Fatal("Please provide a URL using the -url flag")
	}

	// Create a new scraper instance
	s := scraper.New(*url, *outputDir)

	// Scrape the page for archive links
	fmt.Printf("Scanning %s for mailing list archives...\n", *url)
	archives, err := s.ScrapePage(*url)
	if err != nil {
		log.Fatalf("Failed to scrape page: %v", err)
	}

	fmt.Printf("Found %d .txt.gz archive files\n", len(archives))
	if len(archives) == 0 {
		fmt.Println("No mailing list archives found. Make sure the URL points to a valid mailing list archive page.")
		return
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

	fmt.Printf("\nDownload complete!\n")
	fmt.Printf("Successfully downloaded: %d archives\n", downloadCount)
	if errorCount > 0 {
		fmt.Printf("Failed to download: %d archives\n", errorCount)
	}
}
