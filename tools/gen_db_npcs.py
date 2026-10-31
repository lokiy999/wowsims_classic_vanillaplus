#!/usr/bin/python3
"""assets/database/db_npcs.json = the NPC pages db.lokiy.dev has, for the sim UI's NPC links (ui/core/db_npcs.ts):

  {"ids": {creature id: page}, "names": {name: page}, "quests": [quest ids], "zones": [zone ids]}
  (page = the URL's path on db.lokiy.dev, e.g. "/npc/11502"; quests = the ones with a /quest/<id> page, Part GO;
  zones = the ones with a /zone/<id> page, from its api/zones.json, Part HT)

ids from its api/npcs.json (every NPC with a page); names for the sources the sim shows by name only (bosses without
a creature id in the sim DB, the server's own bosses): its bosses in api/npcs.json with a unique name, then
api/mobs.json (every mob the loot tracker saw, incl. Vanilla+-only ones). The URLs are taken as db.lokiy.dev gives
them. NPCs it doesn't have keep Wowhead's link. Run by tools/itemdb/sync_sim.sh. DB_API overrides the folder.

Usage (from the repo root):
  python3 tools/gen_db_npcs.py
"""
import collections
import json
import os

REPO = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
DB_API = os.environ.get("DB_API", "/var/www/db/api")
OUT = os.path.join(REPO, "assets/database/db_npcs.json")
SITE = "https://db.lokiy.dev"


def page(url):
    return url[len(SITE):] if url.startswith(SITE + "/") else url


def main():
    with open(os.path.join(DB_API, "npcs.json"), encoding="utf-8") as f:
        npcs = json.load(f)
    with open(os.path.join(DB_API, "mobs.json"), encoding="utf-8") as f:
        mobs = json.load(f)
    if len(npcs) < 1000:  # sanity floor: a broken export must not switch every NPC link back to Wowhead
        raise SystemExit(f"only {len(npcs)} NPCs in {DB_API}/npcs.json, {OUT} left as it is")
    ids = {int(k): page(v["url"]) for k, v in npcs.items() if v.get("url")}
    count = collections.Counter(v["name"] for v in npcs.values())
    names = {v["name"]: page(v["url"]) for v in npcs.values() if v.get("boss") and v.get("url") and count[v["name"]] == 1}
    names.update({name: page(url) for name, url in mobs.items()})
    quests = []
    if os.path.exists(os.path.join(DB_API, "quests.json")):
        with open(os.path.join(DB_API, "quests.json"), encoding="utf-8") as f:
            quests = sorted(int(k) for k in json.load(f))
    zones = []
    if os.path.exists(os.path.join(DB_API, "zones.json")):
        with open(os.path.join(DB_API, "zones.json"), encoding="utf-8") as f:
            zones = sorted(int(k) for k in json.load(f))
    with open(OUT, "w", encoding="utf-8") as f:
        json.dump({"ids": {str(k): ids[k] for k in sorted(ids)}, "names": dict(sorted(names.items())), "quests": quests,
                   "zones": zones},
                  f, ensure_ascii=False, separators=(",", ":"))
        f.write("\n")
    print(f"wrote {OUT}: {len(ids)} NPC ids, {len(names)} names, {len(quests)} quests, {len(zones)} zones")


if __name__ == "__main__":
    main()
