"""Stats, slot, level, binding and usable classes of a server item, from its in-game tooltip lines (for /search).

parse(item, page_key, section) -> dict with only the non-empty fields.
"""
import re

CLASSES = ["Warrior", "Paladin", "Hunter", "Rogue", "Priest", "Shaman", "Mage", "Warlock", "Druid"]
ALL_CLASSES = (1 << len(CLASSES)) - 1


def mask(*names):
    m = 0
    for n in names:
        m |= 1 << CLASSES.index(n)
    return m


# armor and weapon proficiencies (classic 1.12; hunters / shamans learn mail at 40, still listed)
ARMOR_USERS = {
    "cloth": ALL_CLASSES,
    "leather": mask("Warrior", "Paladin", "Hunter", "Rogue", "Shaman", "Druid"),
    "mail": mask("Warrior", "Paladin", "Hunter", "Shaman"),
    "plate": mask("Warrior", "Paladin"),
}
WEAPON_USERS = {
    "Daggers": mask("Warrior", "Hunter", "Rogue", "Priest", "Shaman", "Mage", "Warlock", "Druid"),
    "Fist Weapons": mask("Warrior", "Hunter", "Rogue", "Shaman", "Druid"),
    "Axes": mask("Warrior", "Paladin", "Hunter", "Shaman"),
    "Maces": mask("Warrior", "Paladin", "Rogue", "Priest", "Shaman", "Druid"),
    "Swords": mask("Warrior", "Paladin", "Hunter", "Rogue", "Mage", "Warlock"),
    "Polearms": mask("Warrior", "Paladin", "Hunter"),
    "Staffs": mask("Warrior", "Hunter", "Priest", "Shaman", "Mage", "Warlock", "Druid"),
    "Bows": mask("Warrior", "Hunter", "Rogue"),
    "Guns": mask("Warrior", "Hunter", "Rogue"),
    "Crossbows": mask("Warrior", "Hunter", "Rogue"),
    "Thrown": mask("Warrior", "Hunter", "Rogue"),
    "Arrows": mask("Warrior", "Hunter", "Rogue"),
    "Bullets": mask("Warrior", "Hunter", "Rogue"),
    "Wands": mask("Priest", "Mage", "Warlock"),
}
TWO_HAND_ONLY = {"Axes": mask("Warrior", "Paladin", "Hunter", "Shaman"),
                 "Maces": mask("Warrior", "Paladin", "Shaman", "Druid"),
                 "Swords": mask("Warrior", "Paladin", "Hunter")}
ONE_HAND_ONLY = {"Axes": mask("Warrior", "Paladin", "Hunter", "Shaman"),
                 "Maces": mask("Warrior", "Paladin", "Rogue", "Priest", "Shaman", "Druid"),
                 "Swords": mask("Warrior", "Paladin", "Hunter", "Rogue", "Mage", "Warlock")}
OFF_HAND_WEAPON = mask("Warrior", "Hunter", "Rogue")
RELIC_USERS = {"Relics: Idols": mask("Druid"), "Relics: Librams": mask("Paladin"), "Relics: Totems": mask("Shaman")}
SHIELD_USERS = mask("Warrior", "Paladin", "Shaman")

# (key, label, regex): the regex's last group is the value; values of the same key add up
STATS = [
    ("str", "Strength", r"^\+(\d+) Strength$"),
    ("agi", "Agility", r"^\+(\d+) Agility$"),
    ("sta", "Stamina", r"^\+(\d+) Stamina$"),
    ("int", "Intellect", r"^\+(\d+) Intellect$"),
    ("spi", "Spirit", r"^\+(\d+) Spirit$"),
    ("armor", "Armor", r"^(\d+) Armor$"),
    ("block", "Block", r"^(\d+) Block$"),
    ("ap", "Attack power", r"^Equip: \+(\d+) Attack Power\.?$"),
    ("rap", "Ranged attack power", r"^Equip: \+(\d+) ranged Attack Power\.?$"),
    ("feral", "Feral attack power", r"^Equip: \+(\d+) Attack Power in Cat"),
    ("sp", "Spell damage & healing", r"^Equip: Increases damage and healing done by magical spells and effects by up to (\d+)"),
    ("heal", "Healing", r"^Equip: Increases healing done by spells and effects by up to (\d+)"),
    ("arcsp", "Arcane spell damage", r"^Equip: Increases damage done by Arcane spells and effects by up to (\d+)"),
    ("firsp", "Fire spell damage", r"^Equip: Increases damage done by Fire spells and effects by up to (\d+)"),
    ("frosp", "Frost spell damage", r"^Equip: Increases damage done by Frost spells and effects by up to (\d+)"),
    ("holsp", "Holy spell damage", r"^Equip: Increases damage done by Holy spells and effects by up to (\d+)"),
    ("natsp", "Nature spell damage", r"^Equip: Increases damage done by Nature spells and effects by up to (\d+)"),
    ("shasp", "Shadow spell damage", r"^Equip: Increases damage done by Shadow spells and effects by up to (\d+)"),
    ("hit", "Hit %", r"^Equip: Improves your chance to hit by (\d+)%"),
    ("crit", "Crit %", r"^Equip: Improves your chance to get a critical strike by (\d+)%"),
    ("shit", "Spell hit %", r"^Equip: Improves your chance to hit with spells by (\d+)%"),
    ("scrit", "Spell crit %", r"^Equip: Improves your chance to get a critical strike with spells by (\d+)%"),
    ("def", "Defense", r"^Equip: Increased Defense \+(\d+)"),
    ("dodge", "Dodge %", r"^Equip: Increases your chance to dodge an attack by (\d+)%"),
    ("parry", "Parry %", r"^Equip: Increases your chance to parry an attack by (\d+)%"),
    ("blockc", "Block chance %", r"^Equip: Increases your chance to block attacks with a shield by (\d+)%"),
    ("bv", "Block value", r"^Equip: Increases the block value of your shield by (\d+)"),
    ("mp5", "Mana per 5 sec", r"^Equip: Restores (\d+) mana per 5 sec"),
    ("hp5", "Health per 5 sec", r"^Equip: Restores (\d+) health per 5 sec"),
    ("pen", "Spell penetration", r"^Equip: Decreases the magical resistances of your spell targets by (\d+)"),
    ("arcres", "Arcane resistance", r"^\+(\d+) Arcane Resistance$"),
    ("firres", "Fire resistance", r"^\+(\d+) Fire Resistance$"),
    ("frores", "Frost resistance", r"^\+(\d+) Frost Resistance$"),
    ("natres", "Nature resistance", r"^\+(\d+) Nature Resistance$"),
    ("shares", "Shadow resistance", r"^\+(\d+) Shadow Resistance$"),
    ("allres", "All resistances", r"^(?:Equip: )?\+(\d+) All Resistances\.?$"),
    ("skill", "Weapon skill", r"^Equip: Increased (?:Daggers|Swords|Two-handed Swords|Axes|Two-handed Axes|Maces|"
                              r"Two-handed Maces|Bows|Guns|Crossbows|Polearms|Staves|Fist Weapons) \+(\d+)"),
    ("arpen", "Armor penetration", r"^Equip: Your attacks ignore (\d+) of your enemies' Armor"),
    ("haste", "Attack & casting speed %", r"^Equip: Increases your attack and casting speed by (\d+)%"),
    ("rhaste", "Ranged attack speed %", r"^Equip: Increases ranged attack speed by (\d+)%"),
    ("slots", "Bag slots", r"^(\d+) Slot (?:Bag|Quiver|Ammo Pouch|Soul Bag|Herb Bag|Enchanting Bag)$"),
]
# lines that count for two stats
BOTH = [
    (("hit", "shit"), re.compile(r"^Equip: Improves your chance to hit with attacks and spells by (\d+)%")),
    (("crit", "scrit"), re.compile(r"^Equip: Improves your critical strike chance for all attacks and spells by (\d+)%")),
]
STAT_RE = [((k,), re.compile(rx)) for k, _, rx in STATS] + BOTH
STAT_LABELS = [(k, label) for k, label, _ in STATS] + [("dps", "Damage per second"), ("speed", "Speed"),
                                                       ("min", "Min damage"), ("max", "Max damage")]
DMG_RE = re.compile(r"^(\d+) - (\d+) (?:\w+ )?Damage, Speed ([\d.]+)$")
SLOT_RE = re.compile(r"^(Head|Neck|Shoulder|Back|Chest|Shirt|Tabard|Wrist|Hands|Waist|Legs|Feet|Finger|Trinket|"
                     r"Main Hand|Off Hand|One-Hand|Two-Hand|Held In Off-hand|Ranged|Projectile|Thrown|Relic|Wand|Gun|"
                     r"Bow|Crossbow)(?:, .+)?$")
SLOT_NAME = {"Wand": "Ranged", "Gun": "Ranged", "Bow": "Ranged", "Crossbow": "Ranged", "Thrown": "Ranged"}
BIND = {"Binds when picked up": "bop", "Binds when equipped": "boe", "Binds when used": "bou", "Quest Item": "quest"}


def usable(page_key, section, lines):
    for ln in lines:
        if ln.startswith("Classes: "):
            return sum(1 << CLASSES.index(c.strip()) for c in ln[9:].split(",") if c.strip() in CLASSES)
    if page_key in ARMOR_USERS:
        return ARMOR_USERS[page_key]
    if section == "Shields":
        return SHIELD_USERS
    if section in RELIC_USERS:
        return RELIC_USERS[section]
    if page_key in ("weapons", "ranged"):
        m = re.match(r"^(.+?)(?: \((.+)\))?$", section)
        kind, hand = m.group(1), m.group(2) or ""
        users = WEAPON_USERS.get(kind, ALL_CLASSES)
        if hand == "Two-Hand":
            users = TWO_HAND_ONLY.get(kind, users)
        elif hand in ("One-Hand", "Main Hand"):
            users = ONE_HAND_ONLY.get(kind, users)
        elif hand == "Off Hand":
            users = ONE_HAND_ONLY.get(kind, users) & OFF_HAND_WEAPON
        return users
    return ALL_CLASSES


def parse(item, page_key, section):
    lines = item["lines"]
    out = {}
    stats = {}
    if page_key != "recipes":  # a recipe's tooltip shows the crafted item's stats
        for ln in lines:
            if ln.startswith("(") and " Set:" in ln[:8]:
                continue  # set bonuses are not the item's own stats
            m = DMG_RE.match(ln)
            if m and "min" not in stats:
                lo, hi, spd = int(m.group(1)), int(m.group(2)), float(m.group(3))
                stats.update(min=lo, max=hi, speed=spd)
                if spd:
                    stats["dps"] = round((lo + hi) / 2 / spd, 1)
                continue
            for keys, rx in STAT_RE:
                m = rx.match(ln)
                if m:
                    for k in keys:
                        stats[k] = stats.get(k, 0) + int(m.group(1))
                    break
        if "allres" in stats:
            for k in ("arcres", "firres", "frores", "natres", "shares"):
                stats[k] = stats.get(k, 0) + stats["allres"]
    if stats:
        out["st"] = stats
    for ln in lines[:6]:
        m = SLOT_RE.match(ln)
        if m and page_key != "recipes":
            out["sl"] = SLOT_NAME.get(m.group(1), m.group(1))
            break
    for ln in lines:
        m = re.match(r"^Requires Level (\d+)$", ln)
        if m:
            out["lv"] = int(m.group(1))
            break
    for ln in lines[:3]:
        if ln in BIND:
            out["b"] = BIND[ln]
            break
    u = usable(page_key, section, lines)
    if u != ALL_CLASSES:
        out["cl"] = u
    if "Unique" in lines[:4] or any(ln.startswith("Unique (") for ln in lines[:4]):
        out["u"] = 1
    return out
