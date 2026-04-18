---
Title: Analysis and implementation guide for post-PR jsverbs CLI polish
Ticket: LOUPE-015
Status: active
Topics:
    - loupedeck
    - javascript
    - jsverbs
    - cli
    - documentation
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: cmd/loupedeck/cmds/run/command.go
      Note: |-
        Current run command still returns a structured status row instead of behaving like a plain runtime command
        run is simplified to a plain runtime command here
    - Path: cmd/loupedeck/cmds/run/session.go
      Note: |-
        Session defaults and schema assembly live here
        Runtime session defaults and help-visible flags live here
    - Path: cmd/loupedeck/cmds/verbs/command.go
      Note: |-
        Generated verbs currently rely on Glazed command wrappers and only show default-section help briefly
        Generated verbs are simplified to bare runtime commands here
    - Path: pkg/scriptmeta/scriptmeta.go
      Note: |-
        The per-file doc and verb lookup scope bugs both live here
        Per-file doc and verb scope fixes live here
ExternalSources:
    - https://github.com/go-go-golems/loupedeck/pull/1
Summary: Narrow design doc for the post-PR cleanup pass covering per-file scope correctness and simplification of loupedeck runtime command UX.
LastUpdated: 2026-04-18T19:50:00-04:00
WhatFor: Explains the intended behavior and implementation boundaries for the small follow-up fixes requested after the jsverbs CLI cutover.
WhenToUse: Read before modifying scriptmeta scope logic, run command construction, or the generated verbs Cobra help/output behavior.
---


# Analysis and implementation guide for post-PR jsverbs CLI polish

## Executive Summary

This ticket is not another large architecture pass. It is a cleanup ticket for correctness and product-shape polish.

Two of the requested changes are correctness fixes from PR comments:

1. per-file doc extraction must stay per-file
2. per-file explicit verb lookup must stay per-file

The rest are UX simplifications:

- make `run` a plain command instead of a structured-output command
- make the generated `verbs` commands feel like runtime commands rather than generic Glazed data commands
- surface loupedeck session flags directly in short help
- default all runtime commands to infinite duration unless the caller opts into a timeout

## Problem Statement

### Problem 1: `doc --script <file>` extracts too much

`pkg/scriptmeta.BuildDocStore(...)` currently walks `target.RootDir` even when the selected target is a single file. That means a request for docs for one script can pull in unrelated docs from sibling files.

This is wrong for the product semantics of `--script <file>` and makes per-script documentation unreliable.

### Problem 2: explicit verb selection is not properly scoped to the selected file

`pkg/scriptmeta.FindVerb(...)` currently does an early `registry.Verb(selector)` lookup against the entire scanned directory before applying the entry-file filter.

That means a command targeting one file can resolve a full-path verb from another file in the same scanned directory, which violates file-target expectations.

### Problem 3: `run` still behaves like a data command

The `run` command currently implements both `cmds.BareCommand` and `cmds.GlazeCommand`, and its long help still advertises `--with-glaze-output --output json` even though the command’s purpose is to run a live hardware/runtime session.

That is unnecessary product surface and makes the CLI feel more framework-driven than task-driven.

### Problem 4: generated verbs still inherit Glazed-structured UX

The generated `verbs` tree currently uses command wrappers that still implement Glazed or writer-command behavior. That keeps the parsing infrastructure convenient, but it also means the commands inherit structured-output framing and Glazed settings sections even though these are runtime-oriented commands rather than reporting/data commands.

### Problem 5: loupedeck session flags are not visible enough in short help

The session flags exist, but the short help currently emphasizes only the default section. For actual runtime usage, `--device`, `--duration`, `--send-interval`, and `--flush-interval` are first-class operational flags and should be visible immediately on generated verb help.

### Problem 6: default duration should be infinite

The current default duration of `15s` is surprising for a live scene runner. The desired default is `0s`, meaning run until interrupted or until the hardware exit path is used.

## Proposed Solution

### 1. Fix per-file doc extraction in `pkg/scriptmeta`

If `ResolveTarget(...)` returns an `EntryFile`, `BuildDocStore(...)` should build its input list from that single file only.

Directory targets should retain the current recursive behavior.

### 2. Fix per-file explicit verb lookup in `pkg/scriptmeta`

`FindVerb(...)` should perform explicit selector lookup only against `EntryVerbs(target, registry)` when `target.EntryFile` is set.

That keeps all selector modes aligned with the file scope:

- exact full path
- short verb name
- function name
- suffix match

### 3. Make `run` a plain command

`run` should stop implementing `cmds.GlazeCommand` and should stop returning a structured status row.

It should simply:

- parse its arguments/flags
- execute the live runtime session
- return success or error

### 4. Keep verbs parseable through schema metadata, but remove structured runtime UX

The generated verbs still need schema-driven argument parsing because jsverbs metadata already defines fields/arguments/sections. So the right move is not to throw away the schema; it is to stop exposing runtime commands as structured-output commands.

The implementation should therefore:

- keep using jsverbs command descriptions for argument/flag generation
- keep merging in loupedeck session fields
- wrap generated commands as bare runtime commands instead of Glazed result commands
- avoid dual-mode / `--with-glaze-output`
- avoid adding Glazed settings sections to the generated runtime commands

### 5. Show session flags in short help

All runtime commands should show both:

- the default/verb argument section
- the `loupedeck` session section

That should apply to:

- `run`
- generated `verbs ...`

### 6. Change default duration to `0s`

`NewSessionSection()` should use `0s` as the default and help text should state clearly that `0s` means run until interrupted.

## Design Decisions

### Decision: keep jsverbs schema parsing, drop structured-output product framing

This is the key nuance.

The user did not ask to remove schema-based parsing. They asked to stop treating runtime commands as structured-output commands.

So the intended design is:

- keep schema-driven flag generation
- drop Glazed reporting UX where it is not useful

### Decision: selector scope must be consistent for all explicit lookups on file targets

If the caller picked one file, every selector path should stay inside that file’s explicit verbs. No early fast path should bypass that filter.

### Decision: infinite duration is the correct operational default

These commands drive a live device session. Time-limited runs are an opt-in debugging/testing choice, not the default product behavior.

## Alternatives Considered

### Alternative 1: leave `run` and `verbs` as dual-mode commands and just hide some help

Rejected.

That would preserve framework complexity the user explicitly said they do not want.

### Alternative 2: eliminate schema sections from verbs entirely and hand-roll Cobra flags

Rejected.

That would create duplication against jsverbs metadata and reintroduce exactly the kind of ad hoc CLI glue the earlier work was supposed to avoid.

### Alternative 3: change selector scope only for implicit lookups

Rejected.

The PR comment is correct: explicit selector lookups also need file scoping for consistency and correctness.

## Implementation Plan

1. Create this ticket and record the requested scope.
2. Fix `pkg/scriptmeta.BuildDocStore(...)` so file targets use only the selected file.
3. Fix `pkg/scriptmeta.FindVerb(...)` so file targets restrict explicit selector resolution to `EntryVerbs(...)`.
4. Add tests for both scope fixes.
5. Simplify `cmd/loupedeck/cmds/run/command.go` so it is a bare runtime command only.
6. Refactor `cmd/loupedeck/cmds/run/session.go` helpers so runtime-only section assembly can be reused without Glazed settings leakage.
7. Refactor `cmd/loupedeck/cmds/verbs/command.go` so generated verbs are exposed as bare runtime commands with session flags visible in short help.
8. Change session default duration to `0s` and update tests/help/docs accordingly.
9. Run targeted tests and then `go test ./...`.
10. Update the implementation diary, changelog, tasks, and ticket index before committing.
