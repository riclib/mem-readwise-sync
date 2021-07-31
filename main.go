package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/boltdb/bolt"
	"gopkg.in/alecthomas/kingpin.v2"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"text/template"
	"time"
)

const (
	lastSyncKey = "last_sync"
)

var (
	configFile = kingpin.Flag("config.file", "Configuration file.").ExistingFile()
)

type Config struct {
	ReadwiseKey       string `yaml:"readwise_key"`
	MemKey            string `yaml:"mem_key"`
	TimestampFormat   string `yaml:"timestamp_format"`
	BookTemplate      string `yaml:"book_template"`
	HighlightTemplate string `yaml:"highlight_template"`
}

type HighlightsOfBook struct {
	TimeStamp string
	Book      Book
	Highlight []Highlight
}

var (
	bookBucket       = []byte("Books")
	lastUpdateBucket = []byte("LastUpdate")
)

type ctx struct {
	db           *bolt.DB
	config       Config
	templates    map[string]*template.Template
	lastSyncTime string
	thisSyncTime string
}

func main() {
	kingpin.Version("mem-readwise-sync 1.0.0")
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	var context ctx

	if *configFile == "" {
		kingpin.Usage()
		os.Exit(1)
	}

	data, err := ioutil.ReadFile(*configFile)
	if err != nil {
		log.Panic("Error reading config", err)
	}

	if err := yaml.Unmarshal(data, &(context.config)); err != nil {
		log.Panic("Error parsing config", err)
	}

	db, err := bolt.Open("mem-readwise-sync.db", 0644, nil)
	if err != nil {
		log.Panic("Opening db", err)
	}
	context.db = db
	context.templates = make(map[string]*template.Template, 2)
	context.templates["book"] = template.Must(template.New("book").Parse(context.config.BookTemplate))
	context.templates["highlight"] = template.Must(template.New("highlight").Parse(context.config.HighlightTemplate))
	context.thisSyncTime = time.Now().UTC().Format(time.RFC3339)

	syncBooks(context)
	syncHighlights(context)
	AddTimeToCache(context, context.thisSyncTime)
}

func syncHighlights(ctx ctx) {
	highlights := GetHighlights(ctx)
	newHighlightsOfBook := make(map[int]HighlightsOfBook)
	lastUpdate, found := GetTimeFromCache(ctx)
	var t time.Time
	var err error
	if found {
		t, err = time.Parse(time.RFC3339, lastUpdate)
		if err != nil {
			log.Print("Couldn't parse last update", err)
		}
	} else {
		t = time.Unix(0, 0)
	}
	for _, highlight := range highlights {
		//		highlightedat := time.Parse(time.RFC3339, highlight.HighlightedAt)
		if highlight.HighlightedAt.Unix() < t.Unix() {
			continue
		}
		//		if i > 10 {break}
		book, found := GetBookFromCache(ctx, highlight.BookId)
		if !found {
			log.Println("Skipping highlight for book that wasn't synced", highlight.BookId)
		} else {
			bookHighlights, found := newHighlightsOfBook[highlight.BookId]
			var newHighlights []Highlight
			if !found { //First highlight of book for this sync
				newHighlights = make([]Highlight, 0)
				newHighlights = append(newHighlights, highlight)
			} else {
				newHighlights = append([]Highlight{highlight}, bookHighlights.Highlight...)
			}

			newHighlightsOfBook[highlight.BookId] = HighlightsOfBook{
				Book:      book,
				Highlight: newHighlights,
			}
		}

	}
	for _, b := range newHighlightsOfBook {
		b.TimeStamp = time.Now().Format(ctx.config.TimestampFormat)
		var buf bytes.Buffer
		err := ctx.templates["highlight"].Execute(&buf, b)
		if err != nil {
			log.Println("Executing Template", err)
		}
		// log.Println(buf.String())
		_, err = sendCreate(ctx, buf.String())
		if err != nil {
			log.Println("Sending Create Mem", err)
		} else {
			log.Println("Synced ", len(b.Highlight), " for ", b.Book.Title)
		}

	}

}

func syncBooks(ctx ctx) {
	books := GetBooks(ctx)
	for _, book := range books {

		//		if i > 0 {
		//			break
		//		}

		_, found := GetBookFromCache(ctx, book.Id)
		//		log.Print("found", book)
		if !found {
			var buf bytes.Buffer
			err := ctx.templates["book"].Execute(&buf, book)
			if err != nil {
				log.Println("Executing Template", err)
			}

			book.MemURL, err = sendCreate(ctx, buf.String())
			if err != nil {
				log.Println("Sending Create Mem", err)
				continue
			}

			if book.MemURL != "" {
				err := AddBookToCache(ctx, book)
				if err != nil {
					log.Println("Couldn't add book", err, book)
				}

			}

		}
	}
}

func GetTimeFromCache(context ctx) (string, bool) {
	var found bool = false
	var t string
	key := []byte(fmt.Sprint(lastSyncKey))
	// Check if the update exists
	err := context.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(lastUpdateBucket)
		if bucket != nil {
			buf := bucket.Get(key)
			if buf != nil {
				found = true
				t = string(buf)
			}
		}
		return nil
	})
	if err != nil {
		log.Panic("Read Book Transaction", err)
	}
	return t, found

}
func AddTimeToCache(context ctx, t string) error {
	key := []byte(lastSyncKey)
	value := []byte(t)
	// Check if the book exists
	err := context.db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(lastUpdateBucket)
		if err != nil {
			log.Panic("CreateBucketIfNotExists", err)
		}
		err = bucket.Put(key, value)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		log.Println("Couldn't add last update", err)
	}
	log.Println("Set Last Update to ", t)
	return err

}

func AddBookToCache(context ctx, book Book) error {
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

func GetBookFromCache(context ctx, id int) (Book, bool) {
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

func GetHighlights(ctx ctx) []Highlight {
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

func GetBooks(ctx ctx) []Book {
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

type memCreateApiInput struct {
	Content string `json:"content"`
}

type memCreateApiResponse struct {
	Url string `json:"url"`
}

func sendCreate(ctx ctx, text string) (string, error) {
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
	return memResponse.Url, err
	// Display Results
}
