package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
)

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

var db *sql.DB

func main() {

	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname,
	)

	var err error

	for i := 0; i < 10; i++ {
		db, err = sql.Open("postgres", dsn)
		if err == nil {
			err = db.Ping()
			if err == nil {
				break
			}
		}
		log.Println("Waiting for database...")
		time.Sleep(3 * time.Second)
	}

	if err != nil {
		log.Fatal("Cannot connect to DB:", err)
	}

	fmt.Println("Connected to Database")
	fmt.Println("Starting the Server on :8080")

	http.HandleFunc("/users", usersHandler)

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func usersHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {

	case "GET":
		rows, err := db.Query("SELECT id, name, email FROM users")
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		defer rows.Close()

		var users []User

		for rows.Next() {
			var u User
			rows.Scan(&u.ID, &u.Name, &u.Email)
			users = append(users, u)
		}

		json.NewEncoder(w).Encode(users)

	case "POST":
		var u User
		json.NewDecoder(r.Body).Decode(&u)

		err := db.QueryRow(
			"INSERT INTO users(name,email) VALUES($1,$2) RETURNING id",
			u.Name, u.Email,
		).Scan(&u.ID)

		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		json.NewEncoder(w).Encode(u)

	case "DELETE":
		id := r.URL.Query().Get("id")

		_, err := db.Exec("DELETE FROM users WHERE id=$1", id)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		w.Write([]byte("Deleted"))

	}
}
