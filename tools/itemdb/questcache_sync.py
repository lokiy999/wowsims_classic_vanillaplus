#!/usr/bin/env python3
"""Merge uploaded questcache.wdb files into the quest store and rebuild db.lokiy.dev when a quest is new or changed
(sim repo Part HK).

    python3 questcache_sync.py              # process ~/questcache/incoming/*.wdb
    python3 questcache_sync.py --no-build   # merge only
    python3 questcache_sync.py --pickup     # take uploads from lokiy.dev/admin/upload (run by sync_db.sh)

Uploads: Lokiy uploads WDB/questcache.wdb on lokiy.dev/admin/upload. The upload service (LootTracker tools/upload,
~/loot-upload) checks the "WQST" magic and keeps it in PICKUP as web_<file time>.wdb. --pickup moves it to incoming/,
runs the merge + rebuild and writes STATUS for that page. STATUS has the same shape as the item cache's:
{"at", "files", "code", "output": [last lines], "changes": [{"date", "id", "name", "new", "fields"}]}.

~/questcache/quests.json    {"updated", "quests": {id: {"quest": record (wdb_quests.py), "firstSeen", "changed",
                           "seen"}}}; gen_site.py reads it. "seen" = time of the newest cache file the quest came from:
                           an older file never overwrites newer data. The cache only holds quests the client has
                           queried, so every upload only adds quests (none is ever removed).
~/questcache/changes.jsonl one line per new or changed quest.
~/questcache/uploads/      the last processed file of each client.
"""
import argparse
import datetime
import fcntl
import glob
import json
import os
import re
import shutil
import subprocess
import sys

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from itemcache_sync import rebuild  # noqa: E402  (db.lokiy.dev, lokiy.dev/loot tooltips, sim.lokiy.dev)
from wdb_quests import parse  # noqa: E402

HOME = os.path.expanduser("~/questcache")
INCOMING = os.path.join(HOME, "incoming")
UPLOADS = os.path.join(HOME, "uploads")
STORE = os.path.join(HOME, "quests.json")
CHANGES = os.path.join(HOME, "changes.jsonl")
PICKUP = os.path.expanduser("~/loot-upload/data/questcache/incoming")
STATUS = os.path.expanduser("~/loot-upload/data/questcache/status.json")


def load_store():
    if not os.path.exists(STORE):
        return {"updated": None, "quests": {}}
    with open(STORE, encoding="utf-8") as fh:
        return json.load(fh)


def save_store(store):
    tmp = STORE + ".new"
    with open(tmp, "w", encoding="utf-8") as fh:
        json.dump(store, fh, ensure_ascii=False, separators=(",", ":"))
    if os.path.exists(STORE):
        shutil.copy2(STORE, STORE + ".bak")
    os.replace(tmp, STORE)


def merge(store, quests, mtime):
    """Add a parsed cache (file mtime) to the store; returns the change log lines."""
    when = datetime.datetime.fromtimestamp(mtime, datetime.timezone.utc).strftime("%Y-%m-%d %H:%M")
    log = []
    for qid, rec in quests.items():
        key = str(qid)
        old = store["quests"].get(key)
        if old and old.get("seen", 0) > mtime:
            continue
        if old and old["quest"] == rec:
            old["seen"] = mtime
            continue
        fields = {k: [old["quest"].get(k), v] for k, v in rec.items() if old["quest"].get(k) != v} if old else {}
        store["quests"][key] = {"quest": rec, "firstSeen": old["firstSeen"] if old else when, "changed": when,
                                "seen": mtime}
        log.append({"date": when, "id": qid, "name": rec.get("title", ""), "new": not old, "fields": fields})
    return log


def pickup():
    """Hand uploads from lokiy.dev/admin/upload to the normal run and write the result for the page."""
    if not os.path.isdir(PICKUP):
        return 0
    with open("/tmp/questcache-pickup.lock", "w") as lock:
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
        print(f"questcache pickup: {', '.join(files)} -> exit {res.returncode}, {len(changes)} new / changed quests")
        return res.returncode


def main():
    ap = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    ap.add_argument("--no-build", action="store_true", help="don't rebuild db.lokiy.dev")
    ap.add_argument("--pickup", action="store_true", help="take uploads from lokiy.dev/admin/upload")
    args = ap.parse_args()
    if args.pickup:
        return pickup()

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
    for mtime, client, path in sorted(files):
        try:
            quests = parse(path)
        except (OSError, ValueError, IndexError) as e:
            print(f"{path}: {e}; renamed to .bad")
            os.replace(path, path + ".bad")
            failed = True
            continue
        if not quests:
            print(f"{path}: no quests; renamed to .bad")
            os.replace(path, path + ".bad")
            failed = True
            continue
        changes = merge(store, quests, mtime)
        log += changes
        print(f"{client}: {len(quests)} quests in cache ({sum(1 for q in quests.values() if q.get('items') or q.get('choices'))}"
              f" with reward items), {sum(c['new'] for c in changes)} new, {sum(not c['new'] for c in changes)} changed")
        os.makedirs(UPLOADS, exist_ok=True)
        for old in glob.glob(os.path.join(UPLOADS, f"{glob.escape(client)}_*.wdb")):
            os.remove(old)
        shutil.move(path, os.path.join(UPLOADS, os.path.basename(path)))

    if log:
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
        print(f"Store: {len(store['quests'])} quests")
        if not args.no_build:
            rebuild()
    return 1 if failed else 0


if __name__ == "__main__":
    sys.exit(main())
