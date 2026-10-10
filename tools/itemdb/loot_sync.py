"""Random affixes ("of the Owl", ...) and Lokiy's own loot (lokiy.dev/loot) for db.lokiy.dev.

- Possible affixes per item: Wowhead's classic gear planner data (assets/db_inputs/wowhead_gearplannerdb.txt):
  item.randomEnchants = random property ids, randomEnchant = {id: {name, effects}}. Classic data: the server can
  have changed an item's affix pool.
- Seen affixes and drops: /var/www/lokiy/loot/loot.json, written by the LootTracker addon's publish_site.py
  (read only). The addon logs the looted name ("Elegant Gloves of the Owl"), not the affix id, so a seen affix is
  matched by name; every tier of that name is marked.
"""
import json
import os
import re

REPO = os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
GEARPLANNER = os.path.join(REPO, "assets", "db_inputs", "wowhead_gearplannerdb.txt")
LOOT_JSON = "/var/www/lokiy/loot/loot.json"


def _page_data(text, name):
    """The JSON object passed to WH.setPageData("<name>", {...})."""
    start = text.index(f'WH.setPageData("{name}"')
    end = text.find("WH.setPageData(", start + 1)
    chunk = text[text.index("{", start):end if end > 0 else len(text)].rstrip().rstrip(";").rstrip().rstrip(")")
    chunk = re.sub(r",(\s*\})\s*$", r"\1", chunk)  # Wowhead leaves a trailing comma after the last entry
    obj, _ = json.JSONDecoder().raw_decode(chunk)
    return obj


def load_random_affixes():
    """{item_id: [(affix name, effects text)]}, tiers of one name sorted by their first number."""
    text = open(GEARPLANNER, encoding="utf-8").read()
    enchants = _page_data(text, "wow.gearPlanner.classic.randomEnchant")
    out = {}
    for item in _page_data(text, "wow.gearPlanner.classic.item").values():
        ids = item.get("randomEnchants")
        if not ids:
            continue
        rows = set()
        for rid in ids:
            e = enchants.get(str(rid))
            if e and e.get("name"):
                rows.add((e["name"], ", ".join(e.get("effects", []))))

        def first_number(r):
            m = re.search(r"\d+", r[1])
            return int(m.group(0)) if m else 0
        out[int(item["id"])] = sorted(rows, key=lambda r: (r[0], first_number(r), r[1]))
    return out


def load_loot():
    """Lokiy's loot tracker: (seen affixes {item_id: {affix name: count}}, drops {item_id: [(source, drops, kills, rate)]},
    generatedAt) or None when the file is missing."""
    try:
        d = json.load(open(LOOT_JSON, encoding="utf-8"))
    except (OSError, ValueError):
        return None
    names = {}  # item id -> {full looted name: count}
    for e in d.get("recentDrops", []):
        if e.get("item") and e.get("name"):
            c = names.setdefault(int(e["item"]), {})
            c[e["name"]] = c.get(e["name"], 0) + int(e.get("quantity") or 1)
    for iid, info in d.get("items", {}).items():  # last looted name; counts only if the log missed it
        if info.get("name"):
            names.setdefault(int(iid), {}).setdefault(info["name"], 1)
    drops = {}
    for kind, rows in (("mob", d.get("mobs", [])), ("chest", d.get("chests", []))):
        for src in rows:
            seen = src.get("kills") if kind == "mob" else src.get("opened")
            for it in src.get("items", []):
                if not it.get("id"):
                    continue
                n = it.get("drops", 0)
                rate = it.get("dropRate")
                if rate is None and seen:
                    rate = n / seen
                drops.setdefault(int(it["id"]), []).append((src["name"], n, seen or 0, rate))
    for v in drops.values():
        v.sort(key=lambda r: -r[1])
    return names, drops, d.get("generatedAt")


def seen_affixes(base_name, looted_names):
    """{affix name: count} from looted names that are the base name plus an affix."""
    out = {}
    for full, n in looted_names.items():
        if full != base_name and full.startswith(base_name + " "):
            affix = full[len(base_name) + 1:]
            out[affix] = out.get(affix, 0) + n
    return out
