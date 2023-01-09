package start

import (
	"context"
	"fmt"
	"github.com/kekzploit/MeMeMinter/pkg/transactions"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

func commitUser(mongoUri string, firstName string, userId int64, address string) bool {

	// instantiate mongodb client
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoUri))

	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

	collection := client.Database("minter").Collection("users")

	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := collection.InsertOne(ctx, bson.D{
		{Key: "first_name", Value: firstName},
		{Key: "user_id", Value: userId},
		{Key: "created", Value: time.Now().Format("02-01-2006")},
		{Key: "spent", Value: 0},
		{Key: "minted", Value: 0},
		{Key: "address", Value: address},
		{Key: "balance", Value: 0}})

	id := res.InsertedID
	if id != "" {
		return true
	}
	return false
}

func Start(firstName string, userId int64, mongoUri string, walletApi string) (bool, string) {

	// create BEAM address
	addressCreated, address := transactions.CreateAddress(userId, walletApi)
	if addressCreated {
		committed := commitUser(mongoUri, firstName, userId, address)
		if committed {
			return true, address
		} else {
			fmt.Println("error committing user to ..") // return as string
			return false, "error committing user to.."
		}
	}
	fmt.Println(address) // return as string
	return false, "error creating BEAM address.."
}
