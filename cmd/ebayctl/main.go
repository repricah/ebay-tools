package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/repricah/ebay-tools"
)

func main() {
	if err := run(context.Background(), os.Args[1:]); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: ebayctl <smoke|inventory-get|inventory-upsert>")
	}

	cfg, err := configFromEnv()
	if err != nil {
		return err
	}
	client, err := ebaytools.NewClient(cfg)
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

func writeJSON(payload any) error {
	encoded, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	_, err = fmt.Println(string(encoded))
	return err
}
