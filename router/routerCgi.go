package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cgi"
	"os"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/mitchellh/mapstructure"
)

func main() {
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
		rest.Post("/aps/routerCgi.cgi/PostData", PostData),
	)
	if err != nil {
		log.Fatal(err)
	}
	api.SetApp(router)
	cgi.Serve(api.MakeHandler())
}

const filePath string = "works.json"

func getAllData(w rest.ResponseWriter, r *rest.Request) {

	bytes, err := readFile(filePath)
	var bodyData ID

	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var Datas Data
	Datas, err = makeJSON(bytes, bodyData)

	if err != nil {
		rest.Error(w, err.Error(), http.StatusBadRequest)
	}

	w.WriteJson(&Datas)
}

func PostData(w rest.ResponseWriter, r *rest.Request) {
	var postData ID
	err := r.DecodeJsonPayload(&postData)
	//送られてきたデータをCountry型に落とし込む。
	//落とし込め得ない場合はたぶんエラー
	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	bytes, err := readFile(filePath)

	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var Datas Data
	Datas, err = makeJSON(bytes, postData)

	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = writeJSON(filePath, Datas)

	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteJson(&Datas)
}

func makeJSON(bytes []byte, body ID) (ids Data, err error) {
	var data map[string]interface{}
	var Datas Data

	if err := json.Unmarshal(bytes, &data); err != nil {
		return Datas, err
	}
	//mapにエンコード

	works := make([]ID, 1)
	err = mapstructure.Decode(data["Id"], &works)
	if err != nil {
		return Datas, err
	}

	if body.WorkTag != "" {
		works = append(works, body)
	}

	Datas.Id = works
	return Datas, nil
}

func writeJSON(Filename string, data Data) (err error) {
	//ファイルへの書き込み
	result, err := json.MarshalIndent(data, "", "	")
	if err != nil {
		return err
	}

	return writeFile(Filename, result)
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
