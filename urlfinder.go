package main

import "regexp"

func IsStreamUrl(txt string) bool {
	r := regexp.MustCompile("^https:\\/\\/(www|playbyplay)\\.sport5\\.co\\.il.*$")
	return r.MatchString(txt)
}
