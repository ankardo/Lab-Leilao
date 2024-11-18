package auction_usecase

import (
	"context"
	"time"

	"github.com/ankardo/Lab-Leilao/configuration/logger"
	"github.com/ankardo/Lab-Leilao/internal/entity/auction_entity"
	"github.com/ankardo/Lab-Leilao/internal/entity/bid_entity"
	"github.com/ankardo/Lab-Leilao/internal/internal_error"
	"github.com/ankardo/Lab-Leilao/internal/usecase/bid_usecase"
	"go.uber.org/zap"
)

type AuctionInputDTO struct {
	ProductName string           `json:"product_name" binding:"required,min=1"`
	Category    string           `json:"category" binding:"required,min=2"`
	Description string           `json:"description" binding:"required,min=10,max=200"`
	Condition   ProductCondition `json:"condition" binding:"oneof=0 1 2"`
}

type AuctionOutputDTO struct {
	Id             string           `json:"id"`
	ProductName    string           `json:"product_name"`
	Category       string           `json:"category"`
	Description    string           `json:"description"`
	Condition      ProductCondition `json:"condition"`
	Status         AuctionStatus    `json:"status"`
	Timestamp      time.Time        `json:"timestamp" time_format:"2006-01-02 15:04:05"`
	ExpirationTime time.Time        `json:"expiration_time" time_format:"2006-01-02 15:04:05"`
}

type WinningInfoOutputDTO struct {
	Auction AuctionOutputDTO          `json:"auction"`
	Bid     *bid_usecase.BidOutputDTO `json:"bid,omitempty"`
}

func NewAuctionUseCase(
	auctionRepositoryInterface auction_entity.AuctionRepositoryInterface,
	bidRepositoryInterface bid_entity.BidEntityRepository,
) AuctionUseCaseInterface {
	return &AuctionUseCase{
		auctionRepositoryInterface: auctionRepositoryInterface,
		bidRepositoryInterface:     bidRepositoryInterface,
	}
}

type AuctionUseCaseInterface interface {
	CreateAuction(
		ctx context.Context,
		auctionInput AuctionInputDTO) *internal_error.InternalError

	FindAuctionById(
		ctx context.Context, id string) (*AuctionOutputDTO, *internal_error.InternalError)

	FindAuctions(
		ctx context.Context,
		status AuctionStatus,
		category, productName string) ([]AuctionOutputDTO, *internal_error.InternalError)

	FindWinningBidByAuctionId(
		ctx context.Context,
		auctionId string) (*WinningInfoOutputDTO, *internal_error.InternalError)
}

type (
	ProductCondition int64
	AuctionStatus    int64
)

type AuctionUseCase struct {
	auctionRepositoryInterface auction_entity.AuctionRepositoryInterface
	bidRepositoryInterface     bid_entity.BidEntityRepository
}

func (au *AuctionUseCase) CreateAuction(
	ctx context.Context,
	auctionInput AuctionInputDTO,
) *internal_error.InternalError {
	auction, err := auction_entity.CreateAuction(
		auctionInput.ProductName,
		auctionInput.Category,
		auctionInput.Description,
		auction_entity.ProductCondition(auctionInput.Condition))
	if err != nil {
		return err
	}

	if err := au.auctionRepositoryInterface.CreateAuction(ctx, auction); err != nil {
		return err
	}

	go au.scheduleAuctionClosure(ctx, auction.Id, auction.ExpirationTime)

	return nil
}

func (au *AuctionUseCase) scheduleAuctionClosure(ctx context.Context, auctionID string, expirationTime time.Time) {
	duration := time.Until(expirationTime)
	if duration <= 0 {
		au.closeAuction(ctx, auctionID)
		return
	}

	timer := time.NewTimer(duration)
	<-timer.C

	au.closeAuction(ctx, auctionID)
}

func (au *AuctionUseCase) closeAuction(ctx context.Context, auctionID string) {
	if err := au.auctionRepositoryInterface.UpdateAuctionsStatus(
		ctx, []string{auctionID}, auction_entity.Completed); err != nil {

		logger.Error("Failed to close auction", err)
		return
	}

	logger.Info("Auction closed successfully", zap.String("auctionID", auctionID))
}
