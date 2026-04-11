
> i feel like the project right now, is in a state where it is so focused on "productive keybinds", that there is nothing but keybinds right now
> will try experimenting building from another direction, rather that starting with vim + tables then add finance, we first do the finance then add vim tables later
> first have a workflow that actually works first, then optimize the workflow with vim keys

---

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

The asset pipeline order and commands live in **`package.json`**: `build:tsc` → `build:tailwind` → `build:bundle` → `build:templ` → `build:wasm`. **`bun run build:assets`** runs that chain in order (single source of truth). Air’s `pre_cmd` calls `build:assets` before each `go build`.

```bash
# install JS/CSS tooling (tailwind, typescript, etc.)
bun install

# dev: watch + rebuild (runs build:assets, then go build, then the binary)
# browser live reload uses Air’s proxy at :3500; the app listens on :3000
bun run dev
# or just run air directly
air

# production-style: build once and run (no watcher, no proxy)
bun run start

# optional: only assets + codegen (no go build)
bun run build:assets

# optional: run one asset step while debugging (same order as build:assets)
bun run build:tsc
bun run build:tailwind
bun run build:bundle
bun run build:templ
bun run build:wasm

# optional: full build (assets + go), does not start the server
bun run build
```

## notes
- consider hydrating on server before passing down
- cuz also needa consider network calls, i dun think its wise to network call from with go wasm...?
    - tho of course... can always rpc to network call js side then back... hm...
- also gotta fix "ciw" can edit header row
    - should have some fundamental architectural thing that guarantees header rows CANNOT be edited!
