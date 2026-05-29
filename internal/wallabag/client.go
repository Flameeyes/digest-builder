// SPDX-FileCopyrightText: 2026 Bob the Skull <bob.github@defp.uk>
// SPDX-License-Identifier: 0BSD

// Package wallabag wraps the wallabago library with convenience methods
// for the digest-builder workflow: fetch pending articles, classify them,
// and mark them consumed after EPUB generation.
package wallabag

import (
	"fmt"
	"log"
	"strings"

	"github.com/Strubbl/wallabago/v8"

	"git.home.flameeyes.family/bob/digest-builder/internal/config"
)

// Article is a simplified view of a Wallabag entry for digest processing.
type Article struct {
	ID          int
	Title       string
	URL         string
	DomainName  string
	Content     string
	Tags        []string
	ReadingTime int
	CreatedAt   string
}

// Client wraps wallabago global state with our config.
type Client struct {
	cfg *config.Config
}

// NewClient initialises the wallabago global config and returns a Client.
func NewClient(cfg *config.Config) (*Client, error) {
	wbCfg := wallabago.NewWallabagConfig(
		strings.TrimRight(cfg.Wallabag.URL, "/"),
		cfg.Wallabag.ClientID,
		cfg.Wallabag.ClientSecret,
		cfg.Wallabag.Username,
		cfg.Wallabag.Password,
	)
	wallabago.SetConfig(wbCfg)

	return &Client{cfg: cfg}, nil
}

// FetchPending retrieves all entries tagged with the configured pending tag.
func (c *Client) FetchPending() ([]Article, error) {
	tag := c.cfg.Digest.PendingTag
	if tag == "" {
		tag = "digest-pending"
	}

	var articles []Article
	page := 1

	for {
		entries, err := wallabago.GetEntries(
			wallabago.APICall, // built-in authenticated HTTP caller
			-1,                // archive: don't filter
			-1,                // starred: don't filter
			"created",         // sort field
			"asc",             // order
			page,              // page number
			50,                // perPage
			tag,               // filter by tag
			0,                 // since (unix timestamp, 0 = all)
			-1,                // public: don't filter
			"full",            // detail level
			"",                // domain_name: don't filter
		)
		if err != nil {
			return nil, fmt.Errorf("GetEntries page %d: %w", page, err)
		}

		for _, item := range entries.Embedded.Items {
			if c.shouldExclude(item.Title) {
				log.Printf("  Excluding: %s", item.Title)
				continue
			}

			tags := make([]string, 0, len(item.Tags))
			for _, t := range item.Tags {
				tags = append(tags, t.Label)
			}

			createdAt := ""
			if item.CreatedAt != nil {
				createdAt = item.CreatedAt.Format("2006-01-02 15:04")
			}

			articles = append(articles, Article{
				ID:          item.ID,
				Title:       item.Title,
				URL:         item.URL,
				DomainName:  item.DomainName,
				Content:     item.Content,
				Tags:        tags,
				ReadingTime: item.ReadingTime,
				CreatedAt:   createdAt,
			})
		}

		if page >= entries.Pages || len(entries.Embedded.Items) == 0 {
			break
		}
		page++
	}

	return articles, nil
}

// shouldExclude checks if an article title matches any exclusion pattern.
func (c *Client) shouldExclude(title string) bool {
	lower := strings.ToLower(title)
	for _, pat := range c.cfg.Digest.ExcludePatterns {
		if strings.Contains(lower, strings.ToLower(pat)) {
			return true
		}
	}
	return false
}

// MarkConsumed removes the pending tag and adds a date-stamped tag to each
// article. Returns the number of successfully updated entries.
func (c *Client) MarkConsumed(articles []Article, dateStr string) int {
	tag := c.cfg.Digest.PendingTag
	if tag == "" {
		tag = "digest-pending"
	}
	consumedTag := fmt.Sprintf("digest-%s", dateStr)

	count := 0
	for _, art := range articles {
		// Fetch current tags to find the ID of the pending tag.
		tags, err := wallabago.GetTagsOfEntry(wallabago.APICall, art.ID)
		if err != nil {
			log.Printf("  GetTagsOfEntry(%d): %v", art.ID, err)
			continue
		}

		tagID := -1
		for _, t := range tags {
			if strings.EqualFold(t.Label, tag) {
				tagID = t.ID
				break
			}
		}

		// Remove the pending tag.
		if tagID >= 0 {
			if err := wallabago.DeleteEntryTag(art.ID, tagID); err != nil {
				log.Printf("  DeleteEntryTag(%d, %d): %v", art.ID, tagID, err)
				// Continue — still try to add the consumed tag.
			}
		}

		// Add the date-stamped consumed tag.
		if err := wallabago.AddEntryTags(art.ID, consumedTag); err != nil {
			log.Printf("  AddEntryTags(%d, %s): %v", art.ID, consumedTag, err)
			continue
		}

		count++
	}
	return count
}
