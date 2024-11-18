package auction_config

import (
	"log"
	"os"
	"time"
)

func GetAuctionInterval() time.Duration {
	auctionInterval := os.Getenv("AUCTION_INTERVAL")
	duration, err := time.ParseDuration(auctionInterval)
	if err != nil {
		log.Printf("Error parsing AUCTION_INTERVAL: %v. Using default 5 minutes.", err)
		return 5 * time.Minute
	}

	return duration
}
