# Private-server spell tooltip rules

How spell tooltip *values* (damage/duration text, not names or icons) are corrected
to match the private server instead of retail Classic Wowhead. Companion to
[private-server-item-rules.md](private-server-item-rules.md), which covers items.

_Last updated: 2026-09-13._

## Background

The item pipeline already pulls stats from the private server's own DBC dump
instead of Wowhead. Spells did not get the same treatment until now: spell
*mechanics* (damage/coefficients/cooldowns) were always hand-coded in the Go sim
(`sim/<class>/*.go`), never derived from any data file, and spell *tooltip text*
(the popup shown when hovering a spell in the UI) was pulled live from Wowhead
with no private-server override at all.

This pipeline fixes the second part: the displayed tooltip *description* and
*channel-cast-time* text now come from the server's own DBC dump
(`CSV's/Spell.csv`, `CSV's/SpellDuration.csv`) when available, instead of
showing retail Wowhead's numbers. It does **not** touch spell names, icons, mana
cost, range, or requirements (still Wowhead's), and it does **not** touch the
Go sim's actual combat math — see "What this does NOT fix" below.

## Pipeline (run order)

```
python tools/gen_spell_value_overrides.py       # corrects assets/db_inputs/wowhead_spell_tooltips.csv
go run ./tools/database/gen_db -outDir=./assets -gen=db
cp -r assets/database/* dist/classic/assets/database/
GOOS=js GOARCH=wasm go build -o ./dist/classic/lib.wasm ./sim/wasm/
cp dist/classic/lib.wasm dist/classic/assets/lib.wasm
npx vite build -m development                    # only needed if ui/**/*.tsx changed
```

`tools/gen_custom_spell_tooltips.py` (older, pre-existing) still separately owns
**talent** spells: it fills in tooltips for talent spell IDs Wowhead has no data
for at all, and patches per-rank `descriptions` arrays into
`ui/core/talents/trees/*.json`. `gen_spell_value_overrides.py` explicitly excludes
talent IDs and handles everything else the sim references.

### How spell IDs are discovered

`gen_spell_value_overrides.py` unions spell IDs from:

- Every literal `SpellID: N` in `sim/**/*.go` (`core.ActionID{SpellID: ...}`).
- Every integer inside a `var FooSpellId = [N]int32{0, id1, id2, ...}` array
  literal in `sim/**/*.go` — **this is the majority of ranked spells**
  (e.g. `sim/priest/mind_flay.go`'s `MindFlaySpellId`), since they're referenced
  dynamically (`MindFlaySpellId[rank]`) rather than as a literal `SpellID:` value.
  A naive `SpellID:\s*\d+` grep alone misses almost every ranked ability.
- `database.SharedSpellsIcons` (`tools/database/overrides.go`).
- All talent spell IDs (unioned in, then subtracted back out — talents stay
  owned by the older script, see above).

### What gets corrected, and how

For each discovered spell ID that's already cached from Wowhead:

1. Resolve the description from `Spell.csv`'s `EffectBasePoints`/`EffectAmplitude`/
   `ProcChance`/`StackAmount`/`DurationIndex` columns (reverse-engineered; see the
   `COL_*` constants in `tools/gen_custom_spell_tooltips.py`), substituting
   Blizzard's `$s`/`$o`/`$d`/`$h`/`$u` tokens.
2. If it differs from the cached Wowhead description, rewrite **only** the
   `<div class="q">...</div>` block inside the cached tooltip HTML.
3. If the raw DBC text ties a `$d` (duration) token to the same effect, and the
   tooltip also has a `Channeled (N sec cast)` line, correct that line's `N` to
   match the same duration — Wowhead scrapes it as a separate field, so it can
   (and did, for Mind Flay) go stale independently of the description.
4. A spell ID with no Wowhead entry at all gets a minimal DBC-only row, same
   fallback pattern the talent script already uses.

Corrections are skipped (never applied) whenever the resolved text still
contains an unresolved `$` token or a bare `X` substitution — either means the
column reverse-engineering doesn't cover that spell/effect shape, and applying
it would silently corrupt an otherwise-correct tooltip. Also skipped: `$o`
(amount-over-duration) corrections on mana-regen text specifically (validated
against SW:Pain/Renew-style periodic damage/heal effects only — spot-checked
wrong for periodic-energize auras like Drink).

## Why regenerating db.json/db.bin isn't enough — the wasm and Rotation-tab gaps

Two separate things had to be fixed beyond the CSV correction itself, both
discovered by end-to-end verification in a real browser (not just re-running
the pipeline):

1. **`db.bin` is compiled into the WASM binary.** `assets/database/loader.go`
   `go:embed`s `db.bin` at build time. Editing `db.json`/`db.bin` on disk has no
   effect on the running app until `lib.wasm` is rebuilt (`GOOS=js GOARCH=wasm
   go build -o ./dist/classic/lib.wasm ./sim/wasm/`) and re-copied to
   `dist/classic/assets/lib.wasm`.
2. **The Rotation tab's spell picker never used local tooltip data at all.**
   `ActionId.trySetLocalTooltip()` already existed (added for the item
   pipeline) and is used by the gear picker (`ui/core/components/gear_picker/item_list.tsx`)
   to prefer local DB tooltip text over the live Wowhead widget when local data
   exists. The APL rotation picker
   (`ui/core/components/individual_sim_ui/apl_helpers.tsx`) never called it —
   it only called `setBackgroundAndHref` + `setWowheadDataset`, which always
   defers to the real, live `wow.zamimg.com/js/tooltips.js` widget fetching
   from `nether.wowhead.com` directly. That means **every** spell hover tooltip
   in the Rotation tab was silently showing live retail Wowhead data, no matter
   what this repo's local DB contained — this is why the first attempt at this
   fix appeared to do nothing when checked in a real browser. Fixed by wiring
   `trySetLocalTooltip()` into `apl_helpers.tsx` with the same
   local-first-then-Wowhead-fallback pattern the item picker uses.

If a future audit adds a new local-tooltip data source and a UI location still
shows stale/live Wowhead data despite the pipeline being correct, check whether
that location actually calls `trySetLocalTooltip()` before assuming the data
pipeline is broken.

## What this does NOT fix

- **Spell names and icons** — deliberately left as Wowhead's, per this being
  scoped to values/tooltip text only.
- **Mana cost, range, requirements** — untouched (only the description div and
  the `Channeled (N sec cast)` line are corrected).
- **The Go sim's actual combat math, except Mind Flay, Mind Blast, Smite, and
  Devouring Plague (see below)** —
  `sim/<class>/*.go` hardcodes damage, coefficients, tick counts, etc. as Go
  literals, independent of any tooltip. A tooltip correction never changes
  simulated numbers on its own. A one-off audit (2026-09-13, not committed as
  code) cross-checked a handful of DoT/channeled spells against `Spell.csv`:
  - Shadow Word: Pain, Moonfire — Go values match the DBC exactly (high confidence).
  - **Mind Flay — FIXED 2026-09-13** (`sim/priest/mind_flay.go`). The private
    server extends Mind Flay to a 5-tick/5-sec channel (retail Classic is
    3-tick/3-sec); confirmed via `Spell.csv`'s `EffectAmplitude`=1000ms,
    `DurationIndex`=7 (5000ms), consistent across all 6 ranks — high
    confidence. `MindFlayTicks` changed 3→5, `MindFlayBaseDamage` changed from
    `{0, 75, 126, 186, 261, 330, 426}` to `{0, 160, 275, 400, 550, 700, 900}`
    (= `(EffectBasePoints+1) * 5`, matching every rank's DBC-derived tooltip
    total exactly), and the `tickIdx == 0` branch's hardcoded `ticks = 3`
    changed to reference the `MindFlayTicks` const instead of duplicating the
    literal. Verified end-to-end in a real sim run (not just the tooltip) —
    Results tab shows Mind Flay ticking with the new damage/tick-count.
    `go test ./sim/priest/...` produced a byte-identical `.results` (the
    default P1 shadow preset doesn't spec Mind Flay, so there was no golden
    to promote).
  - **Mind Blast — FIXED 2026-09-14** (`sim/priest/mind_blast.go`). Found and
    decoded the `EffectDieSides` column (`Spell.csv` index 64) — the min-max
    damage range is `EffectBasePoints[1]+1` to `EffectBasePoints[1]+
    EffectDieSides[0]` (note the cross-slot pairing: the active base-points
    value sits in effect slot 1, but its paired die-sides value sits in slot
    0 — confirmed real, not a bug in the reverse-engineering, by cross-
    checking against Shadow Word: Pain/Devouring Plague/Starshards, which all
    have both values in the same slot 0 and zero variance). Rank 9's
    resulting range (557-589) was independently confirmed against the user's
    own in-game tooltip on the private server before any Go code changed.
    All 9 ranks corrected; old values were under-tuned, growing worse at
    higher ranks (e.g. rank 9 was 508-537, i.e. ~9% low at the low end and
    ~9% low at the high end). `go test ./sim/priest/...` DPS changed from
    239.272 to 246.667 on the default P1 shadow preset — golden `.results`
    promoted (this preset does use Mind Blast, unlike Mind Flay above).
  - **Smite — FIXED 2026-09-14** (`sim/priest/smite.go`), same
    `EffectBasePoints[1]+EffectDieSides[0]` pattern as Mind Blast (identical
    effect-slot structure across all 8 ranks). Applied the same fix, but
    **not independently spot-checked against an in-game tooltip** the way
    Mind Blast rank 9 was — confidence rests on the method being validated
    elsewhere, not on direct confirmation for this specific spell. Verify
    in-game if in doubt.
  - **Devouring Plague — FIXED 2026-09-14** (`sim/priest/devouring_plague.go`).
    Pure periodic DoT, `EffectDieSides[0]` is 1 for every rank (no variance,
    same slot as `EffectBasePoints[0]` — the reliable same-slot pattern, not
    Mind Blast/Smite's cross-slot case), so `(EffectBasePoints+1) * 8 ticks`
    applies directly with no ambiguity, same as Shadow Word: Pain/Starshards.
    All 6 ranks corrected; old values were a consistent ~10-12% too high
    (e.g. rank 6 was 904, corrected to 816). `go test ./sim/priest/...` DPS
    changed from 246.667 to 246.061 on the P1 shadow preset (small - Devouring
    Plague is a minor part of that rotation); golden `.results` promoted.
  - **Holy Fire — checked, NOT fixed.** Has both a direct-damage effect and a
    DoT effect, and unlike Mind Blast/Smite the effect-slot layout is NOT
    consistent across ranks (rank 3 puts the direct-damage value in a
    different slot than ranks 1, 2, 4-8). Rank 1's DoT recomputed cleanly to
    match Go's existing value (30) using an offset pairing
    (`EffectDieSides[i-1]` for `EffectBasePoints[i]`), but rank 2 did not (50
    vs Go's 40) — the pairing rule doesn't generalize. Needs an in-game
    tooltip check (like Mind Blast got) before touching.
  - **Rend** (`sim/warrior/rend.go`) — `EffectDieSides` is 1 for every rank
    (no random variance), which doesn't change the original finding: Go
    per-tick damage is consistently ~25-30% below the DBC-implied value
    across all 4 ranks (medium confidence). **NOT fixed** — flagged for
    follow-up, not yet touched.
  - **Flame Shock** — inconsistent audit result, likely a column-mapping
    issue in the audit itself rather than a real Go bug (low confidence).
    **NOT fixed** — not pursued.
  - **Every other spell in the sim** (the ~330 non-talent spell IDs this
    pipeline touches for tooltip text, and the ~4600+ talent/rotation spells
    overall) has **NOT** been audited against `Spell.csv` for Go-mechanics
    correctness at all. A tooltip now showing a DBC-corrected value does
    **not** imply the Go code computing that spell's actual damage was
    checked or changed. Treat any spell not listed above as fixed as
    unverified against the private server's dump until it's explicitly
    audited.
  - `EffectDieSides` (`Spell.csv` column index 64) is now decoded and gives a
    real min-max range for direct-damage spells, but the effect-slot pairing
    between it and `EffectBasePoints` is confirmed to vary per spell (same
    slot for Rend/SW:Pain/Devouring Plague/Starshards; cross-slot 1↔0 for
    Mind Blast/Smite; inconsistent across ranks for Holy Fire) — do not
    assume a single universal pairing rule without checking each spell's own
    data, and prefer an in-game tooltip confirmation over the formula alone
    when the two disagree.

## Reproducing / extending this

To re-run after `CSV's/Spell.csv` or the sim's spell IDs change:

```
python tools/gen_spell_value_overrides.py --dry-run   # preview counts first
python tools/gen_spell_value_overrides.py
go run ./tools/database/gen_db -outDir=./assets -gen=db
cp -r assets/database/* dist/classic/assets/database/
GOOS=js GOARCH=wasm go build -o ./dist/classic/lib.wasm ./sim/wasm/
cp dist/classic/lib.wasm dist/classic/assets/lib.wasm
go build ./...          # sanity check
go test ./sim/...        # check for unexpected result diffs (tooltip changes
                          # should never change these - if they do, investigate)
```

Then verify end-to-end in an actual browser (not just by inspecting
`db.json`) — hover the spell in the Rotation tab and confirm both the
description text and the `Channeled (N sec cast)` line, since a pipeline-level
"looks correct" check on `db.json` alone would have missed both gaps described
above.
