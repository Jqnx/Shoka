# Metadata sources

Shoka fills an entry by running a library's metadata sources in order and
merging the results: the first source to set a field wins, and later sources
only fill gaps. Source links are the exception — they are collected from every
source and deduplicated. Sources are configured per library under
**Admin → Libraries → [library] → Metadata sources**, where each can be enabled,
reordered, and given its own settings.

Local sources always run before remote ones; only the order among remote
sources is configurable. ComicInfo.xml and Filename are enabled for new
libraries. The two gallery sources start disabled.

| Source | Kind | Enabled by default | Needs | Provides |
| --- | --- | --- | --- | --- |
| ComicInfo.xml | Local | Yes | — | Title, summary, language, category, page count, release date, artists, tags, characters, parodies, circles, web links |
| Filename | Local | Yes | — | Title, artist, circle, parodies, language |
| nHentai | Remote | No | Optional API key | Title, artists, circles, parodies, characters, tags, category, language, page count, upload date, gallery URL |
| E-Hentai | Remote | No | Session cookies | Nothing yet — lookup is not implemented |

## ComicInfo.xml

Reads the `ComicInfo.xml` entry embedded in an archive. This is the
highest-priority source because the file was curated for that archive, so its
values outrank anything scraped.

| ComicInfo element | Fills |
| --- | --- |
| `Title`, `Summary` | Title, summary |
| `LanguageISO` | Language |
| `Genre` | Category |
| `PageCount` | Page count |
| `Year` / `Month` / `Day` | Release date |
| `Writer` | Artists (comma-separated) |
| `Tags` | Tags (comma-separated) |
| `Characters` | Characters (comma-separated) |
| `Series` | Parodies (comma-separated) |
| `Publisher` | Circles (comma-separated) |
| `Web` | Source links (space-separated URLs) |

The source takes no configuration. Saving metadata from the UI writes a
`ComicInfo.xml` back into the archive; for RAR and 7z archives that operation
rewrites the container as `.cbz` (see the warning in the README's Quick start).

## Filename

A local heuristic parser for archives without embedded metadata. It recognises
shapes such as:

```
[Artist] Title (Parody)
[Circle (Artist)] Title (Parody)
[Artist] Title (Parody) [English]
```

A leading `[123456]` gallery id, event tags (`(C91)`, `(Comiket 102)`), and
resolution suffixes (`(x1200)`) are stripped first. The first bracket names the
artist unless it uses the `[Circle (Artist)]` form, which fills both fields; the
first parenthesis that is not blocklisted becomes the parody; a bracket naming a
language (e.g. `[English]`) sets the language. Any other bracket is discarded —
a plain `[Circle]` is not read as a circle — and the remaining text becomes the
title.

This source exposes two blocklists:

- **Magazine blocklist** — case-insensitive prefixes ignored when choosing a
  parody. A built-in list of magazine names (Comic LO, Kairakuten, …) is always
  applied.
- **Misc tag blocklist** — case-insensitive substrings ignored when choosing a
  parody. Defaults include `zenpen`, `kouhen`, `saimin`, and `hypnotized`.

## nHentai

Uses nhentai's v2 gallery API. Automatic fetching searches by title and takes
the first result; picking a search result yourself fetches that gallery by id.

- **API key (optional).** Create one at
  <https://nhentai.net/user/settings#apikeys> and paste it into the source's
  settings. Without a key the anonymous quota applies.
- **Rate limits.** Shoka paces requests at roughly one every six seconds and, on
  a `429`, waits for `Retry-After` (60 seconds by default) and retries once.
  Requests time out after 15 seconds.
- **Fields.** Title prefers the English name, then the romanised form, then the
  Japanese one. Tags are split into artists, groups (circles), parodies,
  characters, language, category, and tags; the gallery's URL, page count, and
  upload date are stored too.

## E-Hentai

Registered and configurable — it has a **session cookie** field and an
on/off switch — but its search and gallery lookup are not implemented yet.
Enabling it currently contributes no metadata. The intent is to authenticate
with the library's e-hentai/exhentai cookies once the API client is finished.

## Adding a source

Sources live in `backend/internal/metadata/sources` and implement
`metadata.Source` (at minimum `Name`, `Priority`, `IsLocal`, `Fetch`).
`SearchableSource` adds manual search, `RemoteSource` adds fetch-by-id, and
`HostAware` maps gallery URLs back to the source. Register the implementation in
`backend/main.go`; its built-in priority decides where it runs before a library
reorders it, and local sources are always pinned ahead of remote ones.
