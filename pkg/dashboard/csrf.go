package dashboard

// CSRFTokenResponse represents the response containing an anti-CSRF token.
type CSRFTokenResponse struct {
	CSRFToken string `json:"csrf_token"`
}
