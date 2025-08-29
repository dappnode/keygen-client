package keygen

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/joho/godotenv"
)

func mustEnv(t *testing.T, k string) string {
	t.Helper()
	v := os.Getenv(k)
	if v == "" {
		t.Fatalf("missing env %s", k)
	}
	return v
}

func TestMain(m *testing.M) {
	// Look for a .env in repo root and/or package dir
	_ = godotenv.Load("../.env", ".env")
	os.Exit(m.Run())
}

func optionalEnv(k, def string) string {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	return v
}

func randomHex(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func TestIntegration_FullFlow(t *testing.T) {
	// Don’t parallelize: keep request rate low to avoid 429s.
	ctx := context.Background()

	accountID := mustEnv(t, "KEYGEN_ACCOUNT_ID")
	apiToken := mustEnv(t, "KEYGEN_API_TOKEN")
	policyID := mustEnv(t, "KEYGEN_POLICY_ID")
	baseEmail := mustEnv(t, "KEYGEN_CUSTOMER_EMAIL")

	// Unique test data to avoid collisions on repeated runs
	suffix := time.Now().UTC().Format("20060102T150405") + "-" + randomHex(4)
	subscriptionID := "it-" + suffix
	customerEmail := strings.Replace(baseEmail, "@", "+"+suffix+"@", 1)
	fingerprint := "fp-" + suffix

	// Build client (use sandbox if provided)
	c := New(accountID, apiToken, WithDefaultMachine("dappnode", "linux"))

	// 1) Create license
	key, err := c.CreateLicense(ctx, policyID, LicenseMetadata{
		SubscriptionID: subscriptionID,
		CustomerEmail:  customerEmail,
	})
	if err != nil {
		t.Fatalf("CreateLicense: %v", err)
	}
	t.Logf("created license key: %s", key)

	// Ensure cleanup even if later assertions fail
	// resolve ID early so we can delete at the end
	licID, err := c.ResolveLicenseID(ctx, key)
	if err != nil {
		t.Fatalf("ResolveLicenseID (post-create): %v", err)
	}
	t.Cleanup(func() {
		// Best-effort cleanup (ignore errors)
		_, _ = c.DeactivateMachine(ctx, key, fingerprint)
		_ = c.DeleteLicense(ctx, licID)
	})

	// 2) Get by subscriptionId metadata
	gotID, err := c.GetLicenseBySubscriptionID(ctx, subscriptionID)
	if err != nil {
		t.Fatalf("GetLicenseBySubscriptionID: %v", err)
	}
	if gotID != licID {
		t.Fatalf("GetLicenseBySubscriptionID mismatch: want %s got %s", licID, gotID)
	}

	// 3) List licenses by policy (should include the one we created)
	lics, err := c.ListLicensesByPolicy(ctx, policyID)
	if err != nil {
		t.Fatalf("ListLicensesByPolicy: %v", err)
	}
	found := false
	for _, l := range lics {
		if l.ID == licID {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("created license %s not found in ListLicensesByPolicy", licID)
	}

	// 4) List license keys by policy (should include the key)
	keys, err := c.ListLicenseKeysByPolicy(ctx, policyID)
	if err != nil {
		t.Fatalf("ListLicenseKeysByPolicy: %v", err)
	}
	hasKey := false
	for _, k := range keys {
		if k == key {
			hasKey = true
			break
		}
	}
	if !hasKey {
		t.Fatalf("created key not found in ListLicenseKeysByPolicy")
	}

	// 5) Activate machine
	if err := c.ActivateMachine(ctx, key, fingerprint, "", ""); err != nil {
		t.Fatalf("ActivateMachine: %v", err)
	}

	// 6) Validate (scoped to fingerprint)
	val, err := c.Validate(ctx, key, fingerprint)
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if !val.Valid {
		t.Fatalf("Validate: expected valid, got %+v", val)
	}

	// 7) List machines (by license)
	ms, err := c.ListMachines(ctx, licID)
	if err != nil {
		t.Fatalf("ListMachines: %v", err)
	}
	mfound := false
	for _, m := range ms {
		if m.Fingerprint == fingerprint {
			mfound = true
			break
		}
	}
	if !mfound {
		t.Fatalf("ListMachines: did not find fingerprint %s", fingerprint)
	}

	// 8) List all machines (should also include it)
	all, err := c.ListAllMachines(ctx)
	if err != nil {
		t.Fatalf("ListAllMachines: %v", err)
	}
	afound := false
	for _, m := range all {
		if m.Fingerprint == fingerprint {
			afound = true
			break
		}
	}
	if !afound {
		t.Fatalf("ListAllMachines: did not find fingerprint %s", fingerprint)
	}

	// 9) Deactivate machine
	foundDel, err := c.DeactivateMachine(ctx, key, fingerprint)
	if err != nil {
		t.Fatalf("DeactivateMachine: %v", err)
	}
	if !foundDel {
		t.Fatalf("DeactivateMachine: machine not found for fingerprint %s", fingerprint)
	}

	// 10) Delete license
	if err := c.DeleteLicense(ctx, licID); err != nil {
		t.Fatalf("DeleteLicense: %v", err)
	}
}
