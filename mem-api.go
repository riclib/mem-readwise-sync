package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

type memCreateApiInput struct {
	Content string `json:"content"`
}

type memCreateApiResponse struct {
	Url string `json:"url"`
	Id  string `json:"id"`
}

func sendCreate(ctx Context, text string) (memCreateApiResponse, error) {
	// Create (POST https://api.mem.ai/v0/mems)

	input := memCreateApiInput{Content: text}
	js, err := json.Marshal(input)
	if err != nil {
		log.Println("Generating Json for API Call:", err)
	}

	body := bytes.NewBuffer(js)

	// Create client
	client := &http.Client{}

	// Create request
	req, err := http.NewRequest("POST", "https://api.mem.ai/v0/mems", body)

	// Headers
	req.Header.Add("Authorization", "ApiAccessToken "+ctx.config.MemKey)
	req.Header.Add("Content-Type", "application/json; charset=utf-8")

	// Fetch Request
	resp, err := client.Do(req)

	if err != nil {
		log.Println("Calling Mem Create", err)
	}

	// Read Response Body
	respBody, _ := ioutil.ReadAll(resp.Body)

	var memResponse memCreateApiResponse
	err = json.Unmarshal(respBody, &memResponse)
	if err != nil {
		log.Println("Unmarshall Mem response:", err)
	}
	return memResponse, err
	// Display Results
}

func sendAppend(ctx Context, memid string, text string) (memCreateApiResponse, error) {
	// Create (POST https://api.mem.ai/v0/mems)

	input := memCreateApiInput{Content: text}
	js, err := json.Marshal(input)
	if err != nil {
		log.Println("Generating Json for API Call:", err)
	}

	body := bytes.NewBuffer(js)

	// Create client
	client := &http.Client{}

	// Create request
	req, err := http.NewRequest("POST", "https://api.mem.ai/v0/mems/"+memid+"/append", body)

	// Headers
	req.Header.Add("Authorization", "ApiAccessToken "+ctx.config.MemKey)
	req.Header.Add("Content-Type", "application/json; charset=utf-8")

	// Fetch Request
	resp, err := client.Do(req)

	if err != nil {
		log.Println("Calling Mem Create", err)
	}

	// Read Response Body
	respBody, _ := ioutil.ReadAll(resp.Body)

	var memResponse memCreateApiResponse
	err = json.Unmarshal(respBody, &memResponse)
	if err != nil {
		log.Println("Unmarshall Mem response:", err)
	}
	return memResponse, err
	// Display Results
}
