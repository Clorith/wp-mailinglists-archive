package web

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

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

type Thread struct {
	FirstMessage    Message
	Replies         []Message
	TotalMessages   int
	LastMessageDate time.Time
}

type ListConfig struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	ArchivePattern string `json:"archive_pattern"`
}

type Config struct {
	Lists []ListConfig `json:"lists"`
}

type MailingList struct {
	Config  ListConfig
	Threads map[string]*Thread
}

type Server struct {
	lists     map[string]*MailingList // key is list_id
	templates *template.Template
	pageSize  int
}

type PageData struct {
	Title       string
	Threads     []*Thread
	CurrentPage int
	TotalPages  int
	HasPrev     bool
	HasNext     bool
	Lists       []ListConfig
	ActiveList  string
}

type IndexData struct {
	Title         string
	Threads       []*Thread
	CurrentPage   int
	TotalPages    int
	HasPrev       bool
	HasNext       bool
	SearchQuery   string
	Lists         []ListConfig
	SelectedLists []string
	ActiveList    string
}

type ThreadData struct {
	Title  string
	Thread *Thread
	List   *ListConfig
}

func New(configFile string) (*Server, error) {
	// Read config file
	configData, err := os.ReadFile("data/config.json")
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(configData, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Initialize server
	s := &Server{
		lists:    make(map[string]*MailingList),
		pageSize: 20,
	}

	// Load each mailing list
	for _, listConfig := range config.Lists {
		// Read messages file for this list
		messagesFile := fmt.Sprintf("data/lists/%s/messages.json", listConfig.ID)
		data, err := os.ReadFile(messagesFile)
		if err != nil {
			log.Printf("Warning: failed to read messages for list %s: %v", listConfig.ID, err)
			continue
		}

		var messages []Message
		if err := json.Unmarshal(data, &messages); err != nil {
			log.Printf("Warning: failed to parse messages for list %s: %v", listConfig.ID, err)
			continue
		}

		// Create threads map for this list
		threads := make(map[string]*Thread)
		for _, msg := range messages {
			if msg.InReplyTo == "" {
				totalMessages := 1 + len(msg.Replies)
				lastMessageDate := msg.Date
				for _, reply := range msg.Replies {
					if reply.Date.After(lastMessageDate) {
						lastMessageDate = reply.Date
					}
				}

				thread := &Thread{
					FirstMessage:    msg,
					Replies:         msg.Replies,
					TotalMessages:   totalMessages,
					LastMessageDate: lastMessageDate,
				}
				threads[msg.MessageID] = thread
			}
		}

		// Add list to server
		s.lists[listConfig.ID] = &MailingList{
			Config:  listConfig,
			Threads: threads,
		}
		log.Printf("Loaded mailing list %s with %d threads", listConfig.ID, len(threads))
	}

	// Create template functions
	funcMap := template.FuncMap{
		"add": func(a, b int) int {
			return a + b
		},
		"subtract": func(a, b int) int {
			return a - b
		},
		"sequence": func(start, end int) []int {
			var result []int
			for i := start; i <= end; i++ {
				result = append(result, i)
			}
			return result
		},
		"splitLines": func(s string) []string {
			return strings.Split(s, "\n")
		},
		"hasPrefix": func(s, prefix string) bool {
			return strings.HasPrefix(s, prefix)
		},
		"md5": func(s string) string {
			hash := md5.Sum([]byte(strings.ToLower(strings.TrimSpace(s))))
			return hex.EncodeToString(hash[:])
		},
		"urlEscape": func(s string) string {
			return url.QueryEscape(s)
		},
		"isListSelected": func(listID string, selectedLists []string) bool {
			for _, id := range selectedLists {
				if id == listID {
					return true
				}
			}
			return false
		},
	}

	// Load templates
	templates, err := template.New("").Funcs(funcMap).ParseFiles("templates/index.html", "templates/thread.html")
	if err != nil {
		return nil, fmt.Errorf("failed to parse templates: %w", err)
	}
	s.templates = templates

	return s, nil
}

func (s *Server) Start(addr string) error {
	// Add middleware to discourage crawling
	http.HandleFunc("/", s.withAntiCrawlerHeaders(s.handleIndex))
	http.HandleFunc("/thread/", s.withAntiCrawlerHeaders(s.handleThread))
	http.HandleFunc("/robots.txt", s.handleRobotsTxt)

	log.Printf("Starting server on %s...\n", addr)
	return http.ListenAndServe(addr, nil)
}

// withAntiCrawlerHeaders adds HTTP headers to discourage crawling
func (s *Server) withAntiCrawlerHeaders(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set headers to discourage crawling
		w.Header().Set("X-Robots-Tag", "noindex, nofollow, noarchive, nosnippet, noimageindex")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Permissions-Policy", "interest-cohort=()")

		// Call the next handler
		next(w, r)
	}
}

// handleRobotsTxt serves the robots.txt file
func (s *Server) handleRobotsTxt(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("User-agent: *\nDisallow: /\n"))
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	page := 1
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	searchQuery := r.URL.Query().Get("q")
	searchQuery = strings.TrimSpace(searchQuery)

	// Get selected lists from query parameters
	selectedLists := r.URL.Query()["lists[]"]
	hasSelectedLists := len(selectedLists) > 0

	// If only one list is selected, use it as the active list
	activeList := ""
	if len(selectedLists) == 1 {
		activeList = selectedLists[0]
	}

	// Get all threads from selected lists or all lists
	var threads []*Thread
	if hasSelectedLists {
		for _, listID := range selectedLists {
			if list, exists := s.lists[listID]; exists {
				for _, thread := range list.Threads {
					threads = append(threads, thread)
				}
			}
		}
	} else {
		for _, mailingList := range s.lists {
			for _, thread := range mailingList.Threads {
				threads = append(threads, thread)
			}
		}
	}

	// Filter threads based on search query if provided
	if searchQuery != "" {
		searchQuery = strings.ToLower(searchQuery)
		filteredThreads := make([]*Thread, 0)
		for _, thread := range threads {
			// Search in title and first message
			if strings.Contains(strings.ToLower(thread.FirstMessage.Subject), searchQuery) ||
				strings.Contains(strings.ToLower(thread.FirstMessage.Body), searchQuery) {
				filteredThreads = append(filteredThreads, thread)
				continue
			}

			// Search in replies
			for _, reply := range thread.Replies {
				if strings.Contains(strings.ToLower(reply.Body), searchQuery) {
					filteredThreads = append(filteredThreads, thread)
					break
				}
			}
		}
		threads = filteredThreads
	}

	// Sort threads by date (newest first)
	sort.Slice(threads, func(i, j int) bool {
		return threads[i].FirstMessage.Date.After(threads[j].FirstMessage.Date)
	})

	// Calculate pagination
	threadsPerPage := s.pageSize
	totalThreads := len(threads)
	totalPages := (totalThreads + threadsPerPage - 1) / threadsPerPage

	if page > totalPages {
		page = totalPages
	}

	start := (page - 1) * threadsPerPage
	end := start + threadsPerPage
	if end > totalThreads {
		end = totalThreads
	}

	// Get list of all mailing lists for the filters
	var lists []ListConfig
	for _, list := range s.lists {
		lists = append(lists, list.Config)
	}
	// Sort lists alphabetically by ID
	sort.Slice(lists, func(i, j int) bool {
		return lists[i].ID < lists[j].ID
	})

	data := IndexData{
		Title:         "WordPress Mailing List Archive",
		Threads:       threads[start:end],
		CurrentPage:   page,
		TotalPages:    totalPages,
		HasPrev:       page > 1,
		HasNext:       page < totalPages,
		SearchQuery:   searchQuery,
		Lists:         lists,
		SelectedLists: selectedLists,
		ActiveList:    activeList,
	}

	if err := s.templates.ExecuteTemplate(w, "index.html", data); err != nil {
		log.Printf("Error executing index template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (s *Server) handleThread(w http.ResponseWriter, r *http.Request) {
	// Extract thread ID from URL path
	parts := strings.Split(r.URL.Path, "/thread/")
	if len(parts) != 2 {
		log.Printf("Invalid thread URL path: %s", r.URL.Path)
		http.NotFound(w, r)
		return
	}

	// Get the raw thread ID and decode it
	threadID, err := url.PathUnescape(parts[1])
	if err != nil {
		log.Printf("Error unescaping thread ID: %v", err)
		http.NotFound(w, r)
		return
	}

	// Remove any trailing slashes
	threadID = strings.TrimRight(threadID, "/")

	if threadID == "" {
		log.Printf("Empty thread ID in URL path: %s", r.URL.Path)
		http.NotFound(w, r)
		return
	}

	log.Printf("Looking up thread with ID: %s", threadID)

	// Find the thread in any of the mailing lists
	var thread *Thread
	var list *MailingList
	for _, ml := range s.lists {
		if t, exists := ml.Threads[threadID]; exists {
			thread = t
			list = ml
			break
		}
	}

	if thread == nil {
		log.Printf("Thread not found: %s", threadID)
		http.NotFound(w, r)
		return
	}

	// Add debug logging
	log.Printf("Found thread: %s with %d replies", thread.FirstMessage.Subject, len(thread.Replies))

	data := ThreadData{
		Title:  thread.FirstMessage.Subject,
		Thread: thread,
		List:   &list.Config,
	}

	if err := s.templates.ExecuteTemplate(w, "thread.html", data); err != nil {
		log.Printf("Error executing thread template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
