package forebears

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

var (
	ErrNeedTwoNames    = errors.New("The forebears.io API wants two name parts: firstname and surname")
	ErrWrongStatusCode = errors.New("The forebears.io API returned the wrong status code")
	ErrNoMatch         = errors.New("No matches for supplied string")
)

type ForebearsCountry struct {
	Jurisdiction string `json:"jurisdiction"`
	Percent      string `json:"percent"`
}

type ForebearsSphere struct {
	Sphere  string `json:"sphere"`
	Percent string `json:"percent"`
}

type ForebearsResult struct {
	Countries []ForebearsCountry `json:"countries"`
	Spheres   []ForebearsSphere  `json:"spheres"`
}

type Client struct {
	apiKey     string
	httpClient *http.Client
}

func NewClient(key string) *Client {
	return &Client{httpClient: &http.Client{}, apiKey: key}
}

func (c *Client) SetAPIKey(key string) {
	c.apiKey = key
}

func (c *Client) Name2Country(nameStr string) (string, error) {
	// If we had no location we try to find one from the given name
	//fmt.Printf("Trying to determine country from name \"%s\"\n", nameStr)
	parts := strings.Split(nameStr, " ")
	query := ""
	if len(parts) < 2 {
		return "", ErrNeedTwoNames
	}
	fn := strings.Join(parts[:len(parts)-1], " ")
	ln := parts[len(parts)-1]
	query = fmt.Sprintf("https://ono.4b.rs/v1/nat?key=%s&fn=%s&sn=%s",
		c.apiKey, url.QueryEscape(fn), url.QueryEscape(ln))
	//fmt.Printf("Requesting: %s\n", query)
	body, resp, err := c.httpGet(query, nil)
	if err != nil {
		//fmt.Printf("LookupName(): Error fetching data from forebears.io: %s\n", err.Error())
		return "", err
	}
	if resp.StatusCode != 200 {
		//fmt.Printf("LookupName(): Error fetching data from forebears.io: got status code %d\n", resp.StatusCode)
		return "", ErrWrongStatusCode
	}
	forebearsResult := ForebearsResult{}
	err = json.Unmarshal(body, &forebearsResult)
	if err != nil {
		//fmt.Printf("LookupName(): Error decoding JSON data from forebears.io: %s\n", err.Error())
		return "", err
	}
	//fmt.Printf("Got result from 4bears: %s\n", string(body))
	//fmt.Printf("Unmarshaled to: %v\n", forebearsResult)
	if len(forebearsResult.Countries) < 1 {
		return "", ErrNoMatch
	}
	locationString := forebearsResult.Countries[0].Jurisdiction
	//fmt.Printf("locationString = %s, looking up using OSM...\n", locationString)
	return locationString, nil
}

// HTTPGet returns response body ([]byte), HTTP headers (map[string][]string) and error
func (c *Client) httpGet(url string, basicAuth []string) ([]byte, *http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if len(basicAuth) > 0 {
		req.SetBasicAuth(basicAuth[0], basicAuth[1])
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return []byte{}, resp, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, resp, err
	}
	return body, resp, nil
}
