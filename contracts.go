package ebaytools

import (
	"fmt"
	"strings"
)

const (
	tradingCardConditionNameID       = "40001"
	tradingCardGraderNameID          = "27501"
	tradingCardGradeNameID           = "27502"
	tradingCardCertificationNumberID = "27503"
	gradedTradingCardConditionEnum   = "LIKE_NEW"
	ungradedTradingCardConditionEnum = "USED_VERY_GOOD"
	defaultFixedPriceFormat          = "FIXED_PRICE"
	defaultGoodTilCanceledDuration   = "GTC"
)

type ListingDraft interface {
	listingDraft()
	baseDraft() BaseDraft
}

type BaseDraft struct {
	SKU          string
	Description  string
	Quantity     int
	ImageURLs    []string
	ExtraAspects map[string][]string
	Listing      ListingOptions
}

type ListingOptions struct {
	MarketplaceID         string
	CategoryID            string
	MerchantLocationKey   string
	ListingDuration       string
	ListingDescription    string
	QuantityLimitPerBuyer int
	Format                string
	Policies              PolicyRefs
	Price                 Money
}

type PolicyRefs struct {
	FulfillmentPolicyID string
	PaymentPolicyID     string
	ReturnPolicyID      string
}

type Money struct {
	Currency string
	Value    string
}

type TradingCardCondition struct {
	UngradedConditionDescriptorID string
	GraderID                      string
	GradeID                       string
	CertificationNumber           string
}

type SingleCardDraft struct {
	BaseDraft
	Title     string
	Game      string
	CardName  string
	SetName   string
	Language  string
	Finish    string
	Condition TradingCardCondition
}

type PlaysetDraft struct {
	BaseDraft
	Title     string
	Game      string
	CardName  string
	SetName   string
	Language  string
	Finish    string
	Condition TradingCardCondition
}

type SealedProductDraft struct {
	BaseDraft
	Title         string
	Game          string
	SetName       string
	ProductName   string
	Configuration string
	Brand         string
	UPC           string
	Condition     string
}

type GenericCollectibleDraft struct {
	BaseDraft
	Title     string
	Condition string
}

func (SingleCardDraft) listingDraft()         {}
func (PlaysetDraft) listingDraft()            {}
func (SealedProductDraft) listingDraft()      {}
func (GenericCollectibleDraft) listingDraft() {}

func (d SingleCardDraft) baseDraft() BaseDraft         { return d.BaseDraft }
func (d PlaysetDraft) baseDraft() BaseDraft            { return d.BaseDraft }
func (d SealedProductDraft) baseDraft() BaseDraft      { return d.BaseDraft }
func (d GenericCollectibleDraft) baseDraft() BaseDraft { return d.BaseDraft }

func BuildInventoryItemFromDraft(draft ListingDraft) (InventoryItem, error) {
	draft, err := normalizeListingDraft(draft)
	if err != nil {
		return InventoryItem{}, err
	}

	switch draft := draft.(type) {
	case SingleCardDraft:
		return buildSingleCardInventoryItem(draft, false)
	case PlaysetDraft:
		return buildSingleCardInventoryItem(singleCardFromPlayset(draft), true)
	case SealedProductDraft:
		return buildSealedProductInventoryItem(draft)
	case GenericCollectibleDraft:
		return buildGenericInventoryItem(draft)
	default:
		return InventoryItem{}, fmt.Errorf("unsupported listing draft type %T", draft)
	}
}

func BuildOfferFromDraft(draft ListingDraft) (Offer, error) {
	draft, err := normalizeListingDraft(draft)
	if err != nil {
		return Offer{}, err
	}

	base := draft.baseDraft()
	if strings.TrimSpace(base.SKU) == "" {
		return Offer{}, fmt.Errorf("sku is required")
	}

	listing := base.Listing
	if strings.TrimSpace(listing.MarketplaceID) == "" {
		return Offer{}, fmt.Errorf("marketplace id is required")
	}
	if strings.TrimSpace(listing.CategoryID) == "" {
		return Offer{}, fmt.Errorf("category id is required")
	}
	if strings.TrimSpace(listing.MerchantLocationKey) == "" {
		return Offer{}, fmt.Errorf("merchant location key is required")
	}
	if strings.TrimSpace(listing.Policies.FulfillmentPolicyID) == "" ||
		strings.TrimSpace(listing.Policies.PaymentPolicyID) == "" ||
		strings.TrimSpace(listing.Policies.ReturnPolicyID) == "" {
		return Offer{}, fmt.Errorf("all policy ids are required")
	}
	if strings.TrimSpace(listing.Price.Currency) == "" || strings.TrimSpace(listing.Price.Value) == "" {
		return Offer{}, fmt.Errorf("price currency and value are required")
	}

	quantity := normalizedQuantity(base.Quantity, 1)
	lotSize := 0
	switch draft.(type) {
	case PlaysetDraft:
		lotSize = 4
	}

	listingDescription := strings.TrimSpace(listing.ListingDescription)
	if listingDescription == "" {
		listingDescription = strings.TrimSpace(base.Description)
	}

	format := strings.TrimSpace(listing.Format)
	if format == "" {
		format = defaultFixedPriceFormat
	}
	duration := strings.TrimSpace(listing.ListingDuration)
	if duration == "" {
		duration = defaultGoodTilCanceledDuration
	}

	return Offer{
		SKU:                   strings.TrimSpace(base.SKU),
		MarketplaceID:         strings.TrimSpace(listing.MarketplaceID),
		Format:                format,
		LotSize:               lotSize,
		AvailableQuantity:     quantity,
		CategoryID:            strings.TrimSpace(listing.CategoryID),
		MerchantLocationKey:   strings.TrimSpace(listing.MerchantLocationKey),
		ListingDescription:    listingDescription,
		ListingDuration:       duration,
		QuantityLimitPerBuyer: listing.QuantityLimitPerBuyer,
		ListingPolicies: &ListingPolicies{
			FulfillmentPolicyID: strings.TrimSpace(listing.Policies.FulfillmentPolicyID),
			PaymentPolicyID:     strings.TrimSpace(listing.Policies.PaymentPolicyID),
			ReturnPolicyID:      strings.TrimSpace(listing.Policies.ReturnPolicyID),
		},
		PricingSummary: &PricingSummary{
			Price: &Amount{
				Currency: strings.TrimSpace(listing.Price.Currency),
				Value:    strings.TrimSpace(listing.Price.Value),
			},
		},
	}, nil
}

func buildSingleCardInventoryItem(draft SingleCardDraft, playset bool) (InventoryItem, error) {
	base := draft.BaseDraft
	if strings.TrimSpace(base.SKU) == "" {
		return InventoryItem{}, fmt.Errorf("sku is required")
	}
	if strings.TrimSpace(draft.CardName) == "" {
		return InventoryItem{}, fmt.Errorf("card name is required")
	}
	if strings.TrimSpace(draft.Game) == "" {
		return InventoryItem{}, fmt.Errorf("game is required")
	}

	condition, descriptors, err := draft.Condition.toEbayCondition()
	if err != nil {
		return InventoryItem{}, err
	}

	title := strings.TrimSpace(draft.Title)
	if title == "" {
		title = strings.TrimSpace(draft.CardName)
		if playset {
			title = "4x " + title
		}
	}

	quantity := normalizedQuantity(base.Quantity, 1)
	if playset {
		quantity = normalizedQuantity(base.Quantity, 1)
	}

	return InventoryItem{
		SKU:                  strings.TrimSpace(base.SKU),
		Condition:            condition,
		ConditionDescriptors: descriptors,
		Availability: &Availability{
			ShipToLocationAvailability: &ShipToLocationAvailability{
				Quantity: quantity,
			},
		},
		Product: &Product{
			Title:       title,
			Description: strings.TrimSpace(base.Description),
			Aspects: mergeAspects(
				base.ExtraAspects,
				map[string][]string{
					"Game":     nonEmptySlice(draft.Game),
					"Set":      nonEmptySlice(draft.SetName),
					"Language": nonEmptySlice(draft.Language),
					"Finish":   nonEmptySlice(draft.Finish),
				},
			),
			ImageURLs: append([]string(nil), base.ImageURLs...),
		},
	}, nil
}

func buildSealedProductInventoryItem(draft SealedProductDraft) (InventoryItem, error) {
	base := draft.BaseDraft
	if strings.TrimSpace(base.SKU) == "" {
		return InventoryItem{}, fmt.Errorf("sku is required")
	}
	if strings.TrimSpace(draft.ProductName) == "" {
		return InventoryItem{}, fmt.Errorf("product name is required")
	}

	title := strings.TrimSpace(draft.Title)
	if title == "" {
		title = strings.TrimSpace(strings.TrimSpace(draft.SetName) + " " + strings.TrimSpace(draft.ProductName))
		title = strings.TrimSpace(title)
	}
	condition := strings.TrimSpace(draft.Condition)
	if condition == "" {
		condition = "NEW"
	}

	return InventoryItem{
		SKU:       strings.TrimSpace(base.SKU),
		Condition: condition,
		Availability: &Availability{
			ShipToLocationAvailability: &ShipToLocationAvailability{
				Quantity: normalizedQuantity(base.Quantity, 1),
			},
		},
		Product: &Product{
			Title:       title,
			Description: strings.TrimSpace(base.Description),
			Aspects: mergeAspects(
				base.ExtraAspects,
				map[string][]string{
					"Game":          nonEmptySlice(draft.Game),
					"Set":           nonEmptySlice(draft.SetName),
					"Configuration": nonEmptySlice(draft.Configuration),
				},
			),
			Brand:     strings.TrimSpace(draft.Brand),
			UPC:       strings.TrimSpace(draft.UPC),
			ImageURLs: append([]string(nil), base.ImageURLs...),
		},
	}, nil
}

func buildGenericInventoryItem(draft GenericCollectibleDraft) (InventoryItem, error) {
	base := draft.BaseDraft
	if strings.TrimSpace(base.SKU) == "" {
		return InventoryItem{}, fmt.Errorf("sku is required")
	}
	if strings.TrimSpace(draft.Title) == "" {
		return InventoryItem{}, fmt.Errorf("title is required")
	}
	if strings.TrimSpace(draft.Condition) == "" {
		return InventoryItem{}, fmt.Errorf("condition is required")
	}

	return InventoryItem{
		SKU:       strings.TrimSpace(base.SKU),
		Condition: strings.TrimSpace(draft.Condition),
		Availability: &Availability{
			ShipToLocationAvailability: &ShipToLocationAvailability{
				Quantity: normalizedQuantity(base.Quantity, 1),
			},
		},
		Product: &Product{
			Title:       strings.TrimSpace(draft.Title),
			Description: strings.TrimSpace(base.Description),
			Aspects:     cloneAspects(base.ExtraAspects),
			ImageURLs:   append([]string(nil), base.ImageURLs...),
		},
	}, nil
}

func (c TradingCardCondition) toEbayCondition() (string, []ConditionDescriptor, error) {
	if strings.TrimSpace(c.GraderID) != "" || strings.TrimSpace(c.GradeID) != "" || strings.TrimSpace(c.CertificationNumber) != "" {
		if strings.TrimSpace(c.GraderID) == "" || strings.TrimSpace(c.GradeID) == "" {
			return "", nil, fmt.Errorf("graded card condition requires grader id and grade id")
		}

		descriptors := []ConditionDescriptor{
			{Name: tradingCardGraderNameID, Values: []string{strings.TrimSpace(c.GraderID)}},
			{Name: tradingCardGradeNameID, Values: []string{strings.TrimSpace(c.GradeID)}},
		}
		if info := strings.TrimSpace(c.CertificationNumber); info != "" {
			descriptors = append(descriptors, ConditionDescriptor{
				Name:           tradingCardCertificationNumberID,
				AdditionalInfo: info,
			})
		}
		return gradedTradingCardConditionEnum, descriptors, nil
	}

	if strings.TrimSpace(c.UngradedConditionDescriptorID) == "" {
		return "", nil, fmt.Errorf("trading card condition requires an ungraded descriptor id or graded details")
	}

	return ungradedTradingCardConditionEnum, []ConditionDescriptor{
		{
			Name:   tradingCardConditionNameID,
			Values: []string{strings.TrimSpace(c.UngradedConditionDescriptorID)},
		},
	}, nil
}

func normalizedQuantity(quantity int, defaultQuantity int) int {
	if quantity > 0 {
		return quantity
	}
	return defaultQuantity
}

func normalizeListingDraft(draft ListingDraft) (ListingDraft, error) {
	if draft == nil {
		return nil, fmt.Errorf("draft is nil")
	}

	switch draft := draft.(type) {
	case *SingleCardDraft:
		if draft == nil {
			return nil, fmt.Errorf("draft is nil")
		}
		return *draft, nil
	case *PlaysetDraft:
		if draft == nil {
			return nil, fmt.Errorf("draft is nil")
		}
		return *draft, nil
	case *SealedProductDraft:
		if draft == nil {
			return nil, fmt.Errorf("draft is nil")
		}
		return *draft, nil
	case *GenericCollectibleDraft:
		if draft == nil {
			return nil, fmt.Errorf("draft is nil")
		}
		return *draft, nil
	default:
		return draft, nil
	}
}

func singleCardFromPlayset(draft PlaysetDraft) SingleCardDraft {
	return SingleCardDraft{
		BaseDraft: draft.BaseDraft,
		Title:     draft.Title,
		Game:      draft.Game,
		CardName:  draft.CardName,
		SetName:   draft.SetName,
		Language:  draft.Language,
		Finish:    draft.Finish,
		Condition: draft.Condition,
	}
}

func nonEmptySlice(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return []string{strings.TrimSpace(value)}
}

func mergeAspects(maps ...map[string][]string) map[string][]string {
	merged := map[string][]string{}
	for _, input := range maps {
		for key, values := range input {
			cleanKey := strings.TrimSpace(key)
			if cleanKey == "" {
				continue
			}
			filtered := make([]string, 0, len(values))
			for _, value := range values {
				if trimmed := strings.TrimSpace(value); trimmed != "" {
					filtered = append(filtered, trimmed)
				}
			}
			if len(filtered) > 0 {
				merged[cleanKey] = filtered
			}
		}
	}
	if len(merged) == 0 {
		return nil
	}
	return merged
}

func cloneAspects(input map[string][]string) map[string][]string {
	return mergeAspects(input)
}
