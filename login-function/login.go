// refference is : https://qiita.com/__init__/items/21b2604db54867f8c543#-signup-%E3%81%AE%E5%AE%9F%E8%A3%852
package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"net/http/cgi"

	"golang.org/x/crypto/bcrypt"
	// データの暗号化
	"github.com/dgrijalva/jwt-go"

	"github.com/gorilla/mux"
	"github.com/lib/pq"
)

var db *sql.DB

func main() {

	i := Info{}

	// Convert
	// https://github.com/lib/pq/blob/master/url.go
	// "postgres://bob:secret@1.2.3.4:5432/mydb?sslmode=verify-full"
	// ->　"user=bob password=secret host=1.2.3.4 port=5432 dbname=mydb sslmode=verify-full"
	pgUrl, err := pq.ParseURL(i.GetDBUrl())

	// 戻り値に err を返してくるので、チェック
	if err != nil {
		// エラーの場合、処理を停止する
		log.Fatal()
	}

	// DB 接続
	db, err = sql.Open("postgres", pgUrl)
	if err != nil {
		log.Fatal(err)
	}

	// DB 疎通確認
	err = db.Ping()

	if err != nil {
		log.Fatal(err)
	}

	// urls.py
	router := mux.NewRouter()

	// endpoint
	router.HandleFunc("/aps/login.cgi/signup", signup).Methods("POST")
	router.HandleFunc("/aps/login.cgi/login", login).Methods("POST")
	// service はあとで記述する

	// log.Fatal は、異常を検知すると処理の実行を止めてくれる
	cgi.Serve(router)
}

// User is DB data
type User struct {
	// 大文字だと Public 扱い
	ID       int    `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Info us dburl
type Info struct {
	dburl string
}

func (u Info) GetDBUrl() string {
	return "postgres://jhxftezm:j97prNfNWSDTgA3WguME6LXwKj_Hpo1S@rajje.db.elephantsql.com:5432/jhxftezm"
}

// JSON 形式で結果を返却
// data interface{} とすると、どのような変数の型でも引数として受け取ることができる
func responseByJSON(w http.ResponseWriter, data interface{}) {
	json.NewEncoder(w).Encode(data)
	return
}

// Token 作成関数
func createToken(user User) (string, error) {
	var err error

	// 鍵となる文字列(多分なんでもいい)
	secret := "secret"

	// Token を作成
	// jwt -> JSON Web Token - JSON をセキュアにやり取りするための仕様
	// jwtの構造 -> {Base64 encoded Header}.{Base64 encoded Payload}.{Signature}
	// HS254 -> 証明生成用(https://ja.wikipedia.org/wiki/JSON_Web_Token)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email": user.Email,
		"iss":   "__init__", // JWT の発行者が入る(文字列(__init__)は任意)
	})

	tokenString, err := token.SignedString([]byte(secret))

	if err != nil {
		log.Fatal(err)
	}

	return tokenString, nil
}

type JWT struct {
	Token string `json:"token"`
}

type Error struct {
	Message string `json:"message"`
}

func errorInResponse(w http.ResponseWriter, status int, error Error) {
	w.WriteHeader(status) // 400 とか 500 などの HTTP status コードが入る
	json.NewEncoder(w).Encode(error)
	return
}

func signup(w http.ResponseWriter, r *http.Request) {
	var user User
	var error Error

	// https://golang.org/pkg/encoding/json/#NewDecoder
	json.NewDecoder(r.Body).Decode(&user)

	if user.Email == "" {
		error.Message = "Email は必須です。"
		errorInResponse(w, http.StatusBadRequest, error)
		return
	}

	if user.Password == "" {
		error.Message = "パスワードは必須です。"
		errorInResponse(w, http.StatusBadRequest, error)
		return
	}

	// パスワードのハッシュを生成
	// https://godoc.org/golang.org/x/crypto/bcrypt#GenerateFromPassword
	hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), 10)

	if err != nil {
		log.Fatal(err)
	}

	user.Password = string(hash)

	sql_query := "INSERT INTO USERS(EMAIL, PASSWORD) VALUES($1, $2) RETURNING ID;"

	// query 発行
	// Scan で、Query 結果を変数に格納
	err = db.QueryRow(sql_query, user.Email, user.Password).Scan(&user.ID)

	if err != nil {
		error.Message = "サーバーエラー"
		errorInResponse(w, http.StatusInternalServerError, error)
		return
	}

	// DB に登録できたらパスワードをからにしておく
	user.Password = ""
	w.Header().Set("Content-Type", "application/json")

	// JSON 形式で結果を返却
	responseByJSON(w, user)
}

func login(w http.ResponseWriter, r *http.Request) {
	var user User
	var error Error
	var jwt JWT

	json.NewDecoder(r.Body).Decode(&user)

	if user.Email == "" {
		error.Message = "Email は必須です。"
		errorInResponse(w, http.StatusBadRequest, error)
		return
	}

	if user.Password == "" {
		error.Message = "パスワードは、必須です。"
		errorInResponse(w, http.StatusBadRequest, error)
	}

	password := user.Password

	// 認証キー(Emal)のユーザー情報をDBから取得
	row := db.QueryRow("SELECT * FROM USERS WHERE email=$1;", user.Email)
	// ハッシュ化している
	err := row.Scan(&user.ID, &user.Email, &user.Password)

	if err != nil {
		if err == sql.ErrNoRows { // https://golang.org/pkg/database/sql/#pkg-variables
			error.Message = "ユーザが存在しません。"
			errorInResponse(w, http.StatusBadRequest, error)
		} else {
			log.Fatal(err)
		}
	}

	hasedPassword := user.Password

	err = bcrypt.CompareHashAndPassword([]byte(hasedPassword), []byte(password))

	if err != nil {
		error.Message = "無効なパスワードです。"
		errorInResponse(w, http.StatusUnauthorized, error)
		return
	}

	token, err := createToken(user)

	if err != nil {
		log.Fatal(err)
	}

	w.WriteHeader(http.StatusOK)
	jwt.Token = token

	responseByJSON(w, jwt)
}
