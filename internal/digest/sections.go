// SPDX-FileCopyrightText: 2026 Bob the Skull <bob.github@defp.uk>
// SPDX-License-Identifier: 0BSD

package digest

import (
	"strings"

	"git.home.flameeyes.family/bob/digest-builder/internal/config"
	"git.home.flameeyes.family/bob/digest-builder/internal/wallabag"
)

// Classify groups articles into sections based on their Wallabag tags.
// Returns a map of section key → articles in that section.
// Articles with no matching section tag fall into "general_tech".
func Classify(articles []wallabag.Article, sections []config.SectionConfig) map[string][]wallabag.Article {
	// Build a case-insensitive tag → section-key lookup.
	tagMap := make(map[string]string)
	for _, sec := range sections {
		for _, tag := range sec.Tags {
			tagMap[strings.ToLower(tag)] = sec.Key
		}
	}

	result := make(map[string][]wallabag.Article)
	for _, art := range articles {
		placed := false
		for _, tag := range art.Tags {
			if key, ok := tagMap[strings.ToLower(tag)]; ok {
				result[key] = append(result[key], art)
				placed = true
				break // first matching section wins
			}
		}
		if !placed {
			result["general_tech"] = append(result["general_tech"], art)
		}
	}
	return result
}
