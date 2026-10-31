# Minor changes — Sim + db.lokiy.dev (wowsims-classic, tools/itemdb)

Hotfixes, data added / removed, wording, settings and rebuilds: anything that doesn't change code. Real changes go
in `docs/CHANGES.md` ("## Part XX"). Collected with every other project's into the hub's `CHANGELOG_MINOR.md` (`tools/changelog.py`).

One line per change under its date (`## YYYY-MM-DD`), newest date at the bottom.

## 2026-10-28

- db.lokiy.dev: tracker mobs without an NPC ID matched (hub): Zul'Gurub's The Nameless Hermit 25030, Azus the Bloodseeker 25031, Fresh / Corrupted Blood Globe 25035 / 25036 and Caverngloom Crocolisk 25040 added to `logged_npcs.json` (DungeonJournal's ZG combat-log CSVs; Azus 25034 is a Blood Tide helper, left out). Scarlet Sharpshooter is 25230; 25233 (a "Dummy" helper that only casts Explosive Shot) is dropped, so the name matches one NPC. Winterax Envoy has no known ID (not in any log, the creature cache or Wowhead) and keeps its name-only /mob page. 587/588 tracker mobs matched.

## 2026-10-29

- db.lokiy.dev: Dragonbreath 50021 (tier set hand-in) also drops from Dire Maul West's Tendris Warpwood, Illyanna Ravenoak, Magister Kalendris and Immol'thar (`MANUAL_DROPS`; the loot tracker saw it drop after each of them, AtlasLoot lists only Prince Tortheldrin there). Gauntlets of Valor 16737 left as is: its one tracked drop came after Shadow Hunter Vosh'gajin (an AtlasLoot source), not Highlord Omokk (hub, from LootTracker).
- db.lokiy.dev: Dragonbreath 50021 also drops from Captain Kromcrush (Dire Maul North; in the same loot as his Libram of Focus / Boots of the Full Moon in an approved upload). The Crimson Defender drop in Lokiy's pre-event data was Cannon Master Willey's loot (with Miniature Cannon Balls and Mantle of the Scarlet Crusade), already a source: no change.

## 2026-10-30

- db.lokiy.dev: 12 more drops from Rist's loot log in `EXTERNAL_DROPS` (the proposal's maybe list, Lokiy; not Living Essence 12803): The Emerald Dream (Pimgib, Wildspawn Felsworn), Codex of Defense, Five of Elementals, Gray Sack of Gems (Nefarian), Plans: Thorium Leggings, Ebonhold Gauntlets, Overlord's Chestplate of the Whale, Mithril Lockbox (Gordok Ogre-Mage, Slavering Ghoul), Arctic Ring of the Wolf, Aquamarine (Ancient Core Hound).
- db.lokiy.dev: Sandals of the Impure Saint 26355 drops from Brother Michael (SM raid; `MANUAL_DROPS` SMMichael): seen in Rist's loot log, confirmed by Lokiy. It had no source before.
- Docs: docs/TODO.md: what is still sourced from Wowhead (counts after Parts HF-HI) and how each could be replaced; the not-on-V+ line updated (left out since HF).
- db.lokiy.dev search: shamans can use polearms on Vanilla+ (`item_stats.py` WEAPON_USERS; Lokiy 2026-10-30).
- db.lokiy.dev: Heavy Junkbox (16885) gets the source "Pickpocketed by rogues from humanoids of level 50 and up (classic)" (`LOKIY_ON_SERVER`); it had none since Part HI.

## 2026-10-31

- db.lokiy.dev / sim: Ring of Critical Testing 1-4 (18968, 18970, 18971, 18982) are test items, left out (`item_sources.py` `TEST_ITEM_IDS`, Lokiy). The sim never had them.
