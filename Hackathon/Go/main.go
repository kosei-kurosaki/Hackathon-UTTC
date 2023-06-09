package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/oklog/ulid/v2"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	_ "strconv"
	"time"
)

var db *sql.DB

func init() {

	//1-1
	mysqlUser := os.Getenv("MYSQL_USER")
	mysqlUserPwd := os.Getenv("MYSQL_PWD")
	mysqlHost := os.Getenv("MYSQL_HOST")
	mysqlDatabase := os.Getenv("MYSQL_DATABASE")
	//1-2

	_db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@%s/%s?parseTime=true", mysqlUser, mysqlUserPwd, mysqlHost, mysqlDatabase))
	if err != nil {
		log.Fatalf("fail: sql.Open, %v\n", err)
	}
	//1-3
	if err := _db.Ping(); err != nil {
		log.Fatalf("fail: _db.Ping, %v\n", err)

	}
	db = _db
}

type Message struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
	UserID    string `json:"userid"`
}

type NewMessage struct {
	Name      string `json:"name"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
}

var newID int = 1

var messages []Message

func getMessages(w http.ResponseWriter, r *http.Request) {

	rows, err := db.Query("SELECT * FROM messages")

	if err != nil {
		log.Printf("fail: db.Query, %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	//2-3
	messages := make([]Message, 0)
	for rows.Next() {
		var m Message
		if err := rows.Scan(&m.ID, &m.Name, &m.Message, &m.Timestamp, &m.UserID); err != nil {
			log.Printf("fail: rows.Scan, %v\n", err)

			if err := rows.Close(); err != nil { // 500を返して終了するが、その前にrowsのClose処理が必要
				log.Printf("fail: rows.Close(), %v\n", err)
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		messages = append(messages, m)
	}

	//2-4
	bytes, err := json.Marshal(messages)
	if err != nil {
		log.Printf("fail: json.Marshal, %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = w.Write(bytes)
	if err != nil {
		log.Printf("fail: w.Write, %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	return

}

func handlePostMessage(w http.ResponseWriter, r *http.Request) {
	var msg Message
	// リクエストボディからメッセージをデコード
	err := json.NewDecoder(r.Body).Decode(&msg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// メッセージをデータベースに保存するコードを後で書

	// レスポンスとしてメッセージをエンコードしてクライアントに送信
	json.NewEncoder(w).Encode(msg)
}

func updateMessage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	for index, msg := range messages {
		if msg.ID == id {
			var updatedMessage Message
			_ = json.NewDecoder(r.Body).Decode(&updatedMessage)
			messages[index].Message = updatedMessage.Message
			json.NewEncoder(w).Encode(messages[index])
			return
		}
	}
	http.Error(w, "Message not found", http.StatusNotFound)
}

func postMessage(w http.ResponseWriter, r *http.Request) {
	var newMessage NewMessage
	reqBody, _ := ioutil.ReadAll(r.Body)
	json.Unmarshal(reqBody, &newMessage)
	//aa
	// You might want to use real timestamp and user name here
	newMessage.Timestamp = time.Now().Format(time.RFC3339)
	newMessage.Name = "Some User"

	newID++
	messages = append(messages, Message{ID: strconv.Itoa(newID), Name: newMessage.Name, Message: newMessage.Message, Timestamp: newMessage.Timestamp})

	json.NewEncoder(w).Encode(newMessage)
}

func generateID() string {
	t := time.Now()
	entropy := ulid.Monotonic(rand.New(rand.NewSource(t.UnixNano())), 0)
	id := ulid.MustNew(ulid.Timestamp(t), entropy)
	return id.String()
}

func handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:59171")
	w.Header().Set("Access-Control-Allow-Origin", "https://hackathon-uttc-ngo5uabimq-uc.a.run.app")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	switch r.Method {
	case "GET":
		getMessages(w, r)
	case "POST":
		postMessage(w, r)
	case "PUT":
		updateMessage(w, r)
	default:
		http.Error(w, "Method not supported", http.StatusMethodNotAllowed)
	}
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("hello!!!!!!"))
		return
	}
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/messages", handler)
	router.HandleFunc("/hello", helloHandler)
	//router.HandleFunc("/messages", handlePostMessage)
	//router.HandleFunc("/messages", getMessages).Methods("GET")
	//router.HandleFunc("/messages/{id}", updateMessage).Methods("PUT")
	//router.HandleFunc("/messages", postMessage).Methods("POST")

}
