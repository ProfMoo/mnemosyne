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

### Eventually

- AI-based music suggestions based on recent "obsessions" (i.e. fav songs of last week or two). Might just rely on integrations w/ Aurral or other service for this tho.
- Integrations with music downloaders to automatically download suggestions
