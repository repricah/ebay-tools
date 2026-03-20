package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	ebaytools "github.com/repricah/ebay-tools"
)

type clientAPI interface {
	RefreshUserAccessToken(ctx context.Context, scopes []string) (*ebaytools.TokenResponse, error)
	GetPrivileges(ctx context.Context, accessToken string) (*ebaytools.SellingPrivileges, error)
	OptInToProgram(ctx context.Context, request ebaytools.OptInToProgramRequest, accessToken string) error
	GetOptedInPrograms(ctx context.Context, accessToken string) (*ebaytools.OptInProgramsResponse, error)
	GetInventoryItem(ctx context.Context, sku, accessToken string) (*ebaytools.InventoryItem, error)
	UpsertInventoryItem(ctx context.Context, sku string, item ebaytools.InventoryItem, accessToken, contentLanguage string) error
	GetInventoryLocations(ctx context.Context, accessToken string) (*ebaytools.InventoryLocationsResponse, error)
	CreateInventoryLocation(ctx context.Context, merchantLocationKey string, location ebaytools.InventoryLocation, accessToken string) error
	GetFulfillmentPolicies(ctx context.Context, marketplaceID, accessToken string) (*ebaytools.FulfillmentPoliciesResponse, error)
	GetPaymentPolicies(ctx context.Context, marketplaceID, accessToken string) (*ebaytools.PaymentPoliciesResponse, error)
	GetReturnPolicies(ctx context.Context, marketplaceID, accessToken string) (*ebaytools.ReturnPoliciesResponse, error)
	CreateFulfillmentPolicy(ctx context.Context, policy ebaytools.FulfillmentPolicy, accessToken string) (*ebaytools.FulfillmentPolicy, error)
	CreatePaymentPolicy(ctx context.Context, policy ebaytools.PaymentPolicy, accessToken string) (*ebaytools.PaymentPolicy, error)
	CreateReturnPolicy(ctx context.Context, policy ebaytools.ReturnPolicy, accessToken string) (*ebaytools.ReturnPolicy, error)
	GetOffers(ctx context.Context, sku, accessToken string) (*ebaytools.OffersResponse, error)
	CreateOffer(ctx context.Context, offer ebaytools.Offer, accessToken string) (*ebaytools.CreateOfferResponse, error)
	PublishOffer(ctx context.Context, offerID, accessToken string) (*ebaytools.PublishOfferResponse, error)
}

var (
	newClient = func(cfg ebaytools.Config) (clientAPI, error) {
		return ebaytools.NewClient(cfg)
	}
	stdout io.Writer = os.Stdout
	stderr io.Writer = os.Stderr
)

func main() {
	if err := run(context.Background(), os.Args[1:]); err != nil {
		_, _ = fmt.Fprintln(stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: ebayctl <smoke|program-opt-in|program-list|policy-list|policy-create|location-list|location-create|inventory-get|inventory-upsert|offer-get|offer-create|offer-publish>")
	}

	cfg, err := configFromEnv()
	if err != nil {
		return err
	}
	client, err := newClient(cfg)
	if err != nil {
		return err
	}

	switch args[0] {
	case "smoke":
		token, err := client.RefreshUserAccessToken(ctx, []string{ebaytools.DefaultReadonlyScope()})
		if err != nil {
			return err
		}
		privileges, err := client.GetPrivileges(ctx, token.AccessToken)
		if err != nil {
			return err
		}
		return writeJSON(map[string]any{
			"token_type":                    token.TokenType,
			"expires_in":                    token.ExpiresIn,
			"seller_registration_completed": privileges.SellerRegistrationCompleted,
			"selling_limit":                 privileges.SellingLimit,
		})
	case "program-opt-in":
		if len(args) < 2 {
			return fmt.Errorf("usage: ebayctl program-opt-in <program-type>")
		}
		token, err := client.RefreshUserAccessToken(ctx, []string{ebaytools.DefaultAccountScope()})
		if err != nil {
			return err
		}
		programType := strings.TrimSpace(args[1])
		if programType == "" {
			return fmt.Errorf("program type is required")
		}
		if err := client.OptInToProgram(ctx, ebaytools.OptInToProgramRequest{ProgramType: programType}, token.AccessToken); err != nil {
			return err
		}
		return writeJSON(map[string]any{
			"programType": programType,
			"status":      "opt-in-requested",
		})
	case "program-list":
		token, err := client.RefreshUserAccessToken(ctx, []string{ebaytools.DefaultAccountScope()})
		if err != nil {
			return err
		}
		response, err := client.GetOptedInPrograms(ctx, token.AccessToken)
		if err != nil {
			return err
		}
		return writeJSON(response)
	case "policy-list":
		if len(args) < 2 {
			return fmt.Errorf("usage: ebayctl policy-list <fulfillment|payment|return> [marketplace-id]")
		}
		token, err := client.RefreshUserAccessToken(ctx, []string{ebaytools.DefaultAccountScope()})
		if err != nil {
			return err
		}
		marketplaceID := "EBAY_US"
		if len(args) >= 3 && strings.TrimSpace(args[2]) != "" {
			marketplaceID = strings.TrimSpace(args[2])
		}
		switch strings.TrimSpace(args[1]) {
		case "fulfillment":
			response, err := client.GetFulfillmentPolicies(ctx, marketplaceID, token.AccessToken)
			if err != nil {
				return err
			}
			return writeJSON(response)
		case "payment":
			response, err := client.GetPaymentPolicies(ctx, marketplaceID, token.AccessToken)
			if err != nil {
				return err
			}
			return writeJSON(response)
		case "return":
			response, err := client.GetReturnPolicies(ctx, marketplaceID, token.AccessToken)
			if err != nil {
				return err
			}
			return writeJSON(response)
		default:
			return fmt.Errorf("unknown policy type %q", args[1])
		}
	case "policy-create":
		if len(args) < 3 {
			return fmt.Errorf("usage: ebayctl policy-create <fulfillment|payment|return> <json-file>")
		}
		token, err := client.RefreshUserAccessToken(ctx, []string{ebaytools.DefaultAccountScope()})
		if err != nil {
			return err
		}
		switch strings.TrimSpace(args[1]) {
		case "fulfillment":
			policy, err := loadFulfillmentPolicy(args[2])
			if err != nil {
				return err
			}
			response, err := client.CreateFulfillmentPolicy(ctx, policy, token.AccessToken)
			if err != nil {
				return err
			}
			return writeJSON(response)
		case "payment":
			policy, err := loadPaymentPolicy(args[2])
			if err != nil {
				return err
			}
			response, err := client.CreatePaymentPolicy(ctx, policy, token.AccessToken)
			if err != nil {
				return err
			}
			return writeJSON(response)
		case "return":
			policy, err := loadReturnPolicy(args[2])
			if err != nil {
				return err
			}
			response, err := client.CreateReturnPolicy(ctx, policy, token.AccessToken)
			if err != nil {
				return err
			}
			return writeJSON(response)
		default:
			return fmt.Errorf("unknown policy type %q", args[1])
		}
	case "location-list":
		token, err := client.RefreshUserAccessToken(ctx, []string{ebaytools.DefaultInventoryReadonlyScope()})
		if err != nil {
			return err
		}
		response, err := client.GetInventoryLocations(ctx, token.AccessToken)
		if err != nil {
			return err
		}
		return writeJSON(response)
	case "location-create":
		if len(args) < 3 {
			return fmt.Errorf("usage: ebayctl location-create <merchant-location-key> <json-file>")
		}
		location, err := loadInventoryLocation(args[2])
		if err != nil {
			return err
		}
		token, err := client.RefreshUserAccessToken(ctx, []string{ebaytools.DefaultInventoryScope()})
		if err != nil {
			return err
		}
		if err := client.CreateInventoryLocation(ctx, args[1], location, token.AccessToken); err != nil {
			return err
		}
		return writeJSON(map[string]any{
			"merchantLocationKey": args[1],
			"status":              "created",
		})
	case "inventory-get":
		if len(args) < 2 {
			return fmt.Errorf("usage: ebayctl inventory-get <sku>")
		}
		token, err := client.RefreshUserAccessToken(ctx, []string{ebaytools.DefaultInventoryReadonlyScope()})
		if err != nil {
			return err
		}
		item, err := client.GetInventoryItem(ctx, args[1], token.AccessToken)
		if err != nil {
			return err
		}
		return writeJSON(item)
	case "inventory-upsert":
		if len(args) < 3 {
			return fmt.Errorf("usage: ebayctl inventory-upsert <sku> <json-file>")
		}
		item, err := loadInventoryItem(args[2])
		if err != nil {
			return err
		}
		token, err := client.RefreshUserAccessToken(ctx, []string{ebaytools.DefaultInventoryScope()})
		if err != nil {
			return err
		}
		if err := client.UpsertInventoryItem(ctx, args[1], item, token.AccessToken, "en-US"); err != nil {
			return err
		}
		return writeJSON(map[string]any{
			"sku":    args[1],
			"status": "upserted",
		})
	case "offer-get":
		if len(args) < 2 {
			return fmt.Errorf("usage: ebayctl offer-get <sku>")
		}
		token, err := client.RefreshUserAccessToken(ctx, []string{ebaytools.DefaultInventoryReadonlyScope()})
		if err != nil {
			return err
		}
		offers, err := client.GetOffers(ctx, args[1], token.AccessToken)
		if err != nil {
			return err
		}
		return writeJSON(offers)
	case "offer-create":
		if len(args) < 2 {
			return fmt.Errorf("usage: ebayctl offer-create <json-file>")
		}
		offer, err := loadOffer(args[1])
		if err != nil {
			return err
		}
		token, err := client.RefreshUserAccessToken(ctx, []string{ebaytools.DefaultInventoryScope()})
		if err != nil {
			return err
		}
		response, err := client.CreateOffer(ctx, offer, token.AccessToken)
		if err != nil {
			return err
		}
		return writeJSON(response)
	case "offer-publish":
		if len(args) < 2 {
			return fmt.Errorf("usage: ebayctl offer-publish <offer-id>")
		}
		token, err := client.RefreshUserAccessToken(ctx, []string{ebaytools.DefaultInventoryScope()})
		if err != nil {
			return err
		}
		response, err := client.PublishOffer(ctx, args[1], token.AccessToken)
		if err != nil {
			return err
		}
		return writeJSON(response)
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func configFromEnv() (ebaytools.Config, error) {
	cfg := ebaytools.Config{
		APIBaseURL:    strings.TrimSpace(os.Getenv("EBAY_API_BASE_URL")),
		OAuthTokenURL: strings.TrimSpace(os.Getenv("EBAY_OAUTH_TOKEN_URL")),
		AppID:         strings.TrimSpace(os.Getenv("EBAY_APP_ID")),
		CertID:        strings.TrimSpace(os.Getenv("EBAY_CERT_ID")),
		RefreshToken:  strings.TrimSpace(os.Getenv("EBAY_REFRESH_TOKEN")),
	}
	if cfg.APIBaseURL == "" {
		cfg.APIBaseURL = "https://api.sandbox.ebay.com"
	}
	if cfg.OAuthTokenURL == "" {
		cfg.OAuthTokenURL = "https://api.sandbox.ebay.com/identity/v1/oauth2/token"
	}
	return cfg, nil
}

func loadInventoryItem(path string) (ebaytools.InventoryItem, error) {
	data, err := os.ReadFile(strings.TrimSpace(path))
	if err != nil {
		return ebaytools.InventoryItem{}, fmt.Errorf("read inventory item file: %w", err)
	}
	var item ebaytools.InventoryItem
	if err := json.Unmarshal(data, &item); err != nil {
		return ebaytools.InventoryItem{}, fmt.Errorf("decode inventory item file: %w", err)
	}
	return item, nil
}

func loadInventoryLocation(path string) (ebaytools.InventoryLocation, error) {
	data, err := os.ReadFile(strings.TrimSpace(path))
	if err != nil {
		return ebaytools.InventoryLocation{}, fmt.Errorf("read inventory location file: %w", err)
	}
	var location ebaytools.InventoryLocation
	if err := json.Unmarshal(data, &location); err != nil {
		return ebaytools.InventoryLocation{}, fmt.Errorf("decode inventory location file: %w", err)
	}
	return location, nil
}

func loadOffer(path string) (ebaytools.Offer, error) {
	data, err := os.ReadFile(strings.TrimSpace(path))
	if err != nil {
		return ebaytools.Offer{}, fmt.Errorf("read offer file: %w", err)
	}
	var offer ebaytools.Offer
	if err := json.Unmarshal(data, &offer); err != nil {
		return ebaytools.Offer{}, fmt.Errorf("decode offer file: %w", err)
	}
	return offer, nil
}

func loadFulfillmentPolicy(path string) (ebaytools.FulfillmentPolicy, error) {
	data, err := os.ReadFile(strings.TrimSpace(path))
	if err != nil {
		return ebaytools.FulfillmentPolicy{}, fmt.Errorf("read fulfillment policy file: %w", err)
	}
	var policy ebaytools.FulfillmentPolicy
	if err := json.Unmarshal(data, &policy); err != nil {
		return ebaytools.FulfillmentPolicy{}, fmt.Errorf("decode fulfillment policy file: %w", err)
	}
	return policy, nil
}

func loadPaymentPolicy(path string) (ebaytools.PaymentPolicy, error) {
	data, err := os.ReadFile(strings.TrimSpace(path))
	if err != nil {
		return ebaytools.PaymentPolicy{}, fmt.Errorf("read payment policy file: %w", err)
	}
	var policy ebaytools.PaymentPolicy
	if err := json.Unmarshal(data, &policy); err != nil {
		return ebaytools.PaymentPolicy{}, fmt.Errorf("decode payment policy file: %w", err)
	}
	return policy, nil
}

func loadReturnPolicy(path string) (ebaytools.ReturnPolicy, error) {
	data, err := os.ReadFile(strings.TrimSpace(path))
	if err != nil {
		return ebaytools.ReturnPolicy{}, fmt.Errorf("read return policy file: %w", err)
	}
	var policy ebaytools.ReturnPolicy
	if err := json.Unmarshal(data, &policy); err != nil {
		return ebaytools.ReturnPolicy{}, fmt.Errorf("decode return policy file: %w", err)
	}
	return policy, nil
}

func writeJSON(payload any) error {
	encoded, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(stdout, string(encoded))
	return err
}
