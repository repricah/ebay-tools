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
	optInProgramPath          = "/sell/account/v1/program/opt_in"
	getOptedInProgramsPath    = "/sell/account/v1/program/get_opted_in_programs"
	fulfillmentPolicyPath     = "/sell/account/v1/fulfillment_policy"
	paymentPolicyPath         = "/sell/account/v1/payment_policy"
	returnPolicyPath          = "/sell/account/v1/return_policy"
	inventoryItemPath         = "/sell/inventory/v1/inventory_item"
	locationPath              = "/sell/inventory/v1/location"
	offerPath                 = "/sell/inventory/v1/offer"
	formContentType           = "application/x-www-form-urlencoded"
	jsonContentType           = "application/json"
	defaultReadonlyScope      = "https://api.ebay.com/oauth/api_scope/sell.account.readonly"
	defaultAccountScope       = "https://api.ebay.com/oauth/api_scope/sell.account"
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

// HTTPClient defines the minimal interface required to execute HTTP requests.
// It matches the method signature of http.Client.Do to allow custom clients.
type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

type Client struct {
	apiBaseURL    string
	oauthTokenURL string
	appID         string
	certID        string
	refreshToken  string
	httpClient    HTTPClient
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
	SKU                  string                `json:"sku,omitempty"`
	Locale               string                `json:"locale,omitempty"`
	Condition            string                `json:"condition,omitempty"`
	ConditionDescription string                `json:"conditionDescription,omitempty"`
	ConditionDescriptors []ConditionDescriptor `json:"conditionDescriptors,omitempty"`
	Availability         *Availability         `json:"availability,omitempty"`
	Product              *Product              `json:"product,omitempty"`
}

type ConditionDescriptor struct {
	Name           string   `json:"name,omitempty"`
	Values         []string `json:"values,omitempty"`
	AdditionalInfo string   `json:"additionalInfo,omitempty"`
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

type InventoryLocationsResponse struct {
	Href      string              `json:"href,omitempty"`
	Total     int                 `json:"total,omitempty"`
	Limit     int                 `json:"limit,omitempty"`
	Offset    int                 `json:"offset,omitempty"`
	Locations []InventoryLocation `json:"locations,omitempty"`
}

type InventoryLocation struct {
	MerchantLocationKey    string           `json:"merchantLocationKey,omitempty"`
	Name                   string           `json:"name,omitempty"`
	MerchantLocationStatus string           `json:"merchantLocationStatus,omitempty"`
	LocationTypes          []string         `json:"locationTypes,omitempty"`
	Location               *LocationDetails `json:"location,omitempty"`
	LocationWebURL         string           `json:"locationWebUrl,omitempty"`
	Phone                  string           `json:"phone,omitempty"`
}

type LocationDetails struct {
	Address *Address `json:"address,omitempty"`
}

type Address struct {
	AddressLine1    string `json:"addressLine1,omitempty"`
	AddressLine2    string `json:"addressLine2,omitempty"`
	City            string `json:"city,omitempty"`
	StateOrProvince string `json:"stateOrProvince,omitempty"`
	PostalCode      string `json:"postalCode,omitempty"`
	Country         string `json:"country,omitempty"`
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

type OptInProgramsResponse struct {
	Programs []Program `json:"programs,omitempty"`
}

type Program struct {
	ProgramType string `json:"programType,omitempty"`
	OptInStatus string `json:"optInStatus,omitempty"`
}

type OptInToProgramRequest struct {
	ProgramType string `json:"programType"`
}

type CategoryType struct {
	Name string `json:"name,omitempty"`
}

type TimeDuration struct {
	Unit  string `json:"unit,omitempty"`
	Value int    `json:"value,omitempty"`
}

type RegionIncluded struct {
	RegionName string `json:"regionName,omitempty"`
}

type ShippingService struct {
	ShippingCarrierCode string  `json:"shippingCarrierCode,omitempty"`
	ShippingServiceCode string  `json:"shippingServiceCode,omitempty"`
	ShippingCost        *Amount `json:"shippingCost,omitempty"`
}

type ShippingOption struct {
	OptionType       string            `json:"optionType,omitempty"`
	CostType         string            `json:"costType,omitempty"`
	ShippingServices []ShippingService `json:"shippingServices,omitempty"`
	RegionIncluded   []RegionIncluded  `json:"regionIncluded,omitempty"`
}

type FulfillmentPolicy struct {
	FulfillmentPolicyID string           `json:"fulfillmentPolicyId,omitempty"`
	Name                string           `json:"name,omitempty"`
	Description         string           `json:"description,omitempty"`
	MarketplaceID       string           `json:"marketplaceId,omitempty"`
	CategoryTypes       []CategoryType   `json:"categoryTypes,omitempty"`
	HandlingTime        *TimeDuration    `json:"handlingTime,omitempty"`
	ShippingOptions     []ShippingOption `json:"shippingOptions,omitempty"`
	GlobalShipping      bool             `json:"globalShipping,omitempty"`
	PickupDropOff       bool             `json:"pickupDropOff,omitempty"`
	FreightShipping     bool             `json:"freightShipping,omitempty"`
}

type PaymentMethod struct {
	PaymentMethodType string `json:"paymentMethodType,omitempty"`
}

type PaymentPolicy struct {
	PaymentPolicyID     string          `json:"paymentPolicyId,omitempty"`
	Name                string          `json:"name,omitempty"`
	Description         string          `json:"description,omitempty"`
	MarketplaceID       string          `json:"marketplaceId,omitempty"`
	CategoryTypes       []CategoryType  `json:"categoryTypes,omitempty"`
	ImmediatePay        bool            `json:"immediatePay,omitempty"`
	PaymentMethods      []PaymentMethod `json:"paymentMethods,omitempty"`
	PaymentInstructions string          `json:"paymentInstructions,omitempty"`
}

type ReturnPolicy struct {
	ReturnPolicyID          string         `json:"returnPolicyId,omitempty"`
	Name                    string         `json:"name,omitempty"`
	Description             string         `json:"description,omitempty"`
	MarketplaceID           string         `json:"marketplaceId,omitempty"`
	CategoryTypes           []CategoryType `json:"categoryTypes,omitempty"`
	ReturnsAccepted         bool           `json:"returnsAccepted,omitempty"`
	ReturnPeriod            *TimeDuration  `json:"returnPeriod,omitempty"`
	ReturnMethod            string         `json:"returnMethod,omitempty"`
	RefundMethod            string         `json:"refundMethod,omitempty"`
	ReturnShippingCostPayer string         `json:"returnShippingCostPayer,omitempty"`
}

type FulfillmentPoliciesResponse struct {
	Href                string              `json:"href,omitempty"`
	Total               int                 `json:"total,omitempty"`
	Limit               int                 `json:"limit,omitempty"`
	Size                int                 `json:"size,omitempty"`
	FulfillmentPolicies []FulfillmentPolicy `json:"fulfillmentPolicies,omitempty"`
}

type PaymentPoliciesResponse struct {
	Href            string          `json:"href,omitempty"`
	Total           int             `json:"total,omitempty"`
	Limit           int             `json:"limit,omitempty"`
	Size            int             `json:"size,omitempty"`
	PaymentPolicies []PaymentPolicy `json:"paymentPolicies,omitempty"`
}

type ReturnPoliciesResponse struct {
	Href           string         `json:"href,omitempty"`
	Total          int            `json:"total,omitempty"`
	Limit          int            `json:"limit,omitempty"`
	Size           int            `json:"size,omitempty"`
	ReturnPolicies []ReturnPolicy `json:"returnPolicies,omitempty"`
}

func WithHTTPClient(httpClient HTTPClient) Option {
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

func DefaultAccountScope() string {
	return defaultAccountScope
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

func (c *Client) OptInToProgram(ctx context.Context, request OptInToProgramRequest, accessToken string) error {
	body, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("marshal opt-in request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.apiBaseURL+optInProgramPath, strings.NewReader(string(body)))
	if err != nil {
		return fmt.Errorf("build optInToProgram request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(accessToken))
	req.Header.Set("Content-Type", jsonContentType)

	return c.doJSON(req, nil)
}

func (c *Client) GetOptedInPrograms(ctx context.Context, accessToken string) (*OptInProgramsResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.apiBaseURL+getOptedInProgramsPath, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("build getOptedInPrograms request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(accessToken))

	var response OptInProgramsResponse
	if err := c.doJSON(req, &response); err != nil {
		return nil, err
	}
	return &response, nil
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

func (c *Client) GetInventoryLocations(ctx context.Context, accessToken string) (*InventoryLocationsResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.apiBaseURL+locationPath, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("build getInventoryLocations request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(accessToken))

	var response InventoryLocationsResponse
	if err := c.doJSON(req, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

func (c *Client) CreateInventoryLocation(ctx context.Context, merchantLocationKey string, location InventoryLocation, accessToken string) error {
	cleanKey := strings.TrimSpace(merchantLocationKey)
	if cleanKey == "" {
		return fmt.Errorf("merchant location key is required")
	}

	body, err := json.Marshal(location)
	if err != nil {
		return fmt.Errorf("marshal inventory location: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.apiBaseURL+locationPath+"/"+url.PathEscape(cleanKey), strings.NewReader(string(body)))
	if err != nil {
		return fmt.Errorf("build createInventoryLocation request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(accessToken))
	req.Header.Set("Content-Type", jsonContentType)

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

func (c *Client) GetFulfillmentPolicies(ctx context.Context, marketplaceID, accessToken string) (*FulfillmentPoliciesResponse, error) {
	return getPolicyList[FulfillmentPoliciesResponse](ctx, c, fulfillmentPolicyPath, marketplaceID, accessToken, "getFulfillmentPolicies")
}

func (c *Client) GetPaymentPolicies(ctx context.Context, marketplaceID, accessToken string) (*PaymentPoliciesResponse, error) {
	return getPolicyList[PaymentPoliciesResponse](ctx, c, paymentPolicyPath, marketplaceID, accessToken, "getPaymentPolicies")
}

func (c *Client) GetReturnPolicies(ctx context.Context, marketplaceID, accessToken string) (*ReturnPoliciesResponse, error) {
	return getPolicyList[ReturnPoliciesResponse](ctx, c, returnPolicyPath, marketplaceID, accessToken, "getReturnPolicies")
}

func (c *Client) CreateFulfillmentPolicy(ctx context.Context, policy FulfillmentPolicy, accessToken string) (*FulfillmentPolicy, error) {
	return createPolicy(ctx, c, fulfillmentPolicyPath, policy, accessToken, "createFulfillmentPolicy")
}

func (c *Client) CreatePaymentPolicy(ctx context.Context, policy PaymentPolicy, accessToken string) (*PaymentPolicy, error) {
	return createPolicy(ctx, c, paymentPolicyPath, policy, accessToken, "createPaymentPolicy")
}

func (c *Client) CreateReturnPolicy(ctx context.Context, policy ReturnPolicy, accessToken string) (*ReturnPolicy, error) {
	return createPolicy(ctx, c, returnPolicyPath, policy, accessToken, "createReturnPolicy")
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
	req.Header.Set("Content-Language", "en-US")

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

func getPolicyList[T any](ctx context.Context, c *Client, path, marketplaceID, accessToken, op string) (*T, error) {
	cleanMarketplaceID := strings.TrimSpace(marketplaceID)
	if cleanMarketplaceID == "" {
		return nil, fmt.Errorf("marketplace id is required")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.apiBaseURL+path, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("build %s request: %w", op, err)
	}
	req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(accessToken))
	query := req.URL.Query()
	query.Set("marketplace_id", cleanMarketplaceID)
	req.URL.RawQuery = query.Encode()

	var response T
	if err := c.doJSON(req, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

func createPolicy[T any](ctx context.Context, c *Client, path string, payload T, accessToken, op string) (*T, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal %s payload: %w", op, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.apiBaseURL+path, strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("build %s request: %w", op, err)
	}
	req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(accessToken))
	req.Header.Set("Content-Type", jsonContentType)

	var response T
	if err := c.doJSON(req, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

func inventoryItemResourcePath(sku string) string {
	return inventoryItemPath + "/" + url.PathEscape(strings.TrimSpace(sku))
}

func publishOfferPath(offerID string) string {
	return offerPath + "/" + url.PathEscape(strings.TrimSpace(offerID)) + "/publish"
}
