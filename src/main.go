package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"net"
	"os"
	"sort"

	"github.com/nmcclain/ldap"
)

//	"github.com/vjeantet/goldap/message"
//	ldap "github.com/vjeantet/ldapserver"

var contacts Contacts

func main() {
	var err error

	credentials := Credentials{
		Username: os.Getenv("HELSINKI_USER"),
		Password: os.Getenv("HELSINKI_PASSWORD"),
	}
	if len(credentials.Username) <= 0 || len(credentials.Password) <= 0 {
		log.Fatal("You need to provide the system environment variables 'HELSINKI_USER' and 'HELSINKI_PASSWORD'")
	}

	contacts, err = GetContacts("https://gw.helsinki.tools/SOGo/dav/{{.Username}}/Contacts/", credentials)
	if err != nil {
		log.Printf("Failed to get contacts", err)
	}
	contacts = FilterDuplicateNumbers(contacts)
	sort.Sort(contacts)
	f, err := os.Create("tbook.csv")
	defer f.Close()
	if err != nil {
		log.Fatal("Failed to open file ", err)
	}
	csvW := csv.NewWriter(f)
	for _, v := range contacts {
		fmt.Println(v.DisplayName, v.PhoneNumber)
		csvW.Write([]string{v.DisplayName, v.PhoneNumber, "none", "active", "", "", "", "", "", "", "", "false", "", "", "sip", ""})
	}

	s := ldap.NewServer()
	handler := ldapHandler{}
	s.BindFunc("", handler)
	s.SearchFunc("", handler)

	// start the server
	listen := "0.0.0.0:10389"
	log.Printf("Starting example LDAP server on %s", listen)
	if err := s.ListenAndServe(listen); err != nil {
		log.Fatalf("LDAP Server Failed: %s", err.Error())
	}
}

type ldapHandler struct {
}

func (h ldapHandler) Bind(bindDN, bindSimplePw string, conn net.Conn) (ldap.LDAPResultCode, error) {
	return ldap.LDAPResultSuccess, nil
}

func (h ldapHandler) Search(boundDN string, searchReq ldap.SearchRequest, conn net.Conn) (ldap.ServerSearchResult, error) {
	log.Print("SEARCH", boundDN, searchReq)
	log.Print(searchReq.Attributes)
	log.Print(searchReq.Filter)
	entries := []*ldap.Entry{}
	for k, v := range contacts {
		entries = append(entries, &ldap.Entry{"cn=" + string(k) + "," + searchReq.BaseDN, []*ldap.EntryAttribute{
			&ldap.EntryAttribute{"cn", []string{string(k)}},
			&ldap.EntryAttribute{"displayName", []string{string(v.DisplayName)}},
			&ldap.EntryAttribute{"telephoneNumber", []string{string(v.PhoneNumber)}},
			&ldap.EntryAttribute{"sn", []string{string(v.DisplayName)}},
		}})
	}
	return ldap.ServerSearchResult{entries, []string{}, searchReq.Controls, ldap.LDAPResultSuccess}, nil
}
