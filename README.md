# keygen-client
keygen API client written in golang

## HTTP Error Handling

The client now returns HTTP status codes for better error handling. When API requests fail, you can extract the HTTP status code to provide specific user feedback:

```go
err := client.ActivateMachine(ctx, licenseKey, fingerprint, "", "")
if httpErr, ok := err.(*keygen.HTTPError); ok {
    switch httpErr.StatusCode {
    case 422:
        // Handle machine limit exceeded
        fmt.Println("Machine limit exceeded - please deactivate an existing machine")
    case 401:
        // Handle unauthorized
        fmt.Println("Invalid API credentials")
    case 404:
        // Handle not found
        fmt.Println("License not found")
    default:
        fmt.Printf("API error (HTTP %d): %s\n", httpErr.StatusCode, err.Error())
    }
} else {
    // Handle non-HTTP errors (network issues, etc.)
    fmt.Printf("Request failed: %s\n", err.Error())
}
```

The HTTPError type includes:
- `StatusCode int` - HTTP status code (422, 401, 404, etc.)
- `Method string` - HTTP method (GET, POST, etc.)
- `Path string` - API endpoint path
- `Body string` - Full error message
- `Err error` - Underlying error for unwrapping

## Backward Compatibility

Existing error handling code continues to work without changes:

```go
// This still works exactly as before
err := client.ActivateMachine(ctx, licenseKey, fingerprint, "", "")
if err != nil {
    log.Printf("Error: %s", err.Error())
}
```

## Development

Command to run the test using keygen envs
    ```bash
    KEYGEN_ACCOUNT_ID= \
    KEYGEN_API_TOKEN= \
    KEYGEN_POLICY_ID= \
    go test ./keygen
    ```