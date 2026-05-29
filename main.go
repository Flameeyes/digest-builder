// SPDX-FileCopyrightText: 2026 Bob the Skull <bob.github@defp.uk>
// SPDX-License-Identifier: 0BSD

// digest-builder fetches articles from Wallabag tagged "digest-pending",
// bundles them into a timestamped EPUB optimised for e-ink readers, and
// marks them consumed.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"git.home.flameeyes.family/bob/digest-builder/internal/config"
	"git.home.flameeyes.family/bob/digest-builder/internal/digest"
	"git.home.flameeyes.family/bob/digest-builder/internal/wallabag"
)

var version = "dev"

func main() {
	configPath := flag.String("config", "", "Path to config.yaml (optional; uses env vars if omitted)")
	dryRun := flag.Bool("dry-run", false, "Fetch and build EPUB but do not mark articles consumed")
	showVersion := flag.Bool("version", false, "Print version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Println("digest-builder", version)
		os.Exit(0)
	}

	log.SetFlags(log.LstdFlags | log.Lmsgprefix)
	log.SetPrefix("[digest-builder] ")

	// ── Load configuration ───────────────────────────────────────────
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	// ── Timestamp for this run ───────────────────────────────────────
	now := time.Now()
	dateStr := now.Format("2006-01-02-1504")
	log.Printf("=== Digest Builder — %s ===", dateStr)

	// ── Initialise Wallabag client ───────────────────────────────────
	wb, err := wallabag.NewClient(cfg)
	if err != nil {
		log.Fatalf("wallabag: %v", err)
	}

	// ── Fetch pending articles ───────────────────────────────────────
	items, err := wb.FetchPending()
	if err != nil {
		log.Fatalf("fetch: %v", err)
	}
	if len(items) == 0 {
		log.Println("No pending articles — nothing to build")
		os.Exit(0)
	}
	log.Printf("Fetched %d pending articles", len(items))

	// ── Group into sections ──────────────────────────────────────────
	sections := digest.Classify(items, cfg.Digest.Sections)
	for _, sec := range cfg.Digest.Sections {
		if arts, ok := sections[sec.Key]; ok {
			log.Printf("  %s: %d articles", sec.Label, len(arts))
		}
	}

	// ── Build EPUB ───────────────────────────────────────────────────
	outDir := cfg.Output.Dir
	if outDir == "" {
		outDir = "."
	}
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		log.Fatalf("output dir: %v", err)
	}

	filename := fmt.Sprintf(cfg.Output.FilenameFormat(), dateStr)
	outPath := outDir + "/" + filename

	if err := digest.BuildEPUB(sections, cfg.Digest.Sections, dateStr, outPath); err != nil {
		log.Fatalf("epub: %v", err)
	}

	fi, _ := os.Stat(outPath)
	log.Printf("EPUB written: %s (%d bytes)", outPath, fi.Size())

	// ── Mark consumed ────────────────────────────────────────────────
	if *dryRun {
		log.Println("Dry run — skipping mark-consumed step")
	} else {
		consumed := wb.MarkConsumed(items, dateStr)
		log.Printf("Marked %d/%d articles consumed", consumed, len(items))
	}

	log.Println("Done!")
}
