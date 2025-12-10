package main

import "strings"

func filterProfanity(msg string) string {
	var filteredWords = map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}

	words := strings.Split(msg, " ")
	for i, word := range words {
		_, isInList := filteredWords[strings.ToLower(word)]
		if isInList {
			words[i] = "****"
		}
	}

	return strings.Join(words, " ")
}
