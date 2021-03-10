package main

import (
	"strings"
)

// hey yeah if it works and its stupid it aint stupid
const characterIndex = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789."

var (
	characterReference = []string{
		"\U000E0050", "\U000E0043", "\U000E0034", "\U000E0035",
		"\U000E002D", "\U000E002A", "\U000E005D", "\U000E002E",
		"\U000E0026", "\U000E0024", "\U000E0058", "\U000E004E",
		"\U000E0037", "\U000E0049", "\U000E0051", "\U000E0041",
		"\U000E0028", "\U000E0027", "\U000E004B", "\U000E005E",
		"\U000E0044", "\U000E0040", "\U000E004D", "\U000E0056",
		"\U000E0060", "\U000E0055", "\U000E0030", "\U000E0023",
		"\U000E0039", "\U000E004F", "\U000E0052", "\U000E002B",
		"\U000E0057", "\U000E003C", "\U000E0053", "\U000E005B",
		"\U000E003F", "\U000E0021", "\U000E003B", "\U000E0046",
		"\U000E0031", "\U000E0059", "\U000E003E", "\U000E0047",
		"\U000E005C", "\U000E003D", "\U000E0054", "\U000E0048",
		"\U000E005F", "\U000E0038", "\U000E003A", "\U000E002F",
		"\U000E005A", "\U000E0020", "\U000E0042", "\U000E0033",
		"\U000E0036", "\U000E004A", "\U000E0022", "\U000E0045",
		"\U000E0032", "\U000E002C", "\U000E0029", "\U000E0025",
		"\U000E004C",
	}
)

func GetCharacterIndex(s string) int {
	return strings.Index(characterIndex, s)
}

func StringToZWS(baseStr string) string {

	var completedStr string

	for i := 0; i < len(baseStr); i++ {
		r := characterReference[GetCharacterIndex(string(baseStr[i]))]
		if len(r) == 0 {
			// means we're trying to create a string without a character in the reference
			return ""
		}
		completedStr += r
	}

	return completedStr
}

// what else am i supposed to call it dumbass
func ZWSToString(encodedStr string) string {

	rL := []rune(encodedStr)
	var finalStr string

	for _, r := range rL {
		match := false
		for i, v := range characterReference {
			if []rune(v)[0] == r {
				match = true
				finalStr += string(characterIndex[i])
				break
			}
		}
		if !match {
			return ""
		}
	}

	return finalStr
}
