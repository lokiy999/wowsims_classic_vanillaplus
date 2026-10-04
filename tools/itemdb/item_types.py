"""Item type of a server item, from its in-game tooltip lines (VPlusItemDB.lua has no class / subclass).

classify(item) -> (page key, section); PAGES lists the type pages in display order.
"""
import re

PAGES = [
    ("weapons", "Weapons"),
    ("ranged", "Ranged weapons & ammo"),
    ("cloth", "Cloth armor"),
    ("leather", "Leather armor"),
    ("mail", "Mail armor"),
    ("plate", "Plate armor"),
    ("accessories", "Cloaks, jewelry & trinkets"),
    ("shields", "Shields, off-hands & relics"),
    ("consumables", "Consumables"),
    ("recipes", "Recipes"),
    ("containers", "Bags & quivers"),
    ("mounts", "Mounts & pets"),
    ("quest", "Quest items"),
    ("other", "Trade goods & other"),
    ("junk", "Junk"),
]
PAGE_NAMES = dict(PAGES)

ARMOR_SLOTS = ["Head", "Shoulder", "Chest", "Wrist", "Hands", "Waist", "Legs", "Feet"]
MATERIALS = {"Cloth": "cloth", "Leather": "leather", "Mail": "mail", "Plate": "plate"}
ACCESSORY = {"Back": "Cloaks", "Neck": "Necks", "Finger": "Rings", "Trinket": "Trinkets", "Shirt": "Shirts",
             "Tabard": "Tabards"}
MELEE_SLOTS = ("One-Hand", "Two-Hand", "Main Hand", "Off Hand")
RANGED = {"Bow": "Bows", "Gun": "Guns", "Crossbow": "Crossbows", "Wand": "Wands", "Thrown": "Thrown",
          "Arrow": "Arrows", "Bullet": "Bullets"}
RECIPE_PREFIX = {"Pattern:": "Patterns", "Plans:": "Plans", "Recipe:": "Recipes", "Formula:": "Formulas",
                 "Schematic:": "Schematics", "Manual:": "Manuals"}
SLOT_RE = re.compile(r"^(Head|Neck|Shoulder|Back|Chest|Shirt|Tabard|Wrist|Hands|Waist|Legs|Feet|Finger|Trinket|"
                     r"Main Hand|Off Hand|One-Hand|Two-Hand|Held In Off-hand|Ranged|Projectile|Thrown|Relic|Wand|Gun|"
                     r"Bow|Crossbow)(?:, (.+))?$")


def plural(s):
    return s if s.endswith("s") else s + ("es" if s.endswith(("x", "sh")) else "s")


def classify(item):
    lines = item["lines"]
    text = "\n".join(lines)
    name = item["name"]
    for prefix, section in RECIPE_PREFIX.items():
        if name.startswith(prefix):
            return "recipes", section  # before the slot check: a recipe's tooltip shows the crafted item
    if re.match(r"^(Libram|Grimoire|Tome|Codex|Book|Guide):|^Grimoire of ", name):
        return "recipes", "Class books & guides"
    for ln in lines[:6]:
        m = SLOT_RE.match(ln)
        if not m:
            continue
        slot, sub = m.group(1), m.group(2) or ""
        if slot in ARMOR_SLOTS:
            if sub in MATERIALS:
                return MATERIALS[sub], slot
            return "accessories", "Other armor"
        if slot in ACCESSORY:
            return "accessories", ACCESSORY[slot]
        if slot == "Held In Off-hand":
            return "shields", "Held in off-hand"
        if slot == "Relic":
            return "shields", f"Relics: {plural(sub)}" if sub else "Relics"
        if slot == "Off Hand" and sub == "Shield":
            return "shields", "Shields"
        if slot in MELEE_SLOTS:
            if sub == "Fishing Pole":
                return "other", "Fishing poles"
            return "weapons", f"{plural(sub) if sub else 'Other'} ({slot})"
        key = sub if slot in ("Ranged", "Projectile", "Thrown") and sub else slot
        return "ranged", RANGED.get(key, plural(key))
    if re.search(r"^\d+ Slot (Bag|Quiver|Ammo Pouch|Soul Bag|Herb Bag|Enchanting Bag|.*Bag)", text, re.M):
        if "Quiver" in text or "Ammo Pouch" in text:
            return "containers", "Quivers & ammo pouches"
        return "containers", "Bags"
    if re.search(r"summon and dismiss|rideable|mount", text, re.I) and "Use:" in text:
        return "mounts", "Mounts" if re.search(r"rideable|mount", text, re.I) else "Pets"
    if "Quest Item" in lines or "This Item Begins a Quest" in lines:
        return "quest", "Quest items"
    use = "Use:" in text
    if use:
        lo = text.lower()
        if "while eating" in lo:
            return "consumables", "Food"
        if "while drinking" in lo:
            return "consumables", "Drinks"
        if "Requires First Aid" in text or "Bandage" in name:
            return "consumables", "Bandages"
        if "Potion" in name or "Draught" in name:
            return "consumables", "Potions"
        if re.search(r"Elixir|Flask", name):
            return "consumables", "Elixirs & flasks"
        if "Scroll" in name:
            return "consumables", "Scrolls"
        if "Classes: Rogue" in text and "Poison" in name:
            return "consumables", "Poisons"
        if re.search(r"Oil|Sharpening|Weightstone|Armor Kit|Scope", name):
            return "consumables", "Item enhancements"
        if "Requires Engineering" in text or re.search(r"Bomb|Grenade|Dynamite|Charge|Firework|Rocket", name):
            return "consumables", "Engineering & fireworks"
        if "Requires Enchanting" in text or "Teaches you" in text:
            return "recipes", "Other recipes"
        if item["quality"] == 0:
            return "junk", "Junk"
        return "consumables", "Other usable items"
    if item["quality"] == 0:
        return "junk", "Junk"
    return "other", "Trade goods & other"
