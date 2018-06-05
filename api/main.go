package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"

	"gopkg.in/mgo.v2"
)

type Post struct {
	Text      string    `json:"text" bson:"text"`
	CreatedAt time.Time `json:"createdAt" bson:"created_at"`
}

type Task struct {
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"createdAt"`
}

var posts *mgo.Collection
var tasks *mgo.Collection

func main() {
	session, err := mgo.Dial("mongo:27017")
	if err != nil {
		log.Fatalln(err)
		log.Fatalln("mongo err")
		os.Exit(1)
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)

	// Get posts collection
	posts = session.DB("app").C("posts")
	tasks = session.DB("app").C("tasks")

	// Set up routes
	r := mux.NewRouter()

	r.HandleFunc("/posts", createPost).Methods("POST")
	r.HandleFunc("/posts", readPosts).Methods("GET")

	r.HandleFunc("/tasks", readTasks).Methods("GET")
	r.HandleFunc("/tasks", createTasks).Methods("POST")

	r.HandleFunc("/echo", echo).Methods("GET")

	err = http.ListenAndServe(":8080", cors.AllowAll().Handler(r))
	if err != nil {
		fmt.Println(err)
		log.Fatalln(err)
		log.Fatalln("server start err")
		os.Exit(1)
	}

	log.Println("Listening on port 8080...")

}

func echo(w http.ResponseWriter, r *http.Request) {
	responseJSON(w, "okay")
}

func createTasks(w http.ResponseWriter, r *http.Request) {
	// Read body
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		responseError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Read post
	task := &Task{}
	err = json.Unmarshal(data, task)
	if err != nil {
		responseError(w, err.Error(), http.StatusBadRequest)
		return
	}
	task.CreatedAt = time.Now().UTC()

	// Insert new post
	if err := tasks.Insert(task); err != nil {
		responseError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	responseJSON(w, task)
}

func readTasks(w http.ResponseWriter, r *http.Request) {
	result := []Task{}
	if err := tasks.Find(nil).Sort("-created_at").All(&result); err != nil {
		responseError(w, err.Error(), http.StatusInternalServerError)
	} else {
		responseJSON(w, result)
	}
}

func readPosts(w http.ResponseWriter, r *http.Request) {
	result := []Post{}
	if err := posts.Find(nil).Sort("-created_at").All(&result); err != nil {
		responseError(w, err.Error(), http.StatusInternalServerError)
	} else {
		responseJSON(w, result)
	}
}

func createPost(w http.ResponseWriter, r *http.Request) {
	// Read body
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		responseError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Read post
	post := &Post{}
	err = json.Unmarshal(data, post)
	if err != nil {
		responseError(w, err.Error(), http.StatusBadRequest)
		return
	}
	post.CreatedAt = time.Now().UTC()

	// Insert new post
	if err := posts.Insert(post); err != nil {
		responseError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	responseJSON(w, post)
}

func responseError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func responseJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
