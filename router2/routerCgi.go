package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cgi"
	"os"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/bitly/go-simplejson"
	"github.com/rs/xid"
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
		rest.Get("/aps/routerCgi.cgi/GetAWork/:request", GetAWork),
		rest.Post("/aps/routerCgi.cgi/PostData", PostData),
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
	var CreatedWorkTag string = (xid.New()).String()
	DesignationUserID := "Id"

	var sendData map[string]interface{}
	err := r.DecodeJsonPayload(&sendData)
	if err != nil {
		log.Fatal(109)
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

	data.Get(DesignationUserID).SetPath([]string{CreatedWorkTag}, sendData)
	data.Get(DesignationUserID).Get(CreatedWorkTag).Set("WorkTag", CreatedWorkTag)

	o, _ := data.EncodePretty()
	err = writeFile(filePath, o)
	if err != nil {
		log.Fatal(136)
		return
	}

	fake, err := data.Get(DesignationUserID).Get(CreatedWorkTag).MarshalJSON()
	if err != nil {
		log.Fatal(141)
		return
	}
	var resultJSON ID
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
