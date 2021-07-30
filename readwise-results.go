package main

import "time"

type ReadwiseBooks struct {
	Count    int           `json:"count"`
	Next     string        `json:"next"`
	Previous interface{}   `json:"previous"`
	Results  []Book `json:"results"`
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
}

type ReadwiseHighlights struct {
	Count    int           `json:"count"`
	Next     string        `json:"next"`
	Previous interface{}   `json:"previous"`
	Results  []Highlight `json:"results"`
}

type Highlight struct {
	Id            int         `json:"id"`
	Text          string      `json:"text"`
	Note          string      `json:"note"`
	Location      int         `json:"location"`
	LocationType  string      `json:"location_type"`
	HighlightedAt interface{} `json:"highlighted_at"`
	Url           interface{} `json:"url"`
	Color         string      `json:"color"`
	Updated       time.Time   `json:"updated"`
	BookId        int         `json:"book_id"`
	Tags          []struct {
		Id   int    `json:"id"`
		Name string `json:"name"`
	} `json:"tags"`
}