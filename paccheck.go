package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"html"
	"strings"
)


const (
	ARCH_LINUX_URL = "https://archlinux.org"
	REGEX = "(?s)<div id=\"news\">.*?<h4>\\s*?<a href=\"[/\\w-]*\"\\s*?title=\"[\\w\\W]*?\">([\\w\\W]*?)</a>.*?<div class=\"article-content\">\\s*?<p>(.*?)</div>"
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
	var filename string
	xdg_config, exists := os.LookupEnv("XDG_CONFIG_HOME")
	if !exists {
		filename = filepath.Join(os.Getenv("HOME"), ".config", ".paccheck", "news")
	} else {
		filename = filepath.Join(xdg_config, ".paccheck", "news")
	}

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
	
	// regex for capturing all HTML tags
	r, err = regexp.Compile("<.*?>")

	if err != nil {
		panic(err)
	}

	// remove all HTML tags
	styledFeed = r.ReplaceAllString(styledFeed, "")

	// unescape HTML entities
	title = html.UnescapeString(title)
	styledFeed = html.UnescapeString(styledFeed)

	// Trim leading and trailing whitespace
	styledFeed = strings.TrimSpace(styledFeed)

	fmt.Println("\n" + title + styledFeed)
	fmt.Print(BOLD + "Acknowledge and save this news feed?" + RESET + " (y/n) ")

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

	// Check if it's a title, if it is prepend a new line
	// so that it looks better
	if tagName == "h2" {
		body = r.ReplaceAllString(body, "\n" + color)
	} else {
		body = r.ReplaceAllString(body, color)
	}

	regex = fmt.Sprintf("</%s>", tagName)
	r, err = regexp.Compile(regex)

	if err != nil {
		panic(err)
	}

	return r.ReplaceAllString(body, RESET)
}

// Returns a string consisting of the HTML page from `url`
// and an error in case something went wrong.
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
