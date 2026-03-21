# ebay-tools

Reusable Go client and thin CLI for eBay selling workflows.

It also exposes a small draft/contracts layer so app code can map domain objects into reusable listing shapes without constructing raw eBay payloads everywhere.

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
- build `InventoryItem` / `Offer` structs from draft archetypes:
  - `SingleCardDraft`
  - `PlaysetDraft`
  - `SealedProductDraft`
  - `GenericCollectibleDraft`

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

Draft/contracts example:

```go
draft := ebaytools.PlaysetDraft{
	BaseDraft: ebaytools.BaseDraft{
		SKU:         "playset-lightning-bolt",
		Description: "Four copies ready to list",
		Listing: ebaytools.ListingOptions{
			MarketplaceID:       "EBAY_US",
			CategoryID:          "183454",
			MerchantLocationKey: "warehouse-1",
			Policies: ebaytools.PolicyRefs{
				FulfillmentPolicyID: "fulfillment-id",
				PaymentPolicyID:     "payment-id",
				ReturnPolicyID:      "return-id",
			},
			Price: ebaytools.Money{Currency: "USD", Value: "3.99"},
		},
	},
	Game:     "Magic: The Gathering",
	CardName: "Lightning Bolt",
	SetName:  "Magic 2010",
	Condition: ebaytools.TradingCardCondition{
		UngradedConditionDescriptorID: "400010",
	},
}

item, err := ebaytools.BuildInventoryItemFromDraft(draft)
if err != nil {
	// handle error
}

offer, err := ebaytools.BuildOfferFromDraft(draft)
if err != nil {
	// handle error
}
```
