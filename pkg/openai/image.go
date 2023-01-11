package openai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Data struct {
	Url string `json:"url"`
}

type Response struct {
	Created int    `json:"created"`
	Data    []Data `json:"data"`
}

func NewImage(openaiApi string, openaiUrl string, payload string) (bool, string) {
	jsonData := fmt.Sprintf(`{
	  "prompt": "%s",
	  "n": 1,
	  "size": "1024x1024"
	}`, payload)

	request, err := http.NewRequest("POST", openaiUrl, bytes.NewBuffer([]byte(jsonData)))
	if err != nil {
		fmt.Println("error") // return meaningful statement
		return false, ""
	}

	request.Header.Set("Content-Type", "application/json; charset=UTF-8")
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", openaiApi))

	client := &http.Client{}
	res, err := client.Do(request)
	if err != nil {
		fmt.Println("error") // return meaningful statement
		return false, ""
	}
	// defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)

	data := Response{}
	_ = json.Unmarshal([]byte(body), &data)

	for _, url := range data.Data {
		return true, url.Url
	}

	return false, ""
}

func CreateImage(openaiApi string, openaiUrl string, payload string) (bool, string) {
	imageCreated, url := NewImage(openaiApi, openaiUrl, payload)
	if imageCreated {
		return true, url
	}
	return false, ""
}
