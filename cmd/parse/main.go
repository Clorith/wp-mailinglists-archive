package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"maillist-scrub/pkg/parser"
)

func main() {
	// Parse command line flags
	inputDir := flag.String("input", "wp-hackers-archives-extracted", "Directory containing extracted archive files")
	outputFile := flag.String("output", "messages.json", "Output JSON file for parsed messages")
	flag.Parse()

	// Convert relative paths to absolute paths
	absInputDir, err := filepath.Abs(*inputDir)
	if err != nil {
		log.Fatalf("Failed to get absolute path for input directory: %v", err)
	}

	// Create parser instance
	p := parser.New(absInputDir)

	// Parse all files
	messages, err := p.ParseAll()
	if err != nil {
		log.Fatalf("Failed to parse files: %v", err)
	}

	// Convert messages to JSON
	jsonData, err := json.MarshalIndent(messages, "", "  ")
	if err != nil {
		log.Fatalf("Failed to convert messages to JSON: %v", err)
	}

	// Write to output file
	if err := os.WriteFile(*outputFile, jsonData, 0644); err != nil {
		log.Fatalf("Failed to write output file: %v", err)
	}

	// Print summary
	fmt.Printf("\nParsing complete!\n")
	fmt.Printf("Total messages parsed: %d\n", len(messages))
	fmt.Printf("Output written to: %s\n", *outputFile)
}
