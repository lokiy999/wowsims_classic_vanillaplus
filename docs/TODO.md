# Open items

Things either of us raised across this work that haven't been done yet. Not a
backlog of ideas — only things actually said and left open. Grouped by how they
came up, most concrete first.

_As of 2026-09-19._

## Raised 2026-09-19 — after the talent-calculator cross-check (CHANGES.md Part AP)

Still open, the calculator confirmed the text but not the mechanics:
- Priest **Power Word: Requital** (implemented 2026-09-19, but coefficient is a placeholder 1.5/3.5 and ranks 2-4 lack DB name/icon; original note): instant, 150 mana, 20s cooldown, 185.5 Holy damage (346-371 if the target is
  feared/stunned/incapacitated). Not implemented: spell power coefficient unknown, and it needs an APL action.
- ~~Warrior **Enrage**:~~ DONE 2026-09-19 (see CHANGES.md Part AP follow-up 2). Old note: "1% melee damage per stack for 15s, up to 10 stacks" at rank 1. Code still uses a flat
  5%/rank with 12 swing-counting stacks. Waiting on rank 2+ text.
- ~~Warrior **Tactical Mastery**:~~ DONE 2026-09-19. Old note: "retain an additional 10 Rage" at rank 1 (code 5/rank).
- Mage **Mind Mastery**: also +20%/rank Arcane Intellect effect (buff pipeline, not done).
- Not modeled: Shaman Armaments of Storm (5%/rank, up to 300 Nature, level scaled), Shamanism, Aftershock;
  Rogue Improved Sinister Strike proc, Coup de Grace, Gaining an Advantage, Brigandage; Hunter Find Weakness,
  Deadeye, Thrill of the Hunt; Warlock Defiler, Prolonged Misery, Demonic Embrace regen; Druid Killer Instincts
  attack speed part, Power of Nature; Warrior Cleaving (Thunder Clap/Whirlwind), Improved Execute.

## Raised 2026-09-19 — warrior talents (CHANGES.md Part AO)

- Not modeled: Improved Rend stacking, Improved Mortal Strike, Improved Bloodthirst, Improved
  Berserker/Battle Stance, Enrage (DBC 1%/stack; current code is 5%/rank, unverified),
  Cleaving, Improved Execute, Improved Bloodrage, Tactical Mastery values, Weapon Expertise
  values, Deep Wounds values, Booming Voice/Improved Combat Shouts (core buff code),
  Duelist rage part, Butterfly Style, Constitution, Shield Mastery, Improved Shield Wall.
- Para Bellum applies to every warrior spell with a cooldown and a spell code; verify it
  does not wrongly shorten stance-change or shared-cooldown timers.
- Warrior `.results` files stale.
- **Whole-pass note:** all nine classes now have stale `.results` (see the 2026-09-18 item
  above); regenerating them is a separate review-sized task.

## Raised 2026-09-19 — rogue talents (CHANGES.md Part AN)

- Not modeled: Improved Sinister Strike extra-hit proc (3/5%), Coup de Grace (+5%/rank
  damage below 20% health), Bloodthirsty (Garrote/Rupture +10%/rank damage, shorter ticks),
  Exhaustion, Combat Rush, Brigandage, Gaining an Advantage, Dazing Bolts, Survivor,
  Physical Prowess, Improved Kidney Shot, Remorseless Attacks, Weapon Expertise values.
- Connivery is applied to all melee damage; DBC says attacks from behind only.
- Rogue `.results` files stale.

## Raised 2026-09-19 — hunter talents (CHANGES.md Part AM)

- Not modeled: Melee Specialization (+30% melee speed, -30% ranged), Dual Wield
  Specialization (OH +20-50%), Weapon Expertise (extra arrow), Find Weakness, Thrill of the
  Hunt, Deadeye, Improved Hunter's Mark, Stalking, Vantage Point, Spirit Bond, Kill Command,
  Improved Tracking, Deep Freeze, Savage Blow, Whirling Axe, most pet utility talents.
- Reconnaissance is coded 3%/rank all damage; DBC spell 34063 (3%) may be flat at every
  rank. Unchecked.
- Hunter `.results` files stale.

## Raised 2026-09-19 — paladin talents (CHANGES.md Part AL)

- Not modeled: Divine Concentration, Divine Grace, Holy Grasp, Illumination, Light's Mercy,
  Improved Purifying, Holy Purge, Divine Might/Improved Sanctity Aura/Sanctity Aura
  (buff pipeline, see the buff audit above), Seal of Command/Fury, Redoubt values,
  Eye for an Eye, Repentance, Unbreakability, Improved Retribution Aura, Codex Holy Light
  cost/cast time, Blessed Strikes threat reduction, Vengeance max stacks (code 10, unchecked).
- Improved Lay on Hands cooldown (code -10 min/rank) unchecked against DBC (-15/-30 min?).
- Paladin `.results` files stale.

## Raised 2026-09-19 — shaman talents (CHANGES.md Part AK)

- Elemental Devastation: DBC aura 30165 gives a flat 10% crit at every rank, code uses 3%/rank.
  Looks odd, so not changed; confirm on the server.
- Not modeled: Static Field, Armaments of Storm, Shamanism, Aftershock, Earthquake, Earth
  Shield, Primal Endurance, Guardian Totems (partly in core), Rockhide, Improved Fire Totems,
  Improved Ghost Wolf, Blood Lust, Elemental Mastery values unchecked, all healing talents
  (Healing Way, Purification damage part, Tidal Mastery speed part), Nature Grace/Guardian.
- Healing Stream Totem base healing formula in `water_totems.go` (`base*purification + restorative`)
  looks wrong (adds instead of multiplies); left alone.
- Shaman `.results` files stale.

## Raised 2026-09-19 — druid talents (CHANGES.md Part AJ)

- ~~**Improved Mark of the Wild conflict:**~~ RESOLVED 2026-09-19 (20%/rank confirmed by user). DBC gives 20/40/60% (3 ranks), but the buff audit
  above expects the improved Gift of the Wild armor to be 285x1.35. One of them is wrong;
  not touched. Check which value is actually live on the server.
- Not modeled: Starlight Wrath/etc. are fine, but Starfall, Power of Nature, Dreamstate,
  Mighty Roots, Hurricane, Cycle of Life, Unity with Nature, Stalking, Predatory Strikes
  (custom DBC values look malformed), Primal Fury, Survival Instincts, Leader of the Pack
  crit value (DBC 3%), Furor values, all healing talents (Improved Rejuvenation/Regrowth,
  Gift of Nature healing part, Tranquil Spirit), Swiftbloom, Naturalist.
- Druid `.results` files stale.

## Raised 2026-09-19 — mage talents (CHANGES.md Part AI)

- Not modeled: Shatter (needs a frozen-target state), Spell Twisting, Impact, Blazing
  Speed, Chain Reaction, Thermal Expansion, Cryo Core, Cold Grip, Rimebound (needs
  spell 34266), Advanced Ice Shielding, Frost Warding/Fire Warding, Practical talents,
  Brilliance Aura, Magic Absorption mana-on-resist, Arcane Intellect +%/rank from Mind Mastery.
- Hot Streak/Pyromania untested in a fire rotation (default mage APL is not fire).
- Mage `.results` files stale.

## Raised 2026-09-19 — warlock talents (CHANGES.md Part AH)

- Not modeled: Defiler (-20% tick time/duration/GCD), Prolonged Misery (+2/4/6s dot
  duration, not a multiple of the 3s tick), Jinx, Improved Drain Soul (DBC: Drain
  Soul cooldown -5/-10s + mana/health on kill; code wrongly uses it as a Curse of Doom
  threat reduction in `curses.go`), Inevitable Doom, Herald of Woe, Sadism, Feeding
  Demons, Fel Pact, Demonic Onslaught, Improved Immolate (2 stacks; current 5% bonus
  is a guess), Pyroclasm/Mayhem/Shock and Awe, Demonic Embrace regen, Soul Link
  damage split (redirect is 20% per DBC; code uses a flat multiplier).
- Master Demonologist Felhunter resist and Voidwalker values were assumed to scale
  3/rank and 2/rank from rank-1 DBC only; ranks 2-5 of the sub-spells not checked.
- Warlock `.results` files stale (see the existing stale-results item above).

## Raised 2026-09-19 — priest talents (CHANGES.md Part AG)

- Healing priest is not simulated: `registerFlashHealSpell`/`GreaterHeal`/`PWS`/
  `PoH`/`Renew` are commented out. The commented code also has stale talent
  values (e.g. Spiritual Healing 2%/rank; server DBC is 3%/rank, Improved Renew
  5%/rank). Fix the values when re-enabling.
- Unmodeled priest talents: Purifying Light undead/demon bonus, Spirit Tap
  on-kill proc chance, Power Word: Requital (new server spell), Improved Inner
  Fire, Wand Specialization, Blackout, Blur, Insanity, Shadow Word: Numb.
- Talent hit/crit (Spell Focus, Force of Will, Holy Specialization) is applied
  per spell, so it does not show on the sidebar Spell Hit/Crit lines.
- Next classes to audit: warlock, mage, druid, shaman, paladin, hunter, rogue,
  warrior.

## Raised 2026-09-19 — player buff audit (table still being worked through; nothing implemented yet)

Going through `sim/core/buffs.go` buff-by-buff with the user. Agreed so far:

- **Devotion Aura is wrong in the sim.** Code uses base 735 and +12.5%/talent
  point (max ×1.25). Should be base **700**, Improved Devotion Aura **+25%/point,
  2 points (max ×1.50 → 1050)**. Libram of Truth adds a flat **+55 to the base
  before the talent multiplier**: (700+55)×1.5 = 1132.5. Make the Libram a
  toggle like the other tristates (default / enhanced / enhanced+ — i.e. none,
  talents, talents+Libram; exact UI wording to be settled). Touches
  `BuffSpellValues[DevotionAura]` and `DevotionAuraAura` in `sim/core/buffs.go`,
  plus the raid-buff picker in `ui/core/components/inputs/buffs_debuffs.ts`
  (proto change likely needed for a third state).
- **Commanding Shout is SoD-only — remove it** (`CommandingShout` in
  `BuffSpellValues` and its raid-buff toggle/proto field).
- **Horn of Lordaeron — remove** (user asked to drop it from the list).
- `IncludeAQ` in `sim/core/config.go` is `false`, but the user's server has the
  AQ spell books — flip it (or make it phase-driven) so Battle Shout / Blessing
  of Might / Grace of Air / Strength of Earth use the AQ rank values.
- **Resistance totems/auras get an Improved state:** Fire/Frost/Nature
  Resistance Totem and Fire/Frost/Shadow Resistance Aura are 60 base, +50% from
  talents → **90**. Currently only a plain 60 in `BuffSpellValues` with boolean
  toggles (no tristate), so needs a proto/UI change to a tristate. Keep the
  existing subtraction of Gift of the Wild's resist (`bonusResist`) working
  against the 90 value.
- **Sanctity Aura is not Alliance-only on this server** — Horde can get it too.
  `sim/core/buffs.go` gates it with `raidBuffs.SanctityAura && isAlliance`;
  drop the faction check (and any UI-side faction hiding in
  `ui/core/components/inputs/buffs_debuffs.ts`). **Confirmed by user: both
  factions can use all faction-gated buffs** — so also drop `isAlliance` /
  `isHorde` gating on Devotion Aura, Retribution Aura, Blessing of Might,
  Stoneskin Totem, Strength of Earth Totem, Grace of Air Totem.
- **Sanctity Aura gets an Improved state:** base +10% Holy damage, talents +5%
  → **15%**. Currently hardcoded ×1.1 with a boolean toggle in
  `SanctityAuraAura`; needs a tristate (proto/UI) and 1.15 for improved.
- **Trueshot Aura is wrong:** code gives +100 AP and +100 RAP; should be
  **+50 melee AP and +100 RAP** (`TrueshotAura` in `sim/core/buffs.go`,
  `meleeAP := 100.0` → 50).
- **Moonkin Aura is wrong:** code gives +3% spell crit; should be **+5% to all
  spell damage and healing** (a damage/healing multiplier, not crit). Fix in
  the `raidBuffs.MoonkinAura` block of `sim/core/buffs.go` (currently
  `SpellCrit` 3×`SpellCritRatingPerCritChance`) — need to add a spell damage +
  healing done multiplier (×1.05) instead. Confirmed by user: stacks
  multiplicatively with other damage multipliers and applies to the Moonkin
  druid's own spells too.
- **Arcane Intellect is wrong and needs Improved states:** code has a flat
  +31 Int with a boolean toggle. Should be **30 base**, talents **+100%** (→60),
  ZG set **+25% more**. User believes the full-improved total is **67**
  (additive, 67.5 floored) — **ask again at the end of the list**; user will
  verify in-game (alternative if multiplicative: 75). Needs a proto/UI change from boolean to
  multi-state (none / talents / talents+ZG set, etc.).
- **When the buff list is finished, consolidate all of the above into one
  to-do** (user's request) — not done yet.
- ~~Fix while here: improved Gift of the Wild armor is 384 (285×1.35 floored).~~ Done 2026-09-19: Improved MotW is 20%/rank (x1.6, armor 456), see CHANGES.md Part AP.

## Raised 2026-09-18 — item pipeline: local `CSV's/` data has drifted from the committed DB

While hand-patching in the 6 new rank-V scrolls (81010-81015, see `docs/CHANGES.md`
Part AE), running the normal full pipeline (`parse_vplus.py` → `gen_include.py` →
`gen_db`) picked up a huge unrelated diff — `custom_items.json` alone changed by
~223k lines, `included_items.json` by ~3900 ids. Root cause: `CSV's/` (the raw
VPlusItemDB/AtlasLoot dump) has been gitignored and untracked since commit
`a96980e6d`, and the local copy on disk has clearly been updated (presumably by
a server patch) since the pipeline was last run against it — nobody has re-run
the full pipeline since, so the committed `assets/db_inputs/*.json`/`assets/database/*`
are stale relative to whatever's currently in `CSV's/`. Avoided doing that full
resync as an unplanned side effect of the scroll task (reverted it, hand-patched
just the 6 new items instead). If you want the DB brought current with your
latest server dump, that's a distinct, reviewable piece of work — expect it to
touch item stats/inclusion/phases broadly, not just a handful of items.

## Raised 2026-09-18 — pre-existing `go test ./sim/...` failures, unrelated to this session

`go test ./sim/...` fails the same ~15 packages (`warrior/tank_warrior`,
`druid/balance`, `druid/feral`, `hunter`, `mage`, `paladin/protection`,
`paladin/retribution`, `priest/shadow`, `rogue/dps_rogue`, `shaman/elemental`,
`shaman/enhancement`, `shaman/warden`, `warlock/dps`, `warrior/dps_warrior`) on
a clean `HEAD` checkout with no changes at all — confirmed by stashing
everything from the Part AE session and re-running; identical failure list both
times. Committed `.results` files are stale relative to current item/racial/etc
data (same situation flagged before, e.g. Part R's "committed sim/*/*.results
are STALE" note). Not touched here since regenerating dozens of `.results`
files is a separate, reviewable piece of work, not a side effect of adding 6
scrolls. Run `make test && make update-tests` (or the equivalent manual
`go test` + promote-`.results.tmp` steps) when you're ready to take that on.

## Raised 2026-09-17 — consumables that exist on the server but aren't in the sim at all

Full survey in `docs/CHANGES.md` Part P; most of it implemented in Part Q.
These are in `VPlusItemDB.lua` with real stat-buff tooltips.

**Implemented (Part Q):** Flask of Indomitable Might, Elixir of Brute
Force, Elixir of Demonslaying, Elixir of Greater Intellect, Elixir of the
Sages, Juju Guile, and all 4 Troll's Blood Potions — new `proto/common.proto`
enum values (`IntellectElixir` is a brand-new category), regenerated
Go/TS bindings, `sim/core/consumes.go` implementation, and
`ui/core/components/inputs/consumables.ts` / `consumes_picker.ts` pickers.

**Explicitly skipped (user's call, not a technical blocker):** Bloodkelp
Elixir of Dodging (22192) / Resistance (22193), Elixir of Wisdom (3383),
Potion of Fervor (1450), Minor Magic Resistance Potion (3384), Combat
Healing/Mana Potion (18839/18841). Still absent from the sim if wanted
later.

**Still genuinely blocked — needs a new engine feature, not just a
consumable add:**

- The 6 `Greater X Protection Potion` + 6 base tiers (Arcane/Fire/Frost/
  Holy/Nature/Shadow) — `Potions` proto enum already has the 6 Greater
  values, and the UI file already has them written out (commented, with an
  existing `TODO: ... Missing school shields and shields don't actually
  absorb damage right now` note — a previously-known gap). They're
  damage-absorb shields; there is no absorb-shield primitive anywhere in
  the sim (confirmed via grep across `sim/core`/`sim/priest` — not even
  Power Word: Shield exists). Needs an actual new mechanic before these can
  be added.

## Explicitly deferred to "adjust manually later"

- **Crafted-item phasing is flat.** Every craftable item is Phase 1 right now.
  We'd originally discussed bumping a crafted item to a later phase when its
  pattern *and* a reagent are both BoP from that phase's content (Dark
  Iron/Molten/Corehound/Flarecore → MC; Chromatic/Dreamscale → BWL), but you
  said to skip that and just set everything to Phase 1, adjusting by hand if a
  specific recipe needs it. Nothing's been adjusted since — if any of those
  MC/BWL-reagent recipes should actually sit later than Phase 1, that's still
  outstanding.

## Known gaps I flagged but didn't build

- ~~**Custom class sets have no set bonuses.**~~ Done 2026-09-13: Talonclaw
  Regalia, Ursoc Armor (druid), Cataclysm Armor, The Stonefury (shaman),
  Righteous Armor (paladin) all implemented (commit `d464da4b3`). See
  [CHANGES.md](CHANGES.md) Part C for detail and reproduction steps.
- ~~**All PvP sets need their bonuses done — priority.**~~ Done earlier in the
  2026-09-13 session (commits `79cb702a3`, `590bfad43`) — every class's PvP
  sets (both blue and epic tier) had a universal 2pc/6pc bonus-value swap bug,
  fixed and verified against the server's real tooltips.
- ~~**Really, every set needs a full pass.**~~ Done 2026-09-13 (commit
  `bfcd64728`): audited every class's PvE tier/dungeon sets, one background
  agent per class, three classes at a time. Found and fixed wrong
  piece-count thresholds and/or wrong bonus values in the large majority of
  already-"implemented" sets across all 9 classes — see CHANGES.md Part C for
  the full list. Remaining known gap: **Cenarion Armor's root-cause data bug
  was fixed, but the same underlying question — why did the generic
  Wowhead-derived pipeline fail to tag `setName` for some items when
  `custom_items.json` had no override for them — was never diagnosed**, only
  patched item-by-item as it was found (Righteous Boots, the 8 Cenarion
  Armor pieces, Striker's Garb, a handful of AQ40/Naxx sets that turned out
  to be content gaps instead). Also, several sets fixed this pass have
  bonuses left as documented no-ops because the underlying spell/mechanic
  isn't simulated at all in this fork (e.g. Feral Tank abilities, Hibernate,
  Life Tap's cooldown) — those are correctly *not* bugs, but would need
  actual new sim features, not a set-bonus fix, if ever wanted.
- **T3-looking sets (Dreadnaught, Cryptstalker, etc.) aren't pinned to a
  phase.** They fall through the normal location rule, which puts them at
  Phase 1 since they have no AtlasLoot raid-instance source on this server.
  Fine as long as the server doesn't actually raid that content — worth
  revisiting only if it turns out to.
- **Memory of Hyjal's (item 81015, rank V protection scroll) "reduces all
  damage received by up to 14" clause isn't modeled** — only its Armor +420
  is. There's no flat per-hit incoming-damage-reduction primitive anywhere in
  this engine (only multiplicative `PseudoStats.DamageTakenMultiplier`), same
  category of gap as the still-unimplemented Protection Potion absorb
  shields (see the "genuinely blocked" section above). Low-impact (14 damage
  per hit is minor against raid-boss hits) but noted rather than silently
  dropped. See `docs/CHANGES.md` Part AE.

## Raised, never actually answered

- **Scarlet Monastery set completeness.** Early on you asked me to check
  whether the Scarlet Crusade set has all its level-60 pieces, and said you
  could supply the last piece yourself if not. I don't have a record of ever
  reporting back on that. Worth re-checking the 5-piece Scarlet Crusade set in
  the current Phase-6 (SM) item list and asking you for the missing piece if
  one's still short.

## Scope limits I stated out loud

- **AtlasLoot-based loot sources only cover dungeons/raids/world bosses.**
  `gen_sources.py` explicitly leaves `Crafting/`, `Factions/`, and `PvP/`
  sources on whatever the generic upstream (retail) data already had, because
  matching those confidently against real faction/spell ids felt like a
  separate, riskier piece of work. If you want those corrected too, that's a
  distinct follow-up, not something quietly finished.

## Minor / cosmetic

- **One world-boss table's display name is ugly.** `ASpiritA` (an AtlasLoot
  world-boss table) isn't in the hand-verified `WB_NAMES` alias list, so any
  item sourced only from it falls back to an auto-prettified "A Spirit A"
  instead of a real boss name. Low-stakes (I don't know what boss this
  actually is without more digging) but noted rather than silently left ugly.
- **A stale line in the pipeline doc.** `docs/private-server-item-rules.md`'s
  "Known implementation wrinkles" section still says `gen_phases.py` "only
  scans `Instances/`" — that was true before the `serverdata.py` rewrite, but
  `gen_phases.py` now goes through `serverdata.atlasloot()` and already covers
  all seven AtlasLoot folders. The prose just never got updated to say so.

## Housekeeping, not content

- **Nothing from this work is committed.** `git log` still shows the same last
  commit as when this all started; `git status` has 216 modified/new files
  covering everything from this session (item pipeline rewrite, Int→Spell
  Power, script move, determinism fixes) plus the earlier talent-tree work.
  It's all just sitting in the working tree. Say the word when you want it
  committed (and how you want it split up, if at all).
- **Stray scratch files at the repo root**, not created by me and not
  referenced by anything built this session: `_add_candidates.json`,
  `_missing_vplus_items.txt`, `_olddb.json`, `_pvp_obsolete.json`,
  `_removelist.json`, `_removelist_aqnaxx.json`, `"Start Server.bat"`. Flagged
  before, still there — delete or keep, your call.

## Sidebar stat breakdown: derived stats are counted as "Gear"

- The Base/Gear/Talents/Buffs/Consumes columns in the left sidebar
  (`ui/core/components/character_stats.tsx`) are built from cumulative
  snapshots in `applyAllEffects` (`sim/core/character.go`), and every snapshot
  runs the stat dependencies. So Agility on a helm shows up as armor (2 per
  Agility), attack power, crit and dodge in the *Gear* column. Example: Helm of
  Endless Rage shows 691 gear armor (643 item + 24 Agility x 2) against the
  tooltip's 643. The total is correct and Toughness only scales the item's own
  armor, so this is a display issue only.
- Proper fix: a separate "from attributes" column for derived stats. That needs
  a new field per phase in `PlayerStats` (proto + regenerated pb.go), a second
  measurement without dependencies in `measureStats`, and the UI column. Moving
  the armor to Base as a stopgap was rejected: Base would then hold armor from
  gear and buffed Agility.
- **Stopgap in place:** `character_stats.tsx` moves the armor from gear Agility
  (2 per Agility, hard-coded to match `character.go`) from the Gear column to
  the Base column. Only armor, only the gear phase. Base therefore includes
  gear-Agility armor until the derived column exists; remove the stopgap when
  that lands.

## Priest talents still not implemented (2026-09-20)

Inner Fire and a no-proc Spirit Tap are done (Part AR). Still to do: Blackout, Wand Specialization, Blur, Focused
Casting, Improved Psychic Scream / Shadow Word: Silence / Shadow Word: Numb, Pilgrimage, Martyrdom, Stratagem,
Improved Dispel Magic, Insanity, and all healing talents (Spiritual Healing, Improved Healing, Improved Renew,
Improved Power Word: Shield, Improved Prayer of Healing, Holy Focus, Holy Reach, Holy Nova, Light's Grace, Blessed
Recovery, Spirit of Redemption, Holy Link), which need the healing spells enabled first. Spirit Tap also needs a
kill trigger with the 50%/100% per-rank proc chance.

## Warlock, open after the second audit (2026-09-20)
- Improved Drain Soul: the Health/Mana regen on kill (stacks to 5) is not modeled.
- Suppression is not in the sidebar (only Affliction spells, unlike the school-based Spell Hit tooltip).
- Bring the Pain (+5%/rank crit on 4 spells) is not in the sidebar.

## Druid spells not in the sim, cooldowns confirmed in game (2026-09-20)
Nature's Swiftness 5 min (`registerNaturesSwiftnessCD` is commented out in `sim/druid/talents.go`), Tranquility 2 min,
Bash 1 min, Frenzied Regeneration 5 min (`sim/druid/_frenzied_regeneration.go` is disabled), Rebirth 30 min.
