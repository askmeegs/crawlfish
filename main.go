package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"gopkg.in/russross/blackfriday.v2"
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

	var items []Content

	err := getJson(url, &items)
	if err != nil {
		return err
	}

	// if it's a directory, add it to crawl. if it's markdown, check if 404.
	for _, i := range items {
		if i.t == "dir" {
			fmt.Println("enqueuing %s", i.url)
			queue = append(queue, i.url) //enQ
		}
		if i.t == "file" && isMarkdown(i.name) {
			brokenLinks, err := brokenLinks(i.url)
			if err != nil {
				log.Printf("Could not check if 404 (%s) - %v ", i.url, err)
			}
			if len(brokenLinks) > 0 {
				results[i.url] = brokenLinks
			}
		}
	}

	return nil

}

func getJson(url string, target interface{}) error {
	r, err := myClient.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	return json.NewDecoder(r.Body).Decode(target)
}

// given a markdown url, return any broken links
func brokenLinks(url string) ([]string, error) {
	// get markdown contents from url

	resp, _ := http.Get(url)
	bytes, _ := ioutil.ReadAll(resp.Body)

	output := blackfriday.Run(bytes)
	fmt.Printf("%#v", output)

	// for each link ...
	// is 404 ?
	return []string{}, nil
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
