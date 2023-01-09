package checks

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type Users struct {
	Id            int    `bson:"userid"`
	Alias         string `bson:"alias"`
	WalletAddress string `bson:"wallet_address"`
	PaymentId     string `bson:"payment_id"`
	Balance       int    `bson:"balance"`
}

func QueryDb(userId int64) (bool, string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))

	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

	collection := client.Database("minter").Collection("users")
	filter := bson.D{{Key: "user_id", Value: userId}}

	var result Users
	err = collection.FindOne(context.TODO(), filter).Decode(&result)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return false, "no documents found"
		}
		return false, "error fetching document"
	}
	return true, "document found"
}

func CheckUserExists(userId int64) (bool, string) {
	userExists, _ := QueryDb(userId)
	if userExists {
		return true, ""
	}
	return false, ""
}
