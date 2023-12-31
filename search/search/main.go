package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"gopkg.in/olivere/elastic.v5"
)

// Elasticsearch configuration
const (
	elasticsearchURL   = "http://localhost:9200"
	elasticsearchIndex = "product-index"
)

type Product struct {
	ID          int64   `json:"index_id"`
	Name        string  `json:"name"`
	SalePrice   float64 `json:"sale_price"`
	MarketPrice float64 `json:"market_price"`
	Type        string  `json:"type"`
	Quantity    int     `json:"quantity"`
	Category    string  `json:"category"`
	SubCategory string  `json:"sub_category"`
	Brand       string  `json:"brand"`
	Rating      float64 `json:"rating"`
	ImageUrl    string  `json:"image_url"`
	ProductUrl  string  `json:"product_url"`
	Description string  `json:"description"`
	IsAvailable bool    `json:"is_available"`
}

func main() {
	// Initialize the Elasticsearch client
	client, err := elastic.NewClient(elastic.SetURL(elasticsearchURL), elastic.SetSniff(false))
	if err != nil {
		log.Fatalf("Error creating Elasticsearch client: %v", err)
	}

	// Create an HTTP handler to insert multiple documents into Elasticsearch
	http.HandleFunc("/insertMany", func(w http.ResponseWriter, r *http.Request) {
		// Parse the request body into a slice of Document structs
		// var docs []Product
		// if err := json.NewDecoder(r.Body).Decode(&docs); err != nil {
		// 	http.Error(w, fmt.Sprintf("Error decoding JSON: %v", err), http.StatusBadRequest)
		// 	return
		// }

		// Create an Elasticsearch context
		ctx := context.Background()

		// Use the Bulk API to insert multiple documents in a single request
		bulk := client.Bulk()

		// Path to the CSV file.
		csvFilePath := "dataset/bigbasket_products.csv"

		// Read the CSV file and parse its contents.
		objects, err := readCSVFile(csvFilePath)
		if err != nil {
			fmt.Printf("Error reading CSV file: %v\n", err)
			return
		}

		// req := elastic.NewBulkIndexRequest().Index(elasticsearchIndex).Doc(docs)
		// bulk = bulk.Add(req)

		for _, doc := range objects {
			req := elastic.NewBulkIndexRequest().Index(elasticsearchIndex).Doc(doc)
			bulk = bulk.Add(req)
		}

		_, err = bulk.Do(ctx)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error indexing documents: %v", err), http.StatusInternalServerError)
			return
		}

		// Respond with a success message
		w.WriteHeader(http.StatusCreated)
		fmt.Fprintln(w, "Documents indexed successfully")
	})

	// Start the HTTP server
	fmt.Println("Server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// Function to read a CSV file and parse its contents into a slice of MyObject.
func readCSVFile(filePath string) ([]Product, error) {
	var objects []Product

	rand.Seed(time.Now().UnixNano())

	// Open the CSV file for reading.
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Create a CSV reader.
	reader := csv.NewReader(file)

	// Read and parse the CSV file line by line.
	for {
		// Read a single row from the CSV.
		row, err := reader.Read()
		if err == io.EOF {
			break // End of file
		} else if err != nil {
			return nil, err
		}

		// Parse the row and create a Product instance.
		var id int64
		var sPrice, mPrice, rating float64
		if id, err = strconv.ParseInt(row[0], 10, 64); err != nil {
			continue
		}
		if sPrice, err = strconv.ParseFloat(row[5], 64); err != nil {
			continue
		}
		if mPrice, err = strconv.ParseFloat(row[6], 64); err != nil {
			continue
		}
		if rating, err = strconv.ParseFloat(row[11], 64); err != nil {
			rating = 2.5
		}
		qty := rand.Intn(35) + 5

		object := Product{
			ID:          id + 1,
			Name:        row[1],
			SalePrice:   sPrice,
			MarketPrice: mPrice,
			Category:    row[2],
			SubCategory: row[3],
			Brand:       row[4],
			Type:        row[9],
			Rating:      rating,
			ImageUrl:    row[7],
			ProductUrl:  row[8],
			Description: row[12],
			Quantity:    qty,
			IsAvailable: true,
		}

		// Append the object to the slice.
		objects = append(objects, object)
	}

	return objects, nil
}
