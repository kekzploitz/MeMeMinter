package checks

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

func Balance(userId int64, mongoUri string, beamCharge int, beamTxFee int) bool {
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
			fmt.Println("unable to get balance")
			return false
		}
		return false
	}

	if result.Balance > (beamCharge + beamTxFee) {
		return true
	}
	return false
}
