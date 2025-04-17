package extractor

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Extractor handles the extraction of gzipped files
type Extractor struct {
	inputDir  string
	outputDir string
}

// New creates a new Extractor instance
func New(inputDir, outputDir string) *Extractor {
	return &Extractor{
		inputDir:  inputDir,
		outputDir: outputDir,
	}
}

// ExtractFile extracts a single gzipped file
func (e *Extractor) ExtractFile(filename string) error {
	// Create output directory if it doesn't exist
	if err := os.MkdirAll(e.outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Open the gzipped file
	inputPath := filepath.Join(e.inputDir, filename)
	inputFile, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("failed to open input file: %w", err)
	}
	defer inputFile.Close()

	// Create gzip reader
	gzReader, err := gzip.NewReader(inputFile)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzReader.Close()

	// Create output file (remove .gz extension)
	outputFilename := strings.TrimSuffix(filename, ".gz")
	outputPath := filepath.Join(e.outputDir, outputFilename)
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outputFile.Close()

	// Copy the decompressed content
	if _, err := io.Copy(outputFile, gzReader); err != nil {
		return fmt.Errorf("failed to write decompressed content: %w", err)
	}

	return nil
}

// ExtractAll extracts all .gz files from the input directory
func (e *Extractor) ExtractAll() (int, int, error) {
	// Read all files in the input directory
	files, err := os.ReadDir(e.inputDir)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to read input directory: %w", err)
	}

	var successCount, failureCount int

	// Process each .gz file
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".gz") {
			fmt.Printf("Extracting %s...\n", file.Name())
			if err := e.ExtractFile(file.Name()); err != nil {
				fmt.Printf("Error extracting %s: %v\n", file.Name(), err)
				failureCount++
				continue
			}
			successCount++
		}
	}

	return successCount, failureCount, nil
}
