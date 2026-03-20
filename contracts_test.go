package ebaytools

import "testing"

func TestBuildInventoryItemFromSingleCardDraft(t *testing.T) {
	t.Parallel()

	item, err := BuildInventoryItemFromDraft(SingleCardDraft{
		BaseDraft: BaseDraft{
			SKU:         "single-1",
			Description: "Ungraded single card",
			Quantity:    1,
			ImageURLs:   []string{"https://example.test/lotus.jpg"},
		},
		Game:     "Magic: The Gathering",
		CardName: "Black Lotus",
		SetName:  "Limited Edition Alpha",
		Language: "English",
		Condition: TradingCardCondition{
			UngradedConditionDescriptorID: "400010",
		},
	})
	if err != nil {
		t.Fatalf("BuildInventoryItemFromDraft: %v", err)
	}
	if item.SKU != "single-1" {
		t.Fatalf("sku = %q", item.SKU)
	}
	if item.Condition != "USED_VERY_GOOD" {
		t.Fatalf("condition = %q", item.Condition)
	}
	if len(item.ConditionDescriptors) != 1 || item.ConditionDescriptors[0].Name != "40001" {
		t.Fatalf("conditionDescriptors = %#v", item.ConditionDescriptors)
	}
	if item.Product == nil || item.Product.Title != "Black Lotus" {
		t.Fatalf("product = %#v", item.Product)
	}
	if item.Product.Aspects["Game"][0] != "Magic: The Gathering" {
		t.Fatalf("aspects = %#v", item.Product.Aspects)
	}
	if item.Availability == nil || item.Availability.ShipToLocationAvailability == nil || item.Availability.ShipToLocationAvailability.Quantity != 1 {
		t.Fatalf("availability = %#v", item.Availability)
	}
}

func TestBuildOfferFromPlaysetDraftDefaultsQuantityAndTitle(t *testing.T) {
	t.Parallel()

	draft := PlaysetDraft{
		BaseDraft: BaseDraft{
			SKU:         "playset-1",
			Description: "Four copies of a sandbox common",
			Listing: ListingOptions{
				MarketplaceID:       "EBAY_US",
				CategoryID:          "183454",
				MerchantLocationKey: "warehouse-1",
				Policies: PolicyRefs{
					FulfillmentPolicyID: "fulfillment-id",
					PaymentPolicyID:     "payment-id",
					ReturnPolicyID:      "return-id",
				},
				Price: Money{Currency: "USD", Value: "3.99"},
			},
		},
		Game:     "Magic: The Gathering",
		CardName: "Lightning Bolt",
		SetName:  "Magic 2010",
		Condition: TradingCardCondition{
			UngradedConditionDescriptorID: "400010",
		},
	}

	item, err := BuildInventoryItemFromDraft(draft)
	if err != nil {
		t.Fatalf("BuildInventoryItemFromDraft: %v", err)
	}
	if item.Product == nil || item.Product.Title != "4x Lightning Bolt" {
		t.Fatalf("product = %#v", item.Product)
	}
	if item.Availability == nil || item.Availability.ShipToLocationAvailability == nil || item.Availability.ShipToLocationAvailability.Quantity != 4 {
		t.Fatalf("availability = %#v", item.Availability)
	}

	offer, err := BuildOfferFromDraft(draft)
	if err != nil {
		t.Fatalf("BuildOfferFromDraft: %v", err)
	}
	if offer.AvailableQuantity != 4 {
		t.Fatalf("available quantity = %d", offer.AvailableQuantity)
	}
	if offer.Format != "FIXED_PRICE" {
		t.Fatalf("format = %q", offer.Format)
	}
	if offer.ListingDuration != "GTC" {
		t.Fatalf("duration = %q", offer.ListingDuration)
	}
	if offer.ListingDescription != "Four copies of a sandbox common" {
		t.Fatalf("listing description = %q", offer.ListingDescription)
	}
}

func TestBuildInventoryItemFromSealedProductDraft(t *testing.T) {
	t.Parallel()

	item, err := BuildInventoryItemFromDraft(SealedProductDraft{
		BaseDraft: BaseDraft{
			SKU:         "sealed-1",
			Description: "Factory sealed booster box",
			Quantity:    2,
		},
		Game:          "Magic: The Gathering",
		SetName:       "Bloomburrow",
		ProductName:   "Collector Booster Box",
		Configuration: "Collector Booster Box",
		Brand:         "Wizards of the Coast",
		UPC:           "123456789012",
	})
	if err != nil {
		t.Fatalf("BuildInventoryItemFromDraft: %v", err)
	}
	if item.Condition != "NEW" {
		t.Fatalf("condition = %q", item.Condition)
	}
	if item.Product == nil || item.Product.Title != "Bloomburrow Collector Booster Box" {
		t.Fatalf("product = %#v", item.Product)
	}
	if item.Product.UPC != "123456789012" || item.Product.Brand != "Wizards of the Coast" {
		t.Fatalf("product = %#v", item.Product)
	}
	if item.Product.Aspects["Configuration"][0] != "Collector Booster Box" {
		t.Fatalf("aspects = %#v", item.Product.Aspects)
	}
}

func TestBuildInventoryItemFromGenericCollectibleDraft(t *testing.T) {
	t.Parallel()

	item, err := BuildInventoryItemFromDraft(GenericCollectibleDraft{
		BaseDraft: BaseDraft{
			SKU:         "generic-1",
			Description: "Loose collectible",
			Quantity:    1,
			ExtraAspects: map[string][]string{
				"Theme": {"Retro"},
			},
		},
		Title:     "Vintage Counter Display",
		Condition: "USED_GOOD",
	})
	if err != nil {
		t.Fatalf("BuildInventoryItemFromDraft: %v", err)
	}
	if item.Product == nil || item.Product.Title != "Vintage Counter Display" {
		t.Fatalf("product = %#v", item.Product)
	}
	if item.Condition != "USED_GOOD" {
		t.Fatalf("condition = %q", item.Condition)
	}
}

func TestBuildInventoryItemFromDraftRejectsMissingCardCondition(t *testing.T) {
	t.Parallel()

	_, err := BuildInventoryItemFromDraft(SingleCardDraft{
		BaseDraft: BaseDraft{SKU: "single-1"},
		Game:      "Magic: The Gathering",
		CardName:  "Black Lotus",
	})
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestBuildOfferFromDraftRejectsMissingListingRefs(t *testing.T) {
	t.Parallel()

	_, err := BuildOfferFromDraft(GenericCollectibleDraft{
		BaseDraft: BaseDraft{SKU: "generic-1"},
		Title:     "Loose collectible",
		Condition: "USED_GOOD",
	})
	if err == nil {
		t.Fatalf("expected error")
	}
}
