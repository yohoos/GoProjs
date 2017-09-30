package main

import (
	"net/http"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
	"log"
)

const (
	user     = "yohoos"
	password = "magicdust50"
	dbname   = "testdb"
)

var db *sql.DB
var err error

var createTableQuery = "CREATE TABLE IF NOT EXISTS users(" +
	"id SERIAL PRIMARY KEY," +
	"username VARCHAR(50)," +
	"password VARCHAR(120)" +
	");"

func homePage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
}

func main() {
	connectionString := fmt.Sprintf("user=%s password=%s dbname=%s", user, password, dbname)

	db, err = sql.Open("postgres", connectionString)
	if err != nil {
		panic(err.Error())
	}

	defer db.Close()

	// make sure the table is created

	err = db.Ping()
	if err != nil {
		fmt.Println("Database connection not working!")
	} else {
		fmt.Println("Database connection working!")
	}

	if _, err := db.Exec(createTableQuery); err != nil {
		log.Fatal(err)
	}
	db.Exec("DELETE FROM users")
	db.Exec("ALTER SEQUENCE users_id_seq RESTART WITH 1")

	http.HandleFunc("/signup", signUpPage)
	http.HandleFunc("/login", login)
	http.HandleFunc("/", homePage)
	http.ListenAndServe(":8080", nil)

}

func login(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.ServeFile(w, r, "login.html")
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	var databaseUsername string
	var databasePassword string

	err := db.QueryRow("SELECT username, password FROM users WHERE username=$1", username).
		Scan(&databaseUsername, &databasePassword)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusMovedPermanently)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(databasePassword), []byte(password))
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusMovedPermanently)
		return
	}

	w.Write([]byte("Hello " + databaseUsername))
}

func signUpPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.ServeFile(w, r, "signup.html")
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	var user string

	err := db.QueryRow("SELECT username FROM users WHERE username=$1", username).Scan(&user)

	switch {
	case err == sql.ErrNoRows:
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "Invalid password!", http.StatusInternalServerError)
			return
		}

		_, err = db.Exec("INSERT INTO users(username, password) VALUES($1, $2)", username, hashedPassword)
		if err != nil {
			http.Error(w, "Unable to create new account!", http.StatusInternalServerError)
			return
		}

		w.Write([]byte("user created!"))
		return
	case err != nil:
		http.Error(w, "Username already exists.", http.StatusInternalServerError)
		return
	default:
		http.Redirect(w, r, "/", http.StatusMovedPermanently)
	}
}
