# blog
My blog service.


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
- [ ] Markdown format support
- [ ] Telegram Instant View support
- [ ] Uploading media
- [ ] RSS
