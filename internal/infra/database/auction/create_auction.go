package auction

import (
	"context"
	"sync"
	"time"

	"github.com/ankardo/Lab-Leilao/configuration/auction_config"
	"github.com/ankardo/Lab-Leilao/internal/entity/auction_entity"
	"github.com/ankardo/Lab-Leilao/internal/internal_error"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type AuctionEntityMongo struct {
	Id             string                          `bson:"_id"`
	ProductName    string                          `bson:"product_name"`
	Category       string                          `bson:"category"`
	Description    string                          `bson:"description"`
	Condition      auction_entity.ProductCondition `bson:"condition"`
	Status         auction_entity.AuctionStatus    `bson:"status"`
	Timestamp      int64                           `bson:"timestamp"`
	ExpirationTime int64                           `bson:"expiration_time"`
}

type AuctionRepository struct {
	Collection            *mongo.Collection
	auctionStatusMap      map[string]auction_entity.AuctionStatus
	auctionEndTimeMap     map[string]time.Time
	auctionStatusMapMutex *sync.Mutex
	auctionEndTimeMutex   *sync.Mutex
}

func NewAuctionRepository(database *mongo.Database) *AuctionRepository {
	return &AuctionRepository{
		Collection:            database.Collection("auctions"),
		auctionStatusMap:      make(map[string]auction_entity.AuctionStatus),
		auctionEndTimeMap:     make(map[string]time.Time),
		auctionStatusMapMutex: &sync.Mutex{},
		auctionEndTimeMutex:   &sync.Mutex{},
	}
}

func (ar *AuctionRepository) CreateAuction(
	ctx context.Context,
	auctionEntity *auction_entity.Auction,
) *internal_error.InternalError {
	ar.auctionStatusMapMutex.Lock()
	defer ar.auctionStatusMapMutex.Unlock()

	_, exists := ar.auctionStatusMap[auctionEntity.Id]
	if exists {
		return internal_error.NewBadRequestError("Auction with this ID already exists")
	}

	auctionEntity.Status = auction_entity.Active
	expirationTime := auctionEntity.Timestamp.Add(auction_config.GetAuctionInterval())

	ar.auctionStatusMap[auctionEntity.Id] = auction_entity.Active
	ar.auctionEndTimeMap[auctionEntity.Id] = expirationTime

	auctionDocument := bson.M{
		"_id":             auctionEntity.Id,
		"productName":     auctionEntity.ProductName,
		"category":        auctionEntity.Category,
		"description":     auctionEntity.Description,
		"condition":       auctionEntity.Condition,
		"status":          auction_entity.Active,
		"timestamp":       auctionEntity.Timestamp.Unix(),
		"expiration_time": expirationTime.Unix(),
	}

	_, err := ar.Collection.InsertOne(ctx, auctionDocument)
	if err != nil {
		return internal_error.NewInternalServerError("Error inserting auction")
	}

	return nil
}
