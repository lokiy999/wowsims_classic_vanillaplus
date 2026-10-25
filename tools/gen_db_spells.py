#!/usr/bin/python3
"""Spell names, icons and tooltips for the sim UI from db.lokiy.dev's api/spells.json, the only spell data source:

  assets/database/db_spell_ids.json   sorted ids of every spell db.lokiy.dev has (ui/core/db_spells.ts loads it with
                                      the database: those link to db.lokiy.dev/spell/<id>)
  assets/spells/<id // 1000>.json     {id: {name, icon, rank, hasBuff, tooltip}} per block of 1000 ids, loaded when a
                                      spell of the block is shown (same origin, so no CORS on db.lokiy.dev needed)

Spells db.lokiy.dev doesn't have (Season of Discovery / later ones the sim DB names) keep Wowhead's data. Run by
tools/itemdb/sync_sim.sh. VPLUS_SPELLS overrides the input path.

Usage (from the repo root):
  python3 tools/gen_db_spells.py
"""
import glob
import html
import json
import os
import re

REPO = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
SPELLS_JSON = os.environ.get("VPLUS_SPELLS", "/var/www/db/api/spells.json")
IDS_OUT = os.path.join(REPO, "assets/database/db_spell_ids.json")
CHUNK_DIR = os.path.join(REPO, "assets/spells")
RANK_RE = re.compile(r"Rank (\d+)")


def text(s):
    return html.escape(s.strip(), quote=False).replace("\n", "<br />")


def tooltip(s):
    """Same layout as the sim's other local spell tooltips: name, rank, description, buff text."""
    parts = [f'<b class="whtt-name">{text(s["name"])}</b>']
    if s.get("rank"):
        parts.append(text(s["rank"]))
    if s.get("desc"):
        parts.append(text(s["desc"]))
    out = "<br />".join(parts)
    if s.get("tip"):
        out += f'<br /><br /><span class="whtt-buff">Buff: {text(s["tip"])}</span>'
    return out


def main():
    with open(SPELLS_JSON, encoding="utf-8") as f:
        spells = {int(k): v for k, v in json.load(f).items()}
    if len(spells) < 10000:  # sanity floor: a broken export must not switch every spell back to Wowhead
        raise SystemExit(f"only {len(spells)} spells in {SPELLS_JSON}, nothing written")
    chunks = {}
    for sid, s in sorted(spells.items()):
        rank = RANK_RE.search(s.get("rank") or "")
        chunks.setdefault(sid // 1000, {})[str(sid)] = {
            "name": s["name"],
            "icon": os.path.basename(s.get("icon") or "")[:-4],
            "rank": int(rank.group(1)) if rank else 0,
            "hasBuff": bool(s.get("tip")),
            "tooltip": tooltip(s),
        }
    os.makedirs(CHUNK_DIR, exist_ok=True)
    for old in glob.glob(os.path.join(CHUNK_DIR, "*.json")):
        if int(os.path.basename(old)[:-5]) not in chunks:
            os.remove(old)
    for n, chunk in chunks.items():
        with open(os.path.join(CHUNK_DIR, f"{n}.json"), "w", encoding="utf-8") as f:
            json.dump(chunk, f, ensure_ascii=False, separators=(",", ":"))
            f.write("\n")
    with open(IDS_OUT, "w", encoding="utf-8") as f:
        json.dump(sorted(spells), f, separators=(",", ":"))
        f.write("\n")
    print(f"wrote {IDS_OUT} and {len(chunks)} files in {CHUNK_DIR}: {len(spells)} spells")


if __name__ == "__main__":
    main()
