package util

import (
	"bufio"
	"regexp"
	"strings"
)

func BytesToArray(input []byte) []string {
	var ret []string

	scanner := bufio.NewScanner(strings.NewReader(string(input)))
	for scanner.Scan() {
		ret = append(ret, scanner.Text())
	}
	return ret
}

func ArrayToMap(input []string) map[string]bool {
	ret := make(map[string]bool)

	for i := 0; i < len(input); i++ {
		key := input[i]
		ret[key] = true
	}

	return ret
}

func ArrayMatch(find string, a []string) []string {
	var ret []string

	for _, k := range a {
		match, _ := regexp.MatchString(find, k)
		if match {
			ret = append(ret, k)
		}
	}

	return ret
}

func ArrayNotMatch(find string, a []string) []string {
	var ret []string

	for _, k := range a {
		match, _ := regexp.MatchString(find, k)
		if !match {
			ret = append(ret, k)
		}
	}

	return ret
}
