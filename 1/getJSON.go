package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cgi"

	"github.com/julienschmidt/httprouter"
	"github.com/mitchellh/mapstructure"
)

func main() {
	router := httprouter.New()

	router.GET("/aps/getJSON.cgi/getAll", getAllData)

	cgi.Serve(router)
}

func readFile(filename string) ([]byte, error) {
	byte, err := ioutil.ReadFile(filename)
	return byte, err
}

func getAllData(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	filePath := "works.json"

	bytes, err := ioutil.ReadFile(filePath)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	var IDs ID
	IDs, err = makeJSON(bytes)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, fmt.Sprintf("%+v\n", IDs))
}

func makeJSON(bytes []byte) (ids ID, err error) {
	var data map[string]interface{}
	var IDs ID

	if err := json.Unmarshal(bytes, &data); err != nil {
		fmt.Println(46)
		return IDs, err
	}
	//mapにエンコード

	var num int = len(data)
	// numはsliceの大きさを指定するため。
	//なお大きさはユーザーの作品数に依存する

	works := make([]WorkID, num)
	err = mapstructure.Decode(data["Id"], &works)
	if err != nil {
		fmt.Println(54)
		return IDs, err
	}
	IDs.WorkIDs = works
	return IDs, nil
}

// WorkID is each works data
type WorkID struct {
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

// ID is each person data
type ID struct {
	WorkIDs []WorkID `json:"WorkID"`
}
