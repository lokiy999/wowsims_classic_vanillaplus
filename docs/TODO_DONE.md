# Done items

Items from [TODO.md](TODO.md) that are finished, answered or decided. Moved here on 2026-09-25 (Lokiy: keep the TODO
list to open items only). Grouped by the TODO section they came from; details are in [CHANGES.md](CHANGES.md) under
the Part named with each item. New finished items are added at the top of their section.

## Questions for Lokiy (answered)

- **36. Buff stacking**: ANSWERED 2026-09-30: Elixir of the Sages and Juju Guile stack with Arcane Intellect, Brute Force stacks with Juju Power (both already so); Stoneskin Totem and Devotion Aura "likely don't stack": now exclusive, the larger armor applies (Part ED).
- **22. Rivenspike**: ANSWERED 2026-09-30 by VPlusItemDB: both "Equip: Your attacks ignore 5% of your enemies' Armor" and "Chance on hit: Reduces targets armor by 300 for 20 sec"; both in the sim now (Part DX).
- **10. Trap tick timing**: ANSWERED 2026-09-29 from Spell.csv (tick interval column): Immolation Trap every 3 sec (7 ticks over 21 sec), Explosive Trap's burn every 2 sec (5 ticks over 10 sec); the sim already used these (Part DR).
- **3. Arcane Missiles "energize"**: ANSWERED 2026-09-24: once per cast, 5 stacks max (as implemented).
- **8. Mana Tide Totem**: ANSWERED 2026-09-24: yes, trainable; now castable by the shaman (Part BV).
- **12. Armaments of Storm** (part of question 12): numbers found in Spell.csv (5%/rank, 5 x level), done in Part DD.
- **14. Cooldowns in Spell.csv**: ANSWERED 2026-09-24: column 20 is the cooldown; applied (Part BZ). Counting from 0: column 19 is the spell's own cooldown, column 20 the category cooldown (column 21 the category); a spell uses one of them (Part DQ).
- **15. Item list resync**: ANSWERED 2026-09-24: follow the dump, never Naxx, rename ok; done (Part BY).
- **17. Flask of Indomitable Might icon**: ANSWERED 2026-09-25: INV_Potion_21 (in-game AtlasLoot), applied.
- **23. Cloak of Untold Secrets**: ANSWERED 2026-09-25: drops from Krixix in BWL, keep it. Source set (Part CP).
- **Item pipeline rerun is not stable**: checked 2026-09-25 (Part CP): the full rerun was right. It drops Ring of
  Swarming Thought (21707, Skeram, AQ40, in no AtlasLoot table, caught by the raid-source rule); the committed list
  still had it and is fixed now. `add_items.json` changing between runs is expected.

## Talents

- Done on 2026-09-30 (Part EA): raid buffs and debuffs compared with the server; Blood Pact, Stormstrike, Curse of Weakness, Insect Swarm, Gift of Arthas and Scorpid Sting fixed.
- Done on 2026-09-30 (Part DY): set bonuses re-checked against VPlusItemDB; Blood Tiger Harness 2% crit, Soulforge Armor 4 crit proc added.
- Done on 2026-09-30 (Part DX): item effects compared with VPlusItemDB; Venomspitter, Keris of Zul'Serak, Skullforge Reaver, Hameya's Slayer, Mark of the Chosen, Rivenspike fixed.
- Done on 2026-09-30 (Part DW): feral druid, paladin and all heal tables compared with the server; Shred 250% weapon damage.
- Done on 2026-09-30 (Part DV): warrior abilities compared with the server; Overpower now also strikes with the off-hand.
- Done on 2026-09-30 (Part DU): level-keyed spell tables (rogue, warrior, druid, paladin, armors) compared with the server; rogue Expose Armor, Instant and Deadly Poison, Mage Armor and Demon Armor fixed.
- Done on 2026-09-29 (Part DS): damage and mana tables of all classes compared with the server by rank; warlock Searing Pain, Siphon Life, Drain Life and Conflagrate ranks fixed; Conflagrate cooldown back to 10 sec.
- Done on 2026-09-29 (Part DR): hunter shot, strike and trap damage and mana tables compared by rank with the server; Serpent Sting damage fixed.
- Done on 2026-09-29 (Part DQ): shaman Elemental Mastery checked against the server (crit only, the spell no longer makes the next spell free; 3 min cooldown matches).
- Done on 2026-09-29 (Part DP): warlock Master Demonologist checked for all five ranks; Felhunter resistance fixed (0.2 x level per rank).
- Done on 2026-09-29 (Part DO): warlock Improved Immolate (Immolate stacks up to 2 times).
- Done on 2026-09-25 (Part DL): boss spell damage target option; core death delay; paladin Second Wind, Stoicism,
  Holy Shield coefficient 0.1 per block (assumed by Lokiy).
- Done on 2026-09-25 (Part DK): paladin Eye for an Eye damage (threat still in the threat batch).
- Done on 2026-09-25 (Part DJ): paladin Seal of Light and Seal of Wisdom (passive healing / mana cost effects,
  judgements), Divine Grace (seal part).
- Done on 2026-09-25 (Part DI follow-up): Judgement of Wisdom 60 mana and 30 sec, Judgement of Light 30 sec.
- Done on 2026-09-25 (Part DI): paladin Blessing of Sanctuary values and level ranges, Guardian's Favor (Sanctuary
  part), Libram of Fervor Holy damage part, Libram of Divinity.
- Done on 2026-09-25 (Part DH): paladin Reckoning (4%/rank on any hit, 20%/rank on crits), Shield of Faith.
- Done on 2026-09-25 (Part DG): paladin Judgement of Righteousness and Judgement of Command ranks, Seal of the
  Crusader (attack power and Holy damage, no attack speed); Consecration, Exorcism, Holy Wrath (60 sec) and Divine
  Favor (5 min) compared, all matching.
- Done on 2026-09-25 (Part DE and follow-up): warlock Fel Pact, Feeding Demons, Herald of Woe.
- Done on 2026-09-25 (Part DD): shaman Armaments of Storm, Bloodlust, Aftershock.
- Done on 2026-09-25: Improved Rend (Part CM), Deep Wounds stacks (Part CR), Improved Berserker Stance GCD (Part CS),
  Para Bellum on every warrior ability (Parts CQ, CS), Connivery "in front of target" option and Exhaustion on
  Rupture / Expose Armor (Part CQ), druid Vengeance feral part (Part DA), Cat Form +5% crit (Part CZ).
- Done on 2026-09-24 (Part CE): Defiler, Inevitable Doom, Feral Instinct (Tiger's Fury part), Death Mark, Seal of
  Command (now needs the talent).
- Warrior Enrage: DONE 2026-09-19 (Part AP follow-up 2): 1% melee damage per stack for 15 sec, up to 10 stacks.
- Warrior Tactical Mastery: DONE 2026-09-19: retains an additional 10 Rage per rank.
- Mage Mind Mastery: also +20%/rank Arcane Intellect effect, done (`sim/mage/mage.go`), confirmed 2026-09-24.
- Shaman Rockhide, Primal Endurance, Nature's Grace (threat) and Nature's Guardian: done (`sim/shaman/talents_server.go`).
- Warlock Demonic Onslaught (pet crit): done (`sim/warlock/talents_server.go`); Improved Drain Soul cooldown and the
  wrong Curse of Doom use fixed 2026-09-20.
- Paladin: Divine Concentration, Improved Purifying, Holy Grasp, Blessed Strikes threat, Sanctity Aura talents, Seal
  of Fury, Improved Retribution Aura, Vengeance (Parts BC and BE). Improved Lay on Hands cooldown checked (-15 min per
  rank, matches). Holy Shock, Hammer of Wrath, Righteous Fury, Redoubt and Holy Shield confirmed (Part AY).
- Druid Improved Mark of the Wild: RESOLVED 2026-09-19, 20%/rank confirmed by Lokiy (x1.6, armor 456, Part AP).
- Priest Inner Fire and a no-proc Spirit Tap: done (Part AR).

## Class notes (checked against the code 2026-09-29, Part DO)

The class notes in TODO.md had items that were already done in earlier parts; moved here:
- Warrior: Maim (10/20/30 sec), Improved Hamstring and Improved Charge are utility, Cleaving, Improved Execute,
  Butterfly Style (dodge and crit).
- Rogue: Improved Sinister Strike extra hit, Coup de Grace, Bloodthirsty, Combat Rush, Brigandage, Gaining an
  Advantage, Physical Prowess (Strength and Evasion cooldown), Weapon Expertise.
- Hunter: Melee Specialization, Dual Wield Specialization, Weapon Expertise, Find Weakness, Deadeye, Stalking, Spirit
  Bond, Improved Tracking; Reconnaissance 3%/rank matches the talent text (15% at 5/5); Kill Command, Savage Blow,
  Whirling Axe, Aspects of the Monkey/Beast/Pack are in; Lethal Shots and crossbow Weapon Expertise show in the
  Ranged crit line.
- Paladin: Divine Grace (seal part), Light's Mercy, Codex of the Silver Hand; Seal of Command values.
- Shaman: Elemental Devastation (+10% melee and spell crit, 5/10/15 sec), Static Field, Guardian Totems, Healing
  Way, Purification, Tidal Mastery.
- Druid: Power of Nature, Dreamstate, Mighty Roots is utility, Unity with Nature, Stalking, Predatory Strikes (300%
  of level at 3/3, matches the text), Primal Fury, Survival Instincts, Furor, Improved Rejuvenation / Regrowth, Gift
  of Nature, Tranquil Spirit, Swiftbloom, Naturalist, Killer Instincts.
- Mage: Spell Twisting, Thermal Expansion, Rimebound, Fire Warding, Brilliance Aura, Magic Absorption.
- Warlock: Prolonged Misery, Sadism, Mayhem, Demonic Embrace, Soul Link, Improved Drain Soul (cooldown part),
  Bring the Pain; Improved Immolate now stacks Immolate twice (Part DO, was a guessed +5%).
- Priest: Purifying Light, Improved Inner Fire, healing talents (Spiritual Healing, Improved Healing, Improved Renew,
  Improved Power Word: Shield, Improved Prayer of Healing).

## Healers and feral tank

- Healing done in Part CG (priest, shaman, druid incl. Nature's Swiftness and Tranquility, paladin); Mana Tide Totem
  castable (Part BV).
- Feral Tank page done (Part CF). Cat Form's passive (3025: +5% crit) done in Part CZ, doubled by Leader of the Pack.
- Gear slots empty on first load: not a bug (checked 2026-09-25). That is the testing switch
  `START_WITH_EMPTY_PRESETS = true` in `ui/core/individual_sim_ui.ts` (Part AT); set it to `false` when the presets
  should come back.

## Sim results, data and pipeline

- Item database resync: done 2026-09-24 (Part BY).
- "Level 60" target used a SoD NPC id; BWL encounter mechanics copied from SoD; `tools/database/atlasloot.go` read the
  AtlasLootClassic_SoD repository: done 2026-09-24 (Part CB).
- Sidebar: armor from gear Agility shown under Base, and derived stats counted as "Gear": fixed 2026-09-24 (Part BW).
- Spell values: direct and periodic damage per rank match the server (Parts BO, BP, BR, BT); heals (Part CG); raid
  buffs, debuffs and world buffs (Part CL); physical and pet abilities (Part CN); resource costs (Part CW); set
  bonuses (Part CX); item effects (Part CY); self-buffs and cooldowns (Part DA); racials (Part DB); weapon enchant
  procs (Part DC); shaman imbues (Part DF).
- Done 2026-09-23: SoD content removed or replaced (Part BI); trinkets (Parts BJ-BM); boss crit only reduced by Defense
  (Part BK); Memory of Hyjal flat damage reduction, Greater Protection Potions (absorb shields), The Black Book pet
  damage reduction, Presence of Might +10 all stats, Law of Nature removed (Part BN).
- Pre-existing `go test ./sim/...` failures: RESOLVED 2026-09-20. All `.results` files were regenerated during the
  second audit round; `go test ./sim/...` passes with and without `--tags=with_db`.
- Everything from the early sessions is committed and pushed (old note: 216 files uncommitted).

## Player buff audit (2026-09-19/20)

- Done 2026-09-20 (Part BC): Devotion Aura 700 and +25%/point, Commanding Shout and Horn of Lordaeron removed,
  `IncludeAQ` on, faction gating dropped, Moonkin Aura +5% spell damage and healing, Trueshot Aura, Sanctity Aura and
  the resistance auras.
- Libram of Truth toggle: done 2026-09-24 (Part CC).
- Arcane Intellect: done 2026-09-20 (30/37/60/67 picker, Part AS).
- Improved Gift of the Wild armor: done 2026-09-19, Improved MotW is 20%/rank (x1.6, armor 456, Part AP).

## Consumables (2026-09-17)

- Implemented (Part Q): Flask of Indomitable Might, Elixir of Brute Force, Elixir of Demonslaying, Elixir of Greater
  Intellect, Elixir of the Sages, Juju Guile, and all 4 Troll's Blood Potions (new proto enum values, `IntellectElixir`
  category, `sim/core/consumes.go`, UI pickers).
- The Greater and base Protection Potions (absorb shields): done in Part BN with the new absorb-shield support.

## Known gaps (2026-09-13)

- Custom class sets have no set bonuses: done 2026-09-13 (Talonclaw Regalia, Ursoc Armor, Cataclysm Armor, The
  Stonefury, Righteous Armor; commit `d464da4b3`, Part C).
- All PvP sets need their bonuses done: done 2026-09-13 (commits `79cb702a3`, `590bfad43`); every PvP set bonus
  checked again against the server in Part CX.
- Every set needs a full pass: done 2026-09-13 (commit `bfcd64728`) and again against the server tooltips in Part CX.
- Memory of Hyjal (81015) "reduces all damage received by up to 14": done in Part BN.

## Shaman totems and imbues (2026-09-20)

- Mana Tide Totem cast by the shaman: done (Part BV, `sim/shaman/mana_tide.go`).
- Rockbiter Weapon healing at level 50 (rank 6) and ranks 2-3: filled from the shifted Spell.csv rows (Part DF).
- Healing Stream Totem base healing formula: fixed 2026-09-20 (Part AW).
- Improved Fire Totems: done 2026-09-24 (Part BR).

## Trinkets and items (2026-09-23)

- The Black Book: done 2026-09-23 (Part BN): pet damage +100% and pet damage taken -100% for 30 sec.
- Done in Parts BJ-BM: Uther's Strength, Force of Will, Ragged John's Cup and Blazing Emblem damage reduction, Heart of
  the Scale thorns, Petrified Scarab decay, Aegis of Preservation heal, Sawtooth Talisman 5%, Reactive Auto-Recaster,
  and every custom-item cooldown (from in-game tooltips): Mark of Bestial Fury, Blood Scarred Scale, Scarlet Battle
  Orders, The Final Gaze, Dark Iron Bookmark, Orb of Chaotic Elements, Everlasting Liver, Ironbark Tea Leaf, Banner of
  Challenge, Hibernation Crystal.
- Royal Seal of Eldre'Thalas (18466) "+12 to all attributes" and the Blue Mottled/Pink Speckled Egg "+10 All
  Resistances": done 2026-09-24 (Part CC, plus 201 other items with "Equip: +N All Resistances"). Onyx Egg "-1% damage
  taken" and the Royal Seal (18471/18472) "-2% spell cost" already in `vplus_trinkets.go` (checked 2026-09-24).
- Weapons coded as their Season of Discovery versions: fixed in Part BI (checked 2026-09-24: Thunderstrike,
  Shadowstrike and Masterwork Stormhammer match the server tooltips). SoD ids for the Mangle debuff and the Primal
  Blessing set proc fixed in Part BI.
- Flame Wrath and Quel'Serrar Use effects: done 2026-09-23 (Part BM, cooldowns from the tooltips).
- Presence of Might: already +10 to all five stats in `enchant_overrides.go` (checked 2026-09-24).
- Enchant Shield - Law of Nature (SoD enchant): removed (Part BN).
- Server spell values that differ from the sim (e.g. Immolation Trap, Arcane Shot): done 2026-09-24 (Parts BO, BP, BR).

## Decided, left as they are

- Local spell names that differ from the server: 713 Summon Incubus (server: Turn Undead), 1098/11725/11726 Subjugate
  Demon (Enslave Demon), 10326 Turn Undead (Turn Evil), 13219 Wound Poison (Wound Poison I), 13948 Enchant Gloves -
  Minor Haste (Lesser Haste), 15487 Silence (Shadow Word: Silence), 28271/28272 Polymorph (Polymorph: Turtle / Pig).
  The Arcanum names and "Judgement Armor 8-piece" are deliberate.
- 13 custom enchants (900101-900212) have no spell row or label; the gear slot shows their name, which reads fine.
- Trinkets that are utility only, fine to leave (62): Abyss Shard, Alchemists' Stone, Ankh of Life, Arcane Infused Gem,
  Arena Grand Master, Barov Peasant Caller, Chronomirage, Dalaran Spellshackles, Darkmoon Card: Twisting Nether,
  Defender of the Timbermaw, Defiler's Talisman, Dimensional Ripper - Everlook, Dog Whip, Elementium Echo Modulator,
  Enamored Water Spirit, Gnomish Universal Remote, Gyrofreeze Ice Reflector, Heart of Noxxion, Hederine Shackles, Hook
  of the Master Angler, Hyper-Radiant Flame Reflector, Insignia of the Alliance/Gladiator/Horde/Unfettered/Unfettered
  Champion, Leyguard Charm, Lifestone, Major/Minor Recombobulator, Orb of Deception, Personal Harm Prevention Field
  Emitter, Piccolo of the Flaming Fire, Scarlet Hound Whistle, Scarlet Triage Kit, Still Eye of the Watcher, Talisman of
  Arathor, The Wall's Chime, Ultra-Flash Shadow Reflector, Ultrasafe Transporter: Gadgetzan, Vial of Elune's Light.
- Warlock Withering Shroud (5 yd around the warlock) and Death and Decay (20 yd): do nothing at range, left out
  (Part DF).

## Lokiy's list (2026-09-30)
- Tears of Teremus: own toggle, stacks with everything (Part EG)
- King's Heart: +200 attack power, own toggle, stacks with everything (Part EG)
