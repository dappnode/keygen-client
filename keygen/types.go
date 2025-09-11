package keygen

// HTTPError represents an error response from the Keygen API that includes the HTTP status code.
type HTTPError struct {
	Method     string // HTTP method (GET, POST, etc.)
	Path       string // API path
	StatusCode int    // HTTP status code
	Body       string // Response body
	Err        error  // Underlying error if any
}

// Error implements the error interface.
func (e *HTTPError) Error() string {
	if e.Body != "" {
		return e.Body
	}
	return e.Err.Error()
}

// Unwrap allows error unwrapping.
func (e *HTTPError) Unwrap() error {
	return e.Err
}

// LicenseMetadata mirrors the structured metadata you already use.
type LicenseMetadata struct {
	SubscriptionID string `json:"subscriptionId"`
	CustomerEmail  string `json:"customerEmail"`
}

// LicenseSummary is a normalized view for listing by policy.
type LicenseSummary struct {
	ID       string         `json:"id"`
	Key      string         `json:"key,omitempty"`    // may be empty depending on API shape
	Status   string         `json:"status,omitempty"` // may be empty depending on API shape
	Metadata map[string]any `json:"metadata,omitempty"`
}

// Machine is a simplified machine representation.
type Machine struct {
	ID          string `json:"id"`
	LicenseId   string `json:"licenseId"`
	Fingerprint string `json:"fingerprint"`
	Platform    string `json:"platform"`
	Name        string `json:"name"`
}

// LicenseValidation unifies the validate-key output
type LicenseValidation struct {
	Key         string `json:"key"`
	Expiry      string `json:"expiry"`
	Status      string `json:"status"`
	Valid       bool   `json:"valid"`
	Code        string `json:"code"`
	Detail      string `json:"detail"`
	Timestamp   string `json:"ts"`
	Fingerprint string `json:"fingerprint"`
}
