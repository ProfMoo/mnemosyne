# Mnemosyne

Your full music listening history, on your own hardware, forever.

## Purpose

Mnemosyne is meant to be a self-hosted, open-source, replacement for any music listening history (aka "scrobbling") solution. It's an alternative to [last.fm](https://www.last.fm/) on the closed source side, and [koito](https://github.com/gabehf/koito) and [maloja](https://github.com/krateng/maloja) on the open-source side.

The primary difference between prior scrobbling apps and Mnemosyne is correctness. Other scrobblers don't ensure correct data on a continual basis. An import as "Skrillex, Naisha" (one artist) is stuck forever. In Mnemosyne, the artist is split as "Skrillex" and "Naisha" during scrobbling. If it's tagged incorrectly during scrobbling, a background process continually checks that data and self-corrects.

## Goals

### Principles

- Mnemosyne will always be self-hosting and homelab friendly.
- Do NOT lose listening history. If it hits our API, it will get ingested.
- Correct tagging and data as #1 priority.

## Roadmap

### Primary

- Able to import from Spotify, Tidal. and ListenBrainz (more later, perhaps)
- Music is tagged from MusicBrainz, with corrections and updates running continually in the background.
- API-based, heavily spec'd. All API calls are defined via OpenAPI. So new frontends and integrations can be built in any language or framework.
- Built to be HIGHLY AVAILABLE (can run multiple at once, to ensure no scrobble is missed)
- Intelligent dedup support (some scrobbling systems double dip). I'll probably disallow scrobbles if they're the same song and are "played"
- Async jobs that provide interesting, useful insight to you about your listening history (think: "what is my top played in 2025 that was new to me during that year?)

### Eventually

- AI-based music suggestions based on recent "obsessions" (i.e. fav songs of last week or two). Might just rely on integrations w/ Aurral or other service for this tho.
- Integrations with music downloaders to automatically download suggestions based on listening history

## Why Build This?

Honestly, there are good alternatives out there. The reasons I'm building this are:

1. I found some features/behaviors of Koito to be underwhelming.
   1. The inability to properly tag music on the way in frustrated me. Tons of incorrect or incomplete data.
   2. Scrobbles were dropped here and there. Seemingly if a scrobble was dropped during import, Koito would just continue onward without issue.
   3. The Koito interface is honestly not very intuitive or insightful. I want more details info about my listening history, not just monthly breakdowns.
2. I have been writing code professionally for over a decade. I basically have no choice but to use AI to write all my code at work... and I don't see that changing anytime soon. This project is meant to be a way for me to still spend time on the nitty gritty code. I'll still use AI, especially for undifferentiated heavy lifting and boring tasks. But the design and the core functionality will be all grass fed, organize, hand-written code like the times of yor. I miss it.
