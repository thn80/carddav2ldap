package main

import (
	"bytes"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"
)

type CarddavData struct {
	Resp []struct {
		Href        string `xml:"href"`
		DisplayName string `xml:"propstat>prop>displayname"`
	} `xml:"response"`
}
type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

//see https://tools.ietf.org/html/rfc6350
func ParseVcard(vcardB []byte, displayName string) (contacts Contacts) {
	lines := bytes.Split(vcardB, []byte("\r\n"))
	var tmpNumbers [][]string
	var tmpFallbackName string
	var tmpFallbackOrg string
	var tmpFallbackFN string
	for _, line := range lines {
		parts := bytes.SplitN(line, []byte(":"), 2)
		key := parts[0]
		value := []byte("")
		if len(parts) > 1 { //handle empty values
			value = parts[1]
		}

		if bytes.Compare(line, []byte("BEGIN:VCARD")) == 0 {
			tmpNumbers = [][]string{}
			tmpFallbackName = ""
			tmpFallbackOrg = ""
			tmpFallbackFN = ""
			continue
		}

		keyParts := bytes.Split(key, []byte(";"))
		switch string(keyParts[0]) {
		case "TEL":
			tmpNumbers = append(tmpNumbers, []string{"", string(NormalizeNumber(value))})
			if len(keyParts) >= 2 && bytes.HasPrefix(keyParts[1], []byte("TYPE=")) {
				tmpNumbers[len(tmpNumbers)-1][0] = string(keyParts[1][5:])
			}
		case "N":
			nameParts := strings.Split(string(value), ";")
			if len(nameParts) >= 2 {
				nameParts[0], nameParts[1] = nameParts[1], nameParts[0]
			}
			tmpFallbackName = strings.Join(nameParts, " ")
		case "ORG":
			tmpFallbackOrg = strings.ReplaceAll(string(value), ";", " ")
		case "FN":
			tmpFallbackFN = string(value)
		}

		if bytes.Compare(line, []byte("END:VCARD")) == 0 {
			displayName = strings.TrimSpace(displayName)
			tmpFallbackName = strings.TrimSpace(tmpFallbackName)
			tmpFallbackOrg = strings.TrimSpace(tmpFallbackOrg)
			tmpFallbackFN = strings.TrimSpace(tmpFallbackFN)
			if len(displayName) <= 0 { // build displayName from ORG, N and FN
				if len(tmpFallbackFN) > 0 {
					displayName = tmpFallbackFN
				} else if len(tmpFallbackName) > 0 {
					displayName = tmpFallbackName
				}
				if len(tmpFallbackOrg) > 0 && len(displayName) > 0 {
					displayName = tmpFallbackOrg + " - " + displayName
				} else if len(tmpFallbackOrg) > 0 {
					displayName = tmpFallbackOrg
				}
				if len(displayName) <= 0 { //skip vCards without displayName
					continue
				}
			}
			for _, v := range tmpNumbers {
				contact := Contact{
					DisplayName: displayName,
					PhoneNumber: v[1],
				}
				numberType := strings.TrimSpace(v[0])
				if len(numberType) > 0 && len(tmpNumbers) > 1 { //append numbertype to displayName for multible numbers and same contact
					contact.DisplayName = contact.DisplayName + " (" + numberType + ")"
				}
				contacts = append(contacts, contact)
				log.Print("Parsed contact\t", contact.DisplayName)
			}
		}
	}
	return contacts
}

func ReadContactBook(httpClient HttpClient, path string) (contacts Contacts, err error) {
	body, err := HttpReq(httpClient, "PROPFIND", path)
	if err != nil {
		log.Print("Failed to get contact book", path, err)
		return nil, err
	}
	cards := CarddavData{}
	if err = XmlParse(body, &cards); err != nil {
		log.Print("Failed to parse contact book", path, err)
		return nil, err
	}
	for _, v := range cards.Resp {
		vcardB, err := HttpReq(httpClient, "GET", v.Href)
		if err != nil {
			log.Print("Failed to get vCard", v.DisplayName, v.Href)
			continue
		}
		contacts = append(contacts, ParseVcard(vcardB, v.DisplayName)...)
	}
	return contacts, nil
}

func GetContacts(carddavUrl string, credentials Credentials) (contacts Contacts, err error) {
	httpClient := HttpClient{}
	tr := &http.Transport{
		MaxIdleConns:    10,
		IdleConnTimeout: 30 * time.Second,
	}
	httpClient.Http = &http.Client{Transport: tr}

	tmpl, err := template.New("URL").Parse(carddavUrl)
	if err != nil {
		log.Panic("Failed to parse url template", err)
	}
	urlB := bytes.Buffer{}
	if err = tmpl.Execute(&urlB, credentials); err != nil {
		log.Panic("Failed to eval template with user data", err)
	}

	if httpClient.Request, err = http.NewRequest("PROPFIND", urlB.String(), nil); err != nil {
		log.Panic("Failed to create http request", err)
	}
	httpClient.Request.Header.Add("Content-Type", "application/xml")
	httpClient.Request.SetBasicAuth(credentials.Username, credentials.Password)

	contactBooks := CarddavData{}
	body, err := HttpReq(httpClient, "PROPFIND", httpClient.Request.URL.Path)
	if err != nil {
		log.Print("Failed to request contact books", err)
	}
	if err = XmlParse(body, &contactBooks); err != nil {
		log.Print("Failed to parse contact book list", err)
	}

	contacts = Contacts{}
	for _, v := range contactBooks.Resp {
		if strings.Compare(v.DisplayName, "Personal Address Book") == 0 {
			log.Print("Skip contact book ", v.DisplayName)
			continue
		}
		log.Printf("Parse contact book: %s", v.DisplayName)
		nContacts, err := ReadContactBook(httpClient, v.Href)
		if err != nil {
			log.Print("Failed to parse contact book", v.DisplayName, v.Href)
			continue
		}
		contacts = append(contacts, nContacts...)
	}
	return contacts, nil
}
