# 0001. Make the repository public

Date: 2026-09-21
Status: Accepted

## Context
Should I make this repository public or private?
I thought If I make it public someone will try to hack it or do something by looking at the source code. But I guess making private doesn't make it safer.
>Hiding the code doesn't make it much safer. Your blog will almost certainly be built on public frameworks and libraries (Next.js, Django, Hugo, whatever you pick), and attackers already know those better than your code. Attacks on small sites are rarely someone reading your repo. They're automated bots that scan the whole internet for known weaknesses: outdated packages, exposed admin panels, default passwords. Those bots hit private and public projects equally. If your code is secure, publishing it doesn't make it insecure. If it isn't, keeping it private only hides the problem a little.

Ok, so real issue is leaking secrets and old libs.

## Decision
I chose to make it public.

## Alternatives considered
Making this repo public makes no sence.

## Consequences
Now I can add this project to my portfolio.
