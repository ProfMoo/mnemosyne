# Mnemosyne

Your full music listening history, on your own hardware, forever.

## Purpose

Mnemosyne is meant to be a self-hosted, open-source, replacement for any music listening history (aka "scrobbling") solution. It's an alternative to [last.fm](https://www.last.fm/) (closed source), and [koito](https://github.com/gabehf/koito) and [maloja](https://github.com/krateng/maloja) (both open source).

The primary difference between prior scrobbling apps and Mnemosyne is:

1. **Eventual correctness.** Other scrobblers don't ensure correct data on a continual basis. An import as "Skrillex, Naisha" (one artist) is stuck forever. In Mnemosyne, the artist is split as "Skrillex" and "Naisha" during scrobbling. If it's tagged incorrectly during scrobbling, a background process continually checks that data and self-corrects.
2. **Self-hosted.** Mnemosyne is optimized for self-hosted setups. Other fully features listening platforms, like ListenBrainz Server, require full DBs, message queues, and data processing deployments to self-host. Mnemosyne only needs a filesystem for SQLite.

## Benchmarking

The reality of self-hosting music apps is that: MusicBrainz is the only good music metadata source (don't try to mix metadata w/ Spotify, Last.FM, etc.). Below I compare Mnemosyne against other solutions, particularly with respect to MusicBrainz tagging.

| Feature | [ListenBrainz Server](https://github.com/metabrainz/listenbrainz-server) | [Koito](https://github.com/gabehf/Koito) | [Maloja](https://github.com/krateng/maloja) | [Multi-Scrobbler](https://github.com/FoxxMD/multi-scrobbler) | [Meelo](https://github.com/Arthi-chaud/Meelo) |
| --- | --- | --- | --- | --- | --- |
| **Primary purpose** | 🟡 Full listening platform | ✅ Scrobble server + stats | ✅ Scrobble server + stats | 🟡 Scrobble collection + routing | 🟡 Music library + player |
| **Intended audience** | 🟡 Community / multi-user | ✅ Personal | ✅ Personal | ✅ Personal | ✅ Personal |
| **Self-hostable** | 🟡 Very complex | ✅ Easy | ✅ Easy | ✅ Easy | 🟡 Moderate |
| **MusicBrainz matching** | ✅ Advanced | 🟡 Moderate | ❌ Minimal | ✅ Good | ✅ Good |
| **Continuous MB rematching** | 🟡 Partial / complex | 🟡 Not verified | ❌ No | ❌ No | 🟡 Library rescanning |
| **Manual match correction** | ✅ Yes | 🟡 Limited | 🟡 Custom rules | ✅ Yes | ✅ Yes |
| **Listening analytics**¹ | ✅ Extensive | ✅ Good | ✅ Good | ❌ Basic | 🟡 Moderate |
| **MB-enriched reporting**² | ✅ Extensive | 🟡 Limited | ❌ Minimal | ❌ Minimal | 🟡 Moderate |

^1: More complex analysis and reports. For example, "your favorite 80s track", or "favorite albums that you heard for this first time this year"

^2: Includes reports that involve metadata not usually available directly in scrobbling DBs, such as "your top tracks from producer XYZ" or "favorite remixes"

## Goals

### Principles

- Mnemosyne will always be self-hosting and homelab friendly.
- Do NOT lose listening history. If it hits the Mnemosyne API, it will get ingested. Scrobbling is eventually consistent.
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
- More complex reports based on MusicBrainz metadata.

## Why Build This?

Honestly, there are good alternatives out there. These are some slight upgrades to be had in this space, outlined above. But primarily I've chosen to write this as a fun personal project 🙂

I have been writing code professionally for over a decade. I basically have no choice but to use AI to write all my code at work... and I don't see that changing anytime soon. This project is meant to be a way for me to still spend time on the nitty gritty code. I'll still use AI, especially for undifferentiated heavy lifting and boring tasks. But the design and the core functionality will be all grass fed, organize, hand-written code like the times of yor. I miss it.
