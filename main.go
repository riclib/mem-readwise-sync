package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/boltdb/bolt"
	"io/ioutil"
	"log"
	"net/http"
	"text/template"
)

const bookTemplate = `# {{.Title}}
#{{.Category}} #readwise
Author: {{.Author}}
![Cover]({{.CoverImageUrl}})
{{if .SourceUrl}} src: {{.SourceUrl}} {{end}}
`

type HighlightsOfBook struct {
	book      Book
	highlight []Highlight
}

var bookBucket = []byte("Books")

type Context struct {
	db        *bolt.DB
	templates map[string]*template.Template
}

func main() {

	var context Context
	db, err := bolt.Open("mem-readwise-sync.db", 0644, nil)
	if err != nil {
		log.Panic("Opening db", err)
	}
	context.db = db
	context.templates = make(map[string]*template.Template, 2)
	context.templates["book"] = template.Must(template.New("book").Parse(bookTemplate))

	syncBooks(context)
	syncHighlights(context)
}

func syncHighlights(context Context) {
	highlights := GetHighlights()
	newHighlightsOfBook := make(map[int]HighlightsOfBook)

	for _, highlight := range highlights {
		//		if i > 10 {break}
		book, found := GetBookFromCache(context, highlight.BookId)
		if !found {
			log.Println("Skipping highlight for book that wasn't synced", highlight.BookId)
		} else {
			bookHighlights, found := newHighlightsOfBook[highlight.BookId]
			var newHighlights []Highlight
			if !found { //First highlight of book for this sync
				newHighlights = make([]Highlight, 0)
				newHighlights = append(newHighlights, highlight)
			} else {
				newHighlights = append(bookHighlights.highlight, highlight)
			}

			newHighlightsOfBook[highlight.BookId] = HighlightsOfBook{
				book:      book,
				highlight: newHighlights,
			}
		}

	}
	for i, b := range newHighlightsOfBook {
		log.Println(i, b.book.Title)
		for j, h := range b.highlight {
			log.Println("---", j, h.Id, h.Text)
		}
	}

}

func syncBooks(context Context) {
	books := GetBooks()
	for i, book := range books {

		//		if i > 0 {
		//			break
		//		}

		_, found := GetBookFromCache(context, book.Id)
		//		log.Print("found", book)
		if !found {
			var buf bytes.Buffer
			err := context.templates["book"].Execute(&buf, book)
			if err != nil {
				log.Println("Executing Template", err)
			}
			/*
				book.MemURL, err = sendCreate(buf.String())
				if err != nil {
					log.Println("Sending Create Mem", err)
					continue
				}
			*/
			book.MemURL = fmt.Sprint(i)
			if book.MemURL != "" {
				err := AddBookToCache(context, book)
				if err != nil {
					log.Println("Couldn't add book", err, book)
				}

			}

		}
	}
}
func AddBookToCache(context Context, book Book) error {
	bookKey := []byte(fmt.Sprint(book.Id))
	bookJson, err := json.Marshal(book)
	if err != nil {
		log.Panic("Marshal book", err, book)
	}
	// Check if the book exists
	err = context.db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(bookBucket)
		if err != nil {
			log.Panic("CreateBucketIfNotExists", err)
		}
		err = bucket.Put(bookKey, bookJson)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		log.Println("Couldn't add book", err)
	}
	log.Println(book.Title)
	return err

}

func GetBookFromCache(context Context, id int) (Book, bool) {
	var found bool = false
	var foundBook Book
	bookKey := []byte(fmt.Sprint(id))
	// Check if the book exists
	err := context.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(bookBucket)
		if bucket != nil {
			bookBytes := bucket.Get(bookKey)
			if bookBytes != nil {
				found = true
				err := json.Unmarshal(bookBytes, &foundBook)
				if err != nil {
					log.Panic("Unmarshal book", err)
				}
			}
		}
		return nil
	})
	if err != nil {
		log.Panic("Read Book Transaction", err)
	}
	return foundBook, found

}

func GetHighlights() []Highlight {
	// Get highlights (GET https://readwise.io/api/v2/highlights/)

	// Create client
	client := &http.Client{}
	var apiResult ReadwiseHighlights
	var result []Highlight

	// Fetch Request
	done := false
	next := "https://readwise.io/api/v2/highlights/"

	for !done {
		req, err := http.NewRequest("GET", next, nil)
		req.Header.Add("Authorization", "Token t3Ns8c1LoVhSzZeOnXpcvYFe0tpKVqVsNrBdJRC5WzaJz5Tdg9")
		req.Header.Add("Cookie", "uniqueCookie=130232-1609617978")
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("Failure : ", err)
		}
		respBody, _ := ioutil.ReadAll(resp.Body)
		err = json.Unmarshal(respBody, &apiResult)
		result = append(result, apiResult.Results...)
		log.Print("Got ", len(result), " of ", apiResult.Count, " highlights")

		if apiResult.Next == next {
			done = true
		} else {
			next = apiResult.Next
		}

	}

	// Read Response Body
	return result
}

func GetBooks() []Book {
	// Get Books (GET https://readwise.io/api/v2/books/)

	// Create client
	client := &http.Client{}
	var apiResult ReadwiseBooks
	var result []Book

	// Fetch Request
	done := false
	next := "https://readwise.io/api/v2/books/"
	for !done {
		req, err := http.NewRequest("GET", next, nil)
		req.Header.Add("Authorization", "Token t3Ns8c1LoVhSzZeOnXpcvYFe0tpKVqVsNrBdJRC5WzaJz5Tdg9")
		req.Header.Add("Cookie", "uniqueCookie=130232-1609617978")
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

		if apiResult.Next == next {
			done = true
		} else {
			next = apiResult.Next
		}
	}
	return result
}

type memCreateApiInput struct {
	Content string `json:"content"`
}

type memCreateApiResponse struct {
	Url string `json:"url"`
}

func sendCreate(text string) (string, error) {
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
	req.Header.Add("Authorization", "ApiAccessToken 2b463e5a-4f63-43ba-9b70-2f6c47f6a700")
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
	return memResponse.Url, err
	// Display Results
}
