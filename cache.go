package main

import (
	"encoding/json"
	"fmt"
	"github.com/boltdb/bolt"
	"log"
)

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

func AddTimeToCache(context Context, t string) error {
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

func GetTimeFromCache(context Context) (string, bool) {
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
