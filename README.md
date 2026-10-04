# blog
My blog service. Just a CRUD project, nothing interesting here.

## Stack
- Go (v1.27.1)
- SQLite ([golib](https://gitlab.com/cznic/sqlite))
- Garage S3

## ENV
Check out [.env.example](./.env.example)

## Rules
- LLM usage must be as little as possible and never use it for writing code, but for example, generating commit msg is ok.
- Follow [COMMIT.md](./docs/COMMIT.md) for making commits
- If you have debatable choice then write [decision file](##Decision)


## Decisions
Here -> [dir](./docs/decisions/)

#### Naming
Use this naming template: `0001-<name>.md`
> [!WARNING] 
> Never renumber them, because that number will be used as reference.

#### Superseding
Don't edit old decisions when you change your mind. Instead create new decision and set old one's status to `Superseded by 0007` and add `-spsd` into it's file name: `0001-<name>-spsd.md`

## Features
- [x] Markdown format support
- [x] Uploading media
- [ ] Telegram Instant View support
- [ ] AUTH for admin (just a middleware and login handler with session not jwt)
- [ ] RSS

## Deliberately done or not:
- No gracefull shutdown: not needed for one user - me
- Observability: I don't need much for this project
- Not enforcing media ownership: post to media is one-to-many rel, and there is only one admin = we good
- Converting markdown into html in tx: I am aware that I can convert outside tx then update post in tx, which would shorten the lock, but I am lazy 
- Lack of some tests: yes, I am lazy and this is not some enterprise service, but pet project done with limited llm usage so I don't forget some skills.
