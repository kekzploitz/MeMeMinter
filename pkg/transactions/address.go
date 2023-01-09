package transactions

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

type AddressCreate struct {
	jsonpc string
	id     int
	result string
}

func CreateAddress(userId int64, walletApi string) (bool, string) {

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
}`, strconv.FormatInt(userId, 10))

	request, err := http.NewRequest("POST", walletApi, bytes.NewBuffer([]byte(jsonData)))
	if err != nil {
		fmt.Println("error") // return meaningful statement
	}
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")

	client := &http.Client{}
	res, err := client.Do(request)
	if err != nil {
		fmt.Println("error") // return meaningful statement
	}
	// defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)

	data := AddressCreate{}
	_ = json.Unmarshal([]byte(body), &data)

	if data.result != "" {
		return true, "address created successfully"
	}

	return false, "unable to create address"
}
