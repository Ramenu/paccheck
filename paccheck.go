package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
)


const (
	ARCH_LINUX_URL = "https://archlinux.org"
	REGEX = "(?s)<div id=\"news\">.*?<h4>\\s*?<a href=\"[/\\w-]*\"\\s*?title=\"[\\w\\W]*?\">([\\w\\W]*?)</a>.*?<div class=\"article-content\">\\s*?<p>(.*?)</p>\\s*</div>"
	RED    = "\033[31m"
	GREEN  = "\033[32m"
	YELLOW = "\033[33m"
	BLUE   = "\033[34m"
	CYAN = "\x1b[36m"
	RESET = "\033[0m"
	BOLD = "\033[1m"
)

func main() {

	fmt.Println(BOLD + BLUE + ":: " + RESET + BOLD + "Checking Arch Linux news..." + RESET)
	body, err := fetch(ARCH_LINUX_URL)

	if err != nil {
		panic(err)
	}

	r, err := regexp.Compile(REGEX)

	if err != nil {
		panic(err)
	}

	feed := r.FindStringSubmatch(body)[2]

	// check if the news feed has been saved locally, if so we can assume that
	// the user has seen the update so we do not need to show it again
	updated := false

	// define the path to the file
	filename := filepath.Join(os.Getenv("HOME"), ".paccheck", "news")

	// create the '.paccheck' directory if it doesn't exist
	if _, err := os.Stat(filepath.Dir(filename)); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(filename), 0755); err != nil {
			panic(err)
		}

		// create/open the file
		file, err := os.Create(filename)

		if err != nil {
			panic(err)
		}

		file.Close()

	} else {

		// read the contents of the file
		bytes, err := os.ReadFile(filename)
		if err != nil {
			panic(err)
		}

		if string(bytes) == feed {
			updated = true
		}
	}

	// news feed hasn't been updated so we can exit
	if updated {
		return
	}

	plainTitle := r.FindStringSubmatch(body)[1]
	dashLine := "\n"
	for range plainTitle {
		dashLine += "-"
	}
	title := BOLD + YELLOW + plainTitle + RESET + BOLD + dashLine + RESET + "\n\n"

	styledFeed := highlightTag(feed, "code", CYAN)
	styledFeed = highlightTag(styledFeed, "h2", BOLD)
	
	// remove all HTML tags
	r, err = regexp.Compile("<.*?>")

	if err != nil {
		panic(err)
	}

	styledFeed = r.ReplaceAllString(styledFeed, "")
	fmt.Println("\n" + title + styledFeed)
	fmt.Print(BOLD + "\nAcknowledge and save this news feed?" + RESET + " (y/n) ")

	var ack string
	_, err = fmt.Scan(&ack)

	if err != nil {
		panic(err)
	}

	// save the news feed if the user types 'y'
	if ack == "y" {
		// update the file with the new news feed
		err = os.WriteFile(filename, []byte(feed), 0644)

		if err != nil {
			panic(err)
		}
	}
}

func highlightTag(body string, tagName string, color string) string {
	regex := fmt.Sprintf("<%s>", tagName)
	r, err := regexp.Compile(regex)

	if err != nil {
		panic(err)
	}

	body = r.ReplaceAllString(body, color)
	regex = fmt.Sprintf("</%s>", tagName)
	r, err = regexp.Compile(regex)

	if err != nil {
		panic(err)
	}

	return r.ReplaceAllString(body, RESET)
}

func fetch(url string) (string, error) {
	resp, err := http.Get(url)

	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	html, err := io.ReadAll(resp.Body)

	if err != nil {
		return "", err
	}

	return string(html), nil
}
