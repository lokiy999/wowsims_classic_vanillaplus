#!/usr/bin/python3
"""Compare the freshly built sim item DB (assets/database/db.json) with the committed one and say whether it is safe
to deploy (sync_sim.sh). Prints a summary for the log; exit 0 = deploy, 1 = changes too big, 2 = no item changes.

Too big: the item count drops by more than 1% or 30 items, or more than 200 items lose their stats or their sources.

Usage (from the repo root):
  python3 tools/itemdb/sim_check.py
"""
import json
import subprocess
import sys

MAX_DROP_PCT, MAX_DROP_ITEMS, MAX_LOST = 1.0, 30, 200


def main():
    old = {i["id"]: i for i in json.loads(subprocess.check_output(
        ["git", "show", "HEAD:assets/database/db.json"]))["items"]}
    with open("assets/database/db.json", encoding="utf-8") as f:
        new = {i["id"]: i for i in json.load(f)["items"]}

    added = sorted(set(new) - set(old))
    removed = sorted(set(old) - set(new))
    both = set(old) & set(new)
    changed = sorted(i for i in both if old[i] != new[i])
    lost_stats = [i for i in both if any(old[i].get("stats") or []) and not any(new[i].get("stats") or [])]
    lost_sources = [i for i in both if old[i].get("sources") and not new[i].get("sources")]
    phase = [i for i in changed if old[i].get("phase") != new[i].get("phase")]
    sources = [i for i in changed if old[i].get("sources") != new[i].get("sources")]

    def names(ids, n=10):
        txt = ", ".join(f"{i} {(new.get(i) or old[i])['name']}" for i in ids[:n])
        return txt + (f" (+{len(ids) - n} more)" if len(ids) > n else "")

    print(f"items {len(old)} -> {len(new)}: +{len(added)} -{len(removed)}, {len(changed)} changed "
          f"({len(phase)} phase, {len(sources)} sources)")
    for label, ids in (("added", added), ("removed", removed), ("phase", phase), ("sources", sources)):
        if ids:
            print(f"  {label}: {names(ids)}")

    drop = len(old) - len(new)
    problems = []
    if drop > MAX_DROP_ITEMS or drop > len(old) * MAX_DROP_PCT / 100:
        problems.append(f"item count drops by {drop}")
    if len(lost_stats) > MAX_LOST:
        problems.append(f"{len(lost_stats)} items lose their stats")
    if len(lost_sources) > MAX_LOST:
        problems.append(f"{len(lost_sources)} items lose their sources")
    if problems:
        print("NOT deployed: " + "; ".join(problems))
        return 1
    return 0 if (added or removed or changed) else 2


if __name__ == "__main__":
    sys.exit(main())
