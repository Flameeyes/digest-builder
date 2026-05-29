// SPDX-FileCopyrightText: 2026 Bob the Skull <bob.github@defp.uk>
// SPDX-License-Identifier: 0BSD

package digest

import (
	"fmt"
	"html"
	"os"
	"path/filepath"
	"strings"

	epub "github.com/go-shiori/go-epub"

	"git.home.flameeyes.family/bob/digest-builder/internal/config"
	"git.home.flameeyes.family/bob/digest-builder/internal/wallabag"
)

// BuildEPUB generates an EPUB file from classified articles.
func BuildEPUB(
	sections map[string][]wallabag.Article,
	sectionOrder []config.SectionConfig,
	dateStr string,
	outPath string,
) error {
	e, err := epub.NewEpub(fmt.Sprintf("Daily Digest — %s", dateStr))
	if err != nil {
		return fmt.Errorf("new epub: %w", err)
	}

	e.SetAuthor("digest-builder")
	e.SetLang("en")
	e.SetIdentifier(fmt.Sprintf("digest-%s", dateStr))

	// Write CSS to a temp file for go-epub to pick up.
	cssPath, err := writeCSSTempFile()
	if err != nil {
		return fmt.Errorf("write css: %w", err)
	}
	defer os.Remove(cssPath)

	internalCSS, err := e.AddCSS(cssPath, "digest.css")
	if err != nil {
		return fmt.Errorf("add css: %w", err)
	}

	// Collect ordered section keys, then append any extras from the data.
	ordered := make([]string, 0, len(sectionOrder))
	labelMap := make(map[string]string)
	for _, sec := range sectionOrder {
		ordered = append(ordered, sec.Key)
		labelMap[sec.Key] = sec.Label
	}
	for key := range sections {
		found := false
		for _, k := range ordered {
			if k == key {
				found = true
				break
			}
		}
		if !found {
			ordered = append(ordered, key)
		}
	}

	chapterIdx := 0

	for _, sectionKey := range ordered {
		articles, ok := sections[sectionKey]
		if !ok || len(articles) == 0 {
			continue
		}

		label := labelMap[sectionKey]
		if label == "" {
			label = strings.ReplaceAll(sectionKey, "_", " ")
			// Simple title case fallback.
			words := strings.Fields(label)
			for i, w := range words {
				if len(w) > 0 {
					words[i] = strings.ToUpper(w[:1]) + w[1:]
				}
			}
			label = strings.Join(words, " ")
		}

		// Section heading page.
		sectionBody := fmt.Sprintf(
			`<h1>%s</h1><p>%d article(s)</p>`,
			html.EscapeString(label), len(articles),
		)
		sectionFile, err := e.AddSection(sectionBody, label, "", internalCSS)
		if err != nil {
			return fmt.Errorf("add section %s: %w", sectionKey, err)
		}

		for _, art := range articles {
			chapterIdx++

			title := art.Title
			if title == "" {
				title = "Untitled"
			}

			source := art.DomainName
			if source == "" {
				source = "Unknown"
			}

			body := fmt.Sprintf(
				`<h1>%s</h1>
<div class="meta">
<strong>%s</strong> — %s<br/>
<a href="%s">%s</a>
</div>
<div class="article-content">
%s
</div>`,
				html.EscapeString(title),
				html.EscapeString(source),
				html.EscapeString(art.CreatedAt),
				html.EscapeString(art.URL),
				html.EscapeString(art.URL),
				art.Content, // Already HTML from Wallabag.
			)

			fname := fmt.Sprintf("ch%03d.xhtml", chapterIdx)
			_, err := e.AddSubSection(sectionFile, body, title, fname, internalCSS)
			if err != nil {
				return fmt.Errorf("add article %d (%s): %w", chapterIdx, title, err)
			}
		}
	}

	if err := e.Write(outPath); err != nil {
		return fmt.Errorf("write epub: %w", err)
	}

	return nil
}

// writeCSSTempFile writes the digest CSS to a temporary file and returns
// its path. The caller is responsible for removing it.
func writeCSSTempFile() (string, error) {
	dir := os.TempDir()
	path := filepath.Join(dir, "digest-builder-style.css")
	if err := os.WriteFile(path, []byte(CSS), 0o644); err != nil {
		return "", err
	}
	return path, nil
}
