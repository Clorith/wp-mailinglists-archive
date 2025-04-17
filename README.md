# Mailing List Archive

A web application for browsing and searching WordPress mailing list archives from pipermailer archives.

<div style="background-color: #fff3cd; color: #856404; padding: 10px; border-radius: 4px; margin: 15px 0; border-left: 4px solid #ffeeba;">
  <strong>Note:</strong> This is a Vibe Code experiment.
</div>

This application contains the archives from the Automattic mailing lists for the WordPress project by default, including wp-hackers, wp-docs, and other WordPress-related mailing lists. You can find the original mailing lists at [lists.automattic.com](https://lists.automattic.com/mailman/listinfo).

## Features

- **Web Interface**: Browse and search mailing list archives through a modern web interface
- **Multiple Mailing Lists**: Support for multiple WordPress mailing lists (wp-hackers, wp-docs, etc.)
- **Thread View**: View complete email threads with all replies
- **Search Functionality**: Search across all mailing lists or filter by specific lists
- **Pagination**: Navigate through threads with intuitive pagination controls
- **Accessibility**: Accessible UI with keyboard navigation and screen reader support
- **Responsive Design**: Works well on desktop and mobile devices
- **Archive Scraping**: Scrapes pipermailer archives for `.txt.gz` files in table rows
- **Archive Extraction**: Extracts compressed archives for processing
- **Email Parsing**: Parses email messages into structured data

## Directory Structure

```
.
├── data/
│   ├── lists/
│   │   ├── wp-hackers/
│   │   │   ├── messages.json
│   │   │   ├── extracted/
│   │   │   └── archives/
│   │   ├── wp-docs/
│   │   │   ├── messages.json
│   │   │   ├── extracted/
│   │   │   └── archives/
│   │   └── ...
│   └── config.json
├── pkg/
│   ├── scraper/     # Web scraping functionality
│   ├── extractor/   # Archive extraction functionality
│   ├── parser/      # Email parsing functionality
│   └── web/         # Web server and templates
├── templates/       # HTML templates
└── cmd/
    ├── scraper/     # Command to scrape mailing list archives
    ├── extract/     # Command to extract .gz archives
    ├── extract_all/ # Command to extract all mailing list archives
    ├── parse/       # Command to parse extracted emails
    ├── parse_all/   # Command to parse all extracted emails
    └── web/         # Command to run the web server
```

## Installation

1. Make sure you have Go installed on your system
2. Clone this repository
3. Install dependencies:
   ```bash
   go mod tidy
   ```

## Usage

### Web Server

Run the web server with:

```bash
go run cmd/web/main.go
```

The server will start on port 8080 by default. You can access the web interface at http://localhost:8080.

### Command Line Tools

#### Scrape Mailing List Archives

```bash
go run cmd/scraper/main.go -list "wp-hackers" -output "data/lists/wp-hackers/archives"
```

#### Extract Archives

```bash
go run cmd/extract/main.go -input "data/lists/wp-hackers/archives" -output "data/lists/wp-hackers/extracted"
```

#### Extract All Archives

```bash
go run cmd/extract_all/main.go
```

#### Parse Extracted Emails

```bash
go run cmd/parse/main.go -input "data/lists/wp-hackers/extracted" -output "data/lists/wp-hackers/messages.json"
```

#### Parse All Extracted Emails

```bash
go run cmd/parse_all/main.go
```

## Configuration

The application uses a configuration file at `data/config.json` to define the mailing lists:

```json
{
  "lists": [
    {
      "id": "wp-hackers",
      "name": "WordPress Hackers",
      "description": "WordPress development mailing list",
      "url": "https://wordpress.org/mailman/listinfo/wp-hackers"
    },
    {
      "id": "wp-docs",
      "name": "WordPress Documentation",
      "description": "WordPress documentation mailing list",
      "url": "https://wordpress.org/mailman/listinfo/wp-docs"
    }
  ]
}
```

## Web Interface

The web interface provides the following features:

- **Home Page**: View all threads from selected mailing lists
- **List View**: View threads from a specific mailing list
- **Thread View**: View a complete email thread with all replies
- **Search**: Search across all mailing lists or filter by specific lists
- **Pagination**: Navigate through threads with intuitive pagination controls
- **Accessibility**: Accessible UI with keyboard navigation and screen reader support

## Development

### Adding a New Mailing List

1. Add the mailing list to `data/config.json`
2. Create the directory structure:
   ```
   data/lists/<list-id>/
   ├── archives/
   ├── extracted/
   └── messages.json
   ```
3. Run the scraper to download archives:
   ```bash
   go run cmd/scraper/main.go -list "<list-id>" -output "data/lists/<list-id>/archives"
   ```
4. Extract the archives:
   ```bash
   go run cmd/extract/main.go -input "data/lists/<list-id>/archives" -output "data/lists/<list-id>/extracted"
   ```
5. Parse the extracted emails:
   ```bash
   go run cmd/parse/main.go -input "data/lists/<list-id>/extracted" -output "data/lists/<list-id>/messages.json"
   ```

## License

This project is licensed under the MIT License - see the LICENSE file for details. 