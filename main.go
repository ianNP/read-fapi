package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Response struct {
	Data []struct {
		Index    int    `json:"index"`
		ID       string `json:"id"`
		Name     string `json:"name"`
		Address  string `json:"address"`
		City     string `json:"city"`
		Postcode string `json:"postcode"`
		Phone    string `json:"phone"`
		Meter    int    `json:"meter"`
	} `json:"data"`
	Total      int `json:"total"`
	Count      int `json:"count"`
	Pagination struct {
		Next     interface{} `json:"next"`
		Previous interface{} `json:"previous"` // Is this acceptable? the 'Next' field should match
	} `json:"pagination"`
}

// Need to modularise this and remove all the code from main()

func main() {
	fmt.Println("Calling API...")
	client := &http.Client{}
	baseUrl := "http://127.0.0.1:8000"
	url := baseUrl + "/customers"
	onLastPage := false

	for {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			fmt.Print(err.Error())
		}
		req.Header.Add("Accept", "application/json")
		req.Header.Add("Content-Type", "application/json")
		resp, err := client.Do(req)
		if err != nil {
			fmt.Print(err.Error())
		}
		defer resp.Body.Close()
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Print(err.Error())
		}
		var responseObject Response
		json.Unmarshal(bodyBytes, &responseObject)
		data := responseObject.Data

		if onLastPage {
			break
		}
		pagination := responseObject.Pagination
		if pagination.Next == nil {
			// Need to GET one more time before exiting the loop
			onLastPage = true
		} else {
			url = baseUrl + pagination.Next.(string)
		}

		fmt.Printf("API Response as struct %+v\n", data)
	}
}
