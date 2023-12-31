package main

// API to accept image and return product details

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"

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

type Request struct {
	ImageUrl string `json:"image_url"`
}

type Dish struct {
	Name     string   `json:"name"`
	Recipe   string   `json:"recipe"`
	Keywords []string `json:"keywords"` // Ingredients
}

type Response struct {
	Dishes   map[string]Dish `json:"dishes"`
	Products []Product       `json:"products"`
}

func main() {
	// Initialize the Elasticsearch client
	client, err := elastic.NewClient(elastic.SetURL(elasticsearchURL), elastic.SetSniff(false))
	if err != nil {
		log.Fatalf("Error creating Elasticsearch client: %v", err)
	}

	// Create an HTTP handler to search for documents in Elasticsearch
	http.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		// Parse the query string
		q := r.URL.Query().Get("query")
		if q == "" {
			http.Error(w, "Query string parameter 'query' is required", http.StatusBadRequest)
			return
		}

		products := search(client, q)
		if products == nil {
			http.Error(w, "Error searching Elasticsearch", http.StatusInternalServerError)
			return
		}

		// Return search results
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(products)

	})

	// Almightly Search
	http.HandleFunc("/khoj", func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		err := r.ParseMultipartForm(10 << 20) // 10 MB limit for the file size
		if err != nil {
			http.Error(w, "Error in MultiFormParsing", http.StatusInternalServerError)
			return
		}

		// Get the file from the form data.
		file, _, err := r.FormFile("image") // "file" is the field name in the form
		if err != nil {
			http.Error(w, "Error in file parsing", http.StatusInternalServerError)
			return
		}
		defer file.Close()

		// Create a unique file name for the uploaded image.
		//fileName := fmt.Sprintf("uploaded_image_%d%s", r.Context().Value(context), filepath.Ext(file.Filename))
		//filePath := filepath.Join("uploads", "_new") // You can customize the directory as needed

		// Create the file on the server.
		out, err := os.Create("image.jpg")
		if err != nil {
			http.Error(w, "Error while creating temp file", http.StatusInternalServerError)
			return
		}
		defer out.Close()

		// Copy the file from the request to the server.
		_, err = io.Copy(out, file)
		if err != nil {
			http.Error(w, "Error while saving temp file", http.StatusInternalServerError)
			return
		}
		var response Response

		response.Dishes = callVisionAPI("image.jpg")
		if response.Dishes == nil {
			http.Error(w, "Error calling vision API", http.StatusInternalServerError)
			return
		}

		dishList := getKeys(response.Dishes)

		response.Products = search(client, response.Dishes[dishList[0]].Keywords[0])
		if response.Products == nil {
			http.Error(w, "Error searching Elasticsearch", http.StatusInternalServerError)
			return
		}

		// Return search results
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)

	})

	// Start the HTTP server
	fmt.Println("Server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// func handleImageUpload(w http.ResponseWriter, r *http.Request) {
// 	// Check if the request method is POST.
// 	if r.Method != http.MethodPost {
// 		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
// 		return
// 	}

// 	// Create a file to store the uploaded image.
// 	file, err := os.Create("uploaded_image.jpg") // You can change the file name and extension as needed.
// 	if err != nil {
// 		http.Error(w, fmt.Sprintf("Error creating file: %v", err), http.StatusInternalServerError)
// 		return
// 	}
// 	defer file.Close()

// 	// Copy the request body (image data) to the file.
// 	_, err = io.Copy(file, r.Body)
// 	if err != nil {
// 		http.Error(w, fmt.Sprintf("Error copying image data: %v", err), http.StatusInternalServerError)
// 		return
// 	}

// 	// Respond with a success message.
// 	w.WriteHeader(http.StatusCreated)
// 	fmt.Fprintln(w, "Image uploaded successfully")
// }

func callVisionAPI(imagePath string) map[string]Dish {

	apiURL := "http://192.168.113.81:6000/detect"

	// Create a buffer to write the multipart request body
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// Add the image file as a part of the request
	file, err := os.Open(imagePath) //
	if err != nil {
		fmt.Println("Error opening file:", err)
		return nil
	}
	defer file.Close()

	part, err := writer.CreateFormFile("image", "image.jpg") // "image" is the form field name, and "image.jpg" is the file name
	if err != nil {
		fmt.Println("Error creating form file:", err)
		return nil
	}

	_, err = io.Copy(part, file)
	if err != nil {
		fmt.Println("Error copying file to part:", err)
		return nil
	}

	// Don't forget to close the multipart writer to finalize the request
	writer.Close()

	// Create a POST request with the multipart body
	req, err := http.NewRequest("POST", apiURL, &requestBody)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return nil
	}

	// Set the Content-Type header to indicate a multipart request
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Make the POST request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error making request:", err)
		return nil
	}
	defer resp.Body.Close()

	// Process the response as needed
	fmt.Println("Vision-Service Response Status Code:", resp.Status)

	// Parse the JSON response.
	dishes := make(map[string]Dish)
	var dishList []Dish
	if err := json.NewDecoder(resp.Body).Decode(&dishList); err != nil {
		fmt.Printf("Error parsing the JSON response: %v\n", err)
		return nil
	}

	for _, dish := range dishList {
		dishes[dish.Name] = dish
	}

	return dishes

}

// func orch(w http.ResponseWriter, r *http.Request) {

// 	if r.Method != http.MethodPost {
// 		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
// 		return
// 	}

// 	// Create a temp file to store the uploaded image.
// 	file, err := ioutil.TempFile("uploaded_image", ".jpg")
// 	if err != nil {
// 		http.Error(w, fmt.Sprintf("Error creating file: %v", err), http.StatusInternalServerError)
// 		return
// 	}

// 	defer os.Remove(file.Name())

// 	// Copy the request body (image data) to the file.
// 	_, err = io.Copy(file, r.Body)
// 	if err != nil {
// 		http.Error(w, fmt.Sprintf("Error copying image data: %v", err), http.StatusInternalServerError)
// 		return
// 	}

// 	var dishes map[string]Dish

// 	dishes = callVisionAPI("uploaded_image")
// 	if dishes == nil {
// 		http.Error(w, fmt.Sprintf("Error calling vision API"), http.StatusInternalServerError)
// 		return
// 	}

// 	//dishList := getKeys(dishes)

// 	products := callSearchService(dishes)

// 	response := Response{dishes, products}

// 	// Respond with a success message.
// 	w.WriteHeader(http.StatusCreated)
// 	json.NewEncoder(w).Encode(response)
// 	log.Println("Search uploaded successfully")

// }

func search(client *elastic.Client, queryString string) []Product {

	log.Printf("Query string: %s\n", queryString)

	// Create an Elasticsearch context
	ctx := context.Background()

	// Search for the query string in the "title" and "text" fields
	query := elastic.NewMultiMatchQuery(queryString, "name", "category", "sub_category", "brand", "type")

	log.Printf("Query: %v\n", query)
	// Execute the search request
	res, err := client.Search().
		Index(elasticsearchIndex).
		Query(query).
		Do(ctx)
	if err != nil {
		return nil
	}

	var product Product
	var products []Product

	for _, hit := range res.Hits.Hits {
		err := json.Unmarshal(*hit.Source, &product)
		if err != nil {
			fmt.Printf("Error unmarshalling product: %v\n", err)
			return nil
		}
		products = append(products, product)
	}

	fmt.Printf("Found a total of %d products\n", res.Hits.TotalHits.Value)
	fmt.Printf("Search-Service Response Status Code: %d\n", 200)

	return products
}

// function to return keys of a map
func getKeys(m map[string]Dish) []string {
	keys := make([]string, len(m))
	i := 0
	for k := range m {
		keys[i] = k
		i++
	}
	return keys
}
