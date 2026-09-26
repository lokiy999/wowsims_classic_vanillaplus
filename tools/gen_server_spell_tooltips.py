#!/usr/bin/python
"""Local (server-data) tooltips for every spell the sim and UI use, not just talents.

Spells outside the talent trees (Fireball, Innervate, procs, consumables, ...) used to fall back to
Wowhead's live tooltip, which shows Classic/SoD values instead of the Vanilla Plus ones. This collects
the spell ids referenced by the sim (Go), the UI (TS) and the rotation APLs, builds a tooltip from
Spell.csv for each (name, rank, description, buff text) and:
  - writes/refreshes its row in assets/db_inputs/wowhead_spell_tooltips.csv (icon kept from Wowhead),
  - lists the ids in assets/db_inputs/server_spell_ids.txt, which gen_db adds to the DB spell icons.
Talent spells are left to gen_custom_spell_tooltips.py.

Usage:
  python tools/gen_server_spell_tooltips.py [--dry-run]
Then regenerate the DB:
  go run ./tools/database/gen_db -outDir=./assets -gen=db
"""
import argparse
import glob
import json
import os
import re
import sys

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from gen_custom_spell_tooltips import (  # noqa: E402
    CLASSES, COL_NAME, COL_SHIFT_FROM, FALLBACK_ICON, ICON_ID_OVERRIDES, NUM_RE, load_durations,
    load_radii, load_spell_csv, load_spell_icon_csv, resolve_desc, talent_spell_ranks)

COL_RANK = 129  # rank subtext ("Rank 10"), 9 after the name
COL_AURA = 147  # aura (buff/debuff) text, 9 after the description

GO_LINE_RE = re.compile(r"spell_?id", re.I)
INT_RE = re.compile(r"\b(\d{2,6})\b")
TS_RES = [re.compile(r"fromSpellId\(\s*(\d+)"), re.compile(r"spellId:\s*(\d+)")]
APL_RE = re.compile(r'"spellId":\s*(\d+)')
ENCHANT_SPELL_RE = re.compile(r"SpellId:\s*(\d+)")


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
    for p in glob.glob(os.path.join(repo, "ui/**/apls/*.json"), recursive=True):
        ids.update(int(x) for x in APL_RE.findall(open(p, encoding="utf-8").read()))
    # Enchant spells (the enchant picker and gear slot show these): the DB's enchants and the overrides.
    db_path = os.path.join(repo, "assets/database/db.json")
    if os.path.exists(db_path):
        db = json.load(open(db_path, encoding="utf-8"))
        ids.update(e["spellId"] for e in db.get("enchants", []) if e.get("spellId"))
    ids.update(int(x) for x in ENCHANT_SPELL_RE.findall(
        open(os.path.join(repo, "tools/database/enchant_overrides.go"), encoding="utf-8").read()))
    return ids


def load_extra_cols(path):
    """rank and aura text per spell, with the same column shift as load_spell_csv."""
    import csv
    ranks, auras = {}, {}
    for r in csv.reader(open(path, encoding="utf-8")):
        if not r or not r[0].isdigit():
            continue
        shift = 0
        for k in range(COL_NAME, min(len(r), COL_NAME + 6)):
            if r[k] and not NUM_RE.match(r[k]):
                shift = k - COL_NAME
                break
        sid = int(r[0])
        for col, out in ((COL_RANK, ranks), (COL_AURA, auras)):
            j = col + (shift if col >= COL_SHIFT_FROM else 0)
            v = r[j] if j < len(r) else ""
            if v and v != "0" and not re.fullmatch(r"0x[0-9A-Fa-f]+", v):
                out[sid] = v
    return ranks, auras


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--csv-dir", default="CSV's")
    ap.add_argument("--repo", default=".")
    ap.add_argument("--dry-run", action="store_true")
    args = ap.parse_args()

    csv_dir = os.path.join(args.repo, args.csv_dir)
    spell_path = os.path.join(csv_dir, "Spell.csv")
    names, icon_ids, raw_descs, spell_data = load_spell_csv(spell_path)
    ranks, auras = load_extra_cols(spell_path)
    durations = load_durations(os.path.join(csv_dir, "SpellDuration.csv"))
    radii = load_radii(os.path.join(csv_dir, "SpellRadius.csv"))
    icon_slugs = load_spell_icon_csv(os.path.join(csv_dir, "SpellIcon.csv"))

    talent_ids = set()
    for cn in CLASSES:
        talent_ids |= {sid for sid, _ in talent_spell_ranks(os.path.join(args.repo, "ui/core/talents/trees", cn + ".json"))}

    tooltip_path = os.path.join(args.repo, "assets/db_inputs/wowhead_spell_tooltips.csv")
    lines = open(tooltip_path, encoding="utf-8").read().splitlines()
    existing = {}
    for n, line in enumerate(lines):
        j = line.find(",")
        if j > 0 and line[:j].isdigit():
            existing[int(line[:j])] = n

    def clean(text, sid):
        if not text:
            return ""
        return resolve_desc(text, sid, spell_data, durations, radii)

    rows, unresolved = {}, []
    for sid in sorted(collect_ids(args.repo) - talent_ids):
        name = names.get(sid, "")
        if not name or name == "0" or NUM_RE.match(name):
            continue
        desc = clean(raw_descs.get(sid) if not re.fullmatch(r"0x[0-9A-Fa-f]+|0|", raw_descs.get(sid, "")) else "", sid)
        aura = clean(auras.get(sid, ""), sid)
        if not desc and not aura:
            continue
        old = {}
        if sid in existing:
            try:
                old = json.loads(lines[existing[sid]].split(",", 1)[1])
            except ValueError:
                old = {}
        icon = old.get("icon") or ICON_ID_OVERRIDES.get(icon_ids.get(sid, -1)) or icon_slugs.get(icon_ids.get(sid, -1)) or FALLBACK_ICON
        tooltip = f'<b class="whtt-name">{name}</b>'
        if sid in ranks:
            tooltip += f"<br />{ranks[sid]}"
        if desc:
            tooltip += f"<br />{desc}"
        if aura and aura != desc:
            tooltip += f'<br /><br /><span class="whtt-buff">Buff: {aura}</span>'
        if re.search(r"\bX\b|\$", tooltip):
            unresolved.append(f"  {sid} {name}: {desc} | {aura}"[:160])
            continue  # keeps the Wowhead tooltip
        rows[sid] = json.dumps({
            "name": name, "quality": None, "icon": icon, "tooltip": tooltip, "tooltip2": "",
            "buff": old.get("buff", "") or (aura and "1") or "", "spells": {}, "buffspells": {},
            "completion_category": 0,
        }, separators=(",", ":"))

    print(f"{len(rows)} server spell tooltips ({sum(1 for s in rows if s in existing)} replace Wowhead rows)")
    if unresolved:
        sys.stderr.write(f"{len(unresolved)} skipped (unresolved tokens, Wowhead tooltip stays):\n" + "\n".join(unresolved) + "\n")
    if args.dry_run:
        for sid in list(rows)[:15]:
            print(sid, rows[sid][:220])
        return

    for sid, row in rows.items():
        if sid in existing:
            lines[existing[sid]] = f"{sid},{row}"
        else:
            lines.append(f"{sid},{row}")
    with open(tooltip_path, "w", encoding="utf-8", newline="") as f:
        f.write("\n".join(lines) + "\n")
    with open(os.path.join(args.repo, "assets/db_inputs/server_spell_ids.txt"), "w", encoding="utf-8") as f:
        f.write("# Non-talent spell ids with local tooltips from Spell.csv. Maintained by tools/gen_server_spell_tooltips.py\n")
        f.write("\n".join(str(s) for s in rows) + "\n")
    print(f"Wrote {tooltip_path} and server_spell_ids.txt")


if __name__ == "__main__":
    main()
