> another rewrite! go to https://github.com/gjtiquia/stuf

---

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
- show keyboard shortcut hints consistently, and architecture makes sure logic and render wont drift (probably refactor to use keymap in bubbletea?)
- develop some concrete user journies (meaning... actually start using it and solve for pain points!)

## high-level questions that users should answer
- how long to save up to x amount of money?
- how much can i save per month?
- how much can i invest per month, without being broke, while still being able to travel?
- how much do i need to allocate/envelop into different categories each month? (credit card, debit card, saving goals, investments)
- how can i evaluate (ideally daily) if my spending is within the envelop amount i assigned myself? (see envelop / zero-based budgeting)
- how can i evaluate the performance of my investments? (given that i can also add records of investments, eg. portfolio overall value, individual trades)
- how can i see the overall trend of my finances? is it growing?

## how it answers those questions

- **aggregation**: query results show total/income/expense and breakdown by tag. press `v` to toggle trend view (monthly/weekly/daily/yearly).
- **targets**: link a saved query to a target amount. the targets view shows actual vs planned, with `???` for untracked categories (progressive disclosure).
- **sign convention**: positive amounts = income, negative = expense. `SUM(amount_cents)` answers everything.

## keybinds

### record creation
- `tab` / `shift+tab` — next/prev field
- `enter` — advance field / confirm
- `esc` — back

### query results
- `j/k` — move cursor
- `n/p` — page down/up
- `g/G` — top/bottom
- `enter` — edit record
- `s` — save query
- `v` — toggle trend view (cycles: monthly → weekly → daily → yearly → list)
- `esc` — back

### targets
- `j/k` — move cursor
- `a` — add target
- `d` — delete target
- `esc` — back to menu

## development

if you are an agent, please first read AGENTS.md
