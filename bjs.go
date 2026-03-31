package main

import (
	"encoding/json"
	"fmt"

	azuretls "github.com/Noooste/azuretls-client"
)

const (
	inventoryURL = "https://api.bjs.com/digital/live/api/v1.2/inventory/club"
	priceURL     = "https://api.bjs.com/digital/live/api/v1.0/product/price"
)

var commonHeaders = azuretls.OrderedHeaders{
	{"accept", "application/json, text/plain, */*"},
	{"accept-language", "en-US,en;q=0.9"},
	{"cache-control", "no-cache"},
	{"origin", "https://www.bjs.com"},
	{"pragma", "no-cache"},
	{"referer", "https://www.bjs.com/"},
	{"sec-ch-ua", `"Chromium";v="146", "Not-A.Brand";v="24", "Google Chrome";v="146"`},
	{"sec-ch-ua-mobile", "?0"},
	{"sec-ch-ua-platform", `"macOS"`},
	{"sec-fetch-dest", "empty"},
	{"sec-fetch-mode", "cors"},
	{"sec-fetch-site", "same-site"},
	{"user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/146.0.0.0 Safari/537.36"},
}

type BJSClient struct {
	session *azuretls.Session
}

func NewBJSClient() *BJSClient {
	session := azuretls.NewSession()
	return &BJSClient{session: session}
}

func (c *BJSClient) Close() {
	c.session.Close()
}

// CheckInventory checks product availability at a specific club.
func (c *BJSClient) CheckInventory(storeID int, articleID, clubID string) (string, int, error) {
	reqBody := fmt.Sprintf(`{"InventoryMicroServiceEnableSwitch":"ON","Body":{"GetInventoryAvailability":{"ApplicationArea":{"BusinessContext":{"ContextData":[{"name":"storeId","text":%d},{"name":"langId","text":"-1"}]}},"PartNumber":"%s","uom":"CS","ExternalIdentifier":"%s"}}}`,
		storeID, articleID, clubID)

	headers := commonHeaders.Clone()
	headers.Add("content-type", "application/json")

	resp, err := c.session.Post(inventoryURL, reqBody, headers)
	if err != nil {
		return "", 0, fmt.Errorf("inventory request: %w", err)
	}

	if resp.StatusCode != 200 {
		return "", 0, fmt.Errorf("inventory API returned %d", resp.StatusCode)
	}

	var result struct {
		Body struct {
			ShowInventoryAvailability struct {
				DataArea struct {
					InventoryAvailability struct {
						InventoryStatus   string `json:"InventoryStatus"`
						AvailableQuantity int    `json:"AvailableQuantity"`
					} `json:"InventoryAvailability"`
				} `json:"DataArea"`
			} `json:"ShowInventoryAvailability"`
		} `json:"Body"`
	}

	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return "", 0, fmt.Errorf("decoding inventory response: %w (body: %s)", err, truncate(resp.Body, 200))
	}

	inv := result.Body.ShowInventoryAvailability.DataArea.InventoryAvailability
	return inv.InventoryStatus, inv.AvailableQuantity, nil
}

// GetPrice fetches the current club price for a product.
func (c *BJSClient) GetPrice(storeID int, productID, clubID string) (float64, error) {
	url := fmt.Sprintf("%s/%d?productId=%s&pageName=PDP&clubId=%s", priceURL, storeID, productID, clubID)

	resp, err := c.session.Get(url, commonHeaders)
	if err != nil {
		return 0, fmt.Errorf("price request: %w", err)
	}

	if resp.StatusCode != 200 {
		return 0, fmt.Errorf("price API returned %d", resp.StatusCode)
	}

	var result struct {
		BJSClubProduct []struct {
			InClubOfferPrice struct {
				Amount float64 `json:"amount"`
			} `json:"inClubOfferPrice"`
		} `json:"bjsClubProduct"`
	}

	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return 0, fmt.Errorf("decoding price response: %w (body: %s)", err, truncate(resp.Body, 200))
	}

	if len(result.BJSClubProduct) == 0 {
		return 0, fmt.Errorf("no price data returned")
	}

	return result.BJSClubProduct[0].InClubOfferPrice.Amount, nil
}

func truncate(b []byte, n int) string {
	if len(b) <= n {
		return string(b)
	}
	return string(b[:n]) + "..."
}
