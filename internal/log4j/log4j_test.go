package log4j

import (
	"testing"
	"net/http"
	"os"
	"io"
	"strings"
)

// CheckFile(zipname, zipfile string) {

func DownloadFile(url string) (string, error) {

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Create the file
	filepath := url[(strings.LastIndex(url, "/") + 1):]
	out, err := os.Create(filepath)
	if err != nil {
		return "", err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return filepath, err
}


func TestCalculate(t *testing.T) {
	testCases := []struct {
		url       string
		version    string
	}{
		{"https://github.com/elastic/apm-agent-java/releases/download/v1.28.0/elastic-apm-java-aws-lambda-layer-1.28.0.zip", "log4j-core-2.12.1"},
		{"https://launcher.mojang.com/v1/objects/952438ac4e01b4d115c5fc38f891710c4941df29/server.jar", "log4j-core-2.0-beta9"},
	}
	for _, testCase := range testCases {
		filename, err := DownloadFile(testCase.url);
		if err != nil {
			t.Fatalf("Failed to Download: %v", err)
		}
		defer os.Remove(filename)
		if version := CheckFile(filename, filename); version  != testCase.version {
			 t.Errorf("Expected %s, got: %s", testCase.version, version)
		}


	}
}
