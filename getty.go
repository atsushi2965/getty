package main

import (
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

const usage string = `Getty ya pirate!

Usage: getty <id | url>
`

var client = &http.Client{}

func download(url string, file *os.File) {
	res, err := client.Get(url)
	if err != nil || res.StatusCode != 200 {
		fmt.Println("Could not download the image")
		os.Exit(1)
	}

	defer res.Body.Close()

	io.Copy(file, res.Body)
}

func getty(id string, wg *sync.WaitGroup) {
	defer wg.Done()

	page := "https://www.gettyimages.com/detail/photo/-/" + id
	req, _ := http.NewRequest("GET", page, nil)
	res, err := client.Do(req)
	if err != nil {
		fmt.Println("Request error: ", err)
		os.Exit(1)
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	re := regexp.MustCompile(`[^"]+s=2048x2048[^"]+`)
	urls := re.FindAllString(string(body), -1)
	if urls == nil {
		fmt.Println("Could not find the image")
		os.Exit(1)
	}
	url := html.UnescapeString(urls[len(urls) - 1])

	resultPath := fmt.Sprintf("%s.jpg", id)
	resultFile, err := os.Create(resultPath)
	if err != nil {
		fmt.Println("Could not write the resulting image")
		os.Exit(1)
	}

	download(url, resultFile)
	defer resultFile.Close()
}

func main() {
	args := os.Args[1:]

	help := false

	if len(args) == 0 || args[0] == "help" {
		help = true
	}

	var links []string

	for _, arg := range args {
		if arg == "-h" || arg == "--help" {
			help = true
		} else if arg[0] == '-' {
			fmt.Println("Unknown argument " + arg)
			os.Exit(1)
		} else {
			links = append(links, arg)
		}
	}

	if help {
		fmt.Println(usage)
		return
	}

	var wg sync.WaitGroup

	for _, link := range links {
		var id string
		if isGettyID(link) {
			id = link
		} else {
			id = idFromURL(link)
		}

		wg.Add(1)
		go getty(id, &wg)
	}

	wg.Wait()
}

func idFromURL(link string) string {
	u, err := url.Parse(link)
	if err != nil {
		fmt.Println("Invalid url: " + link)
		os.Exit(1)
	}

	parts := strings.Split(u.Path, "/")
	id := parts[len(parts)-1]

	if !isGettyID(id) {
		fmt.Println("Invalid url: " + link)
		os.Exit(1)
	}

	return id
}

func isGettyID(id string) bool {
	if _, err := strconv.Atoi(id); err == nil {
		return true
	}

	return false
}
