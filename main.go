package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"io"
	"compress/gzip"
	"encoding/json"
	"strings"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func parse(c echo.Context) error {

	// Decode the jq query
	j, err := url.QueryUnescape(c.QueryParam("jq"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}
	j = strings.Replace(j, "±", "+", -1)

	// Decode the url
	url, err := url.QueryUnescape(c.QueryParam("url"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	// Create the request
	req, err := http.NewRequest(c.Request().Method, url, c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	// Add the headers
	for key := range c.Request().Header {
		req.Header.Add(key, c.Request().Header.Get(key))
	}

	// Add insecure request
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	// Request the url
	resp, err := client.Do(req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	var reader io.ReadCloser
    switch resp.Header.Get("Content-Encoding") {
    case "gzip":
        reader, err = gzip.NewReader(resp.Body)
        if err != nil {
            return c.JSON(http.StatusBadRequest, err)
        }
        reader.Close()
    default:
        reader = resp.Body
        if err != nil {
             return c.JSON(http.StatusBadRequest, err)
        }
    }

	// Convert the body to a slice of byte
	body, err := ioutil.ReadAll(reader)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}
	resp.Body.Close()

	// Copy it in a temp file
	file, err := ioutil.TempFile("/tmp", "jqInput_")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	defer os.Remove(file.Name())

	_, err = file.Write(body)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	// Execute jq
	cmd := exec.Command("bash", "-c", fmt.Sprintf("jq %s '%s'", j, file.Name()))
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	value, err := cmd.Output()
	if err != nil {
		return c.JSON(http.StatusBadRequest, fmt.Sprintf("Error: %s, %s", err, stderr.String()))
	}

    var js json.RawMessage
    if json.Unmarshal(value, &js) == nil {
       c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSONCharsetUTF8)
    }

	return c.String(http.StatusOK, string(value))
}

func main() {

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.Any("*", parse)
	e.Logger.Fatal(e.Start(":8080"))
}
