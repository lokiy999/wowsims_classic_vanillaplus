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

## HPS display and rotation notes (2026-09-14, overheal log 2026-09-15)

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
- **Overhealing shown in the Log tab (2026-09-15).**
  `ui/core/proto_utils/logs_parser.tsx`'s `ResourceChangedLog` parses Go's
  `"Gained %0.3f health from %s (%0.3f --> %0.3f)."` log line
  (`sim/core/health.go`'s `GainHealth`), but its regex never captured the
  raw requested amount (the first `%0.3f`) - only the pre/post health
  values, which are already clamped to max HP. So a fully-overhealed cast
  (e.g. Vampiric Embrace healing a full-HP solo-sim target, or any heal
  landing on a topped-off tank) always displayed "Recovered 0.0 Health"
  with no indication real healing happened, even after the HPS-tracking
  fix above made that same healing show up correctly in the aggregate HPS
  number. Fixed by adding a capturing group around that raw amount
  (shifting every later `match[N]` index up by one - `rawAmount` is now
  `match[4]`, the rest reindexed accordingly) and rendering `(X overheal)`
  next to the line whenever `rawAmount - actualGain > 0.05`. Only applies
  to gains (not spends) - there's no "under-spend" analog.

  **Two more bugs found verifying this in-browser once the dev server came
  back up (2026-09-15):**
  1. `<span className="text-muted">` is Bootstrap's default dark gray,
     meant for light backgrounds - on this app's dark theme it rendered at
     ~invisible contrast (confirmed via `getComputedStyle`:
     `rgba(33,37,41,0.75)` text on a page whose surrounding text is
     `rgb(255,255,255)`). The data was correct in the DOM the whole time
     (confirmed via `el.querySelector('.text-muted').textContent`) but
     unreadable. Fixed by dimming via `opacity: 0.7` on the span instead of
     `.text-muted`, which inherits (and stays visible against) whatever the
     surrounding text color already is - no app-wide dark-theme override
     for `.text-muted` exists to fix this more generally, so any other spot
     using that Bootstrap class should be treated as suspect too.
  2. Debugging *this* surfaced a real, unrelated, pre-existing Vampiric
     Embrace bug that had nothing to do with the log display: nearly every
     `Recovered 0.0 Health` line showed a debug `BaseHealing:0.0`, meaning
     the heal amount computed to genuinely zero, not just "clamped to
     zero". Root cause: `dealDamageInternal` (`sim/core/spell_result.go`)
     only calls `OnSpellHitTaken` for **non-periodic** damage; DoT ticks go
     through `OnPeriodicDamageTaken` instead. Vampiric Embrace's aura only
     hooked `OnSpellHitTaken`, so a Shadow Priest's overwhelmingly-DoT
     damage (Shadow Word: Pain, Devouring Plague, Mind Flay) never
     triggered the heal at all - only the rare direct Mind Blast hit did,
     alongside a flood of harmless-but-noisy zero-damage "spell landed"
     trigger events from every other spell's initial application. **Fixed**
     (`sim/priest/vampiric_embrace.go`) by hooking the same handler on both
     `OnSpellHitTaken` and `OnPeriodicDamageTaken`, plus an early return
     when the computed heal is `<= 0` to cut the log noise. Verified this
     was a real, large-effect bug, not a rounding nit: in-browser HPS for
     the user's live rotation went **43 -> 166** after this fix (nearly
     4x) - Vampiric Embrace was healing for essentially nothing before,
     across every session of this pipeline work back to whenever this
     talent was first implemented, not just something introduced this
     session.

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

## Open issues (TODO, not yet fixed)

- **Mana-cost percentage reductions stack additively instead of
  multiplicatively — flagged 2026-09-15, NOT fixed.** `ManaCostOptions.
  Multiplier` (`sim/core/mana.go:304-317`) is a single `int32` percentage
  that every talent/effect just subtracts from directly (e.g.
  `spell.Cost.Multiplier -= 5*rank` in `applyMentalAgility`,
  `sim/priest/talents.go:98`; `Multiplier: 100 - 5*rank` in Shadow Word:
  Pain's config, `sim/priest/shadow_word_pain.go:55`), then applied once as
  a single factor. Real WoW (confirmed against the user's own in-game log)
  applies each source as its own sequential multiplicative step, rounding
  after each: Shadow Word: Pain Rank 8 with Improved SW:Pain 2/2 (-10%) and
  Mental Agility 3/3 (-15%) should cost `470 × 0.90 = 423` → `423 × 0.85 =
  359.55` → rounds to **360** (matches the user's in-game reading exactly).
  The sim instead computes `470 × (1 - 0.10 - 0.15) = 352.5` — off by
  ~7.5 mana in this case, and the error grows with more stacked
  reductions. Individual per-talent magnitudes (e.g. Mental Agility's
  5/10/15%) were already DBC-audited and are correct in isolation — this
  is purely a stacking-order/architecture issue. **Not priest-specific**:
  `Cost.Multiplier` is touched in 43 places across every class's talent
  files, so a real fix means changing the stacking model (e.g. an
  `int32` sum → a multiplicative float chain) sim-wide, not a one-line
  patch. Needs scoping/buy-in before starting.

- **Mind Flay per-tick damage lower than in-game — flagged 2026-09-15,
  NOT fixed, root cause not yet found.** User's in-game reading: Rank 6
  Mind Flay, 5 ticks (tick count confirmed correct per the FIXED entry
  above), **560 damage per tick**. Sim log for the same cast: **407.14**
  per tick — about 27% low, too large a gap to be float/rounding display
  noise. The total base damage per rank (`MindFlayBaseDamage`, e.g. 900
  for Rank 6) was already independently verified against `Spell.csv`
  (`(EffectBasePoints+1) * 5 ticks`, high confidence, see the FIXED entry
  above) and end-to-end tested in a real sim run, so the base total is
  probably not the culprit. Unexplored so far: the flat `0.15` spell
  coefficient (`spellCoeff` in `sim/priest/mind_flay.go:57`, commented as
  "classic penalty for mf having a slow effect" — no DBC citation) and
  interaction with Shadow Weaving/Darkness/other %-damage talent stacking
  on this specific spell. Needs the caster's Shadow Power/talent state at
  cast time to reconstruct the expected number and isolate which factor
  is off.

  Back-of-envelope estimate (2026-09-15, math only, not verified against
  `Spell.csv` or applied to code): using the sim's own numbers at
  `coeff=0.15` — base damage/tick `900/5=180`, character sheet Shadow
  Power `740`, observed tick `407.14` — solve `(180 + 0.15×740) × M =
  407.14` → `M ≈ 1.3995` (the bundled Shadow Weaving/Darkness/etc.
  multiplier, independent of the coefficient). Holding `base` and `M`
  fixed and solving for the coefficient that yields the in-game 560:
  `(180 + coeff×740) × 1.3995 = 560` → `coeff ≈ 0.298`, i.e. **~0.30**.
  For comparison, `coeff=0.35` overshoots to `(180+0.35×740)×1.3995 ≈
  614`. So if the true issue is purely this flat coefficient being
  under-tuned, **~0.30 fits the single data point much better than
  0.35** — but this is one sample, back-solved algebraically assuming
  `M` and Shadow Power were constant at cast time; it has not been
  cross-checked against `Spell.csv` and should be treated as a lead, not
  a confirmed value.

  **Corroborating in-game samples (2026-09-15/16), all at Shadow Weaving
  1 stack (1.02× taken multiplier) — every test the user runs uses this
  same 1-stack condition:**

  | Shadow Power | Tick damage |
  |---|---|
  | 739 | 539 |
  | 721 | 532 |
  | 733 | 536 |
  | 800 | 565 |

  The first three samples (SP 721-739, a ~2.5% range) initially favored
  `coeff≈0.30` over `0.15`/`0.35` (see prior paragraph), but that range was
  too narrow to be conclusive. The SP=800 sample was deliberately requested
  as a wide-range discriminating test (see reasoning below the table) and
  **it invalidates the 0.30 estimate**: re-solving `M` at `coeff=0.30`
  across all four points now gives 1.340-1.345, a 0.37% spread (up from
  0.15% on 3 points) — the fit degrades once the SP range widens.

  Linear-regressing all 4 points as `tick = A + B×SP` (least squares):
  `B≈0.4230`, `A≈226.5`. Since the model is `tick=(180+coeff×SP)×M`, i.e.
  `A=180M` and `B=coeff×M`: `M=A/180≈1.258`, `coeff=B/M≈0.336` — close to
  the clean fraction **1/3 (0.333)**. Solving `M` per-point at `coeff=1/3`
  exactly:

  | SP | 180+SP/3 | Implied M |
  |---|---|---|
  | 739 | 426.33 | 539/426.33 = 1.2643 |
  | 721 | 420.33 | 532/420.33 = 1.2658 |
  | 733 | 424.33 | 536/424.33 = 1.2631 |
  | 800 | 446.67 | 565/446.67 = 1.2650 |

  Spread: **0.08%** (M≈1.264±0.001) — the tightest fit found so far,
  noticeably better than 0.30's 0.37% or 0.35's ~0.22% once all 4 points
  are used. Current best working hypothesis: **the Mind Flay spell
  coefficient should be ~1/3 (0.333), not the current 0.15**, with roughly
  a 1.264 flat multiplier from other sources (Shadow Weaving 1 stack ×1.02
  is part of this 1.264, the rest presumably Darkness/Shadowform/similar
  talent bonuses) at this talent build. Still unverified against
  `Spell.csv` and not yet applied to code — the reasoning above is pure
  back-solving from in-game samples, not confirmed from spell data.

  **Why SP=800 was the requested test:** with only the narrow 721-739
  range, the coefficient's contribution to the tick barely moved between
  hypotheses, so almost any coefficient could fit with some M. A wide-SP
  sample makes the coefficient term dominate the total, separating the
  hypotheses clearly - exactly what happened here (0.30 held up on the
  narrow range but broke once the 800 SP point was added).

  **Fifth sample, low end (2026-09-16): SP=100, tick=282, same 1-stack
  Shadow Weaving.** This is the low-SP counterpart test suggested above,
  and it reverses the previous conclusion. At `coeff=1/3`:
  `(180+100/3)×M=282` → `M=282/213.33=1.322` — a **4.5% outlier** against
  the other four points' 1.264 cluster (at low SP the flat `180` base
  dominates, so this is exactly where a wrong coefficient gets exposed;
  1/3 fails the test). Re-checking `coeff=0.30` against all 5 points
  instead:

  | SP | 180+0.30×SP | Implied M |
  |---|---|---|
  | 100 | 210.0 | 282/210 = 1.3429 |
  | 721 | 396.3 | 532/396.3 = 1.3424 |
  | 733 | 399.9 | 536/399.9 = 1.3403 |
  | 739 | 401.7 | 539/401.7 = 1.3418 |
  | 800 | 420.0 | 565/420 = 1.3452 |

  All five land within **0.37%** (M≈1.342-1.345), with the new SP=100
  point sitting right in the middle of the cluster instead of breaking it.
  **Revised conclusion: `coeff≈0.30` is confirmed (not 1/3), M≈1.342.**
  The 1/3 hypothesis only looked good on the earlier 4-point set because
  none of those points had a small-enough SP for the coefficient term to
  matter much relative to the flat base - a good illustration of why a
  low-SP sample is the more decisive test, as suggested above.

  **Three more samples, wide low-to-mid range (2026-09-16), same 1-stack
  Shadow Weaving:** SP=45→tick=254 (lowest achievable in current gear),
  SP=246→tick=339, SP=405→tick=404. Combined with the prior 5, this is now
  **8 samples spanning SP 45-800 (an 18x range)**. Linear-regressing all 8
  as `tick=A+B×SP`: `B≈0.4071`, `A≈238.6` → `M=A/180≈1.325`,
  `coeff=B/M≈0.307`. Checking fit quality at `coeff=0.307, M=1.325`:

  | SP | tick | Predicted | Diff |
  |---|---|---|---|
  | 45 | 254 | 256.9 | +1.1% |
  | 100 | 282 | 279.3 | -1.0% |
  | 246 | 339 | 338.6 | -0.1% |
  | 405 | 404 | 403.4 | -0.1% |
  | 721 | 532 | 531.9 | 0.0% |
  | 733 | 536 | 536.8 | +0.1% |
  | 739 | 539 | 539.3 | +0.1% |
  | 800 | 565 | 564.1 | -0.2% |

  All 8 within ~1%, across an 18x SP range - about as solid as back-
  solving from in-game samples alone can get. `1/3` is now decisively
  rejected: it systematically under-predicts every low-SP point by
  3-4.4% (a consistent bias, not noise) - e.g. at SP=45,
  `(180+45/3)×1.264≈246.5` vs actual 254; at SP=100, `≈269.6` vs actual
  282.

  **Current best estimate: `coeff≈0.30-0.31`, `M≈1.32-1.33`** (the ~0.307
  regression value and the earlier ~0.30 both sit inside this band; the
  small residual spread left is consistent with real in-game
  damage-display rounding, not a wrong coefficient).

  **`Spell.csv` checked (2026-09-16) — no coefficient field exists.**
  Searched every `float`-typed column (indices 73-75, 97-99, 112, 167-169,
  identified from the CSV's own type header row) plus the known
  `EffectBasePoints`/`EffectAmplitude` columns, across all 6 Mind Flay
  ranks (15407, 17311-17314, 18807) and their tick sub-spells. All the
  `float` columns are flat `0`/`1` across every rank (boolean-looking
  flags, not a percentage). This 1.12-era DBC dump simply doesn't store a
  spell-power coefficient anywhere - that only became data-driven in
  later expansions; in vanilla/Classic it's a client-hardcoded
  cast-time/channel-time formula, not a DBC field. (One useful
  cross-check survived: rank 6's `EffectBasePoints[0]=179` → `+1=180`,
  confirming the per-tick base used throughout this analysis.) So there is
  no DBC value to verify ~0.30 against - the 8-sample empirical fit above
  is the best evidence available.

  **FIXED 2026-09-16** (`sim/priest/mind_flay.go`) - `spellCoeff` changed
  from the old (undocumented, no citation) `0.15` to **`0.30`**, per the
  empirical derivation above. This is **not DBC-verified** (see previous
  paragraph - no such field exists to check against), it's back-solved
  from live gameplay samples. If a cleaner value later gets independently
  confirmed (e.g. the classic channel-coefficient formula
  `tickLength/3.5 ≈ 0.2857` was tested against the same 8 points and
  fits distinctly worse - a systematic ~5% upward drift in implied `M` as
  SP increases, versus 0.30's flat sub-1% residuals - so it was rejected
  in favor of the empirical value), update this entry and the code
  together.
