package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/buger/jsonparser"
	md "github.com/nikitavoloboev/markdown-parser"
)

type Content struct {
	name string `json:"name"`
	t    string `json:"type"`
	url  string `json:"url"`
}

var myClient = &http.Client{Timeout: 10 * time.Second}
var results = map[string][]string{}
var queue = []string{""}

func main() {

	users := []string{"m-okeefe"}

	for _, u := range users {
		repos, err := getRepos(u)
		if err != nil {
			log.Fatal(err)
		}

		for _, repoName := range repos {
			for len(queue) > 0 {
				p := queue[0]
				// deQ!
				if len(queue) == 1 {
					queue = []string{}
				} else {
					queue = queue[1:]
				}

				err := SearchGithub(u, repoName, p) //start at repo root
				if err != nil {
					log.Fatal(err)
				}
				fmt.Printf("\nBadfiles is now lengh %d...", len(results))
			}
			fmt.Printf("USER=%s, REPO=%s, MARKDOWN FILES w/ 404 LINKS: %#v \n\n\n", u, repoName, results)
		}
	}

	fmt.Println("----- done! -----------")
}

// get all repos for user
// TODO IMPLEMENT FOR REAL
func getRepos(username string) ([]string, error) {
	return []string{"brokenlinks"}, nil
}

// NOTE - github rate limits to 10 requests per minute
// returns map[filepath][list-of-broken-links-in-file]
func SearchGithub(username string, repoName string, path string) error {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s", username, repoName, path)

	fmt.Printf("\n\n Sleeping, then searching: %s", url)
	time.Sleep(time.Second * 2)

	data, err := getJson(url)
	if err != nil {
		return err
	}

	jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		t, _ := jsonparser.GetString(value, "type")
		name, _ := jsonparser.GetString(value, "name")
		url, _ := jsonparser.GetString(value, "download_url")
		fmt.Printf("URL IS %s\n\n", url)
		// Recursive case
		if t == "dir" {
			fmt.Println("enqueuing %s", url)
			queue = append(queue, url) //enQ
		}
		// Base case - i'm a file
		if t == "file" && isMarkdown(name) {
			brokenLinks, err := brokenLinks(url)
			if err != nil {
				log.Printf("Could not check if 404 (%s) - %v ", url, err)
			}
			if len(brokenLinks) > 0 {
				results[url] = brokenLinks
			}
		}

	})

	return nil

}

func getJson(url string) ([]byte, error) {
	r, err := myClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()
	return ioutil.ReadAll(r.Body)
}

// given a markdown url, return any broken links
func brokenLinks(url string) ([]string, error) {

	broken := []string{}

	// get markdown contents from url
	body, err := md.DownloadURL(url)
	if err != nil {
		return broken, err
	}
	fmt.Printf("\n\n GOT MARKDOWN: %s \n\n", body)
	links := md.GetAllLinks(body)
	for k := range links {
		fmt.Printf("\n\nCHECKING IF link=%s is broken\n\n", k)
		broke, err := is404(k)
		if err != nil {
			return broken, fmt.Errorf("Could not check if link %s was 404: %v", k, err)
		}
		if broke {
			broken = append(broken, k)
		}
	}
	return broken, nil
}

func is404(url string) (bool, error) {
	resp, err := http.Get(url)
	if err != nil {
		return false, err
	}
	if resp.StatusCode == http.StatusNotFound {
		return true, nil
	}
	return false, nil
}

func isMarkdown(fn string) bool {
	if len(fn) < 3 {
		return false
	}
	return strings.HasSuffix(fn, ".md")
}
