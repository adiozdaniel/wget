package appState

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	utils "wget/wgetUtils"
)

// taskManager calls to action methods depending on the passed flags
func (app *AppState) taskManager(err error) error {
	if err != nil {
		return err
	}

	// Mirror website handling
	if app.UrlArgs.Mirroring {
		err := app.DownloadAndMirror(app.UrlArgs.URL, app.UrlArgs.RejectFlag, app.UrlArgs.ConvertLinksFlag, app.UrlArgs.ExcludeFlag)
		if err != nil {
			return err
		}
		return nil
	}

	// If no file name is provided, derive it from the URL
	if app.UrlArgs.File == "" && app.UrlArgs.URL != "" {
		urlParts := strings.Split(app.UrlArgs.URL, "/")
		app.UrlArgs.File = urlParts[len(urlParts)-1]
	}

	// Handle the work-in-background flag
	if app.UrlArgs.WorkInBackground {
		err := app.DownloadInBackground(app.UrlArgs.File, app.UrlArgs.URL, app.UrlArgs.RateLimit)
		if err != nil {
			return err
		}
		return nil
	}

	// Handle multiple file downloads from sourcefile
	if app.UrlArgs.Sourcefile != "" {
		err := app.DownloadMultipleFiles(app.UrlArgs.Sourcefile, app.UrlArgs.File, app.UrlArgs.RateLimit, app.UrlArgs.Path)
		if err != nil {
			return err
		}
		return nil
	}

	// Ensure URL is provided
	if app.UrlArgs.URL == "" {
		return fmt.Errorf("error: url not provided")
	}

	// Start downloading the file
	err = app.singleDownloader(app.UrlArgs.File, app.UrlArgs.URL, app.UrlArgs.RateLimit, app.UrlArgs.Path)
	if err != nil {
		return err
	}
	return nil
}

// ParseArgs parses the command-line arguments and returns a UrlArgs struct
func (app *AppState) ParseArgs() error {
	mirrorMode := false
	track := false

	// Iterate over the command-line arguments manually
	for _, arg := range os.Args[1:] {
		if strings.HasPrefix(arg, "-O=") {
			app.UrlArgs.File = arg[len("-O="):]
		} else if strings.HasPrefix(arg, "-P=") {
			app.UrlArgs.Path = arg[len("-P="):]
		} else if strings.HasPrefix(arg, "--rate-limit=") {
			if err := utils.RateLimitValidator(arg); err != nil {
				return err
			}

			app.UrlArgs.RateLimit = arg[len("--rate-limit="):]
		} else if strings.HasPrefix(arg, "--mirror") {
			app.UrlArgs.Mirroring = true
			mirrorMode = true
		} else if strings.HasPrefix(arg, "--convert-links") {
			if !mirrorMode {
				return fmt.Errorf("error: --convert-links can only be used with --mirror")
			}
			app.UrlArgs.ConvertLinksFlag = true
		} else if strings.HasPrefix(arg, "-R=") || strings.HasPrefix(arg, "--reject=") {
			if !mirrorMode {
				return fmt.Errorf("error: --reject can only be used with --mirror")
			}
			if strings.HasPrefix(arg, "-R=") {
				app.UrlArgs.RejectFlag = arg[len("-R="):]
			} else {
				app.UrlArgs.RejectFlag = arg[len("--reject="):]
			}
		} else if strings.HasPrefix(arg, "-X=") || strings.HasPrefix(arg, "--exclude=") {
			if !mirrorMode {
				return fmt.Errorf("error: --exclude can only be used with --mirror")
			}
			if strings.HasPrefix(arg, "-X=") {
				app.UrlArgs.ExcludeFlag = arg[len("-X="):]
			} else {
				app.UrlArgs.ExcludeFlag = arg[len("--exclude="):]
			}
		} else if strings.HasPrefix(arg, "-B") {
			app.UrlArgs.WorkInBackground = true
		} else if strings.HasPrefix(arg, "-i=") {
			app.UrlArgs.Sourcefile = arg[len("-i="):]
			track = true
		} else if strings.HasPrefix(arg, "http") {
			app.UrlArgs.URL = arg
		} else {
			return fmt.Errorf("error: Unrecognized argument'%s'", arg)
		}
	}

	if app.UrlArgs.RateLimit != "" {
		if strings.ToLower(string(app.UrlArgs.RateLimit[len(app.UrlArgs.RateLimit)-1])) != "k" &&
			strings.ToLower(string(app.UrlArgs.RateLimit[len(app.UrlArgs.RateLimit)-1])) != "m" {
			return fmt.Errorf("invalid rateLimit")
		}
	}

	if app.UrlArgs.WorkInBackground {
		if app.UrlArgs.Sourcefile != "" || app.UrlArgs.Path != "" {
			return fmt.Errorf("-B flag shpuld not be used with -i or -P flags")
		}
	}

	// Check for invalid flag combinations if --mirror is provided
	if app.UrlArgs.Mirroring {
		if app.UrlArgs.File != "" || app.UrlArgs.Path != "" || app.UrlArgs.RateLimit != "" || app.UrlArgs.Sourcefile != "" || app.UrlArgs.WorkInBackground {
			return fmt.Errorf("error: --mirror can only be used with --convert-links, --reject, --exclude, and a URL. No other flags are allowed")
		}
	} else {
		if app.UrlArgs.ConvertLinksFlag || app.UrlArgs.RejectFlag != "" || app.UrlArgs.ExcludeFlag != "" {
			return fmt.Errorf("error: --convert-links, --reject, and --exclude can only be used with --mirror")
		}
	}

	// Ensure URL is provided
	if app.UrlArgs.URL == "" && !track {
		return fmt.Errorf("error: URL not provided")
	}

	// Validate the URL
	err := validateURL(app.UrlArgs.URL)
	if err != nil {
		return fmt.Errorf("error: invalid URL provided")
	}

	return nil
}

func validateURL(link string) error {
	_, err := url.ParseRequestURI(link)
	if err != nil {
		return fmt.Errorf("invalid url:\n%v", err)
	}
	return nil
}
