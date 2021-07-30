package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"text/template"
)

const bookTemplate = `Book {{.Id}} {{.Title}}
`



var (

)
func main() {
	t:= template.Must(template.New("book").Parse(bookTemplate))
	books := GetBooks()
	for _, book := range books {
	err := t.Execute(os.Stdout,book)
	if err != nil {
		log.Println("Executing Template", err)
	}
	}
}

func GetBooks() []Book {
	// Get Books (GET https://readwise.io/api/v2/books/)

	// Create client
	client := &http.Client{}
	var apiResult  ReadwiseBooks
	var result []Book

	req, err := http.NewRequest("GET", "https://readwise.io/api/v2/books/", nil)

	// Headers
	req.Header.Add("Authorization", "Token t3Ns8c1LoVhSzZeOnXpcvYFe0tpKVqVsNrBdJRC5WzaJz5Tdg9")

	// Fetch Request
	resp, err := client.Do(req)

	if err != nil {
		fmt.Println("Failure : ", err)
	}

	// Read Response Body
	respBody, _ := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(respBody, &apiResult)
	result = apiResult.Results

	// Display Results
	fmt.Println("response Status : ", resp.Status)
	fmt.Println("response Headers : ", resp.Header)
//	fmt.Println("response Body : ", string(respBody))
	return result
}


