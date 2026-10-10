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
CHANCES = os.path.join(os.path.dirname(os.path.abspath(__file__)), "affix_chances.json")


def _page_data(text, name):
    """The JSON object passed to WH.setPageData("<name>", {...})."""
    start = text.index(f'WH.setPageData("{name}"')
    end = text.find("WH.setPageData(", start + 1)
    chunk = text[text.index("{", start):end if end > 0 else len(text)].rstrip().rstrip(";").rstrip().rstrip(")")
    chunk = re.sub(r",(\s*\})\s*$", r"\1", chunk)  # Wowhead leaves a trailing comma after the last entry
    obj, _ = json.JSONDecoder().raw_decode(chunk)
    return obj


# Wowhead random enchant stat keys -> item_stats.py keys (for the search page)
AFFIX_STAT = {
    "str": "str", "agi": "agi", "sta": "sta", "int": "int", "spi": "spi", "armor": "armor", "def": "def",
    "mleatkpwr": "ap", "rgdatkpwr": "rap", "splheal": "heal", "spldmg": "sp", "blockpct": "blockc", "dodgepct": "dodge",
    "manargn": "mp5", "healthrgn": "hp5", "skillBuff": "skill", "mlecritstrkpct": "crit",
    "arcres": "arcres", "firres": "firres", "frores": "frores", "natres": "natres", "shares": "shares",
    "arcsplpwr": "arcsp", "firsplpwr": "firsp", "frosplpwr": "frosp", "holsplpwr": "holsp", "natsplpwr": "natsp",
    "shasplpwr": "shasp",
}


def merge_tiers(effects):
    """The tiers of one affix name as one line: ["+12 Intellect, +13 Spirit", "+13 Intellect, +12 Spirit"] ->
    "+12-13 Intellect, +12-13 Spirit"; values with a gap are listed: "+15/17 Fire Resistance" (Lokiy). Effects that
    are not "+N Stat" are kept as they are."""
    values, order, other = {}, [], []
    for line in effects:
        for part in [p.strip() for p in line.split(",") if p.strip()]:
            m = re.match(r"^\+(\d+) (.+)$", part)
            if not m:
                if part not in other:
                    other.append(part)
                continue
            v, stat = int(m.group(1)), m.group(2)
            if stat not in values:
                order.append(stat)
                values[stat] = set()
            values[stat].add(v)
    out = []
    for stat in order:
        vs = sorted(values[stat])
        if len(vs) == 1:
            txt = str(vs[0])
        elif vs[-1] - vs[0] == len(vs) - 1:
            txt = f"{vs[0]}-{vs[-1]}"
        else:
            txt = "/".join(str(v) for v in vs)
        out.append(f"+{txt} {stat}")
    return ", ".join(out + other)


def chance_text(names, chances, n_names):
    """Roll chance of getting one of these affix names, as display text: Wowhead's chances added up ("13.2%"); "<4.6%"
    when Wowhead lists other affixes of the item but not these (it only lists the ones it has seen); "~7%" (share of the
    item's affix names) when there is no Wowhead chance data for the item."""
    if chances:
        listed = [chances[n] for n in names if n in chances]
        if listed:
            return f"{round(sum(listed), 1):g}%"
        return f"<{min(chances.values()):g}%"
    return f"~{round(len(names) / (n_names or 1) * 100)}%"


def load_random_affixes():
    """({item_id: [(affix name, effects text)]}, {item_id: {stat: [max value, chance text]}}, {item_id: {name: %}}).

    Tiers of one name are sorted by their first number. Chances: Wowhead's per item affix roll chances
    (affix_chances.json, from fetch_affix_chances.py), see chance_text().
    """
    text = open(GEARPLANNER, encoding="utf-8").read()
    enchants = _page_data(text, "wow.gearPlanner.classic.randomEnchant")
    try:
        all_chances = {int(k): v for k, v in json.load(open(CHANCES, encoding="utf-8")).items()}
    except (OSError, ValueError):
        all_chances = {}
    out, stats = {}, {}
    for item in _page_data(text, "wow.gearPlanner.classic.item").values():
        ids = item.get("randomEnchants")
        if not ids:
            continue
        rows = set()
        by_stat = {}  # stat -> (max value, {affix names})
        for rid in ids:
            e = enchants.get(str(rid))
            if e and e.get("name"):
                rows.add((e["name"], ", ".join(e.get("effects", []))))
                for k, v in e.get("stats", {}).items():
                    key = AFFIX_STAT.get(k)
                    if key and v:
                        mx, nm = by_stat.get(key, (0, set()))
                        by_stat[key] = (max(mx, v), nm | {e["name"]})
        n_names = len({r[0] for r in rows})
        chances = all_chances.get(int(item["id"]), {})
        if by_stat:
            stats[int(item["id"])] = {k: [mx, chance_text(nm, chances, n_names)] for k, (mx, nm) in by_stat.items()}

        def first_number(r):
            m = re.search(r"\d+", r[1])
            return int(m.group(0)) if m else 0
        out[int(item["id"])] = sorted(rows, key=lambda r: (r[0], first_number(r), r[1]))
    return out, stats, all_chances


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
