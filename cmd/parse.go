package cmd

import "regexp"

var reSmbConnStr = regexp.MustCompile(`(?m)smb://(?P<address>[^/]+)/(?P<share>[^/]+)/(?P<path>.+)`)

func parseSmbConnectionString(s string) (address string, share string, path string, valid bool) {
	match := reSmbConnStr.FindStringSubmatch(s)
	for i, name := range reSmbConnStr.SubexpNames() {
		if name == "address" {
			address = match[i]
		} else if name == "share" {
			share = match[i]
		} else if name == "path" {
			path = match[i]
		}
	}
	valid = !(len(match) == 0 || address == "" || share == "")
	return
}
