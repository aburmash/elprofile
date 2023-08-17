package util

import (
	"bufio"
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
