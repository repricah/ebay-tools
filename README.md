# ebay-tools

Reusable Go client and thin CLI for eBay selling workflows.

Current surface:

- refresh a user access token from a stored refresh token
- fetch seller privileges
- get one inventory item by SKU
- create or replace one inventory item by SKU

CLI:

```bash
EBAY_APP_ID=...
EBAY_CERT_ID=...
EBAY_REFRESH_TOKEN=...
go run ./cmd/ebayctl smoke
go run ./cmd/ebayctl inventory-get my-sku
go run ./cmd/ebayctl inventory-upsert my-sku ./item.json
```
