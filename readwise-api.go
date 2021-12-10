package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type ReadwiseBooks struct {
	Count    int         `json:"count"`
	Next     string      `json:"next"`
	Previous interface{} `json:"previous"`
	Results  []Book      `json:"results"`
}

type Book struct {
	Id              int           `json:"id"`
	Title           string        `json:"title"`
	Author          string        `json:"author"`
	Category        string        `json:"category"`
	NumHighlights   int           `json:"num_highlights"`
	LastHighlightAt time.Time     `json:"last_highlight_at"`
	Updated         time.Time     `json:"updated"`
	CoverImageUrl   string        `json:"cover_image_url"`
	HighlightsUrl   string        `json:"highlights_url"`
	SourceUrl       interface{}   `json:"source_url"`
	Asin            string        `json:"asin"`
	Tags            []interface{} `json:"tags"`
	MemURL          string        `json:"mem_url"`
	MemId           string        `json:"mem_id"`
}

type ReadwiseHighlights struct {
	Count    int         `json:"count"`
	Next     string      `json:"next"`
	Previous interface{} `json:"previous"`
	Results  []Highlight `json:"results"`
}

type Highlight struct {
	Id            int       `json:"id"`
	Text          string    `json:"text"`
	Note          string    `json:"note"`
	Location      int       `json:"location"`
	LocationType  string    `json:"location_type"`
	HighlightedAt time.Time `json:"highlighted_at"`
	Url           string    `json:"url"`
	Color         string    `json:"color"`
	Updated       time.Time `json:"updated"`
	BookId        int       `json:"book_id"`
	Tags          []struct {
		Id   int    `json:"id"`
		Name string `json:"name"`
	} `json:"tags"`
}

func GetHighlights(ctx Context) []Highlight {
	// Get highlights (GET https://readwise.io/api/v2/highlights/)

	// Create client
	client := &http.Client{}
	var apiResult ReadwiseHighlights
	var result []Highlight

	// Fetch Request
	done := false
	next := "https://readwise.io/api/v2/highlights/"

	lastUpdate, update := GetTimeFromCache(ctx)
	for !done {
		req, err := http.NewRequest("GET", next, nil)
		req.Header.Add("Authorization", "Token "+ctx.config.ReadwiseKey)
		if update {
			q := req.URL.Query()
			q.Add("updated__gt", lastUpdate)
			req.URL.RawQuery = q.Encode()
			//			log.Println("get highlights URL: ", req.URL.String())
		}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("Failure : ", err)
		}
		respBody, _ := ioutil.ReadAll(resp.Body)
		err = json.Unmarshal(respBody, &apiResult)
		result = append(result, apiResult.Results...)
		log.Print("Got ", len(result), " of ", apiResult.Count, " highlights")

		if apiResult.Next == "" || apiResult.Next == next {
			done = true
		} else {
			next = apiResult.Next
		}

	}

	// Read Response Body
	return result
}

func GetBooks(ctx Context) []Book {
	// Get Books (GET https://readwise.io/api/v2/books/)

	// Create client
	client := &http.Client{}
	var apiResult ReadwiseBooks
	var result []Book

	// Fetch Request
	done := false
	next := "https://readwise.io/api/v2/books/"
	lastUpdate, update := GetTimeFromCache(ctx)
	for !done {
		req, err := http.NewRequest("GET", next, nil)
		req.Header.Add("Authorization", "Token "+ctx.config.ReadwiseKey)
		if update {
			q := req.URL.Query()
			q.Add("updated__gt", lastUpdate)
			req.URL.RawQuery = q.Encode()
			//			log.Println("get books URL: ", req.URL.String())
		}
		resp, err := client.Do(req)

		if err != nil {
			fmt.Println("Failure : ", err)
		}

		// Read Response Body
		respBody, _ := ioutil.ReadAll(resp.Body)
		err = json.Unmarshal(respBody, &apiResult)
		if err != nil {
			log.Print("Error unmarshalling highlights", err)
		}

		result = append(result, apiResult.Results...)
		log.Print("Got ", len(result), " of ", apiResult.Count, " books")

		if apiResult.Next == "" || apiResult.Next == next {
			done = true
		} else {
			next = apiResult.Next
		}
	}
	return result
}
