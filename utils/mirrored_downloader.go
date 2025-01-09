package utils

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/net/html"
)

// DownloadAndMirror downloads the given URL and its assets recursively, with progress.
func (w *WgetValues) DownloadAndMirror() {
	// Ensure url is valid
	if !ValidateURL(w.Url) {
		return
	}

	websiteName, err := getWebsiteName(w.Url)
	if err != nil {
		fmt.Printf("Invalid URL: %v\n", err)
		return
	}

	// Start time
	if !w.MirrorStarted {
		startTime := time.Now()
		fmt.Printf("started at: %s\n", startTime.Format("2006-01-02 15:04:05"))

		w.MirrorStarted = true
	}

	rootDir := filepath.Join(w.OutPutDirectory, "www." + websiteName)

	visited := make(map[string]bool)
	queue := []string{w.Url}

	for len(queue) > 0 {
		currentURL := queue[0]
		queue = queue[1:]

		// Skip if already visited
		if visited[currentURL] {
			continue
		}
		visited[currentURL] = true

		// Fetch the page
		res, err := http.Get(currentURL)
		if err != nil {
			continue
		}
		defer res.Body.Close()

		// Print response status
		fmt.Printf("sending request, awaiting response... status %s\n", res.Status)

		// Parse HTML if content is text/html
		contentType := res.Header.Get("Content-Type")
		if strings.Contains(contentType, "text/html") {
			htmlData, err := io.ReadAll(res.Body)
			if err != nil {
				fmt.Printf("corrupted response body: %v\n", err)
				continue
			}

			fmt.Printf("saving website files to: %s\n", rootDir)

			// Save the HTML file
			err = saveFile(rootDir, "index.html", htmlData)
			if err != nil {
				fmt.Printf("Failed to save file: %v\n", err)
				continue
			}

			// Parse and extract asset links
			assets, newLinks := parseHTMLForAssets(currentURL, htmlData)
			queue = append(queue, newLinks...)

			// Download each asset with progress reporting
			for _, asset := range assets {
				if err := w.downloadAssetWithProgress(asset, rootDir); err != nil {
					fmt.Println("Avoiding broken link...")
				}
			}
		}
	}
}

// downloadAssetWithProgress downloads a single asset and saves it to the output directory with progress.
func(w *WgetValues) downloadAssetWithProgress(assetURL, rootDir string) error {
	res, err := http.Get(assetURL)
	if err != nil {
		return fmt.Errorf("avoiding broken asset link: %v", err)
	}
	defer res.Body.Close()

	// Reader
	var reader io.Reader = res.Body

	// Get the asset's size
	contentLength := res.ContentLength
	if contentLength == -1 {
		contentLength = 0
	}

	// Create directories based on URL path under the root directory
	parsedURL, err := url.Parse(assetURL)
	if err != nil {
		return fmt.Errorf("broken asset url: %v", err)
	}
	assetPath := filepath.Join(rootDir, parsedURL.Path)
	err = os.MkdirAll(filepath.Dir(assetPath), os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create directories: %v", err)
	}

	// Save the file with progress
	file, err := os.Create(assetPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	pr := &ProgressRecoder{
		Reader:           reader,
		Total:            contentLength,
		startTime:        time.Now(),
		ProgressFunction: w.ShowProgress,
	}

	// Read the response body
	_, err = io.Copy(file, pr)
	CheckError(err)

	// End time
	endTime := time.Now()
	fmt.Printf("\nDownload completed at: %s\n", endTime.Format("2006-01-02 15:04:05"))

	return nil
}

// parseHTMLForAssets parses HTML content and extracts asset URLs and links.
func parseHTMLForAssets(baseURL string, htmlData []byte) (assets []string, links []string) {
	doc, err := html.Parse(strings.NewReader(string(htmlData)))
	if err != nil {
		fmt.Printf("Failed to parse HTML: %v\n", err)
		return nil, nil
	}

	var extract func(*html.Node)
	extract = func(n *html.Node) {
		if n.Type == html.ElementNode {
			var link string
			switch n.Data {
			case "link", "img", "script":
				for _, attr := range n.Attr {
					if attr.Key == "href" || attr.Key == "src" {
						link = attr.Val
					}
				}
			case "a":
				for _, attr := range n.Attr {
					if attr.Key == "href" {
						link = attr.Val
					}
				}
			}

			// Normalize the link
			if link != "" {
				fullURL := normalizeURL(baseURL, link)
				if strings.Contains(fullURL, baseURL) {
					links = append(links, fullURL) // Add internal links for recursion
				} else {
					assets = append(assets, fullURL) // Treat external links as assets
				}
			}
		}

		// Recursively process child nodes
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			extract(c)
		}
	}

	extract(doc)
	return assets, links
}

// normalizeURL resolves relative URLs against a base URL.
func normalizeURL(baseURL, relative string) string {
	u, err := url.Parse(relative)
	if err != nil {
		return ""
	}
	base, err := url.Parse(baseURL)
	if err != nil {
		return ""
	}
	return base.ResolveReference(u).String()
}

// saveFile saves the given data to a specified directory and filename.
func saveFile(basePath, fileName string, data []byte) error {
	// Create the directory if it doesn't exist
	err := os.MkdirAll(basePath, os.ModePerm)
	if err != nil {
		return err
	}

	// Create the full file path
	filePath := filepath.Join(basePath, fileName)

	// Write data to the file
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(data)
	return err
}

// getWebsiteName extracts the website name from the given URL.
func getWebsiteName(rawURL string) (string, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %v", err)
	}
	return strings.TrimPrefix(parsedURL.Hostname(), "www."), nil
}
