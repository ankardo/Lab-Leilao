package integration

import (
	"net/http"
	"testing"
	"time"

	"github.com/ankardo/Lab-Leilao/configuration/auction_config"
	"github.com/ankardo/Lab-Leilao/tests/integration/helpers"
)

func TestHandleAuctionExpirationAutomatically(t *testing.T) {
	client := helpers.NewAPIClient("http://auction:8080", t)

	auction := map[string]interface{}{
		"product_name": "Expired Product",
		"category":     "Electronics",
		"description":  "Product description",
		"condition":    1,
		"timestamp":    time.Now().Add(-1 * time.Hour).Unix(),
	}

	resp := client.SendRequest(http.MethodPost, "/auction", auction)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status code 201, got %d", resp.StatusCode)
	}

	waitDuration := auction_config.GetAuctionInterval() + (5 * time.Second)
	time.Sleep(waitDuration)

	resp = client.SendRequest(http.MethodGet, "/auction?status=1", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status code 200, got %d", resp.StatusCode)
	}

	var auctions []map[string]interface{}
	client.ParseResponse(resp, &auctions)

	if len(auctions) == 0 {
		t.Errorf("Expected at least one auction to be closed, but found none.")
	}
}
