package main

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	ebaytools "github.com/repricah/ebay-tools"
)

type stubClient struct{}

func (stubClient) RefreshUserAccessToken(context.Context, []string) (*ebaytools.TokenResponse, error) {
	return &ebaytools.TokenResponse{
		AccessToken: "access-123",
		ExpiresIn:   7200,
		TokenType:   "User Access Token",
	}, nil
}

func (stubClient) GetPrivileges(context.Context, string) (*ebaytools.SellingPrivileges, error) {
	return &ebaytools.SellingPrivileges{SellerRegistrationCompleted: true}, nil
}

func (stubClient) OptInToProgram(context.Context, ebaytools.OptInToProgramRequest, string) error {
	return nil
}

func (stubClient) GetOptedInPrograms(context.Context, string) (*ebaytools.OptInProgramsResponse, error) {
	return &ebaytools.OptInProgramsResponse{
		Programs: []ebaytools.Program{
			{ProgramType: "SELLING_POLICY_MANAGEMENT", OptInStatus: "OPTED_IN"},
		},
	}, nil
}

func (stubClient) GetInventoryItem(context.Context, string, string) (*ebaytools.InventoryItem, error) {
	return &ebaytools.InventoryItem{SKU: "sku-123"}, nil
}

func (stubClient) UpsertInventoryItem(context.Context, string, ebaytools.InventoryItem, string, string) error {
	return nil
}

func (stubClient) GetInventoryLocations(context.Context, string) (*ebaytools.InventoryLocationsResponse, error) {
	return &ebaytools.InventoryLocationsResponse{
		Total: 1,
		Locations: []ebaytools.InventoryLocation{
			{MerchantLocationKey: "warehouse-1", Name: "Sandbox Warehouse"},
		},
	}, nil
}

func (stubClient) CreateInventoryLocation(context.Context, string, ebaytools.InventoryLocation, string) error {
	return nil
}

func (stubClient) GetFulfillmentPolicies(context.Context, string, string) (*ebaytools.FulfillmentPoliciesResponse, error) {
	return &ebaytools.FulfillmentPoliciesResponse{
		Total: 1,
		FulfillmentPolicies: []ebaytools.FulfillmentPolicy{
			{FulfillmentPolicyID: "fulfillment-policy-id", Name: "Default Shipping"},
		},
	}, nil
}

func (stubClient) GetPaymentPolicies(context.Context, string, string) (*ebaytools.PaymentPoliciesResponse, error) {
	return &ebaytools.PaymentPoliciesResponse{
		Total: 1,
		PaymentPolicies: []ebaytools.PaymentPolicy{
			{PaymentPolicyID: "payment-policy-id", Name: "Default Payment"},
		},
	}, nil
}

func (stubClient) GetReturnPolicies(context.Context, string, string) (*ebaytools.ReturnPoliciesResponse, error) {
	return &ebaytools.ReturnPoliciesResponse{
		Total: 1,
		ReturnPolicies: []ebaytools.ReturnPolicy{
			{ReturnPolicyID: "return-policy-id", Name: "Default Return"},
		},
	}, nil
}

func (stubClient) CreateFulfillmentPolicy(context.Context, ebaytools.FulfillmentPolicy, string) (*ebaytools.FulfillmentPolicy, error) {
	return &ebaytools.FulfillmentPolicy{FulfillmentPolicyID: "fulfillment-policy-id"}, nil
}

func (stubClient) CreatePaymentPolicy(context.Context, ebaytools.PaymentPolicy, string) (*ebaytools.PaymentPolicy, error) {
	return &ebaytools.PaymentPolicy{PaymentPolicyID: "payment-policy-id"}, nil
}

func (stubClient) CreateReturnPolicy(context.Context, ebaytools.ReturnPolicy, string) (*ebaytools.ReturnPolicy, error) {
	return &ebaytools.ReturnPolicy{ReturnPolicyID: "return-policy-id"}, nil
}

func (stubClient) GetOffers(context.Context, string, string) (*ebaytools.OffersResponse, error) {
	return &ebaytools.OffersResponse{
		Total: 1,
		Offers: []ebaytools.Offer{
			{
				OfferID:    "offer-123",
				SKU:        "sku-123",
				Status:     "UNPUBLISHED",
				Format:     "FIXED_PRICE",
				CategoryID: "183050",
			},
		},
	}, nil
}

func (stubClient) CreateOffer(context.Context, ebaytools.Offer, string) (*ebaytools.CreateOfferResponse, error) {
	return &ebaytools.CreateOfferResponse{OfferID: "offer-123"}, nil
}

func (stubClient) PublishOffer(context.Context, string, string) (*ebaytools.PublishOfferResponse, error) {
	return &ebaytools.PublishOfferResponse{ListingID: "listing-123"}, nil
}

func TestRunOfferGet(t *testing.T) {
	previous := newClient
	newClient = func(cfg ebaytools.Config) (clientAPI, error) { return stubClient{}, nil }
	t.Cleanup(func() { newClient = previous })

	stdout, stderr, err := captureRun(t, []string{"offer-get", "sku-123"})
	if err != nil {
		t.Fatalf("run: %v stderr=%s", err, stderr)
	}
	if !strings.Contains(stdout, `"offerId": "offer-123"`) {
		t.Fatalf("stdout missing offer id: %s", stdout)
	}
}

func TestRunProgramOptIn(t *testing.T) {
	type captured struct {
		request ebaytools.OptInToProgramRequest
	}
	var got captured

	previous := newClient
	newClient = func(cfg ebaytools.Config) (clientAPI, error) {
		return stubProgramClient{optInFn: func(ctx context.Context, request ebaytools.OptInToProgramRequest, accessToken string) error {
			got.request = request
			return nil
		}}, nil
	}
	t.Cleanup(func() { newClient = previous })

	stdout, stderr, err := captureRun(t, []string{"program-opt-in", "SELLING_POLICY_MANAGEMENT"})
	if err != nil {
		t.Fatalf("run: %v stderr=%s", err, stderr)
	}
	if !strings.Contains(stdout, `"status": "opt-in-requested"`) {
		t.Fatalf("stdout missing status: %s", stdout)
	}
	if got.request.ProgramType != "SELLING_POLICY_MANAGEMENT" {
		t.Fatalf("program type = %#v", got.request)
	}
}

func TestRunProgramList(t *testing.T) {
	previous := newClient
	newClient = func(cfg ebaytools.Config) (clientAPI, error) { return stubClient{}, nil }
	t.Cleanup(func() { newClient = previous })

	stdout, stderr, err := captureRun(t, []string{"program-list"})
	if err != nil {
		t.Fatalf("run: %v stderr=%s", err, stderr)
	}
	if !strings.Contains(stdout, `"programType": "SELLING_POLICY_MANAGEMENT"`) {
		t.Fatalf("stdout missing program type: %s", stdout)
	}
}

func TestRunPolicyList(t *testing.T) {
	previous := newClient
	newClient = func(cfg ebaytools.Config) (clientAPI, error) { return stubClient{}, nil }
	t.Cleanup(func() { newClient = previous })

	stdout, stderr, err := captureRun(t, []string{"policy-list", "fulfillment", "EBAY_US"})
	if err != nil {
		t.Fatalf("run: %v stderr=%s", err, stderr)
	}
	if !strings.Contains(stdout, `"fulfillmentPolicyId": "fulfillment-policy-id"`) {
		t.Fatalf("stdout missing policy id: %s", stdout)
	}
}

func TestRunPolicyCreate(t *testing.T) {
	type captured struct {
		policy ebaytools.FulfillmentPolicy
	}
	var got captured

	previous := newClient
	newClient = func(cfg ebaytools.Config) (clientAPI, error) {
		return stubPolicyCreateClient{createFulfillmentFn: func(ctx context.Context, policy ebaytools.FulfillmentPolicy, accessToken string) (*ebaytools.FulfillmentPolicy, error) {
			got.policy = policy
			return &ebaytools.FulfillmentPolicy{FulfillmentPolicyID: "fulfillment-policy-id"}, nil
		}}, nil
	}
	t.Cleanup(func() { newClient = previous })

	payloadPath := filepath.Join(t.TempDir(), "fulfillment-policy.json")
	if err := os.WriteFile(payloadPath, []byte(`{
		"name": "Default Shipping",
		"marketplaceId": "EBAY_US",
		"handlingTime": {"unit": "DAY", "value": 1}
	}`), 0600); err != nil {
		t.Fatalf("write payload: %v", err)
	}

	stdout, stderr, err := captureRun(t, []string{"policy-create", "fulfillment", payloadPath})
	if err != nil {
		t.Fatalf("run: %v stderr=%s", err, stderr)
	}
	if !strings.Contains(stdout, `"fulfillmentPolicyId": "fulfillment-policy-id"`) {
		t.Fatalf("stdout missing policy id: %s", stdout)
	}
	if got.policy.Name != "Default Shipping" || got.policy.MarketplaceID != "EBAY_US" {
		t.Fatalf("captured policy = %#v", got.policy)
	}
}

func TestRunLocationList(t *testing.T) {
	previous := newClient
	newClient = func(cfg ebaytools.Config) (clientAPI, error) { return stubClient{}, nil }
	t.Cleanup(func() { newClient = previous })

	stdout, stderr, err := captureRun(t, []string{"location-list"})
	if err != nil {
		t.Fatalf("run: %v stderr=%s", err, stderr)
	}
	if !strings.Contains(stdout, `"merchantLocationKey": "warehouse-1"`) {
		t.Fatalf("stdout missing location key: %s", stdout)
	}
}

func TestRunLocationCreate(t *testing.T) {
	type captured struct {
		key      string
		location ebaytools.InventoryLocation
	}
	var got captured

	previous := newClient
	newClient = func(cfg ebaytools.Config) (clientAPI, error) {
		return stubLocationCreateClient{createFn: func(ctx context.Context, merchantLocationKey string, location ebaytools.InventoryLocation, accessToken string) error {
			got.key = merchantLocationKey
			got.location = location
			return nil
		}}, nil
	}
	t.Cleanup(func() { newClient = previous })

	payloadPath := filepath.Join(t.TempDir(), "location.json")
	if err := os.WriteFile(payloadPath, []byte(`{
		"name": "Sandbox Warehouse",
		"locationTypes": ["WAREHOUSE"],
		"location": {
			"address": {
				"country": "US",
				"postalCode": "10001"
			}
		}
	}`), 0600); err != nil {
		t.Fatalf("write payload: %v", err)
	}

	stdout, stderr, err := captureRun(t, []string{"location-create", "warehouse-1", payloadPath})
	if err != nil {
		t.Fatalf("run: %v stderr=%s", err, stderr)
	}
	if !strings.Contains(stdout, `"status": "created"`) {
		t.Fatalf("stdout missing status: %s", stdout)
	}
	if got.key != "warehouse-1" || got.location.Name != "Sandbox Warehouse" {
		t.Fatalf("captured location = %#v", got)
	}
}

func TestRunOfferCreate(t *testing.T) {
	type captured struct {
		offer ebaytools.Offer
	}
	var got captured

	previous := newClient
	newClient = func(cfg ebaytools.Config) (clientAPI, error) {
		return stubCreateClient{createFn: func(ctx context.Context, offer ebaytools.Offer, accessToken string) (*ebaytools.CreateOfferResponse, error) {
			got.offer = offer
			return &ebaytools.CreateOfferResponse{OfferID: "offer-123"}, nil
		}}, nil
	}
	t.Cleanup(func() { newClient = previous })

	payloadPath := filepath.Join(t.TempDir(), "offer.json")
	if err := os.WriteFile(payloadPath, []byte(`{
		"sku": "sku-123",
		"marketplaceId": "EBAY_US",
		"format": "FIXED_PRICE",
		"availableQuantity": 4,
		"categoryId": "183050",
		"merchantLocationKey": "default",
		"listingDescription": "Playset of commons",
		"listingPolicies": {
			"fulfillmentPolicyId": "fulfillment-policy-id",
			"paymentPolicyId": "payment-policy-id",
			"returnPolicyId": "return-policy-id"
		},
		"pricingSummary": {
			"price": {"currency": "USD", "value": "3.99"}
		},
		"listingDuration": "GTC"
	}`), 0600); err != nil {
		t.Fatalf("write payload: %v", err)
	}

	stdout, stderr, err := captureRun(t, []string{"offer-create", payloadPath})
	if err != nil {
		t.Fatalf("run: %v stderr=%s", err, stderr)
	}
	if !strings.Contains(stdout, `"offerId": "offer-123"`) {
		t.Fatalf("stdout missing offer id: %s", stdout)
	}
	if got.offer.SKU != "sku-123" || got.offer.MarketplaceID != "EBAY_US" {
		t.Fatalf("captured offer = %#v", got.offer)
	}
}

func TestRunOfferPublish(t *testing.T) {
	previous := newClient
	newClient = func(cfg ebaytools.Config) (clientAPI, error) { return stubClient{}, nil }
	t.Cleanup(func() { newClient = previous })

	stdout, stderr, err := captureRun(t, []string{"offer-publish", "offer-123"})
	if err != nil {
		t.Fatalf("run: %v stderr=%s", err, stderr)
	}
	if !strings.Contains(stdout, `"listingId": "listing-123"`) {
		t.Fatalf("stdout missing listing id: %s", stdout)
	}
}

type stubCreateClient struct {
	stubClient
	createFn func(context.Context, ebaytools.Offer, string) (*ebaytools.CreateOfferResponse, error)
}

func (c stubCreateClient) CreateOffer(ctx context.Context, offer ebaytools.Offer, accessToken string) (*ebaytools.CreateOfferResponse, error) {
	return c.createFn(ctx, offer, accessToken)
}

type stubProgramClient struct {
	stubClient
	optInFn func(context.Context, ebaytools.OptInToProgramRequest, string) error
}

func (c stubProgramClient) OptInToProgram(ctx context.Context, request ebaytools.OptInToProgramRequest, accessToken string) error {
	return c.optInFn(ctx, request, accessToken)
}

type stubPolicyCreateClient struct {
	stubClient
	createFulfillmentFn func(context.Context, ebaytools.FulfillmentPolicy, string) (*ebaytools.FulfillmentPolicy, error)
}

func (c stubPolicyCreateClient) CreateFulfillmentPolicy(ctx context.Context, policy ebaytools.FulfillmentPolicy, accessToken string) (*ebaytools.FulfillmentPolicy, error) {
	return c.createFulfillmentFn(ctx, policy, accessToken)
}

type stubLocationCreateClient struct {
	stubClient
	createFn func(context.Context, string, ebaytools.InventoryLocation, string) error
}

func (c stubLocationCreateClient) CreateInventoryLocation(ctx context.Context, merchantLocationKey string, location ebaytools.InventoryLocation, accessToken string) error {
	return c.createFn(ctx, merchantLocationKey, location, accessToken)
}

func captureRun(t *testing.T, args []string) (string, string, error) {
	t.Helper()

	oldOSStdout := os.Stdout
	oldOSStderr := os.Stderr
	oldStdout := stdout
	oldStderr := stderr
	oldEnv := map[string]string{
		"EBAY_API_BASE_URL":    os.Getenv("EBAY_API_BASE_URL"),
		"EBAY_OAUTH_TOKEN_URL": os.Getenv("EBAY_OAUTH_TOKEN_URL"),
		"EBAY_APP_ID":          os.Getenv("EBAY_APP_ID"),
		"EBAY_CERT_ID":         os.Getenv("EBAY_CERT_ID"),
		"EBAY_REFRESH_TOKEN":   os.Getenv("EBAY_REFRESH_TOKEN"),
	}
	t.Cleanup(func() {
		os.Stdout = oldOSStdout
		os.Stderr = oldOSStderr
		stdout = oldStdout
		stderr = oldStderr
		for k, v := range oldEnv {
			if v == "" {
				_ = os.Unsetenv(k)
			} else {
				_ = os.Setenv(k, v)
			}
		}
	})

	for _, kv := range []struct{ k, v string }{
		{"EBAY_API_BASE_URL", "https://api.sandbox.ebay.test"},
		{"EBAY_OAUTH_TOKEN_URL", "https://api.sandbox.ebay.test/identity/v1/oauth2/token"},
		{"EBAY_APP_ID", "app-id"},
		{"EBAY_CERT_ID", "cert-id"},
		{"EBAY_REFRESH_TOKEN", "refresh-123"},
	} {
		if err := os.Setenv(kv.k, kv.v); err != nil {
			t.Fatalf("setenv %s: %v", kv.k, err)
		}
	}

	outR, outW, _ := os.Pipe()
	errR, errW, _ := os.Pipe()
	os.Stdout = outW
	os.Stderr = errW
	stdout = outW
	stderr = errW

	runErr := run(context.Background(), args)

	_ = outW.Close()
	_ = errW.Close()

	var stdout, stderr bytes.Buffer
	_, _ = io.Copy(&stdout, outR)
	_, _ = io.Copy(&stderr, errR)

	return stdout.String(), stderr.String(), runErr
}
