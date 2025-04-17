package parser

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// Message represents a single email message from the archive
type Message struct {
	SenderName  string    `json:"sender_name"`
	SenderEmail string    `json:"sender_email"`
	Subject     string    `json:"subject"`
	Body        string    `json:"body"`
	Date        time.Time `json:"date"`
	MessageID   string    `json:"message_id"`
	InReplyTo   string    `json:"in_reply_to,omitempty"`
	References  []string  `json:"references,omitempty"`
	Replies     []Message `json:"replies,omitempty"`
}

// Parser handles the parsing of mailing list archive files
type Parser struct {
	inputDir string
	// Map of MessageID to Message for tracking threads
	messageMap map[string]*Message
}

// New creates a new Parser instance
func New(inputDir string) *Parser {
	return &Parser{
		inputDir:   inputDir,
		messageMap: make(map[string]*Message),
	}
}

// parseFromHeader extracts name and email from a From: header
func parseFromHeader(line string) (name, email string) {
	// Format is typically: "From: email at domain.com (Full Name)"
	// or sometimes just: "From: email at domain.com"

	// Extract the part in parentheses if it exists
	nameMatch := regexp.MustCompile(`\((.*?)\)`).FindStringSubmatch(line)
	if len(nameMatch) > 1 {
		name = nameMatch[1]
	}

	// Extract email part (everything between "From: " and " (" or end of line)
	emailPart := strings.Split(line, "From: ")[1]
	if idx := strings.Index(emailPart, " ("); idx != -1 {
		email = strings.TrimSpace(emailPart[:idx])
	} else {
		email = strings.TrimSpace(emailPart)
	}

	return name, email
}

// parseDate parses the date string from the archive format
func parseDate(dateStr string) (time.Time, error) {
	// Clean up the date string
	dateStr = strings.TrimSpace(dateStr)
	if dateStr == "" {
		return time.Time{}, fmt.Errorf("empty date string")
	}

	// Remove HTML tags
	dateStr = regexp.MustCompile(`<[^>]*>`).ReplaceAllString(dateStr, "")

	// Remove \r and \n
	dateStr = strings.TrimSuffix(dateStr, "\r")
	dateStr = strings.TrimSuffix(dateStr, "\n")

	// Remove time ranges (e.g., "from 10:00 AM to 4:00 PM (PDT)")
	if idx := strings.Index(dateStr, " from "); idx != -1 {
		dateStr = dateStr[:idx]
	}

	// Remove bracketed timezone information
	if idx := strings.Index(dateStr, " ["); idx != -1 {
		dateStr = dateStr[:idx]
	}

	// Normalize month names to title case
	months := []string{
		"january", "february", "march", "april", "may", "june",
		"july", "august", "september", "october", "november", "december",
		"jan", "feb", "mar", "apr", "jun", "jul", "aug", "sep", "oct", "nov", "dec",
	}
	for _, month := range months {
		dateStr = strings.ReplaceAll(dateStr, " "+month+" ", " "+strings.Title(month)+" ")
		dateStr = strings.ReplaceAll(dateStr, " "+strings.ToUpper(month)+" ", " "+strings.Title(month)+" ")
	}

	// Normalize timezones
	tzReplacements := map[string]string{
		"GMT+0000":  "GMT",
		"GMT+00:00": "GMT",
		"GMT+02:00": "+0200",
		"GMT+01:00": "+0100",
		"GMT-07:00": "-0700",
		"GMT-08:00": "-0800",
	}
	for old, new := range tzReplacements {
		dateStr = strings.ReplaceAll(dateStr, old, new)
	}

	// List of date formats to try
	formats := []string{
		"Mon Jan  2 15:04:05 2006",               // "Mon Aug  1 01:30:26 2006"
		"Mon, 2 Jan 2006 15:04:05 -0700",         // "Tue, 21 Aug 2012 11:58:49 +0200"
		"Mon, 2 Jan 2006 15:04:05 MST",           // "Wed, 20 Apr 2005 13:53:00 GMT"
		"on 02 January 2006 at 15:04:05 MST",     // "on 04 April 2005 at 01:22:54 UTC"
		"02/01/2006",                             // "12/04/2005"
		"Mon,  2 Jan 2006 15:04:05 -0700 (MST)",  // "Mon,  4 Jul 2005 22:03:43 +0000 (GMT)"
		"Jan 2, 2006 3:04 PM",                    // "Jul 21, 2006 11:19 PM"
		"January 02, 2006",                       // "June 09, 2006"
		"1/2/2006 3:04 PM",                       // "4/20/2007 3:25 PM"
		"Jan 2006",                               // "Mar 2004"
		"Mon, January 2, 2006 15:04",             // "Sat, November 11, 2006 3:18"
		"2006/1/2",                               // "2012/1/31"
		"Monday, 2 January, 2006, 15:04",         // "Monday, 2 April, 2012, 12:00"
		"Monday, January 2, 2006, 3:04 PM",       // "Monday, June 21, 2010, 2:03 PM"
		"Mon, Jan 2, 2006 at 3:04 PM",            // "Mon, Jul 1, 2013 at 1:00 AM"
		"2 Jan 2006 15:04:05 -0700",              // "30 Dec 2009 16:45:21 +0100"
		"Mon, 2 Jan 2006 15:04:05",               // "Thu, 30 Dec 2010 17:12:28"
		"Monday, 2 January, 2006 15:04",          // "Tuesday, 10 May, 2011 15:46"
		"Monday, January 2, 2006 15:04:05 MST",   // "Tuesday, July 17, 2012 12:11:00 AM GMT"
		"01/02/06 15:04",                         // "04/04/08 19:16"
		"2. January 2006",                        // "21. May 2007"
		"2 January 2006 15:04:05 -0700",          // "16 April 2007 21:26:32 +0200"
		"Monday, January 2, 2006 3:04:05 PM MST", // "Tuesday, July 17, 2012 12:11:00 AM GMT"
		"Mon, Jan 2, 2006 at 15:04",              // "Mon, Mar 26, 2012 at 10:36"
		"Monday, January 2, 2006 3:04 PM",        // "Thursday, April 11, 2013 2:08 PM"
		"Monday, January 2, 2006 at 3:04 PM",     // "Friday, May 16, 2014 at 3:39 PM"
		"2 January 2006 15:04:05 MST",            // "16 April 2007 21:26:32 GMT"
	}

	// Try each format
	for _, format := range formats {
		if date, err := time.Parse(format, dateStr); err == nil {
			return date, nil
		}
	}

	// Try to parse complex formats with AM/PM and timezone
	if strings.Contains(dateStr, " AM ") || strings.Contains(dateStr, " PM ") {
		// Extract date and time parts
		parts := strings.Fields(dateStr)
		if len(parts) >= 7 {
			// Find the AM/PM part
			var ampmIdx int
			for i, part := range parts {
				if part == "AM" || part == "PM" {
					ampmIdx = i
					break
				}
			}
			if ampmIdx > 0 && ampmIdx+1 < len(parts) {
				// Reconstruct the date string in a standard format
				dateTime := strings.Join(parts[:ampmIdx+1], " ")
				tz := parts[ampmIdx+1]
				if date, err := time.Parse("Monday, January 2, 2006 3:04:05 PM", dateTime); err == nil {
					// Apply timezone offset
					if tz == "GMT" || tz == "UTC" {
						return date.UTC(), nil
					}
				}
			}
		}
	}

	// Try to parse UTC format with "at"
	if strings.Contains(dateStr, "at") && strings.HasSuffix(dateStr, "UTC") {
		parts := strings.Split(dateStr, " at ")
		if len(parts) == 2 {
			// Try to parse "on 13 March 2005 at 23:24:22 UTC"
			dateWithoutPrefix := strings.TrimPrefix(parts[0], "on ")
			timeStr := strings.TrimSuffix(parts[1], " UTC")
			fullStr := dateWithoutPrefix + " " + timeStr
			if date, err := time.Parse("02 January 2006 15:04:05", fullStr); err == nil {
				return date, nil
			}
		}
	}

	// Try to parse month-first formats with flexible month names
	if strings.Contains(dateStr, ",") {
		parts := strings.Split(dateStr, ",")
		if len(parts) == 2 {
			monthDay := strings.Fields(strings.TrimSpace(parts[0]))
			if len(monthDay) == 2 {
				// Convert month name to number
				months := map[string]string{
					"Jan": "01", "Feb": "02", "Mar": "03", "Apr": "04",
					"May": "05", "Jun": "06", "Jul": "07", "Aug": "08",
					"Sep": "09", "Oct": "10", "Nov": "11", "Dec": "12",
					"January": "01", "February": "02", "March": "03", "April": "04",
					"June": "06", "July": "07", "August": "08",
					"September": "09", "October": "10", "November": "11", "December": "12",
				}
				if monthNum, ok := months[monthDay[0]]; ok {
					// Try to parse as "MM/DD/YYYY HH:MM PM"
					day := strings.TrimSuffix(monthDay[1], ",")
					yearTime := strings.TrimSpace(parts[1])
					fullStr := fmt.Sprintf("%s/%s/%s", monthNum, day, yearTime)
					if date, err := time.Parse("01/02/2006 3:04 PM", fullStr); err == nil {
						return date, nil
					}
				}
			}
		}
	}

	// Try to parse just month and year
	if len(strings.Fields(dateStr)) == 2 {
		if date, err := time.Parse("Jan 2006", dateStr); err == nil {
			// Set to first day of the month
			return date, nil
		}
	}

	// Try to parse European-style dates (e.g., "21. May 2007")
	if strings.Contains(dateStr, ".") {
		parts := strings.Fields(dateStr)
		if len(parts) == 3 {
			day := strings.TrimSuffix(parts[0], ".")
			month := parts[1]
			year := parts[2]
			months := map[string]string{
				"Jan": "01", "Feb": "02", "Mar": "03", "Apr": "04",
				"May": "05", "Jun": "06", "Jul": "07", "Aug": "08",
				"Sep": "09", "Oct": "10", "Nov": "11", "Dec": "12",
				"January": "01", "February": "02", "March": "03", "April": "04",
				"June": "06", "July": "07", "August": "08",
				"September": "09", "October": "10", "November": "11", "December": "12",
			}
			if monthNum, ok := months[month]; ok {
				fullStr := fmt.Sprintf("%s/%s/%s", monthNum, day, year)
				if date, err := time.Parse("01/02/2006", fullStr); err == nil {
					return date, nil
				}
			}
		}
	}

	// Try to parse short dates (e.g., "04/04/08 19:16")
	if strings.Contains(dateStr, "/") && len(dateStr) < 20 {
		if date, err := time.Parse("01/02/06 15:04", dateStr); err == nil {
			return date, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}

// parseReferences parses the References header into a slice of message IDs
func parseReferences(refs string) []string {
	if refs == "" {
		return nil
	}
	return strings.Fields(refs)
}

// addToThread adds a message to the appropriate thread
func (p *Parser) addToThread(msg *Message) {
	// If this is a reply, find the parent message
	if msg.InReplyTo != "" {
		if parent, exists := p.messageMap[msg.InReplyTo]; exists {
			parent.Replies = append(parent.Replies, *msg)
			return
		}
	}

	// If this is part of a thread via References, find the earliest ancestor
	if len(msg.References) > 0 {
		for _, ref := range msg.References {
			if parent, exists := p.messageMap[ref]; exists {
				parent.Replies = append(parent.Replies, *msg)
				return
			}
		}
	}

	// If no parent found, this is a new thread
	p.messageMap[msg.MessageID] = msg
}

// ParseFile parses a single archive file and returns all messages found
func (p *Parser) ParseFile(filename string) error {
	file, err := os.Open(filepath.Join(p.inputDir, filename))
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	var currentMessage *Message
	var messageBody strings.Builder
	var inMessageBody bool

	for scanner.Scan() {
		line := scanner.Text()

		// Check for start of new message
		if strings.HasPrefix(line, "From ") && !strings.HasPrefix(line, "From: ") {
			// If we were processing a message, save it
			if inMessageBody && currentMessage != nil {
				currentMessage.Body = strings.TrimSpace(messageBody.String())
				p.addToThread(currentMessage)
				messageBody.Reset()
			}

			// Start new message
			currentMessage = &Message{}
			inMessageBody = true
			continue
		}

		if !inMessageBody {
			continue
		}

		// Parse headers
		if strings.HasPrefix(line, "From: ") {
			currentMessage.SenderName, currentMessage.SenderEmail = parseFromHeader(line)
		} else if strings.HasPrefix(line, "Date: ") {
			dateStr := strings.TrimPrefix(line, "Date: ")
			if date, err := parseDate(dateStr); err == nil {
				currentMessage.Date = date
			} else {
				fmt.Printf("Warning: Failed to parse date '%s': %v\n", dateStr, err)
			}
		} else if strings.HasPrefix(line, "Subject: ") {
			currentMessage.Subject = strings.TrimPrefix(line, "Subject: ")
		} else if strings.HasPrefix(line, "Message-ID: ") {
			currentMessage.MessageID = strings.TrimPrefix(line, "Message-ID: ")
			currentMessage.MessageID = strings.Trim(currentMessage.MessageID, "<>")
		} else if strings.HasPrefix(line, "In-Reply-To: ") {
			currentMessage.InReplyTo = strings.TrimPrefix(line, "In-Reply-To: ")
			currentMessage.InReplyTo = strings.Trim(currentMessage.InReplyTo, "<>")
		} else if strings.HasPrefix(line, "References: ") {
			currentMessage.References = parseReferences(strings.TrimPrefix(line, "References: "))
		} else if line == "" {
			// Empty line marks the end of headers and start of body
			continue
		} else {
			// Collect message body
			messageBody.WriteString(line)
			messageBody.WriteString("\n")
		}
	}

	// Don't forget to save the last message
	if inMessageBody && currentMessage != nil {
		currentMessage.Body = strings.TrimSpace(messageBody.String())
		p.addToThread(currentMessage)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	return nil
}

// GetThreads returns all top-level messages (those that are not replies)
func (p *Parser) GetThreads() []Message {
	var threads []Message
	for _, msg := range p.messageMap {
		// Only include messages that are not replies
		if msg.InReplyTo == "" && len(msg.References) == 0 {
			threads = append(threads, *msg)
		}
	}
	return threads
}

// ParseAll processes all text files in the input directory
func (p *Parser) ParseAll() ([]Message, error) {
	files, err := os.ReadDir(p.inputDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read input directory: %w", err)
	}

	// Process all files to build the complete thread structure
	for _, file := range files {
		if !file.IsDir() && !strings.HasSuffix(file.Name(), ".gz") {
			fmt.Printf("Parsing %s...\n", file.Name())
			if err := p.ParseFile(file.Name()); err != nil {
				fmt.Printf("Error parsing %s: %v\n", file.Name(), err)
				continue
			}
		}
	}

	// Return only the top-level threads
	return p.GetThreads(), nil
}
