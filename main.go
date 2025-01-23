package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

type Book struct {
	ID              int    `json:"id"`
	Title           string `json:"title"`
	Description     string `json:"description"`
	Author          string `json:"author"`
	PublicationYear int    `json:"publication_year"`
}

var (
	books = make(map[int]Book)
	mtx   = &sync.Mutex{}
)

func main() {
	http.HandleFunc("/books", getAllBooks)

	http.HandleFunc("/books/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			getBookById(w, r)
		case "POST":
			createBook(w, r)
		case "PUT":
			updateBook(w, r)
		case "DELETE":
			deleteBook(w, r)
		default:
			http.Error(w, "Unsupported method", http.StatusMethodNotAllowed)
		}
	})

	fmt.Println("Server is running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// CRUDs operations
func getAllBooks(w http.ResponseWriter, r *http.Request) {
	mtx.Lock()
	defer mtx.Unlock()

	// Send response
	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(http.StatusOK)

	err := json.NewEncoder(w).Encode(books) // Convert from book as a map to json
	if err != nil {
		http.Error(w, "Failed to encode books data", http.StatusInternalServerError)
		return
	}
}

func getBookById(w http.ResponseWriter, r *http.Request) {
	mtx.Lock()
	defer mtx.Unlock()

	idStr := strings.TrimPrefix(r.URL.Path, "/books/")

	id, err := strconv.Atoi(idStr) // Convet from string to int
	if err != nil {
		http.Error(w, "Invalid book ID", http.StatusBadRequest)
		return
	}

	book, isExist := books[id]
	if !isExist {
		http.Error(w, "Book not found", http.StatusNotFound)
		return
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(book)
	if err != nil {
		http.Error(w, "Failed to encode book data", http.StatusInternalServerError)
		return
	}
}

func createBook(w http.ResponseWriter, r *http.Request) {
	mtx.Lock()
	defer mtx.Unlock()

	var newBook Book
	if err := json.NewDecoder(r.Body).Decode(&newBook); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Check if didn't fill the id
	if newBook.ID == 0 {
		http.Error(w, "Book id is required", http.StatusBadRequest)
		return
	}

	// Check if this id already exist
	if _, exists := books[newBook.ID]; exists {
		http.Error(w, "Book already exists", http.StatusConflict)
		return
	}

	books[newBook.ID] = newBook

	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(newBook); err != nil {
		http.Error(w, "Failed to encode book data", http.StatusInternalServerError)
		return
	}
}

func updateBook(w http.ResponseWriter, r *http.Request) {
	mtx.Lock()
	defer mtx.Unlock()

	idStr := strings.TrimPrefix(r.URL.Path, "/books/")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid book ID", http.StatusBadRequest)
		return
	}

	// Ensure that the book exists
	if _, exists := books[id]; !exists {
		http.Error(w, "Book not found", http.StatusNotFound)
		return
	}

	// Update the specified book
	var updatedBook Book
	if err := json.NewDecoder(r.Body).Decode(&updatedBook); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	if updatedBook.ID != id {
		http.Error(w, "Book id in the URL must match the id in the body", http.StatusBadRequest)
		return
	}

	books[id] = updatedBook

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(updatedBook); err != nil {
		http.Error(w, "Failed to encode updated book data", http.StatusInternalServerError)
		return
	}
}

func deleteBook(w http.ResponseWriter, r *http.Request) {
	mtx.Lock()
	defer mtx.Unlock()

	idStr := strings.TrimPrefix(r.URL.Path, "/books/")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid book ID", http.StatusBadRequest)
		return
	}

	// Ensure that the book exists
	if _, exists := books[id]; !exists {
		http.Error(w, "Book not found", http.StatusNotFound)
		return
	}

	delete(books, id)

	w.WriteHeader(http.StatusNoContent)
}
