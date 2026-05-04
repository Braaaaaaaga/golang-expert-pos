package auction

import (
	"context"
	"fullcycle-auction_go/internal/entity/auction_entity"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestAuctionAutoClose(t *testing.T) {
	ctx := context.Background()

	mongoContainer, err := mongodb.Run(ctx, "mongo:6")
	if err != nil {
		t.Skipf("skipping: could not start mongodb container (Docker unavailable?): %v", err)
	}
	defer mongoContainer.Terminate(ctx)

	connStr, err := mongoContainer.ConnectionString(ctx)
	if err != nil {
		t.Fatalf("failed to get connection string: %v", err)
	}

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(connStr))
	if err != nil {
		t.Fatalf("failed to connect to mongodb: %v", err)
	}
	defer client.Disconnect(ctx)

	os.Setenv("AUCTION_DURATION", "2s")

	repo := NewAuctionRepository(client.Database("auctions_test"))

	auctionEntity, validationErr := auction_entity.CreateAuction(
		"Product Name",
		"Category",
		"Description long enough here",
		auction_entity.New,
	)
	assert.Nil(t, validationErr)

	internalErr := repo.CreateAuction(ctx, auctionEntity)
	assert.Nil(t, internalErr)

	found, internalErr := repo.FindAuctionById(ctx, auctionEntity.Id)
	assert.Nil(t, internalErr)
	assert.Equal(t, auction_entity.Active, found.Status)

	time.Sleep(3 * time.Second)

	found, internalErr = repo.FindAuctionById(ctx, auctionEntity.Id)
	assert.Nil(t, internalErr)
	assert.Equal(t, auction_entity.Completed, found.Status)
}
