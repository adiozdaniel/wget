package WgetUtils

import (
	"fmt"
	"math"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func IsRejected(url, rejectTypes string) bool {
	if rejectTypes == "" {
		return false
	}

	rejectedTypes := strings.Split(rejectTypes, ",")
	for _, ext := range rejectedTypes {
		if strings.HasSuffix(url, ext) {
			return true
		}
	}
	return false
}

func IsRejectedPath(url, pathRejects string) bool {
	if pathRejects == "" {
		return false
	}

	rejects := strings.Split(pathRejects, ",")
	for _, path := range rejects {
		if path[0] != '/' {
			continue
		}
		if contains(url, path) {
			return true
		}
	}

	return false
}

func contains(str, substr string) bool {
	for i := 0; i < len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func FileExists(path string) bool {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return err == nil
}

func ExtractDomain(urlStr string) (string, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return "", err
	}
	return u.Hostname(), nil
}

// isValidAttribute checks if an HTML tag attribute is valid for processing
func IsValidAttribute(tagName, attrKey string) bool {
	return (tagName == "a" && attrKey == "href") ||
		(tagName == "img" && attrKey == "src") ||
		(tagName == "script" && attrKey == "src") ||
		(tagName == "link" && attrKey == "href")
}

func HttpRequest(url string) (*http.Response, error) {
	// Create a new HTTP client
	client := &http.Client{}

	// Create a new request with a User-Agent header
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	// Set headers to mimic a Chrome browser
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.85 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Connection", "keep-alive")

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}

	return resp, err
}

// ExpandPath expands shorthand notations to full paths
func ExpandPath(path string) (string, error) {
	// 1. Expand `~` to the home directory
	if strings.HasPrefix(path, "~") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("error finding home directory:\n %v", err)
		}
		path = strings.Replace(path, "~", homeDir, 1)
	}

	// 2. Expand environment variables like $HOME, $USER, etc.
	path = os.ExpandEnv(path)

	// 3. Convert relative paths (./ or ../) to absolute paths
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("error getting absolute path:\n %v", err)
	}

	return absPath, nil
}

func roundToNearest(value float64) float64 {
	return math.Round(value*100) / 100
}

// Helper function to format the speed in a human-readable format
func FormatSpeed(speed float64) string {
	if speed > 1000000000 {
		res_gb := float64(speed) / 1000000000
		return fmt.Sprintf("~%.2fGB", roundToNearest(res_gb))
	} else if speed > 1000000 {
		res_gb := float64(speed) / 1000000
		return fmt.Sprintf("[~%.2fMB]", roundToNearest(res_gb))
	}
	return fmt.Sprintf("%.0fKiB", speed)
}

// SaveShowProgressState saves the showProgress state to a temporary file.
func SaveShowProgressState(tempConfigFile string, showProgress bool) error {
	data := []byte(strconv.FormatBool(showProgress))
	err := os.WriteFile(tempConfigFile, data, 0o644)
	if err != nil {
		return fmt.Errorf("error saving showProgress state: %v", err)
	}
	return nil
}

// LoadShowProgressState loads the showProgress state from the temporary file if it exists.
func LoadShowProgressState(tempConfigFile string) (bool, error) {
	if _, err := os.Stat(tempConfigFile); os.IsNotExist(err) {
		// File doesn't exist, return default true
		return true, nil
	}

	data, err := os.ReadFile(tempConfigFile)
	if err != nil {
		return false, fmt.Errorf("error reading showProgress state: %v", err)
	}

	// Parse the boolean value
	showProgress, err := strconv.ParseBool(string(data))
	if err != nil {
		return false, fmt.Errorf("error parsing showProgress state: %v", err)
	}

	// Delete the file after retrieving the state
	err = os.Remove(tempConfigFile)
	if err != nil {
		return false, fmt.Errorf("error deleting temp file: %v", err)
	}

	return showProgress, nil
}

func ResolveURL(base, rel string) string {
	// Remove fragment identifiers (anything starting with #)
	if fragmentIndex := strings.Index(rel, "#"); fragmentIndex != -1 {
		rel = rel[:fragmentIndex]
	}

	if strings.HasPrefix(rel, "http") {
		return rel
	}

	if strings.HasPrefix(rel, "//") {
		protocol := "http:"
		if strings.HasPrefix(base, "https") {
			protocol = "https:"
		}
		return protocol + rel
	}

	if strings.HasPrefix(rel, "/") {
		return strings.Join(strings.Split(base, "/")[:3], "/") + rel
	}
	if strings.HasPrefix(rel, "./") {
		return strings.Join(strings.Split(base, "/")[:3], "/") + rel[1:]
	}
	if strings.HasPrefix(rel, "//") && strings.Contains(rel[2:], "/") {
		baseParts := strings.Split(base, "/")
		return baseParts[0] + "//" + baseParts[2] + rel[1:]
	}

	baseParts := strings.Split(base, "/")
	return baseParts[0] + "//" + baseParts[2] + "/" + rel
}
