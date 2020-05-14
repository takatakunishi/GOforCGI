package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cgi"
	"os"

	"database/sql"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/bitly/go-simplejson"
	"github.com/lib/pq"
	"github.com/rs/xid"
	"golang.org/x/crypto/bcrypt"
)

var db *sql.DB

func main() {

	i := Info{}
	pgUrl, err := pq.ParseURL(i.GetDBUrl())
	if err != nil {
		log.Fatal()
	}
	db, err = sql.Open("postgres", pgUrl)
	if err != nil {
		log.Fatal(err)
	}
	// DB 疎通確認
	err = db.Ping()

	if err != nil {
		log.Fatal(err)
	}
	api := rest.NewApi()
	api.Use(rest.DefaultDevStack...)
	api.Use(&rest.CorsMiddleware{
		RejectNonCorsRequests: false,
		OriginValidator: func(origin string, request *rest.Request) bool {
			return true
		},
		AllowedMethods: []string{"GET", "POST", "PUT"},
		AllowedHeaders: []string{
			"Accept", "Content-Type", "X-Custom-Header", "Origin"},
		AccessControlAllowCredentials: true,
		AccessControlMaxAge:           3600,
	})
	router, err := rest.MakeRouter(
		rest.Get("/aps/routerCgi.cgi/getAllData", getAllData),
		rest.Get("/aps/routerCgi.cgi/GetAWork/:request", GetAWork),
		rest.Post("/aps/routerCgi.cgi/PostData", PostData),
		rest.Post("/aps/routerCgi.cgi/signup", signup),
		rest.Post("/aps/routerCgi.cgi/login", login),
	)
	if err != nil {
		log.Fatal(err)
	}
	api.SetApp(router)
	cgi.Serve(api.MakeHandler())
}

const filePath string = "works.json"

// getAllData 作品データをすべて送るAPI
func getAllData(w rest.ResponseWriter, r *rest.Request) {

	rawData, err := getSimpleJSON(filePath)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := rawData.Get("Id")

	works := make([]ID, 0)
	for _, v := range data.MustMap() {
		fake, _ := json.Marshal(v)
		if err != nil {
			fmt.Println(err)
			break
		}
		var box ID
		err = json.Unmarshal(fake, &box)
		if err != nil {
			fmt.Println(err)
			break
		}
		works = append(works, box)
	}

	var result Data

	result.Id = works

	w.WriteJson(&result)
}

// GetAWork リクエストされたデータを返すAPI
func GetAWork(w rest.ResponseWriter, r *rest.Request) {
	request := r.PathParam("request")

	rawData, err := getSimpleJSON(filePath)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, tof := rawData.Get("Id").CheckGet(request)
	if !tof {
		rest.Error(w, "no data", http.StatusInternalServerError)
	}

	works := make([]ID, 0)
	v := data.MustMap()
	fake, _ := json.Marshal(v)
	if err != nil {
		fmt.Println(err)
	}
	var box ID
	err = json.Unmarshal(fake, &box)
	if err != nil {
		fmt.Println(err)
	}
	works = append(works, box)

	var result Data

	result.Id = works

	w.WriteJson(&works)
}

// PostData 送られてきたデータを書き込むAPI
func PostData(w rest.ResponseWriter, r *rest.Request) {
	var CreatedWorkTag string = (xid.New()).String()
	DesignationUserID := "Id"
	var jsonTag = "Work"
	var user User
	var resultJSON ID

	var sendData map[string]interface{}
	err := r.DecodeJsonPayload(&sendData)
	if err != nil {
		log.Fatal(109)
		return
	}

	fake, err := json.Marshal(sendData["UserInfo"])
	if err != nil {
		rest.Error(w, "サーバーエラー1", http.StatusInternalServerError)
		return
	}
	err = json.Unmarshal(fake, &user)
	if err != nil {
		rest.Error(w, "サーバーエラー2", http.StatusInternalServerError)
		return
	}

	loginCheck(w, r, user)

	fake, err = json.Marshal(sendData[jsonTag])
	if err != nil {
		rest.Error(w, "サーバーエラーx", http.StatusInternalServerError)
		return
	}
	err = json.Unmarshal(fake, &resultJSON)
	if err != nil {
		rest.Error(w, "サーバーエラーy", http.StatusInternalServerError)
		return
	}
	if resultJSON.Title == "" {
		rest.Error(w, "failed to data", http.StatusInternalServerError)
		return
	}

	data, err := getSimpleJSON(filePath)

	_, tof := data.Get(DesignationUserID).CheckGet(CreatedWorkTag)

	if tof {
		var i = 0
		for {
			if i < 5 && tof {
				_, tof = data.CheckGet(CreatedWorkTag)
				if i == 4 {
					log.Fatal(123)
					rest.Error(w, "Failed to Tagging. Retry sends data!", http.StatusInternalServerError)
					break
				}
			}
			i++
		}
	}

	data.Get(DesignationUserID).SetPath([]string{CreatedWorkTag}, sendData[jsonTag])
	data.Get(DesignationUserID).Get(CreatedWorkTag).Set("WorkTag", CreatedWorkTag)

	o, _ := data.EncodePretty()
	err = writeFile(filePath, o)
	if err != nil {
		log.Fatal(136)
		return
	}

	fake, err = data.Get(DesignationUserID).Get(CreatedWorkTag).MarshalJSON()
	if err != nil {
		log.Fatal(141)
		return
	}
	err = json.Unmarshal(fake, &resultJSON)
	if err != nil {
		log.Fatal(147)
		return
	}
	err = w.WriteJson(&resultJSON)
	if err != nil {
		log.Fatal(153)
		return
	}
}

func writeJSON(Filename string, data Data) (err error) {
	//ファイルへの書き込み
	result, err := json.MarshalIndent(data, "", "	")
	if err != nil {
		return err
	}

	return writeFile(Filename, result)
}

func getSimpleJSON(filePath string) (j *simplejson.Json, err error) {

	bytes, err := readFile(filePath)
	var rawData *simplejson.Json
	if err != nil {
		fmt.Println(err)
		return rawData, err
	}

	rawData, err = simplejson.NewJson(bytes)

	return rawData, err
}

func writeFile(filename string, bytes []byte) (err error) {
	return ioutil.WriteFile(filename, bytes, os.ModePerm)
}

func readFile(filename string) ([]byte, error) {
	byte, err := ioutil.ReadFile(filename)
	return byte, err
}

// ID is each works data
type ID struct {
	WorkTag     string   `json:"WorkTag"`
	Title       string   `json:"Title"`
	Auth        string   `json:"Auth"`
	Corporator  []string `json:"Corporator"`
	Date        string   `json:"Date"`
	URL         []string `json:"Url"`
	Description string   `json:"Description"`
	Tags        []string `json:"Tags"`
	//Likes is
	Likes struct {
		Amount int `json:"Amount"`
		// Users is about other User of this user
		Users []string `json:"Users"`
	} `json:"Likes"`
}

// Data is each person data
type Data struct {
	Id []ID `json:"Id"`
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

type Res struct {
	Status string `json:"status"`
}

func (u Info) GetDBUrl() string {
	return "postgres://jhxftezm:j97prNfNWSDTgA3WguME6LXwKj_Hpo1S@rajje.db.elephantsql.com:5432/jhxftezm"
}

// JSON 形式で結果を返却
// data interface{} とすると、どのような変数の型でも引数として受け取ることができる
func responseByJSON(w rest.ResponseWriter, data interface{}) {
	w.WriteJson(&data)
	return
}

type Error struct {
	Message string `json:"message"`
}

func errorInResponse(w http.ResponseWriter, status int, error Error) {
	w.WriteHeader(status) // 400 とか 500 などの HTTP status コードが入る
	json.NewEncoder(w).Encode(error)
	return
}

func signup(w rest.ResponseWriter, r *rest.Request) {
	var user User

	// https://golang.org/pkg/encoding/json/#NewDecoder
	json.NewDecoder(r.Body).Decode(&user)

	if user.Email == "" {
		rest.Error(w, "Email は必須です。", http.StatusInternalServerError)
		return
	}

	if user.Password == "" {
		rest.Error(w, "パスワードは必須です。", http.StatusInternalServerError)
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
		rest.Error(w, "サーバーエラー", http.StatusInternalServerError)
		return
	}

	// DB に登録できたらパスワードをからにしておく
	user.Password = ""
	w.Header().Set("Content-Type", "application/json")

	// JSON 形式で結果を返却
	responseByJSON(w, user)
}

func login(w rest.ResponseWriter, r *rest.Request) {
	var res Res
	var user User

	var sendData map[string]interface{}
	err := r.DecodeJsonPayload(&sendData)
	if err != nil {
		rest.Error(w, "サーバーエラー", http.StatusInternalServerError)
		return
	}

	fake, err := json.Marshal(sendData["UserInfo"])
	if err != nil {
		rest.Error(w, "サーバーエラー1", http.StatusInternalServerError)
		return
	}
	err = json.Unmarshal(fake, &user)
	if err != nil {
		rest.Error(w, "サーバーエラー2", http.StatusInternalServerError)
		return
	}

	err = loginSameFunc(w, r, user)
	if err != nil {
		return
	}

	w.WriteHeader(http.StatusOK)

	res.Status = "ok"

	responseByJSON(w, res)
}

func loginCheck(w rest.ResponseWriter, r *rest.Request, user User) {
	err := loginSameFunc(w, r, user)
	if err != nil {
		return
	}
}

func loginSameFunc(w rest.ResponseWriter, r *rest.Request, user User) (err error) {

	if user.Email == "" {
		rest.Error(w, "Email は必須です。", http.StatusBadRequest)
		return err
	}

	if user.Password == "" {
		rest.Error(w, "パスワードは、必須です。", http.StatusBadRequest)
		return err
	}

	password := user.Password

	// 認証キー(Emal)のユーザー情報をDBから取得
	row := db.QueryRow("SELECT * FROM USERS WHERE email=$1;", user.Email)
	// ハッシュ化している
	err = row.Scan(&user.ID, &user.Email, &user.Password)

	if err != nil {
		if err == sql.ErrNoRows { // https://golang.org/pkg/database/sql/#pkg-variables
			rest.Error(w, "ユーザが存在しません。", http.StatusBadRequest)
		} else {
			rest.Error(w, "サーバーエラー4", http.StatusInternalServerError)
		}
		return err
	}

	hasedPassword := user.Password

	err = bcrypt.CompareHashAndPassword([]byte(hasedPassword), []byte(password))

	if err != nil {
		rest.Error(w, "無効なパスワードです。", http.StatusBadRequest)
		return err
	}
	return nil
}
