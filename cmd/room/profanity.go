package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

type ProfanityChecker interface {
	// Check if the provided content contains any profanities.
	Check(content string) bool
}

func newProfanityChecker() ProfanityChecker {
	version := strings.ToLower(os.Getenv("VERSION"))

	switch version {
	case "", "v1":
		return newDummyProfanityChecker()
	case "v2":
		return newRegexProfanityChecker()
	default:
		panic(fmt.Sprintf("unsupported service version: %s", version))
	}
}

type dummyProfanityChecker struct{}

func newDummyProfanityChecker() *dummyProfanityChecker {
	return &dummyProfanityChecker{}
}

func (c *dummyProfanityChecker) Check(content string) bool {
	return false
}

type regexProfanityChecker struct {
	re *regexp.Regexp
}

func newRegexProfanityChecker() *regexProfanityChecker {
	return &regexProfanityChecker{
		re: regexp.MustCompile("(boogers|snot|poop|shucks|argh)"),
	}
}

func (c *regexProfanityChecker) Check(content string) bool {
	return c.re.MatchString(content)
}
