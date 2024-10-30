package main

import (
	"fmt"
	"sort"
	"strconv"
)

type ByNumericSuffix []string

func (s ByNumericSuffix) Len() int {
	return len(s)
}

func (s ByNumericSuffix) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s ByNumericSuffix) Less(i, j int) bool {
	numI := extractNumericSuffix(s[i])
	numJ := extractNumericSuffix(s[j])
	return numI < numJ
}

func extractNumericSuffix(s string) int {
	// Найти последний числовой суффикс в строке
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] < '0' || s[i] > '9' {
			if i == len(s)-1 {
				return 0
			}
			num, err := strconv.Atoi(s[i+1:])
			if err != nil {
				return 0
			}
			return num
		}
	}
	num, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return num
}

func main() {
	ports := []string{"COM1", "COM10", "COM11", "COM12", "COM13", "COM14", "COM15", "COM16", "COM3", "COM2"}
	fmt.Println(ports)
	sort.Sort(ByNumericSuffix(ports))
	fmt.Println(ports)
}
