package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

type User struct {
	Customer_id    int    `json:"Customer_id"`
	Age            int    `json:"Age"`
	Gender         string `json:"Gender"`
	Annual_Income  int    `json:"Annual_Income"`
	Spending_score int    `json:"Spending_score"`
}

// Basic Authentication credentials
var validUsername = "admin"
var validPassword = "password123"

func main() {
	var err error

	// Menghubungkan ke database MySQL
	db, err = sql.Open("mysql", "root:Mimey$123@tcp(localhost:3306)/datamall")
	if err != nil {
		log.Fatal("Failed to connect to MySQL:", err)
	}
	defer db.Close()

	// Endpoint untuk memeriksa status
	http.HandleFunc("/status", statusHandler)
	// Endpoint untuk handle users
	http.HandleFunc("/users", basicAuth(handleUsers))
	// Endpoint untuk handle users by id
	http.HandleFunc("/users/", basicAuth(handleUserByID))

	// Menjalankan server pada port 8080
	fmt.Println("Server running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

// Middleware for Basic Authentication
func basicAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if !ok || !validateCredentials(username, password) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized) // 401
			return
		}
		next(w, r)
	}
}

// Function to validate credentials
func validateCredentials(username, password string) bool {
	return username == validUsername && password == validPassword
}

// Handler untuk endpoint /status
func statusHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "API is up and running"})
}

func handleUsers(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		createUser(w, r)
	case http.MethodGet:
		getUsers(w)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed) // 405
	}
}

func handleUserByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Path[len("/users/"):]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest) // 400
		return
	}

	switch r.Method {
	case http.MethodPut:
		UpdateUser(w, r, id)
	case http.MethodDelete:
		DeleteUser(w, r, id)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed) // 405
	}
}

func createUser(w http.ResponseWriter, r *http.Request) {
	var user User

	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest) // 400
		return
	}

	stmt, err := db.Prepare("INSERT INTO mall (Age, Gender, Annual_Income, Spending_score) VALUES (?, ?, ?, ?)")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError) // 500
		return
	}
	defer stmt.Close()

	res, err := stmt.Exec(user.Age, user.Gender, user.Annual_Income, user.Spending_score)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError) // 500
		return
	}

	lastID, err := res.LastInsertId()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError) // 500
		return
	}

	user.Customer_id = int(lastID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func getUsers(w http.ResponseWriter) {
	rows, err := db.Query("SELECT Customer_id, Age, Gender, Annual_Income, Spending_score FROM mall")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError) // 500
		return
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.Customer_id, &user.Age, &user.Gender, &user.Annual_Income, &user.Spending_score); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError) // 500
			return
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError) // 500
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

// Fungsi untuk memperbarui data pengguna
func UpdateUser(w http.ResponseWriter, r *http.Request, id int) {
	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest) // 400
		return
	}

	// Cek apakah ID ada dalam database
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM mall WHERE Customer_id=?)", id).Scan(&exists)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError) // 500
		return
	}

	if !exists {
		http.Error(w, "User not found", http.StatusNotFound) // 404
		return
	}

	stmt, err := db.Prepare("UPDATE mall SET Age=?, Gender=?, Annual_Income=?, Spending_score=? WHERE Customer_id=?")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError) // 500
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(user.Age, user.Gender, user.Annual_Income, user.Spending_score, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError) // 500
		return
	}

	user.Customer_id = id
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// Fungsi untuk menghapus data pengguna
func DeleteUser(w http.ResponseWriter, r *http.Request, id int) {
	stmt, err := db.Prepare("DELETE FROM mall WHERE Customer_id=?")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError) // 500
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError) // 500
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
