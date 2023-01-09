package transactions

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"
)

type Users struct {
	Id            int    `bson:"userid"`
	Alias         string `bson:"alias"`
	WalletAddress string `bson:"wallet_address"`
	PaymentId     string `bson:"payment_id"`
	Balance       int    `bson:"balance"`
}

type Result struct {
	AssetId          int           `json:"asset_id"`
	Comment          string        `json:"comment"`
	Confirmations    int           `json:"confirmations"`
	CreateTime       int           `json:"create_time"`
	Fee              int           `json:"fee"`
	Height           int           `json:"height"`
	Income           bool          `json:"income"`
	Kernel           string        `json:"kernel"`
	Rates            []interface{} `json:"rates"`
	Receiver         string        `json:"receiver"`
	ReceiverIdentity string        `json:"receiver_identity"`
	Sender           string        `json:"sender"`
	SenderIdentity   string        `json:"sender_identity"`
	Status           int           `json:"status"`
	StatusString     string        `json:"status_string"`
	TxId             string        `json:"txId"`
	TxType           int           `json:"tx_type"`
	TxTypeString     string        `json:"tx_type_string"`
	Value            int           `json:"value"`
}

type TxCheck struct {
	JsonRpc string   `json:"jsonrpc"`
	Id      int      `json:"id"`
	Result  []Result `json:"result"`
}

func UpdateUserBalance(comment string, value int, balance int) {
	newBalance := balance + value

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
		bson.M{"user": strconv.Atoi(comment)}, // comment is userid from BEAM tx
		bson.D{
			{Key: "$set", Value: bson.D{{Key: "balance", Value: newBalance}}},
		},
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Updated %v Documents!\n", result.ModifiedCount)
}

func CommitTxDb(comment string, txId string, value int, balance int, mongoUri string) {
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
		{Key: "user_id", Value: strconv.Atoi(comment)},
		{Key: "tx_id", Value: txId},
		{Key: "spent", Value: value}})

	id := res.InsertedID
	if id != "" {
		UpdateUserBalance(comment, value, balance)
	}
}

func GetTxDb(comment string, txId string, value int, mongoUri string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoUri))

	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

	collection := client.Database("minter").Collection("transactions")
	filter := bson.D{{Key: "tx_id", Value: txId}}

	var result Users
	err = collection.FindOne(context.TODO(), filter).Decode(&result)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			CommitTxDb(comment, txId, value, result.Balance, mongoUri)
		}
	}
}

func GetTxsApi(walletApi string, mongoUri string) {

	jsonData := fmt.Sprintf(`{
    "jsonrpc":"2.0",
    "id": 8,
    "method":"tx_list",
    "params":
    {
        "filter" :
        {
            "status": 3
        },
        "rates": true,
        "skip" : 0,
        "count" : 10
    }
}`)

	request, err := http.NewRequest("POST", walletApi, bytes.NewBuffer([]byte(jsonData)))
	if err != nil {
		fmt.Println("error") // return meaningful statement
		return
	}
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")

	client := &http.Client{}
	res, err := client.Do(request)
	if err != nil {
		fmt.Println("error") // return meaningful statement
		return
	}
	// defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)

	data := TxCheck{}
	_ = json.Unmarshal([]byte(body), &data)

	for _, tx := range data.Result {
		if tx.Income && tx.AssetId == 0 {
			GetTxDb(tx.Comment, tx.TxId, tx.Value, mongoUri)
		}
	}
}

func MonitorTx(walletApi string, mongoUri string) {
	for {
		GetTxsApi(walletApi, mongoUri)
		time.Sleep(60 * time.Second)
	}
}
