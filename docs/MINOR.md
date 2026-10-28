# Minor changes — Sim + db.lokiy.dev (wowsims-classic, tools/itemdb)

Hotfixes, data added / removed, wording, settings and rebuilds: anything that doesn't change code. Real changes go
in `docs/CHANGES.md` ("## Part XX"). Collected with every other project's into the hub's `CHANGELOG_MINOR.md` (`tools/changelog.py`).

One line per change under its date (`## YYYY-MM-DD`), newest date at the bottom.

## 2026-10-28

- db.lokiy.dev: tracker mobs without an NPC ID matched (hub): Zul'Gurub's The Nameless Hermit 25030, Azus the Bloodseeker 25031, Fresh / Corrupted Blood Globe 25035 / 25036 and Caverngloom Crocolisk 25040 added to `logged_npcs.json` (DungeonJournal's ZG combat-log CSVs; Azus 25034 is a Blood Tide helper, left out). Scarlet Sharpshooter is 25230; 25233 (a "Dummy" helper that only casts Explosive Shot) is dropped, so the name matches one NPC. Winterax Envoy has no known ID (not in any log, the creature cache or Wowhead) and keeps its name-only /mob page. 587/588 tracker mobs matched.
