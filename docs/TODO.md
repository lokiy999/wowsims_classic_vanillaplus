# Open items

Things either of us raised across this work that haven't been done yet. Not a
backlog of ideas — only things actually said and left open. Grouped by how they
came up, most concrete first.

_As of 2026-09-19._

## Questions for Lokiy (collected 2026-09-24)

Things only you can answer; everything else I keep working on. Newest at the bottom.

1. **Talent builds** (LOW PRIORITY: Lokiy 2026-09-24, one of the last tasks, do near the end). Every spec's preset (and test) has an empty talent string, so the sim runs every class with no
   talents: Bloodthirst, Mortal Strike, Stormstrike, Aimed Shot, Shadowform, Mind Flay, Vampiric Embrace, Inner Focus,
   Adrenaline Rush, Holy Shield, Shadowburn, Demonic Sacrifice, Amplify Curse, Elemental Mastery etc. are never cast,
   and passive talents do nothing. Send a talent string (the sim's talent calculator export) for each spec you want as
   a preset, or tell me to draft builds from the custom trees myself.
2. **Pyroblast / Flame Shock / Insect Swarm / Arcane Missiles spell power**: cast each on a dummy with and without
   known +spell damage and send the tick numbers (combat log). The new tick counts use the classic total for now.
3. ~~**Arcane Missiles "energize"**~~ ANSWERED 2026-09-24: once per cast, 5 stacks max (as implemented). (+1% Arcane crit for 20 sec, stacks 5): does one cast give one stack, or does every
   missile give one?
4. **Trinkets**: Arcanite Dragonling and Cannonball Runner (tooltip + what the pet does), Six Demon Bag (tooltip and
   which effects it rolls), Orb of Chaotic Elements (how much it heals).
5. **Scarlet Monastery sets**: are all pieces in the game?
6. **`ASpiritA` world boss**: what name should it show?
7. **Scratch files at the sim repo root**: keep or delete?
8. ~~**Mana Tide Totem**~~ ANSWERED 2026-09-24: yes, trainable; now castable by the shaman (Part BV).: it is not in your shaman talent tree. Can a shaman learn it as a normal spell (from a
   trainer or a book)? If yes, the sim can let the shaman cast it.
9. **Execute**: the server text converts extra rage into "$*10;F1" damage per point, which I can't read from the
   data. The sim uses 15 per rage (classic). What does the tooltip say?
10. **Trap tick timing**: Immolation Trap (21 sec) and Explosive Trap (10 sec): how often do they tick?
11. **Illumination** (paladin): the tooltips for ranks 2-4 show the rank 5 text. What % of the mana cost do ranks
    2, 3 and 4 return? (The sim uses 27.5 / 35 / 42.5%.)
12. **Hunter Thrill of the Hunt**, **Shaman Armaments of Storm / Shamanism**: the server text has no proc chance
    ("a chance", "up to 300"). If you know the numbers, they can go in.
13. **More DBC tables**: if you can export `SpellCastTimes.csv` (and `SpellCooldowns`/`SpellCategory` if your
    tool has them) the same way as `Spell.csv`, I can check every cast time and cooldown against the sim too.
14. ~~**Cooldowns in Spell.csv**~~ ANSWERED 2026-09-24: correct, applied (Part BZ).: column 20 looks like the cooldown (Mana Tide 600000 = 10 min). If so, the server
    changed e.g. **Fire Blast to 20 sec** (classic 8), **Pyroblast to 1 min** (classic none) and the **shocks to
    10 sec** (classic 6). Can you check Fire Blast / Pyroblast / Flame Shock cooldowns in your spellbook? If they
    match, I will apply every cooldown from that column.
15. ~~**Item list resync**~~ ANSWERED 2026-09-24: follow the dump, never Naxx, rename ok; done (Part BY).: a full item pipeline run on the current dump (stats unchanged) would remove 349 items
    (mostly low-level crafted and dungeon gear, e.g. Rhahk'Zor's Hammer, Thornspike, Fine Leather Boots) and add 95
    (e.g. Headchopper, Rockfist, the Naxxramas T3 Dreadnaught/Redemption pieces), and rename the Cenarion set to
    "Cenarion Raiment". Do you want the sim to follow the dump (and keep or drop the Naxx items)? Until then the DB
    stays as it is.

## Talents not in the sim code (regenerated 2026-09-24, after Part CE)

Every talent field that no Go code reads (`Talents.<Name>` never used under `sim/`). This replaces the per-class
"not modeled" lists in the older dated sections below, which are stale: about 70 talents listed there were implemented
later (Part BS and others). Regenerate with a grep of `Talents\.` in `sim/<class>/` against `ui/core/talents/trees/<class>.json`
(per class: some names like Survival Instincts and Cold Blood exist in two classes).

Most of these are utility or PvP (stuns, fears, range, movement, stealth, pet utility, healing-only) and don't change a
DPS sim. Bosses can't be stunned, so stun-based bonuses (Improved Kidney Shot, Shock and Awe, Seal of Command's stunned
judgement) are left out on purpose. Done on 2026-09-24 (Part CE): Defiler, Inevitable Doom, Feral Instinct (Tiger's
Fury part), Death Mark, Seal of Command (now needs the talent); Divine Might is covered by the buff options (see below).
Still open for a DPS sim:
- **Divine Might** (paladin): not read by the sim code, but the Blessing of Might "improved" option (+20%) and the
  Blessing of Kings 12%/13% options are exactly Divine Might 2/2 (labels now say so). Rank 1 (+10%) has no option.
- **Improved Hunter's Mark**: covered by the "improved" Hunter's Mark debuff option (+20%/rank, 2/2 = +40%);
  the hunter doesn't cast the mark itself, so the talent field itself stays unused.
- **Wand Specialization** (mage, priest): wands aren't used by any rotation.
- Needs numbers: Shaman Armaments of Storm (25% chance, "up to 300", per hit or per imbue?), Shamanism (chance),
  Bloodlust (cooldown), Aftershock (cooldown, what it consumes), Hunter Thrill of the Hunt (chance), Mage Ice Shards
  (cooldown and spell power coefficient; Cold Grip depends on it), Warlock Withering Shroud and Death and Decay
  (cooldown, AoE only).

**PRIORITY: threat, tank and defensive talents to add or check later** (the user wants these done as one batch; the
feral tank rotation is off, so bear talents wait for Bear Form):
- Druid: Feral Instinct Bear Form threat +5/10/15%; Primal Tenacity and Improved Enrage (Enrage isn't in the sim);
  Custody of the Nature (crit immunity after being crit).
- Warrior: Mocker (Taunt/Challenging Shout/Mocking Blow hit +6%), Improved Defensive Stance (absorb shield on stance
  swap), Concussion Blow and Shield Toss ("high amount of threat": the threat value and Shield Toss attack power
  coefficient are unknown), Piercing Howl threat, Blood Craze, Berserker's Blood (needs the player's health).
- Paladin: Shield of Faith (-15% spell damage taken), Guardian's Favor (Salvation/Sanctuary +20%), Second Wind,
  Stoicism, Eye for an Eye, Improved Concentration Aura.
- Warlock: Damned Vanguard (demon threat +30%, damage taken -30%).
- Hunter: Intimidation threat, Improved Distracting Shot, Deterrence, pet Survival Instincts and Bestial Swiftness.
- Others: Mage Cryo Core, Priest Blur and Martyrdom, Rogue Survivor and Heightened Senses, Shaman Earth Shield,
  Earthquake threat and Tidal Barrier.
- To check in game: Slam flat threat 140, Rockbiter Weapon threat, Shen'dralar Badge of Deterrence and threat-reducing
  trinket effects (Fetish of the Sand Reaver, Grace of Earth, Two-Faced Medallion use effects).

- **Druid** (18 of 59): Cycle Of Life, Mighty Roots, Hurricane, Starfall, Primal Tenacity, Brutal Impact, Feral Charge, Leap, Untamed Heart, Natures Focus, Naturalist, Improved Enrage, Custody Of The Nature, Improved Rejuvenation, Tranquil Spirit, Swiftmend, Catharsis, Improved Regrowth
- **Hunter** (20 of 60): Improved Mend Pet, Aspect Mastery, Improved Revive Pet, Intimidation, Bestial Swiftness, Survival Instincts, Team Play, Terrifying Roar, Improved Hunters Mark, Improved Concussive Shot, Vantage Point, Hawk Eye, Improved Distracting Shot, Scatter Shot, Deterrence, Improved Wing Clip, Trapper, Thrill Of The Hunt, Deep Freeze, Wyvern Sting
- **Mage** (21 of 60): Wand Specialization, Practical Defensive Magic, Practical Offensive Magic, Flame Throwing, Impact, Blazing Speed, Chain Reaction, Frost Warding, Permafrost, Cryo Core, Frostbite, Improved Frost Nova, Cold Blood, Ice Block, Arctic Reach, Shatter, Cold Grip, Ice Mirror, Ice Shards, Ice Barrier, Advanced Ice Shielding
- **Paladin** (18 of 60): Aura Mastery, Spiritual Focus, Shield Of Faith, Inner Light, Divine Grace, Unyielding Faith, Lights Mercy, Holy Purge, Improved Hammer Of Justice, Guardians Favor, Dominance, Second Wind, Improved Concentration Aura, Stoicism, Pursuit Of Justice, Divine Might, Eye For An Eye, Repentance
- **Priest** (20 of 60): Pilgrimage, Wand Specialization, Martyrdom, Improved Dispel Magic, Focused Casting, Stratagem, Holy Focus, Lights Grace, Blessed Recovery, Holy Nova, Holy Reach, Improved Prayer Of Healing, Spirit Of Redemption, Holy Link, Blackout, Improved Psychic Scream, Shadow Word Numb, Improved Shadow Word Silence, Blur, Insanity
- **Rogue** (19 of 60): Remorseless Attacks, Seek And Destroy, Vitality, Improved Kidney Shot, Total Control, Improved Gouge, Improved Sprint, Improved Kick, Dazing Bolts, Survivor, Improved Sap, Master Of Deception, Thug Life, Setup, Camouflage, Heightened Senses, Shadow Cut, Cloak Of Shadows, Enveloping Shadows
- **Shaman** (21 of 60): Storm Reach, Earths Grasp, Sand Blast, Eye Of The Storm, Earthquake, Earth Shield, Aftershock, Improved Ghost Wolf, Bloodlust, Armaments Of Storm, Shamanism, Spiritwalking, Improved Healing Wave, Totemic Mastery, Tidal Barrier, Nature Focus, Ancestral Healing, Focused Mind, Healing Way, Meditation, Cleansing Wave
- **Warlock** (19 of 60): Fel Concentration, Jinx, Dread, Black Speech, Herald Of Woe, Withering Shroud, Death And Decay, Improved Healthstone, Improved Health Funnel, Improved Voidwalker, Master Conjuror, Damned Vanguard, Improved Felhunter, Improved Enslave Demon, Fel Pact, Aftermath, Feeding Demons, Pyroclasm, Shock And Awe
- **Warrior** (12 of 60): Improved Charge, Improved Hamstring, Improved Rend, Combat Endurance, Piercing Howl, Blood Craze, Berserkers Blood, Mocker, Improved Defensive Stance, Concussion Blow, Iron Will, Shield Toss

## Open items at a glance (updated 2026-09-24)

One list of everything still open; the details are in the dated sections below. Items finished on 2026-09-23 are
listed at the end of this overview so the older sections don't need rewriting.

**Affects sim results**
- Spell values: direct and periodic damage in the per-rank arrays now match the server (Part BO). Not compared yet:
  heals and shields (not in the sim). DoTs, missiles, abilities and pet spells done in Parts BP, BR and BT. Pyroblast, Flame Shock, Insect Swarm and Arcane
  Missiles spell power coefficients are a guess (classic total spread over the new tick count); Arcane Missiles'
  +1% Arcane crit is one stack per cast (could be per missile).
- Talents not modeled yet: the damage/threat/mana ones with clear numbers were done in Part BS. Left: utility and
  PvP talents (stuns, fears, movement, range), healing talents (no healing spells in the sim), and ones needing a new
  mechanic (freeze for Shatter/Frostbite/Deep Freeze, target health for Coup de Grace-style effects,
  Aftershock, Chain Reaction, Ice Shards, Withering Shroud, Death and Decay, Earthquake, Improved Rend stacks,
  Berserker's Blood). The per-class lists in the 2026-09-19/20 sections are partly stale.
- Paladin ret/prot rotations are new and basic (written during the SoD removal); tune in game.
- No preset uses Arcane Missiles or Insect Swarm, so the tests don't cover them (checked in the browser instead).
- Talent presets are all empty (question 1 at the top; low priority, one of the last tasks).
- Healing: priest healing spells are disabled; druid Nature's Swiftness, Tranquility, Bash, Frenzied Regeneration and
  Rebirth are not in the sim; the shaman casting Mana Tide Totem is not modeled.
- Shaman totem and weapon imbue items (2026-09-20 sections).
- Consumables the user chose to skip (Part P): still absent if ever wanted.

**Needs data from the game**
- Trinkets: Arcanite Dragonling and Cannonball Runner (pet damage), Six Demon Bag (effects); Orb of Chaotic Elements
  heal amount is assumed equal to its damage.
- Talent and set values marked "confirm in game" in the class sections.
- Scarlet Monastery set completeness (never answered).

**Data / pipeline**
- ~~Item database resync~~ done 2026-09-24 (Part BY).
- Crafted items are all Phase 1 (deferred by the user).
- Local spell names that differ from the server (list in the 2026-09-23 tooltip audit section).
- ~~"Level 60" target uses a SoD NPC id; BWL encounter mechanics copied from SoD~~ done 2026-09-24 (Part CB).
- Unverified classic values (not in the server data): Slam flat threat 140, Lightforge Armor 6-piece proc chance 6%.

**Cosmetic / housekeeping**
- ~~Sidebar: armor from gear Agility is shown under Base (stopgap)~~ fixed 2026-09-24 (Part BW).
- `ASpiritA` world boss name; stray scratch files at the repo root (keep or delete, user's call).

**Done 2026-09-23** (older sections may still mention these): SoD content removed or replaced (Part BI); trinkets
(Parts BJ-BM, only the items listed above remain); boss crit only reduced by Defense (Part BK); Memory of Hyjal flat
damage reduction, Greater Protection Potions (absorb shields), The Black Book pet damage reduction, Presence of Might
+10 all stats, Law of Nature removed (Part BN); the committed-work and pipeline-doc notes below.

## Raised 2026-09-19 — after the talent-calculator cross-check (CHANGES.md Part AP)

Still open, the calculator confirmed the text but not the mechanics:
- Priest **Power Word: Requital** (implemented 2026-09-19, but coefficient is a placeholder 1.5/3.5 and ranks 2-4 lack DB name/icon; original note): instant, 150 mana, 20s cooldown, 185.5 Holy damage (346-371 if the target is
  feared/stunned/incapacitated). Not implemented: spell power coefficient unknown, and it needs an APL action.
- ~~Warrior **Enrage**:~~ DONE 2026-09-19 (see CHANGES.md Part AP follow-up 2). Old note: "1% melee damage per stack for 15s, up to 10 stacks" at rank 1. Code still uses a flat
  5%/rank with 12 swing-counting stacks. Waiting on rank 2+ text.
- ~~Warrior **Tactical Mastery**:~~ DONE 2026-09-19. Old note: "retain an additional 10 Rage" at rank 1 (code 5/rank).
- ~~Mage **Mind Mastery**: also +20%/rank Arcane Intellect effect~~ done (`sim/mage/mage.go`), confirmed 2026-09-24.
- Not modeled: Shaman Armaments of Storm (5%/rank, up to 300 Nature, level scaled), Shamanism, Aftershock;
  Rogue Improved Sinister Strike proc, Coup de Grace, Gaining an Advantage, Brigandage; Hunter Find Weakness,
  Deadeye, Thrill of the Hunt; Warlock Defiler, Prolonged Misery, Demonic Embrace regen, Demonic Onslaught (pet crit, the DBC effect is 20% but the per-rank scaling is unclear); Druid Killer Instincts
  attack speed part; Warrior Cleaving (Thunder Clap/Whirlwind), Improved Execute.

## Raised 2026-09-19 — warrior talents (CHANGES.md Part AO)

- Warrior audit round 2 (CHANGES.md Part BB), still open and needs the game: Maim (10% on auto attacks per the DBC, +5%
  damage taken, duration unknown, all three ranks look identical); Improved Rend (stack behaviour: damage per stack,
  duration); Deep Wounds duration and stacks (sim: 4 ticks of 3s, no stacks); Shield Block cooldown (DBC 20s on rank 1,
  5s on rank 2, sim 5s); Berserker's Blood (needs current health, 1-40% speed); Improved Berserker Stance GCD reduction
  (0.25s/rank); Improved Hamstring, Improved Charge; Butterfly Style rage part; Execute rage-to-damage ratio (sim 15).
- Para Bellum applies to every warrior spell with a cooldown and a spell code; verify it
  does not wrongly shorten stance-change or shared-cooldown timers.

## Raised 2026-09-19 — rogue talents (CHANGES.md Part AN)

- Not modeled: Improved Sinister Strike extra-hit proc (3/5%), Coup de Grace (+5%/rank
  damage below 20% health), Bloodthirsty (Garrote/Rupture +10%/rank damage, shorter ticks),
  Exhaustion, Combat Rush, Brigandage, Gaining an Advantage, Dazing Bolts, Survivor,
  Physical Prowess, Improved Kidney Shot, Remorseless Attacks, Weapon Expertise values.
- Connivery is applied to all melee damage; DBC says attacks from behind only.
- Rogue audit round 2 (CHANGES.md Part BA) open: Exhaustion on Rupture and Expose Armor (Slice and Dice is done),
  Remorseless Attacks (needs kills), Improved Kidney Shot (no Kidney Shot in the sim), Survivor, Sprint (Physical Prowess
  cooldown part, Improved Sprint), Thistle Tea cooldown not in the DBC, Brigandage damage may need a coefficient check.

## Raised 2026-09-19 — hunter talents (CHANGES.md Part AM)

- Not modeled: Melee Specialization (+30% melee speed, -30% ranged), Dual Wield
  Specialization (OH +20-50%), Weapon Expertise (extra arrow), Find Weakness, Thrill of the
  Hunt, Deadeye, Improved Hunter's Mark, Stalking, Vantage Point, Spirit Bond, Kill Command,
  Improved Tracking, Deep Freeze, Savage Blow, Whirling Axe, most pet utility talents.
- Reconnaissance is coded 3%/rank all damage; DBC spell 34063 (3%) may be flat at every
  rank. Unchecked.
- Hunter audit round 2 (CHANGES.md Part AZ) still open: Thrill of the Hunt (proc chance unknown, restores 3 x level mana);
  Kill Command mana cost (assumed free, check in game); Savage Blow damage (placeholder: one main-hand and one off-hand
  weapon hit), range and the Hawk/Cheetah/Wild effects; Whirling Axe cooldown, mana cost and range (none set), slow and
  interrupt not modeled; Aspect of the Beast/Pack stat effects not applied (Monkey is done); Aspect Mastery (later); Hunter's Mark melee AP (+90) not modeled; Team Play, Stalking's kill bonus,
  Deep Freeze, Vantage Point not modeled; shot damage and mana tables by rank not compared; Lethal Shots and Weapon
  Expertise crossbow crit are shown in the Melee Crit tooltip only.

## Raised 2026-09-19 — paladin talents (CHANGES.md Part AL)

- Not modeled (2026-09-20): Divine Grace (needs Seal/Judgement of Light and Wisdom, not in the sim), Illumination (the
  server text looks odd: 20% of the base mana cost back on a crit at rank 1, a chance for 50% at rank 5, ranks 2-4
  unknown; user wants to look at it later), Light's Mercy (Flash of Light is not in the sim), Holy Purge, Eye for an Eye
  (5%/10% of spell damage taken reflected), Repentance, Codex Holy Light cost/cast time, the Judgement of Fury forced
  attack. Done: Divine Concentration, Improved Purifying, Holy Grasp, Blessed Strikes threat, Sanctity Aura talents,
  Seal of Fury, Improved Retribution Aura, Vengeance confirmed (CHANGES.md Parts BC and BE).
- Seal of Command: the 1s internal cooldown is unverified (damage 50%, 12 procs per minute, 120s duration and top rank Judgement of Command 441-475 are confirmed; lower Judgement of Command ranks are still the old values).
- Improved Lay on Hands cooldown checked: DBC is -15 min/rank, matches the code.
- Paladin audit round 2 (CHANGES.md Part AY): Holy Shock, Hammer of Wrath, Righteous Fury, Redoubt and Holy Shield are
  now confirmed. Still open: Consecration and Exorcism rank/damage not compared, Holy Wrath / Avenging Wrath / Divine Favor
  cooldowns not in the DBC, Holy Shield coefficient, Holy Shock healing.

## Raised 2026-09-19 — shaman talents (CHANGES.md Part AK)

- Elemental Devastation: DBC aura 30165 gives a flat 10% crit at every rank, code uses 3%/rank.
  Looks odd, so not changed; confirm on the server.
- Not modeled: Static Field, Armaments of Storm, Shamanism, Aftershock, Earthquake, Earth
  Shield, Primal Endurance, Guardian Totems (partly in core), Rockhide, Improved Fire Totems,
  Improved Ghost Wolf, Blood Lust, Elemental Mastery values unchecked, all healing talents
  (Healing Way, Purification damage part, Tidal Mastery speed part), Nature Grace/Guardian.
- ~~Healing Stream Totem base healing formula~~ fixed 2026-09-20 (CHANGES.md Part AW).

## Raised 2026-09-19 — druid talents (CHANGES.md Part AJ)

- ~~**Improved Mark of the Wild conflict:**~~ RESOLVED 2026-09-19 (20%/rank confirmed by user). DBC gives 20/40/60% (3 ranks), but the buff audit
  above expects the improved Gift of the Wild armor to be 285x1.35. One of them is wrong;
  not touched. Check which value is actually live on the server.
- Not modeled: Starlight Wrath/etc. are fine, but Starfall, Power of Nature, Dreamstate,
  Mighty Roots, Hurricane, Cycle of Life, Unity with Nature, Stalking, Predatory Strikes
  (custom DBC values look malformed), Primal Fury, Survival Instincts, Leader of the Pack
  crit value (DBC 3%), Furor values, all healing talents (Improved Rejuvenation/Regrowth,
  Gift of Nature healing part, Tranquil Spirit), Swiftbloom, Naturalist.

## Raised 2026-09-19 — mage talents (CHANGES.md Part AI)

- Not modeled: Shatter (needs a frozen-target state), Spell Twisting, Impact, Blazing
  Speed, Chain Reaction, Thermal Expansion, Cryo Core, Cold Grip, Rimebound (needs
  spell 34266), Advanced Ice Shielding, Frost Warding/Fire Warding, Practical talents,
  Brilliance Aura, Magic Absorption mana-on-resist, Arcane Intellect +%/rank from Mind Mastery.
- Hot Streak/Pyromania untested in a fire rotation (default mage APL is not fire).

## Raised 2026-09-19 — warlock talents (CHANGES.md Part AH)

- Not modeled: Defiler (-20% tick time/duration/GCD), Prolonged Misery (+2/4/6s dot
  duration, not a multiple of the 3s tick), Jinx, Improved Drain Soul (DBC: Drain
  Soul cooldown -5/-10s + mana/health on kill; the cooldown and the wrong Curse of Doom use were fixed 2026-09-20), Inevitable Doom, Herald of Woe, Sadism, Feeding
  Demons, Fel Pact, Demonic Onslaught, Improved Immolate (2 stacks; current 5% bonus
  is a guess), Pyroclasm/Mayhem/Shock and Awe, Demonic Embrace regen, Soul Link
  damage split (redirect is 20% per DBC; code uses a flat multiplier).
- Master Demonologist Felhunter resist and Voidwalker values were assumed to scale
  3/rank and 2/rank from rank-1 DBC only; ranks 2-5 of the sub-spells not checked.

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

- **Done 2026-09-20 (CHANGES.md Part BC):** Devotion Aura 700 and +25%/point, Commanding Shout and Horn of Lordaeron removed,
  `IncludeAQ` on, faction gating dropped, Moonkin Aura +5% spell damage and healing, Trueshot Aura 50 melee AP.
- **Still open from this list:** ~~Libram of Truth toggle~~ done 2026-09-24 (Part CC). Resistance totems for an outside shaman still have no picker
  (Guardian Totems from a shaman in the sim works). Sanctity Aura and the resistance auras were done (CHANGES.md Part BC).
- ~~Arcane Intellect~~ done 2026-09-20 (30/37/60/67 picker, CHANGES.md Part AS).
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

## Raised 2026-09-18 — pre-existing `go test ./sim/...` failures: RESOLVED 2026-09-20

All `.results` files were regenerated during the second audit round; `go test ./sim/...` passes with and without `--tags=with_db`.

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

- ~~**Nothing from this work is committed.**~~ Resolved: everything is committed and pushed. Old note: `git log` still shows the same last
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

## Sidebar stat breakdown: derived stats are counted as "Gear" (FIXED 2026-09-24, Part BW)

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

## Shaman totems, still open (2026-09-20)
- Mana Spring, Healing Stream and Windfury Totem dropped from the APL still use the hard-coded raid buff values
  (see the TODO comments in `sim/shaman/air_totems.go` and `water_totems.go`).
- Flametongue Totem is only the top rank (58+, +40 spell damage, 12.17 fire damage per second of weapon speed) and only
  for non-shaman classes through the weapon imbue option; there is no Flametongue Totem spell for the shaman to cast.
- Rockbiter Weapon healing at level 50 (rank 6) and the Frostbrand damage of ranks 2 to 4 are not in the server data lookup
  (only ranks 1 and 5 were found); the sim only uses Frostbrand rank 5.
- Rockbiter Weapon extra threat: the sim keeps the old per-rank bonus, the real value is unknown.
- Windfury Weapon keeps a 1.5s internal cooldown between procs: test in game whether that is right.
- Frostbrand's movement speed slow is not modeled (only the attack speed slow).
- Weapon enchants last 5 minutes: the Flametongue, Windfury and Frostbrand procs now stop after 5 minutes, but the stats
  from Flametongue (spell damage) and Rockbiter (Strength, healing) are permanent for the whole fight.
- Mana Tide Totem is only a party buff (100 mana per second for 15s, 10 min cooldown); a shaman casting it (200 mana) is
  not modeled (`registerManaTideTotemCD` in `sim/shaman/talents.go` is commented out).
- Resistance totems get their Guardian Totems bonus (60 -> 90) only from a shaman in the sim; there is no picker in the
  buff settings for an outside shaman. Same for Improved Weapon Totems.
- A shaman dropping his own Mana Spring Totem does not use Restorative Totems (the raid buff "Improved" option is +50%).

## Fire totems (2026-09-20)
The fire totems are in the sim and match the server data at level 60; decide whether anything is missing.
- Searing Totem rank 6: 40-54 Fire damage per attack, spell coefficient 0.083, 170 mana, lasts 55s.
- Magma Totem rank 4 (level 56): 75 area Fire damage per pulse, coefficient 0.033, 650 mana.
- Fire Nova Totem rank 5 (level 52): 413-459 area Fire damage, coefficient 0.143, 520 mana, 15s cooldown.
- Call of Flame (+10%/rank damage) and Elemental Fury (crit damage) already apply; ~~Improved Fire Totems~~ done 2026-09-24 (Part BR).

## Raised 2026-09-23 — spell audit (CHANGES.md Part BG)
- ~~**The Black Book**~~ done 2026-09-23 (Part BN). Old note: the server tooltip says pet damage +100% and pet damage taken -100% for 30s; the sim gives
  +100% pet armor instead of the damage reduction. Pet damage is right, so DPS is unaffected. The sim's 5 min cooldown is
  not in the server tooltip; confirm in game.

## Raised 2026-09-23 — trinkets that do nothing in the sim (CHANGES.md Part BH)

204 trinkets are in the item database; 38 have code (`core.NewItemEffect`). Of the other 166, 39 only give plain
stats (those work without code) and 5 have no effect text. The rest have a Use/Equip effect with no code, so
equipping them does nothing beyond their stats. Effect text is from `VPlusItemDB.lua`.

**Could matter for a DPS/tank sim** (31 Use, 28 Equip):

- Use:
  - Aegis of Preservation (19345): Use: Decreases damage taken by 10%, and heals for 30% of damage taken for 20 sec.
  - Arcanite Dragonling (16022): Use: Activates your Arcanite Dragonling to fight for you for 1 min. It requires an hour to cool down before it can be used again.
  - Banner of Challenge (26340): Use: Forces the target to attack you for 5 sec.
  - Blazing Emblem (2802): Use: Increases Fire resistance by 50 and reduces all Fire damage taken by up to 25 for 15 sec.
  - Blessed Prayer Beads (19990): Use: Increases healing done by spells and effects by up to 190 for 20 sec.
  - Blood Scarred Scale (83075): Use: Increases damage and healing done by magical spells and effects by up to 25 and all resistances by 17 for 30 sec.
  - Cannonball Runner (13382): Use: Summons a cannon that will fire at enemies in front of it for 20 sec.
  - Chained Essence of Eranikus (10455): Use: Poisons all enemies in an 8 yard radius around the caster. Victims of the poison suffer 50 Nature damage every 5 sec for 45 sec.
  - Dark Iron Bookmark (80011): Use: Blasts the enemy for 168 to 202 Fire damage.
  - Everlasting Liver (81041): Use: Restores 3% of your total Health every 3 sec and increases your Spirit by 40. Lasts 30 sec.
  - Fetish of Chitinous Spikes (21488): Use: Spikes sprout from you causing 82 Nature damage to attackers when hit. Lasts 30 sec.
  - Fetish of the Sand Reaver (21647): Use: Reduces the threat you generate by 70% for 20 sec. / Equip: Reduces the threat you generate by 5%.
  - Grace of Earth (21181): Use: Reduces your threat to enemy targets within 30 yards, making them less likely to attack you. / Equip: Reduces the threat you generate by 5%.
  - Gri'lek's Charm of Valor (19952): Use: Increases the critical hit chance of Holy spells and physical attacks by 10% for 30 sec.
  - Hazza'rah's Charm of Healing (19958): Use: Increases the Priest's casting speed by 40% for 15 sec.
  - Heart of the Scale (13164): Use: Increases Fire Resistance by 20 and deals 20 Fire damage to anyone who strikes you with a melee attack for 5 min.
  - Hibernation Crystal (20636): Use: Increases healing done by magical spells and effects by up to 350 for 15 sec.
  - Ironbark Tea Leaf (26070): Use: Instantly heals 500 damage. Also increases armor by 1500 and healing taken by 15% for 30 sec.
  - Mar'li's Eye (19930): Use: Restores 60 mana every 5 sec for 30 sec.
  - Mark of Bestial Fury (26312): Use: Increases melee and ranged attack power by 200 and increases damage done by magical spells and effects by up to 120 for 30 sec.
  - Orb of Chaotic Elements (26229): Use: Releases the power of wild elements, dealing 1 to 1000 Fire, Frost or Nature damage to enemy and healing you.
  - Petrified Scarab (21685): Use: Increases your spell resistances by 100 for 1 min. Every time a hostile spell lands on you, this bonus is reduced by 10 resistance.
  - Ragged John's Neverending Cup (15873): Use: Increases Stamina by 28 and reduces physical damage taken by 22 for 10 min. However, lowers your movement speed by 25%.
  - Ramstein's Lightning Bolts (13515): Use: Harness the power of lightning to strike down all enemies around you for 200 to 440 Nature damage.
  - Scarlet Battle Orders (26323): Use: Increases movement, attack and casting speed by 20% for 20 sec.
  - Shard of the Fallen Star (21891): Use: Calls down a meteor, burning all enemies within the area for 400 to 442 total Fire damage.
  - Six Demon Bag (7734): Use: Blasts enemies in front of you with the power of wind, fire, all that kind of thing!
  - Smokey's Lighter (13171): Use: Deals 125 Fire damage to all targets in a cone in front of the caster.
  - The Final Gaze (26328): Use: Greatly reduces the chance your attacks and spells will miss or be resisted for 10 sec. / Equip: Increases your stealth detection.
  - Two-Faced Medallion (26353): Use: Attempts to disguise you as the targeted dead creature, reducing the radius at which enemies will attack you. / Equip: Reduces the threat you ...
  - Wushoolay's Charm of Nature (19955): Use: Increases damage done and healing of your Nature spells by 20% for 15 sec.
- Equip:
  - Blue Mottled Egg (83006): Equip: +10 All Resistances.
  - Force of Will (11810): Equip: When struck in combat has a 1% chance of reducing all melee damage taken by 25 for 10 sec.
  - Mark of Thirst (26203): Equip: Increases your attack, casting and movement speed by 20% against targets below 20% health.
  - Mark of the Veteran (26159,26168): Equip: Increases the attack power granted by Battle Shout by 42.
  - Mark of the Veteran (26160,26163,26170,26172): Equip: Improves your critical strike chance for all attacks and spells by 2%.
  - Mark of the Veteran (26162,26171): Equip: Reduces the cooldown of your Multi-Shot by 2 sec.
  - Mark of the Veteran (26164,26173): Equip: Increases your Energy regeneration by 2 per tick.
  - Mark of the Veteran (26165,26174): Equip: Grants +5% increased spell hit chance for 20 sec when one of your spells is resisted.
  - Mark of the Veteran (26167,26176): Equip: Reduces the time between periodic ticks of your Corruption spell by 1 sec.
  - Onyx Egg (83003): Equip: Decreases damage taken by 1%.
  - Pimgib's Collar (18354): Equip: Increases the damage of your Imp's Firebolt spell by 18.
  - Pink Speckled Egg (83007): Equip: +10 All Resistances.
  - Reactive Auto-Recaster (26223): Equip: 4% chance to recast instantly the just casted spell.
  - Royal Seal of Eldre'Thalas (18466): Equip: +12 to all attributes.
  - Royal Seal of Eldre'Thalas (18471,18472): Equip: Reduces the cost of your spells by 2%.
  - Sawtooth Talisman (26212): Equip: Your attacks ignore 5% of your enemies' Armor. / Equip: Your attacks ignore 250 of your enemies' Armor.
  - Shen'dralar Badge of Deterrence (26094): Equip: Threat +5%
  - The Lion Horn of Stormwind (14557): Equip: Generates an aura that protects nearby party members by increasing their armor by 250 and magic resistances by 10.
  - Uther's Strength (11302): Equip: When you take damage has a 2% chance to protect you with a holy shield.

**Utility only, fine to leave** (62): Abyss Shard, Alchemists' Stone, Ankh of Life, Arcane Infused Gem, Arena Grand Master, Barov Peasant Caller, Chronomirage, Dalaran Spellshackles, Darkmoon Card: Twisting Nether, Defender of the Timbermaw, Defiler's Talisman, Dimensional Ripper - Everlook, Dog Whip, Elementium Echo Modulator, Enamored Water Spirit, Gnomish Universal Remote, Gyrofreeze Ice Reflector, Heart of Noxxion, Hederine Shackles, Hook of the Master Angler, Hyper-Radiant Flame Reflector, Insignia of the Alliance, Insignia of the Gladiator, Insignia of the Horde, Insignia of the Unfettered, Insignia of the Unfettered Champion, Leyguard Charm, Lifestone, Major Recombobulator, Minor Recombobulator, Orb of Deception, Personal Harm Prevention Field Emitter, Piccolo of the Flaming Fire, Scarlet Hound Whistle, Scarlet Triage Kit, Still Eye of the Watcher, Talisman of Arathor, The Wall's Chime, Ultra-Flash Shadow Reflector, Ultrasafe Transporter: Gadgetzan, Vial of Elune's Light.

Also: Shard of the Flame (17082) has no stats in the DB (server: 16 health per 5 sec, no DPS effect);
Scrolls of Blinding Light (19343) is not in the server data. ~~Royal Seal of Eldre'Thalas (18466) "+12 to all
attributes" and the Blue Mottled/Pink Speckled Egg "+10 All Resistances"~~ done 2026-09-24 (Part CC, plus 201 other
items with "Equip: +N All Resistances"). Onyx Egg "-1% damage taken" and the Royal Seal (18471/18472) "-2% spell cost" are already in
`sim/common/item_effects/vplus_trinkets.go` (checked 2026-09-24).

## Raised 2026-09-23 — tooltip/icon audit (CHANGES.md Part BH)

- ~~**Weapons coded as their Season of Discovery versions**~~ fixed in Part BI (checked 2026-09-24: Thunderstrike,
  Shadowstrike and Masterwork Stormhammer match the server tooltips). Old note: not the server's (proc spell ids are SoD, and the effect
  may be too): Flame Wrath (11809, server: Use: fire shield + 210-250 fire ring; sim: on-hit proc), Quel'Serrar
  (18348, server: +1% hit, Use: +40 defense/+1400 armor for 20s; sim: proc aura 463105), Shadowstrike (17074, server:
  chance to steal 100-180 life), Thunderstrike (17223, server: 150-250 Nature to up to 3 targets), Masterwork
  Stormhammer (12794, server: 110-200 Nature to up to 3 targets), Ravager aura id 433801. Check each against
  `VPlusItemDB.lua`. ~~Also SoD ids still in use: Mangle debuff, Primal Blessing set proc~~ (fixed in Part BI, confirmed 2026-09-24).
- ~~**Presence of Might**: server "+10 to all Stats"~~ already +10 to all five stats in `enchant_overrides.go` (checked 2026-09-24).
- **Enchant Shield - Law of Nature** (7603, item 228982) is a Season of Discovery enchant, not in the server data.
- ~~**Server spell values that differ from the sim**~~ done 2026-09-24 (Parts BO, BP, BR). Old note: e.g. Immolation Trap rank
  5: server 966 Fire over 21 sec, sim 690 over 15 sec (`sim/hunter/immolation_trap.go`); Arcane Shot 14286: server 212.
  The tooltips now show the server values; the sim numbers were not changed.
- **Local spell names that differ from the server, left as they are:** 713 Summon Incubus (server: Turn Undead),
  1098/11725/11726 Subjugate Demon (Enslave Demon), 10326 Turn Undead (Turn Evil), 13219 Wound Poison (Wound
  Poison I), 13948 Enchant Gloves - Minor Haste (Lesser Haste), 15487 Silence (Shadow Word: Silence), 28271/28272
  Polymorph (Polymorph: Turtle / Pig). The Arcanum names and "Judgement Armor 8-piece" are deliberate.
- 13 custom enchants (900101-900212) have no spell row or label; the gear slot shows their name, which reads fine.

## Raised 2026-09-23 — after the SoD removal (CHANGES.md Part BI)

- ~~**Flame Wrath** and **Quel'Serrar**~~ done 2026-09-23 (Part BM, cooldowns from the tooltips). Old note: "Use:" effects on the server (Flame Wrath: fire shield + 210-250 fire ring;
  Quel'Serrar: +40 defense and +1400 armor for 20 sec) but the server data has no cooldown for them, so the sim keeps
  the classic chance-on-hit versions. Needs the cooldowns from the game.
- The paladin rotations are new and simple (no Consecration for ret, no mana management); tune them in game.
- ~~The "Level 60" preset target uses NPC id 213336~~ and ~~BWL encounter mechanics copied from SoD~~: done 2026-09-24 (Part CB).
- The disabled feral tank and healing priest rotations use spell ids from a later expansion (48xxx), not SoD.
- ~~`tools/database/atlasloot.go` reads the AtlasLootClassic_SoD repository; SoD comments in core~~ done 2026-09-24 (Part CB).

## Trinkets, status after CHANGES.md Part BJ (2026-09-23)

The trinket list above is partly done. Still open:
- **Need the cooldown from the game** (custom server items, not on Wowhead): Mark of Bestial Fury (26312), Blood
  Scarred Scale (83075), Scarlet Battle Orders (26323), The Final Gaze (26328), Dark Iron Bookmark (80011), Orb of
  Chaotic Elements (26229), Everlasting Liver (81041), Ironbark Tea Leaf (26070), Banner of Challenge (26340); also
  Hibernation Crystal (20636, Wowhead shows no cooldown).
- **Need a new mechanic:** Arcanite Dragonling and Cannonball Runner (summons), Six Demon Bag (random effects),
  Reactive Auto-Recaster (spell recast), Force of Will and Ragged John's Cup (flat damage reduction), Uther's Strength
  and Aegis of Preservation's heal (absorb/heal), Sawtooth Talisman's 5% armor ignore, Petrified Scarab's decay.
- Use effects of Grace of Earth and Two-Faced Medallion (threat/aggro radius) are not modeled.
- Cooldowns assumed from classic where the server changed the effect: Gri'lek's, Wushoolay's and Hazza'rah's charms
  (3 min), Aegis of Preservation (5 min). Confirm in game.

## Trinkets, status after CHANGES.md Part BL (2026-09-23)

Done since Part BJ: Uther's Strength, Force of Will, Ragged John's Cup and Blazing Emblem damage reduction, Heart of
the Scale thorns, Petrified Scarab decay, Aegis of Preservation heal, Sawtooth Talisman 5%, Reactive Auto-Recaster.
Still open: the custom items that need a cooldown from the game (list above), Arcanite Dragonling and Cannonball Runner
(summons, no pet data), Six Demon Bag (random effects), Grace of Earth and Two-Faced Medallion use effects.
Other things now possible with `addFlatDamageReduction`: Memory of Hyjal's "reduces all damage received by up to 14"
(item 81015) and absorb shields such as the Protection potions.

## Trinkets, status after CHANGES.md Part BM (2026-09-23)

All custom-item cooldowns are in (from in-game tooltips). Still open: Arcanite Dragonling and Cannonball Runner
(summons, no pet data), Six Demon Bag (random effects), Grace of Earth and Two-Faced Medallion use effects, Banner of
Challenge (taunt). The Orb of Chaotic Elements heal is assumed to equal the damage dealt.
