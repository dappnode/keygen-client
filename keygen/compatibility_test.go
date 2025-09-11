package keygen

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestErrorMessageFormat verifies that error messages maintain the exact same format
// as before to ensure backward compatibility
func TestErrorMessageFormat(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		responseBody   string
		expectedFormat string
	}{
		{
			name:           "error with body",
			statusCode:     422,
			responseBody:   `{"errors":[{"code":"MACHINE_LIMIT_EXCEEDED"}]}`,
			expectedFormat: `keygen: POST /accounts/test-account/licenses -> HTTP 422: {"errors":[{"code":"MACHINE_LIMIT_EXCEEDED"}]}`,
		},
		{
			name:           "error without body",
			statusCode:     500,
			responseBody:   "",
			expectedFormat: "keygen: POST /accounts/test-account/licenses -> HTTP 500",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				if tt.responseBody != "" {
					w.Write([]byte(tt.responseBody))
				}
			}))
			defer server.Close()

			client := New("test-account", "test-token", WithBaseURL(server.URL))
			ctx := context.Background()

			_, err := client.CreateLicense(ctx, "test-policy", LicenseMetadata{})

			if err == nil {
				t.Fatal("Expected an error")
			}

			// Check that it's an HTTPError
			httpErr, ok := err.(*HTTPError)
			if !ok {
				t.Fatalf("Expected HTTPError, got %T", err)
			}

			// Check that the error message matches expected format exactly
			if httpErr.Error() != tt.expectedFormat {
				t.Errorf("Expected error message:\n%q\nGot:\n%q", tt.expectedFormat, httpErr.Error())
			}

			// Check that the status code is accessible
			if httpErr.StatusCode != tt.statusCode {
				t.Errorf("Expected status code %d, got %d", tt.statusCode, httpErr.StatusCode)
			}
		})
	}
}

// TestAllClientMethods verifies that all client methods can return HTTPError
func TestAllClientMethods_ReturnHTTPError(t *testing.T) {
	// Create server that always returns 422
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(422)
		w.Write([]byte(`{"errors":[{"code":"TEST_ERROR"}]}`))
	}))
	defer server.Close()

	client := New("test-account", "test-token", WithBaseURL(server.URL))
	ctx := context.Background()

	// Test methods that return errors (not exhaustive but representative)
	methodTests := []struct {
		name   string
		method func() error
	}{
		{
			name: "CreateLicense",
			method: func() error {
				_, err := client.CreateLicense(ctx, "policy", LicenseMetadata{})
				return err
			},
		},
		{
			name: "DeleteLicense",
			method: func() error {
				return client.DeleteLicense(ctx, "license-id")
			},
		},
		{
			name: "ActivateMachine",
			method: func() error {
				return client.ActivateMachine(ctx, "license-key", "fingerprint", "", "")
			},
		},
		{
			name: "GetLicenseBySubscriptionID",
			method: func() error {
				_, err := client.GetLicenseBySubscriptionID(ctx, "sub-id")
				return err
			},
		},
	}

	for _, tt := range methodTests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.method()
			if err == nil {
				t.Fatal("Expected an error")
			}

			httpErr, ok := err.(*HTTPError)
			if !ok {
				t.Fatalf("Expected HTTPError, got %T: %v", err, err)
			}

			if httpErr.StatusCode != 422 {
				t.Errorf("Expected status code 422, got %d", httpErr.StatusCode)
			}

			// Verify error message contains expected parts
			errorMsg := err.Error()
			if !strings.Contains(errorMsg, "HTTP 422") {
				t.Errorf("Error message should contain 'HTTP 422': %s", errorMsg)
			}
		})
	}
}