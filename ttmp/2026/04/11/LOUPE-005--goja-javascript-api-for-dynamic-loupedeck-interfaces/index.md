---
Title: Goja JavaScript API for dynamic Loupedeck interfaces
Ticket: LOUPE-005
Status: active
Topics:
    - loupedeck
    - go
    - goja
    - javascript
    - animation
    - rendering
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: display.go
      Note: Current display blit entry point the JS layer must eventually target indirectly
    - Path: renderer.go
      Note: Current keyed invalidation scheduler that a future JS UI runtime should build on top of
    - Path: writer.go
      Note: Current writer ownership/pacing layer that must remain beneath any JS API
    - Path: cmd/loupe-svg-buttons/main.go
      Note: Current dynamic-ish button bank demo that motivates a scriptable higher-level runtime
    - Path: svg_icons.go
      Note: Existing asset pipeline that a JS API may want to expose as icons/images instead of raw SVG strings
    - Path: ttmp/2026/04/11/LOUPE-005--goja-javascript-api-for-dynamic-loupedeck-interfaces/design-doc/02-textbook-reactive-goja-ui-runtime-for-dynamic-loupedeck-interfaces.md
      Note: Intern-oriented conceptual deep dive for the preferred reactive design
    - Path: ttmp/2026/04/11/LOUPE-005--goja-javascript-api-for-dynamic-loupedeck-interfaces/design-doc/03-implementation-plan-reactive-goja-ui-runtime-for-dynamic-loupedeck-interfaces.md
      Note: Intern-oriented execution roadmap for implementing the preferred reactive runtime in phases
ExternalSources: []
Summary: Design ticket for adding a goja-based JavaScript API above the current Loupedeck Go rendering and transport layers, now including a broad brainstorm, example scripts, a reactive-runtime textbook, and a detailed phased implementation plan for a new engineer.
LastUpdated: 2026-04-11T20:40:45-04:00
WhatFor: Track the design work for a future embedded JavaScript runtime that can build dynamic animated interfaces on the Loupedeck without exposing raw transport details.
WhenToUse: Use when evaluating how a goja scripting layer should map onto the current renderer/writer stack, onboarding a new engineer to the preferred reactive approach, or finding example JS APIs and implementation phases.
---

# Goja JavaScript API for dynamic Loupedeck interfaces

## Overview

LOUPE-005 explores what a scriptable JavaScript runtime should look like if the root `github.com/go-go-golems/loupedeck` package grows from a transport-safe Go frontend into a programmable dynamic UI platform. The central question is not just “can Go expose functions to goja?” but “what is the right JavaScript abstraction layer above the existing display, renderer, and writer architecture?”

The ticket focuses on:

- API shape options
- scene/state/animation models
- easing and timeline primitives
- different user personas and scenarios
- concrete example scripts for multiple design styles

## Key Links

- **Brainstorm design doc**: `design-doc/01-brainstorm-goja-javascript-api-approaches-for-dynamic-loupedeck-interfaces.md`
- **Reactive textbook**: `design-doc/02-textbook-reactive-goja-ui-runtime-for-dynamic-loupedeck-interfaces.md`
- **Implementation plan**: `design-doc/03-implementation-plan-reactive-goja-ui-runtime-for-dynamic-loupedeck-interfaces.md`
- **Example scripts**: `reference/01-javascript-api-example-scripts.md`
- **Diary**: `reference/02-implementation-diary.md`

## Status

Current status: **active**

Current completion state:
- Ticket created
- Broad brainstorm design document written
- Example-script reference written
- Reactive textbook written
- Detailed phased implementation plan written
- Diary started
- Implementation intentionally not started yet

## Topics

- loupedeck
- go
- goja
- javascript
- animation
- rendering

## Tasks

See [tasks.md](./tasks.md) for the current task list.

## Changelog

See [changelog.md](./changelog.md) for recent changes and decisions.
