package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"cloud.google.com/go/bigquery"
)

type Customer struct {
	Index    int    `json:"index"`
	ID       string `json:"id"`
	Name     string `json:"name"`
	Address  string `json:"address"`
	City     string `json:"city"`
	Postcode string `json:"postcode"`
	Phone    string `json:"phone"`
	Meter    int    `json:"meter"`
}

func (c Customer) String() string {
	cust := fmt.Sprintf("Index: %v\n", c.Index)
	cust += fmt.Sprintf("\tId: %v\n", c.ID)
	cust += fmt.Sprintf("\tName: %v\n", c.Name)
	contact := fmt.Sprintf("\tContact: %v, %v, %v, %v", c.Address, c.City, c.Postcode, c.Phone)
	contact = strings.ReplaceAll(contact, "\n", ", ")
	cust += contact + "\n"
	cust += fmt.Sprintf("\tMeter reading: %v\n", c.Meter)
	return cust
}

type Response struct {
	Data       []Customer `json:"data"`
	Total      int        `json:"total"`
	Count      int        `json:"count"`
	Pagination struct {
		Next     interface{} `json:"next"`     // This may return a nil
		Previous interface{} `json:"previous"` // This may return a nil
	} `json:"pagination"`
}

// Need to modularise this and remove all the code from main()

func main() {
	fmt.Println("Calling API...")
	client := &http.Client{}
	baseUrl := "http://127.0.0.1:8000"
	url := baseUrl + "/customers"
	onLastPage := false

	ctx := context.Background()
	bqClient, err := bigquery.NewClient(ctx, "ian-meikle-playground")
	if err != nil {
		log.Fatal(err)
	}
	table := bqClient.Dataset("test_bq_api").Table("customer_insert_test_table")
	schema, err := bigquery.InferSchema(Customer{})
	if err != nil {
		log.Fatal(err)
	}
	if err := table.Create(ctx, &bigquery.TableMetadata{Schema: schema}); err != nil {
		fmt.Printf("Table creation: %v\n", err)
	}

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

		err = table.Inserter().Put(ctx, data)
		fmt.Println(err) // Should handle this better

		// for _, c := range data {
		// 	fmt.Print(c)
		// }
	}
}
