# Sealed eBay Research Handoff for ChatGPT

Use this prompt when you have sealed prerelease or oddball product photos and want a research handoff for a local Codex agent that will drive `github.com/repricah/ebay-tools`.

## Goal

Given the attached photos, identify the product and prepare a structured eBay draft handoff for a local coding agent. Do not assume this goes through `repricah.com`. This is for a local operator workflow using `ebay-tools`.

## What I Need Back

Return a concise handoff with these sections:

1. `Identification`
   - product name
   - game
   - set / release
   - item type
   - edition / variant details
   - confidence level
   - any unresolved ambiguity

2. `Suggested eBay Listing`
   - title
   - short description
   - condition
   - category guess
   - quantity
   - notable aspects / specifics
   - whether it should be treated as sealed product, promo kit, prerelease card bundle, or generic collectible

3. `Evidence`
   - OCR text you extracted
   - visual details that support the identification
   - if you infer anything, label it clearly as inference

4. `Image Checklist`
   - front photo needed?
   - back photo needed?
   - side / seal photo needed?
   - damage / wear close-up needed?
   - barcode / UPC needed?

5. `Agent Handoff JSON`
   - Return JSON only inside one fenced `json` block with this shape:

```json
{
  "product_type": "sealed_product | prerelease_card | promo_bundle | generic_collectible",
  "title": "",
  "description": "",
  "condition": "NEW | USED_LIKE_NEW | USED_VERY_GOOD | USED_GOOD | USED_ACCEPTABLE",
  "sku_hint": "",
  "quantity": 1,
  "category_hint": "",
  "aspects": {
    "Game": [],
    "Set": [],
    "Language": [],
    "Finish": [],
    "Card Size": [],
    "Features": []
  },
  "notes_for_agent": [],
  "open_questions": [],
  "confidence": "high | medium | low"
}
```

## Rules

- Prefer accuracy over forced completion.
- If the photos are insufficient, say exactly what is missing.
- Do not fabricate UPCs, release dates, card numbers, or marketplace category IDs.
- If the item appears too ambiguous for safe listing, say so.
- If this looks like a sealed prerelease promo or loose prerelease card not well suited for automated catalog matching, say that explicitly.
- Avoid long prose. The handoff should be something a local agent can act on.

## Follow-up for Local Codex Agent

After ChatGPT gives the handoff, the local agent should:

1. Validate the proposed product type.
2. Normalize the listing fields for `ebay-tools`.
3. Ask only for missing operator inputs such as:
   - final price
   - shipping policy choice
   - images to upload
4. Create an eBay draft in Sandbox first.
5. Show the final payload before publish.
