package transactions

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type UsersDb struct {
	Firtsname string `bson:"first_name"`
	Id        int    `bson:"user_id"`
	Created   string `bson:"created"`
	Spent     int    `bson:"spent"`
	Minted    int    `bson:"minted"`
	Address   string `bson:"address"`
	Balance   int    `bson:"balance"`
}

func GetAddress(userId int64, mongoUri string) (bool, string, string) {
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

	var result UsersDb
	err = collection.FindOne(context.TODO(), filter).Decode(&result)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			fmt.Println("failed to get address")
		}
		return false, "", "failed"
	}

	return true, result.Address, "success"
}
