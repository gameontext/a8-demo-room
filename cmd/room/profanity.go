package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

////////////////////////////////////////////////////////////////////
//        .-"-.            .-"-.            .-"-.           .-"-.
//      _/_-.-_\_        _/.-.-.\_        _/.-.-.\_       _/.-.-.\_
//     / __} {__ \      /|( o o )|\      ( ( o o ) )     ( ( o o ) )
//    / //  "  \\ \    | //  "  \\ |      |/  "  \|       |/  "  \|
//   / / \'---'/ \ \  / / \'---'/ \ \      \'/^\'/         \ .-. /
//   \ \_/`"""`\_/ /  \ \_/`"""`\_/ /      /`\ /`\         /`"""`\
//    \           /    \           /      /  /|\  \       /       \
////////////////////////////////////////////////////////////////////

var profanities = []string{
	"boogers",
	"snot",
	"poop",
	"shucks",
	"argh",
	"dang",
	"boob",
	"crap",
	"woo",
	"merde",
}

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
	regex := fmt.Sprintf("(%s)", strings.Join(profanities, "|"))
	return &regexProfanityChecker{
		re: regexp.MustCompile(regex),
	}
}

func (c *regexProfanityChecker) Check(content string) bool {
	lowercased := strings.ToLower(content)
	return c.re.MatchString(lowercased)
}
