package auction

import (
	"context"

	"github.com/ankardo/Lab-Leilao/internal/entity/auction_entity"
	"github.com/ankardo/Lab-Leilao/internal/internal_error"
	"go.mongodb.org/mongo-driver/bson"
)

func (ar *AuctionRepository) UpdateAuctionsStatus(ctx context.Context, ids []string, status auction_entity.AuctionStatus) *internal_error.InternalError {
	filter := bson.M{
		"_id": bson.M{"$in": ids},
	}
	update := bson.M{
		"$set": bson.M{
			"status": status,
		},
	}

	_, err := ar.Collection.UpdateMany(ctx, filter, update)
	if err != nil {
		return internal_error.NewInternalServerError("Error updating auction statuses")
	}

	return nil
}
