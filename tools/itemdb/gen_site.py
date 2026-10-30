#!/usr/bin/python3
"""Static item database site for db.lokiy.dev (all Vanilla Plus items).

Reads the server data in the repo's CSV's/ folder:
  - VPlusItemDB.lua: in-game tooltip lines of every server item (dump of 2026-08-26)
  - ~/itemcache/items.json: the game client's item cache (itemcache_sync.py), the server's current data; its items
    replace the dump's and add the ones the dump doesn't have
  - AtlasLoot (the server's fork): loot tables, their boss/instance names and item icons
and writes plain HTML pages (Caddy serves them from disk):
  /index.html             search + browse by instance / boss
  /item/<id>.html         one page per item (tooltip, icon, where it drops), with Open Graph tags for link previews
  /loot/<table>.html      one page per loot table (boss, trash, set, ...) that has custom or reworked items
  /reworked.html          reworked classic items by instance / boss
  /instances.html         every loot table by instance
  /type/<type>.html       every item by type (item_types.py)
  /search.html            filter all items (search.js over /search.json; stats from item_stats.py)
  /tables.json            the loot tables for the header search (site.js: items, NPCs, spells, quests, zones,
                          loot tables, grouped; Part GQ)
  /items.json             search index; /tt/<id // 500>.json hover tooltips
  /api/items.json         every item for other sites (lokiy.dev/loot): name, quality, icon URL, tooltip lines with
                          their colour class, custom flag, random affixes. db.lokiy.dev is the only item source.
  /spell/<id>.html        every Vanilla+ spell (spells.py: lokiy.dev/spells' database, its classic.json diff against
                          stock 1.12.1), the items that use it; /spells.html a search over /spells.json
  /api/spells.json        every spell: name, rank, icon, description, buff text, status (new / changed / same),
                          cost / range / cast / cd tooltip lines (empty fields left out)
  (items no Vanilla+ client has seen: classic Wowhead tooltips, tagged "not seen on Vanilla+ yet", api flag
   "classicOnly"; add_classic_items, Part GF)
  /npc/<id>.html          every NPC pfQuest knows (npcs.py): level, rank, zone, DungeonJournal abilities / adds /
                          tactics (dungeonjournal.json, uploaded by DungeonJournal), Lokiy's loot, its AtlasLoot table,
                          classic drops / vendor items (pfQuest); /npcs.html a search over /npcs.json
  /api/npcs.json          {creature id: {name, level, zone, boss, url}}
  /quest/<id>.html, /quests.html, /zone/<id>.html, /zones.html, /api/quests.json, /api/zones.json: quests.py (Part GO)
  /api/mobs.json          {mob name: URL of its page} for every mob the loot tracker saw (lokiy.dev/loot links there;
                          /npc/<id> when the name matched a creature id)
  /api/sources.json       where items come from, for other projects (the sim): every AtlasLoot table with its name,
                          instance, folder and items (incl. the drops the loot tracker added), and the Wowhead classic
                          sources (wowhead_sources.json) for what AtlasLoot doesn't cover. The sim's only source data.
                          "items": every item's sources by layer (item_sources.py: AtlasLoot, the item workbook, the
                          loot tracker, pfQuest, Wowhead), each {"text", "type", "from"}; "notOnServer": {item id:
                          why} for items not on Vanilla+ (AQ, Naxxramas, Alterac Valley, Tier 2, ...).
                          "webdkp": guild raid loot awards per item (webdkp_awards.json): count, first / last date.
  /style.css, /site.js

Scope (Lokiy 2026-10-04): custom items = item ids that do not exist in classic 1.12 (above 24283) plus the items
AtlasLoot flags as new ("N"); reworked items = classic ids whose server tooltip differs from Wowhead's classic one
(reworked.py). Every other server item gets a page too (Lokiy 2026-10-04: "all other items as well").

Usage:
  python3 tools/itemdb/gen_site.py [--out /var/www/db]
"""
import argparse
import glob
import html
import json
import os
import re
import shutil
import sys
from datetime import date, datetime

REPO = os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
sys.path.insert(0, os.path.join(REPO, "docs"))
from parse_vplus import parse_lua  # noqa: E402
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from reworked import RECIPE, find_reworked, load_wowhead, norm  # noqa: E402
from item_types import ARMOR_SLOTS, PAGES, classify  # noqa: E402
from item_stats import ALL_CLASSES, CLASSES, STAT_LABELS, parse as parse_stats  # noqa: E402
from loot_sync import (load_affix_ids, load_loot, load_mob_drops, load_mob_loot, load_random_affixes, load_rolls,  # noqa: E402
                       merge_tiers, seen_affixes)
from wdb_tooltip import TRIGGERS, Lookups, tooltip  # noqa: E402
from npcs import RANKS, NpcMatcher, load_journal, load_logged, load_npcs  # noqa: E402
from quests import Quests  # noqa: E402
from spells import MAIN_FIELDS, STATUS_LABEL, link_tokens, load_spells, tooltip_lines  # noqa: E402
from affixes import Affixes  # noqa: E402
from item_sources import (CUT_TEXT, FROM_LABEL, atlas_item_sources, merge_atlas_sources, LOKIY_DROPPED_IDS, SHARED_WB_BOSSES, SHARED_WB_KEY, TO_CHECK, TYPE_LABEL,  # noqa: E402
                          Sources, add_shared_world_boss_loot)

CSV = os.path.join(REPO, "CSV's")
ATLAS = os.path.join(CSV, "AtlasLoot")
ITEMCACHE = os.path.expanduser("~/itemcache/items.json")
SPELL_DB = "/var/www/lokiy/spells/db"
CLASSIC_MAX_ITEM_ID = 24283
TRACKED_MIN_QUALITY = 3  # rare+: lower drops are mostly trade goods / trash loot every mob of the zone has
SITE_NAME = "Vanilla Plus Database"
SITE_URL = "https://db.lokiy.dev"
QUALITY = {"Poor": 0, "Common": 1, "Uncommon": 2, "Rare": 3, "Epic": 4, "Legendary": 5, "Artifact": 6}
Q_COLOR = {0: "#9d9d9d", 1: "#ffffff", 2: "#1eff00", 3: "#0070dd", 4: "#a335ee", 5: "#ff8000", 6: "#e6cc80"}
GREEN = ("Equip:", "Use:", "Chance on hit:", "(", "Set:")
HERE = os.path.dirname(os.path.abspath(__file__))
# Server-made icons extracted from the client's patch MPQs (<name>.jpg 56px, <name>_s.jpg 18px): the first 42 by hand
# in tools/itemdb/icons/ (git), the rest uploaded by Lokiy's PC to CSV's/icons/ (ItemCacheSync upload_itemcache.py,
# asks itemcache_sync.py --missing-icons). Everything else comes from the public icon CDN.
ICON_DIRS = [os.path.join(HERE, "icons"), os.path.join(os.path.dirname(os.path.dirname(HERE)), "CSV's", "icons")]
LOCAL_ICONS = {f[:-4] for d in ICON_DIRS if os.path.isdir(d) for f in os.listdir(d)
               if f.endswith(".jpg") and not f.endswith("_s.jpg")}
# Typos in the server's AtlasLoot icon names.
ICON_ALIAS = {
    "inv_brancer_11": "inv_bracer_11",
    "inv_misc bomb_03": "inv_misc_bomb_03",
    "inv_misc_armorkit 27": "inv_misc_armorkit_27",
    "inv_throwing knife_01": "inv_throwingknife_01",
}

esc = html.escape
# Blizzard / server test and placeholder items: they keep their page but stay out of search and browse lists.
DEV_ITEM = re.compile(r"^QA|\(Test\)|\bTest (Item|Bow|Crit|Defense|Stationery)|TEST ITEM|Test Beatdown|"
                      r"Properties Test|^Deprecated |^Monster - |\[PH\]|\[DEP\]|\bUNUSED\b|^Unused |^DEBUG ")
# type page sections: armor slots in equipment order, the rest alphabetical
SECTION_ORDER = {s: n for n, s in enumerate(ARMOR_SLOTS)}


def slug(s):
    return re.sub(r"[^a-z0-9]+", "-", s.lower()).strip("-")


def icon_url(icon, size="large", absolute=False):
    icon = ICON_ALIAS.get(icon, icon or "inv_misc_questionmark")
    local = icon.replace(" ", "_")
    if local not in LOCAL_ICONS:  # every client icon is local (Part FS); an unknown name gets the question mark
        local = "inv_misc_questionmark"
    return f"{SITE_URL if absolute else ''}/icons/{local}{'_s' if size == 'small' else ''}.jpg"


# --------------------------------------------------------------------------------------------- data

def load_items():
    """{id: {name, quality, lines (tooltip lines without the Rarity line)}}"""
    raw = parse_lua(os.path.join(CSV, "VPlusItemDB.lua"))
    items = {}
    for iid, lines in raw.items():
        quality = 1
        keep = []
        for ln in lines:
            m = re.match(r"^Rarity: (\w+)$", ln)
            if m:
                quality = QUALITY.get(m.group(1), 1)
            else:
                keep.append(ln)
        if keep:
            items[iid] = {"id": iid, "name": keep[0], "quality": quality, "lines": keep[1:]}
    return items


# names of classic leftovers that were never real items (on top of DEV_ITEM)
CLASSIC_JUNK = re.compile(r"^(OLD|Deptecated |NPC Equip|zz)|\((old\d*|test|TEST)\)|\bTEST\b|DOOM!", re.I)
WOWHEAD_ONLY_LINES = re.compile(r"^(Item Level \d+|Durability \d+ / \d+|Sell Price:.*|Max Stack:.*)$")


def add_classic_items(items, wowhead, tables):
    """Classic 1.12 items no Vanilla+ client has seen yet: not in the VPlusItemDB dump nor the item cache (Lokiy
    2026-10-25: a classic item exists on Vanilla+ unless it's marked not on Vanilla+; e.g. Serpent's Kiss, Blade of
    Hanna). From Wowhead's classic tooltips (wowhead_item_tooltips.csv), only classic ids (<= 24283; higher Wowhead ids
    are later Classic items, and Vanilla+ uses that range for its own) that had a classic source (AtlasLoot, pfQuest,
    the Wowhead sources) and aren't placeholders. Tagged "classic_only"; once the item cache sees one, its real server
    data is used instead (add_cache_items runs first). Returns the ids added."""
    known = set()
    for name in ("pfquest_sources.json", "wowhead_sources.json"):
        path = os.path.join(HERE, name)
        if os.path.exists(path):
            d = json.load(open(path, encoding="utf-8"))
            known |= {int(i) for i in (d["items"] if name.startswith("pfquest") else d)}
    known |= {r[0] for rows in tables.values() for r in rows if r[0]}
    added = []
    for iid, w in wowhead.items():
        if iid > CLASSIC_MAX_ITEM_ID or iid in items or iid not in known or not w["lines"]:
            continue
        if DEV_ITEM.search(w["name"]) or CLASSIC_JUNK.search(w["name"]):
            continue
        # (Wowhead writes "(3) Set :"; the game and the set-bonus rule in item_sources.py use "(3) Set:")
        lines = [re.sub(r"^\((\d+)\) Set :", r"(\1) Set:", ln) for ln in w["lines"][1:] if not WOWHEAD_ONLY_LINES.match(ln)]
        ilvl = next((int(ln.split()[-1]) for ln in w["lines"] if re.match(r"^Item Level \d+$", ln)), 0)
        items[iid] = {"id": iid, "name": w["name"], "quality": w["quality"] if w["quality"] is not None else 1,
                      "lines": lines, "classic_only": True, "ilvl": ilvl}
        added.append(iid)
    return added


def add_cache_items(items):
    """Items from the game client's item cache (the server's current data) replace the dump's tooltips.
    Returns {id: icon} from the items' display ids."""
    try:
        with open(ITEMCACHE, encoding="utf-8") as f:
            store = json.load(f)["items"]
    except (OSError, ValueError):
        return {}
    recs = {int(k): v["item"] for k, v in store.items()}
    lk = Lookups(SPELL_DB, os.path.join(CSV, "dbc"))
    for iid, rec in recs.items():
        lines = tooltip(rec, lk, recs)
        items[iid] = {"id": iid, "name": lines[0], "quality": rec["quality"], "lines": lines[1:-1]}
        # tooltip line -> spell id (Use / Equip / Chance on hit, set bonuses), for the links to /spell/<id>
        spell_lines = {TRIGGERS[sp["trigger"]] + lk.desc(sp["id"]): sp["id"] for sp in rec["spells"]
                       if sp["id"] and sp["trigger"] in TRIGGERS and lk.desc(sp["id"])}
        for threshold, sid in lk.sets.get(rec["itemSet"], {}).get("bonuses", []):
            if lk.desc(sid):
                spell_lines[f"({threshold}) Set: {lk.desc(sid)}"] = sid
        if spell_lines:
            items[iid]["spell_lines"] = spell_lines
    return {iid: lk.icons[rec["displayId"]] for iid, rec in recs.items() if rec["displayId"] in lk.icons}


def add_addon_items(items):
    """Items LootTracker saw that aren't in the VPlusItemDB dump (added to the server later): a basic entry
    from loot.json (name, quality, bind) so every looted item gets a page. Returns {id: icon} the addon saved
    from the game (the last icon fallback)."""
    try:
        seen = json.load(open("/var/www/lokiy/loot/loot.json", encoding="utf-8")).get("items", {})
    except (OSError, ValueError):
        return {}
    binds = {"bop": "Binds when picked up", "boe": "Binds when equipped", "bou": "Binds when used",
             "quest": "Quest Item"}
    for k, v in seen.items():
        iid = int(k)
        if iid in items or not v.get("name"):
            continue
        lines = [binds[v["bind"]]] if v.get("bind") in binds else []
        lines.append("(Seen by LootTracker; full tooltip not in the item dump yet)")
        items[iid] = {"id": iid, "name": v["name"], "quality": v.get("quality", 1), "lines": lines}
    return {int(k): v["icon"].lower() for k, v in seen.items() if v.get("icon")}


def lua_string_expr(expr):
    """AL["Trash Mobs"].." ("..BZ["Molten Core"]..")"  ->  Trash Mobs (Molten Core)"""
    parts = re.findall(r'\w+\["((?:[^"\\]|\\.)*)"\]|"((?:[^"\\]|\\.)*)"', expr)
    return "".join(a or b for a, b in parts).strip()


# AtlasLoot tables whose boss has another name on Vanilla+: {table key: Vanilla+ name}
TABLE_NAMES = {
    "SMHoundmasterLoksey": "Loksey",  # the SM raid's 25225 (2026-10-29)
}


def load_table_names():
    """{table_key: (display name, category)}; category is the instance / group name in AtlasLoot's boss list."""
    txt = open(os.path.join(ATLAS, "TableRegister", "loottables.en.lua"), encoding="utf-8", errors="replace").read()
    names = {}
    head, _, boss = txt.partition("AtlasLoot_TableNamesBoss")
    for key, expr in re.findall(r'^\s*\["(\w+)"\]\s*=\s*\{\s*(.*?),\s*"[^"]*"\s*\},?\s*$', head, re.M):
        names[key] = (lua_string_expr(expr), "Other")
    category = "Other"
    order = []
    for ln in boss.split("\n"):
        m = re.match(r'^\t\[(.*)\]\s*=\s*\{\s*$', ln)
        if m:
            category = lua_string_expr(m.group(1))
            order.append(category)
            continue
        m = re.match(r'^\s*\["(\w+)"\]\s*=\s*\{\s*(.*?),\s*"[^"]*"\s*\},?\s*$', ln)
        if m:
            names[m.group(1)] = (lua_string_expr(m.group(2)) or names.get(m.group(1), ("", ""))[0], category)
    for key, name in TABLE_NAMES.items():
        if key in names:
            names[key] = (name, names[key][1])
    return names, order


def load_text_codes():
    """AtlasLoot's #codes# (e.g. #r3# -> Honored, #a2# -> Leather) from Core/TextParsing.lua."""
    txt = open(os.path.join(ATLAS, "Core", "TextParsing.lua"), encoding="utf-8", errors="replace").read()
    codes = {}
    for code, expr in re.findall(r'gsub\(\s*text\s*,\s*"(#[^"]+#)"\s*,\s*(.+?)\);?\s*$', txt, re.M):
        val = lua_string_expr(expr)
        if val and code not in codes:
            codes[code] = val
    return codes


def clean_label(s, codes):
    s = re.sub(r"=[a-z]+\d*=", "", s)
    s = re.sub(r"#[^#\s]+#", lambda m: codes.get(m.group(0), ""), s)
    return re.sub(r"\s+", " ", s).strip(" ,")


def craft_items():
    """{spell id: crafted item id} from AtlasLoot's Core/Spells.lua (crafting rows name the recipe spell, "s2329", or
    the enchanting spell, "e14293", whose made item is ["item"])."""
    txt = open(os.path.join(ATLAS, "Core", "Spells.lua"), encoding="utf-8", errors="replace").read()
    out = {}
    for m in re.finditer(r'\[(\d+)\]\s*=\s*\{(.*?)\n\t*\}', txt, re.S):
        c = re.search(r'\["(?:craftItem|item)"\]\s*=\s*(\d+)', m.group(2))
        if c:
            out[int(m.group(1))] = int(c.group(1))
    return out


# {table key: [(rest of the AtlasLoot row after the name, recipe spell id or 0)]}, one per row of the table (headers too),
# filled by load_loot_tables():
# the crafted / reputation sources of api/sources.json come from it (item_sources.atlas_item_sources, Part HB)
ROW_INFO = {}


def load_loot_tables():
    """({table_key: [(item_id, icon, display_name, is_new)]} in AtlasLoot order (id 0 rows are spacers / headers),
    {table_key: AtlasLoot folder (Instances, WorldBosses, Crafting, ...)}).

    Some tables open one tab deeper (BWLRazorgore, AQ40Skeram, ...) or without spaces ("BWLKrixix ={"), and some rows
    end in ";". Crafting rows name the
    recipe spell ("s2329", "e14293"): they get the crafted item (no icon: theirs is the spell's), and are left out when
    it makes none (enchants)."""
    craft = craft_items()
    tables, groups = {}, {}
    for path in sorted(glob.glob(os.path.join(ATLAS, "**", "*.lua"), recursive=True)):
        rel = os.path.relpath(path, ATLAS)
        if rel.startswith(("Locale", "TableRegister", "Core", "Libs")):
            continue
        txt = open(path, encoding="utf-8", errors="replace").read()
        for m in re.finditer(r'^\t\t?(\w+) ?= ?\{\n(.*?)^\t\},?', txt, re.M | re.S):
            rows, info = [], []
            for e in re.finditer(r'\{\s*"?([se]?)(\d+)"?,\s*"([^"]*)",\s*"([^"]*)"(.*?)\}[,;]?\s*$', m.group(2), re.M):
                iid, spell = int(e.group(2)), 0
                if e.group(1):
                    spell, iid = iid, craft.get(iid)
                    if not iid:
                        continue
                info.append((e.group(5), spell))
                # a crafting row's icon is the recipe spell's, not the item's
                rows.append((iid, "" if e.group(1) else e.group(3).lower(), re.sub(r"^=q\d=", "", e.group(4)),
                             bool(re.search(r'"N"\s*,?\s*$', e.group(5).strip()))))
            if any(r[0] for r in rows) and m.group(1) not in tables:
                tables[m.group(1)] = rows
                ROW_INFO[m.group(1)] = info
                groups[m.group(1)] = rel.split(os.sep)[0]
    return tables, groups


def tables_by_name(names, tables):
    """{boss / table display name: [(table key, instance)]}"""
    by_name = {}
    for key, (name, cat) in names.items():
        if key in tables:
            by_name.setdefault(name, []).append((key, cat))
    return by_name


def mob_table(by_name, mob, zones):
    """The loot table of a mob the loot tracker saw, matched by name; when several tables share the name, by the
    table's instance and the zones it was killed in. None when there is no single match."""
    keys = by_name.get(mob, [])
    if len(keys) > 1:
        keys = [(k, c) for k, c in keys if c in zones]
    return keys[0][0] if len(keys) == 1 else None


# Drops that neither AtlasLoot nor the loot tracker has, from Lokiy: {table key: [item ids]}. Added to the end of the
# table like the tracked drops.
MANUAL_DROPS = {
    "BWLKrixix": [26231],  # Cloak of Untold Secrets (2026-09-25)
    "SMHoundmasterLoksey": [26318],  # Flintrock Shoulders, a confirmed drop (2026-10-28)
    # Dragonbreath (tier set hand-in): the loot tracker saw it drop from these Dire Maul bosses; AtlasLoot lists only
    # Prince Tortheldrin and King Gordok there (2026-10-29)
    "DMWTendrisWarpwood": [50021], "DMWIllyannaRavenoak": [50021], "DMWMagisterKalendris": [50021],
    "DMWImmolthar": [50021],
    "DMNCaptainKromcrush": [50021],
    # Primordial Hakkari Bijou drops from every Zul'Gurub boss (Lokiy 2026-10-30; AtlasLoot: Azus / Nameless Hermit)
    **{k: [26191] for k in ("ZGJeklik", "ZGVenoxis", "ZGMarli", "ZGMandokir", "ZGGrilek", "ZGHazzarah", "ZGRenataki",
                            "ZGWushoolay", "ZGGahzranka", "ZGThekal", "ZGArlokk", "ZGJindo", "ZGHakkar")},  # in the same loot as his Libram of Focus / Boots of the Full Moon (upload)
    # Highlord Omokk drops the Tier 0 boots and gloves, not the bracers (Lokiy 2026-10-29)
    "LBRSOmokk": [16684, 16692, 16705, 16712, 16717, 16676, 16672, 16737, 16724],
}
# AtlasLoot rows that are wrong for Vanilla+, from Lokiy: {table key: [item ids]}, taken out of the table.
MANUAL_NOT_DROPS = {
    "LBRSOmokk": [16683, 16697, 16703, 16710, 16714, 16681, 16671, 16735, 16722],  # Tier 0 bracers (2026-10-29)
}


def add_manual_drops(tables, items):
    for key, ids in MANUAL_NOT_DROPS.items():
        tables[key] = [r for r in tables.get(key, []) if r[0] not in ids]
    for key, ids in MANUAL_DROPS.items():
        have = {r[0] for r in tables.get(key, [])}
        tables[key] = tables.get(key, []) + [(i, "", items[i]["name"], False) for i in ids
                                             if i in items and i not in have]


def add_tracked_drops(tables, names, items):
    """Items Lokiy's loot tracker saw a boss drop that AtlasLoot has nowhere (new server items such as Hospitaller
    Codex from Brigitte Abbendis) are added to the end of that boss's loot table as ordinary drops (Lokiy 2026-10-23:
    no separate heading). Only rare+ items that are
    in no table at all, so trash loot and items that belong to another table (zone-wide drops) stay out. The mob is
    matched by name; when several tables share the name, by the table's instance and the zones it was killed in.
    Returns {table key: [item ids added]}."""
    in_tables = {r[0] for rows in tables.values() for r in rows if r[0]}
    by_name = tables_by_name(names, tables)
    added = {}
    for mob, zones, drops in load_mob_drops():
        key = mob_table(by_name, mob, zones)
        if not key:
            continue
        new = [iid for iid, q in drops if q >= TRACKED_MIN_QUALITY and iid in items and iid not in in_tables
               and iid not in added.get(key, [])]
        if new:
            added.setdefault(key, []).extend(new)
    for key, ids in added.items():
        tables[key] = tables[key] + [(i, "", items[i]["name"], False) for i in ids]
    return added


# --------------------------------------------------------------------------------------------- raid loot awards

DKP_ZONES = {"The Molten Core": "Molten Core"}  # WebDKP's zone name -> AtlasLoot's instance name
DKP_RAIDS = {"Molten Core", "Blackwing Lair", "Onyxia's Lair", "Zul'Gurub", "Ruins of Ahn'Qiraj",
             "Temple of Ahn'Qiraj", "Naxxramas", "World Bosses"}
DKP_RAID_GAP = 12 * 3600  # awards of one guild less than 12 h apart are one raid
# Majordomo only comes after these bosses' runes are doused: every Ragnaros kill means they died in that lockout
MC_RUNE_BOSSES = ("MCMagmadar", "MCGehennas", "MCGarr", "MCShazzrah", "MCGeddon", "MCGolemagg", "MCSulfuron",
                  "MCMajordomo")


def load_dkp_awards(tables, names):
    """Guild raid loot awards from WebDKP (tools/itemdb/webdkp_awards.json, from webdkp_awards.py).

    Returns ({item id: [(label, awards, kills, True)]} rows for the Observed drops table, the file's source list,
    {item id: {random suffix: awards}}, {item id: {"n", "first", "last"}} for api/sources.json).
    Kills are worked out from the awards: a boss was killed in a raid (one guild's awards less than 12 h apart) in
    which an item only that boss drops (in no other raid table) was awarded; an MC rune boss also in every raid with
    a Ragnaros kill. A low bound: a kill where nobody took any of the boss's own items isn't seen, and a boss without
    own items has no kill count. An item of one boss table is credited to that boss, an item several bosses drop to
    all of them together (label "A / B", kills = their kills added up)."""
    path = os.path.join(HERE, "webdkp_awards.json")
    if not os.path.exists(path):
        return {}, {}, {}, {}
    data = json.load(open(path, encoding="utf-8"))
    raid_keys = {k for k, (nm, cat) in names.items() if cat in DKP_RAIDS and k in tables
                 and not nm.startswith("Trash")}
    where = {}
    for key in raid_keys:
        for r in tables[key]:
            if r[0]:
                where.setdefault(r[0], set()).add(key)
    own = {}
    for iid, keys in where.items():
        if len(keys) == 1:
            own.setdefault(next(iter(keys)), set()).add(iid)
    suffixes, api, raids, last = {}, {}, [], {}
    for iid, when, zone, source, *suffix in sorted(data["awards"], key=lambda a: (a[3], a[1])):
        if suffix:
            s = suffixes.setdefault(iid, {})
            s[suffix[0]] = s.get(suffix[0], 0) + 1
        a = api.setdefault(iid, {"n": 0, "first": when[:10], "last": when[:10]})
        a["n"] += 1
        a["first"], a["last"] = min(a["first"], when[:10]), max(a["last"], when[:10])
        t = datetime.strptime(when, "%Y-%m-%d %H:%M").timestamp()
        if source not in last or t - last[source] > DKP_RAID_GAP:
            raids.append(set())
        last[source] = t
        raids[-1].add(iid)
    kills = {key: {n for n, r in enumerate(raids) if r & items} for key, items in own.items()}
    if SHARED_WB_KEY in raid_keys:  # one loot set the four world bosses share: rates over all their kills
        kills[SHARED_WB_KEY] = kills.get(SHARED_WB_KEY, set()).union(*(kills.get(k, set()) for k, _ in SHARED_WB_BOSSES))
    for key in MC_RUNE_BOSSES:
        if key in raid_keys:
            kills[key] = kills.get(key, set()) | kills.get("MCRagnaros", set())
    rows = {}
    for iid, a in api.items():
        keys = where.get(iid)
        if not keys:
            continue
        label = " / ".join(names[k][0] for k in sorted(keys, key=lambda k: names[k][0]))
        rows[iid] = [(label, a["n"], sum(len(kills.get(k, ())) for k in keys), True)]
    return rows, data["sources"], suffixes, api


# --------------------------------------------------------------------------------------------- observed drops

def norm_zone(z):
    z = re.sub(r"\s*\(.*?\)\s*", "", z or "").strip().lower()
    return re.sub(r"^(lower|upper) ", "", z)


def observed_drops(iid, keys, names, tables_by_cat, obs, boss_names, guess=True):
    """Credit this item's logged loot events to its AtlasLoot sources.

    Returns [(label, drops, kills, estimated)]. A loot is credited to an AtlasLoot boss of the item if that boss died
    in the 10 min before it; to a trash table if a non-boss creature of that zone died before it; items AtlasLoot
    has no source for are credited to the creature that died last (within 2 min), always as an estimate (only with
    guess: classic items without a source are often vendor, crafted or world drops, where the last kill means nothing).
    """
    events = obs["loots"].get(str(iid), [])
    if not events:
        return []
    allowed, trash_zones = set(), {}
    for key in keys:
        name, cat = names.get(key, (key, "Other"))
        m = re.match(r"^Trash Mobs \((.*)\)$", name)
        if m:
            # The logs only know the zone ("Dire Maul", "Blackrock Spire"), not the wing.
            zone = re.sub(r"\s*\(.*?\)\s*", "", m.group(1)).strip()
            zone = re.sub(r"^(Lower|Upper) ", "", zone)
            trash_zones[norm_zone(m.group(1))] = f"Trash Mobs ({zone})"
        elif "random" in name.lower() or "RANDOM" in key:
            allowed |= {names[k][0] for k in tables_by_cat.get(cat, []) if not names[k][0].startswith("Trash")}
        else:
            allowed.add(name)
    credit, guessed = {}, set()
    for cands in events:
        target = None
        for name, _, _ in cands:
            if name in allowed:
                target = name
                break
        if not target and trash_zones:
            for name, _, zone in cands:
                if name not in boss_names and norm_zone(zone) in trash_zones:
                    target = "trash:" + norm_zone(zone)
                    break
        if guess and not target and not keys and cands and cands[0][1] <= 120:
            target = cands[0][0]
            guessed.add(target)
        if target:
            credit[target] = credit.get(target, 0) + 1
    out = []
    for target, n in sorted(credit.items(), key=lambda kv: -kv[1]):
        if target.startswith("trash:"):
            zone = target[6:]
            label = trash_zones[zone]
            kills = sum(v for z, v in obs["zone_kills"].items() if norm_zone(z) == zone)
            estimated = True  # all trash of the zone counted as kills
        else:
            label, kills = target, obs["kills"].get(target, 0)
            estimated = target in guessed
        rate = n / kills if kills else 1
        if kills < 10 or rate > 0.5:
            estimated = True  # few kills or a suspiciously high rate
        out.append((label, n, kills, estimated))
    return out


def rate_text(drops, kills, estimated):
    if not kills:
        return "~?"
    r = drops / kills * 100
    if r > 100:
        return "~100%+"  # more drops than logged kills: a logger missed kills, or it can drop more than once
    s = f"{r:.0f}%" if r >= 10 else f"{r:.1f}%"
    return ("~" if estimated else "") + s


# --------------------------------------------------------------------------------------------- html

def line_class(ln):
    """Colour class of a tooltip line: tt-white, tt-grey, tt-green or tt-flavor."""
    if ln.endswith(" damage per second)"):
        return "tt-white"  # the game shows the dps line white; GREEN's "(" is meant for proc / set lines
    if re.match(r"^\(\d+\) Set:", ln) or ln.startswith("  ") or re.match(r"^.+\(\d+/\d+\)$", ln):
        return "tt-grey"
    if ln.startswith(GREEN):
        return "tt-green"
    if ln.startswith('"'):
        return "tt-flavor"
    if ln.startswith(("Requires ", "Classes:", "Races:")):
        return "tt-grey"
    return "tt-white"


def tooltip_html(item, mark=(), mark_cls="tt-new"):
    """mark: normalized lines to highlight (the changes against the other version of a reworked item)."""
    color = Q_COLOR.get(item["quality"], "#ffffff")
    name_cls = f' {mark_cls}' if norm(item["name"]) in mark else ""
    out = [f'<div class="tt-name{name_cls}" style="color:{color}">{esc(item["name"])}</div>']
    spell_lines = item.get("spell_lines", {})
    for ln in item["lines"]:
        cls = line_class(ln)
        if mark and norm(ln) in mark:
            cls += " " + mark_cls
        text = esc(ln)
        if ln in spell_lines:  # the spell's page on this site
            text = f'<a class="tt-spell" href="/spell/{spell_lines[ln]}">{text}</a>'
        out.append(f'<div class="{cls}">{text}</div>')
    return "".join(out)


def icon_img(icon, size="medium", cls="icon"):
    # An icon that is neither local nor on the CDN falls back to the question mark.
    fallback = icon_url("inv_misc_questionmark", size)
    return (f'<img class="{cls}" src="{esc(icon_url(icon, size))}" alt="" loading="lazy" '
            f'width="36" height="36" onerror="this.onerror=null;this.src=\'{fallback}\'">')


def page(title, body, *, description="", image="", color="#a335ee", path="", extra_head=""):
    og = [
        f'<meta property="og:site_name" content="{SITE_NAME}">',
        f'<meta property="og:title" content="{esc(title)}">',
        f'<meta property="og:type" content="website">',
        f'<meta property="og:url" content="{SITE_URL}{esc(path)}">',
        f'<meta name="theme-color" content="{color}">',
    ]
    if description:
        og.append(f'<meta property="og:description" content="{esc(description)}">')
        og.append(f'<meta name="description" content="{esc(description[:300])}">')
    if image:
        og.append(f'<meta property="og:image" content="{esc(image)}">')
    return f"""<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>{esc(title)} - {SITE_NAME}</title>
{chr(10).join(og)}
<link rel="icon" href="{icon_url('inv_misc_book_09', 'small')}">
<link rel="stylesheet" href="/style.css">
{extra_head}
</head>
<body>
<header class="top">
  <a class="home" href="https://lokiy.dev" title="Back to lokiy.dev">&larr; lokiy.dev</a>
  <a class="brand" href="/">{icon_img('inv_misc_book_09', 'small', 'brand-icon')}<span>{SITE_NAME}</span></a>
  <nav class="nav"><a href="/">Database</a><a href="/reworked">Reworked</a><a href="/type/weapons">Types</a><a href="/instances">Instances</a><a href="/zones">Zones</a><a href="/unseen">Unseen</a><a href="/removed">Not on V+</a><div class="nav-drop"><button type="button" aria-expanded="false" aria-haspopup="true">Search &#9662;</button><div class="drop"><a href="/search">Items <span class="muted">filters, stats</span></a><a href="/npcs">NPCs</a><a href="/spells">Spells</a><a href="/quests">Quests</a><a href="/zones">Zones</a><a href="/instances">Loot tables</a></div></div></nav>
  <form class="search" action="/" method="get" role="search">
    <input type="search" name="q" placeholder="Search items, NPCs, spells, quests, zones..." aria-label="Search everything" autocomplete="off">
  </form>
</header>
<main>
{body}
</main>
<footer>Vanilla Plus items, from the server's item data and AtlasLoot. Updated {date.today().isoformat()}.
&middot; <a href="/unseen">Items not seen on Vanilla+ yet</a> &middot; <a href="/removed">Items not on Vanilla+</a>
&middot; Part of <a href="https://lokiy.dev">lokiy.dev</a>: <a href="https://sim.lokiy.dev">Sim</a>,
<a href="https://calc.lokiy.dev">Calculators</a>, <a href="https://addons.lokiy.dev">Addons</a>,
<a href="https://collection.lokiy.dev">Collection</a></footer>
<script src="/site.js" defer></script>
</body>
</html>
"""


NOT_ON_SERVER = {}  # item id -> why it isn't on Vanilla+ (item_sources.py), set by main()


def item_link(item, icon):
    color = Q_COLOR.get(item["quality"], "#ffffff")
    nos = NOT_ON_SERVER.get(item["id"])
    cls, title = (' nos', f' title="Not on Vanilla+: {esc(nos)}"') if nos else ("", "")
    return (f'<a class="item{cls}" href="/item/{item["id"]}" data-item="{item["id"]}"{title}>'
            f'{icon_img(icon, "small", "icon-s")}<span style="color:{color}">{esc(item["name"])}</span></a>')


def sources_html(srcs):
    """The Source card's list: each source with its type and, unless AtlasLoot, where it came from."""
    lis = []
    for e in srcs:
        text = esc(e["text"])
        if e.get("href"):
            rel = ' rel="noopener"' if e["href"].startswith("http") else ""
            text = f'<a href="{esc(e["href"])}"{rel}>{text}</a>'
        if e.get("npc"):
            text += f' <a class="src-npc" href="{esc(e["npc"])}">NPC</a>'
        frm = "" if e["from"] == "atlasloot" else f' <span class="muted src-from">{esc(FROM_LABEL[e["from"]])}</span>'
        lis.append(f'<li><span class="src-type t-{e["type"].replace(" ", "-")}">{esc(TYPE_LABEL.get(e["type"], e["type"]))}'
                   f'</span> {text}{frm}</li>')
    return f'<ul class="sources typed">{"".join(lis)}</ul>'


# --------------------------------------------------------------------------------------------- build

def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--out", default="/var/www/db")
    args = ap.parse_args()
    out = args.out

    items = load_items()
    cache_icons = add_cache_items(items)
    addon_icons = add_addon_items(items)
    names, cat_order = load_table_names()
    tables, table_groups = load_loot_tables()
    tracked = add_tracked_drops(tables, names, items)
    add_manual_drops(tables, items)
    add_shared_world_boss_loot(tables, names, table_groups, items)
    codes = load_text_codes()
    wowhead = load_wowhead()
    classic_only = set(add_classic_items(items, wowhead, tables))

    flagged_new = {r[0] for rows in tables.values() for r in rows if r[3]}

    def is_custom(iid):
        return iid in items and (iid > CLASSIC_MAX_ITEM_ID or iid in flagged_new)

    reworked = find_reworked(items, CLASSIC_MAX_ITEM_ID, skip=flagged_new | classic_only, wh=wowhead)
    all_items = sorted(items)
    custom = [i for i in all_items if is_custom(i)]
    browsable = [i for i in all_items if not DEV_ITEM.search(items[i]["name"])]
    types = {i: classify(items[i]) for i in all_items}

    icons, sources = {}, {}
    for key, rows in tables.items():
        for iid, icon, _, _ in rows:
            if iid and icon and iid not in icons:
                icons[iid] = icon
            if iid in items:
                sources.setdefault(iid, [])
                if key not in sources[iid]:
                    sources[iid].append(key)
    for iid in all_items:
        if iid not in icons and cache_icons.get(iid):
            icons[iid] = cache_icons[iid]
        elif iid not in icons and wowhead.get(iid, {}).get("icon"):
            icons[iid] = wowhead[iid]["icon"]
        elif iid not in icons and addon_icons.get(iid):
            icons[iid] = addon_icons[iid]
    loot_keys = [k for k in tables if any(r[0] in items for r in tables[k])]

    def table_label(key):
        name, cat = names.get(key, (key, "Other"))
        return name or key, cat

    def kind(iid):
        if iid in classic_only:
            return ("Classic item, dropped on Vanilla+ (Lokiy)" if iid in LOKIY_DROPPED_IDS
                    else "Classic item, not seen on Vanilla+ yet")
        if DEV_ITEM.search(items[iid]["name"]):
            return "Test / unused item"
        if is_custom(iid):
            return "Custom item"
        if iid in reworked:
            return "Reworked classic item"
        if iid in wowhead and not items[iid]["name"].startswith(RECIPE):
            return "Classic item, same as classic"
        return "Classic item"

    # observed drops from players' combat logs (tools/itemdb/observed_loot.json, from loot_from_logs.py)
    obs_path = os.path.join(HERE, "observed_loot.json")
    obs = json.load(open(obs_path)) if os.path.exists(obs_path) else {"loots": {}, "kills": {}, "zone_kills": {}, "logs": 0}
    tables_by_cat = {}
    for key, (nm, cat) in names.items():
        tables_by_cat.setdefault(cat, []).append(key)
    boss_names = {nm for nm, cat in names.values() if nm and not nm.startswith("Trash")}
    observed = {i: observed_drops(i, sources.get(i, []), names, tables_by_cat, obs, boss_names, guess=is_custom(i))
                for i in all_items}
    # Lokiy's loot tracker (lokiy.dev/loot) and the possible random affixes per item
    loot = load_loot()
    loot_names, loot_drops, loot_time = loot if loot else ({}, {}, None)
    # where each item comes from (AtlasLoot, item workbook, loot tracker, pfQuest, Wowhead) and what isn't on Vanilla+
    srcs = Sources(items, tables, names, table_groups, loot_drops)
    NOT_ON_SERVER.clear()
    NOT_ON_SERVER.update(srcs.not_on_server)
    item_srcs = {i: srcs.sources(i) for i in all_items}
    affixes, affix_stats, affix_chances = load_random_affixes()
    # suffix names and values from the server's DBCs; which ones an item rolls: Wowhead's list + the loot tracker's,
    # pooled per item cache group (affixes.py)
    affix_ids = load_affix_ids()
    afx = Affixes({i: [r[0] for r in rows] for i, rows in affix_ids.items()})

    def afx_rows_of(i):  # with the suffix names awarded / seen on the item, for borrowing from its group
        wanted = set(seen_affixes(items[i]["name"], loot_names.get(i, {}))) | set(dkp_suffixes.get(i, {}))
        return afx.rows(i, wanted)
    loot_rolls = load_rolls()
    loot_note = ('<p class="muted small">From <a href="https://lokiy.dev/loot/">Lokiy\'s loot tracker</a>'
                 + (f', updated {date.fromtimestamp(loot_time).isoformat()}' if loot_time else '')
                 + '. Rate = drops per kill for items that go to a group roll (green and better, or anything seen '
                 + 'rolled; the roll shows them from every corpse), else drops per looted corpse (only seen on corpses '
                 + 'Lokiy opened).</p>')
    obs_note = (f'<p class="muted small">From {obs["logs"]} players\' combat logs. Rate = drops / kills seen in those logs. '
                f'<b>~</b> = estimate (few kills, a guessed source, or a rate that looks too high).</p>')
    # raid loot awards (WebDKP), added to the observed drops: a source in both gets one row with drops and kills
    # added up (the logs and the awards are different kills: awards end 2025-04, the logs start 2026-07)
    dkp, dkp_sources, dkp_suffixes, dkp_api = load_dkp_awards(tables, names)
    in_logs = {i for i, rows in observed.items() if rows}
    for i, rows in dkp.items():
        if i not in observed:
            continue
        merged = {label: [n, k, est] for label, n, k, est in observed[i]}
        for label, n, k, est in rows:
            m = merged.setdefault(label, [0, 0, est])
            m[0], m[1], m[2] = m[0] + n, m[1] + k, m[2] or est
        observed[i] = [(label, n, k, est) for label, (n, k, est) in sorted(merged.items(), key=lambda kv: -kv[1][0])]
    dkp_note = ('<p class="muted small">Counted too: items raid leaders gave out for DKP ('
                + ", ".join(f'{esc(s)} {v["first"][:4]}-{v["last"][:4]}' for s, v in dkp_sources.items())
                + ', from WebDKP), with kills = raids in which an item only that boss drops was given out (Molten Core '
                + 'bosses also count every Ragnaros kill). These are low bounds: items nobody wanted are missing.</p>')

    # every mob / boss Lokiy's loot tracker saw: its drops on the boss's loot table page, else on its own /mob page
    mob_loot = load_mob_loot()
    by_name = tables_by_name(names, tables)
    mob_key = {}
    for m in mob_loot:
        key = mob_table(by_name, m["name"], set(m["zones"]))
        if key in loot_keys:
            mob_key[m["name"]] = key
    loot_key_mobs = {}
    for m in mob_loot:
        if m["name"] in mob_key:
            loot_key_mobs.setdefault(mob_key[m["name"]], []).append(m)

    # creature ids (npcs.py, Part GJ): the tracker's mobs and AtlasLoot's boss tables only know names
    npcs = load_npcs()
    logged = load_logged()  # creature ids from combat logs; Vanilla+'s own ones get a page of their own (Part GL)
    for nid, r in logged.items():
        if nid not in npcs:
            npcs[nid] = {"name": r["name"], "lvl": "", "zone": r["zone"], "vplus": True, **({"rank": 3} if r.get("boss") else {})}
    matcher = NpcMatcher(npcs, logged)
    mob_npc = {}
    for m in mob_loot:
        nid = matcher.match(m["name"], list(m["zones"]), [r[0] for r in m["items"]])
        if nid:
            mob_npc[m["name"]] = nid
    table_npc = {}  # boss loot table key -> creature id
    for key in loot_keys:
        nm, cat = names.get(key, ("", "Other"))
        if (not nm or table_groups.get(key) not in ("Instances", "WorldBosses") or key == SHARED_WB_KEY
                or re.match(r"^(Trash|Random|.* Sets?$)", nm)):
            continue
        nid = matcher.match(re.sub(r"\s*\((Rare|.*Scourge Invasion.*)\)$", "", nm), [cat],
                            [r[0] for r in tables[key] if r[0]])
        if nid:
            table_npc.setdefault(key, nid)
    npc_tables, npc_mob = {}, {}
    for key, nid in table_npc.items():
        npc_tables.setdefault(nid, []).append(key)
    for m in mob_loot:
        if m["name"] in mob_npc:
            npc_mob[mob_npc[m["name"]]] = m

    spells = load_spells()  # also for the NPC pages' abilities (Part GN)

    # DungeonJournal's abilities / tactics (npcs.load_journal; Part GN), on the NPC pages
    journal, journal_flags = load_journal()
    quests = Quests(npcs)  # quest / zone pages (quests.py; Part GO)
    quests.add_rewards(item_srcs)

    def journal_icon(path):
        base = (path or "").replace("/", "\\").split("\\")
        if len(base) < 2 or base[-2].lower() != "icons" or base[-1].lower() == "temp" or "AddOns" in base:
            return "inv_misc_questionmark"
        return base[-1].lower()

    def ability_html(a):
        if a.get("separator"):
            return f'<li class="dj-sep">{esc(a["name"])}</li>'
        sid = a.get("spellId")
        nm = esc(a["name"])
        if sid and sid in spells:
            nm = f'<a href="/spell/{sid}">{nm}</a>'
        tags = "".join(f'<span class="tag">{esc(x)}</span> ' for x in
                       ([a["school"]] if a.get("school") else []) + [r.capitalize() for r in a.get("roles", [])])
        warn = '<b class="dj-warn" title="Important">!</b> ' if a.get("warning") else ""
        lines = "".join(f'<p>{esc(ln)}</p>' for ln in a.get("lines", []))
        return (f'<li>{icon_img(journal_icon(a.get("icon")), "small", "icon-s")} {warn}<b>{nm}</b> {tags}'
                f'<div class="dj-lines">{lines}</div></li>')

    def add_li(a):
        nm = esc(a["name"])
        if a.get("npcId"):
            nm = f'<a href="/npc/{a["npcId"]}">{nm}</a>'
        return f'<li>{nm}{" x" + str(a["count"]) if a.get("count") else ""}</li>'

    def tactic_li(ln):
        if isinstance(ln, str):
            return f'<li>{esc(ln)}</li>'
        if ln.get("separator"):
            return f'</ul><h3>{esc(ln.get("name", ""))}</h3><ul class="dj-tactics">'
        text = esc(ln.get("text", "").replace("{boss}: ", ""))
        return f'<li><b class="dj-warn">!</b> <b>{text}</b></li>' if ln.get("warning") else f'<li>{text}</li>'

    def journal_html(nid):
        parts = []
        for raid, e, kind, boss in journal.get(nid, []):
            flags = " ".join(f'<span class="tag" title="{esc(journal_flags.get(f, {}).get("desc", ""))}">'
                             f'{esc(journal_flags.get(f, {}).get("name", f))}</span>' for f in e.get("flags", []))
            st = e.get("stats") or {}
            stats = " &middot; ".join(f'{k.capitalize()} {v}' for k, v in st.items())
            where = f'{esc(raid)}' + (f' &middot; add of {esc(boss)}' if boss else "")
            head = (f'<p class="muted small">{where}{" &middot; " + stats if stats else ""}</p>'
                    + (f'<p>{flags}</p>' if flags else ""))
            if e.get("lines"):
                head += "".join(f'<p>{esc(ln)}</p>' for ln in e["lines"])
            abil = "".join(ability_html(a) for a in e.get("abilities", []))
            if abil:
                parts.append(f'<section class="card dj"><h2>Abilities</h2>{head}<ul class="dj-abilities">{abil}</ul></section>')
            elif head:
                parts.append(f'<section class="card dj"><h2>Encounter</h2>{head}</section>')
            adds = [a for a in e.get("adds", []) if a.get("name") and not a.get("separator")]
            if adds:
                lis = "".join(add_li(a) for a in adds)
                parts.append(f'<section class="card"><h2>Adds</h2><ul class="loot">{lis}</ul></section>')
            t = e.get("tactics")
            if t and (t.get("lines") or t.get("resistances")):
                res = ", ".join(f'{v} {k.capitalize()}' for k, v in (t.get("resistances") or {}).items())
                lis = "".join(tactic_li(ln) for ln in t.get("lines", []))
                parts.append(f'<section class="card dj"><h2>Tactics</h2>'
                             + (f'<p><b>Resistance:</b> {res}</p>' if res else "")
                             + f'<ul class="dj-tactics">{lis}</ul></section>')
        if parts:
            parts.append('<p class="muted small">Abilities and tactics from the <a href="https://github.com/lokiy999/'
                         'DungeonJournal" rel="noopener">DungeonJournal</a> addon.</p>')
        return parts

    # item sources: db.lokiy.dev's NPC pages instead of Wowhead's, and the NPC behind a loot table (Part GN)
    WH_NPC = "https://www.wowhead.com/classic/npc="

    def local_source(e):
        e = dict(e)
        h = e.get("href", "")
        if h.startswith(WH_NPC) and h[len(WH_NPC):].isdigit() and int(h[len(WH_NPC):]) in npcs:
            e["href"] = f"/npc/{h[len(WH_NPC):]}"
        elif "/quest=" in h and h.rsplit("=", 1)[1].isdigit() and quests.href(int(h.rsplit("=", 1)[1])):
            e["href"] = quests.href(int(h.rsplit("=", 1)[1]))
        elif h.startswith("/loot/") and h[6:] in table_npc:
            e["npc"] = f"/npc/{table_npc[h[6:]]}"
        elif not h and e["from"] == "tracker" and e["text"].startswith("Drops from ") and e["text"][11:] in mob_npc:
            e["href"] = f"/npc/{mob_npc[e['text'][11:]]}"
        return e

    def src_cell(src):
        return f'<a href="/npc/{mob_npc[src]}">{esc(src)}</a>' if src in mob_npc else esc(src)

    def mob_href(m):
        if m["name"] in mob_npc:
            return f'/npc/{mob_npc[m["name"]]}'
        return f'/loot/{mob_key[m["name"]]}' if m["name"] in mob_key else f'/mob/{slug(m["name"])}'

    def mob_rate(r):
        return (f"{min(r, 1) * 100:.0f}%" if r >= 0.1 else f"{r * 100:.1f}%") if r is not None else "?"

    rough_note = ('<p class="notice"><b>Very rough.</b> These drops come from very little data: one player\'s '
                  'loot tracker and a handful of kills, so rates can be far off and many drops are still missing. '
                  'They get better as more kills are tracked.</p>')

    def mob_drops_html(m, heading="Drops seen"):
        seen = (f'Killed {m["kills"]} time{"s" if m["kills"] != 1 else ""}, {m["looted"]} '
                f'corpse{"s" if m["looted"] != 1 else ""} looted')
        if not m["items"]:
            return (f'<section class="card mob-drops"><h2>{esc(heading)}</h2>{rough_note}'
                    f'<p class="muted">{seen}. No drops seen yet.</p></section>')
        trs = []
        for iid, nm, q, n, r in m["items"]:
            if iid in items:
                cell = item_link(items[iid], icons.get(iid, "inv_misc_questionmark"))
            else:
                cell = f'<span style="color:{Q_COLOR.get(q, "#ffffff")}">{esc(nm or str(iid))}</span>'
            trs.append(f'<tr><td>{cell}</td><td>{n}</td><td class="rate">~{mob_rate(r)}</td></tr>')
        return (f'<section class="card mob-drops"><h2>{esc(heading)}</h2>{rough_note}'
                f'<p class="muted small">{seen}, from <a href="https://lokiy.dev/loot/">Lokiy\'s loot tracker</a>.</p>'
                f'<table class="obs-table"><thead><tr><th>Item</th><th>Drops</th><th>Rate</th></tr></thead>'
                f'<tbody>{"".join(trs)}</tbody></table>'
                f'<p class="muted small">Rate = drops per kill for green and better items (the group roll shows them '
                f'from every corpse), else drops per looted corpse. ~ = estimate.</p></section>')

    # clean output (only our generated files)
    for sub in ("item", "loot", "type", "tt", "mob", "spell", "npc", "quest", "zone"):
        shutil.rmtree(os.path.join(out, sub), ignore_errors=True)
        os.makedirs(os.path.join(out, sub), exist_ok=True)

    # ---- item pages
    for iid in all_items:
        it = items[iid]
        rw = reworked.get(iid)
        icon = icons.get(iid, "inv_misc_questionmark")
        color = Q_COLOR.get(it["quality"], "#ffffff")
        page_key, section = types[iid]
        nos = NOT_ON_SERVER.get(iid)
        nos_html = f'<p class="notice nos-banner"><b>Not on Vanilla+.</b> {esc(nos)}</p>' if nos else ""
        if iid in classic_only and not nos:
            nos_html = ('<p class="notice"><b>Classic data, not seen on Vanilla+ yet.</b> No player\'s client has seen '
                        'this item on the server, so the tooltip is classic 1.12 (from Wowhead); it may differ on Vanilla+. '
                        'It is replaced by the server\'s own data as soon as an item cache upload contains it.</p>')
            if iid in LOKIY_DROPPED_IDS:
                nos_html = ('<p class="notice"><b>Has dropped on Vanilla+</b> (Lokiy), but no item cache upload contains it '
                            'yet, so the tooltip is classic 1.12 (from Wowhead); it may differ on Vanilla+. It is replaced by '
                            'the server\'s own data as soon as an item cache upload contains it.</p>')
        if item_srcs[iid]:
            weak = item_srcs[iid][0]["from"] in ("pfquest", "wowhead")
            src_html = (f'<section class="card"><h2>Source</h2>{nos_html}{sources_html([local_source(e) for e in item_srcs[iid]])}'
                        + ('<p class="muted small">Classic data: not checked against Vanilla+ yet.</p>' if weak else "")
                        + '</section>')
        else:
            src_html = f'<section class="card"><h2>Source</h2>{nos_html}<p class="muted">No known source yet.</p></section>'
        src_html += quests.item_card(iid)
        if observed[iid]:
            trs = "".join(f'<tr><td>{esc(label)}</td><td>{n}</td><td>{k}</td><td class="rate">{rate_text(n, k, est)}</td></tr>'
                          for label, n, k, est in observed[iid])
            notes = (obs_note if iid in in_logs else "") + (dkp_note if dkp.get(iid) else "")
            src_html += (f'<section class="card obs"><h2>Observed drops</h2><table class="obs-table"><thead><tr><th>Source</th>'
                         f'<th>Drops</th><th>Kills</th><th>Rate</th></tr></thead><tbody>{trs}</tbody></table>{notes}</section>')
        if loot_drops.get(iid):
            trs = "".join(f'<tr><td>{src_cell(src)}</td><td>{n}</td><td>{lt}</td><td>{k}</td><td class="rate">'
                          f'{(f"{min(r, 1) * 100:.0f}%" if r >= 0.1 else f"{r * 100:.1f}%") if r is not None else ""}</td></tr>'
                          for src, n, lt, k, r in loot_drops[iid])
            src_html += (f'<section class="card obs"><h2>Lokiy\'s loot</h2><table class="obs-table"><thead><tr>'
                         f'<th>Source</th><th>Drops</th><th>Looted</th><th>Kills</th><th>Rate</th></tr></thead><tbody>{trs}</tbody></table>'
                         f'{loot_note}</section>')
        seen =seen_affixes(it["name"], loot_names.get(iid, {}))
        awarded = dkp_suffixes.get(iid, {})
        afx_rows = afx_rows_of(iid)
        if afx_rows or seen or awarded:
            groups, group_seen, only_seen, donor = {}, {}, {}, {}
            for sid, name, effs, n_item, n_group, from_wh, from_item in afx_rows:
                groups.setdefault(name, []).append(", ".join(effs))
                group_seen[name] = group_seen.get(name, 0) + n_group - n_item  # on other items of its group
                only_seen[name] = only_seen.get(name, True) and not from_wh and not from_item
                if from_item and not from_wh:
                    donor.setdefault(name, from_item)
            for name in list(seen) + list(awarded):
                groups.setdefault(name, [])
            chances = affix_chances.get(iid, {})
            lis = []
            for name in sorted(groups, key=lambda n: (-chances.get(n, -1), -seen.get(n, 0) - awarded.get(n, 0), n)):
                effs = esc(merge_tiers(groups[name])) or '<span class="muted">stats unknown</span>'
                tag = f'<span class="tag">seen {seen[name]}&times;</span>' if name in seen else ""
                if name in awarded:
                    tag += f'<span class="tag">awarded {awarded[name]}&times;</span>'
                if group_seen.get(name):
                    tag += (f'<span class="tag" title="Seen on other items that roll from the same suffix list">'
                            f'{group_seen[name]}&times; on its group</span>')
                if only_seen.get(name):
                    tag += '<span class="tag" title="Not in Wowhead\'s list for this item">seen on Vanilla+</span>'
                if name in donor and name not in chances:
                    d = items.get(donor[name], {"name": str(donor[name])})
                    tag += (f'<span class="tag" title="Not in Wowhead\'s list for this item: from the list of an item that '
                            f'rolls from the same server list">from its group (<a href="/item/{donor[name]}">'
                            f'{esc(d["name"])}</a>)</span>')
                rolls = loot_rolls.get(iid, {}).get(f'{it["name"]} {name}', {})
                if rolls:  # the exact tiers seen, e.g. "rolled +15 Arcane Resistance" (or "+14 ... 1x, +15 ... 2x")
                    tag += ('<span class="muted"> rolled ' + esc(", ".join(r if len(rolls) == 1 else f"{r} {n}x"
                                                                     for r, n in sorted(rolls.items()))) + '</span>')
                pct = (f'<span class="rate">{chances[name]:g}%</span>' if name in chances
                       else f'<span class="rate muted">&lt;{min(chances.values()):g}%</span>' if chances and name not in donor
                       else "")
                lis.append(f'<li class="{"seen" if name in seen or name in awarded or group_seen.get(name) else ""}"><span class="affix" style="color:{color}">{esc(it["name"])} '
                           f'<b>{esc(name)}</b></span>{tag}{pct}<span class="muted affix-stats">{effs}</span></li>')
            chance_note = (" Roll chances from Wowhead; it only lists affixes it has seen drop, the others are rarer "
                           "(shown as &lt; the lowest listed chance)." if chances else "")
            src_html += (f'<section class="card affixes"><h2>Random affixes</h2><ul class="affix-list">{"".join(lis)}</ul>'
                         f'<p class="muted small">Affix names and values from the server\'s own client data (Patch 3); '
                         f'which affixes the item can roll from Wowhead\'s classic list plus every affix '
                         f'<a href="https://lokiy.dev/loot/">Lokiy\'s loot tracker</a> saw on it or on another item that '
                         f'rolls from the same list ("on its group"; "seen on Vanilla+" = not in Wowhead\'s list). An affix '
                         f'with several values has tiers (12-13 = any value in between, 15/17 = one of those).{chance_note} '
                         f'"Seen" = looted in the loot tracker, matched by name; "rolled" = the exact value, for drops logged '
                         f'since the addon saves it.'
                         f'{" &quot;Awarded&quot; = given out in raids with this suffix (WebDKP)." if awarded else ""}'
                         f'</p></section>')
        if rw:
            classic_item = {"name": rw["classic_name"], "quality": it["quality"], "lines": rw["classic"]}
            top_html = f"""<section class="card changes">
  <h2>Changed from classic</h2>
  <div class="compare">
    <div><h3>Vanilla Plus</h3><div class="tooltip">{tooltip_html(it, set(rw["added"]), "tt-new")}</div></div>
    <div><h3>Classic</h3><div class="tooltip">{tooltip_html(classic_item, set(rw["removed"]), "tt-old")}</div></div>
  </div>
  <p class="muted small">Highlighted lines differ. Classic = WoW Classic data from
  <a href="https://www.wowhead.com/classic/item={iid}" rel="noopener">Wowhead</a>.</p>
</section>"""
        else:
            top_html = f'<div class="tooltip">{tooltip_html(it)}</div>'
        wh_link = (f' &middot; <a href="https://www.wowhead.com/classic/item={iid}" rel="noopener">Wowhead</a>'
                   if iid in wowhead and not is_custom(iid) else "")
        desc = "\n".join(it["lines"])
        body = f"""
<article class="item-page">
  <div class="item-head">{icon_img(icon, 'large', 'icon-l')}
    <div><h1 style="color:{color}">{esc(it['name'])}</h1>
    <div class="muted">{kind(iid)} {iid} &middot; <a href="/type/{page_key}#{slug(section)}">{esc(section)}</a>{wh_link}</div></div>
  </div>
  {top_html}
  {src_html}
</article>"""
        with open(os.path.join(out, "item", f"{iid}.html"), "w", encoding="utf-8") as f:
            f.write(page(it["name"], body, description=desc, image=icon_url(icon, "large", absolute=True),
                         color=color, path=f"/item/{iid}"))

    def shared_note(key):
        boss_links = ", ".join(f'<a href="/loot/{k}">{esc(n)}</a>' for k, n in SHARED_WB_BOSSES)
        if key == SHARED_WB_KEY:
            return (f'<p class="notice">One loot set that {boss_links} share (Lokiy): each of them can drop these, and '
                    f'the drop rates are shared between them, so a rate here counts every kill of all four bosses '
                    f'(guild raid awards, WebDKP).</p>')
        if key in dict(SHARED_WB_BOSSES):
            return (f'<p class="notice">Also drops the <a href="/loot/{SHARED_WB_KEY}">World Boss Shared Loot</a> '
                    f'(the AQ class books and three glove formulas), shared with {boss_links}.</p>')
        return ""

    # ---- loot table pages
    for key in loot_keys:
        name, cat = table_label(key)
        # items in item ID order under each sub-heading (Lokiy 2026-10-23); blank spacer rows are dropped
        groups = [[None, []]]
        for r in tables[key]:
            if not r[0]:
                header = clean_label(r[2], codes)
                if header:
                    groups.append([header, []])
            else:
                groups[-1][1].append(r)
        rows = []
        for header, members in groups:
            if header:
                rows.append(f'<li class="sub">{esc(header)}</li>')
            for iid, icon, label, _ in sorted(members, key=lambda r: r[0]):
                if iid in items:
                    rate = next((rate_text(n, k, est) for label, n, k, est in observed[iid] if label == name), "")
                    rate_html = f'<span class="rate" title="Observed drop rate in players\' combat logs">{rate}</span>' if rate else ""
                    tag, cls = "", ""
                    if is_custom(iid):
                        tag, cls = '<span class="tag">new</span>', ' class="custom"'
                    elif iid in reworked:
                        tag, cls = '<span class="tag">changed</span>', ' class="reworked"'
                    rows.append(f'<li{cls}>{item_link(items[iid], icons.get(iid, icon))}{tag}{rate_html}</li>')
                else:
                    rows.append(f'<li><a class="item ext" href="https://www.wowhead.com/classic/item={iid}" rel="noopener">'
                                f'{icon_img(icon, "small", "icon-s")}<span>{esc(label)}</span></a>'
                                f'<span class="muted ext-note">Wowhead</span></li>')
        n_custom = sum(1 for r in tables[key] if is_custom(r[0]))
        n_rw = sum(1 for r in tables[key] if r[0] in reworked)
        counts = ", ".join(x for x in (f"{n_custom} new" if n_custom else "", f"{n_rw} changed" if n_rw else "") if x)
        counts = f"{counts} item{'s' if n_custom + n_rw != 1 else ''} on Vanilla Plus." if counts else "No changes on Vanilla Plus."
        nk = obs["kills"].get(name, 0)
        kills_note = (f"Killed {nk} times in players' combat logs; percentages are the observed drop rates "
                      f"(~ = estimate).") if nk else ""
        body = f"""
<article>
  <div class="crumbs"><a href="/instances">Instances</a>{' / ' + esc(cat) if cat and cat != 'Other' else ''}</div>
  <h1>{esc(name)}</h1>
  {f'<p class="notice nos-banner"><b>Not on Vanilla+.</b> {esc(CUT_TEXT[srcs.cut_table(key)])}</p>' if srcs.cut_table(key) else ''}
  <p class="muted">{counts} {kills_note}</p>
  {shared_note(key)}
  {f'<p class="muted">NPC: <a href="/npc/{table_npc[key]}">{esc(npcs[table_npc[key]]["name"])}</a></p>' if key in table_npc else ''}
  <ul class="loot">{''.join(rows)}</ul>
  {''.join(mob_drops_html(m, f"Drops seen from {m['name']}") for m in loot_key_mobs.get(key, []))}
</article>"""
        with open(os.path.join(out, "loot", f"{key}.html"), "w", encoding="utf-8") as f:
            f.write(page(name, body, description=f"{name} loot on Vanilla Plus. {counts}", path=f"/loot/{key}"))

    # ---- browse pages: custom items (index) and reworked items, by instance / boss
    by_cat_all = {}
    for key in loot_keys:
        by_cat_all.setdefault(table_label(key)[1], []).append(key)
    cats_all = [c for c in cat_order if c in by_cat_all] + [c for c in by_cat_all if c not in cat_order]

    def browse_blocks(wanted):
        blocks = []
        for cat in cats_all:
            tbl = []
            for key in by_cat_all[cat]:
                name, _ = table_label(key)
                seen, links = set(), []
                for r in tables[key]:
                    if r[0] in items and wanted(r[0]) and r[0] not in seen:
                        seen.add(r[0])
                        links.append(f'<li>{item_link(items[r[0]], icons.get(r[0]))}</li>')
                if links:
                    tbl.append(f'<div class="table"><h3><a href="/loot/{key}">{esc(name)}</a></h3><ul class="loot">{"".join(links)}</ul></div>')
            if tbl:
                blocks.append(f'<section class="cat" data-cat><h2>{esc(cat)}</h2>{"".join(tbl)}</section>')
        other = [items[i] for i in browsable if wanted(i) and not sources.get(i) and item_srcs[i]]
        nosrc = [items[i] for i in browsable if wanted(i) and not sources.get(i) and not item_srcs[i]]
        for title, lst in (("Other sources", other), ("No known source", nosrc)):
            if lst:
                lst.sort(key=lambda it: (-it["quality"], it["name"]))
                blocks.append(f'<section class="cat" data-cat><h2>{title}</h2><div class="table wide"><ul class="loot cols">'
                              + "".join(f'<li>{item_link(it, icons.get(it["id"]))}</li>' for it in lst) + "</ul></div></section>")
        return "".join(blocks), len(nosrc)

    type_counts = {}
    for iid in browsable:
        type_counts[types[iid][0]] = type_counts.get(types[iid][0], 0) + 1
    type_nav = ('<section class="card"><h2>Browse by type</h2><ul class="type-grid">'
                + "".join(f'<li><a href="/type/{k}">{esc(nm)}</a> <span class="muted">{type_counts[k]}</span></li>'
                          for k, nm in PAGES if type_counts.get(k)) + "</ul></section>")

    search_box = ('<section id="results" class="card" hidden><h2>Search results</h2><nav class="type-tabs" id="result-tabs"></nav>'
                  '<div id="result-groups"></div></section>')
    blocks, n_nosrc = browse_blocks(is_custom)
    body = f"""
<section class="intro">
  <h1>Vanilla Plus item database</h1>
  <p class="muted">All {len(all_items)} items on the server with their in-game tooltips and where they drop:
  {len(custom)} new items (below), <a href="/reworked">{len(reworked)} reworked classic items</a>, and every other item
  <a href="/type/consumables">by type</a> or <a href="/instances">by instance</a>.
  Link any item: <code>db.lokiy.dev/item/&lt;id&gt;</code></p>
</section>
{search_box}
<div id="browse">{type_nav}<h2 class="section-title">New on Vanilla Plus</h2>{blocks}</div>"""
    with open(os.path.join(out, "index.html"), "w", encoding="utf-8") as f:
        f.write(page("Item database", body, path="/",
                     description=f"All {len(all_items)} Vanilla Plus items: {len(custom)} new, {len(reworked)} reworked, with tooltips and drop sources."))

    blocks, n_rw_nosrc = browse_blocks(lambda i: i in reworked)
    body = f"""
<section class="intro">
  <h1>Reworked classic items</h1>
  <p class="muted">{len(reworked)} classic items whose stats or effects are different on Vanilla Plus. Each item page
  shows the Vanilla Plus and the classic tooltip side by side. Compared against WoW Classic data from Wowhead, so a few
  differences can be Classic-only changes rather than Vanilla Plus ones.</p>
</section>
{search_box}
<div id="browse">{blocks}</div>"""
    with open(os.path.join(out, "reworked.html"), "w", encoding="utf-8") as f:
        f.write(page("Reworked classic items", body, path="/reworked",
                     description=f"{len(reworked)} classic items with changed stats on Vanilla Plus, compared with classic."))

    # ---- mob pages (mobs without a loot table) and the mobs index
    def main_zone(m):
        return max(m["zones"], key=lambda z: m["zones"][z]) if m["zones"] else "Unknown zone"

    by_zone = {}
    for m in mob_loot:
        by_zone.setdefault(main_zone(m), []).append(m)
        if m["name"] in mob_npc:  # the page moved to /npc/<id> (Part GJ): keep the old address working
            target = f'/npc/{mob_npc[m["name"]]}'
            with open(os.path.join(out, "mob", f"{slug(m['name'])}.html"), "w", encoding="utf-8") as f:
                f.write(page(m["name"], f'<p>Moved to <a href="{target}">{esc(m["name"])}</a>.</p>',
                             path=f"/mob/{slug(m['name'])}",
                             extra_head=f'<meta http-equiv="refresh" content="0; url={target}">'
                                        f'<link rel="canonical" href="{SITE_URL}{target}">'))
            continue
        if m["name"] in mob_key:
            continue
        info = ", ".join(x for x in (f'Level {m["level"]}' if m["level"] else "",
                                     m["elite"].replace("rareelite", "rare elite").capitalize()) if x)
        body = f"""
<article>
  <div class="crumbs"><a href="/mobs">Mobs</a> / {esc(main_zone(m))}</div>
  <h1>{esc(m['name'])}</h1>
  <p class="muted">{esc(info)}</p>
  {mob_drops_html(m)}
</article>"""
        with open(os.path.join(out, "mob", f"{slug(m['name'])}.html"), "w", encoding="utf-8") as f:
            f.write(page(m["name"], body, path=f"/mob/{slug(m['name'])}",
                         description=f"{m['name']} drops on Vanilla Plus, seen by Lokiy's loot tracker (very rough)."))
    blocks = []
    for zone in sorted(by_zone, key=lambda z: (z == "Unknown zone", z)):
        lis = "".join(f'<li><a href="{mob_href(m)}">{esc(m["name"])}</a> <span class="muted">'
                      f'{m["kills"]} kills, {len(m["items"])} item{"s" if len(m["items"]) != 1 else ""}</span></li>'
                      for m in sorted(by_zone[zone], key=lambda m: m["name"]))
        blocks.append(f'<section class="cat"><h2>{esc(zone)}</h2><ul class="loot cols wide-list">{lis}</ul></section>')
    body = f"""
<section class="intro"><h1>Mobs &amp; bosses</h1>
<p class="muted">Every mob and boss <a href="https://lokiy.dev/loot/">Lokiy's loot tracker</a> has seen, with what
it dropped. Mobs link to their <a href="/npcs">NPC page</a> when their creature ID is known, else to their loot table
or their own page.</p>{rough_note}</section>
{"".join(blocks)}"""
    with open(os.path.join(out, "mobs.html"), "w", encoding="utf-8") as f:
        f.write(page("Mobs", body, path="/mobs", description="Vanilla Plus mobs and bosses and their drops (very rough)."))
    # lokiy.dev/loot links each mob name to its page here
    os.makedirs(os.path.join(out, "api"), exist_ok=True)
    with open(os.path.join(out, "api", "mobs.json.tmp"), "w", encoding="utf-8") as f:
        json.dump({m["name"]: SITE_URL + mob_href(m) for m in mob_loot}, f, ensure_ascii=False, separators=(",", ":"))
    os.replace(os.path.join(out, "api", "mobs.json.tmp"), os.path.join(out, "api", "mobs.json"))

    # ---- NPC pages /npc/<creature id>, the NPC search and api/npcs.json (npcs.py; Part GJ)
    def item_cell(i):
        if i in items:
            return item_link(items[i], icons.get(i, "inv_misc_questionmark"))
        w = wowhead.get(i, {"name": f"Item {i}", "quality": 1})
        return (f'<a class="item ext" href="https://www.wowhead.com/classic/item={i}" rel="noopener">'
                f'<span style="color:{Q_COLOR.get(w["quality"] or 1, "#ffffff")}">{esc(w["name"])}</span></a>')

    def zone_crumb(zone):
        zid = quests.zone_by_name.get(zone)
        return f' / <a href="/zone/{zid}">{esc(zone)}</a>' if zid else (f' / {esc(zone)}' if zone else "")

    npc_api, npc_index = {}, []
    for nid, n in sorted(npcs.items()):
        m = npc_mob.get(nid)
        keys = npc_tables.get(nid, [])
        rank = RANKS.get(n.get("rank"), "")
        boss = n.get("rank") == 3 or bool(keys) or (m and m["elite"] == "boss")
        if boss and not rank:
            rank = "Boss"
        zone = n["zone"] or (main_zone(m) if m and m["zones"] else "") or (table_label(keys[0])[1] if keys else "")
        level = n["lvl"] or (m["level"] if m else "")
        info = " &middot; ".join(esc(x) for x in (f"Level {level}" if level else "", rank, zone) if x)
        parts = journal_html(nid)
        if m:
            parts.append(mob_drops_html(m, "Lokiy's loot"))
        for key in keys:
            lis = "".join(f'<li>{item_cell(r[0])}</li>' for r in tables[key] if r[0])
            parts.append(f'<section class="card"><h2>Loot table: <a href="/loot/{key}">{esc(table_label(key)[0])}</a></h2>'
                         f'<ul class="loot">{lis}</ul></section>')
        drops = n.get("drops", [])
        if drops:
            trs = "".join(f'<tr><td>{item_cell(i)}</td><td class="rate">{c:g}%</td></tr>' for i, c in drops[:80])
            more = f'<p class="muted small">{len(drops) - 80} more with lower chances.</p>' if len(drops) > 80 else ""
            parts.append(f'<section class="card"><h2>Classic drops</h2><table class="obs-table"><thead><tr><th>Item</th>'
                         f'<th>Chance</th></tr></thead><tbody>{trs}</tbody></table>{more}<p class="muted small">Classic '
                         f'1.12 data (pfQuest / vmangos), not checked against Vanilla+; shared world-drop pools and '
                         f'raid loot tables aren\'t listed here.</p></section>')
        if quests.npc_card(nid):
            parts.append(quests.npc_card(nid))
        if n.get("sells"):
            lis = "".join(f'<li>{item_cell(i)}</li>' for i in n["sells"][:150])
            parts.append(f'<section class="card"><h2>Sells</h2><ul class="loot cols">{lis}</ul>'
                         f'<p class="muted small">Classic 1.12 data (pfQuest).</p></section>')
        body = f"""
<article>
  <div class="crumbs"><a href="/npcs">NPCs</a>{zone_crumb(zone)}</div>
  <h1>{esc(n['name'])}</h1>
  <p class="muted">{info}{' &middot; ' if info else ''}NPC {nid} &middot;
  {'<span class="tag">Vanilla+ NPC</span> creature ID from combat logs' if n.get('vplus') else
   f'<a href="https://www.wowhead.com/classic/npc={nid}" rel="noopener">Wowhead</a>'}</p>
  {''.join(parts) or '<p class="muted">No loot or vendor data.</p>'}
</article>"""
        with open(os.path.join(out, "npc", f"{nid}.html"), "w", encoding="utf-8") as f:
            f.write(page(n["name"], body, path=f"/npc/{nid}",
                         description=f"{n['name']}" + (f", level {level}" if level else "") + (f", {zone}" if zone else "")
                         + " on Vanilla Plus."))
        npc_api[nid] = {"name": n["name"], "level": level, "zone": zone, "boss": bool(boss), "url": f"{SITE_URL}/npc/{nid}"}
        npc_index.append([nid, n["name"], level, zone, rank, 1 if (m or keys or drops) else 0])
    with open(os.path.join(out, "npcs.json"), "w", encoding="utf-8") as f:
        json.dump(npc_index, f, ensure_ascii=False, separators=(",", ":"))
    with open(os.path.join(out, "api", "npcs.json.tmp"), "w", encoding="utf-8") as f:
        json.dump(npc_api, f, ensure_ascii=False, separators=(",", ":"))
    os.replace(os.path.join(out, "api", "npcs.json.tmp"), os.path.join(out, "api", "npcs.json"))
    unmatched = "".join(f'<li>{esc(nm)} <span class="muted">{esc(why)}</span></li>'
                        for nm, why in sorted(matcher.unmatched.items()))
    body = f"""
<section class="intro"><h1>NPCs</h1>
<p class="muted">{len(npc_api)} NPCs by creature ID, from pfQuest's classic data (level, rank, zone, classic drops and
vendor items), with <a href="https://lokiy.dev/loot/">Lokiy's loot tracker</a> and the AtlasLoot boss tables where the
name matched one creature ID. The loot tracker's mobs by zone: <a href="/mobs">Mobs</a>.</p>
<form class="spell-search" onsubmit="return false">
  <input id="nq" type="search" placeholder="Name or ID..." autocomplete="off">
  <select id="nr"><option value="">Any rank</option><option>Boss</option><option>Elite</option><option>Rare elite</option><option>Rare</option></select>
  <label class="muted"><input id="nl" type="checkbox"> with loot data</label>
</form></section>
<section class="card"><p class="muted small" id="ncount"></p><ul class="loot" id="nlist"></ul></section>
<section class="card"><details><summary><h2 style="display:inline">Names not matched to a creature ID
<span class="muted">{len(matcher.unmatched)}</span></h2></summary><p class="muted small">Loot tracker mobs and AtlasLoot
tables whose name isn't one pfQuest NPC: Vanilla+'s own NPCs, other names, or several NPCs of that name with no zone or
drop to tell them apart. They keep their /mob or /loot page.</p><ul class="loot cols">{unmatched}</ul></details></section>
<script>
(async () => {{
  const all = await (await fetch('/npcs.json')).json();
  const $ = id => document.getElementById(id), MAX = 300;
  const esc = t => String(t).replace(/[&<>"]/g, c => ({{'&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;'}})[c]);
  const p = new URLSearchParams(location.search);
  $('nq').value = p.get('q') || ''; $('nr').value = p.get('rank') || ''; $('nl').checked = p.get('loot') === '1';
  function show() {{
    const q = $('nq').value.trim().toLowerCase(), r = $('nr').value, l = $('nl').checked;
    const hits = all.filter(n => (!r || n[4] === r) && (!l || n[5])
      && (!q || String(n[0]) === q || n[1].toLowerCase().includes(q) || n[3].toLowerCase().includes(q)));
    $('ncount').textContent = `${{hits.length}} NPCs` + (hits.length > MAX ? ` (first ${{MAX}} shown)` : '');
    $('nlist').innerHTML = hits.slice(0, MAX).map(n => `<li><a href="/npc/${{n[0]}}">${{esc(n[1])}}</a> <span class="muted">`
      + [n[2] ? 'Level ' + n[2] : '', n[4], n[3], '#' + n[0]].filter(Boolean).map(esc).join(' · ') + '</span></li>').join('');
    const u = new URLSearchParams(); if (q) u.set('q', q); if (r) u.set('rank', r); if (l) u.set('loot', '1');
    history.replaceState(null, '', u.toString() ? '?' + u : location.pathname);
  }}
  ['nq', 'nr', 'nl'].forEach(id => $(id).addEventListener('input', show));
  show();
}})();
</script>"""
    with open(os.path.join(out, "npcs.html"), "w", encoding="utf-8") as f:
        f.write(page("NPCs", body, path="/npcs", description=f"{len(npc_api)} NPCs on Vanilla Plus."))

    # ---- quest and zone pages (quests.py; Part GO)
    zone_tables = {}
    for key in loot_keys:
        nm, cat = names.get(key, ("", "Other"))
        if nm and table_groups.get(key) in ("Instances", "WorldBosses"):
            zone_tables.setdefault(cat, []).append((key, nm))
    n_quests, n_zones = quests.write(out, page, item_cell, npc_api, zone_tables)
    print(f"{n_quests} quest pages, {n_zones} zone pages")

    # ---- spell pages, the spell search and api/spells.json (spells.py; Part GE); spells loaded above
    spell_items = {}  # spell id -> [(item id, how: "Use" / "Equip" / ...)]
    for iid in all_items:
        for ln, sid in items[iid].get("spell_lines", {}).items():
            how = ln.split(":", 1)[0]
            spell_items.setdefault(sid, []).append((iid, how))
    item_names = {i: esc(items[i]["name"]) for i in all_items}

    def field_rows(rows):
        return "".join(f'<tr><th>{esc(k)}</th><td>{link_tokens(esc(v), item_names)}</td></tr>' for k, v in rows)

    for sid, sp in spells.items():
        title = sp["name"] + (f" ({sp['rank']})" if sp["rank"] else "")
        tip = (f'<div class="tooltip spell-tt"><div class="tt-name">{esc(sp["name"])}'
               + (f' <span class="tt-grey">{esc(sp["rank"])}</span>' if sp["rank"] else "") + '</div>'
               + (f'<div class="tt-gold">{esc(sp["desc"])}</div>' if sp["desc"] else "")
               + (f'<div class="tt-grey">Buff: {esc(sp["tip"])}</div>' if sp["tip"] else "") + '</div>')
        by_label = dict(sp["fields"])
        main = [(k, by_label[k]) for k in MAIN_FIELDS if by_label.get(k)]
        effects, rest = {}, []
        for k, v in sp["fields"]:
            m = re.match(r"^Effect (\d) · (.+)$", k)
            if m:
                effects.setdefault(m.group(1), []).append((m.group(2), v))
            elif k not in MAIN_FIELDS:
                rest.append((k, v))
        eff_html = "".join(f'<h3>Effect {n}</h3><table class="spell-fields">{field_rows(rows)}</table>'
                           for n, rows in sorted(effects.items()))
        status = sp["status"]
        badge = f'<span class="tag spell-{status}">{STATUS_LABEL[status]}</span>' if status else ""
        diff_html = ""
        if sp["diff"]:
            trs = "".join(f'<tr><th>{esc(r[0])}</th><td class="old">{link_tokens(esc(r[1] or "—"), item_names)}</td>'
                          f'<td>{link_tokens(esc(r[2] or "—"), item_names)}'
                          + (f'<div class="muted small">{esc(r[3])}</div>' if len(r) > 3 and r[3] else "") + '</td></tr>'
                          for r in sp["diff"])
            diff_html = (f'<section class="card"><h2>Changed from classic</h2><table class="spell-diff"><thead><tr><th></th>'
                         f'<th>Classic 1.12.1</th><th>Vanilla+</th></tr></thead><tbody>{trs}</tbody></table></section>')
        used = spell_items.get(sid, [])
        used_html = ""
        if used:
            lis = "".join(f'<li>{item_link(items[i], icons.get(i, "inv_misc_questionmark"))} '
                          f'<span class="muted">{esc(how)}</span></li>' for i, how in sorted(used))
            used_html = f'<section class="card"><h2>Used by items ({len(used)})</h2><ul class="loot">{lis}</ul></section>'
        body = f"""
<article class="spell-page">
  <div class="item-head">{icon_img(sp['icon'] or 'inv_misc_questionmark', 'large', 'icon-l')}
    <div><h1>{esc(sp['name'] or '(no name)')} {f'<span class="muted">{esc(sp["rank"])}</span>' if sp['rank'] else ''}</h1>
    <div class="muted">Spell {sid}{' &middot; ' + esc(sp['class']) if sp['class'] else ''} {badge}
    &middot; <a href="https://lokiy.dev/spells/#view=db&amp;id={sid}">all fields &amp; changelog</a></div></div>
  </div>
  {tip}
  <section class="card"><h2>Details</h2><table class="spell-fields">{field_rows(main) or '<tr><td class="muted">No details.</td></tr>'}</table>
  {eff_html}
  <details><summary class="muted">All other fields ({len(rest)})</summary><table class="spell-fields">{field_rows(rest)}</table></details></section>
  {diff_html}
  {used_html}
</article>"""
        with open(os.path.join(out, "spell", f"{sid}.html"), "w", encoding="utf-8") as f:
            f.write(page(title, body, description=sp["desc"] or title, image=icon_url(sp["icon"], "large", absolute=True),
                         path=f"/spell/{sid}"))
    n_status = {k: sum(1 for sp in spells.values() if sp["status"] == k) for k in STATUS_LABEL}
    with open(os.path.join(out, "spells.json"), "w", encoding="utf-8") as f:
        json.dump([[sid, sp["name"], sp["rank"], sp["icon"], sp["status"], sp["class"]] for sid, sp in spells.items()],
                  f, ensure_ascii=False, separators=(",", ":"))
    body = f"""
<section class="intro"><h1>Spells</h1>
<p class="muted">Every spell of the Vanilla+ client ({len(spells)}): {n_status['new']} new on Vanilla+,
{n_status['changed']} changed from classic 1.12.1, {n_status['same']} the same. From the client's spell data, as on
<a href="https://lokiy.dev/spells/">lokiy.dev/spells</a> (the changelog per patch).</p>
<form class="spell-search" onsubmit="return false">
  <input id="sq" type="search" placeholder="Name or ID..." autocomplete="off">
  <select id="ss"><option value="">All</option><option value="new">New on Vanilla+</option>
    <option value="changed">Changed from classic</option><option value="same">Same as classic</option></select>
  <select id="sc"><option value="">Any class</option></select>
</form></section>
<section class="card"><p class="muted small" id="scount"></p><ul class="loot" id="slist"></ul></section>
<script>
(async () => {{
  const all = await (await fetch('/spells.json')).json();
  const $ = id => document.getElementById(id), LABEL = {json.dumps(STATUS_LABEL)}, MAX = 300;
  const classes = [...new Set(all.map(s => s[5]).filter(Boolean))].sort();
  $('sc').innerHTML += classes.map(c => `<option>${{c}}</option>`).join('');
  const esc = t => String(t).replace(/[&<>"]/g, c => ({{'&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;'}})[c]);
  const p = new URLSearchParams(location.search);
  $('sq').value = p.get('q') || ''; $('ss').value = p.get('status') || ''; $('sc').value = p.get('class') || '';
  function show() {{
    const q = $('sq').value.trim().toLowerCase(), st = $('ss').value, cl = $('sc').value;
    const hits = all.filter(s => (!st || s[4] === st) && (!cl || s[5] === cl)
      && (!q || String(s[0]) === q || s[1].toLowerCase().includes(q)));
    $('scount').textContent = `${{hits.length}} spells` + (hits.length > MAX ? ` (first ${{MAX}} shown)` : '');
    $('slist').innerHTML = hits.slice(0, MAX).map(s => `<li><a class="item" href="/spell/${{s[0]}}">`
      + `<img class="icon-s" src="/icons/${{s[3] || 'inv_misc_questionmark'}}_s.jpg" alt="" loading="lazy" `
      + `onerror="this.onerror=null;this.src='/icons/inv_misc_questionmark_s.jpg'">`
      + `<span>${{esc(s[1] || '(no name)')}}</span></a> <span class="muted">${{esc(s[2])}} #${{s[0]}}`
      + `${{s[5] ? ' · ' + esc(s[5]) : ''}}</span>${{s[4] && s[4] !== 'same' ? ` <span class="tag spell-${{s[4]}}">${{LABEL[s[4]]}}</span>` : ''}}</li>`).join('');
    const u = new URLSearchParams(); if (q) u.set('q', q); if (st) u.set('status', st); if (cl) u.set('class', cl);
    history.replaceState(null, '', u.toString() ? '?' + u : location.pathname);
  }}
  ['sq', 'ss', 'sc'].forEach(id => $(id).addEventListener('input', show));
  show();
}})();
</script>"""
    with open(os.path.join(out, "spells.html"), "w", encoding="utf-8") as f:
        f.write(page("Spells", body, path="/spells", description=f"All {len(spells)} Vanilla Plus spells."))
    api_spells = {}
    for sid, sp in spells.items():
        e = {"name": sp["name"], "rank": sp["rank"], "icon": icon_url(sp["icon"], "medium", absolute=True)
             if sp["icon"] else "", "desc": sp["desc"], "tip": sp["tip"], "status": sp["status"]}
        e.update(tooltip_lines(sp["fields"]))  # cost / range / cast / cd, game-style text, only when set (Part GI)
        api_spells[sid] = e
    os.makedirs(os.path.join(out, "api"), exist_ok=True)
    with open(os.path.join(out, "api", "spells.json.tmp"), "w", encoding="utf-8") as f:
        json.dump(api_spells, f, ensure_ascii=False, separators=(",", ":"))
    os.replace(os.path.join(out, "api", "spells.json.tmp"), os.path.join(out, "api", "spells.json"))

    # ---- /removed: every item marked "not on Vanilla+", by reason, with the rule that marked it (for Lokiy to check)
    def removed_row(iid, why):
        it = items[iid]
        return (f'<tr><td><a class="item" href="/item/{iid}" data-item="{iid}">'
                f'{icon_img(icons.get(iid, "inv_misc_questionmark"), "small", "icon-s")}'
                f'<span style="color:{Q_COLOR.get(it["quality"], "#ffffff")}">{esc(it["name"])}</span></a></td>'
                f'<td>{iid}</td><td class="muted rule">{esc(why)}</td></tr>')

    by_reason = {}
    for iid, text in NOT_ON_SERVER.items():
        by_reason.setdefault(text, []).append(iid)
    head = '<thead><tr><th>Item</th><th>ID</th><th>Marked by</th></tr></thead>'
    check = "".join(removed_row(i, why) for i, why in TO_CHECK.items() if i in items)
    blocks = [f'<section class="card"><h2>To check: not marked yet ({len(TO_CHECK)})</h2>'
              f'<table class="obs-table removed-table">{head}<tbody>{check}</tbody></table></section>'] if TO_CHECK else []
    toc = []
    for n, (text, ids) in enumerate(sorted(by_reason.items(), key=lambda kv: (-len(kv[1]), kv[0]))):
        rows = "".join(removed_row(i, srcs.mark_rule.get(i, "")) for i in sorted(ids))
        toc.append(f'<li><a href="#r{n}">{esc(text)}</a> <span class="muted">{len(ids)}</span></li>')
        blocks.append(f'<section class="card" id="r{n}"><h2>{esc(text)} <span class="muted">{len(ids)}</span></h2>'
                      f'<table class="obs-table removed-table">{head}<tbody>{rows}</tbody></table></section>')
    body = f"""
<section class="intro"><h1>Items not on Vanilla+</h1>
<p class="muted">Every item the database marks "not on Vanilla+" ({len(NOT_ON_SERVER)}), grouped by reason. "Marked by" is
the rule that marked it: an AtlasLoot table of content that isn't on the server, the workbook's AQ / Naxxramas names
list or Removed Items sheet, the set-bonus rule (the 3/5/8 version of a set is the old one), Lokiy's answers, or
Wowhead sources that are all in cut content. An item Lokiy's loot tracker has seen drop is never marked.</p>
<ul class="removed-toc">{''.join(toc)}</ul></section>
{''.join(blocks)}"""
    with open(os.path.join(out, "removed.html"), "w", encoding="utf-8") as f:
        f.write(page("Not on Vanilla+", body, path="/removed",
                     description=f"{len(NOT_ON_SERVER)} items that aren't on Vanilla Plus, by reason."))

    # ---- /unseen: classic items no Vanilla+ client has seen yet (Part GG), by where they come from
    UNSEEN_ORDER = ["raid", "world boss", "dungeon", "rare mob", "drop", "world drop", "container", "quest", "vendor",
                    "crafting", "reputation", "pvp", "event", "set", "other", "none"]
    req_level = lambda it: next((int(m.group(1)) for ln in it["lines"] for m in [re.match(r"^Requires Level (\d+)$", ln)]
                                 if m), 0)

    def unseen_row(iid):
        it = items[iid]
        srcs_html = "; ".join((f'<a href="{esc(e["href"])}">{esc(e["text"])}</a>' if e.get("href", "").startswith("/")
                               else esc(e["text"])) for e in map(local_source, item_srcs[iid][:3])) or '<span class="muted">unknown</span>'
        lv = " / ".join(x for x in (f'ilvl {it.get("ilvl")}' if it.get("ilvl") else "",
                                    f'req {req_level(it)}' if req_level(it) else "") if x)
        return (f'<tr><td><a class="item" href="/item/{iid}" data-item="{iid}">'
                f'{icon_img(icons.get(iid, "inv_misc_questionmark"), "small", "icon-s")}'
                f'<span style="color:{Q_COLOR.get(it["quality"], "#ffffff")}">{esc(it["name"])}</span></a></td>'
                f'<td>{iid}</td><td class="muted">{lv}</td><td class="muted">{esc(types[iid][1])}</td>'
                f'<td class="rule">{srcs_html}</td></tr>')

    unseen = [i for i in classic_only if i not in NOT_ON_SERVER and i not in LOKIY_DROPPED_IDS]
    unseen_groups = {}
    for i in unseen:
        unseen_groups.setdefault(item_srcs[i][0]["type"] if item_srcs[i] else "none", []).append(i)
    head = '<thead><tr><th>Item</th><th>ID</th><th>Level</th><th>Type</th><th>Source</th></tr></thead>'
    toc, blocks = [], []
    for kind_ in sorted(unseen_groups, key=lambda k: UNSEEN_ORDER.index(k) if k in UNSEEN_ORDER else 99):
        ids = sorted(unseen_groups[kind_], key=lambda i: (-items[i]["quality"], items[i]["name"]))
        label = "No known source" if kind_ == "none" else TYPE_LABEL.get(kind_, kind_.capitalize())
        anchor = slug(label)
        toc.append(f'<li><a href="#{anchor}">{esc(label)}</a> <span class="muted">{len(ids)}</span></li>')
        blocks.append(f'<section class="card" id="{anchor}"><h2>{esc(label)} <span class="muted">{len(ids)}</span></h2>'
                      f'<table class="obs-table removed-table unseen-table">{head}'
                      f'<tbody>{"".join(unseen_row(i) for i in ids)}</tbody></table></section>')
    cut = sorted((i for i in classic_only if i in NOT_ON_SERVER), key=lambda i: (-items[i]["quality"], items[i]["name"]))
    if cut:
        blocks.append(f'<section class="card"><details><summary><h2 style="display:inline">Also not seen, but not on '
                      f'Vanilla+ <span class="muted">{len(cut)}</span></h2></summary><p class="muted small">Classic items of '
                      f'content that isn\'t on the server (reasons on <a href="/removed">Not on Vanilla+</a>).</p>'
                      f'<table class="obs-table removed-table unseen-table">{head}'
                      f'<tbody>{"".join(unseen_row(i) for i in cut)}</tbody></table></details></section>')
    body = f"""
<section class="intro"><h1>Not seen on Vanilla+ yet</h1>
<p class="muted">{len(unseen)} classic items that exist on the server (Lokiy: a classic item is on Vanilla+ unless it's
marked <a href="/removed">not on Vanilla+</a>) but are in no uploaded item cache yet, so nobody's client has seen them
on Vanilla+. Their tooltips are classic 1.12 (Wowhead) and may differ on the server. An item leaves this page as
soon as someone's item cache (<a href="https://lokiy.dev/admin/upload">lokiy.dev/admin/upload</a>) contains it.
Grouped by where it comes from; also in the <a href="/search?st=unseen">search</a>.</p>
<ul class="removed-toc">{''.join(toc)}</ul></section>
{''.join(blocks)}"""
    with open(os.path.join(out, "unseen.html"), "w", encoding="utf-8") as f:
        f.write(page("Not seen on Vanilla+ yet", body, path="/unseen",
                     description=f"{len(unseen)} classic items not seen on Vanilla Plus yet, by source."))

    # ---- instances: every loot table by instance / group
    blocks = []
    for cat in cats_all:
        lis = []
        for key in by_cat_all[cat]:
            name, _ = table_label(key)
            n_new = sum(1 for r in tables[key] if is_custom(r[0]) or r[0] in reworked)
            badge = f' <span class="tag">{n_new} new/changed</span>' if n_new else ""
            if srcs.cut_table(key):
                badge += ' <span class="tag nos-tag">not on Vanilla+</span>'
            lis.append(f'<li><a href="/loot/{key}">{esc(name)}</a>{badge}</li>')
        blocks.append(f'<section class="cat"><h2>{esc(cat)}</h2><ul class="loot cols wide-list">{"".join(lis)}</ul></section>')
    body = f"""
<section class="intro"><h1>Instances &amp; loot tables</h1>
<p class="muted">Every AtlasLoot table of the server. The tag counts the new and changed items in it.
Items of content that isn't on the server are listed on <a href="/removed">Not on Vanilla+</a>.</p></section>
{''.join(blocks)}"""
    with open(os.path.join(out, "instances.html"), "w", encoding="utf-8") as f:
        f.write(page("Instances", body, path="/instances", description="Vanilla Plus loot tables by instance and boss."))
    with open(os.path.join(out, "tables.json"), "w", encoding="utf-8") as f:  # the header search's loot tables (Part GQ)
        json.dump([[key, table_label(key)[0], cat, 1 if srcs.cut_table(key) else 0] for cat in cats_all for key in by_cat_all[cat]],
                  f, ensure_ascii=False, separators=(",", ":"))

    # ---- type pages: every item, by type and section
    for page_key, page_name in PAGES:
        ids = [i for i in browsable if types[i][0] == page_key]
        if not ids:
            continue
        secs = {}
        for i in ids:
            secs.setdefault(types[i][1], []).append(i)
        order = sorted(secs, key=lambda s: (SECTION_ORDER.get(s, 50), s))
        toc = " &middot; ".join(f'<a href="#{slug(s)}">{esc(s)}</a> <span class="muted">{len(secs[s])}</span>' for s in order)
        blocks = []
        for s in order:
            lst = sorted(secs[s], key=lambda i: (-items[i]["quality"], items[i]["name"]))
            lis = []
            for i in lst:
                tag = '<span class="tag">new</span>' if is_custom(i) else ('<span class="tag">changed</span>' if i in reworked else "")
                lis.append(f'<li>{item_link(items[i], icons.get(i))}{tag}</li>')
            blocks.append(f'<section class="cat" id="{slug(s)}"><h2>{esc(s)}</h2><div class="table wide"><ul class="loot cols">'
                          f'{"".join(lis)}</ul></div></section>')
        tabs = " ".join(f'<a class="{"on" if k == page_key else ""}" href="/type/{k}">{esc(nm)}</a>'
                        for k, nm in PAGES if type_counts.get(k))
        body = f"""
<nav class="type-tabs">{tabs}</nav>
<section class="intro"><h1>{esc(page_name)}</h1>
<p class="muted">{len(ids)} items, best quality first. Tags: <span class="tag">new</span> on Vanilla Plus,
<span class="tag">changed</span> from classic.</p><p class="toc">{toc}</p></section>
{''.join(blocks)}"""
        with open(os.path.join(out, "type", f"{page_key}.html"), "w", encoding="utf-8") as f:
            f.write(page(page_name, body, path=f"/type/{page_key}", description=f"{len(ids)} Vanilla Plus {page_name.lower()}."))

    # ---- advanced search: /search.html + /search.json (search.js filters it in the browser)
    sec_names = sorted({types[i][1] for i in browsable})
    sec_idx = {s: n for n, s in enumerate(sec_names)}
    type_idx = {k: n for n, (k, _) in enumerate(PAGES)}
    cat_idx = {c: n for n, c in enumerate(cats_all)}
    src_names, src_idx = [], {}  # drop sources (loot table names, as in items.json "s"), so the name box matches them too

    def src_no(nm):
        if nm not in src_idx:
            src_idx[nm] = len(src_names)
            src_names.append(nm)
        return src_idx[nm]

    srows = []
    for i in browsable:
        it = items[i]
        pk, sec = types[i]
        info = parse_stats(it, pk, sec)
        status = (4 if i in classic_only and i not in LOKIY_DROPPED_IDS else 1 if i in classic_only else 3 if is_custom(i) else 2 if i in reworked
                  else 0 if kind(i).endswith("same as classic") else 1)
        cats = sorted({cat_idx[table_label(k)[1]] for k in sources.get(i, []) if table_label(k)[1] in cat_idx})
        ic = icon_url(icons.get(i), "small")  # /icons/<name>_s.jpg
        srows.append([i, it["name"], it["quality"], type_idx[pk], sec_idx[sec], info.get("lv", 0), info.get("sl", ""),
                      info.get("b", ""), info.get("cl", ALL_CLASSES), status, info.get("st", 0), cats or 0, ic,
                      1 if affixes.get(i) or afx_rows_of(i) or dkp_suffixes.get(i) or seen_affixes(it["name"], loot_names.get(i, {})) else 0,
                      affix_stats.get(i, 0),
                      [src_no(nm) for nm in dict.fromkeys(table_label(k)[0] for k in sources.get(i, []))] or 0])
    with open(os.path.join(out, "search.json"), "w", encoding="utf-8") as f:
        json.dump({"types": PAGES, "sections": sec_names, "cats": cats_all, "classes": CLASSES,
                   "stats": STAT_LABELS, "srcs": src_names, "items": srows}, f, separators=(",", ":"))
    body = """
<section class="intro"><h1>Item search</h1>
<p class="muted">Filter all items. Stats are read from the in-game tooltips; set bonuses are not counted. The URL keeps
your filters, so you can share a search.</p></section>
<form id="filters" class="card filters loading" autocomplete="off">
  <label class="f-name">Name or source<input name="q" type="search" placeholder="Item name, id, or where it drops (boss, table)"></label>
  <fieldset class="f-quality"><legend>Quality</legend><div id="qualities" class="chips"></div></fieldset>
  <label>Type<select name="type"></select></label>
  <label>Subtype<select name="sec"></select></label>
  <label>Slot<select name="slot"></select></label>
  <label>Usable by<select name="cls"></select></label>
  <label>Required level<span class="range"><input name="lmin" type="number" min="0" max="60" placeholder="min" aria-label="Minimum level"><span class="muted">-</span><input name="lmax" type="number" min="0" max="60" placeholder="max" aria-label="Maximum level"></span></label>
  <label>Binds<select name="bind"></select></label>
  <label>Vanilla Plus<select name="st"><option value="">Any</option><option value="vp">New or changed</option><option value="new">New items</option><option value="changed">Changed from classic</option><option value="same">Same as classic</option><option value="unseen">Not seen on Vanilla+ yet</option></select></label>
  <label>Drops in<select name="src"></select></label>
  <label>Random affixes<select name="ra"><option value="">Any</option><option value="1">Has random affixes</option><option value="0">No random affixes</option></select></label>
  <fieldset class="f-stats"><legend>Stats (item has the stat, at least the value)</legend><div id="stat-rows"></div>
    <label class="check"><input type="checkbox" name="aff" checked> Also match random affixes <span class="muted">(shown as <span class="aff">17 <small>1.7%</small></span>: best affix value and the roll chance on Wowhead; &lt; = rarer than Wowhead lists, ~ = estimate)</span></label>
    <div class="f-actions"><button type="button" id="add-stat">+ Add stat filter</button><button type="button" id="reset">Reset all</button></div></fieldset>
</form>
<div id="results-top" class="results-head"><strong id="count">Loading...</strong><div class="pager"></div></div>
<div class="table-wrap"><table id="results-table" class="results"></table></div>
<div class="results-foot"><div class="pager"></div></div>"""
    with open(os.path.join(out, "search.html"), "w", encoding="utf-8") as f:
        f.write(page("Item search", body, path="/search",
                     description="Search all Vanilla Plus items by quality, type, slot, level, class and stats.",
                     extra_head='<script src="/search.js" defer></script>'))

    # ---- 404
    with open(os.path.join(out, "404.html"), "w", encoding="utf-8") as f:
        f.write(page("Not found", '<article><h1>Not found</h1><p class="muted">That item or page is not in the database.</p>'
                     '<p><a href="/">Back to all items</a></p></article>'))

    # ---- data: search index (small, all items) and hover tooltips in chunks of 500 ids (/tt/<id // 500>.json)
    data, chunks = [], {}
    for i in browsable:
        d = {"id": i, "n": items[i]["name"], "q": items[i]["quality"], "i": icon_url(icons.get(i), "small")}
        src = [table_label(k)[0] for k in sources.get(i, [])]
        if src:
            d["s"] = src
        if is_custom(i):
            d["c"] = 1
        elif i in reworked:
            d["r"] = 1
        data.append(d)
        chunks.setdefault(i // 500, {})[i] = tooltip_html(items[i])
    with open(os.path.join(out, "items.json"), "w", encoding="utf-8") as f:
        json.dump(data, f, separators=(",", ":"))

    # ---- data for other sites: every item, the one place they get item data from
    api = {}
    for i in all_items:
        it = items[i]
        e = {"name": it["name"], "q": it["quality"], "lines": [[line_class(ln), ln] for ln in it["lines"]]}
        if icons.get(i):
            e["icon"] = icon_url(icons[i], "medium", absolute=True)
        if is_custom(i):
            e["custom"] = 1
        if i in classic_only:
            e["classicOnly"] = 1  # classic tooltip; no Vanilla+ client has seen the item yet
        rows_i = afx_rows_of(i)
        if rows_i:  # [[suffix id, name, [effect lines]]] from the DBCs, incl. tracker-seen and group-borrowed ones
            e["affixes"] = [[sid, name, effs] for sid, name, effs, *_ in rows_i]
        api[i] = e
    os.makedirs(os.path.join(out, "api"), exist_ok=True)
    tmp = os.path.join(out, "api", "items.json.tmp")
    with open(tmp, "w", encoding="utf-8") as f:
        json.dump(api, f, ensure_ascii=False, separators=(",", ":"))
    os.replace(tmp, os.path.join(out, "api", "items.json"))
    # ---- where items come from, for other projects (the sim reads only this): AtlasLoot tables incl. the loot
    # tracker's drops, then the Wowhead classic sources (sim gen_db -dumpWowheadSources, UIItemSource JSON by item id)
    src_api = {"tables": {}, "wowhead": {}, "webdkp": {i: a for i, a in dkp_api.items() if i in items},
               "items": {i: [{k: v for k, v in e.items() if k != "href" or not v.startswith("/")} for e in item_srcs[i]]
                         for i in all_items if item_srcs[i]},
               "notOnServer": dict(sorted(NOT_ON_SERVER.items()))}
    for key, rows in tables.items():
        name, cat = names.get(key, ("", "Other"))
        src_api["tables"][key] = {"name": name, "instance": cat, "group": table_groups.get(key, ""),
                                  "items": [[r[0], r[2]] for r in rows if r[0]]}
    wowhead_sources = os.path.join(HERE, "wowhead_sources.json")
    if os.path.exists(wowhead_sources):
        with open(wowhead_sources, encoding="utf-8") as f:
            src_api["wowhead"] = json.load(f)
    # crafted / reputation sources from AtlasLoot's own tables replace Wowhead's of that kind (Part HB); the key stays
    # "wowhead" so the sim (docs/gen_sources.py) reads them unchanged
    atlas = atlas_item_sources(tables, table_groups, names, ROW_INFO, craft_items())
    src_api["wowhead"], moved = merge_atlas_sources(src_api.get("wowhead", {}), atlas)
    print("api/sources.json from AtlasLoot:", ", ".join(f"{k} {v}" for k, v in sorted(moved.items())))
    tmp = os.path.join(out, "api", "sources.json.tmp")
    with open(tmp, "w", encoding="utf-8") as f:
        json.dump(src_api, f, ensure_ascii=False, separators=(",", ":"))
    os.replace(tmp, os.path.join(out, "api", "sources.json"))
    for n, chunk in chunks.items():
        with open(os.path.join(out, "tt", f"{n}.json"), "w", encoding="utf-8") as f:
            json.dump(chunk, f, separators=(",", ":"))

    for asset in ("style.css", "site.js", "search.js"):
        shutil.copy(os.path.join(HERE, asset), os.path.join(out, asset))
    shutil.rmtree(os.path.join(out, "icons"), ignore_errors=True)
    for d in ICON_DIRS:
        if os.path.isdir(d):
            shutil.copytree(d, os.path.join(out, "icons"), dirs_exist_ok=True)

    print(f"{len(all_items)} items ({len(custom)} custom, {len(reworked)} reworked), {len(loot_keys)} loot tables, "
          f"{n_nosrc} custom / {n_rw_nosrc} reworked without source -> {out}")
    by_from = {}
    for i in all_items:
        f = item_srcs[i][0]["from"] if item_srcs[i] else "none"
        by_from[f] = by_from.get(f, 0) + 1
    print(f"{len(classic_only)} classic items not seen on Vanilla+ yet added "
          f"({sum(1 for i in classic_only if types[i][0] in ('weapons', 'ranged', 'cloth', 'leather', 'mail', 'plate', 'accessories', 'shields'))} gear)")
    print(f"{len(npc_api)} NPC pages; matched {len(mob_npc)}/{len(mob_loot)} tracker mobs, {len(table_npc)} boss tables; "
          f"{len(matcher.unmatched)} names not matched")
    print(f"sources by layer: {by_from}; {len(NOT_ON_SERVER)} items not on Vanilla+; {len(spells)} spells "
          f"({n_status}), {len(spell_items)} used by items")
    for iid, why in srcs.conflicts:
        print(f"  not marked (the loot tracker saw it drop): {iid} {items[iid]['name']}: {why}")


if __name__ == "__main__":
    main()
