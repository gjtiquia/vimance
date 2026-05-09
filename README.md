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
- links between records (implemented) for top-down drill-down: balance snapshots → transaction breakdowns. multiple parents supported.

## todos
- menus no need type to search, is only useful if there are A LOT of menu items. 
  but if within 10, can navigate via numbers and arrows, vim-style by default, no insert mode.
  this makes things much simpler too on the code-side
- input fields should change appearance based on whether it is focused or not.
  specifically for create record, the date should only "explode" to three fields when either one of the three fields are focused, and collapse into a single date when other fields are focused.
  same with tags / links, should only show the prompt to type when it is focused, and simplify it when its not focused.
  simplified in a way that is similar to the confirmation page. and on simplify, perhaps also do validation and show validation errors / warnings.
  should use the same "focus" language as the menu selection. > carets/cursors basically
  and since its the same language, main menu should also support tab / shift tab as keybinds.
- UX on adding links. we should do "queries" within create record flow, 
  basically users can use saved queries or build a temp query to find the record they want to link to

## development

if you are an agent, please first read AGENTS.md
