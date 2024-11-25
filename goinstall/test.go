package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

var database *sql.DB

type Config struct {
	DBHost  string
	DBPort  string
	DBUser  string
	DBPass  string
	DBDBase string
}

type Page struct {
	Title   string
	Content string
	Date    string
}

const (
	Port = ":8080"
)

func LoadEnv() error {
	currentDir, err := os.Getwd()
	if err != nil {
		return err
	}

	parentDir := filepath.Dir(currentDir)

	err = godotenv.Load(filepath.Join(parentDir, ".env"))
	if err != nil {
		return err
	}

	return nil
}

func LoadConfig() *Config {
	return &Config{
		DBHost:  os.Getenv("DB_HOST"),
		DBPort:  os.Getenv("DB_PORT"),
		DBUser:  os.Getenv("DB_USER"),
		DBPass:  os.Getenv("DB_PASS"),
		DBDBase: os.Getenv("DBD_BASE"),
	}
}

func (c *Config) GetConnectionString() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s%s/%s?sslmode=disable",
		c.DBUser,
		c.DBPass,
		c.DBHost,
		c.DBPort,
		c.DBDBase,
	)
}

func (c *Config) Validate() error {
	if c.DBHost == "" {
		return fmt.Errorf("DB Host is required")
	}
	return nil
}

// func serveDynamic(w http.ResponseWriter, _ *http.Request) {
// 	response := "The time is now " + time.Now().String()
// 	fmt.Fprintln(w, response)
// }

// func serveStatic(w http.ResponseWriter, r *http.Request) {
// 	http.ServeFile(w, r, "static.html")
// }

func pageHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pageId := vars["id"]
	thisPage := Page{}
	fmt.Println(pageId)
	err := database.QueryRow(
		"SELECT page_title, page_content, page_date FROM pages WHERE id = $1", pageId).Scan(
		&thisPage.Title, &thisPage.Content, &thisPage.Date)

	if err != nil {
		http.Error(w, http.StatusText(404), http.StatusNotFound)
		log.Println("Couldn't get page: " + pageId)
	}
	fmt.Printf("Single page: %+v\n", thisPage)

	html := `<html><head><title>` + thisPage.Title +
		`</title></head><body><h1>` + thisPage.Title + `</h1><div>` + thisPage.Content + `</div></body></html>`

	fmt.Fprintln(w, html)

	// fileName := "files/" + pageId + ".html"
	// fileInfo, err := os.Stat(fileName)
	// fmt.Println(fileInfo)
	// fmt.Println(err)
	// if err != nil {
	// 	fileName = "files/404.html"
	// }
	// http.ServeFile(w, r, fileName)
}

func main() {
	// dbConn := fmt.Sprintf("%s:%s@tcp(%s)/%s", DBUser, DBPass, DBHost, DBDBase)
	// fmt.Println(dbConn)
	// connStr := fmt.Sprintf("postgres://%s:%s@%s%s/%s?sslmode=disable", DBUser, DBPass, DBHost, DBPort, DBDBase)
	err := LoadEnv()
	if err != nil {
		log.Fatal("Error loading .env file:", err)
	}

	config := LoadConfig()
	connStr := config.GetConnectionString()
	fmt.Println(connStr)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Println("Couldn't connect!")
		log.Println(err.Error)
	}

	defer db.Close()

	// Test the connection
	err = db.Ping()
	if err != nil {
		log.Fatal("Error pinging the database: ", err)
	}

	fmt.Println("Successfully connected to database!")

	database = db

	router := mux.NewRouter()
	http.Handle("/", router)
	router.HandleFunc("/pages/{id:[0-9]+}", pageHandler)
	router.HandleFunc("/homepage", pageHandler)
	router.HandleFunc("/contact", pageHandler)
	// router.HandleFunc("/static", serveStatic)
	// router.HandleFunc("/", serveDynamic)
	fmt.Println("Everything is set up!")
	http.ListenAndServe(Port, nil)
}
