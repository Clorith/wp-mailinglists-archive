# Web Scraper and File Downloader

A Go application that scrapes a webpage for `.txt.gz` files in table rows and downloads them.

## Features

- Scrapes a webpage for `.txt.gz` files specifically in table rows
- Looks for links in the last cell of each table row
- Downloads only `.txt.gz` files
- Handles both absolute and relative URLs
- Creates output directory if it doesn't exist
- Provides progress feedback during downloads

## Installation

1. Make sure you have Go installed on your system
2. Clone this repository
3. Install dependencies:
   ```bash
   go mod tidy
   ```

## Usage

Run the application with the following command:

```bash
go run cmd/scraper/main.go -url "https://example.com" -output "downloads"
```

### Command Line Flags

- `-url`: The URL of the webpage to scrape (required)
- `-output`: Directory where downloaded files will be saved (default: "downloads")

## Example

```bash
# Scrape example.com for .txt.gz files in table rows and save them to the downloads directory
go run cmd/scraper/main.go -url "https://example.com" -output "downloads"
```

## Expected HTML Structure

The application expects the webpage to have a table structure where:
- Each row represents a month and year
- The last cell in each row contains a link to a `.txt.gz` file

Example HTML structure:
```html
<table>
  <tr>
    <td>January 2023</td>
    <td>Some data</td>
    <td><a href="file_2023_01.txt.gz">Download</a></td>
  </tr>
  <tr>
    <td>February 2023</td>
    <td>Some data</td>
    <td><a href="file_2023_02.txt.gz">Download</a></td>
  </tr>
</table>
```

## Error Handling

The application will:
- Log errors for failed downloads but continue with remaining files
- Create the output directory if it doesn't exist
- Handle both absolute and relative URLs
- Provide feedback on the progress of downloads
- Check if any `.txt.gz` files were found before attempting downloads

# Mailing List Archive

A web application for browsing mailing list archives.

## Directory Structure

```
.
├── data/
│   ├── lists/
│   │   ├── wp-hackers/
│   │   │   ├── messages.json
│   │   │   └── archives/
│   │   │       ├── 2006/
│   │   │       ├── 2007/
│   │   │       └── ...
│   │   ├── wp-docs/
│   │   │   ├── messages.json
│   │   │   └── archives/
│   │   └── ...
│   └── config.json
├── pkg/
│   ├── scraper/
│   ├── extractor/
│   └── web/
├── templates/
└── cmd/
    └── web/
```

## Data Format

### config.json
```json
{
  "lists": [
    {
      "id": "wp-hackers",
      "name": "WordPress Hackers",
      "description": "Discussion about WordPress development",
      "archive_pattern": "https://lists.wordpress.org/pipermail/wp-hackers/%Y-%B.txt.gz"
    },
    {
      "id": "wp-docs",
      "name": "WordPress Documentation",
      "description": "Discussion about WordPress documentation",
      "archive_pattern": "https://lists.wordpress.org/pipermail/wp-docs/%Y-%B.txt.gz"
    }
  ]
}
```

### messages.json
Each mailing list has its own messages.json file with the following structure:
```json
{
  "list_id": "wp-hackers",
  "messages": [
    {
      "sender_name": "John Doe",
      "sender_email": "john@example.com",
      "subject": "Thread Subject",
      "body": "Message body...",
      "date": "2023-04-17T10:26:24Z",
      "message_id": "unique-message-id",
      "in_reply_to": "parent-message-id",
      "references": ["ref1", "ref2"],
      "replies": []
    }
  ]
}
```

## Usage

1. Add new mailing list configurations to `data/config.json`
2. Run the scraper for each list: `go run cmd/scraper/main.go -list wp-hackers`
3. Start the web server: `go run cmd/web/main.go` 