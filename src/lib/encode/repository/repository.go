package repository

import "net/url"

// Encode encode the repository name
func Encode(repo string) string {
	return url.PathEscape(url.PathEscape(repo))
}
