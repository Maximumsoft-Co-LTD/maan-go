package main

import (
	"context"
	"fmt"
	"log"

	"github.com/Maximumsoft-Co-LTD/maango"
	"github.com/Maximumsoft-Co-LTD/maango/internal/entities"
	repository "github.com/Maximumsoft-Co-LTD/maango/internal/repo"

	"go.mongodb.org/mongo-driver/bson"
)

func main() {
	ctx := context.Background()
	mongoClient, err := maango.NewClient(
		ctx,
		maango.WithWriteURI("mongodb://localhost:27017"),
		maango.WithReadURI("mongodb://localhost:27017"),
		maango.WithDatabase("test"),
	)
	if err != nil {
		log.Fatalf("Failed to create mongo client: %v", err)
	}
	defer mongoClient.Close()

	repo := repository.NewMongoRepo(ctx, mongoClient)
	defer repo.Close()

	bankConfig := repo.BankConfig(ctx)
	var v entities.BankConfig
	bankConfig.FindOne(bson.M{}).Res(&v)
	fmt.Println("Hello, World:", v)
}
