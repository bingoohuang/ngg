package ss

import (
	"unicode"
	"unicode/utf8"
)

type CamelCaseEntry struct {
	Class CamelCaseEntryClass
	Entry string
}

func (c CamelCaseEntry) String() string {
	return c.Entry
}

type CamelCaseEntries []CamelCaseEntry

func (c CamelCaseEntries) Entries() []string {
	entries := []string{}
	for _, entry := range c {
		entries = append(entries, entry.Entry)
	}
	return entries
}

type CamelCaseEntryClass int

const (
	_ CamelCaseEntryClass = iota
	CamelCaseEntryLower
	CamelCaseEntryUpper
	CamelCaseEntryDigit
	CamelCaseEntryOther
)

// SplitCamelcase splits the camelcase word and returns a list of words. It also
// supports digits. Both lower camel case and upper camel case are supported.
// For more info please check: http://en.wikipedia.org/wiki/CamelCase
// https://github.com/proproto/camelcase/blob/master/camelcase.go
//
// Examples
//
//	"" =>                     [""]
//	"lowercase" =>            ["lowercase"]
//	"Class" =>                ["Class"]
//	"MyClass" =>              ["My", "Class"]
//	"MyC" =>                  ["My", "C"]
//	"HTML" =>                 ["HTML"]
//	"PDFLoader" =>            ["PDF", "Loader"]
//	"AString" =>              ["A", "String"]
//	"SimpleXMLParser" =>      ["Simple", "XML", "Parser"]
//	"vimRPCPlugin" =>         ["vim", "RPC", "Plugin"]
//	"GL11Version" =>          ["GL", "11", "Version"]
//	"99Bottles" =>            ["99", "Bottles"]
//	"May5" =>                 ["May", "5"]
//	"BFG9000" =>              ["BFG", "9000"]
//	"BöseÜberraschung" =>     ["Böse", "Überraschung"]
//	"Two  spaces" =>          ["Two", "  ", "spaces"]
//	"BadUTF8\xe2\xe2\xa1" =>  ["BadUTF8\xe2\xe2\xa1"]
//
// Splitting rules
//
//  1. If string is not valid UTF-8, return it without splitting as
//     single item array.
//  2. Assign all unicode characters into one of 4 sets: lower case
//     letters, upper case letters, numbers, and all other characters.
//  3. Iterate through characters of string, introducing splits
//     between adjacent characters that belong to different sets.
//  4. Iterate through array of split strings, and if a given string
//     is upper case:
//     if subsequent string is lower case:
//     move last character of upper case string to beginning of
//     lower case string
func SplitCamelcase(src string) (entries CamelCaseEntries) {
	// don't split invalid utf8
	if !utf8.ValidString(src) {
		return []CamelCaseEntry{{Entry: src}}
	}
	entries = []CamelCaseEntry{}
	var runes [][]rune
	var runeClasses []CamelCaseEntryClass
	var lastClass CamelCaseEntryClass
	class := CamelCaseEntryOther
	// split into fields based on class of unicode character
	for _, r := range src {
		switch true {
		case unicode.IsLower(r):
			class = CamelCaseEntryLower
		case unicode.IsUpper(r):
			class = CamelCaseEntryUpper
		case unicode.IsDigit(r):
			class = CamelCaseEntryDigit
		default:
			class = CamelCaseEntryOther
		}
		if class == lastClass {
			runes[len(runes)-1] = append(runes[len(runes)-1], r)
		} else {
			runes = append(runes, []rune{r})
			runeClasses = append(runeClasses, class)
		}
		lastClass = class
	}
	// handle upper case -> lower case sequences, e.g.
	// "PDFL", "oader" -> "PDF", "Loader"
	for i := 0; i < len(runes)-1; i++ {
		if unicode.IsUpper(runes[i][0]) && unicode.IsLower(runes[i+1][0]) {
			runes[i+1] = append([]rune{runes[i][len(runes[i])-1]}, runes[i+1]...)
			runes[i] = runes[i][:len(runes[i])-1]
		}
	}
	// construct []string from results
	for i, s := range runes {
		if len(s) > 0 {
			entries = append(entries, CamelCaseEntry{
				Class: runeClasses[i],
				Entry: string(s),
			})
		}
	}
	return entries
}
