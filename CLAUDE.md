# Instructions for Claude

At the start of every conversation in this repo (wowsims_classic_vanillaplus),
read these two files first, before doing anything else:

- [docs/CHANGES.md](docs/CHANGES.md) — plain-language, file-by-file record of
  what's changed in prior sessions and why. Read this so you don't re-derive
  context that's already been written down, and so you don't redo work (or
  undo an intentional decision) that a past session already made.
- [docs/TODO.md](docs/TODO.md) — open items raised across past sessions that
  haven't been resolved yet. Check this before starting new work in case it's
  already related to something outstanding.

When the user asks about items, check VPlusItemDB (the item database used by
the VPlus pipeline) for new/updated items. When the user asks about icons or
drop sources for items, check Atlas/AtlasLoot.

When you make a nontrivial change, append to `docs/CHANGES.md` (new dated
`## Part <letter>` section, following the existing format) instead of starting
a separate file. If something comes up that's worth doing later but isn't part
of the current task, add it to `docs/TODO.md` rather than letting it drop.

Whenever you add or change something that affects the sim (a talent, buff, item effect,
stat or spell modifier), make sure the sidebar stats also show it. Check the spec's
`displayStats` and `modifyDisplayStats` in `ui/<spec>/sim.ts`: effects that are applied
per spell or per proc (hit, crit, damage modifiers) are not in the sim's stat totals, so
they need adding to `modifyDisplayStats` by hand. Verify in the browser that the sidebar
number moves when the talent or buff changes, and note any deliberate exception in
`docs/TODO.md`.
