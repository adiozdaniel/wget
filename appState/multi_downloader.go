package appState

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	utils "wget/wgetUtils"
)

func (app *AppState) DownloadMultipleFiles(filePath, outputFile, limit, directory string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("error opening file:\n%v", err)
	}
	defer file.Close()

	var wg sync.WaitGroup
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		url := strings.TrimSpace(scanner.Text())

		if url == "" {
			continue // Skip empty lines
		}
		wg.Add(1)
		go func(url string) error {
			defer wg.Done()
			err := app.AsyncDownload(outputFile, url, limit, directory)
			if err != nil {
				return err
			}
			return nil
		}(url)
	}
	wg.Wait()

	return nil
}

func (app *AppState) AsyncDownload(outputFileName, url, limit, directory string) error {
	path, err := utils.ExpandPath(directory)
	if err != nil {
		return err
	}

	resp, err := utils.HttpRequest(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error: status %s url:\n[%s]", resp.Status, url)
	}

	if outputFileName == "" {
		urlParts := strings.Split(url, "/")
		fileName := urlParts[len(urlParts)-1]
		outputFileName = filepath.Join(path, fileName)
	} else {
		outputFileName = filepath.Join(path, outputFileName)
	}

	if path != "" {
		err = os.MkdirAll(path, 0o755)
		if err != nil {
			return fmt.Errorf("error creating directory:\n%v", err)
		}
	}

	var out *os.File
	out, err = os.Create(outputFileName)
	if err != nil {
		return fmt.Errorf("error creating file:\n%v", err)
	}
	defer out.Close()

	var reader io.Reader = resp.Body
	if limit != "" {
		reader = utils.NewRateLimitedReader(resp.Body, limit)
	}

	buffer := make([]byte, 32*1024)
	fmt.Printf("Downloading.... [%s]\n", url)
	var downloaded int64
	for {
		n, err := reader.Read(buffer)
		if err != nil && err != io.EOF {
			return fmt.Errorf("oops! error reading response body")
		}

		if n > 0 {
			if _, err := out.Write(buffer[:n]); err != nil {
				return fmt.Errorf("error writing to file:\n%v", err)
			}
			downloaded += int64(n)
		}

		if err == io.EOF {
			break
		}
	}

	// endTime := time.Now()
	fmt.Printf("\033[32mDownloaded\033[0m [%s]\n", url)

	return nil
}
