# Open items

Things either of us raised that are still open. Finished, answered or decided items move to
[TODO_DONE.md](TODO_DONE.md) (Lokiy 2026-09-25); details of each change are in [CHANGES.md](CHANGES.md).

_Cleaned up 2026-09-25 (after Part DH)._

## Questions for Lokiy (collected 2026-09-24)

Things only you can answer; everything else I keep working on. Newest at the bottom. Numbers are kept as they were
(CHANGES.md refers to them); answered ones are in TODO_DONE.md.

1. **Talent builds** (LOW PRIORITY: Lokiy 2026-09-24, one of the last tasks, do near the end). Every spec's preset (and test) has an empty talent string, so the sim runs every class with no
   talents: Bloodthirst, Mortal Strike, Stormstrike, Aimed Shot, Shadowform, Mind Flay, Vampiric Embrace, Inner Focus,
   Adrenaline Rush, Holy Shield, Shadowburn, Demonic Sacrifice, Amplify Curse, Elemental Mastery etc. are never cast,
   and passive talents do nothing. Send a talent string (the sim's talent calculator export) for each spec you want as
   a preset, or tell me to draft builds from the custom trees myself.
2. **Pyroblast / Flame Shock / Insect Swarm / Arcane Missiles spell power**: cast each on a dummy with and without
   known +spell damage and send the tick numbers (combat log). The new tick counts use the classic total for now.
4. **Trinkets**: Arcanite Dragonling and Cannonball Runner (tooltip + what the pet does), Six Demon Bag (tooltip and
   which effects it rolls), Orb of Chaotic Elements (how much it heals).
5. **Scarlet Monastery sets**: are all pieces in the game? (Early request: check whether the 5-piece Scarlet Crusade
   set has all its level-60 pieces; you offered to supply a missing piece. Never reported back.)
6. **`ASpiritA` world boss**: what name should it show? (An AtlasLoot world-boss table not in the `WB_NAMES` list, so
   its items show "A Spirit A".)
7. **Scratch files at the sim repo root**: keep or delete? `_add_candidates.json`, `_missing_vplus_items.txt`,
   `_olddb.json`, `_pvp_obsolete.json`, `_removelist.json`, `_removelist_aqnaxx.json`, `"Start Server.bat"`.
9. **Execute**: the server text converts extra rage into "$*10;F1" damage per point, which I can't read from the
   data. The sim uses 15 per rage (classic). What does the tooltip say? (Same as question 20.)
10. **Trap tick timing**: Immolation Trap (21 sec) and Explosive Trap (10 sec): how often do they tick?
11. **Illumination** (paladin): the tooltips for ranks 2-4 show the rank 5 text. What % of the mana cost do ranks
    2, 3 and 4 return? (The sim uses 27.5 / 35 / 42.5%.)
12. **Hunter Thrill of the Hunt**, **Shaman Shamanism**: the spell data has no proc chance ("a chance"); probably a
    procs-per-minute value from a table that is not exported. If you know the numbers, they can go in.
13. **More DBC tables**: if you can export `SpellCastTimes.csv` (and `SpellCooldowns`/`SpellCategory` if your
    tool has them) the same way as `Spell.csv`, I can check every cast time and cooldown against the sim too. A
    percentage-of-base-mana cost column is also missing (Bestial Wrath, Bloodlust, Aftershock, Seal of Fury).
16. **Feral Tank values not in the server data** (Part CF): how much does Swipe gain from attack power ("increased by
    Attack Power", sim: none yet)? How much rage does Grizzly's Fury give when you are hit? Does Leader of the Pack
    really double Dire Bear Form (armor from items +720% instead of +360%, health +40%, +360 attack power)? Compare
    the character sheet armor with and without the talent. Enrage's armor loss (sim: classic 16% in Dire Bear Form).
18. **Trueshot Aura ranks** (Part CL): the talent gives rank 1 (100 ranged / 50 melee attack power); Spell.csv also
    has ranks 2 (150/75) and 3 (200/100). Can hunters train ranks 2-3 on the server? The sim now uses rank 3.
19. **Hemorrhage charges** (Part CN): the server spell 17348 gives +25 Physical damage taken for 15 sec and has no
    charge count (classic: 30 hits). Does the debuff really last the full 15 sec, or does it still end after N hits?
20. **Execute extra rage** (Part CN): how much damage does each extra point of rage add? The tooltip value (`$f1`)
    is not in the dump; the sim uses classic 15 per rage.
21. **Weapon procs without a proc rate** (Part CO): these epic weapons have a "Chance on hit" / proc line but no code
    in the sim, and the dump has no procs-per-minute: Axe of the Deep Woods (90-126 Nature), Kang the Decapitator (560
    bleed over 30 sec), Taran Icebreaker (180-220 Fire + 36), Brain Hacker (200-300 + Intellect -75), Sul'thraze the
    Lasher, Blackfury (+20% attack speed 10 sec), Hurricane / Dwarven Hand Cannon (ranged procs), Ancient Hakkari
    Manslayer (48-54 life steal), Sliverblade (45 Frost), and server items Magmastrike (+90 attack power 20 sec),
    Aquastrike (mana), World Breaker (armor -50% for 3 attacks), Thunderstrike, Shadowstrike. Do you know their proc
    rates? Otherwise the sim could use 1 proc per minute for all of them (upstream uses about that for similar weapons).
22. **Rivenspike** (Part CO): the server text is "Equip: Your attacks ignore 5% of your enemies' Armor", the sim still
    has the classic proc (reduce target armor, 2 per minute). Is there still a proc on the server, or only the 5%?
24. **Shield Block** (Part CR): the server has 2565 (+50% block for 10 sec, 20 sec cooldown, blocks 1 attack) and 12169
    (+75% for 5 sec, 5 sec cooldown). Which one does a level 60 warrior cast (is 12169 a talent or a second rank)? The
    sim uses 75% / 5 sec / 5 sec.
25. **Ignite** (Part CU): with 5/5 Ignite, does a 1000-damage Fire crit give 4 ticks of 100 (40% over 8 sec, what the
    sim does now, from the server text), or 2 ticks of 400 (the old classic sim model)?
26. **Demonic Sacrifice** (Part CU): the server buffs last 5 min (classic 30 min). Can the warlock resummon and
    sacrifice again in a fight, or should long fights just lose the buff?
27. **Battlegear of Might 8** (Part CX): "Your Overpower, Revenge and Execute abilities increases the damage of your
    next offensive ability by 8%." How long does that buff last (sim: 15 sec)? Is it used up by the next ability?
28. **Gauntlets of Might (16863) and Arcanist Boots (16800)** (Part CX): the dump still has these classic pieces with
    the old set bonuses, next to the renumbered pieces. Can players still get them, or should they leave the DB?
29. **Set proc chances** (Part CX): Lightforge 6 / Soulforge 4 (+5% crit for 10 sec) and The Elements 6 (+100 attack
    power, damage and healing) say "chance on offensive action" without a number; the sim uses 6% / 6% / 4%.
30. **Items whose effect type changed** (Part CY): Seeping Willow ("Use: Lowers all stats by 50, armor by 400 and deals
    20 Nature damage every 3 sec ... for 30 sec") and Ragehammer ("Use: Instantly generates 80 rage and increases damage
    done by 10 and attack speed by 5% for 15 sec") are Use effects on the server, the sim still has them as on-hit
    procs. Eskhandar's Left Claw now "causes 55 to 95 damage" (the sim: a bleed). Cooldowns of those Use effects?
31. **Leader of the Pack and Cat Form** (Part CZ): the talent "doubles the effects of your Bear and Cat Forms". The
    sim now doubles Cat Form's +5% crit to +10%, but keeps its threat at -29% (not -58%). Is that how it works on
    the server? (Bear Form: armor, health and attack power are all doubled.)
32. **Flurry and Lightning Shield timings** (Part DA): the server's Flurry buff lasts 8 sec (classic 15 sec) and
    Lightning Shield has 5 charges (classic 3). The sim now uses the server values; a quick check in game would
    confirm the 8 sec. Also: does Lightning Shield have a hidden cooldown between procs? The sim uses 3.5 sec.
33. **Racials** (Part DB): (a) Orc Command's tooltip says "and your damage if an ally is nearby by 1%", but the spell
    data only has the pet bonus. Is the 1% real (then the sim adds it, since a raid always has allies nearby)?
    (b) Troll Berserking's casting speed is a flat +5% in the data (melee speed scales 10-30% with missing health);
    does the casting speed also scale with health? (c) Which races have "Light Weapons Specialization" (20558) and
    "Hunting Weapons Specialization" (26290)? The sim gives both to Trolls (classic ids of the troll racials).
34. **Shaman talents** (Part DD): (a) Can Bloodlust be cast on the shaman itself? The sim assumes yes and keeps it up
    (3 min buff, 1 min cooldown). (b) What do Aftershock and Bloodlust cost? Their cost is a percentage of base mana,
    which the spell export does not include (the sim has them free). (c) Does Armaments of Storm scale with spell
    power, and can it crit? The sim uses a flat 5 x level and lets it crit like a spell.

## Talents not in the sim code (regenerated 2026-09-25, after Part CL; updated after Part DH)

Every talent field that no Go code reads (`Talents.<Name>` never used under `sim/`). Regenerate with a grep of
`Talents\.` in `sim/<class>/` against `ui/core/talents/trees/<class>.json` (per class: some names like Survival
Instincts and Cold Blood exist in two classes).

Most of these are utility or PvP (stuns, fears, range, movement, stealth, pet utility, healing-only) and don't change a
DPS sim. Bosses can't be stunned, so stun-based bonuses (Improved Kidney Shot, Shock and Awe, Seal of Command's stunned
judgement) are left out on purpose. Still open for a DPS sim:
- **Divine Might** (paladin): not read by the sim code, but the Blessing of Might "improved" option (+20%) and the
  Blessing of Kings 12%/13% options are exactly Divine Might 2/2 (labels say so). Rank 1 (+10%) has no option.
- **Improved Hunter's Mark**: covered by the "improved" Hunter's Mark debuff option (+20%/rank, 2/2 = +40%);
  the hunter doesn't cast the mark itself, so the talent field itself stays unused.
- **Wand Specialization** (mage, priest): wands aren't used by any rotation.
- Needs numbers: Shaman Shamanism (chance), Hunter Thrill of the Hunt (chance), Mage Ice Shards (cast time, mana cost
  and spell power coefficient; Cold Grip depends on it).
- Needs a new mechanic: freeze for Shatter / Frostbite / Deep Freeze, Chain Reaction, Earthquake, Berserker's Blood
  (needs the player's health).

**LOW PRIORITY: threat, tank and defensive talents** (Lokiy 2026-09-24: rotations and threat calculations wait until
all the other data is done; do these as one batch then):
- Druid: Feral Instinct Bear Form threat +5/10/15%; Primal Tenacity and Improved Enrage (Enrage isn't in the sim);
  Custody of the Nature (crit immunity after being crit).
- Warrior: Mocker (Taunt/Challenging Shout/Mocking Blow hit +6%), Improved Defensive Stance (absorb shield on stance
  swap), Concussion Blow and Shield Toss ("high amount of threat": the threat value and Shield Toss attack power
  coefficient are unknown), Piercing Howl threat, Blood Craze, Berserker's Blood (needs the player's health).
- Paladin: Guardian's Favor Salvation part (Sanctuary part done in Part DI), Second Wind, Stoicism, Eye for an Eye, Improved
  Concentration Aura.
- Warlock: Damned Vanguard (demon threat +30%, damage taken -30%).
- Hunter: Intimidation threat, Improved Distracting Shot, Deterrence, pet Survival Instincts and Bestial Swiftness.
- Others: Mage Cryo Core, Priest Blur and Martyrdom, Rogue Survivor and Heightened Senses, Shaman Earth Shield,
  Earthquake threat and Tidal Barrier.
- To check in game: Slam flat threat 140, Rockbiter Weapon threat, Shen'dralar Badge of Deterrence and threat-reducing
  trinket effects (Fetish of the Sand Reaver, Grace of Earth, Two-Faced Medallion use effects).

- **Druid** (11 of 59): Cycle Of Life, Mighty Roots, Hurricane, Starfall, Brutal Impact, Feral Charge, Leap, Untamed Heart, Natures Focus, Custody Of The Nature, Catharsis
- **Hunter** (20 of 60): Improved Mend Pet, Aspect Mastery, Improved Revive Pet, Intimidation, Bestial Swiftness, Survival Instincts, Team Play, Terrifying Roar, Improved Hunters Mark, Improved Concussive Shot, Vantage Point, Hawk Eye, Improved Distracting Shot, Scatter Shot, Deterrence, Improved Wing Clip, Trapper, Thrill Of The Hunt, Deep Freeze, Wyvern Sting
- **Mage** (21 of 60): Wand Specialization, Practical Defensive Magic, Practical Offensive Magic, Flame Throwing, Impact, Blazing Speed, Chain Reaction, Frost Warding, Permafrost, Cryo Core, Frostbite, Improved Frost Nova, Cold Blood, Ice Block, Arctic Reach, Shatter, Cold Grip, Ice Mirror, Ice Shards, Ice Barrier, Advanced Ice Shielding
- **Paladin** (15 of 60): Aura Mastery, Spiritual Focus, Inner Light, Divine Grace, Unyielding Faith, Holy Purge, Improved Hammer Of Justice, Dominance, Second Wind, Improved Concentration Aura, Stoicism, Pursuit Of Justice, Divine Might, Eye For An Eye, Repentance
- **Priest** (19 of 60): Pilgrimage, Wand Specialization, Martyrdom, Improved Dispel Magic, Focused Casting, Stratagem, Holy Focus, Lights Grace, Blessed Recovery, Holy Nova, Holy Reach, Spirit Of Redemption, Holy Link, Blackout, Improved Psychic Scream, Shadow Word Numb, Improved Shadow Word Silence, Blur, Insanity
- **Rogue** (19 of 60): Remorseless Attacks, Seek And Destroy, Vitality, Improved Kidney Shot, Total Control, Improved Gouge, Improved Sprint, Improved Kick, Dazing Bolts, Survivor, Improved Sap, Master Of Deception, Thug Life, Setup, Camouflage, Heightened Senses, Shadow Cut, Cloak Of Shadows, Enveloping Shadows
- **Shaman** (15 of 60): Storm Reach, Earths Grasp, Sand Blast, Eye Of The Storm, Earthquake, Earth Shield, Improved Ghost Wolf, Shamanism, Spiritwalking, Totemic Mastery, Tidal Barrier, Nature Focus, Ancestral Healing, Focused Mind, Cleansing Wave
- **Warlock** (16 of 60): Fel Concentration, Jinx, Dread, Black Speech, Withering Shroud, Death And Decay, Improved Healthstone, Improved Health Funnel, Improved Voidwalker, Master Conjuror, Damned Vanguard, Improved Felhunter, Improved Enslave Demon, Aftermath, Pyroclasm, Shock And Awe
- **Warrior** (11 of 60): Improved Charge, Improved Hamstring, Combat Endurance, Piercing Howl, Blood Craze, Berserkers Blood, Mocker, Improved Defensive Stance, Concussion Blow, Iron Will, Shield Toss

## Healers (added 2026-09-24)

All four heal pages work (Healing Priest, Restoration Shaman, Restoration Druid, Holy Paladin) with the server's heal
values. Left for later, roughly in order:
- **Healer gear presets**: every healer page uses a placeholder set from a DPS caster page (shadow priest, elemental,
  balance). Make real healing sets (+healing gear) once the item list is checked.
- **Healing rotations** (low priority like all rotations): fixed priorities on one target dummy, so there is no
  overhealing and HPS is limited by mana. A healing model (damage taken by the tank / raid) would make HPS meaningful.
- **Cast times**: taken from the server cast time index, mapped by comparing with known spells (16 = 1.5, 5 = 2.0,
  20 = 2.5, 14 = 3.0, 22 = 3.5 sec; 19 assumed 2.5). Server Chain Heal uses index 22 (3.5 sec, classic 2.5) and 70%
  per jump; Healing Touch ranks 1-8 are faster than classic. Confirm with a `SpellCastTimes.csv` export (question 13).
- **Spell power coefficients** are the classic rule (cast time / 3.5, HoTs duration / 15). Guesses: Tranquility (server
  heals every second), Swiftmend (also its mana cost, 16% of base mana), Prayer of Healing and Chain Heal splits.
- **Not in the sim yet**: Holy Nova, Lightwell, Holy Link, Inspiration, Ancestral Healing, Earth Shield, Healing Stream
  Totem as a healer spell choice, Lay on Hands as a heal, Blessing of Light, Improved/Lesser Heal, Cleansing Wave,
  Holy Purge, Spirit of Redemption. Holy Shock (heal) has its own cooldown separate from the damage version.
- **Paladin Illumination** ranks 2-4 still interpolated (question 11); Light's Mercy stack effect (20% per stack) is
  from the tooltip wording.
- Raid sim is still unlaunched (only the individual pages are Phase 1 / Alpha).
- Priest healing talents (Spiritual Healing, Improved Healing, Improved Renew, Improved Power Word: Shield, Improved
  Prayer of Healing, Holy Focus, Holy Reach, Holy Nova, Light's Grace, Blessed Recovery, Spirit of Redemption, Holy
  Link) are not modeled.

## Feral tank (added 2026-09-24)

Bear threat values (Maul 1.75, Demoralizing Roar 42) are classic guesses (low priority). The bear gear preset is the
cat pre-raid set as a placeholder. Bash, Growl, Challenging Roar, Feral Charge, Barkskin, Frenzied Regeneration
(`sim/druid/_frenzied_regeneration.go` is disabled, 5 min cooldown confirmed in game) and Rebirth (30 min) are not in
the sim.

## Open items at a glance (updated 2026-09-25)

**Affects sim results**
- Pyroblast, Flame Shock, Insect Swarm and Arcane Missiles spell power coefficients are a guess (classic total spread
  over the new tick count); Arcane Missiles' +1% Arcane crit is one stack per cast.
- LOW PRIORITY (Lokiy 2026-09-24: rotations and threat need all other data first): paladin ret/prot rotations are
  new and basic (no Consecration for ret, no mana management); all rotations and threat values need tuning in game.
- The default elemental shaman rotation never casts Flame Shock; adding it (when its DoT is not active) is +6% in the
  test setup, and it is needed for Aftershock (Part DD).
- No preset uses Arcane Missiles or Insect Swarm, so the tests don't cover them (checked in the browser instead).
- Talent presets are all empty (question 1 at the top; low priority, one of the last tasks).
- Consumables the user chose to skip (Part P): Bloodkelp Elixir of Dodging (22192) / Resistance (22193), Elixir of
  Wisdom (3383), Potion of Fervor (1450), Minor Magic Resistance Potion (3384), Combat Healing/Mana Potion
  (18839/18841). Still absent if ever wanted.

**Needs data from the game**
- Trinkets: Arcanite Dragonling and Cannonball Runner (pet damage), Six Demon Bag (effects); Orb of Chaotic Elements
  heal amount is assumed equal to its damage (question 4).
- Talent and set values marked "confirm in game" in the class sections below.

**Data / pipeline**
- Crafted items are all Phase 1 (deferred by the user: "adjust manually later"). The idea was to move a crafted item
  to a later phase when its pattern and a reagent are both BoP from that phase's content (Dark Iron / Molten /
  Corehound / Flarecore -> MC; Chromatic / Dreamscale -> BWL). Nothing adjusted since.
- Unverified classic values (not in the server data): Slam flat threat 140.
- T3-looking sets (Dreadnaught, Cryptstalker, etc.) aren't pinned to a phase; they fall through to Phase 1 since they
  have no AtlasLoot raid source on this server. Fine unless the server raids that content.
- Why the generic Wowhead-derived pipeline failed to tag `setName` for some items without a `custom_items.json`
  override was never diagnosed (patched item by item: Righteous Boots, Cenarion Armor, Striker's Garb, ...).
- AtlasLoot-based loot sources only cover dungeons, raids and world bosses; `Crafting/`, `Factions/` and `PvP/` sources
  keep the generic (retail) data. Correcting those is a separate follow-up.
- `docs/private-server-item-rules.md` "Known implementation wrinkles" still says `gen_phases.py` "only scans
  `Instances/`"; it now goes through `serverdata.atlasloot()` and covers all seven AtlasLoot folders.

## Class notes still open (from the 2026-09-19/20 audits)

**Warrior**
- Maim (10% on auto attacks per the DBC, +5% damage taken, duration unknown, all three ranks look identical); Shield
  Block (question 24); Berserker's Blood; Improved Hamstring, Improved Charge; Butterfly Style rage part; Execute
  rage-to-damage ratio (question 20); Cleaving (Thunder Clap/Whirlwind), Improved Execute.

**Rogue**
- Not modeled: Improved Sinister Strike extra-hit proc (3/5%), Coup de Grace (+5%/rank damage below 20% health),
  Bloodthirsty (Garrote/Rupture +10%/rank damage, shorter ticks), Combat Rush, Brigandage (damage may need a
  coefficient check), Gaining an Advantage, Dazing Bolts, Survivor, Physical Prowess (and the Sprint cooldown part,
  Improved Sprint), Improved Kidney Shot, Remorseless Attacks (needs kills), Weapon Expertise values. Thistle Tea
  cooldown not in the DBC.

**Hunter**
- Not modeled: Melee Specialization (+30% melee speed, -30% ranged), Dual Wield Specialization (OH +20-50%), Weapon
  Expertise (extra arrow), Find Weakness, Thrill of the Hunt (question 12), Deadeye, Stalking's kill bonus, Vantage
  Point, Spirit Bond, Improved Tracking, Deep Freeze, Team Play, Aspect Mastery, most pet utility talents.
- Reconnaissance is coded 3%/rank all damage; DBC spell 34063 (3%) may be flat at every rank. Unchecked.
- Kill Command mana cost (assumed free, check in game); Savage Blow damage (placeholder: one main-hand and one
  off-hand weapon hit), range and the Hawk/Cheetah/Wild effects; Whirling Axe cooldown, mana cost and range (none
  set), slow and interrupt not modeled; Aspect of the Beast/Pack stat effects not applied; Hunter's Mark melee AP
  (+90) not modeled; shot damage and mana tables by rank not compared; Lethal Shots and Weapon Expertise crossbow
  crit are shown in the Melee Crit tooltip only.

**Paladin**
- Not modeled: Divine Grace (needs Seal/Judgement of Light and Wisdom, not in the sim), Light's Mercy (Flash of Light
  only on the heal page), Holy Purge, Eye for an Eye (10% of spell damage taken reflected), Repentance, Codex Holy
  Light cost/cast time, the Judgement of Fury forced attack. Illumination: question 11.
- Seal of Command: the 1 sec internal cooldown is unverified (damage 50%, 12 procs per minute, 120 sec duration and
  top rank Judgement of Command 441-475 are confirmed).
- Holy Shield spell power coefficient (sim 0.05), Holy Shock healing.

**Shaman**
- Elemental Devastation: DBC aura 30165 gives a flat 10% crit at every rank, code uses 3%/rank. Looks odd, so not
  changed; confirm on the server.
- Not modeled: Static Field, Earthquake, Earth Shield, Guardian Totems (partly in core), Improved Ghost Wolf,
  Elemental Mastery values unchecked, healing talents (Healing Way, Purification damage part, Tidal Mastery speed part).

**Druid**
- Not modeled: Starfall, Power of Nature, Dreamstate, Mighty Roots, Hurricane, Cycle of Life, Unity with Nature,
  Stalking, Predatory Strikes (custom DBC values look malformed), Primal Fury, Survival Instincts, Furor values,
  healing talents (Improved Rejuvenation/Regrowth, Gift of Nature healing part, Tranquil Spirit), Swiftbloom,
  Naturalist, Killer Instincts attack speed part.

**Mage**
- Not modeled: Shatter (needs a frozen-target state), Spell Twisting, Impact, Blazing Speed, Chain Reaction, Thermal
  Expansion, Cryo Core, Cold Grip, Rimebound (needs spell 34266), Advanced Ice Shielding, Frost Warding/Fire Warding,
  Practical talents, Brilliance Aura, Magic Absorption mana-on-resist.
- Hot Streak/Pyromania untested in a fire rotation (default mage APL is not fire).

**Warlock**
- Not modeled: Prolonged Misery (+2/4/6 sec DoT duration, not a multiple of the 3 sec tick), Jinx, Sadism,
  Improved Immolate (2 stacks; current 5% bonus is a guess), Pyroclasm/Mayhem/Shock and Awe, Demonic Embrace regen,
  Soul Link damage split (redirect is 20% per DBC; code uses a flat multiplier), Improved Drain Soul health/mana regen
  on kill (stacks to 5).
- Master Demonologist Felhunter resist and Voidwalker values were assumed to scale 3/rank and 2/rank from rank-1 DBC
  only; ranks 2-5 of the sub-spells not checked.
- Suppression and Bring the Pain (+5%/rank crit on 4 spells) are not in the sidebar.

**Priest**
- Power Word: Requital: implemented 2026-09-19, but the coefficient is a placeholder (1.5/3.5) and ranks 2-4 lack a
  DB name/icon.
- Not modeled: Purifying Light undead/demon bonus, Spirit Tap on-kill proc (50%/100% per rank), Improved Inner Fire,
  Wand Specialization, Blackout, Blur, Focused Casting, Improved Psychic Scream / Shadow Word: Silence / Shadow Word:
  Numb, Pilgrimage, Martyrdom, Stratagem, Improved Dispel Magic, Insanity.
- Talent hit/crit (Spell Focus, Force of Will, Holy Specialization) is applied per spell, so it does not show on the
  sidebar Spell Hit/Crit lines.

## Buffs

- Resistance totems for an outside shaman have no picker (Guardian Totems from a shaman in the sim works: 60 -> 90).
  Same for Improved Weapon Totems.
- When the buff list is finished, consolidate the buff notes into one to-do (Lokiy's request) — not done yet.

## Shaman totems and imbues (2026-09-20)

- Mana Spring, Healing Stream and Windfury Totem dropped from the APL still use the hard-coded raid buff values
  (see the TODO comments in `sim/shaman/air_totems.go` and `water_totems.go`).
- Flametongue Totem is only the top rank (58+, +40 spell damage, 12.17 fire damage per second of weapon speed) and only
  for non-shaman classes through the weapon imbue option; there is no Flametongue Totem spell for the shaman to cast.
- Frostbrand damage of ranks 2 to 4 not used (the sim only uses Frostbrand rank 5); Frostbrand's movement slow is not
  modeled (only the attack speed slow).
- Rockbiter Weapon extra threat: the sim keeps the old per-rank bonus, the real value is unknown.
- Windfury Weapon keeps a 1.5 sec internal cooldown between procs: test in game whether that is right.
- Weapon enchants last 5 minutes: the Flametongue, Windfury and Frostbrand procs stop after 5 minutes, but the stats
  from Flametongue (spell damage) and Rockbiter (Strength, healing) are permanent for the whole fight.
- A shaman dropping his own Mana Spring Totem does not use Restorative Totems (the raid buff "Improved" option is +50%).
- Fire totems match the server data at level 60 (Searing Totem rank 6, Magma Totem rank 4, Fire Nova Totem rank 5);
  decide whether anything is missing.

## Trinkets and items still open

- Arcanite Dragonling and Cannonball Runner (summons, no pet data), Six Demon Bag (random effects), Grace of Earth and
  Two-Faced Medallion use effects (threat / aggro radius), Banner of Challenge (taunt). The Orb of Chaotic Elements
  heal is assumed to equal the damage dealt (question 4).
- Cooldowns assumed from classic where the server changed the effect: Gri'lek's, Wushoolay's and Hazza'rah's charms
  (3 min), Aegis of Preservation (5 min). Confirm in game.
- Shard of the Flame (17082) has no stats in the DB (server: 16 health per 5 sec, no DPS effect); Scrolls of Blinding
  Light (19343) is not in the server data.
