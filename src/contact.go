package main

import "sort"

type Contact struct {
	DisplayName string
	PhoneNumber string
}
type Contacts []Contact

func (c Contacts) Len() int      { return len(c) }
func (c Contacts) Swap(i, j int) { c[i], c[j] = c[j], c[i] }
func (c Contacts) Less(i, j int) bool {
	return sort.StringSlice([]string{c[i].DisplayName, c[j].DisplayName}).Less(0, 1)
}

//filter out duplicate entries for the same number, keep the last definition
func FilterDuplicateNumbers(contacts Contacts) (nContacts Contacts) {
	numberMap := make(map[string]string)
	for _, v := range contacts {
		numberMap[v.PhoneNumber] = v.DisplayName
	}
	for number, name := range numberMap {
		nContacts = append(nContacts, Contact{
			DisplayName: name,
			PhoneNumber: number,
		})
	}
	return nContacts
}
