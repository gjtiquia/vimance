# vimance

a vim-centric finance workflow

optimized for productivity, efficiency and effectiveness

meticulously designed, from each individual keybinding to the overarching workflow

heavily opinionated, a personal tool shared with the world

## ramblings

### the idea

vim + google sheets + actual budget (zero-based budgeting) all-in-one

make it fast to update finances, so that finances will be updated and not lag behind

tho vim-centric, there should also be a strong emphasis on the mobile experience

most financial transactions occur away from the keyboard, while most financial analysis occurs near the keyboard

### the tech stack

go - cuz why not, and also im currently learning it, and also want to leave the option open for TUI-client

htmx - cuz why not

## development

### prerequisites

- [go](https://go.dev/doc/install)
- [tinygo](https://tinygo.org/getting-started/install/)
- [templ](https://templ.guide/quick-start/installation)
- [air](https://github.com/air-verse/air?tab=readme-ov-file#installation)
- [bun](https://bun.sh/)

### commands

The asset pipeline (TypeScript, Tailwind, bundling, `templ generate`, TinyGo wasm) lives in **`bun run build:assets`** — a single source of truth for order and commands. Air’s `pre_cmd` calls that script before each `go build`.

```bash
# install JS/CSS tooling (tailwind, typescript, etc.)
bun install

# dev: watch + rebuild (runs build:assets, then go build, then the binary)
# browser live reload uses Air’s proxy at :3500; the app listens on :3000
bun run dev

# production-style: build once and run (no watcher, no proxy)
bun run start

# optional: only assets + codegen (no go build)
bun run build:assets

# optional: full build (assets + go), does not start the server
bun run build
```
