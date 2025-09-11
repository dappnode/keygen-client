package keygen

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Example demonstrating how consumers can extract HTTP status codes from errors
func ExampleHTTPError_StatusCode() {
	// Create a test server that returns HTTP 422 (machine limit exceeded)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.api+json")
		w.WriteHeader(422)
		w.Write([]byte(`{"errors":[{"title":"Unprocessable resource","detail":"machine count has exceeded maximum allowed for license (1)","code":"MACHINE_LIMIT_EXCEEDED"}]}`))
	}))
	defer server.Close()

	client := New("test-account", "test-token", WithBaseURL(server.URL))

	// Try to activate a machine (this will fail with 422)
	ctx := context.Background()
	err := client.ActivateMachine(ctx, "test-license-key", "test-fingerprint", "", "")

	if err != nil {
		// Check if it's an HTTPError to get the status code
		if httpErr, ok := err.(*HTTPError); ok {
			switch httpErr.StatusCode {
			case 422:
				fmt.Println("Machine limit exceeded - user should deactivate an existing machine")
			case 401:
				fmt.Println("Unauthorized - invalid API token")
			case 404:
				fmt.Println("License not found")
			default:
				fmt.Printf("HTTP error %d: %s\n", httpErr.StatusCode, err.Error())
			}
		} else {
			fmt.Printf("Non-HTTP error: %s\n", err.Error())
		}
	}

	// Output:
	// Machine limit exceeded - user should deactivate an existing machine
}

// Test demonstrating backward compatibility - existing error checking still works
func TestBackwardCompatibility(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(422)
		w.Write([]byte(`{"errors":[{"code":"MACHINE_LIMIT_EXCEEDED"}]}`))
	}))
	defer server.Close()

	client := New("test-account", "test-token", WithBaseURL(server.URL))
	ctx := context.Background()

	// This existing pattern still works exactly the same
	err := client.ActivateMachine(ctx, "test-key", "test-fingerprint", "", "")
	if err != nil {
		t.Logf("Error occurred as expected: %s", err.Error())
		// The error message still contains all the same information
		if err.Error() == "" {
			t.Error("Error message should not be empty")
		}
	} else {
		t.Error("Expected an error but got none")
	}

	// But now consumers can also get the HTTP status code if they want
	if httpErr, ok := err.(*HTTPError); ok {
		t.Logf("HTTP Status Code: %d", httpErr.StatusCode)
		if httpErr.StatusCode != 422 {
			t.Errorf("Expected status code 422, got %d", httpErr.StatusCode)
		}
	} else {
		t.Error("Expected HTTPError")
	}
}