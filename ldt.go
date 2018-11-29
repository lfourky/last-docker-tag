package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"
)

type Response struct {
	Name string   `json:"name"`
	Tags []string `json:"tags"`
}

// If successful, will print the latest repository tag based on prefix or regex, or fallback to "latest" in any case
func main() {
	var (
		repoVersion = flag.String("v", "v2", "Docker repository version")
		repoHost    = flag.String("h", "", "Docker repository address; e.g. http://192.168.0.130:5000")
		repoName    = flag.String("n", "", "Full repository name; e.g. library/mysql")
		tagPrefix   = flag.String("p", "", "Target tag prefix; e.g. build_ if you're targeting build_23, build_24, etc...")
		tagRegex    = flag.String("r", "", `Target tag regex; e.g. ^build_\d+$`)
	)

	flag.Parse()

	if len(*tagPrefix) > 0 && len(*tagRegex) > 0 {
		fmt.Fprintln(os.Stderr, "can't set both prefix & regex")
		printDefault()
	}

	if len(*repoHost) == 0 {
		fmt.Fprintln(os.Stderr, "repository host needs to be set")
		printDefault()
	}

	if len(*repoName) == 0 {
		fmt.Fprintln(os.Stderr, "repository name needs to be set")
		printDefault()
	}

	uri := fmt.Sprintf("%s/%s/%s/tags/list", *repoHost, *repoVersion, *repoName)
	if _, err := url.Parse(uri); err != nil {
		fmt.Fprintf(os.Stderr, "failed to parse uri: %s\n", err.Error())
		printDefault()
	}

	client := http.Client{
		// Timeout & exit after 5 seconds of waiting for the tags to be fetched
		Timeout: time.Duration(time.Second * 5),
	}

	resp, err := client.Get(uri)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot GET %s: %s\n", uri, err.Error())
		printDefault()
	}
	defer resp.Body.Close()

	var response Response
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		fmt.Fprintf(os.Stderr, "err decoding response: %s\n", err.Error())
		printDefault()
	}

	if len(response.Tags) == 0 {
		fmt.Fprintln(os.Stderr, "couldn't find any tags")
		printDefault()
	}

	// Only one of the filtering will take place, either by prefix or by regex
	matchedTags := make([]string, 0, len(response.Tags))

	// Handle case where prefix was given
	if len(*tagPrefix) > 0 {
		for _, t := range response.Tags {
			if matchedByPrefix(t, *tagPrefix) {
				matchedTags = append(matchedTags, t)
			}
		}
	}

	// Handle case where regex was given
	if len(*tagRegex) > 0 {
		rgx, err := regexp.Compile(*tagRegex)
		if err != nil {
			fmt.Fprintf(os.Stderr, "cannot compile/parse regex: %s\n", err.Error())
			printDefault()
		}
		for _, t := range response.Tags {
			if matchedByRegex(t, rgx) {
				matchedTags = append(matchedTags, t)
			}
		}
	}

	if len(matchedTags) == 0 {
		fmt.Fprintln(os.Stderr, "couldn't match any tags")
		printDefault()
	}

	sort.Strings(matchedTags)
	fmt.Println(matchedTags[len(matchedTags)-1])
}

func matchedByPrefix(tag, prefix string) bool {
	return strings.HasPrefix(tag, prefix)
}

func matchedByRegex(tag string, rgx *regexp.Regexp) bool {
	return rgx.MatchString(tag)
}

func printDefault() {
	// In case there's an error, fallback to the value "latest" so as to give a valid docker tag in any case
	fmt.Println("latest")
	os.Exit(-1)
}
