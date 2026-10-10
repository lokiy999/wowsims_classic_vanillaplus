"""Fetch random affix roll chances from Wowhead classic item pages -> tools/itemdb/affix_chances.json.

{item_id: {affix name: chance %}}, e.g. {"9640": {"of the Bear": 41.0, "of Power": 13.2, ...}}. Only items with
random affixes (Wowhead gear planner data) are fetched; items already in the file are skipped, so it can be
re-run to resume. One request every DELAY seconds.

Usage: python3 tools/itemdb/fetch_affix_chances.py [--limit N]
"""
import argparse
import json
import os
import re
import sys
import time
import urllib.request

HERE = os.path.dirname(os.path.abspath(__file__))
sys.path.insert(0, HERE)
from loot_sync import load_random_affixes  # noqa: E402

OUT = os.path.join(HERE, "affix_chances.json")
DELAY = 1.0
LI_RE = re.compile(r'<span class="q\d">\.\.\.(.+?)</span>\s*<small class="q0">\(([\d.]+)% chance\)</small>', re.S)


def fetch(iid):
    req = urllib.request.Request(f"https://www.wowhead.com/classic/item={iid}",
                                 headers={"User-Agent": "Mozilla/5.0 (db.lokiy.dev affix chances)"})
    with urllib.request.urlopen(req, timeout=30) as r:
        page = r.read().decode("utf-8", "replace")
    i = page.find('class="random-enchantments"')
    if i < 0:
        return {}
    block = page[i:page.find("</ul>", i)]
    return {name.strip(): float(pct) for name, pct in LI_RE.findall(block)}


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--limit", type=int, default=0)
    args = ap.parse_args()
    affixes = load_random_affixes()[0]
    done = json.load(open(OUT)) if os.path.exists(OUT) else {}
    todo = [i for i in sorted(affixes) if str(i) not in done]
    if args.limit:
        todo = todo[:args.limit]
    print(f"{len(done)} cached, {len(todo)} to fetch", flush=True)
    for n, iid in enumerate(todo, 1):
        try:
            done[str(iid)] = fetch(iid)
        except OSError as e:
            print(f"{iid}: {e}", flush=True)
            time.sleep(10)
            continue
        if n % 25 == 0 or n == len(todo):
            json.dump(done, open(OUT + ".tmp", "w"), separators=(",", ":"), sort_keys=True)
            os.replace(OUT + ".tmp", OUT)
            print(f"{n}/{len(todo)}", flush=True)
        time.sleep(DELAY)
    json.dump(done, open(OUT + ".tmp", "w"), separators=(",", ":"), sort_keys=True)
    os.replace(OUT + ".tmp", OUT)


if __name__ == "__main__":
    main()
