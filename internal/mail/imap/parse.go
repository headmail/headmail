package imap

import (
	"fmt"
	"strings"
)

func parseAddress(s string) (string, error) {
	s = removeComments(s)
	t := strings.SplitN(s, ";", 2)
	// ../rfc/3464:513 ../rfc/6533:250
	addrType := strings.ToLower(strings.TrimSpace(t[0]))
	if len(t) != 2 {
		return "", fmt.Errorf("missing semicolon that splits address type and address")
	} else if addrType != "rfc822" {
		return "", fmt.Errorf("unrecognized address type %q, expected rfc822", addrType)
	}
	return strings.TrimSpace(t[1]), nil
}

func removeComments(s string) string {
	n := 0
	r := ""
	for _, c := range s {
		if c == '(' {
			n++
		} else if c == ')' && n > 0 {
			n--
		} else if n == 0 {
			r += string(c)
		}
	}
	return r
}
