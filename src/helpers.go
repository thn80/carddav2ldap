package main

import (
	"bytes"
	"encoding/xml"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
)

type HttpClient struct {
	Http    *http.Client
	Request *http.Request
}

func HttpReq(client HttpClient, method string, path string) (body []byte, err error) {
	client.Request.URL.Path = path
	client.Request.Method = method
	client.Http.Do(client.Request)
	resp, err := client.Http.Do(client.Request)
	if err != nil {
		log.Fatalf("Failed to execute request")
		return []byte(""), err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Fatalf("Unexpected http response %d\n", resp.StatusCode)
		return []byte(""), errors.New("Unexpected http response")
	}
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read body: %v", err)
		return body, err
	}
	return body, nil
}

func XmlParse(data []byte, target interface{}) error {
	err := xml.Unmarshal(data, target)
	if err != nil {
		log.Fatalf("Failed to unpack: %v", err)
		return err
	}
	return nil
}

func NormalizeNumber(number []byte) []byte {
	number = bytes.TrimSpace(number)
	nNumber := []byte("")
	for _, d := range number {
		if (d >= '0' && d <= '9') || d == '*' || d == '#' {
			nNumber = append(nNumber, d)
		}
	}
	if bytes.HasPrefix(number, []byte("+49")) {
		nNumber = bytes.Replace(nNumber, []byte("49"), []byte("0"), 1) //+ is already removed
	}
	return nNumber
}
