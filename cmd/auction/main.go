package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/ankardo/Lab-Leilao/configuration/database/mongodb"
	"github.com/ankardo/Lab-Leilao/configuration/logger"
	"github.com/ankardo/Lab-Leilao/internal/infra/api/web/controller/auction_controller"
	"github.com/ankardo/Lab-Leilao/internal/infra/api/web/controller/bid_controller"
	"github.com/ankardo/Lab-Leilao/internal/infra/api/web/controller/user_controller"
	"github.com/ankardo/Lab-Leilao/internal/infra/database/auction"
	"github.com/ankardo/Lab-Leilao/internal/infra/database/bid"
	"github.com/ankardo/Lab-Leilao/internal/infra/database/user"
	"github.com/ankardo/Lab-Leilao/internal/usecase/auction_usecase"
	"github.com/ankardo/Lab-Leilao/internal/usecase/bid_usecase"
	"github.com/ankardo/Lab-Leilao/internal/usecase/user_usecase"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	loadEnv()
	initLogger()

	database := initDatabase(ctx)
	defer func() {
		if err := database.Client().Disconnect(ctx); err != nil {
			logger.Error("Error disconnecting from MongoDB", err)
		}
	}()

	userController, bidController, auctionController := initDependencies(ctx, database)

	router := gin.Default()
	router.GET("/auction", auctionController.FindAuctions)
	router.GET("/auction/:auctionId", auctionController.FindAuctionById)
	router.POST("/auction", auctionController.CreateAuction)
	router.GET("/auction/winner/:auctionId", auctionController.FindWinningBidByAuctionId)
	router.POST("/bid", bidController.CreateBid)
	router.GET("/bid/:auctionId", bidController.FindBidByAuctionId)
	router.GET("/user/:userId", userController.FindUserById)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server error", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, os.Kill)
	<-quit

	logger.Info("Shutting down server...")

	ctxShutdown, cancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelShutdown()

	if err := srv.Shutdown(ctxShutdown); err != nil {
		logger.Error("Server forced to shutdown", err)
	} else {
		logger.Info("Server exited gracefully")
	}
}

func loadEnv() {
	if err := godotenv.Load("cmd/auction/.env"); err != nil {
		if err := godotenv.Load("/app/.env"); err != nil {
			logger.Error("Error loading environment variables", err)
			os.Exit(1)
		}
	}
}

func initDatabase(ctx context.Context) *mongo.Database {
	database, err := mongodb.NewMongoDBConnection(ctx)
	if err != nil {
		logger.Error("Error connecting to MongoDB", err)
		os.Exit(1)
	}
	return database
}

func initLogger() {
	gin.DefaultWriter = logger.GetZapWriter()
	gin.DefaultErrorWriter = logger.GetZapWriter()
	defer logger.GetZapLogger().Sync()
}

func initDependencies(
	ctx context.Context,
	database *mongo.Database,
) (
	userController *user_controller.UserController,
	bidController *bid_controller.BidController,
	auctionController *auction_controller.AuctionController,
) {
	auctionRepo := auction.NewAuctionRepository(database)
	bidRepo := bid.NewBidRepository(database, auctionRepo)
	userRepo := user.NewUserRepository(database)

	userController = user_controller.NewUserController(
		user_usecase.NewUserUseCase(userRepo),
	)
	auctionController = auction_controller.NewAuctionController(
		auction_usecase.NewAuctionUseCase(auctionRepo, bidRepo),
	)
	bidController = bid_controller.NewBidController(
		bid_usecase.NewBidUseCase(bidRepo),
	)
	return
}
