package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"flag"
)

var (
	name   = flag.String("n", "", "name of the go project")
	path   = flag.String("p", "github.com/advincze", "(std)path of the go project, e.g: github.com/name")
	gistID = flag.String("g", "hw", "predefined project (hw|srv|fileserv) or public gist id to use")
)

var gists = []struct {
	Name string
	ID   string
}{
	{"hw", "5530811"},
	{"srv", "ed2641eb8470f17b846e"},
	{"fileserv", "b7be0da8adeebdc97143"},
}

func main() {
	flag.Parse()

	if *name == "" {
		*name = fmt.Sprintf("test%d", time.Now().Unix())
	}

	gID := *gistID
	for _, gist := range gists {
		if gist.Name == *gistID {
			gID = gist.ID
		}
	}

	gist, err := getGist(gID)
	if err != nil {
		log.Println("error fetching gist", gID)
		return
	}

	gopath := os.Getenv("GOPATH")

	fullPath := gopath + "/src/" + *path + "/" + *name
	err = os.Mkdir(fullPath, 0755)
	if err != nil {
		log.Println("error creating dir", gID)
		return
	}

	for filename, content := range gist {
		err = ioutil.WriteFile(fullPath+"/"+filename, content, 0644)
		if err != nil {
			log.Println("error writing file", fullPath+"/"+filename, err)
			continue
		}
	}

	fmt.Println(fullPath)
}

func getGist(id string) (map[string][]byte, error) {
	gists := map[string][]byte{}
	resp, err := http.Get(fmt.Sprintf("https://api.github.com/gists/%s", id))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var v struct {
		Files map[string]struct {
			URL string `json:"raw_url"`
		} `json:"files"`
	}

	err = json.NewDecoder(resp.Body).Decode(&v)
	if err != nil {
		return nil, err
	}

	for filename, file := range v.Files {
		fileResp, err := http.Get(file.URL)
		if err != nil {
			log.Println("error fetching ", file.URL)
			continue
		}
		defer fileResp.Body.Close()

		data, err := ioutil.ReadAll(fileResp.Body)
		if err != nil {
			log.Println("error reading response ", file.URL)
			continue
		}
		gists[filename] = data
	}

	return gists, nil
}
