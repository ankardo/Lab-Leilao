package auction

import (
	"context"
	"fmt"
	"time"

	"github.com/ankardo/Lab-Leilao/configuration/logger"
	"github.com/ankardo/Lab-Leilao/internal/entity/auction_entity"
	"github.com/ankardo/Lab-Leilao/internal/internal_error"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

func (ar *AuctionRepository) findWithFilter(
	ctx context.Context,
	filter bson.M,
	operation string,
) ([]auction_entity.Auction, *internal_error.InternalError) {
	cursor, err := ar.Collection.Find(ctx, filter)
	if err != nil {
		logger.Error("Error during operation", err, zap.String("operation", operation))
		return nil, internal_error.NewInternalServerError("Error finding auctions during " + operation)
	}
	defer cursor.Close(ctx)

	var auctionsMongo []AuctionEntityMongo
	if err := cursor.All(ctx, &auctionsMongo); err != nil {
		logger.Error("Error decoding auctions", err, zap.String("operation", operation))
		return nil, internal_error.NewInternalServerError("Error decoding auctions during " + operation)
	}

	var auctionsEntity []auction_entity.Auction
	for _, auction := range auctionsMongo {
		auctionsEntity = append(auctionsEntity, auction_entity.Auction{
			Id:          auction.Id,
			ProductName: auction.ProductName,
			Category:    auction.Category,
			Status:      auction.Status,
			Description: auction.Description,
			Condition:   auction.Condition,
			Timestamp:   time.Unix(auction.Timestamp, 0),
		})
	}

	logger.Debug("Successfully fetched auctions", zap.String("operation", operation), zap.Int("count", len(auctionsEntity)))

	return auctionsEntity, nil
}

func (ar *AuctionRepository) FindAuctionById(
	ctx context.Context, id string,
) (*auction_entity.Auction, *internal_error.InternalError) {
	filter := bson.M{"_id": id}

	var auctionEntityMongo AuctionEntityMongo
	if err := ar.Collection.FindOne(ctx, filter).Decode(&auctionEntityMongo); err != nil {
		logger.Error(fmt.Sprintf("Error trying to find auction by id = %s", id), err)
		return nil, internal_error.NewInternalServerError("Error trying to find auction by id")
	}

	return &auction_entity.Auction{
		Id:          auctionEntityMongo.Id,
		ProductName: auctionEntityMongo.ProductName,
		Category:    auctionEntityMongo.Category,
		Description: auctionEntityMongo.Description,
		Condition:   auctionEntityMongo.Condition,
		Status:      auctionEntityMongo.Status,
		Timestamp:   time.Unix(auctionEntityMongo.Timestamp, 0),
	}, nil
}

func (ar *AuctionRepository) FindAuctions(
	ctx context.Context,
	status auction_entity.AuctionStatus,
	category string,
	productName string,
) ([]auction_entity.Auction, *internal_error.InternalError) {
	filter := bson.M{}

	if status != 0 {
		filter["status"] = status
	}

	if category != "" {
		filter["category"] = category
	}

	if productName != "" {
		filter["productName"] = primitive.Regex{Pattern: productName, Options: "i"}
	}

	return ar.findWithFilter(ctx, filter, "FindAuctions")
}

func (ar *AuctionRepository) FindExpiredAuctions(
	ctx context.Context,
	expirationTime time.Time,
) ([]auction_entity.Auction, *internal_error.InternalError) {
	filter := bson.M{
		"status":          auction_entity.Active,
		"expiration_time": bson.M{"$lt": expirationTime.Unix()},
	}

	return ar.findWithFilter(ctx, filter, "FindExpiredAuctions")
}
