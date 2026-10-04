"""Collect kills and loot events from Chronicle combat logs (CustomData/Chronicle_*.txt) for db.lokiy.dev.

No player names are written. Output observed_loot.json:
  {"logs": n,
   "kills": {creature: n}, "zone_kills": {zone: n}, "creature_zone": {creature: zone},
   "loots": {item_id: [[[creature, seconds_before_loot, zone], ...], ...]}}   # one entry per loot event
Each loot event lists the creatures that died in the WINDOW seconds before it (newest first, one per name), so the site
generator can credit the loot using AtlasLoot's known sources. Events seen in several logs (merged logs, several loggers
in one raid) are counted once: deaths by GUID within DEDUP sec, loots by (item, looter) within DEDUP sec.
"""
import glob
import json
import os
import re
from collections import defaultdict

LOGS = r"C:\Programs\WoW_VPlus_Client\World of Warcraft Vanilla+\CustomData"
WINDOW = 600.0
DEDUP = 15.0
ITEM_RE = re.compile(r"Hitem:(\d+):")

deaths, loots = [], []
files = sorted(glob.glob(os.path.join(LOGS, "Chronicle_*.txt")))
for path in files:
    names, zone, recent = {}, "", []
    for raw in open(path, encoding="utf-8", errors="replace"):
        p = raw.rstrip("\n").split("|")
        if len(p) < 3:
            continue
        try:
            t = int(p[0]) / 1000.0
        except ValueError:
            continue
        ev = p[1]
        if ev == "ZONE_INFO":
            zone = p[2]
        elif ev == "UNIT_INFO" and p[2].startswith("0xF130") and len(p) > 4 and p[4] and p[4] != "Unknown":
            names[p[2]] = p[4]
        elif ev == "DEATH" and p[2].startswith("0xF130"):
            name = names.get(p[2])
            if name:
                deaths.append((t, p[2], name, zone))
                recent.append((t, name, zone))
                recent = [r for r in recent if t - r[0] <= WINDOW]
        elif ev == "LOOT":
            m = ITEM_RE.search(raw)
            if not m:
                continue
            cands, seen = [], set()
            for (tt, name, z) in reversed(recent):
                if 0 <= t - tt <= WINDOW and name not in seen:
                    seen.add(name)
                    cands.append([name, round(t - tt, 1), z])
            loots.append((t, int(m.group(1)), p[2], cands[:40]))

deaths.sort()
seen_death, kills, zone_kills, creature_zone = {}, defaultdict(int), defaultdict(int), {}
for t, guid, name, zone in deaths:
    if guid in seen_death and t - seen_death[guid] <= DEDUP:
        continue
    seen_death[guid] = t
    kills[name] += 1
    if zone:
        zone_kills[zone] += 1
        creature_zone.setdefault(name, zone)

loots.sort(key=lambda x: x[0])
seen_loot, out_loots = {}, defaultdict(list)
for t, item, looter, cands in loots:
    k = (item, looter)
    if k in seen_loot and t - seen_loot[k] <= DEDUP:
        continue
    seen_loot[k] = t
    out_loots[item].append(cands)

json.dump({"logs": len(files), "kills": kills, "zone_kills": zone_kills, "creature_zone": creature_zone,
           "loots": out_loots}, open("observed_loot.json", "w"), separators=(",", ":"))
print(f"{len(files)} logs, {sum(kills.values())} kills of {len(kills)} creatures in {len(zone_kills)} zones, "
      f"{sum(len(v) for v in out_loots.values())} loot events of {len(out_loots)} items")
