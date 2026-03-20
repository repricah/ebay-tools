package ebaytools

import (
	"context"
	"encoding/base64"
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

func TestRefreshUserAccessToken(t *testing.T) {
	t.Parallel()

	var capturedAuth string
	var capturedBody string
	httpClient := &http.Client{
		Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			if r.Method != http.MethodPost {
				t.Fatalf("method = %q, want %q", r.Method, http.MethodPost)
			}
			if got := r.Header.Get("Content-Type"); got != formContentType {
				t.Fatalf("content-type = %q, want %q", got, formContentType)
			}
			capturedAuth = r.Header.Get("Authorization")

			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("read body: %v", err)
			}
			capturedBody = string(body)

			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(strings.NewReader(`{"access_token":"access-123","expires_in":7200,"token_type":"User Access Token"}`)),
			}, nil
		}),
	}

	client, err := NewClient(Config{
		APIBaseURL:    "https://api.sandbox.ebay.test",
		OAuthTokenURL: "https://api.sandbox.ebay.test/identity/v1/oauth2/token",
		AppID:         "app-id",
		CertID:        "cert-id",
		RefreshToken:  "refresh-123",
	}, WithHTTPClient(httpClient))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	token, err := client.RefreshUserAccessToken(context.Background(), []string{DefaultReadonlyScope()})
	if err != nil {
		t.Fatalf("RefreshUserAccessToken: %v", err)
	}
	if token.AccessToken != "access-123" {
		t.Fatalf("access token = %q, want access-123", token.AccessToken)
	}
	if token.ExpiresIn != 7200 {
		t.Fatalf("expires_in = %d, want 7200", token.ExpiresIn)
	}
	if token.TokenType != "User Access Token" {
		t.Fatalf("token type = %q, want User Access Token", token.TokenType)
	}
	if capturedAuth != "Basic "+base64.StdEncoding.EncodeToString([]byte("app-id:cert-id")) {
		t.Fatalf("authorization = %q", capturedAuth)
	}
	if !strings.Contains(capturedBody, "grant_type=refresh_token") {
		t.Fatalf("body missing grant_type: %q", capturedBody)
	}
	if !strings.Contains(capturedBody, "refresh_token=refresh-123") {
		t.Fatalf("body missing refresh token: %q", capturedBody)
	}
}

func TestGetPrivileges(t *testing.T) {
	t.Parallel()

	var capturedAuthorization string
	httpClient := &http.Client{
		Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			if r.Method != http.MethodGet {
				t.Fatalf("method = %q, want %q", r.Method, http.MethodGet)
			}
			if r.URL.Path != getPrivilegesPath {
				t.Fatalf("path = %q, want %q", r.URL.Path, getPrivilegesPath)
			}
			capturedAuthorization = r.Header.Get("Authorization")

			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(strings.NewReader(`{"sellerRegistrationCompleted":true,"sellingLimit":{"amount":{"currency":"USD","value":"100.0"},"quantity":10}}`)),
			}, nil
		}),
	}

	client, err := NewClient(Config{
		APIBaseURL:    "https://api.sandbox.ebay.test",
		OAuthTokenURL: "https://api.sandbox.ebay.test/identity/v1/oauth2/token",
		AppID:         "app-id",
		CertID:        "cert-id",
		RefreshToken:  "refresh-123",
	}, WithHTTPClient(httpClient))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	privileges, err := client.GetPrivileges(context.Background(), "access-123")
	if err != nil {
		t.Fatalf("GetPrivileges: %v", err)
	}
	if capturedAuthorization != "Bearer access-123" {
		t.Fatalf("authorization = %q", capturedAuthorization)
	}
	if !privileges.SellerRegistrationCompleted {
		t.Fatalf("seller registration should be true")
	}
	if privileges.SellingLimit == nil || privileges.SellingLimit.Amount == nil {
		t.Fatalf("selling limit missing: %#v", privileges.SellingLimit)
	}
	if privileges.SellingLimit.Amount.Currency != "USD" {
		t.Fatalf("currency = %q, want USD", privileges.SellingLimit.Amount.Currency)
	}
	if privileges.SellingLimit.Quantity != 10 {
		t.Fatalf("quantity = %d, want 10", privileges.SellingLimit.Quantity)
	}
}

func TestGetInventoryItem(t *testing.T) {
	t.Parallel()

	httpClient := &http.Client{
		Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			if r.Method != http.MethodGet {
				t.Fatalf("method = %q, want %q", r.Method, http.MethodGet)
			}
			if r.URL.EscapedPath() != "/sell/inventory/v1/inventory_item/sku%2F123" {
				t.Fatalf("escaped path = %q", r.URL.EscapedPath())
			}
			if got := r.Header.Get("Authorization"); got != "Bearer access-123" {
				t.Fatalf("authorization = %q", got)
			}

			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body: io.NopCloser(strings.NewReader(`{
					"sku":"sku/123",
					"condition":"NEW",
					"availability":{"shipToLocationAvailability":{"quantity":4}},
					"product":{"title":"Black Lotus","imageUrls":["https://example.test/lotus.jpg"]}
				}`)),
			}, nil
		}),
	}

	client, err := NewClient(Config{
		APIBaseURL:    "https://api.sandbox.ebay.test",
		OAuthTokenURL: "https://api.sandbox.ebay.test/identity/v1/oauth2/token",
		AppID:         "app-id",
		CertID:        "cert-id",
		RefreshToken:  "refresh-123",
	}, WithHTTPClient(httpClient))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	item, err := client.GetInventoryItem(context.Background(), "sku/123", "access-123")
	if err != nil {
		t.Fatalf("GetInventoryItem: %v", err)
	}
	if item.SKU != "sku/123" {
		t.Fatalf("sku = %q", item.SKU)
	}
	if item.Product == nil || item.Product.Title != "Black Lotus" {
		t.Fatalf("product = %#v", item.Product)
	}
	if item.Availability == nil || item.Availability.ShipToLocationAvailability == nil || item.Availability.ShipToLocationAvailability.Quantity != 4 {
		t.Fatalf("availability = %#v", item.Availability)
	}
}

func TestUpsertInventoryItem(t *testing.T) {
	t.Parallel()

	var capturedBody string
	httpClient := &http.Client{
		Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			if r.Method != http.MethodPut {
				t.Fatalf("method = %q, want %q", r.Method, http.MethodPut)
			}
			if r.URL.EscapedPath() != "/sell/inventory/v1/inventory_item/sku-123" {
				t.Fatalf("escaped path = %q", r.URL.EscapedPath())
			}
			if got := r.Header.Get("Authorization"); got != "Bearer access-123" {
				t.Fatalf("authorization = %q", got)
			}
			if got := r.Header.Get("Content-Language"); got != "en-US" {
				t.Fatalf("content-language = %q", got)
			}
			if got := r.Header.Get("Content-Type"); got != jsonContentType {
				t.Fatalf("content-type = %q", got)
			}

			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("read body: %v", err)
			}
			capturedBody = string(body)

			return &http.Response{
				StatusCode: http.StatusNoContent,
				Body:       io.NopCloser(strings.NewReader("")),
			}, nil
		}),
	}

	client, err := NewClient(Config{
		APIBaseURL:    "https://api.sandbox.ebay.test",
		OAuthTokenURL: "https://api.sandbox.ebay.test/identity/v1/oauth2/token",
		AppID:         "app-id",
		CertID:        "cert-id",
		RefreshToken:  "refresh-123",
	}, WithHTTPClient(httpClient))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	err = client.UpsertInventoryItem(context.Background(), "sku-123", InventoryItem{
		Condition: "NEW",
		Product: &Product{
			Title: "Black Lotus",
		},
		Availability: &Availability{
			ShipToLocationAvailability: &ShipToLocationAvailability{Quantity: 1},
		},
	}, "access-123", "en-US")
	if err != nil {
		t.Fatalf("UpsertInventoryItem: %v", err)
	}
	if !strings.Contains(capturedBody, `"title":"Black Lotus"`) {
		t.Fatalf("body missing title: %q", capturedBody)
	}
	if !strings.Contains(capturedBody, `"quantity":1`) {
		t.Fatalf("body missing quantity: %q", capturedBody)
	}
}

func TestGetOffers(t *testing.T) {
	t.Parallel()

	httpClient := &http.Client{
		Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			if r.Method != http.MethodGet {
				t.Fatalf("method = %q, want %q", r.Method, http.MethodGet)
			}
			if r.URL.Path != offerPath {
				t.Fatalf("path = %q, want %q", r.URL.Path, offerPath)
			}
			if got := r.URL.Query().Get("sku"); got != "sku-123" {
				t.Fatalf("sku query = %q", got)
			}
			if got := r.Header.Get("Authorization"); got != "Bearer access-123" {
				t.Fatalf("authorization = %q", got)
			}

			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body: io.NopCloser(strings.NewReader(`{
					"href":"https://api.sandbox.ebay.test/sell/inventory/v1/offer?sku=sku-123",
					"total":1,
					"size":1,
					"offers":[{
						"offerId":"offer-123",
						"sku":"sku-123",
						"marketplaceId":"EBAY_US",
						"format":"FIXED_PRICE",
						"availableQuantity":4,
						"status":"UNPUBLISHED",
						"pricingSummary":{"price":{"currency":"USD","value":"19.99"}}
					}]
				}`)),
			}, nil
		}),
	}

	client, err := NewClient(Config{
		APIBaseURL:    "https://api.sandbox.ebay.test",
		OAuthTokenURL: "https://api.sandbox.ebay.test/identity/v1/oauth2/token",
		AppID:         "app-id",
		CertID:        "cert-id",
		RefreshToken:  "refresh-123",
	}, WithHTTPClient(httpClient))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	result, err := client.GetOffers(context.Background(), "sku-123", "access-123")
	if err != nil {
		t.Fatalf("GetOffers: %v", err)
	}
	if result.Total != 1 || len(result.Offers) != 1 {
		t.Fatalf("unexpected offers result: %#v", result)
	}
	if result.Offers[0].OfferID != "offer-123" {
		t.Fatalf("offer id = %q", result.Offers[0].OfferID)
	}
	if result.Offers[0].PricingSummary == nil || result.Offers[0].PricingSummary.Price == nil || result.Offers[0].PricingSummary.Price.Value != "19.99" {
		t.Fatalf("pricing = %#v", result.Offers[0].PricingSummary)
	}
}

func TestCreateOffer(t *testing.T) {
	t.Parallel()

	var capturedBody string
	httpClient := &http.Client{
		Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			if r.Method != http.MethodPost {
				t.Fatalf("method = %q, want %q", r.Method, http.MethodPost)
			}
			if r.URL.Path != offerPath {
				t.Fatalf("path = %q, want %q", r.URL.Path, offerPath)
			}
			if got := r.Header.Get("Authorization"); got != "Bearer access-123" {
				t.Fatalf("authorization = %q", got)
			}
			if got := r.Header.Get("Content-Type"); got != jsonContentType {
				t.Fatalf("content-type = %q", got)
			}

			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("read body: %v", err)
			}
			capturedBody = string(body)

			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(strings.NewReader(`{"offerId":"offer-123"}`)),
			}, nil
		}),
	}

	client, err := NewClient(Config{
		APIBaseURL:    "https://api.sandbox.ebay.test",
		OAuthTokenURL: "https://api.sandbox.ebay.test/identity/v1/oauth2/token",
		AppID:         "app-id",
		CertID:        "cert-id",
		RefreshToken:  "refresh-123",
	}, WithHTTPClient(httpClient))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	response, err := client.CreateOffer(context.Background(), Offer{
		SKU:                "sku-123",
		MarketplaceID:      "EBAY_US",
		Format:             "FIXED_PRICE",
		AvailableQuantity:  4,
		CategoryID:         "183050",
		ListingDescription: "Demo listing",
		ListingPolicies: &ListingPolicies{
			FulfillmentPolicyID: "fulfillment-policy-id",
			PaymentPolicyID:     "payment-policy-id",
			ReturnPolicyID:      "return-policy-id",
		},
		PricingSummary: &PricingSummary{
			Price: &Amount{Currency: "USD", Value: "19.99"},
		},
	}, "access-123")
	if err != nil {
		t.Fatalf("CreateOffer: %v", err)
	}
	if response.OfferID != "offer-123" {
		t.Fatalf("offer id = %q", response.OfferID)
	}
	if !strings.Contains(capturedBody, `"sku":"sku-123"`) {
		t.Fatalf("body missing sku: %q", capturedBody)
	}
	if !strings.Contains(capturedBody, `"marketplaceId":"EBAY_US"`) {
		t.Fatalf("body missing marketplace: %q", capturedBody)
	}
	if !strings.Contains(capturedBody, `"fulfillmentPolicyId":"fulfillment-policy-id"`) {
		t.Fatalf("body missing listing policies: %q", capturedBody)
	}
}

func TestPublishOffer(t *testing.T) {
	t.Parallel()

	httpClient := &http.Client{
		Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			if r.Method != http.MethodPost {
				t.Fatalf("method = %q, want %q", r.Method, http.MethodPost)
			}
			if r.URL.EscapedPath() != "/sell/inventory/v1/offer/offer-123/publish" {
				t.Fatalf("escaped path = %q", r.URL.EscapedPath())
			}
			if got := r.Header.Get("Authorization"); got != "Bearer access-123" {
				t.Fatalf("authorization = %q", got)
			}

			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(strings.NewReader(`{"listingId":"listing-123"}`)),
			}, nil
		}),
	}

	client, err := NewClient(Config{
		APIBaseURL:    "https://api.sandbox.ebay.test",
		OAuthTokenURL: "https://api.sandbox.ebay.test/identity/v1/oauth2/token",
		AppID:         "app-id",
		CertID:        "cert-id",
		RefreshToken:  "refresh-123",
	}, WithHTTPClient(httpClient))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	response, err := client.PublishOffer(context.Background(), "offer-123", "access-123")
	if err != nil {
		t.Fatalf("PublishOffer: %v", err)
	}
	if response.ListingID != "listing-123" {
		t.Fatalf("listing id = %q", response.ListingID)
	}
}

func TestNewClientRequiresConfig(t *testing.T) {
	t.Parallel()

	_, err := NewClient(Config{})
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "api base url is required") {
		t.Fatalf("unexpected error: %v", err)
	}
}
