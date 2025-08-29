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
	result, err := client.CreateLicense(ctx, policyID, meta)
	if err != nil {
		t.Fatalf("CreateLicense error: %v", err)
	}
	t.Logf("CreateLicense result: %v", result)
}

func TestDeleteLicense(t *testing.T) {
	ctx := context.Background()
	client := getTestClient()
	licenseID := os.Getenv("KEYGEN_LICENSE_ID")
	err := client.DeleteLicense(ctx, licenseID)
	if err != nil {
		t.Fatalf("DeleteLicense error: %v", err)
	}
	t.Log("DeleteLicense succeeded")
}

func TestGetLicenseBySubscriptionID(t *testing.T) {
	ctx := context.Background()
	client := getTestClient()
	subscriptionID := os.Getenv("KEYGEN_SUBSCRIPTION_ID")
	result, err := client.GetLicenseBySubscriptionID(ctx, subscriptionID)
	if err != nil {
		t.Fatalf("GetLicenseBySubscriptionID error: %v", err)
	}
	t.Logf("GetLicenseBySubscriptionID result: %v", result)
}

func TestListLicensesByPolicy(t *testing.T) {
	ctx := context.Background()
	client := getTestClient()
	policyID := os.Getenv("KEYGEN_POLICY_ID")
	result, err := client.ListLicensesByPolicy(ctx, policyID)
	if err != nil {
		t.Fatalf("ListLicensesByPolicy error: %v", err)
	}
	t.Logf("ListLicensesByPolicy result: %+v", result)
}

func TestListLicenseKeysByPolicy(t *testing.T) {
	ctx := context.Background()
	client := getTestClient()
	policyID := os.Getenv("KEYGEN_POLICY_ID")
	result, err := client.ListLicenseKeysByPolicy(ctx, policyID)
	if err != nil {
		t.Fatalf("ListLicenseKeysByPolicy error: %v", err)
	}
	t.Logf("ListLicenseKeysByPolicy result: %+v", result)
}

func TestActivateMachine(t *testing.T) {
	ctx := context.Background()
	client := getTestClient()
	licenseKey := os.Getenv("KEYGEN_LICENSE_KEY")
	fingerprint := os.Getenv("KEYGEN_FINGERPRINT")
	err := client.ActivateMachine(ctx, licenseKey, fingerprint, "", "")
	if err != nil {
		t.Fatalf("ActivateMachine error: %v", err)
	}
	t.Log("ActivateMachine succeeded")
}

func TestDeactivateMachine(t *testing.T) {
	ctx := context.Background()
	client := getTestClient()
	licenseKey := os.Getenv("KEYGEN_LICENSE_KEY")
	fingerprint := os.Getenv("KEYGEN_FINGERPRINT")
	found, err := client.DeactivateMachine(ctx, licenseKey, fingerprint)
	if err != nil {
		t.Fatalf("DeactivateMachine error: %v", err)
	}
	t.Logf("DeactivateMachine found: %v", found)
}

func TestListMachines(t *testing.T) {
	ctx := context.Background()
	client := getTestClient()
	licenseID := os.Getenv("KEYGEN_LICENSE_ID")
	result, err := client.ListMachines(ctx, licenseID)
	if err != nil {
		t.Fatalf("ListMachines error: %v", err)
	}
	t.Logf("ListMachines result: %+v", result)
}

func TestListAllMachines(t *testing.T) {
	ctx := context.Background()
	client := getTestClient()
	result, err := client.ListAllMachines(ctx)
	if err != nil {
		t.Fatalf("ListAllMachines error: %v", err)
	}
	t.Logf("ListAllMachines result: %+v", result)
}

func TestValidate(t *testing.T) {
	ctx := context.Background()
	client := getTestClient()
	licenseKey := os.Getenv("KEYGEN_LICENSE_KEY")
	fingerprint := os.Getenv("KEYGEN_FINGERPRINT")
	result, err := client.Validate(ctx, licenseKey, fingerprint)
	if err != nil {
		t.Fatalf("Validate error: %v", err)
	}
	t.Logf("Validate result: %+v", result)
}

func TestResolveLicenseID(t *testing.T) {
	ctx := context.Background()
	client := getTestClient()
	licenseKey := os.Getenv("KEYGEN_LICENSE_KEY")
	result, err := client.ResolveLicenseID(ctx, licenseKey)
	if err != nil {
		t.Fatalf("ResolveLicenseID error: %v", err)
	}
	t.Logf("ResolveLicenseID result: %v", result)
}
