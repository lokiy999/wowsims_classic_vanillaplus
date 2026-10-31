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
# test / placeholder NPCs; "zz" and "OLD " only in their dev spelling (Part HU: Old Icebeard, Zzarc' Vul are real)
JUNK = re.compile(r"\(Only GM|\bDND\b|\[DND\]|Trigger|\[PH\]|(?-i:^zz)|(?-i:^OLD )|Dummy|\bTEST\b|\(UNUSED\)|"
                  r"\[UNUSED\]", re.I)


def norm_zone(z):
    return re.sub(r"\s*\(.*?\)\s*|^(lower|upper)\s+", "", (z or "").lower()).strip()


def load_logged():
    """{creature id: {"name", "zone", "boss"?, "aliases"?}}: creature ids seen in combat logs (logged_npcs.json).
    They beat a pfQuest name match; the ones pfQuest doesn't have (Vanilla+'s own, e.g. the SM raid 25200-25264) get
    their own /npc page from our data (Part GL)."""
    path = os.path.join(HERE, "logged_npcs.json")
    if not os.path.exists(path):
        return {}
    return {int(k): v for k, v in json.load(open(path, encoding="utf-8"))["npcs"].items()}


def load_npcs():
    """{creature id: {"name", "lvl", "zone", "rank", "fac", "drops": [[item, %]], "sells": [items]}} without junk."""
    path = os.path.join(HERE, "pfquest_npcs.json")
    if not os.path.exists(path):
        return {}
    raw = json.load(open(path, encoding="utf-8"))
    return {int(k): v for k, v in raw.items() if not JUNK.search(v["name"])}


def load_journal():
    """{creature id: [(raid name, entry, kind, boss name)]} from DungeonJournal's export (dungeonjournal.json, uploaded
    by its tools/export_json.py): kind = "boss", "add" (boss = the boss it belongs to) or "trash" (Part GN)."""
    path = os.path.join(HERE, "dungeonjournal.json")
    if not os.path.exists(path):
        return {}, {}
    data = json.load(open(path, encoding="utf-8"))
    out = {}

    def put(raid, e, kind, boss=""):
        for nid in e.get("npcIds") or ([e["npcId"]] if e.get("npcId") else []):
            out.setdefault(int(nid), []).append((raid, e, kind, boss))

    for r in data.get("raids", []):
        for b in r.get("bosses", []):
            put(r["name"], b, "boss")
            for a in b.get("adds", []):
                put(r["name"], a, "add", b["name"])
        for t in r.get("trash", []):
            if not t.get("separator"):
                put(r["name"], t, "trash")
    return out, data.get("flags", {})


class NpcMatcher:
    def __init__(self, npcs, logged=None):
        self.npcs = npcs
        self.by_name = {}
        for nid, n in npcs.items():
            if not n.get("vplus"):
                self.by_name.setdefault(n["name"].lower(), []).append(nid)
        self.logged = {}  # (name, zone) -> [creature ids seen in combat logs]
        for nid, r in (logged or {}).items():
            for nm in [r["name"]] + r.get("aliases", []):
                self.logged.setdefault((nm.lower(), norm_zone(r["zone"])), []).append(nid)
        self.unmatched = {}  # name -> why (for Lokiy)

    def match(self, name, zones=(), items=()):
        """The creature id of an NPC known by name, or None (and the reason in self.unmatched). A creature id from a
        combat log in that zone comes first, then pfQuest."""
        if not zones:  # no zone known (e.g. a tracked mob nobody looted): a name only one logged creature has
            ids = {i for (nm, _), v in self.logged.items() if nm == (name or "").lower() for i in v}
            if len(ids) == 1:
                return ids.pop()
        for z in zones:
            ids = self.logged.get(((name or "").lower(), norm_zone(z)), [])
            if len(ids) == 1:
                return ids[0]
            if ids:
                self.unmatched[name] = f"several creatures of this name in the {z} combat logs ({', '.join(map(str, ids))})"
                return None
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
