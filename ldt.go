package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	version "github.com/hashicorp/go-version"
)

type RepositoryResponse struct {
	Name string   `json:"name"`
	Tags []string `json:"tags"`
}

// If successful, will print the latest repository tag based on prefix or regex, or fallback to "latest" in any case
func main() {
	var (
		repoVersion    = flag.String("v", "v2", "Docker repository version")
		repoHost       = flag.String("h", "", "Docker repository address; e.g. http://192.168.0.130:5000")
		imageName      = flag.String("n", "", "Full repository name; e.g. library/mysql")
		tagPrefix      = flag.String("p", "", "Tag prefix to strip when matching tags; e.g. build_ if you're targeting build_23, build_24, etc...")
		verboseLogging = flag.Bool("l", false, "Set to true if you want to log tag version parsing errors, otherwise they're skipped")
	)

	flag.Parse()

	log.SetOutput(os.Stderr)

	if len(*repoHost) == 0 {
		log.Println("repository host needs to be set")
		printDefault()
	}

	if len(*imageName) == 0 {
		log.Println("image name needs to be set")
		printDefault()
	}

	uri := fmt.Sprintf("%s/%s/%s/tags/list", *repoHost, *repoVersion, *imageName)
	if _, err := url.Parse(uri); err != nil {
		log.Printf("failed to parse uri: %s", err.Error())
		printDefault()
	}

	client := http.Client{
		// Timeout & exit after 5 seconds of waiting for the tags to be fetched
		Timeout: time.Duration(time.Second * 5),
	}

	resp, err := client.Get(uri)
	if err != nil {
		log.Printf("cannot GET %s: %s", uri, err.Error())
		printDefault()
	}
	defer resp.Body.Close()

	var response RepositoryResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		log.Printf("err decoding response: %s", err.Error())
		printDefault()
	}

	if len(response.Tags) == 0 {
		log.Printf("couldn't find any tags for %s", uri)
		printDefault()
	}

	matchedTags := make([]string, 0, len(response.Tags))

	// Handle case where prefix was given
	if len(*tagPrefix) > 0 {
		for _, t := range response.Tags {
			if strings.HasPrefix(t, *tagPrefix) {
				matchedTags = append(matchedTags, strings.TrimPrefix(t, *tagPrefix))
			}
		}
	}

	if len(matchedTags) == 0 {
		log.Printf("couldn't match any tags for %s", uri)
		printDefault()
	}

	// Sort for SemVer
	versions := make([]*version.Version, 0, len(matchedTags))
	for _, tag := range matchedTags {
		versionedTag, err := version.NewVersion(tag)
		if err != nil {
			if *verboseLogging {
				log.Printf("cannot parse tag: %s from %s, err: %s", tag, uri, err.Error())
			}
			continue
		}
		versions = append(versions, versionedTag)
	}

	if len(versions) == 0 {
		log.Printf("couldn't parse any versions from retrieved tags for %s", uri)
		printDefault()
	}

	sort.Sort(version.Collection(versions))

	// Get the last tag from the sorted list
	targetTag := versions[len(versions)-1].Original()
	if len(*tagPrefix) > 0 {
		// Append the prefix back to the version number if the tag prefix was set
		fmt.Printf("%s%s", *tagPrefix, targetTag)
		return
	}

	// Otherwise, just print the SemVer
	fmt.Print(targetTag)
}

func printDefault() {
	// In case there's an error, fallback to the value "latest" so as to give a valid docker tag in any case
	fmt.Print("latest")
	os.Exit(-1)
}
