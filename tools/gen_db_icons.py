#!/usr/bin/python3
"""assets/database/db_icons.json = sorted list of the icon names db.lokiy.dev has (/icons/<name>.jpg 56 px and
<name>_s.jpg 18 px: every icon of the game client, the server-made ones and the ones tools/import_icons.py added).
The sim UI (ui/core/db_icons.ts) builds the list in and shows every listed icon from db.lokiy.dev.
Run by tools/itemdb/sync_sim.sh with the item pipeline, after import_icons.py. ICONS_DIR overrides the folder.

Usage (from the repo root):
  python3 tools/gen_db_icons.py
"""
import json
import os

REPO = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
ICONS_DIR = os.environ.get("ICONS_DIR", "/var/www/db/icons")
OUT = os.path.join(REPO, "assets/database/db_icons.json")


def main():
    names = {f[:-4].lower() for f in os.listdir(ICONS_DIR) if f.endswith(".jpg")}
    # <name>_s.jpg is the small copy of <name>.jpg, not an icon of its own
    names = sorted(n for n in names if not (n.endswith("_s") and n[:-2] in names))
    if len(names) < 1000:  # sanity floor: a missing / half-copied folder must not switch every icon back to Wowhead
        raise SystemExit(f"only {len(names)} icons in {ICONS_DIR}, {OUT} left as it is")
    with open(OUT, "w", encoding="utf-8") as f:
        json.dump(names, f, separators=(",", ":"))
        f.write("\n")
    print(f"wrote {OUT}: {len(names)} icons")


if __name__ == "__main__":
    main()
