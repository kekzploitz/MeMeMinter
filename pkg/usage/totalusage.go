package usage

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
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

func UpdateDb(userId int64, mongoUri string, beamCharge int, beamTxFee int, spent int, minted int) {
	newSpent := spent + beamCharge
	newMinted := minted + 1

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))

	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

	podcastsCollection := client.Database("minter").Collection("users")
	result, err := podcastsCollection.UpdateOne(
		ctx,
		bson.M{"user_id": userId}, // comment is userid from BEAM tx
		bson.D{
			{Key: "$set", Value: bson.D{{Key: "spent", Value: newSpent}, {Key: "minted", Value: newMinted}}},
		},
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Updated %v Documents!\n", result.ModifiedCount)
}

func CurrentValues(userId int64, mongoUri string, beamCharge int, beamTxFee int) {
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
			fmt.Println("updating users balance")
		}
	}
	UpdateDb(userId, mongoUri, beamCharge, beamTxFee, result.Spent, result.Minted)
}

func UpdateUsage(userId int64, mongoUri string, beamCharge int, beamTxFee int) {
	CurrentValues(userId, mongoUri, beamCharge, beamTxFee)
}
