package checks

import (
	"crypto/tls"
	"io"
	"math/rand"
	"net/http"
	"regexp"
	"strconv"
	"github.com/corpix/uarand"
)

type Web struct {
	Service
	Url    []urlData
	Scheme string
}

type urlData struct {
	Path        string
	Status      int
	Diff        int
	Regex       string
	CompareFile string // TODO implement
}

func (c Web) Run(teamID uint, boxIp string, res chan Result, service Service) {
	u := c.Url[rand.Intn(len(c.Url))]

	// Select a random user agent from uarand
	ua := uarand.GetRandom()

	tr := &http.Transport{
		MaxIdleConns:      1,
		IdleConnTimeout:   GlobalTimeout, // address this
		DisableKeepAlives: true,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	client := &http.Client{Transport: tr}
	req, err := http.NewRequest("GET", c.Scheme + "://" + boxIp + ":" + strconv.Itoa(service.Port) + u.Path, nil)
	if err != nil {
		res <- Result{
			Error: "failed to create request",
			Debug: err.Error(),
		}
		return
	}

	// Set the User-Agent header to the random user agent
	req.Header.Set("User-Agent", ua)

	resp, err := client.Do(req)
	if err != nil {
		res <- Result{
			Error: "web request errored out",
			Debug: err.Error() + " for url " + u.Path,
		}
		return
	}

	if u.Status != 0 && resp.StatusCode != u.Status {
		res <- Result{
			Error: "status returned by webserver was incorrect",
			Debug: "status was " + strconv.Itoa(resp.StatusCode) + " wanted " + strconv.Itoa(u.Status) + " for url " + u.Path,
		}
		return
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		res <- Result{
			Error: "error reading page content",
			Debug: "error was '" + err.Error() + "' for url " + u.Path,
		}
		return
	}

	if u.Regex != "" {
		re, err := regexp.Compile(u.Regex)
		if err != nil {
			res <- Result{
				Error: "error compiling regex to match for web page",
				Debug: err.Error(),
			}
			return
		}
		reFind := re.Find(body)
		if reFind == nil {
			res <- Result{
				Error: "didn't find regex on page",
				Debug: "couldn't find regex \"" + u.Regex + "\" for " + u.Path,
			}
			return
		} else {
			res <- Result{
				Status: true,
				Error:  "page matched regex!",
				Debug:  "matched regex \"" + u.Regex + "\" for " + u.Path,
			}
			return
		}

	}

	res <- Result{
		Status: true,
	}
}
