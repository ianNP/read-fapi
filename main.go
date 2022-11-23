package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

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

	ctx := context.Background()
	bqClient, err := bigquery.NewClient(ctx, "ian-meikle-playground")
	if err != nil {
		log.Fatal(err)
	}
	table := bqClient.Dataset("test_bq_api").Table("customer_insert_test_table_channeled")
	schema, err := bigquery.InferSchema(Customer{})
	if err != nil {
		log.Fatal(err)
	}
	if err := table.Create(ctx, &bigquery.TableMetadata{Schema: schema}); err != nil {
		// log.Fatalf("Table creation: %v\n", err)
		fmt.Printf("Table creation: %v\n", err)
	}

	start := time.Now()
	pageMax := 100

	ch := make(chan struct{})
	for index := 1; index <= pageMax; index++ {
		tempUrl := fmt.Sprintf("%v?page_num=%d&page_size=100", url, index)
		go func(tempUrl string, client *http.Client, table *bigquery.Table, ctx context.Context) {
			data := readAPI(tempUrl, client)
			writeRows(data, table, ctx)
			ch <- struct{}{}
		}(tempUrl, client, table, ctx)

	}
	for index := 1; index <= pageMax; index++ {
		<-ch
	}

	elapsed := time.Since(start)
	log.Printf("Program took %s", elapsed)
}

func readAPI(tUrl string, client *http.Client) []Customer {
	req, err := http.NewRequest("GET", tUrl, nil)
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
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Print(err.Error())
	}
	var responseObject Response
	json.Unmarshal(bodyBytes, &responseObject)
	return responseObject.Data
}

func writeRows(data []Customer, table *bigquery.Table, ctx context.Context) {
	err := table.Inserter().Put(ctx, data)
	if err != nil {
		// log.Fatal(err)
		fmt.Println(err) // Should handle this better
	}
	first := data[0].Index
	last := data[len(data)-1].Index
	fmt.Printf("Handling items from index %d to %d\n", first, last)
}
