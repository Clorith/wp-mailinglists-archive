package scraper

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/net/html"
)

// Scraper represents a web scraper instance
type Scraper struct {
	client    *http.Client
	baseURL   string
	outputDir string
}

// New creates a new Scraper instance
func New(baseURL, outputDir string) *Scraper {
	// Create a custom transport with more permissive TLS settings
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
			MinVersion:         tls.VersionTLS10,
		},
	}

	// Create a client with the custom transport and longer timeout
	client := &http.Client{
		Transport: tr,
		Timeout:   30 * time.Second,
	}

	return &Scraper{
		client:    client,
		baseURL:   baseURL,
		outputDir: outputDir,
	}
}

// execCurl executes curl command with the given URL and returns the response body
func (s *Scraper) execCurl(url string) ([]byte, error) {
	cmd := exec.Command("curl", "-k", "-L", url)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("curl command failed: %w", err)
	}

	return out.Bytes(), nil
}

// findMailingListArchives finds all .txt.gz archive links in the mailing list page
func findMailingListArchives(n *html.Node) []string {
	var archives []string
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "tr" {
			var monthCell, linkCell *html.Node
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				if c.Type == html.ElementNode && c.Data == "td" {
					if monthCell == nil {
						monthCell = c
					} else {
						linkCell = c
					}
				}
			}

			// Look for .txt.gz links in the last cell
			if linkCell != nil {
				for a := linkCell.FirstChild; a != nil; a = a.NextSibling {
					if a.Type == html.ElementNode && a.Data == "a" {
						for _, attr := range a.Attr {
							if attr.Key == "href" && strings.HasSuffix(attr.Val, ".txt.gz") {
								archives = append(archives, attr.Val)
								break
							}
						}
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(n)
	return archives
}

// ScrapePage downloads the webpage and extracts .txt.gz archive links
func (s *Scraper) ScrapePage(pageURL string) ([]string, error) {
	body, err := s.execCurl(pageURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch page: %w", err)
	}

	doc, err := html.Parse(bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	archives := findMailingListArchives(doc)
	return archives, nil
}

// DownloadFile downloads a file from the given URL using curl
func (s *Scraper) DownloadFile(fileURL string) error {
	// Create absolute URL if relative
	if !strings.HasPrefix(fileURL, "http") {
		base, err := url.Parse(s.baseURL)
		if err != nil {
			return fmt.Errorf("failed to parse base URL: %w", err)
		}
		rel, err := url.Parse(fileURL)
		if err != nil {
			return fmt.Errorf("failed to parse relative URL: %w", err)
		}
		fileURL = base.ResolveReference(rel).String()
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(s.outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Get filename from URL
	filename := filepath.Base(fileURL)
	if filename == "" {
		filename = "downloaded_file"
	}
	outputPath := filepath.Join(s.outputDir, filename)

	// Download file using curl
	cmd := exec.Command("curl", "-k", "-L", "-o", outputPath, fileURL)
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}

	return nil
}
