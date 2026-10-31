#!/usr/bin/python3
"""Every icon the sim shows comes from db.lokiy.dev. This adds the ones its icon set (the game client's, see
ItemCacheSync extract_icons.py / extract_ui_images.py) doesn't have: icon names of later expansions and Season of
Discovery that the sim's DB (assets/database/*.json) or UI code (ui/**) uses. Each is copied once from Wowhead's
icon CDN (56 px -> <name>.jpg, 18 px -> <name>_s.jpg) into CSV's/icons/ (gen_site.py copies that folder on every
rebuild) and /var/www/db/icons/ (live right away). Names the CDN doesn't have are remembered in
CSV's/icons/missing.json and not asked for again. Run by tools/itemdb/sync_sim.sh before gen_db_icons.py.

Usage (from the repo root):
  python3 tools/import_icons.py [--dry-run]
"""
import glob
import json
import os
import re
import sys
import urllib.error
import urllib.request

REPO = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
ICON_DIRS = [os.path.join(REPO, "tools/itemdb/icons"), os.path.join(REPO, "CSV's/icons")]
STORE = ICON_DIRS[1]
LIVE = os.environ.get("ICONS_DIR", "/var/www/db/icons")
MISSING = os.path.join(STORE, "missing.json")
CDN = "https://wow.zamimg.com/images/wow/icons/{size}/{name}.jpg"
SIZES = (("large", ""), ("small", "_s"))
# an icon name in UI code: 'name.jpg' or .../icons/<size>/name.jpg
CODE_ICON = re.compile(r"[\"'/]((?:inv|spell|ability|class|achievement|trade|racial|temp)_[a-z0-9_]+)\.jpg[\"']", re.I)


def have():
    return {f[:-4].lower() for d in ICON_DIRS if os.path.isdir(d) for f in os.listdir(d) if f.endswith(".jpg")}


def wanted():
    names = set()
    for path in glob.glob(os.path.join(REPO, "assets/database/*db.json")):
        with open(path, encoding="utf-8") as f:
            db = json.load(f)
        for key in ("items", "itemIcons", "spellIcons"):
            names |= {x["icon"].lower() for x in db.get(key, []) if x.get("icon")}
    for path in glob.glob(os.path.join(REPO, "ui/**/*.ts*"), recursive=True):
        with open(path, encoding="utf-8") as f:
            names |= {n.lower() for n in CODE_ICON.findall(f.read())}
    return names


def fetch(url):
    req = urllib.request.Request(url, headers={"User-Agent": "lokiy.dev icon import"})
    with urllib.request.urlopen(req, timeout=20) as r:
        return r.read()


def main():
    dry = "--dry-run" in sys.argv
    missing = set()
    if os.path.exists(MISSING):
        with open(MISSING, encoding="utf-8") as f:
            missing = set(json.load(f))
    todo = sorted(wanted() - have() - missing)
    print(f"{len(todo)} icons to import{' (dry run)' if dry else ''}: {' '.join(todo)}")
    if dry:
        return
    added = 0
    for name in todo:
        try:
            files = {f"{name}{suffix}.jpg": fetch(CDN.format(size=size, name=name)) for size, suffix in SIZES}
        except urllib.error.HTTPError as e:
            print(f"  {name}: {e.code}, not on the CDN")
            missing.add(name)
            continue
        for d in (STORE, LIVE):
            if os.path.isdir(d):
                for fn, data in files.items():
                    with open(os.path.join(d, fn), "wb") as f:
                        f.write(data)
        added += 1
    with open(MISSING, "w", encoding="utf-8") as f:
        json.dump(sorted(missing), f, indent=0)
        f.write("\n")
    print(f"imported {added}, {len(missing)} names known missing")


if __name__ == "__main__":
    main()
