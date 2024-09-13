package gudu

import "net/http"

// SessionLoadAndSave load and save session data for requests
func (g *Gudu) SessionLoadAndSave(next http.Handler) http.Handler {
	return g.Sessions.LoadAndSave(next)
}
