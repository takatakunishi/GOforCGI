package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cgi"
	"os"

	"github.com/julienschmidt/httprouter"
	"github.com/mitchellh/mapstructure"
)

func main() {
	router := httprouter.New()

	router.POST("/aps/writeJSON.cgi/getAll", getAllData)
	cgi.Serve(router)
}

func readFile(filename string) ([]byte, error) {
	byte, err := ioutil.ReadFile(filename)
	return byte, err
}

func getAllData(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	filePath := "works.json"

	bytes, err := readFile(filePath)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var Datas Data
	Datas, err = makeJSON(bytes, body)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = writeJSON(filePath, Datas)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, fmt.Sprintf("%+v\n", Datas))
}

func makeJSON(bytes []byte, body []byte) (ids Data, err error) {
	var data map[string]interface{}
	var Datas Data

	if err := json.Unmarshal(bytes, &data); err != nil {
		return Datas, err
	}
	//mapにエンコード

	var num int = len(data) + 1
	// numはsliceの大きさを指定するため。
	//なお大きさはユーザーの作品数に依存する

	works := make([]ID, num)
	err = mapstructure.Decode(data["Id"], &works)
	if err != nil {
		return Datas, err
	}
	var work ID
	if err := json.Unmarshal(body, &work); err != nil {
		return Datas, err
	}
	// err = mapstructure.Decode(data, &work)
	if err != nil {
		return Datas, err
	}
	works = append(works, work)
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
