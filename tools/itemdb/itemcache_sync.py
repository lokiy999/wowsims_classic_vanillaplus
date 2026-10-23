#!/usr/bin/env python3
"""Merge uploaded itemcache.wdb files into the item store and rebuild db.lokiy.dev when an item is new or changed.

    python3 itemcache_sync.py              # process ~/itemcache/incoming/*.wdb
    python3 itemcache_sync.py --no-build   # merge only

Lokiy's PC uploads each client's WDB/itemcache.wdb to ~/itemcache/incoming/<client>_<file mtime>.wdb
(ItemCacheSync/upload_itemcache.py). Files are merged oldest first, so the newest copy of an item wins.

~/itemcache/items.json    {"updated", "items": {id: {"item": record, "firstSeen", "changed", "seen"}}}; gen_site.py reads
                          it. "seen" = mtime of the newest cache file the item came from: an older file (the other
                          client, not played for a while) never overwrites newer data.
~/itemcache/changes.jsonl one line per new or changed item: {"date", "id", "name", "new", "fields": {k: [old, new]}}
~/itemcache/uploads/      the last processed file of each client
"""

import argparse
import datetime
import glob
import json
import os
import re
import shutil
import subprocess
import sys

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from wdb_items import read_wdb  # noqa: E402

HOME = os.path.expanduser("~/itemcache")
INCOMING = os.path.join(HOME, "incoming")
UPLOADS = os.path.join(HOME, "uploads")
STORE = os.path.join(HOME, "items.json")
CHANGES = os.path.join(HOME, "changes.jsonl")
GEN_SITE = os.path.join(os.path.dirname(os.path.abspath(__file__)), "gen_site.py")
ENRICH = os.path.expanduser("~/loot-sync/enrich.py")  # lokiy.dev/loot item-info.json (tooltips from this store)
LOCK = "/tmp/db-sync.lock"  # shared with sync_db.sh, so two rebuilds never run at once


def load_store():
    if not os.path.exists(STORE):
        return {"updated": None, "items": {}}
    with open(STORE, encoding="utf-8") as fh:
        return json.load(fh)


def save_store(store):
    tmp = STORE + ".new"
    with open(tmp, "w", encoding="utf-8") as fh:
        json.dump(store, fh, ensure_ascii=False, separators=(",", ":"))
    if os.path.exists(STORE):
        shutil.copy2(STORE, STORE + ".bak")
    os.replace(tmp, STORE)


def merge(store, items, mtime):
    """Add a parsed cache (file mtime) to the store; returns the change log lines."""
    when = datetime.datetime.fromtimestamp(mtime, datetime.timezone.utc).strftime("%Y-%m-%d %H:%M")
    log = []
    for iid, rec in items.items():
        key = str(iid)
        old = store["items"].get(key)
        if old and old.get("seen", 0) > mtime:
            continue
        if old and old["item"] == rec:
            old["seen"] = mtime
            continue
        fields = {k: [old["item"].get(k), v] for k, v in rec.items() if old["item"].get(k) != v} if old else {}
        store["items"][key] = {"item": rec, "firstSeen": old["firstSeen"] if old else when, "changed": when,
                               "seen": mtime}
        log.append({"date": when, "id": iid, "name": rec["name"], "new": not old, "fields": fields})
    return log


def main():
    ap = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    ap.add_argument("--no-build", action="store_true", help="don't rebuild db.lokiy.dev")
    args = ap.parse_args()

    files = []
    for path in glob.glob(os.path.join(INCOMING, "*.wdb")):
        m = re.match(r"^(.+)_(\d+)\.wdb$", os.path.basename(path))
        if m:
            files.append((int(m.group(2)), m.group(1), path))
    if not files:
        print("Nothing in", INCOMING)
        return 0

    store = load_store()
    log, failed = [], False
    baseline, store_changed = not store["items"], False
    for mtime, client, path in sorted(files):
        try:
            res = read_wdb(path)
        except (OSError, ValueError) as e:
            print(f"{path}: {e}; renamed to .bad")
            os.replace(path, path + ".bad")
            failed = True
            continue
        n = len(res["items"])
        if res["bad"]:
            # a file WoW was still writing ends in a cut-off record; the records before it are fine
            print(f"{path}: {len(res['bad'])} unreadable records (e.g. {res['bad'][:3]}) left out")
        if not n:
            print(f"{path}: no items; renamed to .bad")
            os.replace(path, path + ".bad")
            failed = True
            continue
        changes = merge(store, res["items"], mtime)
        if not baseline:
            log += changes
        elif changes:
            print(f"First import: {len(changes)} items stored as the baseline, not logged as new")
            baseline = False
            store_changed = True
        print(f"{client}: {n} items in cache, {sum(c['new'] for c in changes)} new, "
              f"{sum(not c['new'] for c in changes)} changed")
        os.makedirs(UPLOADS, exist_ok=True)
        for old in glob.glob(os.path.join(UPLOADS, f"{glob.escape(client)}_*.wdb")):
            os.remove(old)
        shutil.move(path, os.path.join(UPLOADS, os.path.basename(path)))

    if log or store_changed:
        store["updated"] = datetime.datetime.now(datetime.timezone.utc).strftime("%Y-%m-%d %H:%M")
        save_store(store)
        with open(CHANGES, "a", encoding="utf-8") as fh:
            for c in log:
                fh.write(json.dumps(c, ensure_ascii=False) + "\n")
        for c in log[:20]:
            print(("  new " if c["new"] else "  changed ") + f"{c['id']} {c['name']}"
                  + ("" if c["new"] else ": " + ", ".join(c["fields"])))
        if len(log) > 20:
            print(f"  ... {len(log) - 20} more in {CHANGES}")
        print(f"Store: {len(store['items'])} items")
        if not args.no_build:
            print("Rebuilding db.lokiy.dev")
            subprocess.run(["flock", LOCK, sys.executable, GEN_SITE], check=True)
            if os.path.exists(ENRICH):
                print("Refreshing lokiy.dev/loot tooltips")
                subprocess.run([sys.executable, ENRICH], check=True)
    else:
        print("No item changed")
    return 1 if failed else 0


if __name__ == "__main__":
    sys.exit(main())
