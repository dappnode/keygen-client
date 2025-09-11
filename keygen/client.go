package keygen

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

// Client is a thin Keygen API wrapper.
type Client struct {
	accountID          string
	apiToken           string
	baseURL            string
	http               *http.Client
	defaultMachineName string
	defaultPlatform    string
}

// Option configures the Client.
type Option func(*Client)

// WithHTTPClient sets a custom http.Client.
func WithHTTPClient(h *http.Client) Option {
	return func(c *Client) { c.http = h }
}

// WithBaseURL overrides the API base URL (default: https://api.keygen.sh/v1).
func WithBaseURL(u string) Option {
	return func(c *Client) { c.baseURL = u }
}

// WithDefaultMachine sets defaults used by ActivateMachine.
func WithDefaultMachine(name, platform string) Option {
	return func(c *Client) {
		if name != "" {
			c.defaultMachineName = name
		}
		if platform != "" {
			c.defaultPlatform = platform
		}
	}
}

// New creates a new Client.
func New(accountID, apiToken string, opts ...Option) *Client {
	c := &Client{
		accountID:          accountID,
		apiToken:           apiToken,
		baseURL:            "https://api.keygen.sh/v1",
		http:               http.DefaultClient,
		defaultMachineName: "dappnode",
		defaultPlatform:    "linux",
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// --- Licenses ---

// CreateLicense creates a new license under a policy, returning its key.
func (c *Client) CreateLicense(ctx context.Context, policyID string, meta LicenseMetadata) (string, int, error) {
	path := fmt.Sprintf("/accounts/%s/licenses", c.accountID)
	req := licenseCreateRequest{
		Data: licenseCreateData{
			Type: "licenses",
			Attributes: licenseCreateAttributes{
				Metadata: meta, // subscriptionId + customerEmail
			},
			Relationships: licenseCreateRelationships{
				Policy: licenseRelationship{
					Data: relationshipData{Type: "policies", ID: policyID},
				},
			},
		},
	}

	var resp licenseCreateResponse
	httpCode, err := c.do(ctx, http.MethodPost, path, req, &resp)
	if err != nil {
		return "", httpCode, err
	}
	key := resp.Data.Attributes.Key
	if key == "" {
		return "", httpCode, fmt.Errorf("keygen: license creation returned empty key")
	}
	return key, httpCode, nil
}

// DeleteLicense deletes a license by ID (204 on success).
func (c *Client) DeleteLicense(ctx context.Context, licenseID string) (int, error) {
	path := fmt.Sprintf("/accounts/%s/licenses/%s", c.accountID, licenseID)
	httpCode, err := c.do(ctx, http.MethodDelete, path, nil, nil)
	return httpCode, err
}

// GetLicenseBySubscriptionID returns the license ID for a metadata[subscriptionId].
func (c *Client) GetLicenseBySubscriptionID(ctx context.Context, subscriptionID string) (string, int, error) {
	q := url.Values{}
	q.Set("metadata[subscriptionId]", subscriptionID)
	path := fmt.Sprintf("/accounts/%s/licenses?%s", c.accountID, q.Encode())

	var resp getLicenseBySubscriptionResponse
	httpCode, err := c.do(ctx, http.MethodGet, path, nil, &resp)
	if err != nil {
		return "", httpCode, err
	}

	if len(resp.Data) == 0 {
		return "", httpCode, nil // keep current behavior: not an error if missing
	}
	if len(resp.Data) > 1 {
		return "", httpCode, fmt.Errorf("keygen: multiple licenses for subscriptionId %s", subscriptionID)
	}
	if resp.Data[0].ID == "" {
		return "", httpCode, fmt.Errorf("keygen: empty license id for subscriptionId %s", subscriptionID)
	}
	return resp.Data[0].ID, httpCode, nil
}

// ListLicensesByPolicy returns a rich view (ID, Key*, Status*, Metadata).
// Key/Status may be empty when the API/resource view omits them.
func (c *Client) ListLicensesByPolicy(ctx context.Context, policyID string) ([]LicenseSummary, int, error) {
	var out []LicenseSummary
	page := 1
	var lastHttpCode int

	for {
		q := url.Values{}
		q.Set("policy", policyID)
		q.Set("page[number]", strconv.Itoa(page))
		q.Set("page[size]", "100")
		path := fmt.Sprintf("/accounts/%s/licenses?%s", c.accountID, q.Encode())

		var resp listLicensesByPolicyResponse
		httpCode, err := c.do(ctx, http.MethodGet, path, nil, &resp)
		lastHttpCode = httpCode
		if err != nil {
			return nil, httpCode, err
		}
		for _, d := range resp.Data {
			out = append(out, LicenseSummary{
				ID:       d.ID,
				Key:      d.Attributes.Key,
				Status:   d.Attributes.Status,
				Metadata: d.Attributes.Metadata,
			})
		}
		if resp.Links.Next == nil {
			break
		}
		page++
	}
	return out, lastHttpCode, nil
}

// ListLicenseKeysByPolicy is a convenience wrapper returning only keys.
func (c *Client) ListLicenseKeysByPolicy(ctx context.Context, policyID string) ([]string, int, error) {
	items, httpCode, err := c.ListLicensesByPolicy(ctx, policyID)
	if err != nil {
		return nil, httpCode, err
	}
	keys := make([]string, 0, len(items))
	for _, it := range items {
		if it.Key != "" {
			keys = append(keys, it.Key)
		}
	}
	return keys, httpCode, nil
}

// --- Machines ---

// ActivateMachine creates a machine bound to the license (by key).
// name/platform default to the client defaults if empty.
func (c *Client) ActivateMachine(ctx context.Context, licenseKey, fingerprint, name, platform string) (int, error) {
	licenseID, httpCode, err := c.ResolveLicenseID(ctx, licenseKey)
	if err != nil {
		return httpCode, err
	}
	if name == "" {
		name = c.defaultMachineName
	}
	if platform == "" {
		platform = c.defaultPlatform
	}

	req := createMachineRequest{
		Data: machineData{
			Type: "machines",
			Attributes: machineAttributes{
				Fingerprint: fingerprint,
				Platform:    platform,
				Name:        name,
			},
			Relationships: machineRelationships{
				License: licenseRelationship{
					Data: relationshipData{Type: "licenses", ID: licenseID},
				},
			},
		},
	}
	httpCode, err = c.do(ctx, http.MethodPost, fmt.Sprintf("/accounts/%s/machines", c.accountID), req, &machineResponse{})
	return httpCode, err
}

// DeactivateMachine deletes a machine (by matching fingerprint) from the license.
// Returns (found, error). When found==false and err==nil, no machine matched.
func (c *Client) DeactivateMachine(ctx context.Context, licenseKey, fingerprint string) (bool, int, error) {
	licenseID, httpCode, err := c.ResolveLicenseID(ctx, licenseKey)
	if err != nil {
		return false, httpCode, err
	}

	list, httpCode, err := c.ListMachines(ctx, licenseID)
	if err != nil {
		return false, httpCode, err
	}

	for _, m := range list {
		if m.Fingerprint == fingerprint {
			httpCode, err := c.do(ctx, http.MethodDelete,
				fmt.Sprintf("/accounts/%s/machines/%s", c.accountID, m.ID), nil, nil)
			return true, httpCode, err
		}
	}
	return false, httpCode, nil
}

// ListMachines lists machines for a license by licenseID.
// If no machines exist, returns an empty slice.
func (c *Client) ListMachines(ctx context.Context, licenseID string) ([]Machine, int, error) {
	q := url.Values{}
	page := 1
	q.Set("license", licenseID)
	q.Set("page[number]", strconv.Itoa(page))
	q.Set("page[size]", "100")
	path := fmt.Sprintf("/accounts/%s/machines?%s", c.accountID, q.Encode())

	var resp machinesListResponse
	httpCode, err := c.do(ctx, http.MethodGet, path, nil, &resp)
	if err != nil {
		return nil, httpCode, err
	}
	out := make([]Machine, 0, len(resp.Data))
	for _, d := range resp.Data {
		out = append(out, Machine{
			ID:          d.ID,
			Fingerprint: d.Attributes.Fingerprint,
			Platform:    d.Attributes.Platform,
			Name:        d.Attributes.Name,
		})
	}
	return out, httpCode, nil
}

// ListAllMachines lists all machines for the account.
func (c *Client) ListAllMachines(ctx context.Context) ([]Machine, int, error) {
	var out []Machine
	page := 1
	var lastHttpCode int

	for {
		q := url.Values{}
		q.Set("page[number]", strconv.Itoa(page))
		q.Set("page[size]", "100")
		path := fmt.Sprintf("/accounts/%s/machines?%s", c.accountID, q.Encode())

		var resp machinesListResponse
		httpCode, err := c.do(ctx, http.MethodGet, path, nil, &resp)
		lastHttpCode = httpCode
		if err != nil {
			return nil, httpCode, err
		}
		for _, d := range resp.Data {
			out = append(out, Machine{
				ID:          d.ID,
				LicenseId:   d.Relationships.License.Data.ID,
				Fingerprint: d.Attributes.Fingerprint,
				Platform:    d.Attributes.Platform,
				Name:        d.Attributes.Name,
			})
		}
		// Get out of the loop if no more pages
		if resp.Links.Next == nil {
			break
		}
		page++
	}
	return out, lastHttpCode, nil
}

// --- Validation ---

// Validate checks a key within a fingerprint scope.
func (c *Client) Validate(ctx context.Context, licenseKey, fingerprint string) (LicenseValidation, int, error) {
	req := validateLicenseRequest{
		Meta: validateMeta{
			Key: licenseKey,
			Scope: fingerprintScope{
				Fingerprint: fingerprint,
			},
		},
	}

	var resp licenseValidationResponse
	httpCode, err := c.do(ctx, http.MethodPost,
		fmt.Sprintf("/accounts/%s/licenses/actions/validate-key", c.accountID),
		req, &resp)
	if err != nil {
		return LicenseValidation{}, httpCode, err
	}

	return LicenseValidation{
		Key:         resp.Data.Attributes.Key,
		Expiry:      resp.Data.Attributes.Expiry,
		Status:      resp.Data.Attributes.Status,
		Valid:       resp.Meta.Valid,
		Code:        resp.Meta.Code,
		Detail:      resp.Meta.Detail,
		Timestamp:   resp.Meta.Timestamp,
		Fingerprint: resp.Meta.Scope.Fingerprint,
	}, httpCode, nil
}

// ResolveLicenseID gets the license ID from a key using validate-key.
func (c *Client) ResolveLicenseID(ctx context.Context, licenseKey string) (string, int, error) {
	req := resolveLicenseIDRequest{}
	req.Meta.Key = licenseKey

	var resp licenseValidationResponse
	httpCode, err := c.do(ctx, http.MethodPost,
		fmt.Sprintf("/accounts/%s/licenses/actions/validate-key", c.accountID),
		req, &resp)
	if err != nil {
		return "", httpCode, err
	}
	if resp.Data.ID == "" {
		return "", httpCode, fmt.Errorf("keygen: no license id found of licenseKey %s", licenseKey)
	}
	return resp.Data.ID, httpCode, nil
}

// --- HTTP plumbing ---

func (c *Client) do(ctx context.Context, method, path string, in any, out any) (int, error) {
	var body io.Reader
	if in != nil {
		var buf bytes.Buffer
		if err := json.NewEncoder(&buf).Encode(in); err != nil {
			return 0, fmt.Errorf("keygen: encode request: %w", err)
		}
		body = &buf
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return 0, fmt.Errorf("keygen: new request: %w", err)
	}
	if in != nil {
		req.Header.Set("Content-Type", "application/vnd.api+json")
	}
	req.Header.Set("Accept", "application/vnd.api+json")
	req.Header.Set("Authorization", "Bearer "+c.apiToken)

	resp, err := c.http.Do(req)
	if err != nil {
		return 0, fmt.Errorf("keygen: do request: %w", err)
	}
	defer resp.Body.Close()

	httpCode := resp.StatusCode

	// Non-2xx => return body as plain error string (no custom types)
	if httpCode < 200 || httpCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		if len(b) == 0 {
			return httpCode, fmt.Errorf("keygen: %s %s -> HTTP %d", method, path, httpCode)
		}
		return httpCode, fmt.Errorf("keygen: %s %s -> HTTP %d: %s", method, path, httpCode, string(b))
	}

	if out == nil {
		// Drain for keep-alives anyway
		io.Copy(io.Discard, resp.Body)
		return httpCode, nil
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return httpCode, fmt.Errorf("keygen: decode response: %w", err)
	}
	return httpCode, nil
}
