package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func main() {
	// connect to database
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// create table if it doesn't exist
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS users (id SERIAL PRIMARY KEY, name TEXT, email TEXT)")

	if err != nil {
		log.Fatal(err)
	}

	// router
	router := mux.NewRouter()
	router.HandleFunc("/users", getUsers(db)).Methods("GET")
	router.HandleFunc("/users/{id}", getUser(db)).Methods("GET")
	router.HandleFunc("/users", createUser(db)).Methods("POST")
	router.HandleFunc("/users/{id}", updateUser(db)).Methods("PUT")
	router.HandleFunc("/users/{id}", deleteUser(db)).Methods("DELETE")

	// start server
	log.Fatal(http.ListenAndServe(":8000", jsonContentTypeMiddleware(router))) //ensure header is set to json

	//send a log when server is running
}

func jsonContentTypeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json") //set header to json
		next.ServeHTTP(w, r)                               //call next handler, similar to next() in express
	})
}

func getUsers(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT * FROM users")
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close() //defer functions will be called when the function returns

		users := []User{} //slice to hold users

		// iterate through rows
		for rows.Next() {
			var u User
			err := rows.Scan(&u.ID, &u.Name, &u.Email)
			if err != nil {
				log.Fatal(err)
			}
			users = append(users, u)
		}
		// check for errors after iterating through rows
		err = rows.Err()
		if err != nil {
			log.Fatal(err)
		}

		// write response
		json.NewEncoder(w).Encode(users)
	}
}

func getUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r) //get url params
		id := vars["id"]    //get id from params

		var u User
		// $1 is a placeholder for the id variable
		err := db.QueryRow("SELECT * FROM users WHERE id=$1", id).Scan(&u.ID, &u.Name, &u.Email) //scan method populates user struct
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		// write response
		json.NewEncoder(w).Encode(u)
	}
}

func createUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var u User
		json.NewDecoder(r.Body).Decode(&u) //decode request body into user struct

		// insert user into database
		err := db.QueryRow("INSERT INTO users(name, email) VALUES($1, $2) RETURNING id", u.Name, u.Email).Scan(&u.ID)
		if err != nil {
			log.Fatal(err)
		}

		// write response
		json.NewEncoder(w).Encode(u)
	}
}

func updateUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r) //get url params
		id := vars["id"]    //get id from params

		var u User
		json.NewDecoder(r.Body).Decode(&u) //decode request body into user struct

		// update user in database
		err := db.QueryRow("UPDATE users SET name=$1, email=$2 WHERE id=$3 RETURNING id", u.Name, u.Email, id).Scan(&u.ID)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		// write response
		json.NewEncoder(w).Encode(u)
	}
}

func deleteUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r) //get url params
		id := vars["id"]    //get id from params

		// delete user from database
		var u User
		err := db.QueryRow("SELECT * FROM users WHERE id = $1", id).Scan(&u.ID, &u.Name, &u.Email)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		} else {
			_, err := db.Exec("DELETE FROM users WHERE id = $1", id)
			if err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusNotFound)
				return
			}

			json.NewEncoder(w).Encode("User deleted")
		}
	}
}
