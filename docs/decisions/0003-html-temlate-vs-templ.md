# 0003. html/template vs. github.com/a-h/templ

Date: 2026-09-23
Status: Accepted

## Context
What template engine should I chose for rendering html?
html/template or templ?

html/template (standard library) parses template text at runtime.
templ (github.com/a-h/templ) is a templating language that compiles to Go code.
type safety on templ side, because in templ parameters are typed and errors appear at compile time.

## Decision
I chose templ.

## Alternatives considered
Rejected html/template, because it's reflection based, not generated.
Typo might go through, but in templ it's build error.

## Consequences

