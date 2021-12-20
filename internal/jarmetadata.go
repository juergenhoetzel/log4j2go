package main

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"regexp"

	"net/http"
	"strings"

	"bytes"
	"github.com/juergenhoetzel/log4j2go/internal/log4j"
	"golang.org/x/net/html"
)

const parentURL = "https://repo1.maven.org/maven2/org/apache/logging/log4j/log4j-core/"

func getVersions() ([]string, error) {

	// Get the data
	resp, err := http.Get(parentURL)
	if err != nil {
		return []string{""}, err
	}
	defer resp.Body.Close()

	doc, err := html.Parse(resp.Body)
	if err != nil {
		log.Fatalf("Failed to parse: %v", err)
	}

	vrx, _ := regexp.Compile("([0-9][0-9a-z.-]+)/")
	versions := []string{}
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			if vrx.MatchString(n.FirstChild.Data) {
				versions = append(versions, (strings.TrimRight(n.FirstChild.Data, "/")))

			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)
	return versions, nil
}

const urlFormat = "https://repo1.maven.org/maven2/org/apache/logging/log4j/log4j-core/%s/log4j-core-%s.jar"

func getReader(version string) (*zip.Reader, error) {
	resp, err := http.Get(fmt.Sprintf(urlFormat, version, version))
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	ba, err := io.ReadAll(resp.Body)
	return zip.NewReader(bytes.NewReader(ba), int64(len(ba)))
}

func main() {
	vers, err := getVersions()
	if err != nil {
		log.Fatalf("Failed to get versions: %v", err)
	}
	fmt.Println("{")
	for _, s := range vers {
		r, err := getReader(s)
		if err != nil {
			log.Fatal(err)
		}

		key, ok := log4j.Log4jHash(r.File)
		if ok {
			fmt.Printf("%#v: %q,\n", key, fmt.Sprintf("log4j-core-%s", s))
		}
	}
	fmt.Println("}")
}
