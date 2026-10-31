"""Spells for db.lokiy.dev's /spell/<id> pages (sim repo Part GE, 2026-10-25).

Reads the spell database that lokiy.dev/spells' generator builds (the spell-changelog project, spell_changelog.py
`db`, deployed to /var/www/lokiy/spells/db): index.json ([id, name, rank, icon, class] per spell), <id // 200>.json
(readable fields, rendered description / buff text, raw columns) and classic.json (every spell against stock 1.12.1:
"new" ids, "changed" {id: [[label, classic, Vanilla+, flipped flags?]]}). The parsing lives there, not here.
"""
import json
import os
import re

SPELL_DB = "/var/www/lokiy/spells/db"
STATUS_LABEL = {"new": "New on Vanilla+", "changed": "Changed from classic", "same": "Same as classic"}
# the fields shown in the "Details" table (in this order); effects get their own table, the rest is folded away
MAIN_FIELDS = ["School", "Cast time", "Range", "Cost", "Cost (% of base mana)", "Cooldown", "Category cooldown",
               "Global cooldown", "Duration", "Spell level", "Max level", "Proc chance", "Proc charges", "Stacks",
               "Dispel type", "Mechanic", "Reagent 1", "Reagent 2", "Reagent 3", "Reagent 4", "Reagent 5",
               "Reagent 6", "Reagent 7", "Talent", "Skill line", "Triggered by"]


def load_spells():
    """{spell id: {"name", "rank", "icon", "class", "fields", "desc", "tip", "status", "diff"}}; {} if the spell
    database isn't there."""
    path = os.path.join(SPELL_DB, "index.json")
    if not os.path.exists(path):
        return {}
    index = json.load(open(path, encoding="utf-8"))
    records = {}
    for fn in os.listdir(SPELL_DB):
        if re.fullmatch(r"\d+\.json", fn):
            records.update({int(k): v for k, v in json.load(open(os.path.join(SPELL_DB, fn), encoding="utf-8")).items()})
    try:
        classic = json.load(open(os.path.join(SPELL_DB, "classic.json"), encoding="utf-8"))
    except (OSError, ValueError):
        classic = {}
    new = set(classic.get("new", []))
    changed = {int(k): v for k, v in classic.get("changed", {}).items()}
    out = {}
    for sid, name, rank, icon, cls in index["spells"]:
        r = records.get(sid, {})
        status = ("new" if sid in new else "changed" if sid in changed else "same") if classic else ""
        out[sid] = {"name": name, "rank": rank, "icon": (icon or "").lower(), "class": cls,
                    "fields": r.get("fields", []), "desc": r.get("desc") or "", "tip": r.get("tip") or "",
                    "status": status, "diff": changed.get(sid, [])}
    return out


SPELL_TOKEN = re.compile(r"\[spell:(\d+)\]\s*([^·\[]*?)(?=\s*(?:·|\[|$))")
ITEM_TOKEN = re.compile(r"\[item:(\d+)\]")
ENCHANT_TOKEN = re.compile(r"\[enchant:\d+\]\s*")


def link_tokens(text, item_names, spell_ids=None):
    """Escaped field text with the generator's [spell:N] Name / [item:N] / [enchant:N] Name tokens as links (a spell
    only when spell_ids has it: one without a page stays text)."""
    text = SPELL_TOKEN.sub(lambda m: f'<a href="/spell/{m.group(1)}">{m.group(2) or "#" + m.group(1)}</a>'
                           if spell_ids is None or int(m.group(1)) in spell_ids else (m.group(2) or "#" + m.group(1)),
                           text)
    # an item the database has gets a link; anything else (a class mask in an Item field, a removed item) stays text
    text = ITEM_TOKEN.sub(lambda m: f'<a href="/item/{m.group(1)}">{item_names[int(m.group(1))]}</a>'
                          if int(m.group(1)) in item_names else m.group(1), text)
    return ENCHANT_TOKEN.sub("", text)


def tooltip_lines(fields):
    """The spell tooltip's cost / range / cast / cooldown lines as game-style text, for api/spells.json (Part GH):
    {"cost": "90 Mana", "range": "30 yd", "cast": "1.5 sec cast" / "Instant", "cd": "1 min cooldown"}; empty ones
    left out (no range for Self, no cast / range for passive spells)."""
    f = dict(fields)
    out = {}
    cost = f.get("Cost") or (f"{f['Cost (% of base mana)']} of base mana" if f.get("Cost (% of base mana)") else "")
    if cost:
        out["cost"] = cost
    if "Passive" not in (f.get("Attributes") or ""):
        if f.get("Range") and f["Range"] != "Self":
            out["range"] = f["Range"]
        cast = f.get("Cast time")
        if cast:
            out["cast"] = cast if cast == "Instant" else f"{cast} cast"
    cd = f.get("Cooldown") or f.get("Category cooldown")
    if cd:
        out["cd"] = f"{cd} cooldown"
    return out
