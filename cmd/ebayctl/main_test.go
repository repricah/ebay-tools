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

func (stubClient) GetInventoryItem(context.Context, string, string) (*ebaytools.InventoryItem, error) {
	return &ebaytools.InventoryItem{SKU: "sku-123"}, nil
}

func (stubClient) UpsertInventoryItem(context.Context, string, ebaytools.InventoryItem, string, string) error {
	return nil
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
