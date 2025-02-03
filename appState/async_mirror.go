package appState

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	utils "wget/wgetUtils"
)

func (app *AppState) mirrorAsyncDownload(outputFileName, urlStr, directory string) error {
	app.ProcessedURLs.Lock()
	if processed, exists := app.ProcessedURLs.URLs[urlStr]; exists && processed {
		app.ProcessedURLs.Unlock()
		return fmt.Errorf("URL already processed:\n%s", urlStr)
	}
	app.ProcessedURLs.Unlock()

	// Parse the URL to get the path components
	u, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("error parsing URL:\n%v", err)
	}

	// Create the necessary directories based on the URL path
	rootPath, err := utils.ExpandPath(directory)
	if err != nil {
		return err
	}

	pathComponents := strings.Split(strings.Trim(u.Path, "/"), "/")
	relativeDirPath := filepath.Join(pathComponents[:len(pathComponents)-1]...)
	fullDirPath := filepath.Join(rootPath, relativeDirPath)
	fileName := pathComponents[len(pathComponents)-1]

	resp, err := utils.HttpRequest(urlStr)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error: status %s\nurl: %s", resp.Status, urlStr)
	}

	contentType := resp.Header.Get("Content-Type")

	if outputFileName == "" {
		if fileName == "" || strings.HasSuffix(urlStr, "/") {
			fileName = "index.html"
		} else if contentType == "text/html" && !strings.HasSuffix(fileName, ".html") {
			fileName += ".html"
		}
		outputFileName = filepath.Join(fullDirPath, fileName)
	} else {
		if contentType == "text/html" && !strings.HasSuffix(outputFileName, ".html") {
			outputFileName += ".html"
		}
		outputFileName = filepath.Join(fullDirPath, outputFileName)
	}

	if fullDirPath != "" {
		if _, err := os.Stat(fullDirPath); os.IsNotExist(err) {
			err = os.MkdirAll(fullDirPath, 0o755)
			if err != nil {
				return fmt.Errorf("error creating path:\n%v", err)
			}
		}
	}
	if utils.FileExists(outputFileName) {
		return nil
	}

	var out *os.File
	out, err = os.Create(outputFileName)
	if err != nil {
		return fmt.Errorf("error creating file:\n%v", err)
	}
	defer out.Close()

	var reader io.Reader = resp.Body
	var totalSize int64

	// Get the content length for the progress (if available)
	if length := resp.Header.Get("Content-Length"); length != "" {
		totalSize, err = strconv.ParseInt(length, 10, 64)
		if err != nil {
			return fmt.Errorf("error parsing Content-Length:\n%v", err)
		}
	}

	buffer := make([]byte, 32*1024) // 32 KB buffer size
	var downloaded int64
	startTime := time.Now()

	// Download the file while showing progress
	for {
		n, err := reader.Read(buffer)
		if err != nil && err != io.EOF {
			return fmt.Errorf("error reading response body")
		}

		if n > 0 {
			if _, err := out.Write(buffer[:n]); err != nil {
				return fmt.Errorf("error writing to file:\n%v", err)
			}
			downloaded += int64(n)
			app.showProgress(downloaded, totalSize, startTime) // Display progress
		}

		if err == io.EOF {
			break
		}
	}

	fmt.Printf("\n\033[32mDownloaded [%s]\033[0m\n", urlStr)

	// Mark the URL as processed
	app.ProcessedURLs.Lock()
	app.ProcessedURLs.URLs[urlStr] = true
	app.ProcessedURLs.Unlock()

	return nil
}

// Update the ShowProgress function with the correct speed format
func (app *AppState) showProgress(progress, total int64, startTime time.Time) {
	const length = 50
	if total <= 0 {
		return
	}
	percent := float64(progress) / float64(total) * 100
	numBars := int((percent / 100) * length)

	// Calculate speed (bytes per second)
	elapsed := time.Since(startTime).Seconds()
	speed := float64(progress) / elapsed

	// Calculate estimated time remaining
	var eta string
	if speed > 0 {
		remaining := float64(total-progress) / speed
		eta = fmt.Sprintf("%02d:%02d:%02d", int(remaining/3600), int(remaining/60)%60, int(remaining)%60)
	} else {
		eta = "--:--:--"
	}

	// Print the output with custom format
	if !app.UrlArgs.WorkInBackground {
		out := fmt.Sprintf("%.2f KiB / %.2f KiB [%s%s] %.0f%% %s %s",
			float64(progress)/1024, float64(total)/1024,
			strings.Repeat("=", numBars), strings.Repeat(" ", length-numBars),
			percent, utils.FormatSpeed(speed/1024), eta)
		fmt.Printf("\r%s", out)
	}
}
