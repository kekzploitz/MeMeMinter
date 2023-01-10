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
	"time"
)

type ExternalSend struct {
	TxId string `json:"txId"`
}

type SendExternal struct {
	Id         int          `json:"id"`
	Jsonrpc    string       `json:"jsonrpc"`
	ResultSend ExternalSend `json:"result"`
}

type Users struct {
	Id            int    `bson:"userid"`
	Alias         string `bson:"alias"`
	WalletAddress string `bson:"wallet_address"`
	PaymentId     string `bson:"payment_id"`
	Balance       int    `bson:"balance"`
}

func WithdrawToExternal(walletApi string, beamCharge int, beamTxFee int, walletAddress string, primaryAddress string) {
	jsonData := fmt.Sprintf(`{
		"jsonrpc":"2.0", 
		"id": 2,
		"method":"tx_send", 
		"params":
		{
			"value": %d,
			"fee": %d,
			"from": "387aa3176cab08823876432472b1b954315356bc203724623bb5fd624ed3948a80a",
			"address": "%s",
			"comment": "MeMe Mint External Withdrawal",
			"asset_id": 0,
			"offline": false
		}
	}`, beamCharge, beamTxFee, primaryAddress)

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

	data := SendExternal{}
	_ = json.Unmarshal([]byte(body), &data)

	fmt.Println(data)

	if data.ResultSend.TxId != "" {
		fmt.Println(data.ResultSend.TxId)
	} else {
		fmt.Println("tx failed")
	}

}

func UpdateBalance(userId int64, balance int, beamCharge int, beamTxFee int, mongoUri string, walletApi string, walletAddress string, primaryAddress string) {

	totalCharge := beamCharge + beamTxFee
	newBalance := balance - totalCharge

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoUri))

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
			{Key: "$set", Value: bson.D{{Key: "balance", Value: newBalance}}},
		},
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Updated %v Documents!\n", result.ModifiedCount)
	WithdrawToExternal(walletApi, beamCharge, beamTxFee, walletAddress, primaryAddress)
}

func GetBalance(userId int64, mongoUri string, beamCharge int, beamTxFee int, walletApi string, primaryAddress string) {
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
	}
	UpdateBalance(userId, result.Balance, beamCharge, beamTxFee, mongoUri, walletApi, result.WalletAddress, primaryAddress)
}

func DeductAndSendTx(userId int64, mongoUri string, beamCharge int, beamTxFee int, walletApi string, primaryAddress string) {
	GetBalance(userId, mongoUri, beamCharge, beamTxFee, walletApi, primaryAddress)
}
