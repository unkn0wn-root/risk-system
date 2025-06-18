package utils

import "regexp"

// maskPassword obscures password information in database URLs for secure logging.
// It replaces the password parameter value with asterisks to prevent credential exposure.
func MaskPassword(databaseURL string) string {
	re := regexp.MustCompile(`password=([^&\s]+)`)
	return re.ReplaceAllString(databaseURL, "password=***")
}
