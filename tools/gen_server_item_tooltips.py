#!/usr/bin/python
"""Local (server-data) tooltips for the non-gear items the sim and UI show: consumables (potions, flasks,
elixirs, food, mana gems, explosives), enchant items and other item icons.

Gear items already carry the in-game tooltip from `CSV's/VPlusItemDB.lua` (docs/parse_vplus.py). Everything
else fell back to Wowhead's live tooltip. This collects the item ids referenced by the UI (TS), the sim (Go)
and the DB item icons, takes each one's tooltip lines from the dump and writes
assets/db_inputs/server_item_tooltips.json ({id: {name, icon, tooltip}}); gen_db uses it for those item
icons. Name and icon only matter for server items Wowhead does not know (icon from the AtlasLoot tables).

Usage:
  python tools/gen_server_item_tooltips.py [--dry-run]
Then regenerate the DB:
  go run ./tools/database/gen_db -outDir=./assets -gen=db
"""
import argparse
import glob
import json
import os
import re
import sys

REPO = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
sys.path.insert(0, os.path.join(REPO, "docs"))
from parse_vplus import atlasloot_icons, build_tooltip, parse_item, parse_lua  # noqa: E402

GO_LINE_RE = re.compile(r"item_?id", re.I)
INT_RE = re.compile(r"\b(\d{3,6})\b")
TS_RES = [re.compile(r"fromItemId\(\s*(\d+)"), re.compile(r"itemId:\s*(\d+)")]
# Server items that neither Wowhead nor the AtlasLoot tables have an icon for (placeholder, see docs/TODO.md).
ICON_OVERRIDES = {
    34323: "inv_potion_24",  # Flask of Indomitable Might
}
EXTRA_RE = re.compile(r"var ExtraItemIcons = \[\]int32\{(.*?)\n\}", re.S)


def collect_ids(repo):
    ids = set()
    for p in glob.glob(os.path.join(repo, "sim/**/*.go"), recursive=True):
        if p.endswith("_test.go"):
            continue
        for line in open(p, encoding="utf-8"):
            if GO_LINE_RE.search(line):
                ids.update(int(x) for x in INT_RE.findall(line.split("//")[0]))
    for p in glob.glob(os.path.join(repo, "ui/**/*.ts*"), recursive=True):
        if "node_modules" in p:
            continue
        s = open(p, encoding="utf-8").read()
        for r in TS_RES:
            ids.update(int(x) for x in r.findall(s))
    m = EXTRA_RE.search(open(os.path.join(repo, "tools/database/overrides.go"), encoding="utf-8").read())
    if m:
        ids.update(int(x) for x in re.findall(r"^\s*(\d+),", m.group(1), re.M))
    db = json.load(open(os.path.join(repo, "assets/database/db.json"), encoding="utf-8"))
    ids.update(i["id"] for i in db.get("itemIcons", []))
    ids.update(e["itemId"] for e in db.get("enchants", []) if e.get("itemId"))
    gear = {i["id"] for i in db.get("items", [])}
    return ids - gear


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--repo", default=REPO)
    ap.add_argument("--dry-run", action="store_true")
    args = ap.parse_args()

    dump = parse_lua(os.path.join(args.repo, "CSV's/VPlusItemDB.lua"))
    cwd = os.getcwd()
    os.chdir(args.repo)  # atlasloot_icons reads repo-relative paths
    icons = atlasloot_icons()
    os.chdir(cwd)
    out = {}
    for iid in sorted(collect_ids(args.repo)):
        full = dump.get(iid) or []
        lines = [ln for ln in full if not ln.startswith("Rarity:")]
        if len(lines) < 2:
            continue
        out[str(iid)] = {
            "name": lines[0],
            "icon": ICON_OVERRIDES.get(iid) or icons.get(iid, ""),
            "tooltip": build_tooltip(lines, parse_item(full).get("quality")),
        }

    print(f"{len(out)} server item tooltips")
    if args.dry_run:
        for k in list(out)[:8]:
            print(k, json.dumps(out[k])[:300])
        return
    path = os.path.join(args.repo, "assets/db_inputs/server_item_tooltips.json")
    with open(path, "w", encoding="utf-8") as f:
        json.dump(out, f, indent=1, sort_keys=True)
        f.write("\n")
    print(f"Wrote {path}")


if __name__ == "__main__":
    main()
