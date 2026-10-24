#!/usr/bin/env python3
"""Merge uploaded itemcache.wdb files into the item store and rebuild db.lokiy.dev when an item is new or changed.

    python3 itemcache_sync.py              # process ~/itemcache/incoming/*.wdb
    python3 itemcache_sync.py --no-build   # merge only
    python3 itemcache_sync.py --missing-icons  # icon names db.lokiy.dev uses that neither the CDN nor it has
    python3 itemcache_sync.py --rebuild    # rebuild db.lokiy.dev and lokiy.dev/loot (after uploading icons)
    python3 itemcache_sync.py --pickup     # take uploads from lokiy.dev/admin/upload (run by sync_db.sh)

Uploads: Lokiy uploads itemcache.wdb on lokiy.dev/admin/upload; the upload service (~/loot-upload, sandboxed to its
data folder) keeps it in PICKUP. --pickup moves it to incoming/, runs the merge + rebuild and writes STATUS (time,
files, exit code, output, the new / changed items) for that page.

Lokiy's PC uploads each client's WDB/itemcache.wdb to ~/itemcache/incoming/<client>_<file mtime>.wdb
(ItemCacheSync/upload_itemcache.py). Files are merged oldest first, so the newest copy of an item wins.

~/itemcache/items.json    {"updated", "items": {id: {"item": record, "firstSeen", "changed", "seen"}}}; gen_site.py reads
                          it. "seen" = mtime of the newest cache file the item came from: an older file (the other
                          client, not played for a while) never overwrites newer data.
~/itemcache/changes.jsonl one line per new or changed item: {"date", "id", "name", "new", "fields": {k: [old, new]}}
~/itemcache/uploads/      the last processed file of each client
"""

import argparse
import fcntl
import datetime
import glob
import json
import os
import re
import shutil
import subprocess
import sys
import urllib.error
import urllib.request
from concurrent.futures import ThreadPoolExecutor

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from wdb_items import read_wdb  # noqa: E402

HOME = os.path.expanduser("~/itemcache")
INCOMING = os.path.join(HOME, "incoming")
UPLOADS = os.path.join(HOME, "uploads")
STORE = os.path.join(HOME, "items.json")
CHANGES = os.path.join(HOME, "changes.jsonl")
GEN_SITE = os.path.join(os.path.dirname(os.path.abspath(__file__)), "gen_site.py")
ENRICH = os.path.expanduser("~/loot-sync/enrich.py")  # lokiy.dev/loot item-info.json (tooltips from this store)
SYNC_SIM = os.path.join(os.path.dirname(os.path.abspath(__file__)), "sync_sim.sh")  # sim.lokiy.dev
LOCK = "/tmp/db-sync.lock"  # shared with sync_db.sh, so two rebuilds never run at once
DB_API = "/var/www/db/api/items.json"
PICKUP = os.path.expanduser("~/loot-upload/data/itemcache/incoming")
STATUS = os.path.expanduser("~/loot-upload/data/itemcache/status.json")
SYNC_SIM = os.path.join(os.path.dirname(os.path.abspath(__file__)), "sync_sim.sh")
CDN_PREFIX = "https://wow.zamimg.com/images/wow/icons/medium/"
CDN_SEEN = os.path.join(HOME, "cdn_icons.json")  # {icon name: True if the CDN has it}; each name is checked once


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


def rebuild():
    print("Rebuilding db.lokiy.dev")
    subprocess.run(["flock", LOCK, sys.executable, GEN_SITE], check=True)
    if os.path.exists(ENRICH):
        print("Refreshing lokiy.dev/loot tooltips")
        subprocess.run([sys.executable, ENRICH], check=True)
    if os.path.exists(SYNC_SIM):  # sim.lokiy.dev: rebuilt in the background when its item data changed
        print("sim.lokiy.dev sync started")
        subprocess.Popen([SYNC_SIM], stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL, start_new_session=True)


def pickup():
    """Hand uploads from lokiy.dev/admin/upload to the normal run and write the result for the page."""
    if not os.path.isdir(PICKUP):
        return 0
    with open("/tmp/itemcache-pickup.lock", "w") as lock:
        try:
            fcntl.flock(lock, fcntl.LOCK_EX | fcntl.LOCK_NB)
        except OSError:
            return 0  # another pickup is running
        files = sorted(n for n in os.listdir(PICKUP) if n.endswith(".wdb"))
        if not files:
            return 0
        os.makedirs(INCOMING, exist_ok=True)
        for n in files:
            shutil.move(os.path.join(PICKUP, n), os.path.join(INCOMING, n))
        before = sum(1 for _ in open(CHANGES, encoding="utf-8")) if os.path.exists(CHANGES) else 0
        res = subprocess.run([sys.executable, os.path.abspath(__file__)], capture_output=True, text=True)
        changes = []
        if os.path.exists(CHANGES):
            with open(CHANGES, encoding="utf-8") as fh:
                changes = [json.loads(ln) for i, ln in enumerate(fh) if i >= before and ln.strip()]
        status = {"at": int(datetime.datetime.now().timestamp()), "files": files, "code": res.returncode,
                  "output": (res.stdout + res.stderr).strip().splitlines()[-60:], "changes": changes}
        tmp = STATUS + ".tmp"
        with open(tmp, "w", encoding="utf-8") as fh:
            json.dump(status, fh, ensure_ascii=False)
        os.replace(tmp, STATUS)
        print(f"pickup: {', '.join(files)} -> exit {res.returncode}, {len(changes)} new / changed items")
        return res.returncode
    # sim.lokiy.dev: rebuilt in the background when the item or source data changed
    subprocess.Popen([SYNC_SIM], start_new_session=True, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)


def on_cdn(name):
    req = urllib.request.Request(CDN_PREFIX + name + ".jpg", method="HEAD")
    try:
        with urllib.request.urlopen(req, timeout=20) as r:
            return r.status == 200
    except urllib.error.HTTPError as e:
        return False if e.code == 404 else None
    except OSError:
        return None  # network trouble: ask again next time


def missing_icons():
    """Icon names items on db.lokiy.dev use that the public CDN doesn't have and db.lokiy.dev doesn't host yet."""
    with open(DB_API, encoding="utf-8") as fh:
        names = {v["icon"][len(CDN_PREFIX):-4] for v in json.load(fh).values()
                 if v.get("icon", "").startswith(CDN_PREFIX)}
    try:
        with open(CDN_SEEN, encoding="utf-8") as fh:
            seen = json.load(fh)
    except (OSError, ValueError):
        seen = {}
    todo = sorted(n for n in names if n not in seen)
    with ThreadPoolExecutor(16) as pool:
        for name, ok in zip(todo, pool.map(on_cdn, todo)):
            if ok is not None:
                seen[name] = ok
    with open(CDN_SEEN, "w", encoding="utf-8") as fh:
        json.dump(seen, fh, sort_keys=True)
    return sorted(n for n in names if seen.get(n) is False)


def main():
    ap = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    ap.add_argument("--no-build", action="store_true", help="don't rebuild db.lokiy.dev")
    ap.add_argument("--missing-icons", action="store_true", help="print icon names the site can't show, one per line")
    ap.add_argument("--rebuild", action="store_true", help="only rebuild db.lokiy.dev and lokiy.dev/loot")
    ap.add_argument("--pickup", action="store_true", help="take uploads from lokiy.dev/admin/upload")
    args = ap.parse_args()
    if args.pickup:
        return pickup()
    if args.missing_icons:
        print("\n".join(missing_icons()))
        return 0
    if args.rebuild:
        rebuild()
        return 0

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
            rebuild()
    else:
        print("No item changed")
    return 1 if failed else 0


if __name__ == "__main__":
    sys.exit(main())
