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
      Note: Current display blit entry point that the new retained renderer bridge now targets indirectly through `Draw(image, x, y)`
    - Path: renderer.go
      Note: Current keyed invalidation scheduler that still remains beneath the retained UI and JS runtime layers
    - Path: writer.go
      Note: Current writer ownership/pacing layer that remains beneath the JS runtime
    - Path: runtime/reactive/runtime.go
      Note: Pure-Go reactive core for signals, batching, and eager watches/effects
    - Path: runtime/ui/ui.go
      Note: Retained page/tile UI model with active-page tracking and dirty-tile collection
    - Path: runtime/render/visual_runtime.go
      Note: Tile-to-display renderer bridge using the existing Draw(image, x, y) boundary
    - Path: runtime/host/runtime.go
      Note: Host runtime shell for event routing, page hooks, timers, and replay entry points
    - Path: runtime/js/runtime.go
      Note: goja bootstrap and native-module registration for the first JS slice
    - Path: cmd/loupe-js-demo/main.go
      Note: First end-to-end JS page demo command that renders script-defined tiles to PNG files
    - Path: ttmp/2026/04/11/LOUPE-005--goja-javascript-api-for-dynamic-loupedeck-interfaces/design-doc/02-textbook-reactive-goja-ui-runtime-for-dynamic-loupedeck-interfaces.md
      Note: Intern-oriented conceptual deep dive for the preferred reactive design
    - Path: ttmp/2026/04/11/LOUPE-005--goja-javascript-api-for-dynamic-loupedeck-interfaces/design-doc/03-implementation-plan-reactive-goja-ui-runtime-for-dynamic-loupedeck-interfaces.md
      Note: Intern-oriented execution roadmap that guided the milestone-by-milestone implementation
    - Path: ttmp/2026/04/11/LOUPE-005--goja-javascript-api-for-dynamic-loupedeck-interfaces/design-doc/04-implementation-plan-converge-the-loupedeck-js-runtime-onto-go-go-goja-runtime-ownership.md
      Note: Next-phase plan for replacing the ad hoc JS execution model with go-go-goja runtime ownership patterns
ExternalSources: []
Summary: Ticket for the goja-based JavaScript runtime above the current Loupedeck Go rendering and transport layers, now including the design package plus a first implemented runtime stack: reactive core, retained UI, retained renderer bridge, host shell, goja modules, animation/easing, replay semantics, and a JS demo command.
LastUpdated: 2026-04-11T20:40:45-04:00
WhatFor: Track both the design work and the first implementation pass for an embedded JavaScript runtime that can build dynamic animated interfaces on the Loupedeck without exposing raw transport details.
WhenToUse: Use when evaluating how the goja scripting layer maps onto the current renderer/writer stack, onboarding a new engineer to the preferred reactive approach, or locating the concrete runtime packages that now implement the first slice.
---

# Goja JavaScript API for dynamic Loupedeck interfaces

## Overview

LOUPE-005 explores what a scriptable JavaScript runtime should look like if the root `github.com/go-go-golems/loupedeck` package grows from a transport-safe Go frontend into a programmable dynamic UI platform. The central question is not just “can Go expose functions to goja?” but “what is the right JavaScript abstraction layer above the existing display, renderer, and writer architecture?”

The ticket now focuses on both design and implementation:

- API shape options and rationale
- scene/state/animation models
- easing and timeline primitives
- concrete example scripts for multiple design styles
- milestone-by-milestone runtime implementation progress
- the first working JS runtime slice on top of the retained Go stack

## Key Links

- **Brainstorm design doc**: `design-doc/01-brainstorm-goja-javascript-api-approaches-for-dynamic-loupedeck-interfaces.md`
- **Reactive textbook**: `design-doc/02-textbook-reactive-goja-ui-runtime-for-dynamic-loupedeck-interfaces.md`
- **Implementation plan**: `design-doc/03-implementation-plan-reactive-goja-ui-runtime-for-dynamic-loupedeck-interfaces.md`
- **go-go-goja convergence plan**: `design-doc/04-implementation-plan-converge-the-loupedeck-js-runtime-onto-go-go-goja-runtime-ownership.md`
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
- Milestone A complete: pure-Go reactive core (`runtime/reactive`)
- Milestone B complete: retained page/tile UI model (`runtime/ui`)
- Milestone C complete: retained tile renderer bridge (`runtime/render`)
- Milestone D complete: host runtime shell (`runtime/host`)
- Milestone E complete: first goja modules (`loupedeck/state`, `loupedeck/ui`) and JS demo command
- Milestone F complete: animation/easing packages and JS modules (`loupedeck/anim`, `loupedeck/easing`)
- Milestone G complete: retained replay semantics for reconnect-safe redraws
- Convergence phase H started: local `runtimeowner` port adopted as the first owner-thread step; runtime-scoped bindings and callback refits remain next
- Diary actively maintained with per-milestone commits and validation evidence

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
