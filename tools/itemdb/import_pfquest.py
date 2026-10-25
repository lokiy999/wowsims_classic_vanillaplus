"""One-time import of pfQuest's classic database (vmangos data, not Vanilla+; PC: data\\pfquest, pfQuest 7.0.1) into
pfquest_sources.json: which NPCs drop each item (with the classic drop chance), which vendors sell it and which
objects (chests, ore, ...) hold it. db.lokiy.dev uses it only for items no own data (AtlasLoot, the workbook import,
the loot tracker) has a source for, marked "classic data (pfQuest)".

    python3 import_pfquest.py <pfquest folder>

Writes {"items": {item id: {"drops": [[npc id, name, zone, chance %, level]], "n": NPCs in all, "lvl": [min, max]
level of those NPCs, "vendors": [[npc id, name, zone]], "objects": [[object id, name, zone]]}}, "npcs": {npc id:
[name, zone]}, "zones": {area id: name}}; drops by chance, at most MAX_LIST of each. Drops through a reference loot
table (pfQuest "R", shared world-drop pools) count in "n" without a chance. "npcs" / "zones" name the NPCs and zones
of the Wowhead sources.
"""
import json
import os
import re
import sys
from collections import Counter

HERE = os.path.dirname(os.path.abspath(__file__))
MAX_LIST = 8
TOKEN = re.compile(r'\s*(?:(--[^\n]*)|(\{)|(\})|(\[)|(\])|(=)|(,)|("(?:[^"\\]|\\.)*")|(-?\d+(?:\.\d+)?)|(\w+))')


def parse_lua(path):
    """{(key1, key2): table} for every pfDB["key1"]["key2"] = { ... } in the file."""
    txt = open(path, encoding="utf-8", errors="replace").read()
    res = {}
    for m in re.finditer(r'pfDB\["(\w+)"\]\["(\w+)"\]\s*=\s*', txt):
        val, _ = parse_value(txt, m.end())
        res[(m.group(1), m.group(2))] = val
    return res


def parse_value(txt, pos):
    m = TOKEN.match(txt, pos)
    while m.group(1):  # comment
        pos = m.end()
        m = TOKEN.match(txt, pos)
    if m.group(2):
        return parse_table(txt, m.end())
    if m.group(8):
        return m.group(8)[1:-1].replace("\\'", "'").replace('\\"', '"'), m.end()
    if m.group(9):
        n = m.group(9)
        return (float(n) if "." in n else int(n)), m.end()
    return m.group(10), m.end()


def parse_table(txt, pos):
    """A Lua table: dict when it has keys, list when it has none."""
    d, lst = {}, []
    while True:
        m = TOKEN.match(txt, pos)
        if m.group(1) or m.group(7):
            pos = m.end()
            continue
        if m.group(3):
            return (d if d else lst), m.end()
        if m.group(4):
            key, pos = parse_value(txt, m.end())
            pos = TOKEN.match(txt, pos).end()  # ]
            pos = TOKEN.match(txt, pos).end()  # =
            d[key], pos = parse_value(txt, pos)
        else:
            v, pos = parse_value(txt, pos)
            lst.append(v)


def main(folder):
    def load(name):
        return parse_lua(os.path.join(folder, name))

    items = load("items.lua")[("items", "data")]
    units = load("units.lua")[("units", "data")]
    objects = load("objects.lua")[("objects", "data")]
    refloot = load("refloot.lua")[("refloot", "data")]
    unit_names = load(os.path.join("enUS", "units.lua"))[("units", "enUS")]
    obj_names = load(os.path.join("enUS", "objects.lua"))[("objects", "enUS")]
    zone_names = load(os.path.join("enUS", "zones.lua"))[("zones", "enUS")]

    def zone(entry):
        coords = (entry or {}).get("coords") or []
        zs = Counter(c[2] for c in (coords.values() if isinstance(coords, dict) else coords) if len(c) > 2)
        return zone_names.get(zs.most_common(1)[0][0], "") if zs else ""

    def level(uid):
        lv = re.findall(r"\d+", str((units.get(uid) or {}).get("lvl", "")))
        return [int(x) for x in lv]

    out = {}
    for iid, e in items.items():
        if not isinstance(e, dict) or not e:
            continue
        rec = {}
        drops = dict(e.get("U") or {})
        all_units = set(drops)
        for ref in (e.get("R") or {}):
            all_units |= set((refloot.get(ref) or {}).get("U") or {})
        if drops:
            rows = sorted(drops.items(), key=lambda kv: -kv[1])[:MAX_LIST]
            rec["drops"] = [[u, unit_names.get(u, f"NPC {u}"), zone(units.get(u)), c,
                             "-".join(str(x) for x in sorted(set(level(u))))] for u, c in rows]
        if all_units:
            rec["n"] = len(all_units)
            lv = [x for u in all_units for x in level(u)]
            if lv:
                rec["lvl"] = [min(lv), max(lv)]
        if e.get("V"):
            rec["vendors"] = [[u, unit_names.get(u, f"NPC {u}"), zone(units.get(u))]
                              for u in sorted(e["V"])[:MAX_LIST]]
        if e.get("O"):
            rows = sorted(e["O"].items(), key=lambda kv: -kv[1])[:MAX_LIST]
            rec["objects"] = [[o, obj_names.get(o, f"Object {o}"), zone(objects.get(o))] for o, _ in rows]
        if rec:
            out[str(iid)] = rec
    npcs = {str(u): [n, zone(units.get(u))] for u, n in unit_names.items() if isinstance(u, int)}
    with open(os.path.join(HERE, "pfquest_sources.json"), "w", encoding="utf-8") as f:
        json.dump({"items": out, "npcs": npcs, "zones": {str(z): n for z, n in zone_names.items()}}, f,
                  ensure_ascii=False, separators=(",", ":"), sort_keys=True)
    print(f"{len(out)} items: {sum('drops' in r for r in out.values())} with NPC drops, "
          f"{sum('vendors' in r for r in out.values())} sold, {sum('objects' in r for r in out.values())} in objects")


if __name__ == "__main__":
    main(sys.argv[1])
