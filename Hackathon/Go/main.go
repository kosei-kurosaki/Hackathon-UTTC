package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"github.com/oklog/ulid/v2"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	_ "strconv"
	"syscall"
	"time"
)

var db *sql.DB

func init() {
	err := godotenv.Load(".env_mysql")

	//1-1
	mysqlUser := os.Getenv("MYSQL_USER")
	mysqlUserPwd := os.Getenv("MYSQL_PASSWORD")
	// mysqlHost := os.Getenv("MYSQL_HOST")
	mysqlDatabase := os.Getenv("MYSQL_DATABASE")

	//1-2
	_db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@(localhost:3306)/%s", mysqlUser, mysqlUserPwd, mysqlDatabase))
	if err != nil {
		log.Fatalf("fail: sql.Open, %v\n", err)
	}
	//1-3
	if err := _db.Ping(); err != nil {
		log.Fatalf("fail: _db.Ping, %v\n", err)
	}
	db = _db

}

type UserResForHTTPGet struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age"`
}

// 2
//func getDBConnectionString() string {
//	mysqlUser := os.Getenv("MYSQL_USER")
//	mysqlUserPwd := os.Getenv("MYSQL_USER_PWD")
//	mysqlDatabase := os.Getenv("MYSQL_DATABASE")
//
//	return fmt.Sprintf("%s:%s@tcp(localhost:8002)/%s", mysqlUser, mysqlUserPwd, mysqlDatabase)
//}

type Post struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

//// responseの定義
//type Ping struct {
//	result string
//}

func handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, X-Requested-With")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	switch r.Method {
	case http.MethodGet:
		//2-1
		//nameの取得
		//name := r.URL.Query().Get("name") // To be filled
		////nameが長すぎるか空ならエラー
		//if name == "" {
		//	log.Println("fail: name is empty")
		//	w.WriteHeader(http.StatusBadRequest)
		//	return
		//}
		//
		//if len(name) > 50 {
		//	w.WriteHeader(http.StatusBadRequest)
		//	return
		//}
		////ageの取得
		//age := r.URL.Query().Get("age") // To be filled
		////ageが範囲外ならエラー
		//intage, err := strconv.Atoi(age)
		//if age == "" {
		//	log.Println("fail: age is empty")
		//	w.WriteHeader(http.StatusBadRequest)
		//	return
		//}
		//if intage < 20 || intage > 80 {
		//	w.WriteHeader(http.StatusBadRequest)
		//	return
		//}
		//2-2
		rows, err := db.Query("SELECT id, name, age FROM user")

		if err != nil {
			log.Printf("fail: db.Query, %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		//2-3
		users := make([]UserResForHTTPGet, 0)
		for rows.Next() {
			var u UserResForHTTPGet
			if err := rows.Scan(&u.ID, &u.Name, &u.Age); err != nil {
				log.Printf("fail: rows.Scan, %v\n", err)

				if err := rows.Close(); err != nil { // 500を返して終了するが、その前にrowsのClose処理が必要
					log.Printf("fail: rows.Close(), %v\n", err)
				}
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			users = append(users, u)
		}

		//2-4
		bytes, err := json.Marshal(users)
		if err != nil {
			log.Printf("fail: json.Marshal, %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")

		_, err = w.Write(bytes)
		if err != nil {
			// エラーハンドリングの処理
			// エラーをログに記録したり、エラーメッセージを返したりする
		}

	case http.MethodPost:
		tx, err := db.Begin()
		if err != nil {
			fmt.Print("Begin failure")
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var posts Post

		if err := json.Unmarshal(body, &posts); err != nil {
			log.Fatal(err)
		}

		// nameの制約
		//if posts.Name == "" || len(posts.Name) > 50 {
		//	http.Error(w, err.Error(), http.StatusBadRequest)
		//	return
		//}
		//if posts.Age < 20 || posts.Age > 80 {
		//	http.Error(w, err.Error(), http.StatusBadRequest)
		//	return
		//}
		//fmt.Print(posts.Name)

		// データベースに接続
		//db, err := sql.Open("mysql", getDBConnectionString())
		//if err != nil {
		//	http.Error(w, "Failed to connect to database", http.StatusInternalServerError)
		//	return
		//
		// INSERT文

		var ID string
		ID = generateID()
		//fmt.Print("generatedid:%v\n", ID)

		_, err = tx.Exec("INSERT INTO user (id,name,age) VALUES (?, ?, ?)", ID, posts.Name, posts.Age)
		if err != nil {

			http.Error(w, "Failed to prepare INSERT statement", http.StatusInternalServerError)
			log.Printf("err:%v\n", err)
			return
		}

		if err := tx.Commit(); err != nil {
			fmt.Print("commit error")
		}

		//// INSERT文を実行
		//_, err = stmt.Exec(posts.Name, posts.Age)
		//if err != nil {
		//	http.Error(w, "Failed to execute INSERT statement", http.StatusInternalServerError)
		//	return
		//
		//
		//成功レスポンス
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		//fmt.Printf("User inserted successfully")
		result := fmt.Sprintf("id:%s", ID)
		_, err = w.Write([]byte(result))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Printf("fail in marshal")
			return
		}
		//var res = Ping{result}
		//pin, err := json.Marshal(res)
		////fmt.Printf("w.Write(pin)のまえ")
		//
		//if err != nil {
		//	http.Error(w, err.Error(), http.StatusInternalServerError)
		//	log.Printf("fail in marshal")
		//	return
		//}
		//_, err = w.Write(pin)
		//if err != nil {
		//	// エラーハンドリングの処理を行う
		//	// 例えば、ログ出力やエラーレスポンスの送信など
		//	http.Error(w, err.Error(), http.StatusInternalServerError)
		//	return
		//}

	default:
		log.Printf("fail: HTTP Method is %s\n", r.Method)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}
func main() {
	//2
	http.HandleFunc("/user", handler)
	//3
	closeDBWithSysCall()
	log.Println("Listening...")
	err := http.ListenAndServe(":8003", nil)
	if err != nil {
		log.Fatal(err)
		return
	}

}

// 3
func closeDBWithSysCall() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		s := <-sig
		log.Printf("received syscall, %v", s)
		if err := db.Close(); err != nil {
			log.Fatal(err)

		}
		log.Printf("success: db.Close()")
		os.Exit(0)
	}()
}

//idの生成

func generateID() string {
	t := time.Now()
	entropy := ulid.Monotonic(rand.New(rand.NewSource(t.UnixNano())), 0)
	id := ulid.MustNew(ulid.Timestamp(t), entropy)
	return id.String()
}
