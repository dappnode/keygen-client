package keygen

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPError_Interface(t *testing.T) {
	// Test that HTTPError implements the error interface
	var err error = &HTTPError{
		Method:     "POST",
		Path:       "/test",
		StatusCode: 422,
		Body:       "keygen: POST /test -> HTTP 422: test error",
		Err:        fmt.Errorf("HTTP 422: test error"),
	}

	if err.Error() != "keygen: POST /test -> HTTP 422: test error" {
		t.Errorf("Expected error message to be 'keygen: POST /test -> HTTP 422: test error', got %q", err.Error())
	}

	// Test unwrapping
	httpErr := err.(*HTTPError)
	if httpErr.StatusCode != 422 {
		t.Errorf("Expected status code 422, got %d", httpErr.StatusCode)
	}
	if httpErr.Method != "POST" {
		t.Errorf("Expected method POST, got %s", httpErr.Method)
	}
	if httpErr.Path != "/test" {
		t.Errorf("Expected path /test, got %s", httpErr.Path)
	}
}

func TestClient_HTTPErrorIntegration(t *testing.T) {
	// Create a test server that returns HTTP 422
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.api+json")
		w.WriteHeader(422)
		w.Write([]byte(`{"errors":[{"title":"Unprocessable resource","detail":"machine count has exceeded maximum allowed for license (1)","code":"MACHINE_LIMIT_EXCEEDED"}]}`))
	}))
	defer server.Close()

	// Create client with test server URL
	client := New("test-account", "test-token", WithBaseURL(server.URL))

	// Try to create a license (this will hit our test server and return 422)
	ctx := context.Background()
	_, err := client.CreateLicense(ctx, "test-policy", LicenseMetadata{
		SubscriptionID: "test-sub",
		CustomerEmail:  "test@example.com",
	})

	// Verify we get an HTTPError
	if err == nil {
		t.Fatal("Expected an error, got nil")
	}

	httpErr, ok := err.(*HTTPError)
	if !ok {
		t.Fatalf("Expected HTTPError, got %T: %v", err, err)
	}

	// Verify error properties
	if httpErr.StatusCode != 422 {
		t.Errorf("Expected status code 422, got %d", httpErr.StatusCode)
	}
	if httpErr.Method != "POST" {
		t.Errorf("Expected method POST, got %s", httpErr.Method)
	}
	if httpErr.Path != "/accounts/test-account/licenses" {
		t.Errorf("Expected path /accounts/test-account/licenses, got %s", httpErr.Path)
	}

	// Verify error message contains expected information
	errorMsg := err.Error()
	if errorMsg == "" {
		t.Error("Error message should not be empty")
	}
	t.Logf("Error message: %s", errorMsg)
}

func TestClient_SuccessfulRequestNoHTTPError(t *testing.T) {
	// Create a test server that returns HTTP 200
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.api+json")
		w.WriteHeader(200)
		w.Write([]byte(`{"data":{"attributes":{"key":"test-license-key"}}}`))
	}))
	defer server.Close()

	// Create client with test server URL
	client := New("test-account", "test-token", WithBaseURL(server.URL))

	// Try to create a license (this should succeed)
	ctx := context.Background()
	key, err := client.CreateLicense(ctx, "test-policy", LicenseMetadata{
		SubscriptionID: "test-sub",
		CustomerEmail:  "test@example.com",
	})

	// Verify no error for successful request
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if key != "test-license-key" {
		t.Errorf("Expected key 'test-license-key', got %s", key)
	}
}