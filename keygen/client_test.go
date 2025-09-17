package keygen

import (
	"context"
	"os"
	"testing"
)

func getTestClient() *Client {
	accountID := os.Getenv("KEYGEN_ACCOUNT_ID")
	apiToken := os.Getenv("KEYGEN_API_TOKEN")
	return New(accountID, apiToken, WithDefaultMachine("dappnode", "linux"))
}

func TestCreateLicense(t *testing.T) {
	ctx := context.Background()
	client := getTestClient()
	policyID := os.Getenv("KEYGEN_POLICY_ID")
	subscriptionID := os.Getenv("KEYGEN_SUBSCRIPTION_ID")
	customerEmail := os.Getenv("KEYGEN_CUSTOMER_EMAIL")
	meta := LicenseMetadata{
		SubscriptionID: subscriptionID,
		CustomerEmail:  customerEmail,
	}
	result, httpCode, err := client.CreateLicense(ctx, policyID, meta)
	if err != nil {
		t.Fatalf("CreateLicense error: %v (HTTP %d)", err, httpCode)
	}
	t.Logf("CreateLicense result: %v, HTTP code: %d", result, httpCode)
}

func TestDeleteLicense(t *testing.T) {
	ctx := context.Background()
	client := getTestClient()
	licenseID := os.Getenv("KEYGEN_LICENSE_ID")
	httpCode, err := client.DeleteLicense(ctx, licenseID)
	if err != nil {
		t.Fatalf("DeleteLicense error: %v (HTTP %d)", err, httpCode)
	}
	t.Logf("DeleteLicense succeeded, HTTP code: %d", httpCode)
}

func TestGetLicenseBySubscriptionID(t *testing.T) {
	ctx := context.Background()
	client := getTestClient()
	subscriptionID := os.Getenv("KEYGEN_SUBSCRIPTION_ID")
	result, httpCode, err := client.GetLicenseBySubscriptionID(ctx, subscriptionID)
	if err != nil {
		t.Fatalf("GetLicenseBySubscriptionID error: %v (HTTP %d)", err, httpCode)
	}
	t.Logf("GetLicenseBySubscriptionID result: %v, HTTP code: %d", result, httpCode)
}

func TestListLicensesByPolicy(t *testing.T) {
	ctx := context.Background()
	client := getTestClient()
	policyID := os.Getenv("KEYGEN_POLICY_ID")
	result, httpCode, err := client.ListLicensesByPolicy(ctx, policyID)
	if err != nil {
		t.Fatalf("ListLicensesByPolicy error: %v (HTTP %d)", err, httpCode)
	}
	t.Logf("ListLicensesByPolicy result: %+v, HTTP code: %d", result, httpCode)
}

func TestListLicenseKeysByPolicy(t *testing.T) {
	ctx := context.Background()
	client := getTestClient()
	policyID := os.Getenv("KEYGEN_POLICY_ID")
	result, httpCode, err := client.ListLicenseKeysByPolicy(ctx, policyID)
	if err != nil {
		t.Fatalf("ListLicenseKeysByPolicy error: %v (HTTP %d)", err, httpCode)
	}
	t.Logf("ListLicenseKeysByPolicy result: %+v, HTTP code: %d", result, httpCode)
}

func TestActivateMachine(t *testing.T) {
	ctx := context.Background()
	client := getTestClient()
	licenseKey := os.Getenv("KEYGEN_LICENSE_KEY")
	fingerprint := os.Getenv("KEYGEN_FINGERPRINT")
	httpCode, err := client.ActivateMachine(ctx, licenseKey, fingerprint, "", "")
	if err != nil {
		t.Fatalf("ActivateMachine error: %v (HTTP %d)", err, httpCode)
	}
	t.Logf("ActivateMachine succeeded, HTTP code: %d", httpCode)
}

func TestDeactivateMachine(t *testing.T) {
	ctx := context.Background()
	client := getTestClient()
	licenseKey := os.Getenv("KEYGEN_LICENSE_KEY")
	fingerprint := os.Getenv("KEYGEN_FINGERPRINT")
	found, httpCode, err := client.DeactivateMachine(ctx, licenseKey, fingerprint)
	if err != nil {
		t.Fatalf("DeactivateMachine error: %v (HTTP %d)", err, httpCode)
	}
	t.Logf("DeactivateMachine found: %v, HTTP code: %d", found, httpCode)
}

func TestListMachines(t *testing.T) {
	ctx := context.Background()
	client := getTestClient()
	licenseID := os.Getenv("KEYGEN_LICENSE_ID")
	result, httpCode, err := client.ListMachines(ctx, licenseID)
	if err != nil {
		t.Fatalf("ListMachines error: %v (HTTP %d)", err, httpCode)
	}
	t.Logf("ListMachines result: %+v, HTTP code: %d", result, httpCode)
}

func TestListAllMachines(t *testing.T) {
	ctx := context.Background()
	client := getTestClient()
	result, httpCode, err := client.ListAllMachines(ctx)
	if err != nil {
		t.Fatalf("ListAllMachines error: %v (HTTP %d)", err, httpCode)
	}
	t.Logf("ListAllMachines result: %+v, HTTP code: %d", result, httpCode)
}

func TestValidate(t *testing.T) {
	ctx := context.Background()
	client := getTestClient()
	licenseKey := os.Getenv("KEYGEN_LICENSE_KEY")
	fingerprint := os.Getenv("KEYGEN_FINGERPRINT")
	result, httpCode, err := client.Validate(ctx, licenseKey, fingerprint)
	if err != nil {
		t.Fatalf("Validate error: %v (HTTP %d)", err, httpCode)
	}
	t.Logf("Validate result: %+v, HTTP code: %d", result, httpCode)
}

func TestResolveLicenseID(t *testing.T) {
	ctx := context.Background()
	client := getTestClient()
	licenseKey := os.Getenv("KEYGEN_LICENSE_KEY")
	result, httpCode, err := client.ResolveLicenseID(ctx, licenseKey)
	if err != nil {
		t.Fatalf("ResolveLicenseID error: %v (HTTP %d)", err, httpCode)
	}
	t.Logf("ResolveLicenseID result: %v, HTTP code: %d", result, httpCode)
}
