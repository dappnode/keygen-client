package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"

	"github.com/dappnode/keygen-client/keygen"
)

func main() {
	fmt.Println("🚀 Keygen Client HTTPError Demo")
	fmt.Println("================================")

	// Demo 1: Machine Limit Exceeded (HTTP 422)
	fmt.Println("\n📋 Demo 1: Machine Limit Exceeded (HTTP 422)")
	demo422()

	// Demo 2: Unauthorized (HTTP 401)
	fmt.Println("\n📋 Demo 2: Unauthorized (HTTP 401)")
	demo401()

	// Demo 3: Successful Request (HTTP 200)
	fmt.Println("\n📋 Demo 3: Successful Request (HTTP 200)")
	demoSuccess()

	// Demo 4: Backward Compatibility
	fmt.Println("\n📋 Demo 4: Backward Compatibility")
	demoBackwardCompatibility()
}

func demo422() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.api+json")
		w.WriteHeader(422)
		w.Write([]byte(`{"errors":[{"title":"Unprocessable resource","detail":"machine count has exceeded maximum allowed for license (1)","code":"MACHINE_LIMIT_EXCEEDED","source":{"pointer":"/data"},"links":{"about":"https://keygen.sh/docs/api/machines/#machines-object"}}]}`))
	}))
	defer server.Close()

	client := keygen.New("test-account", "test-token", keygen.WithBaseURL(server.URL))
	err := client.ActivateMachine(context.Background(), "test-license", "test-fingerprint", "", "")

	if httpErr, ok := err.(*keygen.HTTPError); ok {
		fmt.Printf("   ✅ Detected HTTPError with status code: %d\n", httpErr.StatusCode)
		fmt.Printf("   📍 Method: %s, Path: %s\n", httpErr.Method, httpErr.Path)
		
		switch httpErr.StatusCode {
		case 422:
			fmt.Printf("   💡 User feedback: Machine limit exceeded - please deactivate an existing machine\n")
		}
	} else {
		log.Printf("   ❌ Expected HTTPError, got: %T", err)
	}
}

func demo401() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
		w.Write([]byte(`{"errors":[{"title":"Unauthorized","detail":"Invalid API token"}]}`))
	}))
	defer server.Close()

	client := keygen.New("test-account", "invalid-token", keygen.WithBaseURL(server.URL))
	_, err := client.CreateLicense(context.Background(), "test-policy", keygen.LicenseMetadata{})

	if httpErr, ok := err.(*keygen.HTTPError); ok {
		fmt.Printf("   ✅ Detected HTTPError with status code: %d\n", httpErr.StatusCode)
		
		switch httpErr.StatusCode {
		case 401:
			fmt.Printf("   💡 User feedback: Invalid API credentials - please check your token\n")
		}
	}
}

func demoSuccess() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.api+json")
		w.WriteHeader(200)
		w.Write([]byte(`{"data":{"attributes":{"key":"demo-license-key-12345"}}}`))
	}))
	defer server.Close()

	client := keygen.New("test-account", "test-token", keygen.WithBaseURL(server.URL))
	key, err := client.CreateLicense(context.Background(), "test-policy", keygen.LicenseMetadata{
		SubscriptionID: "demo-subscription",
		CustomerEmail:  "demo@example.com",
	})

	if err != nil {
		fmt.Printf("   ❌ Unexpected error: %v\n", err)
	} else {
		fmt.Printf("   ✅ License created successfully: %s\n", key)
		fmt.Printf("   💡 No HTTPError for successful requests\n")
	}
}

func demoBackwardCompatibility() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		w.Write([]byte(`{"errors":[{"title":"Not Found","detail":"License not found"}]}`))
	}))
	defer server.Close()

	client := keygen.New("test-account", "test-token", keygen.WithBaseURL(server.URL))
	err := client.DeleteLicense(context.Background(), "non-existent-license")

	// Traditional error handling (still works!)
	if err != nil {
		fmt.Printf("   ✅ Traditional error handling: %s\n", err.Error())
		
		// But now we can also get the HTTP status code when needed
		if httpErr, ok := err.(*keygen.HTTPError); ok {
			fmt.Printf("   🔧 Enhanced: HTTP status code %d available for detailed handling\n", httpErr.StatusCode)
		}
	}
}