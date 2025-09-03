package keygen

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"
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
	otherPolicyID := mustEnv(t, "KEYGEN_POLICY_ID2") // now mandatory
	baseEmail := mustEnv(t, "KEYGEN_CUSTOMER_EMAIL")

	// Unique test data for each license
	suffix1 := time.Now().UTC().Format("20060102T150405") + "-" + randomHex(4)
	suffix2 := time.Now().UTC().Format("20060102T150405") + "-" + randomHex(4)
	suffix3 := time.Now().UTC().Format("20060102T150405") + "-" + randomHex(4)
	subscriptionID1 := "it1-" + suffix1
	subscriptionID2 := "it2-" + suffix2
	subscriptionID3 := "it3-" + suffix3
	customerEmail1 := strings.Replace(baseEmail, "@", "+"+suffix1+"@", 1)
	customerEmail2 := strings.Replace(baseEmail, "@", "+"+suffix2+"@", 1)
	customerEmail3 := strings.Replace(baseEmail, "@", "+"+suffix3+"@", 1)
	fingerprint1 := "fp1-" + suffix1
	fingerprint2 := "fp2-" + suffix2
	fingerprint3 := "fp3-" + suffix3

	t.Logf("Test setup: accountID=%s, policyID=%s, otherPolicyID=%s", accountID, policyID, otherPolicyID)

	c := New(accountID, apiToken, WithDefaultMachine("dappnode", "linux"))

	// --- Create licenses ---
	t.Logf("Creating license1 with policyID=%s, metadata={SubscriptionID:%s, CustomerEmail:%s}", policyID, subscriptionID1, customerEmail1)
	key1, err := c.CreateLicense(ctx, policyID, LicenseMetadata{
		SubscriptionID: subscriptionID1,
		CustomerEmail:  customerEmail1,
	})
	if err != nil {
		t.Fatalf("CreateLicense1: %v", err)
	}
	t.Logf("created license1 key: %s", key1)

	t.Logf("Creating license2 with policyID=%s, metadata={SubscriptionID:%s, CustomerEmail:%s}", policyID, subscriptionID2, customerEmail2)
	key2, err := c.CreateLicense(ctx, policyID, LicenseMetadata{
		SubscriptionID: subscriptionID2,
		CustomerEmail:  customerEmail2,
	})
	if err != nil {
		t.Fatalf("CreateLicense2: %v", err)
	}
	t.Logf("created license2 key: %s", key2)

	t.Logf("Creating license3 with otherPolicyID=%s, metadata={SubscriptionID:%s, CustomerEmail:%s}", otherPolicyID, subscriptionID3, customerEmail3)
	key3, err := c.CreateLicense(ctx, otherPolicyID, LicenseMetadata{
		SubscriptionID: subscriptionID3,
		CustomerEmail:  customerEmail3,
	})
	if err != nil {
		t.Fatalf("CreateLicense3: %v", err)
	}
	t.Logf("created license3 key: %s", key3)

	// --- Resolve license IDs ---
	t.Logf("Resolving license IDs from keys")
	licID1, err := c.ResolveLicenseID(ctx, key1)
	if err != nil {
		t.Fatalf("ResolveLicenseID1: %v", err)
	}
	licID2, err := c.ResolveLicenseID(ctx, key2)
	if err != nil {
		t.Fatalf("ResolveLicenseID2: %v", err)
	}
	licID3, err := c.ResolveLicenseID(ctx, key3)
	if err != nil {
		t.Fatalf("ResolveLicenseID3: %v", err)
	}

	t.Cleanup(func() {
		t.Logf("Cleanup: Deactivating machines and deleting licenses")
		_, _ = c.DeactivateMachine(ctx, key1, fingerprint1)
		_, _ = c.DeactivateMachine(ctx, key2, fingerprint2)
		_, _ = c.DeactivateMachine(ctx, key3, fingerprint3)
		_ = c.DeleteLicense(ctx, licID1)
		_ = c.DeleteLicense(ctx, licID2)
		_ = c.DeleteLicense(ctx, licID3)
	})

	// --- Activate machines for each license ---
	t.Logf("Activating machine1: licenseKey=%s, fingerprint=%s", key1, fingerprint1)
	if err := c.ActivateMachine(ctx, key1, fingerprint1, "", ""); err != nil {
		t.Fatalf("ActivateMachine1: %v", err)
	}
	t.Logf("Activating machine2: licenseKey=%s, fingerprint=%s", key2, fingerprint2)
	if err := c.ActivateMachine(ctx, key2, fingerprint2, "", ""); err != nil {
		t.Fatalf("ActivateMachine2: %v", err)
	}
	t.Logf("Activating machine3: licenseKey=%s, fingerprint=%s", key3, fingerprint3)
	if err := c.ActivateMachine(ctx, key3, fingerprint3, "", ""); err != nil {
		t.Fatalf("ActivateMachine3: %v", err)
	}

	// --- Validate each license with correct and incorrect fingerprints ---
	t.Logf("Validating license1 with correct fingerprint")
	val, err := c.Validate(ctx, key1, fingerprint1)
	if err != nil || !val.Valid {
		t.Fatalf("Validate license1 (correct fp): %v %+v", err, val)
	}
	t.Logf("Validating license1 with license2's fingerprint (should be invalid)")
	val, err = c.Validate(ctx, key1, fingerprint2)
	if err != nil {
		t.Fatalf("Validate license1 (wrong fp): %v", err)
	}
	if val.Valid {
		t.Fatalf("Validate license1 (wrong fp): expected invalid, got %+v", val)
	}
	t.Logf("Validating license1 with license3's fingerprint (should be invalid)")
	val, err = c.Validate(ctx, key1, fingerprint3)
	if err != nil {
		t.Fatalf("Validate license1 (wrong fp3): %v", err)
	}
	if val.Valid {
		t.Fatalf("Validate license1 (wrong fp3): expected invalid, got %+v", val)
	}
	t.Logf("Validating license3 with license1's fingerprint (should be invalid)")
	val, err = c.Validate(ctx, key3, fingerprint1)
	if err != nil {
		t.Fatalf("Validate license3 (wrong fp): %v", err)
	}
	if val.Valid {
		t.Fatalf("Validate license3 (wrong fp): expected invalid, got %+v", val)
	}

	// --- Additional: Validate with a random (never activated) fingerprint ---
	randomFP := "fp-random-" + randomHex(8)
	t.Logf("Validating license1 with a random, never-activated fingerprint: %s", randomFP)
	val, err = c.Validate(ctx, key1, randomFP)
	if err != nil {
		t.Fatalf("Validate license1 (random fp): %v", err)
	}
	if val.Valid {
		t.Fatalf("Validate license1 (random fp): expected invalid, got %+v", val)
	}

	// --- Additional: Validate a newly created license (not yet activated) ---
	t.Logf("Creating a new license4 (not activating any machine)")
	key4, err := c.CreateLicense(ctx, policyID, LicenseMetadata{
		SubscriptionID: "it4-" + randomHex(8),
		CustomerEmail:  strings.Replace(baseEmail, "@", "+notactivated@", 1),
	})
	if err != nil {
		t.Fatalf("CreateLicense4: %v", err)
	}
	t.Logf("created license4 key: %s", key4)
	// Try to validate with a fingerprint that has not been activated
	notActivatedFP := "fp-not-activated-" + randomHex(8)
	t.Logf("Validating license4 (not activated) with fingerprint: %s", notActivatedFP)
	val, err = c.Validate(ctx, key4, notActivatedFP)
	if err != nil {
		t.Fatalf("Validate license4 (not activated): %v", err)
	}
	if val.Valid {
		t.Fatalf("Validate license4 (not activated): expected invalid, got %+v", val)
	}
	// Cleanup license4
	licID4, _ := c.ResolveLicenseID(ctx, key4)
	_ = c.DeleteLicense(ctx, licID4)

	// --- List licenses by policy and check all are present and isolated ---
	t.Logf("Listing licenses by policyID: %s", policyID)
	lics, err := c.ListLicensesByPolicy(ctx, policyID)
	if err != nil {
		t.Fatalf("ListLicensesByPolicy: %v", err)
	}
	fmt.Println(lics)
	var found1, found2, found3 bool
	for _, l := range lics {
		if l.ID == licID1 {
			found1 = true
		}
		if l.ID == licID2 {
			found2 = true
		}
		if l.ID == licID3 {
			found3 = true
		}
	}
	if !found1 || !found2 {
		t.Fatalf("Did not find both licenses in ListLicensesByPolicy: found1=%v, found2=%v", found1, found2)
	}
	if found3 {
		t.Fatalf("license3 (other policy) should not be in policyID list, got found3=%v", found3)
	}

	t.Logf("Listing licenses by otherPolicyID: %s", otherPolicyID)
	lics3, err := c.ListLicensesByPolicy(ctx, otherPolicyID)
	if err != nil {
		t.Fatalf("ListLicensesByPolicy (other): %v", err)
	}
	found3 = false
	for _, l := range lics3 {
		if l.ID == licID3 {
			found3 = true
		}
		if l.ID == licID1 || l.ID == licID2 {
			t.Fatalf("license1 or license2 should not be in otherPolicyID list")
		}
	}
	if !found3 {
		t.Fatalf("Did not find license3 in ListLicensesByPolicy (other)")
	}

	// --- Try to activate a machine with the wrong license (should fail) ---
	t.Logf("Attempting to activate machine1 with license2's key (should fail)")
	err = c.ActivateMachine(ctx, key2, fingerprint1, "", "")
	if err == nil {
		t.Fatalf("Expected error activating machine1 with license2's key, got none")
	} else {
		t.Logf("Got expected error: %v", err)
	}
	t.Logf("Attempting to activate machine3 with license1's key (should fail)")
	err = c.ActivateMachine(ctx, key1, fingerprint3, "", "")
	if err == nil {
		t.Fatalf("Expected error activating machine3 with license1's key, got none")
	} else {
		t.Logf("Got expected error: %v", err)
	}

	// --- Edge: Try to get license by subscriptionID for all three ---
	t.Logf("Getting license by subscriptionID1: %s", subscriptionID1)
	gotID, err := c.GetLicenseBySubscriptionID(ctx, subscriptionID1)
	if err != nil {
		t.Fatalf("GetLicenseBySubscriptionID1: %v", err)
	}
	if gotID != licID1 {
		t.Fatalf("GetLicenseBySubscriptionID1 mismatch: want %s got %s", licID1, gotID)
	}
	t.Logf("Getting license by subscriptionID2: %s", subscriptionID2)
	gotID, err = c.GetLicenseBySubscriptionID(ctx, subscriptionID2)
	if err != nil {
		t.Fatalf("GetLicenseBySubscriptionID2: %v", err)
	}
	if gotID != licID2 {
		t.Fatalf("GetLicenseBySubscriptionID2 mismatch: want %s got %s", licID2, gotID)
	}
	t.Logf("Getting license by subscriptionID3: %s", subscriptionID3)
	gotID, err = c.GetLicenseBySubscriptionID(ctx, subscriptionID3)
	if err != nil {
		t.Fatalf("GetLicenseBySubscriptionID3: %v", err)
	}
	if gotID != licID3 {
		t.Fatalf("GetLicenseBySubscriptionID3 mismatch: want %s got %s", licID3, gotID)
	}

	// --- Edge: List machines for each license, ensure isolation ---
	t.Logf("Listing machines by licenseID1: %s", licID1)
	ms1, err := c.ListMachines(ctx, licID1)
	if err != nil {
		t.Fatalf("ListMachines1: %v", err)
	}
	if len(ms1) != 1 || ms1[0].Fingerprint != fingerprint1 {
		t.Fatalf("ListMachines1: expected 1 machine with fingerprint1, got %+v", ms1)
	}
	// Ensure all machines in ms1 belong to fingerprint1
	for _, m := range ms1 {
		if m.Fingerprint != fingerprint1 {
			t.Fatalf("ListMachines1: found unexpected machine %+v", m)
		}
	}

	t.Logf("Listing machines by licenseID2: %s", licID2)
	ms2, err := c.ListMachines(ctx, licID2)
	if err != nil {
		t.Fatalf("ListMachines2: %v", err)
	}
	if len(ms2) != 1 || ms2[0].Fingerprint != fingerprint2 {
		t.Fatalf("ListMachines2: expected 1 machine with fingerprint2, got %+v", ms2)
	}
	for _, m := range ms2 {
		if m.Fingerprint != fingerprint2 {
			t.Fatalf("ListMachines2: found unexpected machine %+v", m)
		}
	}

	t.Logf("Listing machines by licenseID3: %s", licID3)
	ms3, err := c.ListMachines(ctx, licID3)
	if err != nil {
		t.Fatalf("ListMachines3: %v", err)
	}
	if len(ms3) != 1 || ms3[0].Fingerprint != fingerprint3 {
		t.Fatalf("ListMachines3: expected 1 machine with fingerprint3, got %+v", ms3)
	}
	for _, m := range ms3 {
		if m.Fingerprint != fingerprint3 {
			t.Fatalf("ListMachines3: found unexpected machine %+v", m)
		}
	}

	// --- List all machines and check all are present ---
	t.Logf("Listing all machines")
	allMachines, err := c.ListAllMachines(ctx)
	if err != nil {
		t.Fatalf("ListAllMachines: %v", err)
	}
	// Build a map of fingerprint -> licenseID for quick lookup
	expected := map[string]string{
		fingerprint1: licID1,
		fingerprint2: licID2,
		fingerprint3: licID3,
	}
	found := map[string]bool{
		fingerprint1: false,
		fingerprint2: false,
		fingerprint3: false,
	}
	for _, m := range allMachines {
		lic, ok := expected[m.Fingerprint]
		if ok {
			// Only check LicenseId if present in Machine struct
			if m.LicenseId != "" && m.LicenseId != lic {
				t.Fatalf("ListAllMachines: machine %+v has wrong licenseId, want %s", m, lic)
			}
			found[m.Fingerprint] = true
		}
	}
	for fp, ok := range found {
		if !ok {
			t.Fatalf("ListAllMachines: did not find machine with fingerprint %s", fp)
		}
	}

	// --- Delete license1, ensure license2 and license3 are unaffected ---
	t.Logf("Deleting license1: %s", licID1)
	if err := c.DeleteLicense(ctx, licID1); err != nil {
		t.Fatalf("DeleteLicense1: %v", err)
	}
	t.Logf("Ensuring license2 is still valid after deleting license1")
	val, err = c.Validate(ctx, key2, fingerprint2)
	if err != nil || !val.Valid {
		t.Fatalf("Validate license2 after deleting license1: %v %+v", err, val)
	}
	t.Logf("Ensuring license3 is still valid after deleting license1")
	val, err = c.Validate(ctx, key3, fingerprint3)
	if err != nil || !val.Valid {
		t.Fatalf("Validate license3 after deleting license1: %v %+v", err, val)
	}

	// --- Ensure license1 is gone ---
	t.Logf("Ensuring deleted license1 cannot be resolved")
	_, err = c.ResolveLicenseID(ctx, key1)
	if err == nil {
		t.Fatalf("Expected error resolving deleted license1")
	}

	// --- Edge: Try to deactivate machine1 again when its license has already been deleted
	t.Logf("Deactivating machine1 again (should be no-op)")
	foundAgain, err := c.DeactivateMachine(ctx, key1, fingerprint1)
	if err != nil {
		// We expect "no license id found" as a valid error when license is already deleted
		if !strings.Contains(err.Error(), "no license id found") {
			t.Fatalf("DeactivateMachine1 again: %v", err)
		} else {
			t.Logf("DeactivateMachine1 again: got expected error after license deletion: %v", err)
		}
	} else if foundAgain {
		t.Fatalf("DeactivateMachine1 again: expected not found")
	}

	// --- Clean up license2 and license3 ---
	t.Logf("Deleting license2: %s", licID2)
	if err := c.DeleteLicense(ctx, licID2); err != nil {
		t.Fatalf("DeleteLicense2: %v", err)
	}
	t.Logf("Deleting license3: %s", licID3)
	if err := c.DeleteLicense(ctx, licID3); err != nil {
		t.Fatalf("DeleteLicense3: %v", err)
	}

	// --- Pagination tests ---
	// IMPORTANT: Pagination tests are not 100% reliable unless we have >100 items to ensure multiple pages.
	// We are mocking the request directly to test pagination structure. We are not calling the client methods
	// because they automatically paginate internally and return all results.

	// --- Pagination test for ListLicensesByPolicy ---
	t.Logf("Testing pagination for ListLicensesByPolicy")
	const extraLicenses = 3 // create a few extra licenses to ensure multiple pages
	var extraKeys []string
	for i := 0; i < extraLicenses; i++ {
		key, err := c.CreateLicense(ctx, policyID, LicenseMetadata{
			SubscriptionID: fmt.Sprintf("it-extra-%d-%s", i, randomHex(4)),
			CustomerEmail:  strings.Replace(baseEmail, "@", fmt.Sprintf("+extra%d@", i), 1),
		})
		if err != nil {
			t.Fatalf("CreateLicense extra %d: %v", i, err)
		}
		extraKeys = append(extraKeys, key)
	}
	// Clean up extra licenses
	defer func() {
		for _, key := range extraKeys {
			licID, _ := c.ResolveLicenseID(ctx, key)
			_ = c.DeleteLicense(ctx, licID)
		}
	}()

	// Fetch first page with small page size
	q := url.Values{}
	q.Set("policy", policyID)
	q.Set("page[number]", "1")
	q.Set("page[size]", "2")
	path := fmt.Sprintf("/accounts/%s/licenses?%s", accountID, q.Encode())
	var respPage1 struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
		Links struct {
			Next *string `json:"next"`
		} `json:"links"`
	}
	if err := c.do(ctx, http.MethodGet, path, nil, &respPage1); err != nil {
		t.Fatalf("Pagination: fetch page 1: %v", err)
	}
	if len(respPage1.Data) != 2 {
		t.Fatalf("Pagination: expected 2 licenses on page 1, got %d", len(respPage1.Data))
	}
	if respPage1.Links.Next == nil {
		t.Fatalf("Pagination: expected next page link, got nil")
	}
	t.Logf("Pagination: page 1 OK, next: %s", *respPage1.Links.Next)

	// Fetch second page
	q.Set("page[number]", "2")
	path = fmt.Sprintf("/accounts/%s/licenses?%s", accountID, q.Encode())
	var respPage2 struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
		Links struct {
			Next *string `json:"next"`
		} `json:"links"`
	}
	if err := c.do(ctx, http.MethodGet, path, nil, &respPage2); err != nil {
		t.Fatalf("Pagination: fetch page 2: %v", err)
	}
	if len(respPage2.Data) != 1 {
		t.Fatalf("Pagination: expected 1 licenses on page 2, got %d", len(respPage2.Data))
	}
	t.Logf("Pagination: page 2 OK, licenses: %v", respPage2.Data)

	// test that next page is nil once we have read two pages of size 2 (we created 3 licenses above)
	if respPage2.Links.Next == nil {
		t.Logf("Pagination: reached end of pages after page 2 as expected")
	} else {
		// error
		t.Fatalf("Pagination: expected no next page after reading all licenses, got %s", *respPage2.Links.Next)
	}

	// --- Pagination test for ListAllMachines ---
	t.Logf("Testing pagination for ListAllMachines")
	q = url.Values{}
	q.Set("page[number]", "1")
	q.Set("page[size]", "2")
	path = fmt.Sprintf("/accounts/%s/machines?%s", accountID, q.Encode())
	var respMachinesPage1 struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
		Links struct {
			Next *string `json:"next"`
		} `json:"links"`
	}
	if err := c.do(ctx, http.MethodGet, path, nil, &respMachinesPage1); err != nil {
		t.Fatalf("Pagination: fetch machines page 1: %v", err)
	}
	if len(respMachinesPage1.Data) != 2 && len(respMachinesPage1.Data) != 0 {
		// Accept 0 if there are not enough machines, otherwise expect 2
		t.Fatalf("Pagination: expected 2 machines on page 1 (or 0 if not enough), got %d", len(respMachinesPage1.Data))
	}
	if len(respMachinesPage1.Data) == 2 && respMachinesPage1.Links.Next == nil {
		t.Fatalf("Pagination: expected next page link for machines, got nil")
	}
	t.Logf("Pagination: machines page 1 OK, next: %v", respMachinesPage1.Links.Next)
}
