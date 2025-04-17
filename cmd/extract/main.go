package main

import (
	"flag"
	"fmt"
	"log"
	"path/filepath"

	"maillist-scrub/pkg/extractor"
)

func main() {
	// Parse command line flags
	inputDir := flag.String("input", "wp-hackers-archives", "Directory containing gzipped files")
	outputDir := flag.String("output", "wp-hackers-archives-extracted", "Directory for extracted files")
	flag.Parse()

	// Convert relative paths to absolute paths
	absInputDir, err := filepath.Abs(*inputDir)
	if err != nil {
		log.Fatalf("Failed to get absolute path for input directory: %v", err)
	}

	absOutputDir, err := filepath.Abs(*outputDir)
	if err != nil {
		log.Fatalf("Failed to get absolute path for output directory: %v", err)
	}

	// Create extractor instance
	ext := extractor.New(absInputDir, absOutputDir)

	// Extract all files
	successCount, failureCount, err := ext.ExtractAll()
	if err != nil {
		log.Fatalf("Failed to extract files: %v", err)
	}

	// Print summary
	fmt.Printf("\nExtraction complete!\n")
	fmt.Printf("Successfully extracted: %d files\n", successCount)
	if failureCount > 0 {
		fmt.Printf("Failed to extract: %d files\n", failureCount)
	}
}
