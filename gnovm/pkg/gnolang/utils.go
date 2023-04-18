package gnolang

import "strings"

func contains(list []string, item string) bool {
	for _, i := range list {
		if i == item {
			return true
		}
	}
	return false
}

func endsWith(list []string, item string) bool {
	for _, i := range list {
		if strings.HasSuffix(item, i) {
			return true
		}
	}
	return false
}
