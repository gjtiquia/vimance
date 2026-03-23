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
- [templ](https://templ.guide/quick-start/installation)
- [air](https://github.com/air-verse/air?tab=readme-ov-file#installation)

### commands

```bash

# start dev server
air

# generate templ components
templ generate

# start server
go run .
```
