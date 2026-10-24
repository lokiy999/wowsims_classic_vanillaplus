#!/usr/bin/python3
"""Guild raid loot awards from WebDKP for db.lokiy.dev: which items were given out, when and where.

Inputs (in ~/webdkp/, not in git: they hold player names):
  - backup/*.csv          WebDKP addon backups (Lokiy's export, 2022-10 .. 2025-05); POINTHISTORY rows with
                          foritem = 1 are item awards: table, points, item, 1, zone, awarded by, date, player, ...
  - pages/<guild>/<n>.html  the public loot lists on webdkp.com (no login; 30 awards a page), fetched by --fetch
                          (one page a second, pages already saved are kept, except page 1 which holds the newest)
Output tools/itemdb/webdkp_awards.json (no player names), read by gen_site.py:
  {"updated": "...", "sources": {source: {"awards": n, "first": date, "last": date}},
   "awards": [[item_id, "YYYY-MM-DD", zone, source(, suffix)], ...],  # zone "" when the source doesn't say;
                                                                    # suffix ("of the Owl") for random-suffix items
   "unmatched": {name: n}}                                          # award names that are no item (or a typo)
An award is the item a raid leader gave a player for DKP, so it shows an item dropped there and roughly how often,
not every drop (items nobody wanted aren't in it). Names are matched to db.lokiy.dev's items (api/items.json):
"[!]" is WebDKP's comma, notes in brackets and "x3" counts are dropped, a random suffix ("Ragnaros Core of the
Eagle") is cut off, ALIASES fixes typos. Unmatched: AQ / ZG formulas the server has no item for, and non-items.
The same award in two inputs (same item, player and minute) is counted once.

Usage:
  python3 tools/itemdb/webdkp_awards.py [--fetch]
"""
import argparse
import csv
import glob
import json
import os
import re
import time
import urllib.request
from collections import Counter
from datetime import datetime, timezone

HERE = os.path.dirname(os.path.abspath(__file__))
DATA = os.path.expanduser("~/webdkp")
OUT = os.path.join(HERE, "webdkp_awards.json")
ITEMS = os.environ.get("VPLUS_ITEMS", "/var/www/db/api/items.json")
# public webdkp.com loot lists (Lokiy 2026-10-24): {source: URL of its Loot page}
GUILDS = {
    "VanillaPlus": "https://webdkp.com/dkp/wow/VanillaPlus/Loot",
    "VanillaPlus Horde": "https://webdkp.com/dkp/Hellscream/VanillaPlus+Horde/Loot",
}
CSV_SOURCE = "Cope (addon backup)"
ALIASES = {  # award name as typed -> item name
    "Scarbrite Axe": "Scabrite Axe",
    "Sclaed Sand Reaver Leggings": "Scaled Sand Reaver Leggings",
    "Eye of Divinity": "The Eye of Divinity",
    "Eye of Sulfaras": "Eye of Sulfuras",
    "Sawblade": "The Sawblade",
    "Sayaad WingsSayaad Wings": "Sayaad Wings",
    "Manual of Battle Shout Book VII": "Manual of Battle Shout VII",
    "Tablet: Grace of Air": "Tablet of Grace of Air Totem III",
    "lava cores2x": "Lava Core",
}
NOTE_RE = re.compile(r"\s*\(.*\)$")  # "Puissant Cape (Feb 9)", "Nexus Crystalx3 (for Franshals)"
ROW_RE = re.compile(r"table\.Add\((\{.*?\})\);")
PAGES_RE = re.compile(r"SetPageData\((\d+),\s*(\d+)\)")
QTY_RE = re.compile(r"\s*(?:x\s*\d+|\d+\s*x)$", re.I)


def get(url):
    req = urllib.request.Request(url, headers={"User-Agent": "db.lokiy.dev WebDKP import (lokiy.dev)"})
    with urllib.request.urlopen(req, timeout=30) as r:
        return r.read().decode("utf-8", "replace")


def fetch(source, url):
    """Saves every page of a loot list that isn't saved yet (page 1 always, it has the newest awards)."""
    folder = os.path.join(DATA, "pages", source.replace(" ", "_"))
    os.makedirs(folder, exist_ok=True)
    page, last = 1, 1
    while page <= last:
        path = os.path.join(folder, f"{page}.html")
        if page == 1 or not os.path.exists(path):
            text = get(url if page == 1 else f"{url}/{page}")
            with open(path, "w", encoding="utf-8") as f:
                f.write(text)
            time.sleep(1.0)
        else:
            text = open(path, encoding="utf-8").read()
        m = PAGES_RE.search(text)
        last = int(m.group(2)) if m else page
        page += 1
    print(f"{source}: {last} pages")


def read_pages(source):
    """[(name, player, unix time)] from the saved pages of one loot list."""
    rows = []
    for path in glob.glob(os.path.join(DATA, "pages", source.replace(" ", "_"), "*.html")):
        for m in ROW_RE.finditer(open(path, encoding="utf-8").read()):
            r = json.loads(m.group(1))
            rows.append((r["name"], r["player"], int(r["date"]), ""))
    return rows


def read_csv(path):
    """[(name, player, unix time, zone)] of the item awards in a WebDKP backup (times are the raid leader's clock,
    read as UTC; only the date is kept)."""
    rows, section = [], ""
    with open(path, encoding="utf-8", errors="replace", newline="") as f:
        for p in csv.reader(f):
            if len(p) == 1 and p[0].isupper():
                section = p[0]
            elif section == "POINTHISTORY" and len(p) >= 8 and p[3].strip() == "1":
                t = datetime.strptime(p[6].strip(), "%Y-%m-%d %H:%M:%S").replace(tzinfo=timezone.utc)
                zone = p[4].strip()
                rows.append((p[2].strip(), p[7].strip(), int(t.timestamp()), "" if zone == "WebDKP" else zone))
    return rows


def name_index():
    items = json.load(open(ITEMS, encoding="utf-8"))
    by_name = {}
    for iid in sorted(items, key=int):  # the lowest id wins when two items share a name
        by_name.setdefault(items[iid]["name"].lower(), int(iid))
    return by_name


def match(name, by_name):
    """(item id, random suffix or "") for an award name, or (None, "")."""
    name = NOTE_RE.sub("", name.strip())
    name = ALIASES.get(name, name).replace("[!]", ",").strip()
    for cand in (name, QTY_RE.sub("", name)):
        if cand.lower() in by_name:
            return by_name[cand.lower()], ""
    # random suffix: "<item> of the Eagle" / "<item> of Spirit"
    m = re.match(r"(.+?) of (?:the )?\w+(?: \w+)?$", name)
    while m:
        if m.group(1).lower() in by_name:
            return by_name[m.group(1).lower()], name[len(m.group(1)) + 1:]
        m = re.match(r"(.+?) of (?:the )?\w+(?: \w+)?$", m.group(1))
    return None, ""


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--fetch", action="store_true", help="fetch the webdkp.com loot lists first")
    args = ap.parse_args()
    if args.fetch:
        for source, url in GUILDS.items():
            fetch(source, url)
    inputs = [(source, read_pages(source)) for source in GUILDS]
    inputs += [(CSV_SOURCE, rows) for path in sorted(glob.glob(os.path.join(DATA, "backup", "*.csv")))
               for rows in [read_csv(path)]]
    by_name = name_index()
    seen, awards, unmatched, stats = set(), [], Counter(), {}
    for source, rows in inputs:
        for name, player, t, zone in rows:
            key = (name.lower(), player.lower(), t // 60)
            if key in seen:
                continue
            seen.add(key)
            iid, suffix = match(name, by_name)
            if iid is None:
                unmatched[name] += 1
                continue
            day = datetime.fromtimestamp(t, timezone.utc).date().isoformat()
            awards.append([iid, day, zone, source] + ([suffix] if suffix else []))
            s = stats.setdefault(source, {"awards": 0, "first": day, "last": day})
            s["awards"] += 1
            s["first"], s["last"] = min(s["first"], day), max(s["last"], day)
    awards.sort(key=lambda a: (a[1], a[0]))
    head = {"updated": datetime.now(timezone.utc).date().isoformat(), "sources": stats,
            "unmatched": dict(unmatched.most_common())}
    with open(OUT, "w", encoding="utf-8") as f:  # one award a line, so a new backup shows as a short git diff
        f.write(json.dumps(head, ensure_ascii=False)[:-1] + ',\n"awards": [\n')
        f.write(",\n".join(json.dumps(a, ensure_ascii=False) for a in awards))
        f.write("\n]}\n")
    total = len(awards) + sum(unmatched.values())
    print(f"{len(awards)} of {total} awards matched to items ({len(unmatched)} names unmatched) -> {OUT}")
    for source, s in stats.items():
        print(f"  {source}: {s['awards']} ({s['first']} .. {s['last']})")


if __name__ == "__main__":
    main()
