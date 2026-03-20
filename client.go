package ebaytools

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	defaultTimeout            = 15 * time.Second
	getPrivilegesPath         = "/sell/account/v1/privilege"
	inventoryItemPath         = "/sell/inventory/v1/inventory_item"
	offerPath                 = "/sell/inventory/v1/offer"
	formContentType           = "application/x-www-form-urlencoded"
	jsonContentType           = "application/json"
	defaultReadonlyScope      = "https://api.ebay.com/oauth/api_scope/sell.account.readonly"
	defaultInventoryScope     = "https://api.ebay.com/oauth/api_scope/sell.inventory"
	defaultInventoryReadScope = "https://api.ebay.com/oauth/api_scope/sell.inventory.readonly"
)

type Config struct {
	APIBaseURL    string
	OAuthTokenURL string
	AppID         string
	CertID        string
	RefreshToken  string
}

type Option func(*Client)

type Client struct {
	apiBaseURL    string
	oauthTokenURL string
	appID         string
	certID        string
	refreshToken  string
	httpClient    *http.Client
}

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

type SellingPrivileges struct {
	SellerRegistrationCompleted bool          `json:"sellerRegistrationCompleted"`
	SellingLimit                *SellingLimit `json:"sellingLimit,omitempty"`
}

type SellingLimit struct {
	Amount   *Amount `json:"amount,omitempty"`
	Quantity int     `json:"quantity,omitempty"`
}

type Amount struct {
	Currency string `json:"currency"`
	Value    string `json:"value"`
}

type InventoryItem struct {
	SKU                  string        `json:"sku,omitempty"`
	Locale               string        `json:"locale,omitempty"`
	Condition            string        `json:"condition,omitempty"`
	ConditionDescription string        `json:"conditionDescription,omitempty"`
	Availability         *Availability `json:"availability,omitempty"`
	Product              *Product      `json:"product,omitempty"`
}

type Availability struct {
	ShipToLocationAvailability *ShipToLocationAvailability `json:"shipToLocationAvailability,omitempty"`
}

type ShipToLocationAvailability struct {
	Quantity int `json:"quantity,omitempty"`
}

type Product struct {
	Title       string              `json:"title,omitempty"`
	Description string              `json:"description,omitempty"`
	Aspects     map[string][]string `json:"aspects,omitempty"`
	Brand       string              `json:"brand,omitempty"`
	MPN         string              `json:"mpn,omitempty"`
	ImageURLs   []string            `json:"imageUrls,omitempty"`
	Subtitle    string              `json:"subtitle,omitempty"`
	UPC         string              `json:"upc,omitempty"`
}

type OffersResponse struct {
	Href   string  `json:"href,omitempty"`
	Total  int     `json:"total,omitempty"`
	Limit  int     `json:"limit,omitempty"`
	Size   int     `json:"size,omitempty"`
	Offers []Offer `json:"offers,omitempty"`
}

type Offer struct {
	OfferID               string           `json:"offerId,omitempty"`
	SKU                   string           `json:"sku,omitempty"`
	MarketplaceID         string           `json:"marketplaceId,omitempty"`
	Format                string           `json:"format,omitempty"`
	AvailableQuantity     int              `json:"availableQuantity,omitempty"`
	CategoryID            string           `json:"categoryId,omitempty"`
	MerchantLocationKey   string           `json:"merchantLocationKey,omitempty"`
	ListingDescription    string           `json:"listingDescription,omitempty"`
	ListingDuration       string           `json:"listingDuration,omitempty"`
	QuantityLimitPerBuyer int              `json:"quantityLimitPerBuyer,omitempty"`
	Status                string           `json:"status,omitempty"`
	ListingPolicies       *ListingPolicies `json:"listingPolicies,omitempty"`
	PricingSummary        *PricingSummary  `json:"pricingSummary,omitempty"`
}

type ListingPolicies struct {
	FulfillmentPolicyID string `json:"fulfillmentPolicyId,omitempty"`
	PaymentPolicyID     string `json:"paymentPolicyId,omitempty"`
	ReturnPolicyID      string `json:"returnPolicyId,omitempty"`
}

type PricingSummary struct {
	Price *Amount `json:"price,omitempty"`
}

type CreateOfferResponse struct {
	OfferID string `json:"offerId"`
}

type PublishOfferResponse struct {
	ListingID string `json:"listingId"`
}

func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		if httpClient != nil {
			c.httpClient = httpClient
		}
	}
}

func NewClient(cfg Config, opts ...Option) (*Client, error) {
	client := &Client{
		apiBaseURL:    strings.TrimRight(strings.TrimSpace(cfg.APIBaseURL), "/"),
		oauthTokenURL: strings.TrimSpace(cfg.OAuthTokenURL),
		appID:         strings.TrimSpace(cfg.AppID),
		certID:        strings.TrimSpace(cfg.CertID),
		refreshToken:  strings.TrimSpace(cfg.RefreshToken),
		httpClient:    &http.Client{Timeout: defaultTimeout},
	}
	for _, opt := range opts {
		opt(client)
	}

	if client.apiBaseURL == "" {
		return nil, fmt.Errorf("api base url is required")
	}
	if client.oauthTokenURL == "" {
		return nil, fmt.Errorf("oauth token url is required")
	}
	if client.appID == "" {
		return nil, fmt.Errorf("app id is required")
	}
	if client.certID == "" {
		return nil, fmt.Errorf("cert id is required")
	}
	if client.refreshToken == "" {
		return nil, fmt.Errorf("refresh token is required")
	}

	return client, nil
}

func DefaultReadonlyScope() string {
	return defaultReadonlyScope
}

func DefaultInventoryScope() string {
	return defaultInventoryScope
}

func DefaultInventoryReadonlyScope() string {
	return defaultInventoryReadScope
}

func (c *Client) RefreshUserAccessToken(ctx context.Context, scopes []string) (*TokenResponse, error) {
	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", c.refreshToken)
	if joinedScopes := strings.TrimSpace(strings.Join(scopes, " ")); joinedScopes != "" {
		form.Set("scope", joinedScopes)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.oauthTokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("build refresh token request: %w", err)
	}
	req.Header.Set("Content-Type", formContentType)
	req.Header.Set("Authorization", "Basic "+c.basicAuth())

	var token TokenResponse
	if err := c.doJSON(req, &token); err != nil {
		return nil, err
	}
	if strings.TrimSpace(token.AccessToken) == "" {
		return nil, fmt.Errorf("refresh token response did not include access_token")
	}
	return &token, nil
}

func (c *Client) GetPrivileges(ctx context.Context, accessToken string) (*SellingPrivileges, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.apiBaseURL+getPrivilegesPath, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("build getPrivileges request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(accessToken))

	var privileges SellingPrivileges
	if err := c.doJSON(req, &privileges); err != nil {
		return nil, err
	}
	return &privileges, nil
}

func (c *Client) GetInventoryItem(ctx context.Context, sku, accessToken string) (*InventoryItem, error) {
	cleanSKU := strings.TrimSpace(sku)
	if cleanSKU == "" {
		return nil, fmt.Errorf("sku is required")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.apiBaseURL+inventoryItemResourcePath(cleanSKU), http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("build getInventoryItem request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(accessToken))

	var item InventoryItem
	if err := c.doJSON(req, &item); err != nil {
		return nil, err
	}
	return &item, nil
}

func (c *Client) UpsertInventoryItem(ctx context.Context, sku string, item InventoryItem, accessToken, contentLanguage string) error {
	cleanSKU := strings.TrimSpace(sku)
	if cleanSKU == "" {
		return fmt.Errorf("sku is required")
	}

	body, err := json.Marshal(item)
	if err != nil {
		return fmt.Errorf("marshal inventory item: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, c.apiBaseURL+inventoryItemResourcePath(cleanSKU), strings.NewReader(string(body)))
	if err != nil {
		return fmt.Errorf("build upsertInventoryItem request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(accessToken))
	req.Header.Set("Content-Type", jsonContentType)

	cleanLanguage := strings.TrimSpace(contentLanguage)
	if cleanLanguage == "" {
		cleanLanguage = "en-US"
	}
	req.Header.Set("Content-Language", cleanLanguage)

	return c.doJSON(req, nil)
}

func (c *Client) GetOffers(ctx context.Context, sku, accessToken string) (*OffersResponse, error) {
	cleanSKU := strings.TrimSpace(sku)
	if cleanSKU == "" {
		return nil, fmt.Errorf("sku is required")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.apiBaseURL+offerPath, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("build getOffers request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(accessToken))

	query := req.URL.Query()
	query.Set("sku", cleanSKU)
	req.URL.RawQuery = query.Encode()

	var offers OffersResponse
	if err := c.doJSON(req, &offers); err != nil {
		return nil, err
	}
	return &offers, nil
}

func (c *Client) CreateOffer(ctx context.Context, offer Offer, accessToken string) (*CreateOfferResponse, error) {
	body, err := json.Marshal(offer)
	if err != nil {
		return nil, fmt.Errorf("marshal offer: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.apiBaseURL+offerPath, strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("build createOffer request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(accessToken))
	req.Header.Set("Content-Type", jsonContentType)

	var response CreateOfferResponse
	if err := c.doJSON(req, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

func (c *Client) PublishOffer(ctx context.Context, offerID, accessToken string) (*PublishOfferResponse, error) {
	cleanOfferID := strings.TrimSpace(offerID)
	if cleanOfferID == "" {
		return nil, fmt.Errorf("offer id is required")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.apiBaseURL+publishOfferPath(cleanOfferID), http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("build publishOffer request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(accessToken))

	var response PublishOfferResponse
	if err := c.doJSON(req, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

func (c *Client) doJSON(req *http.Request, dst any) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request %s %s failed: %w", req.Method, req.URL.String(), err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("%s %s returned %d: %s", req.Method, req.URL.Path, resp.StatusCode, strings.TrimSpace(string(body)))
	}

	if dst == nil || resp.StatusCode == http.StatusNoContent {
		_, _ = io.Copy(io.Discard, resp.Body)
		return nil
	}

	if err := json.NewDecoder(resp.Body).Decode(dst); err != nil {
		if err == io.EOF {
			return nil
		}
		return fmt.Errorf("decode %s %s response: %w", req.Method, req.URL.Path, err)
	}
	return nil
}

func (c *Client) basicAuth() string {
	raw := c.appID + ":" + c.certID
	return base64.StdEncoding.EncodeToString([]byte(raw))
}

func inventoryItemResourcePath(sku string) string {
	return inventoryItemPath + "/" + url.PathEscape(strings.TrimSpace(sku))
}

func publishOfferPath(offerID string) string {
	return offerPath + "/" + url.PathEscape(strings.TrimSpace(offerID)) + "/publish"
}
