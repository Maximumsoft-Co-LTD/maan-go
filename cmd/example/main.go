package main

import (
	"context"
	"fmt"
	"log"

	"maan-go/internal/entities"
	repository "maan-go/internal/repo"
	"maan-go/pkg/mongo"

	"go.mongodb.org/mongo-driver/bson"
)

func main() {
	ctx := context.Background()
	mongoClient, err := mongo.NewClient(
		ctx,
		mongo.WithWriteURI("mongodb://localhost:27017"),
		mongo.WithReadURI("mongodb://localhost:27017"),
		mongo.WithDatabase("test"),
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
