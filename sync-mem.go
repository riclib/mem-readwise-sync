package main

import (
	"bytes"
	"log"
	"time"
)

func syncMem(context Context) {
	syncBooksToMem(context)
	syncHighlightsToMem(context)
}

func syncHighlightsToMem(ctx Context) {
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
		_, err = sendAppend(ctx, b.Book.MemId, buf.String())
		if err != nil {
			log.Println("Sending Create Mem", err)
		} else {
			log.Println("Synced ", len(b.Highlight), " for ", b.Book.Title)
		}

	}

}

func syncBooksToMem(ctx Context) {
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

			createResponse, err := sendCreate(ctx, buf.String())
			if err != nil {
				log.Println("Sending Create Mem", err)
				continue
			}
			book.MemURL = createResponse.Url
			book.MemId = createResponse.Id

			if book.MemURL != "" {
				err := AddBookToCache(ctx, book)
				if err != nil {
					log.Println("Couldn't add book", err, book)
				}

			}

		}
	}
}
