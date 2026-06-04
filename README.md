# claude-news-plugin

A Claude Code skill that fetches the latest **hacker / AI / security / tech** news
from RSS feeds and summarises it in **Traditional Chinese**.

A fast Go binary does the heavy lifting (concurrent fetch + RSS/Atom/JSON parsing);
Claude does the classification and summarisation. Invoke it with `/news`.

```
/news                # all topics, last 24h
/news ai             # AI/LLM only
/news security 12h   # security, last 12h
/news all 72h        # everything, last 72h
```

Output is a Traditional Chinese digest grouped by topic, each item a one-line
summary with a link and source.

---

## How it works

```
/news ──► run.sh ──► news-fetch (Go) ──► feeds.json
                          │ concurrent fetch (goroutines), gofeed parsing
                          ▼
                     clean JSON to stdout
                          │
                          ▼
                Claude classifies + summarises in 繁中
```

- **`news-fetch`** (Go + [gofeed](https://github.com/mmcdole/gofeed)) — fetches all
  feeds concurrently with a browser User-Agent, filters by `-category` and `-since`,
  strips HTML, and prints compact JSON. A failed feed is logged to stderr and skipped,
  never aborting the run.
- **`run.sh`** — maps friendly args (`ai`, `12h`) to flags. Self-healing: if the
  binary is missing it builds it from bundled `src/` on first run (needs Go).
- **`SKILL.md`** — feeds the JSON into Claude with the Traditional Chinese output spec.

---

## Install

### Option A — as a plugin (recommended for sharing)

Requires [Go](https://go.dev/dl/) on the target machine (the binary is built locally,
so it works on any OS/arch).

```
/plugin marketplace add mfeo/claude-news-plugin
/plugin install news
```

On first `/news` the wrapper auto-builds the binary from the bundled source.

### Option B — manual (user-level skill)

```bash
mkdir -p ~/.claude/skills/news
cp -r skill/* ~/.claude/skills/news/      # SKILL.md, run.sh, feeds.json
cp -r src     ~/.claude/skills/news/      # Go source for self-build
# optional: prebuild now instead of on first run
cd ~/.claude/skills/news/src && go build -o ../bin/news-fetch .
```

Then `/news` is available in any session.

---

## Feeds

Edit `feeds.json` to add/remove sources. Each entry:

```json
{ "category": "ai", "source": "OpenAI News", "url": "https://openai.com/news/rss.xml" }
```

`category` is one of `hacker`, `ai`, `security`, `tech` (add your own freely — pass it
to `/news <category>`). Default feeds (all verified working as of 2026-06):

| Category | Sources |
|----------|---------|
| hacker   | Hacker News front page, Hacker News 100+ points (via [hnrss.org](https://hnrss.github.io/)) |
| ai       | OpenAI News, Simon Willison, TLDR AI, Import AI |
| security | Krebs on Security, The Hacker News, Bruce Schneier, CISA Advisories, BleepingComputer |
| tech     | Ars Technica, The Verge, TechCrunch, Pragmatic Engineer |

> **Note on BleepingComputer**: its feed returns `403` to data-center IPs (anti-bot).
> It usually works fine from a normal home/office network. A 403 is logged as a
> `warn:` and does not affect other feeds.

> **Note on Anthropic**: anthropic.com has no public RSS feed. OpenAI's feed plus
> Simon Willison's blog already cover most LLM news; use an RSS bridge (RSSHub) if
> you specifically need Anthropic posts.

See [`RESEARCH.md`](./RESEARCH.md) for the full feed-vs-reader research and the
URL verification log.

---

## Direct CLI use (without Claude)

The Go binary is a standalone tool:

```bash
cd src && go build -o ../bin/news-fetch .
./bin/news-fetch -feeds ./feeds.json -category ai -since 24h -max 25
```

Flags: `-feeds`, `-category` (comma-separated or `all`), `-since` (Go duration),
`-max` (per feed), `-timeout` (per feed).

---

## Layout

```
claude-news-plugin/
├── README.md            # this file
├── RESEARCH.md          # feed sources + reader comparison + verification log
├── feeds.json           # feed list (4 topics, verified)
├── .claude-plugin/      # marketplace manifest (must be at repo root)
│   └── marketplace.json #   → points at ./plugin
├── src/                 # Go source (gofeed concurrent fetcher)
│   ├── go.mod
│   └── main.go
├── skill/               # the skill, ready to drop into ~/.claude/skills/news
│   ├── SKILL.md
│   ├── run.sh
│   └── feeds.json
└── plugin/              # the plugin itself (installed by marketplace)
    ├── .claude-plugin/plugin.json
    ├── hooks/hooks.json # PostInstall: pre-build the binary (best-effort)
    └── skills/news/     # bundled skill + src for self-build
```

## License

MIT
