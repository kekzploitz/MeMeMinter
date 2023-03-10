package transactions

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

func BalanceCheck(userId int64, mongoUri string) (bool, int, string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoUri))

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
			fmt.Println("updating users balance")
		}
		return false, 0, "failed"
	}

	return true, result.Balance, "success"
}
