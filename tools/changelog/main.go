package main

import (
	"bufio"
	"os"
	"strings"
)

const (
	// conventional commit
	feat     = "feat:"
	chore    = "chore:"
	docs     = "docs:"
	fix      = "fix:"
	test     = "test:"
	refactor = "refactor:"

	// no category (no conventional commit followed)
	minor = "minor"

	// output md sections
	features        = "🚀 New Features"
	bugs            = "🐛 Bug Fixes"
	maintenance     = "🧰 Maintenance"
	minorChanges    = "🪜 Minor changes"
	newContributors = "👋 New Contributors"

	// gh changelog sections
	newContributorsText = "## New Contributors"
	whatsChanged        = "## What's Changed"
)

var (
	tags            = []string{feat, chore, docs, fix, test}
	orderedSections = []string{features, bugs, maintenance, minorChanges, newContributors}
	labels          = map[string]string{
		feat:     features,
		fix:      bugs,
		docs:     maintenance,
		chore:    maintenance,
		test:     maintenance,
		refactor: maintenance,
		minor:    minorChanges,
	}
)

func main() {
	changeLog, err := os.Open("input.md")
	if err != nil {
		panic(err)
	}

	defer changeLog.Close() //nolint

	output, err := os.Create("changelog.md")
	if err != nil {
		panic(err)
	}

	defer output.Close() //nolint

	fileScanner := bufio.NewScanner(changeLog)
	fileScanner.Split(bufio.ScanLines)

	components := map[string][]string{}
	eoPRs := false // end of PRs section
	for fileScanner.Scan() {
		text := fileScanner.Text()

		if text == whatsChanged {
			_, _ = output.WriteString(whatsChanged + "\n")
			continue
		}

		if text == newContributorsText {
			eoPRs = true
			continue
		}

		if eoPRs {
			components[newContributors] = append(components[newContributors], text)
			continue
		}

		found := false
		for _, tag := range tags {
			if strings.Contains(text, tag) {
				components[labels[tag]] = append(components[labels[tag]], text)
				found = true
				break
			}
		}

		if !found {
			components[labels[minor]] = append(components[labels[minor]], text)
		}
	}

	for _, section := range orderedSections {
		if _, ok := components[section]; ok {
			_, _ = output.WriteString("## " + section + "\n")

			for _, line := range components[section] {
				_, _ = output.WriteString(line + "\n")
			}

			_, _ = output.WriteString("\n")
		}
	}
}
