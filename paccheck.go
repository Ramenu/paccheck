package main

import (
	"fmt"
	"html"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)


const (
	ARCH_LINUX_URL = "https://archlinux.org"
	REGEX = `(?s)<h4>\s*?<a href=\"[/\w-]*\"\s*?title=\"[\w\W]*?\">([\w\W]*?)</a>.*?<div class=\"article-content\">\s*?<p>(.*?)</div>`
	RED    = "\033[31m"
	GREEN  = "\033[32m"
	YELLOW = "\033[33m"
	BLUE   = "\033[34m"
	CYAN = "\x1b[36m"
	RESET = "\033[0m"
	BOLD = "\033[1m"
	RW_R_R = 0644
	RWX_RX_RX = 0755
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

	// We need to find 2 submatches, one for the top article and one for the second. The second
	// article is needed to check if the local article saved matches the second article, if it doesnt
	// then we can be sure more than one article appeared since the last time the news feed was saved
	matches := r.FindAllStringSubmatch(body, 2)

	// no matches found?
	if matches == nil {
		errorMsg("an internal problem with paccheck occurred with the regex parsing. Please report this as a bug")
		os.Exit(1)
	}
	feed := matches[0][2]

	// check if the news feed has been saved locally, if so we can assume that
	// the user has seen the update so we do not need to show it again
	newsfeedPath := findPaccheckFile("news")
	justCreatedNewsfeed := false
	var localFeed string

	// create all the directories if it doesn't exist
	if _, err := os.Stat(filepath.Dir(newsfeedPath)); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(newsfeedPath), RWX_RX_RX); err != nil {
			panic(err)
		}

		// create/open the file
		file, err := os.Create(newsfeedPath)

		if err != nil {
			panic(err)
		}

		file.Close()
		justCreatedNewsfeed = true

	} else {

		// read the contents of the file
		bytes, err := os.ReadFile(newsfeedPath)
		if err != nil {
			panic(err)
		}

		localFeed = string(bytes)
		if localFeed == feed {
			return
		}
	}

	// this is a really terrible check, it doesnt even check for string bounds
	// but the only point of it is to check if the start of the previous feed is
	// the same as the local feed. Why can't we just compare the entire feed? Well,
	// sometimes Arch doesn't show the entire feed for older articles for example 
	// (you have to manually go to the URL and check it, which would be more expensive
	// to do and require a rewrite of the code). Therefore, it's much simpler to just
	// check if the prefix matches. It is very unlikely that prefixes of other articles
	// will be the same, and it is unlikely that the article is less than 10 characters long.
	previousFeed := matches[1][2][:10]
	multipleUpdatesOccurred := false

	// we need to check if the next match is the same article
	// present in the local news feed. If it isnt, this means multiple
	// updates occurred and we should notify the user of this
	if !strings.HasPrefix(localFeed, previousFeed) {
		multipleUpdatesOccurred = true
	}

	plainTitle := matches[0][1]
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
		err = os.WriteFile(newsfeedPath, []byte(feed), RW_R_R)

		if err != nil {
			panic(err)
		}
	}
	if multipleUpdatesOccurred && !justCreatedNewsfeed {
		noteMsg("More than one news alert notifications have occurred while you were away. Please visit https://archlinux.org for more information.")
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

// Returns the full path to `filename`.
func findPaccheckFile(filename string) string {
	var path string

	// obey XDG standard if the variable is set
	xdg_config, exists := os.LookupEnv("XDG_CONFIG_HOME")
	if !exists {
		path = filepath.Join(os.Getenv("HOME"), ".config", ".paccheck", "news")
	} else {
		path = filepath.Join(xdg_config, ".paccheck", "news")
	}
	return path
}

// Prints an error message `msg` to stderr.
func errorMsg(msg string) {
	fmt.Fprintf(os.Stderr, BOLD + RED + "error" + RESET + BOLD + ":" + RESET + "%s\n", msg)
}

// Prints a notification message `msg` to stdout.
func noteMsg(msg string) {
	fmt.Println(BOLD + YELLOW + "note" + RESET + BOLD + ": " + RESET + msg)
}

