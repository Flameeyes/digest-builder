// SPDX-FileCopyrightText: 2026 Bob the Skull <bob.github@defp.uk>
// SPDX-License-Identifier: 0BSD

// Package digest builds timestamped EPUB digests from classified articles.
package digest

// CSS is the e-ink-optimised stylesheet embedded in every generated EPUB.
// Matches the style from the original digest.py pipeline.
const CSS = `
body {
    font-family: Georgia, 'Times New Roman', serif;
    line-height: 1.6;
    margin: 1em;
    color: #1a1a1a;
    background: white;
}
h1 { font-size: 1.5em; margin-bottom: 0.3em; }
h2 { font-size: 1.3em; margin-bottom: 0.3em; }
h3 { font-size: 1.1em; }
.meta {
    font-size: 0.85em;
    color: #555;
    margin-bottom: 1em;
    border-bottom: 1px solid #ccc;
    padding-bottom: 0.5em;
}
.meta a { color: #555; }
.article-content img { max-width: 100%; height: auto; }
a { color: #333; }
blockquote {
    border-left: 3px solid #ccc;
    margin-left: 0;
    padding-left: 1em;
    color: #444;
}
pre, code {
    font-family: 'Courier New', monospace;
    font-size: 0.9em;
    background: #f4f4f4;
    padding: 0.2em 0.4em;
}
pre { padding: 0.8em; overflow-x: auto; }
/* Force e-ink safe rendering — no dark backgrounds */
div, aside, section, figure, span, p {
    background: white !important;
    color: #1a1a1a !important;
}
`
