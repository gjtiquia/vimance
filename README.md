> rewrite in progress...
>
> current direction would be hyper-focusing on the workflow first, making things work, then further optimizing later
>
> and currently i think the way to get super close to the data, without any overhead about frontend UI design, would be simple strings and CLI
>
> but at the same time... i cant accept not having quality-of-life stuff like, at the very least minimal keybinds and some sort of suggestion
>
> hence... will experiment with bubbletea! had a look and seems will be a good fit, while being productive quickly

---

# vimance

## commands

```bash
go run .          # run app
go test ./...     # run tests
sqlc generate     # regenerate db code
```

## ideas
- tag templates? commonly used group of tags together (or perhaps a whole record template itself)

## todos
- UX on adding links. we should do "queries" within create record flow, 
  basically users can use saved queries or build a temp query to find the record they want to link to
- show keyboard shortcut hints consistently, and architecture makes sure logic and render wont drift (probably refactor to use keymap in bubbletea?)
- queries should hv a way to... check trends (+/- how much per month), viewing sums...?
- develop some concrete user journies (meaning... actually start using it and solve for pain points!)

## development

if you are an agent, please first read AGENTS.md
