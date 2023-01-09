package transactions

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type AddressCreate struct {
	JsonRpc string `json:"jsonrpc"`
	Id      int    `json:"id"`
	Result  string `json:"result"`
}

func CreateAddress(userId int64, walletApi string) (bool, string) {

	comment := fmt.Sprintf("User %v's address", userId)

	jsonData := fmt.Sprintf(`{
    "jsonrpc": "2.0", 
    "id": 1,
    "method": "create_address", 
    "params":
    {
        "type": "regular",
        "expiration": "never",
        "comment": "%s",
        "new_style_regular" : true
    }
}`, comment)

	request, err := http.NewRequest("POST", walletApi, bytes.NewBuffer([]byte(jsonData)))
	if err != nil {
		fmt.Println("error") // return meaningful statement
		return false, ""
	}
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")

	client := &http.Client{}
	res, err := client.Do(request)
	if err != nil {
		fmt.Println("error") // return meaningful statement
		return false, ""
	}
	// defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)

	data := AddressCreate{}
	_ = json.Unmarshal([]byte(body), &data)

	if data.Result != "" {
		return true, data.Result
	}

	return false, data.Result
}
