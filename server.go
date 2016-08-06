package main

// References:
// http://bit.ly/2aEN67x
// http://bit.ly/2aENrXK

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

// Password defines the scheme of the passwords.json file
type Password struct {
	Regular string `json:"regular"`
	Bonus   string `json:"bonus"`
}

var pw *Password

func init() {
	fd, err := ioutil.ReadFile("passwords.json")
	if err != nil {
		log.Fatalf("Unable to read password file: %s", err.Error())
	}

	err = json.Unmarshal(fd, &pw)
	if err != nil {
		log.Fatalf("Unable to decode password JSON file: %s", err.Error())
	}
}

func executeTemplate(templateFile string, w http.ResponseWriter, data interface{}) error {
	t, err := template.ParseFiles(templateFile)
	if err != nil {
		http.Error(w, "Error parsing template file", http.StatusInternalServerError)
		return err
	}

	t.Execute(w, data)
	return nil
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		if err := executeTemplate("templates/challenge.html", w, nil); err != nil {
			return
		}
	case "POST":
		r.ParseForm()
		pwEntered := r.Form.Get("passphrase")

		switch pwEntered {
		case pw.Regular:
			if err := executeTemplate("templates/success.html", w, pwEntered); err != nil {
				return
			}
		case pw.Bonus:
			if err := executeTemplate("templates/bonus.html", w, nil); err != nil {
				return
			}
		default:
			log.Printf("Password entered (%s) is incorrect", pwEntered)
			http.Redirect(w, r, "/", http.StatusMovedPermanently)
		}
	}
}

func main() {
	r := mux.NewRouter()
	port := 8000

	r.HandleFunc("/", indexHandler)
	r.HandleFunc("/static/img/giphy.gif", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, r.URL.Path[1:])
	})

	loggedRouter := handlers.LoggingHandler(os.Stdout, r)
	log.Printf("Starting server on :%d", port)

	srv := &http.Server{
		Handler:      loggedRouter,
		Addr:         fmt.Sprintf(":%d", port),
		WriteTimeout: 30 * time.Second,
		ReadTimeout:  30 * time.Second,
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
