"""Item cache records (wdb_items.py) -> in-game tooltip lines, in the same form as VPlusItemDB.lua, so the rest of
the item database (item_stats, item_types, reworked) reads them unchanged.

Needs the server's spell data (lokiy.dev/spells db: names, descriptions, effects) and a few client tables
(CSV's/dbc: ItemSet, SkillLine, Faction, AreaTable, Map, ItemDisplayInfo).
"""

import glob
import json
import math
import os
import re
import struct

QUALITY_NAMES = ["Poor", "Common", "Uncommon", "Rare", "Epic", "Legendary", "Artifact"]
BONDING = {1: "Binds when picked up", 2: "Binds when equipped", 3: "Binds when used", 4: "Quest Item",
           5: "Quest Item"}
SLOTS = {1: "Head", 2: "Neck", 3: "Shoulder", 4: "Shirt", 5: "Chest", 6: "Waist", 7: "Legs", 8: "Feet", 9: "Wrist",
         10: "Hands", 11: "Finger", 12: "Trinket", 13: "One-Hand", 14: "Off Hand", 15: "Ranged", 16: "Back",
         17: "Two-Hand", 19: "Tabard", 20: "Chest", 21: "Main Hand", 22: "Off Hand", 23: "Held In Off-hand",
         24: "Projectile", 25: "Thrown", 28: "Relic"}
WEAPONS = {0: "Axe", 1: "Axe", 2: "Bow", 3: "Gun", 4: "Mace", 5: "Mace", 6: "Polearm", 7: "Sword", 8: "Sword",
           10: "Staff", 13: "Fist Weapon", 15: "Dagger", 16: "Thrown", 17: "Spear", 18: "Crossbow", 19: "Wand",
           20: "Fishing Pole"}
ARMOR = {1: "Cloth", 2: "Leather", 3: "Mail", 4: "Plate", 6: "Shield", 7: "Libram", 8: "Idol", 9: "Totem"}
PROJECTILES = {2: "Arrow", 3: "Bullet"}
BAGS = {(1, 0): "Bag", (1, 1): "Soul Bag", (1, 2): "Herb Bag", (1, 3): "Enchanting Bag", (1, 4): "Engineering Bag",
        (11, 2): "Quiver", (11, 3): "Ammo Pouch"}
SCHOOLS = ["", "Holy", "Fire", "Nature", "Frost", "Shadow", "Arcane"]
RESISTS = ["Holy", "Fire", "Nature", "Frost", "Shadow", "Arcane"]  # item record order
RESIST_ORDER = [5, 1, 2, 3, 4, 0]  # the order the game client lists them in
STATS = {0: "Mana", 1: "Health", 3: "Agility", 4: "Strength", 5: "Intellect", 6: "Spirit", 7: "Stamina"}
STAT_ORDER = [4, 3, 7, 5, 6, 1, 0]  # the order the game client lists them in
CLASSES = {1: "Warrior", 2: "Paladin", 3: "Hunter", 4: "Rogue", 5: "Priest", 7: "Shaman", 8: "Mage", 9: "Warlock",
           11: "Druid"}
RACES = {1: "Human", 2: "Orc", 3: "Dwarf", 4: "Night Elf", 5: "Undead", 6: "Tauren", 7: "Gnome", 8: "Troll"}
ALL_CLASSES = sum(1 << (c - 1) for c in CLASSES)
ALL_RACES = sum(1 << (r - 1) for r in RACES)
ALLIANCE_RACES = (1 << 0) | (1 << 2) | (1 << 3) | (1 << 6)
REP_RANKS = ["Hated", "Hostile", "Unfriendly", "Neutral", "Friendly", "Honored", "Revered", "Exalted"]
# PvP rank titles by required honor rank (5 = first rank); the dump was scanned on a Horde character
HORDE_RANKS = ["Scout", "Grunt", "Sergeant", "Senior Sergeant", "First Sergeant", "Stone Guard", "Blood Guard",
               "Legionnaire", "Centurion", "Champion", "Lieutenant General", "General", "Warlord", "High Warlord"]
ALLIANCE_RANKS = ["Private", "Corporal", "Sergeant", "Master Sergeant", "Sergeant Major", "Knight",
                  "Knight-Lieutenant", "Knight-Captain", "Knight-Champion", "Lieutenant Commander", "Commander",
                  "Marshal", "Field Marshal", "Grand Marshal"]
TRIGGERS = {0: "Use: ", 1: "Equip: ", 2: "Chance on hit: ", 5: "Use: "}
SPELL_LEARN = 36  # spell effect: teaches the spell in EffectTriggerSpell
REGEN_AURAS = (84, 85)  # Mod Regen / Mod Power Regen: tick every 5 sec when the spell sets no period
FLAG_CONJURED = 2

# spell db columns (lokiy.dev/spells db/index.json)
C_REAGENT, C_REAGENT_COUNT, C_EFFECT, C_BASE_POINTS, C_AMPLITUDE, C_AURA = 42, 50, 61, 76, 94, 91
C_ITEM_TYPE, C_TRIGGER_SPELL, C_DESC = 103, 109, 122


def read_dbc(path):
    """Rows of a WDBC file as lists of ints, plus a function that reads a string field."""
    with open(path, "rb") as fh:
        buf = fh.read()
    n, fields, size, _ = struct.unpack_from("<4I", buf, 4)
    strings = 20 + n * size
    rows = [list(struct.unpack_from(f"<{fields}i", buf, 20 + i * size)) for i in range(n)]

    def text(off):
        return buf[strings + off:buf.index(b"\0", strings + off)].decode("utf-8", errors="replace")
    return rows, text


class Lookups:
    def __init__(self, spell_db, dbc_dir):
        self.spells = {}
        for path in glob.glob(os.path.join(spell_db, "*.json")):
            if not path.endswith("index.json"):
                with open(path, encoding="utf-8") as fh:
                    self.spells.update({int(k): v for k, v in json.load(fh).items()})
        self.sets, self.skills, self.factions, self.areas, self.maps, self.icons = {}, {}, {}, {}, {}, {}
        rows, text = read_dbc(os.path.join(dbc_dir, "ItemSet.dbc"))
        for r in rows:
            self.sets[r[0]] = {"name": text(r[1]), "items": [i for i in r[10:20] if i],
                               "bonuses": [(t, s) for s, t in zip(r[27:35], r[35:43]) if s],
                               "skill": r[43], "skillRank": r[44]}
        for name, col, dest in (("SkillLine", 3, self.skills), ("Faction", 19, self.factions),
                                ("AreaTable", 11, self.areas), ("Map", 4, self.maps),
                                ("ItemDisplayInfo", 5, self.icons)):
            rows, text = read_dbc(os.path.join(dbc_dir, name + ".dbc"))
            dest.update({r[0]: text(r[col]) for r in rows})
        self.icons = {k: v.lower() for k, v in self.icons.items() if v}

    def desc(self, sid):
        s = self.spells.get(sid)
        if not s:
            return ""
        text = (s.get("desc") or "").strip()
        raw = s["raw"]
        # The spells site counts $oN (total over the duration) as one tick when the effect has no period; food
        # and drink (Mod Regen) tick every 5 sec. Each $ token in the template is one number in the text.
        numbers = list(re.finditer(r"\d+", text))
        tokens = re.findall(r"\$(\d*)([a-z])(\d?)", raw[C_DESC])
        for k in reversed(range(min(len(tokens), len(numbers)))):
            ref, letter, n = tokens[k]
            src = self.spells.get(int(ref)) if ref else s
            if letter != "o" or not n or not src:
                continue
            r, n = src["raw"], int(n) - 1
            m = re.match(r"^(\d+) sec$", next((v for f, v in src.get("fields", []) if f == "Duration"), ""))
            per_tick = r[C_BASE_POINTS + n] + 1
            if m and r[C_AURA + n] in REGEN_AURAS and not r[C_AMPLITUDE + n] and numbers[k].group() == str(per_tick):
                text = text[:numbers[k].start()] + str(per_tick * int(m.group(1)) // 5) + text[numbers[k].end():]
        text = text.replace("$z", "your home")  # Hearthstone: the player's bind location
        # the client writes "1.50 min" for 1 min 30 sec, "1 hour", "2 hrs"
        text = re.sub(r"\b(\d+) min (\d+) sec\b", lambda x: f"{int(x.group(1)) + int(x.group(2)) / 60:.2f} min", text)
        return re.sub(r"\b(\d+) hr\b", lambda x: "1 hour" if x.group(1) == "1" else x.group(1) + " hrs", text)

    def name(self, sid):
        s = self.spells.get(sid)
        return s["raw"][120] if s else ""


def num(v):
    return f"{v:g}"


def slot_line(it):
    cls, sub, inv = it["class"], it["subclass"], it["inventoryType"]
    if (cls, sub) in BAGS and it["containerSlots"]:
        return f"{it['containerSlots']} Slot {BAGS[cls, sub]}"
    if inv == 26:  # guns, crossbows, wands: the client shows only the type
        return WEAPONS.get(sub, "") if cls == 2 else ""
    left = SLOTS.get(inv, "")
    right = WEAPONS.get(sub, "") if cls == 2 else ARMOR.get(sub, "") if cls == 4 else \
        PROJECTILES.get(sub, "") if cls == 6 else ""
    if cls == 2 and sub == 16:
        right = "Thrown"
    if inv == 16:  # cloaks: no armor type
        right = ""
    return ", ".join(x for x in (left, right) if x)


def damage_lines(it):
    out = []
    dmg, speed = it["damage"], it["delay"] / 1000
    if not dmg:
        return out
    if it["inventoryType"] == 24:  # ammo: average damage per second it adds
        lo, hi, _ = dmg[0]
        return [f"Adds {num(round((lo + hi) / 2, 1))} damage per second"]
    for n, (lo, hi, school) in enumerate(dmg):
        what = f"{SCHOOLS[school]} Damage" if school else "Damage"
        lo, hi = int(lo), math.ceil(hi - 1e-4)  # the client rounds the minimum down and the maximum up
        span = f"{lo} - {hi}" if lo != hi else f"{lo}"
        if n:
            out.append(f"+ {span} {what}")
        else:
            out.append(f"{span} {what}, Speed {speed:.2f}" if it["class"] == 2 else f"{span} {what}")
    if speed and it["class"] == 2:
        dps = sum((lo + hi) / 2 for lo, hi, _ in dmg) / speed
        out.append(f"({dps - 1e-6:.1f} damage per second)")  # the client rounds exact halves down
    return out


def requirement_lines(it, lk):
    out = []
    cmask = it["allowableClass"]
    if cmask > 0 and cmask & ALL_CLASSES != ALL_CLASSES:
        out.append("Classes: " + ", ".join(n for c, n in CLASSES.items() if cmask & (1 << (c - 1))))
    rmask = it["allowableRace"]
    if rmask > 0 and rmask & ALL_RACES != ALL_RACES:
        out.append("Races: " + ", ".join(n for r, n in RACES.items() if rmask & (1 << (r - 1))))
    if it["requiredLevel"] > 1:
        out.append(f"Requires Level {it['requiredLevel']}")
    if it["requiredSkill"]:
        out.append(f"Requires {lk.skills.get(it['requiredSkill'], '?')} ({it['requiredSkillRank']})")
    if it["requiredSpell"]:
        out.append(f"Requires {lk.name(it['requiredSpell'])}")
    if it["requiredHonorRank"] >= 5:
        alliance = rmask > 0 and rmask & ALL_RACES == ALLIANCE_RACES
        titles = ALLIANCE_RANKS if alliance else HORDE_RANKS
        out.append(f"Requires {titles[min(it['requiredHonorRank'] - 5, len(titles) - 1)]}")
    if it["requiredRepFaction"]:
        out.append(f"Requires {lk.factions.get(it['requiredRepFaction'], '?')} - "
                   f"{REP_RANKS[min(it['requiredRepValue'], 7)]}")
    return out


def body_lines(it, lk):
    """Everything below the name and above the spells."""
    out = []
    if it["flags"] & FLAG_CONJURED:
        out.append("Conjured Item")
    if it["area"]:
        out.append(lk.areas.get(it["area"], ""))
    if it["map"]:
        out.append(lk.maps.get(it["map"], ""))
    if it["bonding"] in BONDING:
        out.append(BONDING[it["bonding"]])
    if it["maxCount"] == 1:
        out.append("Unique")
    elif it["maxCount"] > 1:
        out.append(f"Unique ({it['maxCount']})")
    if it["startQuest"]:
        out.append("This Item Begins a Quest")
    slot = slot_line(it)
    if slot:
        out.append(slot)
    out += damage_lines(it)
    if it["armor"]:
        out.append(f"{it['armor']} Armor")
    if it["block"]:
        out.append(f"{it['block']} Block")
    stats = dict((t, v) for t, v in it["stats"])
    for t in STAT_ORDER:
        if stats.get(t):
            out.append(f"{stats[t]:+d} {STATS[t]}")
    for n in RESIST_ORDER:
        if it["resist"][n]:
            out.append(f"{it['resist'][n]:+d} {RESISTS[n]} Resistance")
    return [ln for ln in out if ln]


def recipe_lines(sp, lk, items):
    """A recipe's 'Use: Teaches you...' spell -> the crafted item's tooltip and the reagents."""
    raw = lk.spells.get(sp["id"], {}).get("raw")
    if not raw or SPELL_LEARN not in raw[C_EFFECT:C_EFFECT + 3]:
        return []
    taught = lk.spells.get(raw[C_TRIGGER_SPELL + raw[C_EFFECT:C_EFFECT + 3].index(SPELL_LEARN)], {}).get("raw")
    made = items.get(taught[C_ITEM_TYPE]) if taught else None
    if not made:
        return []
    out = [made["name"]] + body_lines(made, lk) + requirement_lines(made, lk) + spell_lines(made, lk, items)
    if made["description"]:
        out.append(f'"{made["description"]}"')
    out += set_lines(made, lk, items)
    reagents = []
    for rid, n in zip(taught[C_REAGENT:C_REAGENT + 8], taught[C_REAGENT_COUNT:C_REAGENT_COUNT + 8]):
        if rid > 0 and rid in items:
            reagents.append(items[rid]["name"] + (f" ({n})" if n > 1 else ""))
    if reagents:
        out.append("Requires " + ", ".join(reagents))
    return out


def spell_lines(it, lk, items, recipes=False):
    out = []
    for sp in it["spells"]:
        prefix = TRIGGERS.get(sp["trigger"])
        text = lk.desc(sp["id"])
        if prefix and text:
            out.append(prefix + text)
        if recipes and sp["trigger"] == 0:
            out += recipe_lines(sp, lk, items)
    return out


def set_lines(it, lk, items):
    s = lk.sets.get(it["itemSet"])
    if not s:
        return []
    out = [f"{s['name']} (0/{len(s['items'])})"]
    if s["skill"]:
        out.append(f"Requires {lk.skills.get(s['skill'], '?')} ({s['skillRank']})")
    out += [items[i]["name"] for i in s["items"] if i in items]
    for threshold, sid in sorted(s["bonuses"], key=lambda b: b[0]):
        text = lk.desc(sid)
        if text:
            out.append(f"({threshold}) Set: {text}")
    return out


def tooltip(it, lk, items):
    """Tooltip lines like VPlusItemDB.lua: name first, 'Rarity: X' last."""
    out = [it["name"]] + body_lines(it, lk) + requirement_lines(it, lk) + spell_lines(it, lk, items, recipes=True)
    if it["description"]:
        out.append(f'"{it["description"]}"')
    out += set_lines(it, lk, items)
    out.append("Rarity: " + QUALITY_NAMES[min(it["quality"], 6)])
    return [ln.strip() for ln in out if ln.strip()]
