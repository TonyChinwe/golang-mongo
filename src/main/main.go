package main

import (
	"context"
	"encoding/json"
	"golang-mongo/src/models"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var booksDB *mongo.Collection

var Client *mongo.Client
var ctx1 context.Context

func main() {
	initRepository()
	initRoutes()
	defer Client.Disconnect(ctx1)
}

func initRepository() {

	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://127.0.0.1:27017"))
	if err != nil {
		log.Fatal(err)
	}
	Client = client
	ctx := context.Background()

	ctx1 = ctx
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	database := client.Database("bookstore")
	booksDB = database.Collection("books")

}

func initRoutes() {
	r := mux.NewRouter().StrictSlash(true)
	r.HandleFunc("/create-book", CreateBook).Methods("POST")
	r.HandleFunc("/get-all-books", GetAllBooks).Methods("GET")
	r.HandleFunc("/get-book-by-title/{title}", GetBookByTitle).Methods("GET")
	r.HandleFunc("/get-book-by-isbn/{isbn}", GetBookByIsbn).Methods("GET")
	r.HandleFunc("/update-book", UpdateBook).Methods("PUT")
	r.HandleFunc("/delete-book-by-isbn/{isbn}", DeleteBookBasedOnIsbn).Methods("DELETE")
	http.ListenAndServe(":8080", r)
}

func CreateBook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var book models.Book
	json.NewDecoder(r.Body).Decode(&book)

	if book.Title == "" || len(book.Authors) == 0 || book.Isbn == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("The required fields are not complete"))
		return
	}
	var bookInDb models.Book
	err := booksDB.FindOne(ctx1, bson.M{"isbn": book.Isbn}).Decode(&bookInDb)
	if err == nil {
		w.WriteHeader(http.StatusConflict)
		w.Write([]byte("A book with similar isbn already exist"))
		return
	}
	res, insertErr := booksDB.InsertOne(ctx1, book)
	if insertErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(insertErr.Error()))
		return
	}
	json.NewEncoder(w).Encode(res)

}

func GetAllBooks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var books []models.Book
	bookcursor, err := booksDB.Find(ctx1, bson.M{})
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	err2 := bookcursor.All(ctx1, &books)
	if err2 != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	if len(books) == 0 {
		emptybooks := []models.Book{}
		json.NewEncoder(w).Encode(emptybooks)
		return
	}
	json.NewEncoder(w).Encode(books)

}

func GetBookByTitle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var path = mux.Vars(r)["title"]
	var book models.Book
	err := booksDB.FindOne(ctx1, bson.M{"title": path}).Decode(&book)

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(err.Error()))
		return
	} else {
		json.NewEncoder(w).Encode(book)
	}

}

func GetBookByIsbn(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var path = mux.Vars(r)["isbn"]
	var book models.Book
	err := booksDB.FindOne(ctx1, bson.M{"isbn": path}).Decode(&book)

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(err.Error()))
		return
	} else {
		json.NewEncoder(w).Encode(book)
	}

}

func UpdateBook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var book models.Book
	json.NewDecoder(r.Body).Decode(&book)
	if book.Title == "" || len(book.Authors) == 0 || book.Isbn == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("The required fields are not complete"))
		return
	}
	result, err := booksDB.ReplaceOne(ctx1, bson.M{"isbn": book.Isbn},
		bson.M{"Title": book.Title, "authors": book.Authors, "isbn": book.Isbn},
	)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(err.Error()))
		return
	} else {
		json.NewEncoder(w).Encode("Document " + strconv.Itoa(int(result.ModifiedCount)) + " updated")
	}
}

func DeleteBookBasedOnIsbn(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var path = mux.Vars(r)["isbn"]
	result, err := booksDB.DeleteOne(ctx1, bson.M{"isbn": path})
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(err.Error()))
		return
	}
	json.NewEncoder(w).Encode("Document " + strconv.Itoa(int(result.DeletedCount)) + " deleted")
}
