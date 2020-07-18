package apiserver

import (
	"regexp"
	"strconv"
)

var (
	notNumberRegexp = regexp.MustCompile("[^0-9]+")
	dateRegexp      = regexp.MustCompile("^(1[0-2]|[1-9])[\\/][0-9][0-9]$")
	codeRegexp      = regexp.MustCompile("^[0-9][0-9][0-9]$")
)

func IsCreditCard(s string) bool {
	sanitized := notNumberRegexp.ReplaceAllString(s, "")

	var (
		digit        string
		sum          int
		tmpNum       int
		shouldDouble bool
	)

	if len(sanitized) < 13 || len(sanitized) > 19 {
		return false
	}

	for i := len(sanitized) - 1; i >= 0; i-- {
		digit = sanitized[i:(i + 1)]
		tmpNum, _ = strconv.Atoi(digit)
		if shouldDouble {
			tmpNum *= 2
			if tmpNum >= 10 {
				sum += ((tmpNum % 10) + 1)
			} else {
				sum += tmpNum
			}
		} else {
			sum += tmpNum
		}
		shouldDouble = !shouldDouble
	}

	return sum%10 == 0
}

func IsCardDate(s string) bool {
	return dateRegexp.MatchString(s)
}

func IsCardCode(s string) bool {
	return codeRegexp.MatchString(s)
}
