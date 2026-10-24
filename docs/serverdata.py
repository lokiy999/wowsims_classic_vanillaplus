"""Shared helpers for the private-server item pipeline.

See docs/private-server-item-rules.md for the ruleset these implement.

  parse_vplus.py     -> custom_items.json + renumber.json  (stats + server IDs, from db.lokiy.dev's items)
  gen_include.py     -> included_items.json + add_items.json  (Rule 1 allowlist)
  gen_phases.py      -> item_phases.json  (Rule 3, 6-phase by drop location)
  gen_sources.py     -> item_sources.json  (every item's sources, from db.lokiy.dev's)
  gen_removed_items.py -> removed_items.json  (Rule 2/3 hard exclusions)

Item data and source data come only from db.lokiy.dev (api/items.json, api/sources.json): the sim reads no AtlasLoot,
Wowhead or other copy of its own.
"""
import json
import os
import re
import subprocess

REPO = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
# db.lokiy.dev's item list (tools/itemdb/gen_site.py), the only source of item data. VPLUS_ITEMS overrides the path.
ITEMS_JSON = os.environ.get("VPLUS_ITEMS", "/var/www/db/api/items.json")
# db.lokiy.dev's item sources (tools/itemdb/gen_site.py): its AtlasLoot tables, with the drops the loot tracker added,
# and the Wowhead classic sources. The only source data. VPLUS_SOURCES overrides the path.
SOURCES_JSON = os.environ.get("VPLUS_SOURCES", "/var/www/db/api/sources.json")
QUALITY_NAMES = {0: "Poor", 1: "Common", 2: "Uncommon", 3: "Rare", 4: "Epic", 5: "Legendary"}
DB_JSON = os.path.join(REPO, "assets/database/db.json")

# ---------------------------------------------------------------------------
# pristine sim DB (pre-pipeline) -- read from git so every tool is idempotent
# ---------------------------------------------------------------------------

def pristine_db():
    try:
        raw = subprocess.check_output(
            ["git", "show", "HEAD:assets/database/db.json"], cwd=REPO)
        return {i["id"]: i for i in json.loads(raw)["items"]}
    except Exception:
        return {i["id"]: i for i in json.load(open(DB_JSON))["items"]}


# ---------------------------------------------------------------------------
# db.lokiy.dev items (tooltip lines from the game client's item cache)
# ---------------------------------------------------------------------------

def item_lines():
    """{item_id: [name, tooltip lines..., "Rarity: <quality>"]} from db.lokiy.dev's api/items.json.

    Same line format as the old VPlusItemDB.lua dump (that parse_item() reads), with the Rarity line the dump ended
    each item with."""
    with open(ITEMS_JSON, encoding="utf8") as f:
        items = json.load(f)
    out = {}
    for iid, it in items.items():
        lines = [it["name"]] + [text for _, text in it["lines"]]
        if it.get("q") in QUALITY_NAMES:
            lines.append("Rarity: " + QUALITY_NAMES[it["q"]])
        out[int(iid)] = lines
    return out


# ---------------------------------------------------------------------------
# db.lokiy.dev item sources
# ---------------------------------------------------------------------------

# Tables we never read: AQ20 / AQ40 / Naxxramas, the Atiesh (Naxx questline) frame
# drop in Stratholme, the Scourge Invasion pre-Naxx world event, and all Alterac
# Valley content (drops, rep rewards, and the Stormpike/Frostwolf insignia trinkets).
IGNORED_TABLE = re.compile(
    r'^(AQ|NAX|Naxxramas|TheRuinsofAhnQiraj|TheTempleofAhnQiraj'
    r'|STRATAtiesh|WBScourgeInvasion|ScourgeInvasion'
    r'|AlteracValley|AVRep|Stormpike|Frostwolf)')
AQ_NAXX = IGNORED_TABLE  # backwards-compatible alias

# Phase buckets by AtlasLoot table-name prefix. None => ignore.
def table_bucket(t):
    if IGNORED_TABLE.match(t):
        return None
    if t.startswith(("ZG", "ZulGurub")) or t.startswith("Zandalar"):
        return 3
    if t.startswith(("MC", "MoltenCore", "Onyxia")):
        return 4
    if t.startswith(("BWL", "BlackwingLair")):
        return 5
    if t.startswith(("SM", "ScarletMonastery", "Scarlet")):
        return 6
    return 1  # every 5-man dungeon, Crafting, PvP, WorldEvents, non-ZG Factions


def _sources():
    with open(SOURCES_JSON, encoding="utf8") as f:
        return json.load(f)


def loot_tables():
    """Returns (id_tables, names) from db.lokiy.dev's loot tables (AtlasLoot's, plus the loot tracker's drops).

      id_tables : {item_id: set(table_name)}   -- every table an item appears in
      names     : set(lower-case item names the tables list)  -- for name matching

    WorldBosses tables get a synthetic 'WB' prefix so they bucket to phase 2.
    """
    id_tables = {}
    names = set()
    for key, t in _sources()["tables"].items():
        tname = "WB" + key if t["group"] == "WorldBosses" else key
        for iid, name in t["items"]:
            id_tables.setdefault(iid, set()).add(tname)
            if name and table_bucket(tname) is not None:   # skip AQ/Naxx/Scourge names
                names.add(name.strip().lower())
    return id_tables, names


def table_info():
    """{table_name ('WB' prefix like loot_tables()): (display name, AtlasLoot folder)} of db.lokiy.dev's loot tables."""
    return {("WB" + key if t["group"] == "WorldBosses" else key): (t["name"], t["group"])
            for key, t in _sources()["tables"].items()}


def wowhead_sources():
    """{item_id: [UIItemSource dicts]}: db.lokiy.dev's Wowhead classic sources, for what its loot tables don't cover."""
    return {int(k): v for k, v in _sources()["wowhead"].items()}


def wb_prefix_ok(t):
    """WorldBosses table -> phase 2 (all of them: dragons, Kazzak, Azuregos,
    Hederine, Kurinnaxx)."""
    return t.startswith("WB")


# convenience: bucket a set of table names (None entries dropped) -> min phase
def phase_from_tables(tables):
    buckets = set()
    for t in tables:
        if t.startswith("WB"):
            buckets.add(2)
            continue
        b = table_bucket(t)
        if b is not None:
            buckets.add(b)
    return min(buckets) if buckets else None
