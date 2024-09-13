package mails

import (
	"fmt"
	"net/http"
	"time"
)

// Tracker tracks email opens and clicks
type Tracker struct {
	BaseURL string
}

// NewTracker creates a new Tracker
func NewTracker(baseURL string) *Tracker {
	return &Tracker{BaseURL: baseURL}
}

// TrackOpen generates a URL for tracking email opens
func (t *Tracker) TrackOpen(emailID string) string {
	return fmt.Sprintf("%s/open/%s", t.BaseURL, emailID)
}

// TrackClick generates a URL for tracking email clicks
func (t *Tracker) TrackClick(emailID, url string) string {
	return fmt.Sprintf("%s/click/%s?url=%s", t.BaseURL, emailID, url)
}

// HandleOpen handles email open tracking
func (t *Tracker) HandleOpen(w http.ResponseWriter, r *http.Request) {
	emailID := r.URL.Path[len("/open/"):]
	fmt.Printf("Email %s opened at %v\n", emailID, time.Now())
	http.ServeFile(w, r, "pixel.png")
}

// HandleClick handles email click tracking
func (t *Tracker) HandleClick(w http.ResponseWriter, r *http.Request) {
	emailID := r.URL.Path[len("/click/"):]
	url := r.URL.Query().Get("url")
	fmt.Printf("Email %s link clicked: %s at %v\n", emailID, url, time.Now())
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}
