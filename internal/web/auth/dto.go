package auth

type SessionDTO struct {
	LoggedIn bool   `json:"logged_in"`
	LoginURL string `json:"login_url,omitempty"`
	Message  string `json:"message,omitempty"`
}
