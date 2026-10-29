# Minor changes — Sim + db.lokiy.dev (wowsims-classic, tools/itemdb)

Hotfixes, data added / removed, wording, settings and rebuilds: anything that doesn't change code. Real changes go
in `docs/CHANGES.md` ("## Part XX"). Collected with every other project's into the hub's `CHANGELOG_MINOR.md` (`tools/changelog.py`).

One line per change under its date (`## YYYY-MM-DD`), newest date at the bottom.

## 2026-10-28

- db.lokiy.dev: tracker mobs without an NPC ID matched (hub): Zul'Gurub's The Nameless Hermit 25030, Azus the Bloodseeker 25031, Fresh / Corrupted Blood Globe 25035 / 25036 and Caverngloom Crocolisk 25040 added to `logged_npcs.json` (DungeonJournal's ZG combat-log CSVs; Azus 25034 is a Blood Tide helper, left out). Scarlet Sharpshooter is 25230; 25233 (a "Dummy" helper that only casts Explosive Shot) is dropped, so the name matches one NPC. Winterax Envoy has no known ID (not in any log, the creature cache or Wowhead) and keeps its name-only /mob page. 587/588 tracker mobs matched.

## 2026-10-29

- db.lokiy.dev: Dragonbreath 50021 (tier set hand-in) also drops from Dire Maul West's Tendris Warpwood, Illyanna Ravenoak, Magister Kalendris and Immol'thar (`MANUAL_DROPS`; the loot tracker saw it drop after each of them, AtlasLoot lists only Prince Tortheldrin there). Gauntlets of Valor 16737 left as is: its one tracked drop came after Shadow Hunter Vosh'gajin (an AtlasLoot source), not Highlord Omokk (hub, from LootTracker).
