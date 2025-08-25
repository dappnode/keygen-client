package keygen

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
