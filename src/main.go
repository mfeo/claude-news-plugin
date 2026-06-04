// news-fetch concurrently fetches RSS/Atom/JSON feeds defined in feeds.json,
// filters by category and time window, and prints clean JSON to stdout for an
// LLM to classify and summarise. Per-feed failures are logged to stderr and do
// not abort the run.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/mmcdole/gofeed"
)

// feedSpec is one entry in feeds.json.
type feedSpec struct {
	Category string `json:"category"`
	Source   string `json:"source"`
	URL      string `json:"url"`
}

type feedsFile struct {
	Feeds []feedSpec `json:"feeds"`
}

// item is one normalised article emitted to stdout.
type item struct {
	Category  string `json:"category"`
	Source    string `json:"source"`
	Title     string `json:"title"`
	Link      string `json:"link"`
	Published string `json:"published"`
	Summary   string `json:"summary"`
}

// browser-like UA so feeds with basic anti-bot rules (e.g. BleepingComputer)
// don't 403.
const userAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) " +
	"AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0 Safari/537.36"

func main() {
	var (
		feedsPath  = flag.String("feeds", "", "path to feeds.json (default: alongside the binary)")
		categories = flag.String("category", "all", "comma-separated: hacker,ai,security,tech,all")
		since      = flag.Duration("since", 24*time.Hour, "only items newer than this window (e.g. 24h, 12h)")
		max        = flag.Int("max", 30, "max items per feed")
		timeout    = flag.Duration("timeout", 20*time.Second, "per-feed fetch timeout")
	)
	flag.Parse()

	path, err := resolveFeedsPath(*feedsPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	specs, err := loadFeeds(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error: load feeds:", err)
		os.Exit(1)
	}

	wanted := parseCategories(*categories)
	cutoff := time.Now().Add(-*since)

	items := fetchAll(filterSpecs(specs, wanted), *max, *timeout, cutoff)

	// Newest first across all sources.
	sort.Slice(items, func(i, j int) bool { return items[i].Published > items[j].Published })

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(items); err != nil {
		fmt.Fprintln(os.Stderr, "error: encode:", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "fetched %d items from %d feeds (window=%s)\n", len(items), len(filterSpecs(specs, wanted)), since.String())
}

// resolveFeedsPath returns the explicit path, or feeds.json next to the binary.
func resolveFeedsPath(explicit string) (string, error) {
	if explicit != "" {
		return explicit, nil
	}
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Join(filepath.Dir(exe), "feeds.json"), nil
}

func loadFeeds(path string) ([]feedSpec, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var f feedsFile
	if err := json.Unmarshal(data, &f); err != nil {
		return nil, err
	}
	return f.Feeds, nil
}

func parseCategories(s string) map[string]bool {
	set := map[string]bool{}
	for _, c := range strings.Split(s, ",") {
		c = strings.TrimSpace(strings.ToLower(c))
		if c != "" {
			set[c] = true
		}
	}
	return set
}

func filterSpecs(specs []feedSpec, wanted map[string]bool) []feedSpec {
	if wanted["all"] || len(wanted) == 0 {
		return specs
	}
	var out []feedSpec
	for _, s := range specs {
		if wanted[strings.ToLower(s.Category)] {
			out = append(out, s)
		}
	}
	return out
}

// fetchAll fetches every spec concurrently and returns all items within cutoff.
func fetchAll(specs []feedSpec, max int, timeout time.Duration, cutoff time.Time) []item {
	var (
		wg sync.WaitGroup
		mu sync.Mutex
		// non-nil so an empty result encodes as [] rather than null
		out = []item{}
	)
	for _, spec := range specs {
		wg.Add(1)
		go func(spec feedSpec) {
			defer wg.Done()
			got, err := fetchOne(spec, max, timeout, cutoff)
			if err != nil {
				fmt.Fprintf(os.Stderr, "warn: %s (%s): %v\n", spec.Source, spec.URL, err)
				return
			}
			mu.Lock()
			out = append(out, got...)
			mu.Unlock()
		}(spec)
	}
	wg.Wait()
	return out
}

func fetchOne(spec feedSpec, max int, timeout time.Duration, cutoff time.Time) ([]item, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	fp := gofeed.NewParser()
	fp.UserAgent = userAgent

	feed, err := fp.ParseURLWithContext(spec.URL, ctx)
	if err != nil {
		return nil, err
	}

	var out []item
	for _, e := range feed.Items {
		if len(out) >= max {
			break
		}
		published := publishedTime(e)
		if !published.IsZero() && published.Before(cutoff) {
			continue
		}
		out = append(out, item{
			Category:  spec.Category,
			Source:    spec.Source,
			Title:     strings.TrimSpace(e.Title),
			Link:      e.Link,
			Published: formatTime(published),
			Summary:   cleanSummary(e),
		})
	}
	return out, nil
}

func publishedTime(e *gofeed.Item) time.Time {
	if e.PublishedParsed != nil {
		return *e.PublishedParsed
	}
	if e.UpdatedParsed != nil {
		return *e.UpdatedParsed
	}
	return time.Time{}
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}

// cleanSummary trims and caps the description so the JSON stays compact; the LLM
// does the real summarising.
func cleanSummary(e *gofeed.Item) string {
	s := e.Description
	if s == "" {
		s = e.Content
	}
	s = stripTags(s)
	s = strings.Join(strings.Fields(s), " ")
	const cap = 500
	if len(s) > cap {
		s = s[:cap] + "…"
	}
	return s
}

// stripTags removes naive HTML tags without pulling in a dependency.
func stripTags(s string) string {
	var b strings.Builder
	depth := 0
	for _, r := range s {
		switch r {
		case '<':
			depth++
		case '>':
			if depth > 0 {
				depth--
			}
		default:
			if depth == 0 {
				b.WriteRune(r)
			}
		}
	}
	return b.String()
}
