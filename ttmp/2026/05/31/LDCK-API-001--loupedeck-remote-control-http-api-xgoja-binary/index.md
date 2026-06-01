---
Title: Loupedeck Remote Control HTTP API - xgoja Binary
Ticket: LDCK-API-001
Status: active
Topics:
    - loupedeck
    - xgoja
    - http-api
    - agent-integration
DocType: index
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: "Design for a standalone xgoja binary (loupedeck-server) that exposes the Loupedeck Live hardware as a REST HTTP API. Combines loupedeck, express (gojahttp), and database (SQLite) providers. Supports polling and webhook event delivery for remote coding agent integration."
LastUpdated: 2026-05-31T12:17:21.313155232-04:00
WhatFor: "Control a Loupedeck device remotely from a coding agent via HTTP — show status, receive button/knob/touch events, set button colors and brightness."
WhenToUse: "When you need to bridge a remote coding agent to a physically connected Loupedeck device."
---

# Loupedeck Remote Control HTTP API - xgoja Binary

## Overview

This ticket designs and plans the implementation of `loupedeck-server`, a standalone xgoja-generated binary that combines the Loupedeck hardware driver, an express HTTP server (gojahttp), and a SQLite database to expose the Loupedeck Live as a REST API. The API supports:

- **Display control**: Create pages/tiles, draw grids, switch pages
- **Button control**: Set LED colors, brightness
- **Event polling**: `GET /api/v1/events?since=N` drains queued hardware events from SQLite
- **Webhooks**: Register callback URLs, receive POSTed events with HMAC signatures

**Current status**: Design document complete and uploaded to reMarkable. Implementation not yet started.

## Key Links

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field

## Status

Current status: **active**

## Topics

- loupedeck
- xgoja
- http-api
- agent-integration

## Tasks

See [tasks.md](./tasks.md) for the current task list.

## Changelog

See [changelog.md](./changelog.md) for recent changes and decisions.

## Structure

- design/ - Architecture and design documents
- reference/ - Prompt packs, API contracts, context summaries
- playbooks/ - Command sequences and test procedures
- scripts/ - Temporary code and tooling
- various/ - Working notes and research
- archive/ - Deprecated or reference-only artifacts
