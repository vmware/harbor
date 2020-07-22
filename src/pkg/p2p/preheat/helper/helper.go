package helper

import (
	"strings"
)

// ImageRepository represents the image repository name
// e.g: library/ubuntu:latest
type ImageRepository string

// Valid checks if the repository name is valid
func (ir ImageRepository) Valid() bool {
	if len(ir) == 0 {
		return false
	}

	trimName := strings.TrimSpace(string(ir))
	segments := strings.SplitN(trimName, "/", 2)
	if len(segments) != 2 {
		return false
	}

	nameAndTag := segments[1]
	subSegments := strings.SplitN(nameAndTag, ":", 2)
	if len(subSegments) != 2 {
		return false
	}

	return true
}

// Name returns the name of the image repository
func (ir ImageRepository) Name() string {
	// No check here, should call Valid() before calling name
	segments := strings.SplitN(string(ir), ":", 2)
	if len(segments) == 0 {
		return ""
	}

	return segments[0]
}

// Tag returns the tag of the image repository
func (ir ImageRepository) Tag() string {
	// No check here, should call Valid() before calling name
	segments := strings.SplitN(string(ir), ":", 2)
	if len(segments) < 2 {
		return ""
	}

	return segments[1]
}
