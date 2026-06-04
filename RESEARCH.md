# News RSS Tool — 研究與選型文件

> 目標：(1) 取得 hacker / AI / security 三大主題的最新 news RSS；(2) 決定閱讀方式（CLI 工具 vs 瀏覽器外掛 vs 其他）。
>
> 本文件為「選型決策紀錄」，尚未安裝任何工具。Feed URL 已於 2026-06 用 `curl` 驗證（結果見文末）。

---

## 一、關鍵發現：filter 不是瀏覽器外掛的弱點

研究中先釐清了使用者最在意的問題——**瀏覽器外掛能不能 filter？**

答案是：**可以，而且 Feedbro 的 filter 能力與 CLI 的 Newsboat 同級**：

- 支援 **關鍵字 / 正規表示式（regex）** 比對
- 支援 **all / any / none**（= AND / OR / NOT）布林邏輯
- 比對範圍可選 **標題、作者、內文**
- 符合條件後可自動 **隱藏 / 刪除 / 加標籤 / 高亮 / 桌面通知**
- 規則在文章進來時 **自動套用**

> 兩個取捨：① 外掛的過濾發生在「下載之後」（可搭配 hnrss 來源端 `?points=` 先粗篩）；② 規則綁定單一瀏覽器，換裝置要重設（除非接 Inoreader 雲端同步）。

因此決策重心從「能不能 filter」轉為 **介面型態、可腳本化程度、是否需要 AI 摘要/中文日報、跨裝置同步**。

---

## 二、方案比較表

| 維度 | Feedbro（瀏覽器外掛） | Newsboat（CLI） | 自架 Miniflux/FreshRSS | 自寫 Bun/TS 工具 |
|------|----------------------|-----------------|------------------------|------------------|
| **介面** | 圖文 GUI、看得到縮圖 | 純文字 TUI | Web UI（可手機看） | 自訂（CLI/Markdown/推播） |
| **Filter 能力** | ⭐⭐⭐⭐⭐ Rule Engine：keyword/regex/布林/作者/內文 | ⭐⭐⭐⭐⭐ `query` filter + `ignore-article` | ⭐⭐⭐⭐ 規則 + 全文搜尋 | ⭐⭐⭐⭐⭐ 完全自訂 |
| **過濾發生點** | 下載後（可搭配 hnrss `?points=` 粗篩） | 下載後 + 來源端 | 下載後 | 任意（可在抓取時就篩） |
| **離線閱讀** | 部分（已抓的可看） | ✅ 全離線 | ✅（自架伺服器） | 視實作 |
| **跨裝置同步** | ❌ 綁瀏覽器（除非接 Inoreader） | ⚠️ 可接 TT-RSS/Nextcloud | ✅ 內建 | 視實作 |
| **可腳本化/自動化** | ❌ 難 | ✅ 可接外部指令 | ✅ 有 API | ⭐⭐⭐⭐⭐ 最強 |
| **AI 摘要 / 中文日報** | ❌ | ❌（需外掛腳本） | ❌ | ✅ 可直接整合 Claude API |
| **設定門檻** | 最低（裝外掛、貼 feed） | 中（編設定檔） | 中高（Docker） | 高（要開發維護） |
| **維護成本** | 低 | 低 | 中（要顧伺服器） | 高 |
| **最適合** | 想快速、要圖文、單機使用 | 重度 terminal 使用者、要離線 | 多裝置、要 Web/手機 | 想加 AI 摘要、自動推播、整進 workflow |

### 三方案一句話總結
- **Feedbro**：零開發、filter 夠強、有圖文 → 最快上手，但綁單一瀏覽器、難自動化。
- **Newsboat**：純終端機、可離線、可接腳本 → 符合 CLI 重度使用習慣（與本機環境最契合）。
- **自寫 Bun/TS**：唯一能做「AI 摘要 + 自動產繁中每日新聞日報 + 推播」的路線 → 開發量最大但最貼合 Bun 技術棧。

---

## 三、推薦 Feed 清單（可直接複製）

### Hacker News — 用 hnrss.org（可在來源端就 filter）
| 用途 | URL |
|------|-----|
| 前頁 | `https://hnrss.org/frontpage` |
| 高分過濾（≥100 分） | `https://hnrss.org/frontpage?points=100` |
| 關鍵字搜尋 | `https://hnrss.org/newest?q=rust+OR+ai` |
| 高留言討論 | `https://hnrss.org/ask?comments=25` |

> 參數可用 `&` 組合；任何端點可加 `.atom` / `.jsonfeed`；可加 `count=N`、`description=0`。

### AI / LLM
| 來源 | URL | 狀態 |
|------|-----|------|
| OpenAI News | `https://openai.com/news/rss.xml` | ✅ 200 |
| Simon Willison | `https://simonwillison.net/atom/everything/` | ✅ 200 |
| TLDR AI | `https://tldr.tech/api/rss/ai` | ✅ 200 |
| Import AI | `https://importai.substack.com/feed` | ✅ 200 |
| MIT Tech Review | `https://www.technologyreview.com/feed/` | （未測，常見有效） |
| ByteByteGo | `https://blog.bytebytego.com/feed` | （未測，常見有效） |
| **Anthropic News** | **無官方 RSS**（`/rss.xml` 等路徑皆 404） | ❌ 見備註 |

> **Anthropic 備註**：官網無公開 RSS。替代方案：① 用 RSS 橋接服務（如自架 RSSHub 的 `/anthropic/news`，公開 demo 站常回 403）；② 直接訂閱 OpenAI feed 已能涵蓋多數 LLM 新聞；③ 用自寫工具直接抓 `anthropic.com/news` 頁面解析。

### 資安 Security
| 來源 | URL | 狀態 |
|------|-----|------|
| Krebs on Security | `https://krebsonsecurity.com/feed/` | ✅ 200 |
| The Hacker News | `https://feeds.feedburner.com/TheHackersNews` | ✅ 200 |
| Bruce Schneier | `https://www.schneier.com/feed/atom/` | ✅ 200 |
| CISA 公告 | `https://www.cisa.gov/cybersecurity-advisories/all.xml` | ✅ 200 |
| BleepingComputer | `https://www.bleepingcomputer.com/feed/` | ⚠️ 403（見備註） |

### 一般科技（選配）
| 來源 | URL | 狀態 |
|------|-----|------|
| Ars Technica | `https://feeds.arstechnica.com/arstechnica/index` | ✅ 200 |
| The Verge | `https://www.theverge.com/rss/index.xml` | （未測，常見有效） |
| TechCrunch | `https://techcrunch.com/feed/` | （未測，常見有效） |
| The Pragmatic Engineer | `https://newsletter.pragmaticengineer.com/feed` | （未測，常見有效） |

> **403 備註**：BleepingComputer（及 RSSHub demo）對雲端/`curl` IP 回 403，屬反爬蟲，**feed 本身存在**；實際用瀏覽器外掛或 Newsboat（帶正常 User-Agent）通常可正常讀取。

---

## 四、推薦（依使用者情境）

使用者環境特徵：重度 terminal、偏好 Bun、輸出習慣繁體中文、想要一個可重複使用的工具（後定名為 **claude-news-plugin**）而非單純訂閱。

- **主推：自寫 Bun/TS 工具** — 若想要 AI 摘要 / 繁中每日日報 / 自動化。唯一能整合 Claude API 產出繁中摘要的路線，最貼合技術棧。
- **次選：Newsboat** — 若只想單純讀、零維護。與本機 CLI 環境最契合、零開發。
- Feedbro 雖最快上手，但「綁瀏覽器、難自動化」與 CLI/腳本傾向不合，僅在「想要圖文、單機快速看」時推薦。

---

## 五、驗證方式（落地後如何確認可行）

1. **Feed 有效性**：`curl -sI -A "Mozilla/5.0" <url>` 確認 200；`curl -s <url> | head` 看是否為合法 RSS/Atom XML。
2. **Newsboat 路線**：`brew install newsboat` → 寫 `~/.newsboat/urls` → `newsboat -r` 確認抓到三主題 → 加一條 filter 規則驗證過濾生效。
3. **Feedbro 路線**：裝外掛 → 匯入 OPML → 建一條 Rule（標題含關鍵字則高亮）→ 確認自動套用。
4. **自寫 Bun/TS 路線**：`bun init` → 抓取+解析一個 feed → 輸出標題清單 → 加 filter →（選配）接 Claude API 產繁中摘要。

---

## 六、Feed URL 驗證紀錄（2026-06，`curl -L -A "Mozilla/5.0"`）

```
404  https://www.anthropic.com/rss.xml          （及 /news/rss.xml、/feed.xml、/index.xml 皆 404 → 無官方 RSS）
200  https://openai.com/news/rss.xml
200  https://feeds.feedburner.com/TheHackersNews
200  https://www.cisa.gov/cybersecurity-advisories/all.xml
200  https://krebsonsecurity.com/feed/
200  https://simonwillison.net/atom/everything/
200  https://tldr.tech/api/rss/ai
200  https://importai.substack.com/feed
200  https://www.schneier.com/feed/atom/
200  https://feeds.arstechnica.com/arstechnica/index
200  https://hnrss.org/frontpage
403  https://www.bleepingcomputer.com/feed/     （反爬蟲，feed 仍存在）
403  https://rsshub.app/anthropic/news          （demo 站擋雲端 IP）
```

---

## 待決事項

最終要落地哪一個方案（Newsboat / Feedbro / 自寫 Bun/TS），待使用者拍板後，再進入對應的安裝或開發步驟。
