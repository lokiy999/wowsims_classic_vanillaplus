#!/usr/bin/python3
"""Season of Discovery doesn't exist on Vanilla+ (Lokiy 2026-10-31, sim Part HV). Takes it out of the sim's inputs:

- assets/db_inputs/atlasloot_db.json (the sim's NPC / zone roster, made once from the SoD-era AtlasLoot data):
  every SoD NPC (creature id >= SOD_NPC_MIN) becomes the classic NPC of the same name that db.lokiy.dev has a page
  for (its api/npcs.json; e.g. Lady Sarevess 204068 -> 4831), or goes when there is none (SoD-only encounters like
  "Dreamscythe and Weaver"); the SoD zone Demon Fall Canyon goes; item sources naming either follow; SoD items
  (item id >= SOD_ITEM_MIN) go.
- tools/database/gen_db/main.go: SoD spell ids (>= SOD_SPELL_MIN; Vanilla+'s own spells end below 40000) out of the
  hard-coded per-class spell lists.

Run once from the repo root; running it again changes nothing. DB_API overrides db.lokiy.dev's api folder.
"""
import json
import os
import re

REPO = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
DB_API = os.environ.get("DB_API", "/var/www/db/api")
ATLAS = os.path.join(REPO, "assets/db_inputs/atlasloot_db.json")
GEN_DB = os.path.join(REPO, "tools/database/gen_db/main.go")
SOD_NPC_MIN = 200000  # as gen_site.py
SOD_SPELL_MIN = 400000
SOD_ITEM_MIN = 200000  # Vanilla+'s items end below 90000 (db.lokiy.dev api/items.json)
SOD_ZONES = {15475}  # Demon Fall Canyon


def strip_atlas():
    with open(ATLAS, encoding="utf-8") as f:
        db = json.load(f)
    with open(os.path.join(DB_API, "npcs.json"), encoding="utf-8") as f:
        pages = json.load(f)
    by_name = {}
    for k, v in pages.items():
        by_name.setdefault(v["name"], []).append(int(k))
    have = {n["id"] for n in db["npcs"]}
    remap, gone, npcs = {}, [], []
    for n in db["npcs"]:
        if n["id"] < SOD_NPC_MIN and n.get("zoneId") not in SOD_ZONES:
            npcs.append(n)
            continue
        classic = by_name.get(n["name"], [])
        if len(classic) == 1 and classic[0] < SOD_NPC_MIN and n.get("zoneId") not in SOD_ZONES:
            remap[n["id"]] = classic[0]
            if classic[0] not in have:
                npcs.append({**n, "id": classic[0]})
                have.add(classic[0])
        else:
            gone.append(n["id"])
    db["npcs"] = sorted(npcs, key=lambda n: n["id"])
    db["zones"] = [z for z in db["zones"] if z["id"] not in SOD_ZONES]
    n_items = len(db["items"])
    db["items"] = [it for it in db["items"] if it["id"] < SOD_ITEM_MIN]
    n_items -= len(db["items"])
    n_src = 0
    for it in db["items"]:
        out = []
        for s in it.get("sources", []):
            d = s.get("drop")
            if d and (d.get("npcId") in remap):
                s = {**s, "drop": {**d, "npcId": remap[d["npcId"]]}}
                n_src += 1
            elif d and (d.get("npcId") in gone or d.get("zoneId") in SOD_ZONES):
                n_src += 1
                continue
            out.append(s)
        if "sources" in it:
            it["sources"] = out
    with open(ATLAS, "w", encoding="utf-8") as f:  # gen_db's layout: one entry per line
        f.write("{\n" + ",\n".join(f'"{k}":[\n' + ",\n".join(json.dumps(e, ensure_ascii=False, separators=(",", ":"))
                                                              for e in v) + ("\n" if v else "") + "]"
                                   for k, v in db.items()) + "\n}")
    print(f"atlasloot_db.json: {len(remap)} SoD NPCs -> classic ones, {len(gone)} removed, {n_items} SoD items "
          f"removed, {n_src} item sources changed")


def strip_gen_db():
    with open(GEN_DB, encoding="utf-8") as f:
        src = f.read()
    n = 0

    def clean(m):
        nonlocal n
        ids = [x.strip() for x in m.group(2).split(",") if x.strip()]
        keep = [x for x in ids if int(x) < SOD_SPELL_MIN]
        n += len(ids) - len(keep)
        return m.group(1) + ", ".join(keep) + "}"

    src = re.sub(r'(ret_db\["[a-z ]+ wowhead"\] = \[\]int32\{)([0-9, ]*)\}', clean, src)
    with open(GEN_DB, "w", encoding="utf-8") as f:
        f.write(src)
    print(f"gen_db/main.go: {n} SoD spell ids out of the class spell lists")


if __name__ == "__main__":
    strip_atlas()
    strip_gen_db()
