"""NPCs for db.lokiy.dev's /npc/<creature id> pages (sim repo Part GJ, 2026-10-25).

Creature IDs, names, levels, ranks, zones, classic drops and vendor items come from pfQuest's classic database
(pfquest_npcs.json, import_pfquest.py). Lokiy's loot tracker and AtlasLoot only know NPCs by name: match_npc() maps a
name to one creature ID only when the evidence says so (a unique name, else the one in the right zone, else the one
pfQuest says drops the items seen); every other name is reported, not guessed. Vanilla+'s own NPCs aren't in pfQuest
and keep their name-based /mob pages.
"""
import json
import os
import re

HERE = os.path.dirname(os.path.abspath(__file__))
RANKS = {1: "Elite", 2: "Rare elite", 3: "Boss", 4: "Rare"}
# pfQuest units that are no real NPC (GM waypoints, triggers, placeholders)
JUNK = re.compile(r"\(Only GM|\bDND\b|\[DND\]|Trigger|\[PH\]|^zz|^OLD |Dummy|\bTEST\b|\(UNUSED\)", re.I)


def norm_zone(z):
    return re.sub(r"\s*\(.*?\)\s*|^(lower|upper)\s+", "", (z or "").lower()).strip()


def load_npcs():
    """{creature id: {"name", "lvl", "zone", "rank", "fac", "drops": [[item, %]], "sells": [items]}} without junk."""
    path = os.path.join(HERE, "pfquest_npcs.json")
    if not os.path.exists(path):
        return {}
    raw = json.load(open(path, encoding="utf-8"))
    return {int(k): v for k, v in raw.items() if not JUNK.search(v["name"])}


class NpcMatcher:
    def __init__(self, npcs):
        self.npcs = npcs
        self.by_name = {}
        for nid, n in npcs.items():
            self.by_name.setdefault(n["name"].lower(), []).append(nid)
        self.unmatched = {}  # name -> why (for Lokiy)

    def match(self, name, zones=(), items=()):
        """The creature id of an NPC known by name, or None (and the reason in self.unmatched)."""
        cands = self.by_name.get((name or "").lower(), [])
        if not cands:
            self.unmatched[name] = "not in pfQuest (a Vanilla+ NPC, or named differently)"
            return None
        if len(cands) == 1:
            return cands[0]
        zs = {norm_zone(z) for z in zones if z}
        by_zone = [c for c in cands if zs and norm_zone(self.npcs[c]["zone"]) in zs]
        if len(by_zone) == 1:
            return by_zone[0]
        pool = by_zone or cands
        items = set(items)
        by_drops = [c for c in pool if items & {d[0] for d in self.npcs[c].get("drops", [])}]
        if len(by_drops) == 1:
            return by_drops[0]
        self.unmatched[name] = f"{len(cands)} pfQuest NPCs with this name ({', '.join(map(str, sorted(cands)[:6]))})"
        return None
