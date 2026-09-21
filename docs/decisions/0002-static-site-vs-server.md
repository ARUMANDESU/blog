# 0002. Static site vs. Server

Date: 2026-09-21
Status: Approved

## Context
Should it be Static site using [github pages](https://docs.github.com/en/pages/getting-started-with-github-pages/creating-a-github-pages-site) or Server written in golang.

Static site is a web-site with pre-build files (HTML, CSS, JS) and displays the same content to every visitor.
In order to update it's content you have to update repo, thus new post it a new HTML file in the repo.

Server can show different interface for admin and visitors, which is good.
I can update posts very easily, but I should implement that logic into backend server.

## Decision
I chose to write my own server using golang, which gives me more controll over blog service and opportunity to add new features in the future.

## Alternatives considered
Static site is very easy to make but I need more features that Static site is not able to support.

## Consequences
Easier to customize and add features.
