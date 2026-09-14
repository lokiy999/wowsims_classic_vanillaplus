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
  - **Improved Shadow Word: Pain (talent) — mana-cost half FIXED 2026-09-14**
    (`sim/priest/shadow_word_pain.go`). This talent has two effects (spells
    15275/15317): +3/+6 sec duration (already correctly implemented as +1/+2
    extra ticks) and a **5%/10% mana cost reduction that was TODO'd out and
    never implemented**. Added `Multiplier: 100 -
    5*priest.Talents.ImprovedShadowWordPain` to the spell's `ManaCostOptions`.
    `go test ./sim/priest/...` produced a byte-identical `.results` — the P1
    shadow preset's default talent build has 0 points in this talent, so
    the mana-cost multiplier is a no-op (100) there; verify separately with a
    talent build that actually invests in it.
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
  - **Smite — FIXED 2026-09-14, confirmed 2026-09-14** (`sim/priest/smite.go`),
    same `EffectBasePoints[1]+EffectDieSides[0]` pattern as Mind Blast
    (identical effect-slot structure across all 8 ranks). Rank 8 (371-415)
    confirmed correct against the user's in-game tooltip.
  - **Devouring Plague — FIXED 2026-09-14** (`sim/priest/devouring_plague.go`).
    Pure periodic DoT, `EffectDieSides[0]` is 1 for every rank (no variance,
    same slot as `EffectBasePoints[0]` — the reliable same-slot pattern, not
    Mind Blast/Smite's cross-slot case), so `(EffectBasePoints+1) * 8 ticks`
    applies directly with no ambiguity, same as Shadow Word: Pain/Starshards.
    All 6 ranks corrected; old values were a consistent ~10-12% too high
    (e.g. rank 6 was 904, corrected to 816). `go test ./sim/priest/...` DPS
    changed from 246.667 to 246.061 on the P1 shadow preset (small - Devouring
    Plague is a minor part of that rotation); golden `.results` promoted.
  - **Holy Fire — FIXED 2026-09-14** (`sim/priest/holy_fire.go`). Rank 8's
    direct damage (481-604) and DoT total (225) were confirmed against the
    user's own in-game tooltip screenshot, which also revealed a **10 sec
    cooldown the Go code didn't implement at all** (added). The earlier
    "inconsistent effect-slot layout" finding wasn't actually a bug in the
    data - `EffectDieSides` is ROLE-based, not slot-based: `die[0]` always
    pairs with whichever `EffectBasePoints` slot holds the direct-hit value
    for that rank (slot 1 for every rank except rank 3, which uses slot 0),
    and `die[1]` always pairs with the DoT slot. All 8 ranks recomputed with
    this rule and all reconcile cleanly (rank 8 confirmed exactly; ranks 1-7
    follow the identical formula but weren't independently screenshot-
    verified). Old values were under-tuned on both the direct hit (rank 8:
    355-449, ~24% low) and the DoT (rank 8: 145 vs 225, ~36% low), and the
    missing cooldown meant the sim could spam Holy Fire far faster than the
    game allows. `go test ./sim/priest/...` produced a byte-identical
    `.results` (the P1 shadow preset doesn't cast Holy Fire).
  - **Vampiric Embrace — logic bug FIXED 2026-09-14, missing cooldown ADDED
    2026-09-14** (`sim/priest/vampiric_embrace.go`). Values (10%/rank heal,
    60s duration) were already correct. But the `OnSpellHitTaken` hook fired
    for **any** landed Shadow spell hitting the debuffed target, regardless
    of caster — the DBC text is explicit ("of any Shadow spell damage
    **you** deal"). In a solo sim this rarely mattered (no other
    Shadow-damage source hitting the same target), but in a multi-caster
    raid sim it would over-heal by counting every Shadow-damage source, not
    just the priest who cast it. Added a `spell.Unit == &priest.Unit` check.
    Separately, per the user's own in-game report, added a **15 sec
    cooldown** the Go code didn't have at all (same `Timer`/`Cooldown`
    pattern as Mind Blast/Devouring Plague/Holy Fire) — not independently
    re-derived from `Spell.csv` (no cooldown/recharge column has been
    located in the dump yet; Holy Fire's cooldown was confirmed the same
    way, from an in-game screenshot, not a decoded column).
  - **Vampiric Embrace — HPS-tracking bug FIXED 2026-09-14.** Its healing
    used `Unit.GainHealth()` directly, which updates the health bar but
    never populates `SpellMetrics.TotalHealing` - the field the sim's `hps`
    aggregate is actually summed from (`unitMetrics.hps.Total +=
    spellTargetMetrics.TotalHealing + ...TotalShielding`,
    `sim/core/metrics_aggregator.go:433`). So Vampiric Embrace's healing was
    always invisible to HPS, independent of overhealing. Fixed by routing
    it through a proper internal healing spell (`ActionID.WithTag(1)`,
    `ProcMask: ProcMaskSpellHealing`) using `CalcAndDealHealing` - the same
    pattern `sim/shaman/water_totems.go`'s Healing Stream Totem uses -
    instead of calling `GainHealth` directly. `dealHealingInternal`
    (`sim/core/spell_result.go:544`) records the **pre-clamp** healing
    amount into `TotalHealing`, so overhealing is now correctly included,
    matching the HPS tooltip's stated design ("Healing+Shielding Per
    Second, **including overhealing**"). Verified end-to-end in the
    browser: added Vampiric Embrace to the live rotation (see "APL
    priority" note below), ran a sim, and confirmed the sidebar now shows
    a nonzero HPS next to DPS.
  - **Talent audit 2026-09-14** — every currently-implemented priest talent
    in `sim/priest/talents.go` was cross-checked against `Spell.csv`. Most
    check out exactly: SilentResolve, ImprovedPowerWordFortitude, Meditation,
    MentalStrength, ImprovedMemory, MentalAgility, ForceOfWill,
    HolySpecialization, SearingLight, SpellWarding, SpiritualGuidance, Faith,
    InnerFocus, Shadowform (the implemented +20% Shadow damage part; its
    unimplemented -20% Shadow damage taken TODO is DBC-confirmed accurate),
    Concentration (proc% and Clearcasting duration), MindOverlord,
    ImprovedMindFlay, BurntSoul (proc%, magnitude, and duration),
    ImprovedVampiricEmbrace, Darkness, ShadowAffinity.
    - **Twin Disciplines — FIXED 2026-09-14.** Proc chance was hardcoded flat
      at 10% for all 3 ranks; the DBC (spells 33822/33823/33824) confirms
      10/20/30% per rank, the same progression every other ranked proc
      talent in this file uses. Also fixed the granted buff's duration: Go
      had 15s, but the actual applied buff (spell 33832) has
      `DurationIndex`=1 → 10000ms (10s). `go test ./sim/priest/...` produced
      a byte-identical `.results` (the P1 shadow preset has 0 points here).
    - **Known partial implementations, not fixed** (out of scope for this
      pass - flagged, not touched): Spell Focus's target-resistance-
      reduction sub-effect ($s2, -10%/-20% target resist chance) isn't
      modeled at all, only the hit-chance sub-effect is; Purifying Light's
      Undead/Demon damage bonus sub-effect ($s2, +5%/+10%) isn't modeled;
      Shadow Focus's Shadow-spell-range sub-effect ($s2) isn't modeled
      (irrelevant to damage output, likely low priority). Spirit Tap's
      simplification (flat aura instead of a real 50%/100%-per-rank
      on-kill proc) was already a documented `TODO` before this audit and
      remains one — the audit did confirm its underlying magnitude/duration
      values (100% Spirit, 15s) are correct.
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
    real min-max range for direct-damage spells. The pairing between it and
    `EffectBasePoints` is **ROLE-based, not slot-index-based**: for a
    single-effect spell (Rend/SW:Pain/Devouring Plague/Starshards) both
    values sit in slot 0 so this looks like same-slot pairing; for a
    single-damage-effect spell whose value happens to live in a different
    slot (Mind Blast/Smite, always slot 1), `die[0]` still pairs with it;
    for a two-effect spell (Holy Fire), `die[0]` = direct-hit role and
    `die[1]` = DoT role, tracking whichever `EffectBasePoints` slot holds
    that role even when it moves between ranks (Holy Fire rank 3 vs. the
    rest). Always resolve by role (which effect is the direct hit / the DoT
    / etc.), never by raw column index, and prefer an in-game tooltip
    confirmation over the formula alone when in doubt.

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

## HPS display and rotation notes (2026-09-14)

- **`ui/core/components/raid_sim_action.tsx`**: the individual-sim results
  sidebar now shows an `HPS` line next to `DPS` whenever a sim actually
  produces healing (gated on `hpsMetrics.avg`), not just for specs flagged
  as "healing specs". Two things had to be fixed together for this to
  actually work end-to-end, not just for a designated healer:
  1. `makeToplineResultsContent()`'s single-player branch never pushed an
     `HPS` `resultColumn` at all — only `DPS`/`DPASP`/`TPS`/`DTPS`/`TMI`/`COD`.
  2. Even after adding it, the line stayed invisible: `.hide-healing-metrics`
     (`ui/scss/core/sim_ui/_shared.scss`) force-hides every `.healing-metrics`
     -classed element site-wide whenever `Sim.showHealingMetrics` is false,
     which `Sim.applyDefaults()` sets purely from spec type
     (`showHealingMetrics: isHealingSim`) — independent of whether *this*
     particular sim run has any real healing data. Fixed by omitting the
     `healing-metrics` category class from just this one topline push (kept
     `results-sim-hps` for styling/tooltip lookup), so it's gated only by
     the `hpsMetrics.avg` check, not by spec type. The old
     `getResultsLineClasses('hps')` behavior (which the full-detail healing
     tab and the healer specs still use) is untouched.
- **APL priority-list pattern for "keep this buff up" style actions**: to
  add "cast Vampiric Embrace if it's not on the current target, or has
  <=15s remaining" to a rotation, the condition is
  `or(not(auraIsActive(sourceUnit: CurrentTarget, auraId: <spellId>)),
  cmp(auraRemainingTime(sourceUnit: CurrentTarget, auraId: <spellId>) <=
  <seconds>))`. Raw JSON example (from `ui/shadow_priest/apls/p1.apl.json`):
  ```json
  {"action":{"condition":{"or":{"vals":[
    {"not":{"val":{"auraIsActive":{"sourceUnit":{"type":"CurrentTarget"},"auraId":{"spellId":15286}}}}},
    {"cmp":{"op":"OpLe","lhs":{"auraRemainingTime":{"sourceUnit":{"type":"CurrentTarget"},"auraId":{"spellId":15286}}},"rhs":{"const":{"val":"15"}}}}
  ]}},"castSpell":{"spellId":{"spellId":15286}}}}
  ```
  Added to the checked-in `p1.apl.json` default preset (currently a no-op
  there — that test's `P1Talents` is `""`, zero points, so Vampiric Embrace
  never registers) and, via the app's own JSON Export → splice → Import
  flow (far more reliable than drag-and-drop in the APL editor UI), to the
  user's actual talented live rotation, where it correctly increased
  Vampiric Embrace's cast rate from ~0.1 casts/fight (buried at the bottom
  of the priority list, essentially never reached) to ~4-5 casts/fight.

## TODO

- **Talent-aware tooltips for Shadow Word: Pain and Vampiric Embrace**
  (raised 2026-09-14, not started). Right now every spell's tooltip is
  static text baked once from `Spell.csv`/Wowhead, completely independent
  of the player's current talent build — real Wowhead works the same way
  (a talent's effect shows on the talent's own tooltip, not merged into the
  base spell's). The ask is to make SW:Pain's tooltip reflect Improved
  Shadow Word: Pain's actual duration/mana-cost change, and Vampiric
  Embrace's tooltip reflect Improved Vampiric Embrace's heal % change, live,
  based on current talent points. There's no existing mechanism for this —
  `spellIdTooltipOverride` (`ui/core/proto_utils/action_id.ts:552`) is the
  closest existing thing, but it's a static spellId→spellId map, not
  talent-reactive. Building this would mean, at tooltip-render time for
  these two specific spells: read the live `priest.Talents.*` value,
  recompute the real modified duration/mana-cost/heal% (already known -
  see the FIXED entries above for SW:Pain and Vampiric Embrace), and
  rewrite the relevant part of the tooltip HTML string before display
  (targeted regex substitution, client-side, analogous to what
  `gen_spell_value_overrides.py` already does once at build time for the
  `Channeled (N sec cast)` fix, but done dynamically instead). Scope is
  small per-spell, not a generic system - no Go/simulation changes needed,
  this is purely cosmetic/display.
