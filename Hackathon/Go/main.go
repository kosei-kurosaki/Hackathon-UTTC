package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/oklog/ulid/v2"
	"log"
	"math/rand"
	"net/http"
	"os"
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

	dsn := fmt.Sprintf("%s:%s@%s/%s?parseTime=true", mysqlUser, mysqlUserPwd, mysqlHost, mysqlDatabase)
	_db, err := sql.Open("mysql", dsn)
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
	ChannelId string `json:"channelid"`
}

type MessageEdit struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}

type NewMessage struct {
	Name      string `json:"name"`
	Message   string `json:"message"`
	UserId    string `json:"userid"`
	ChannelId string `json:"channelid"`
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
		if err := rows.Scan(&m.ID, &m.Name, &m.Message, &m.Timestamp, &m.UserID, &m.ChannelId); err != nil {
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

func updateMessage(w http.ResponseWriter, r *http.Request) {

	// リクエストボディからデータを読み込みます
	var updatedMessage MessageEdit
	err := json.NewDecoder(r.Body).Decode(&updatedMessage)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// SQLクエリを実行します
	stmt, err := db.Prepare("UPDATE messages SET Message = updatedMessage.message WHERE ID = updatedMessage.id")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(updatedMessage.Message, updatedMessage.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 成功した場合は200ステータスコードを返します
	w.WriteHeader(http.StatusOK)
}

func postMessage(w http.ResponseWriter, r *http.Request) {
	// リクiiエストからMessageを取得
	var msg NewMessage
	err := json.NewDecoder(r.Body).Decode(&msg)
	fmt.Printf("リクエストをmsgに入れ流ことはできた")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// Timestampの設定
	var Timestamp = time.Now().Format(time.RFC3339)
	//IDの設定
	var ID = generateID()
	fmt.Printf(ID, Timestamp)
	// データベースへの書き込み
	_, err = db.Exec(`INSERT INTO messages (id, name, message, timestamp, userid, channelid) VALUES (?, ?, ?, ?, ?, ?)`,
		ID, msg.Name, msg.Message, Timestamp, msg.UserId, msg.ChannelId)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// レスポンスの作成と送信
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(msg)
}

func deleteMessage(w http.ResponseWriter, r *http.Request) {
	deleteid := r.URL.Query().Get("messageid")
	delForm, err := db.Prepare("DELETE FROM messages WHERE id=deleteid")
	if err != nil {
		panic(err.Error())
	}

	_, err = delForm.Exec(deleteid)
	fmt.Print(deleteid)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Failed to delete message with ID = %s", deleteid)
		return
	}

	fmt.Fprintf(w, "Message with ID = %s was deleted", deleteid)
	// 成功した場合は200ステータスコードを返します
	w.WriteHeader(http.StatusOK)
}

func generateID() string {
	t := time.Now()
	entropy := ulid.Monotonic(rand.New(rand.NewSource(t.UnixNano())), 0)
	id := ulid.MustNew(ulid.Timestamp(t), entropy)
	return id.String()
}

func handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Credential", "true")
	w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE")
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
	case "DELETE":
		deleteMessage(w, r)

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
	http.HandleFunc("/messages", handler)
	http.HandleFunc("/hello", helloHandler)
	//router.HandleFunc("/messages", handlePostMessage)
	//router.HandleFunc("/messages", getMessages).Methods("GET")
	//router.HandleFunc("/messages/{id}", updateMessage).Methods("PUT")
	//router.HandleFunc("/messages", postMessage).Methods("POST")

	log.Println("Listening...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
