package main

import "strings"

func isCommand(str, substr string) bool {
	trimmed := strings.ToUpper(strings.Trim(str, " \n\r"))
	return strings.Index(trimmed, substr) == 0
}
