package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cgi"
	"os"

	"github.com/bitly/go-simplejson"

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
		rest.Get("/aps/routerCgi.cgi/PostData/:request", GetAWork),
		rest.Post("/aps/routerCgi.cgi/PostData", PostData),
	)
	if err != nil {
		log.Fatal(err)
	}
	api.SetApp(router)
	cgi.Serve(api.MakeHandler())
}

const filePath string = "works3.json"

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

	data := rawData.GetPath("Id", request)
	var result ID

	b, err := data.MarshalJSON()
	if err := json.Unmarshal(b, &result); err != nil {
		rest.Error(w, err.Error(), http.StatusBadRequest)
	}

	w.WriteJson(&result)
}

// PostData 送られてきたデータを書き込むAPI
func PostData(w rest.ResponseWriter, r *rest.Request) {
	CreatedWorkTag := "WorkTag6"
	DesignationUserID := "Id"

	var sendData map[string]interface{}
	err := r.DecodeJsonPayload(&sendData)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, err := getSimpleJSON(filePath)

	data.Get(DesignationUserID).SetPath([]string{CreatedWorkTag}, sendData)

	works := make([]ID, 0)
	for _, v := range data.MustMap() {
		fake, _ := json.Marshal(v)
		var box ID
		err = json.Unmarshal(fake, &box)
		if err != nil {
			rest.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		works = append(works, box)
	}

	o, _ := data.EncodePretty()
	err = writeFile(filePath, o)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
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
