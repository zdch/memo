package util

import (
	"regexp"
)

func ValidateBitcoinLegacyAddress(addr string) bool {
	var re = regexp.MustCompile(`[a-zA-Z0-9]{26,35}`)
	return re.MatchString(addr)
}

func ValidateImageHeight(height string) bool {
	var re = regexp.MustCompile(`[0-9]{1,4}`)
	return re.MatchString(height)
}