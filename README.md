# ebay-tools

Reusable Go client and thin CLI for eBay selling workflows.

Current surface:

- refresh a user access token from a stored refresh token
- fetch seller privileges
- opt into seller account programs
- list and create business policies
- create inventory locations
- get one inventory item by SKU
- create or replace one inventory item by SKU
- get offers for one SKU
- create and publish one offer

Operator workflow docs:

- sealed photo-research handoff for ChatGPT/Codex:
  - [`docs/sealed-research-handoff.md`](docs/sealed-research-handoff.md)

CLI:

```bash
EBAY_APP_ID=...
EBAY_CERT_ID=...
EBAY_REFRESH_TOKEN=...
go run ./cmd/ebayctl smoke
go run ./cmd/ebayctl program-opt-in SELLING_POLICY_MANAGEMENT
go run ./cmd/ebayctl program-list
go run ./cmd/ebayctl policy-list fulfillment
go run ./cmd/ebayctl policy-create fulfillment ./fulfillment-policy.json
go run ./cmd/ebayctl location-list
go run ./cmd/ebayctl location-create warehouse-1 ./location.json
go run ./cmd/ebayctl inventory-get my-sku
go run ./cmd/ebayctl inventory-upsert my-sku ./item.json
go run ./cmd/ebayctl offer-get my-sku
go run ./cmd/ebayctl offer-create ./offer.json
go run ./cmd/ebayctl offer-publish offer-id
```

Typical sandbox bootstrap flow:

1. `program-opt-in SELLING_POLICY_MANAGEMENT`
2. `policy-create fulfillment ./fulfillment-policy.json`
3. `policy-create payment ./payment-policy.json`
4. `policy-create return ./return-policy.json`
5. `location-create warehouse-1 ./location.json`
6. `inventory-upsert my-sku ./item.json`
7. `offer-create ./offer.json`
8. `offer-publish offer-id`
