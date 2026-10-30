"""Where items come from, and which items are not on Vanilla+ (db.lokiy.dev; sim repo Part FR, 2026-10-25).

An item's sources, by layer (Lokiy 2026-10-22 / 24: db.lokiy.dev is the source of truth; Wowhead only for what our
own data is missing):
  1. AtlasLoot (the server's fork) loot tables, incl. the boss drops the loot tracker added (gen_site.py).
  2. The old item workbook (workbook_sources.json, a one-time import of Item Template.xlsx by import_workbook.py):
     the hand-checked Source text and its type. Used when AtlasLoot has no table for the item, or only set tables
     (the workbook knows the tier set's raid and its token).
  3. Lokiy's loot tracker: the mobs it saw drop the item.
  4. pfQuest's classic database (pfquest_sources.json, import_pfquest.py): NPC drops with the classic chance,
     vendors, objects. Classic data, not Vanilla+.
  5. Wowhead's classic sources (wowhead_sources.json), marked "from Wowhead".
The first layer that has something is the item's source; layers 2-5 are only gap fillers.

Not on Vanilla+ (Lokiy 2026-10-22, 2026-10-25; left out of the database since 2026-10-30, gen_site.py Part HF): AQ20, AQ40 and Naxxramas (their
loot tables, the ~130 names of AQ_T3_REMOVE_NAMES, and Wowhead-only drops there), except Kurinnaxx (a world boss in
Silithus); Alterac Valley; the Brood of Nozdormu ring; Ashbringer and the Warglaives; the workbook's Removed Items
sheet; the old (classic) versions of the tier sets (Lokiy 2026-10-25: there is no Tier 2 on Vanilla+; each class has
2-4 Tier 1 sets with 2/4/6/8-piece bonuses, incl. Wrath, Transcendence, Judgement, Stormrage, Ten Storms,
Netherwind, Nemesis, Dragonstalker, Bloodfang; a version of the same set with 3/5/8-piece bonuses is the old one);
items Lokiy named (LOKIY_NOT_ON_SERVER). An item the loot tracker has seen drop is never marked: it is reported as a
conflict instead.
"""
import json
import os
import re

HERE = os.path.dirname(os.path.abspath(__file__))
CLASSIC_MAX_ITEM_ID = 24283

RAIDS = {"Molten Core", "Blackwing Lair", "Zul'Gurub", "Onyxia's Lair", "Scarlet Monastery",  # SM: a 10-man raid here
         "Ruins of Ahn'Qiraj", "Temple of Ahn'Qiraj", "Naxxramas"}
GROUP_TYPE = {"WorldBosses": "world boss", "PvP": "pvp", "Crafting": "crafting", "Factions": "reputation",
              "WorldEvents": "event", "Sets": "set"}

# ---- not on Vanilla+
CUT_INSTANCES = {"Ruins of Ahn'Qiraj": "AQ20", "Ruins of Ahn'Qiraj Sets": "AQ20", "Temple of Ahn'Qiraj": "AQ40",
                 "Temple of Ahn'Qiraj Sets": "AQ40", "Naxxramas": "Naxxramas", "Tier 3 Sets": "Naxxramas"}
CUT_TEXT = {"AQ20": "Ruins of Ahn'Qiraj (AQ20) is not on Vanilla+.",
            "AQ40": "Temple of Ahn'Qiraj (AQ40) is not on Vanilla+.",
            "Naxxramas": "Naxxramas is not on Vanilla+.",
            "AV": "Alterac Valley is not on Vanilla+.",
            "Brood": "The Brood of Nozdormu ring (an AQ40 reputation) is not on Vanilla+.",
            "old SM": "An item of the old Scarlet Monastery (the classic 5-man; on Vanilla+ it is a level 60 raid). It "
                      "can't drop anymore; a BoE that dropped back then may still exist on the server, but maybe none.",
            "old set": "The old (classic) version of this set, with 3/5/8-piece bonuses. Vanilla+ has its own "
                       "Tier 1 version with 2/4/6/8-piece bonuses.",
            "AQ/Naxx": "From AQ20, AQ40 or Naxxramas, which are not on Vanilla+.",
            "old pvp": "The old (before patch 1.12) version of this PvP rank set, sold by Captain Dirgehammer / Lady "
                       "Palanseer. Vanilla+ has the patch 1.12 version of the set (AtlasLoot's PvP sets), with other "
                       "item IDs."}
KURINNAXX = "Kurinnaxx"  # AQ20's first boss is a world boss in Silithus here (13 items)
KURINNAXX_NPC = 15348
AV_TABLES = {"AVRepFriendly", "AVRepHonored", "AVRepRevered", "AVRepExalted", "Stormpike1", "Frostwolf1"}
BROOD_TABLES = {"AQBroodRings"}
# Wowhead: zones / reputations of the content above (AQ40 3428, AQ20 3429, Naxxramas 3456, Alterac Valley 2597;
# Frostwolf Clan 729, Stormpike Guard 730, Brood of Nozdormu 910)
WH_CUT_ZONES = {3428: "AQ40", 3429: "AQ20", 3456: "Naxxramas", 2597: "AV"}
WH_CUT_REPS = {729: "AV", 730: "AV", 910: "Brood"}
FORCE_IDS = {
    13262: "Ashbringer can't be obtained: its quest chain was never completable.",
    18582: "The Warglaives of Azzinoth are level 70 (TBC) items and can't be obtained.",
    18583: "The Warglaives of Azzinoth are level 70 (TBC) items and can't be obtained.",
    18584: "The Warglaives of Azzinoth are level 70 (TBC) items and can't be obtained.",
    22736: "Can't be obtained: part of the Atiesh quest line, it doesn't drop (Lokiy 2026-10-28).",
    12947: "Can't be obtained (Lokiy 2026-10-28).",  # Alex's Ring of Audacity
    18943: "Can't be obtained, and neither can its quest (Lokiy 2026-10-28).",  # Dark Iron Pillow
    21038: CUT_TEXT["AV"],  # Hardpacked Snowball (Lokiy 2026-10-28)
}
# Lokiy's answers (2026-10-25): not on Vanilla+ by name
LOKIY_NOT_ON_SERVER = {"Black Grasp of the Destroyer", "Plans: Black Grasp of the Destroyer",
                       # Lokiy 2026-10-25 (from the to-check list): unavailable; their plans with them
                       "Obsidian Mail Tunic", "Plans: Obsidian Mail Tunic",
                       "Thick Obsidian Breastplate", "Plans: Thick Obsidian Breastplate",
                       "Black Qiraji Resonating Crystal"}
# Alterac Valley items with no other source (Lokiy 2026-10-30, Part HI; hub data\\av_removal_list.md): medals, flesh,
# hides, banners, supplies, Skull of Korrak, Steamsaw, Intricate Bauble
AV_ONLY_IDS = {5430, 17326, 17327, 17328, 17411, 17502, 17503, 17522, 17542, 17642, 17643, 17849, 17850, 18148}
# Lokiy 2026-10-28: confirmed not on Vanilla+, by item ID
LOKIY_NOT_ON_SERVER_IDS = {
    # Kurinnaxx's classic AQ20 loot (on Vanilla+ he is a world boss)
    21498, 21499, 21500, 21502, 21503,
    # old Scarlet Monastery 5-man: the rares (Azshir the Sleepless, Fallen Champion, Ironspine) and Bloodmage Thalnos
    7685, 7686, 7687, 7688, 7689, 7690, 7691, 7708, 7709, 7731,
    # AQ40 mount crystals (Blue, Red, Green; Black is in LOKIY_NOT_ON_SERVER)
    21218, 21321, 21323,
    7484,  # Ranger Wristguards (a classic world drop)
    3777,  # Lethargy Root (Lokiy 2026-10-28)
    # Lokiy 2026-10-28: the McWeaksauce sodas, Fake Mistletoe, Ring of the Dawn, Bonebiter, Whipweed Heart, Venture Co.
    # Ledger
    23578, 23579, 17195, 13812, 6830, 19911, 7389,
    # Lokiy 2026-10-28 (second list): the old SM 5-man bosses' loot and the old Scarlet set's leggings
    7682, 7684, 7710, 7711, 7712, 7718, 7719, 7756, 10330,
    13175,  # Voone's Twitchbow
    13247,  # Quartermaster Zigris' Footlocker
    15886,  # Timolain's Phylactery
    19882,  # The Hexxer's Head
    21324,  # Yellow Qiraji Resonating Crystal
    # Lokiy 2026-10-28 (third list): the Scourge Invasion letters, Crude Map, Large Rope Net
    22970, 22972, 22973, 22974, 22975, 22977, 21037, 835,
    # the AV junk drops and the AQ junk / Qiraji raid materials (Lokiy: unobtainable)
    18225, 18226, 18227, 18228, 18229, 18230, 18231, 18233, 18234, 21222, 21223, 21224, 21225, 21226, 21227,
}
# The rank 7-10 PvP sets (rare) exist twice: the old vendor version (16xxx / 17xxx) and patch 1.12's (22843-23319),
# which is the one on Vanilla+ (in AtlasLoot's PvP set tables, seen in item caches). Lokiy 2026-10-28: mark the old.
OLD_PVP_SET = re.compile(r"^(Knight-Lieutenant|Knight-Captain|Lieutenant Commander|Blood Guard|Legionnaire|Champion)'s ")
OLD_PVP_MAX_ID = 18000
# Lokiy 2026-10-28: 99% sure not on Vanilla+
LOKIY_PROBABLY_NOT_IDS = {11768, 16999}  # Incendic Bracers, Royal Seal of Alexis
# Lokiy 2026-10-28: source type by item (every source of the item gets it): Scorn's loot and the Scourge Invasion set
# pieces are listed as dungeon (Candy Corn 20389 was a typo for 23089)
ITEM_TYPE = {i: "dungeon" for i in (23168, 23169, 23170, 23085, 23087, 23088, 23089, 23090, 23091, 23092, 23093)}
# Lokiy 2026-10-28: "all items in No known source are unobtainable" (/unseen's group of classic items without a source)
LOKIY_UNOBTAINABLE_IDS = {7683, 17733, 19105, 19109, 19110, 19111, 19112, 19113, 19508}
# Lokiy 2026-10-28: classic items that have dropped on Vanilla+, though no item cache upload has them yet: off /unseen,
# the tooltip stays classic until a cache has it (gen_site.py)
# AtlasLoot lists these on one boss, but they are classic world drops and the loot tracker saw them drop from trash on
# Vanilla+ too: pfQuest's world drop line follows the AtlasLoot sources
ALSO_WORLD_DROP = {18335}  # Pristine Black Diamond (LBRS Blackhand trash, 2026-10-29)
# Drops seen in other players' loot logs that Lokiy confirmed: {item id: [(npc id, mob name, zone, times seen, log)]}.
# Added after the item's other sources as "Drops from <mob> (<log>, seen n x)"; no rate (the logs' kill counts don't
# line up with their loot). Proposal: data\external loot\rist_proposal.md in the hub.
EXTERNAL_DROPS = {
    12662: [(11454, "Wildspawn Betrayer", "Dire Maul", 1, "Rist's log"),  # Demonic Rune (Lokiy 2026-10-30)
            (13276, "Wildspawn Imp", "Dire Maul", 2, "Rist's log")],
    # the rest of the proposal's "maybe" list (Lokiy 2026-10-30; not Living Essence)
    18364: [(14349, "Pimgib", "Dire Maul", 1, "Rist's log"),  # The Emerald Dream
            (11455, "Wildspawn Felsworn", "Dire Maul", 1, "Rist's log")],
    18357: [(11455, "Wildspawn Felsworn", "Dire Maul", 1, "Rist's log")],  # Codex of Defense
    19272: [(11443, "Gordok Ogre-Mage", "Dire Maul", 1, "Rist's log")],  # Five of Elementals
    17964: [(11583, "Nefarian", "Blackwing Lair", 1, "Rist's log")],  # Gray Sack of Gems
    12704: [(11451, "Wildspawn Satyr", "Dire Maul", 1, "Rist's log")],  # Plans: Thorium Leggings
    8267: [(11443, "Gordok Ogre-Mage", "Dire Maul", 1, "Rist's log")],  # Ebonhold Gauntlets
    10203: [(11443, "Gordok Ogre-Mage", "Dire Maul", 1, "Rist's log")],  # Overlord's Chestplate
    5758: [(11443, "Gordok Ogre-Mage", "Dire Maul", 1, "Rist's log"),  # Mithril Lockbox
           (1791, "Slavering Ghoul", "Western Plaguelands", 1, "Rist's log")],
    12014: [(11443, "Gordok Ogre-Mage", "Dire Maul", 1, "Rist's log")],  # Arctic Ring
    7909: [(11673, "Ancient Core Hound", "Molten Core", 1, "Rist's log")],  # Aquamarine
}
LOKIY_DROPPED_IDS = {18259, 18252, 18290, 14456, 18292, 22739, 20635, 20736, 12714, 20734, 20735, 20728, 20727, 20731}
# Lokiy 2026-10-25: on Vanilla+ after all, with their real source (overrides every rule and every other source)
LOKIY_ON_SERVER = {
    "The Lobotomizer": ("Upper Blackrock Spire - Lord Valthalak (seen once)", "dungeon"),
    # its only data source was an Alterac Valley trogg (gone since Part HI); hub, Lokiy 2026-10-30
    "Heavy Junkbox": ("Pickpocketed by rogues from humanoids of level 50 and up (classic)", "other"),
}
# Lokiy 2026-10-25: the AQ class books and three glove formulas are one loot set that the four world bosses share
# (their drop rates are shared between them): its own loot table, built by add_shared_world_boss_loot()
SHARED_WB_KEY = "WBSharedLoot"
SHARED_WB_NAME = "World Boss Shared Loot"
SHARED_WB_BOSSES = [("KKazzak", "Lord Kazzak"), ("WBKurinnaxx", "Kurinnaxx"), ("Hederine", "Lady Hederine"),
                    ("AAzuregos", "Azuregos")]
SHARED_WB_TABLES = ["AQ20ClassBooks"]  # every item of these tables moves to the shared table
# the glove formulas by item ID: AtlasLoot has 20727 / 20728 / 20731 in AQ20 and 26021 (a Vanilla+ formula) on
# Azuregos; those four aren't in the item data yet (WebDKP awards name them). Enchant Gloves - Living Power (spell
# 34347) is the 8th; its formula's item ID isn't known yet.
SHARED_WB_ITEMS = {20726: "Formula: Enchant Gloves - Threat", 20729: "Formula: Enchant Gloves - Fire Power",
                   20730: "Formula: Enchant Gloves - Healing Power", 20727: "Formula: Enchant Gloves - Shadow Power",
                   20728: "Formula: Enchant Gloves - Frost Power", 20731: "Formula: Enchant Gloves - Superior Agility",
                   26021: "Formula: Enchant Gloves - Arcane Power",
                   # the AQ cloak formulas too (Lokiy 2026-10-25: "cloak yes, spellpower no"): 20734 / 20736 were
                   # in AtlasLoot's AQEnchants, 20735 (AQ40 in classic) nowhere
                   20734: "Formula: Enchant Cloak - Stealth", 20735: "Formula: Enchant Cloak - Subtlety",
                   20736: "Formula: Enchant Cloak - Dodge"}
SHARED_WB_TEXT = "World bosses: " + ", ".join(n for _, n in SHARED_WB_BOSSES) + " (shared loot)"
# Old SM blue BoEs that Wowhead Classic also shows dropping outside the instance (checked 2026-10-25, Lokiy: "blue
# BoEs from SM could be obtainable"): on Vanilla+, with that source. Sunblaze Coif and Sliverblade drop only in SM.
OUTSIDE_SM = {
    7728: "Tirisfal Glades - Scarlet Disciple, Scarlet Augur (outside Scarlet Monastery)",
    7754: "Tirisfal Glades - Scarlet Augur, Scarlet Preserver, Scarlet Disciple, Scarlet Scout (outside Scarlet "
          "Monastery)",
}
# The old Scarlet Monastery 5-man's loot (Lokiy 2026-10-25: every non-level-60 SM item is unobtainable): its trash
# tables, and items whose Wowhead sources are all drops there (zone 796)
OLD_SM_TABLES = {"SMATrash", "SMCTrash", "SMGTrash", "SMLTrash"}
SM_ZONE = 796
# deliberately not marked yet, for Lokiy to check (/removed lists them on top)
# (Flintrock Shoulders 26318, on the workbook's Removed Items sheet: Lokiy 2026-10-28, a confirmed drop from Houndmaster
# Loksey in SM; gen_site.py MANUAL_DROPS)
TO_CHECK = {
    23168: "Lokiy 2026-10-28: maybe on Vanilla+, only during special events; may be reworked (Scorn, Scourge Invasion).",
    23169: "Lokiy 2026-10-28: maybe on Vanilla+, only during special events; may be reworked (Scorn, Scourge Invasion).",
    23170: "Lokiy 2026-10-28: maybe on Vanilla+, only during special events; may be reworked (Scorn, Scourge Invasion).",
}
# Removed Items sheet rows that are not removals (they stay on the site)
REMOVED_SHEET_KEEP = {"Malfurion's Signet Ring"}  # a source fix (a quest reward, not an Emerald Dragons drop)

# Items AtlasLoot never listed as AQ / Naxx (workbook write_all.py AQ_T3_REMOVE_NAMES)
AQ_T3_REMOVE_NAMES = {
    # AQ20 class quest reward items (weapon/cloak-or-cape/ring per class)
    'Blade of Vaulted Secrets', 'Drape of Vaulted Secrets', 'Band of Vaulted Secrets',
    'Kris of Unspoken Names', 'Shroud of Unspoken Names', 'Ring of Unspoken Names',
    'Gavel of Infinite Wisdom', 'Shroud of Infinite Wisdom', 'Ring of Infinite Wisdom',
    'Dagger of Veiled Shadows', 'Cloak of Veiled Shadows', 'Band of Veiled Shadows',
    'Mace of Unending Life', 'Cloak of Unending Life', 'Band of Unending Life',
    'Scythe of the Unseen Path', 'Cloak of the Unseen Path', 'Signet of the Unseen Path',
    'Blade of Eternal Justice', 'Cape of Eternal Justice', 'Ring of Eternal Justice',
    'Hammer of the Gathering Storm', 'Cloak of the Gathering Storm', 'Ring of the Gathering Storm',
    'Sickle of Unyielding Strength', 'Drape of Unyielding Strength', 'Signet of Unyielding Strength',
    # AQ40 tier sets
    'Enigma Robes', 'Enigma Circlet', 'Enigma Leggings', 'Enigma Shoulderpads', 'Enigma Boots',
    'Vestments of the Oracle', 'Tiara of the Oracle', 'Trousers of the Oracle', 'Mantle of the Oracle',
    "Footwraps of the Oracle",
    "Doomcaller's Robes", "Doomcaller's Circlet", "Doomcaller's Trousers", "Doomcaller's Mantle",
    "Doomcaller's Footwraps",
    "Deathdealer's Vest", "Deathdealer's Helm", "Deathdealer's Leggings", "Deathdealer's Spaulders",
    "Deathdealer's Boots",
    'Genesis Vest', 'Genesis Helm', 'Genesis Trousers', 'Genesis Shoulderpads', 'Genesis Boots',
    "Stormcaller's Hauberk", "Stormcaller's Diadem", "Stormcaller's Leggings", "Stormcaller's Pauldrons",
    "Stormcaller's Footguards",
    "Striker's Hauberk", "Striker's Diadem", "Striker's Leggings", "Striker's Pauldrons", "Striker's Footguards",
    "Conqueror's Breastplate", "Conqueror's Crown", "Conqueror's Legguards", "Conqueror's Spaulders",
    "Conqueror's Greaves",
    "Avenger's Breastplate", "Avenger's Crown", "Avenger's Legguards", "Avenger's Pauldrons", "Avenger's Greaves",
    # T3 (Naxxramas) tier sets
    'Frostfire Circlet', 'Frostfire Shoulderpads', 'Frostfire Robe', 'Frostfire Bindings', 'Frostfire Gloves',
    'Frostfire Belt', 'Frostfire Leggings', 'Frostfire Sandals', 'Frostfire Ring',
    'Plagueheart Circlet', 'Plagueheart Shoulderpads', 'Plagueheart Robe', 'Plagueheart Bindings',
    'Plagueheart Gloves', 'Plagueheart Belt', 'Plagueheart Leggings', 'Plagueheart Sandals', 'Plagueheart Ring',
    'Circlet of Faith', 'Shoulderpads of Faith', 'Robe of Faith', 'Bindings of Faith', 'Gloves of Faith',
    'Belt of Faith', 'Leggings of Faith', 'Sandals of Faith', 'Ring of Faith',
    'Bonescythe Helmet', 'Bonescythe Pauldrons', 'Bonescythe Breastplate', 'Bonescythe Bracers',
    'Bonescythe Gauntlets', 'Bonescythe Waistguard', 'Bonescythe Legplates', 'Bonescythe Sabatons', 'Bonescythe Ring',
    'Dreamwalker Headpiece', 'Dreamwalker Spaulders', 'Dreamwalker Tunic', 'Dreamwalker Wristguards',
    'Dreamwalker Handguards', 'Dreamwalker Girdle', 'Dreamwalker Legguards', 'Dreamwalker Boots',
    'Ring of the Dreamwalker',
    'Cryptstalker Headpiece', 'Cryptstalker Spaulders', 'Cryptstalker Tunic', 'Cryptstalker Wristguards',
    'Cryptstalker Handguards', 'Cryptstalker Girdle', 'Cryptstalker Legguards', 'Cryptstalker Boots',
    'Ring of the Cryptstalker',
    'Earthshatter Headpiece', 'Earthshatter Spaulders', 'Earthshatter Tunic', 'Earthshatter Wristguards',
    'Earthshatter Handguards', 'Earthshatter Girdle', 'Earthshatter Legguards', 'Earthshatter Boots',
    'Ring of the Earthshatterer',
    'Redemption Headpiece', 'Redemption Spaulders', 'Redemption Tunic', 'Redemption Wristguards',
    'Redemption Handguards', 'Redemption Girdle', 'Redemption Legguards', 'Redemption Boots', 'Ring of Redemption',
    'Dreadnaught Helmet', 'Dreadnaught Pauldrons', 'Dreadnaught Breastplate', 'Dreadnaught Bracers',
    'Dreadnaught Gauntlets', 'Dreadnaught Waistguard', 'Dreadnaught Legplates', 'Dreadnaught Sabatons',
    'Ring of the Dreadnaught',
    # Naxxramas frost resistance crafts (patterns / materials only in Naxxramas)
    'Glacial Cloak', 'Glacial Gloves', 'Glacial Leggings', 'Glacial Vest', 'Glacial Wrists',
    'Icebane Bracers', 'Icebane Breastplate', 'Icebane Gauntlets', 'Icebane Leggings',
    'Icy Scale Bracers', 'Icy Scale Breastplate', 'Icy Scale Gauntlets', 'Icy Scale Leggings',
    'Polar Bracers', 'Polar Gloves', 'Polar Leggings', 'Polar Tunic',
    "Ramaladni's Icy Grasp",  # quest reward that needs a Frozen Rune (Naxxramas only)
    # Atiesh: its splinters drop only in Naxxramas (AtlasLoot also lists it outside Naxxramas)
    'Atiesh, Greatstaff of the Guardian', 'Splinter of Atiesh',
}

PROFESSIONS = {1: "Alchemy", 2: "Blacksmithing", 3: "Enchanting", 4: "Engineering", 5: "Herbalism",
               8: "Leatherworking", 9: "Mining", 10: "Skinning", 11: "Tailoring"}
REP_LEVELS = {1: "Hated", 2: "Hostile", 3: "Unfriendly", 4: "Neutral", 5: "Friendly", 6: "Honored", 7: "Revered",
              8: "Exalted"}
FACTIONS = {21: "Booty Bay", 47: "Ironforge", 54: "Gnomeregan Exiles", 59: "Thorium Brotherhood", 68: "Undercity",
            69: "Darnassus", 72: "Stormwind", 76: "Orgrimmar", 81: "Thunder Bluff", 87: "Bloodsail Buccaneers",
            92: "Gelkis Clan Centaur", 93: "Magram Clan Centaur", 270: "Zandalar Tribe", 349: "Ravenholdt",
            369: "Gadgetzan", 470: "Ratchet", 509: "The League of Arathor", 510: "The Defilers",
            529: "Argent Dawn", 530: "Darkspear Trolls", 576: "Timbermaw Hold", 577: "Everlook",
            589: "Wintersaber Trainers", 609: "Cenarion Circle", 729: "Frostwolf Clan", 730: "Stormpike Guard",
            749: "Hydraxian Waterlords", 809: "Shen'dralar", 889: "Warsong Outriders", 890: "Silverwing Sentinels",
            909: "Darkmoon Faire", 910: "Brood of Nozdormu"}
TYPE_LABEL = {"raid": "Raid", "dungeon": "Dungeon", "world boss": "World boss", "reputation": "Reputation",
              "crafting": "Crafted", "pvp": "PvP", "quest": "Quest", "world drop": "World drop", "drop": "Drop",
              "vendor": "Vendor", "container": "Container", "event": "Event", "set": "Set", "rare mob": "Rare mob",
              "other": "Other"}
FROM_LABEL = {"lokiy": "Lokiy", "external": "another player's loot log", "atlasloot": "AtlasLoot", "workbook": "item workbook", "tracker": "Lokiy's loot tracker",
              "pfquest": "classic data (pfQuest)", "wowhead": "from Wowhead",
              "questcache": "Vanilla+ quest cache"}


def _load(name, default):
    path = os.path.join(HERE, name)
    if not os.path.exists(path):
        return default
    with open(path, encoding="utf-8") as f:
        return json.load(f)


def set_info(item):
    """(set name, its bonus piece counts) from the tooltip: "Judgement Armor (0/8)", "(2) Set: ..." -> ("Judgement
    Armor", (2, 4, 6, 8)); (None, ()) for an item outside a set."""
    name, bonuses = None, set()
    for ln in item.get("lines", []):
        m = re.match(r"^(.+?) \(\d+/\d+\)$", ln)
        if m:
            name = m.group(1)
        m = re.match(r"^\((\d+)\) Set:", ln)
        if m:
            bonuses.add(int(m.group(1)))
    return name, tuple(sorted(bonuses))


def add_shared_world_boss_loot(tables, names, table_groups, items):
    """The shared world boss loot table (Lokiy 2026-10-25): the AQ class books (SHARED_WB_TABLES) and the glove
    formulas (SHARED_WB_ITEMS) move out of their AtlasLoot tables into SHARED_WB_KEY, in the World Bosses group."""
    rows = []
    for key in SHARED_WB_TABLES:
        rows.append((0, "", "Class books", False))
        rows += [r for r in tables.get(key, []) if r[0]]
    known = {r[0]: r for rows_ in tables.values() for r in rows_ if r[0]}  # AtlasLoot's row (with its icon)
    rows.append((0, "", "Enchanting formulas", False))
    rows += [known.get(i, (i, "inv_misc_note_01", name, False)) for i, name in sorted(SHARED_WB_ITEMS.items())]
    moved = {r[0] for r in rows if r[0]}
    for key in list(tables):
        tables[key] = [r for r in tables[key] if r[0] not in moved]
        if not any(r[0] for r in tables[key]):
            del tables[key]
    tables[SHARED_WB_KEY] = rows
    names[SHARED_WB_KEY] = (SHARED_WB_NAME, "World Bosses")
    table_groups[SHARED_WB_KEY] = "WorldBosses"
    return moved


# ---- crafted / reputation sources from AtlasLoot's tables for api/sources.json (Lokiy 2026-10-28, Part HB), instead of
# Wowhead classic's. In UIItemSource shape: {"crafted": {"profession", "spellId"}}, {"rep": {"repFactionId",
# "repLevel", "playerFaction"?}}. PvP rank items keep Wowhead's vendor: AtlasLoot only has the rank (#rank12#) and
# UIItemSource has no field for it.
CRAFT_PROF = [("Alchemy", 1), ("Smithing", 2), ("Armorsmith", 2), ("Weaponsmith", 2), ("Axesmith", 2),
              ("Hammersmith", 2), ("Swordsmith", 2), ("Enchanting", 3), ("Engineering", 4), ("Gnomish", 4),
              ("Goblin", 4), ("Leather", 8), ("Dragonscale", 8), ("Elemental", 8), ("Tribal", 8), ("Tailoring", 11),
              ("Smelting", 9)]  # table key prefix -> profession (Herbalism / Mining list gathered herbs and ores)
PROF_CODE = {"#p1#": 1, "#p2#": 2, "#p4#": 3, "#p5#": 4, "#p7#": 8, "#p8#": 11, "#p9#": 8, "#p10#": 8, "#p11#": 8,
             "#p14#": 4, "#p15#": 4, "#p17#": 8}  # AtlasLoot's profession codes in a row (cooking / first aid: none)
SET_PROF = {"Blacksmithing": 2, "Leatherworking": 8, "Tailoring": 11}  # "Crafted Sets - <profession>" tables
# Factions table key (without its number) -> faction id; Argent1 (token hand-ins) and Darkmoon (tickets) are no standing
FACTION_TABLES = {"Argent": 529, "Timbermaw": 576, "Thorium": 59, "Cenarion": 609, "Zandalar": 270,
                  "Wintersaber": 589, "Stormpike": 730, "Frostwolf": 729, "WaterLords": 749, "Bloodsail": 87,
                  "Defilers": 510, "GelkisClan": 92, "LeagueofArathor": 509, "MagramClan": 93, "Shendralar": 809,
                  "AQBroodRings": 910}
NO_STANDING_TABLES = {"Argent1", "Darkmoon"}
# battleground reward tables (PvP folder): a header names the side's faction (#b1# Stormpike Guard, ...) ->
# (faction id, player faction 1 Alliance / 2 Horde)
BG_REP = ("WSGRep", "ABRep", "AVRep")
BG_SIDE = {"#b1#": (730, 1), "#b2#": (729, 2), "#b3#": (890, 1), "#b4#": (889, 2), "#b5#": (509, 1), "#b6#": (510, 2)}
STANDING_CODE = {"#r1#": 4, "#r2#": 5, "#r3#": 6, "#r4#": 7, "#r5#": 8}
STANDING_WORD = {"Neutral": 4, "Friendly": 5, "Honored": 6, "Revered": 7, "Exalted": 8}


def atlas_item_sources(tables, groups, names, row_info, craft):
    """{item id: [UIItemSource]} from AtlasLoot's Crafting, Factions and battleground tables. tables / groups / names as
    in gen_site, row_info {key: [(rest of the row, recipe spell)] per row}, craft {spell: made item}."""
    item_spell = {}
    for sp, iid in sorted(craft.items()):
        item_spell.setdefault(iid, sp)
    out = {}

    def add(iid, e):
        if e not in out.setdefault(iid, []):
            out[iid].append(e)

    for key, rows in tables.items():
        group, (name, inst) = groups.get(key, ""), names.get(key, ("", ""))
        info = row_info.get(key, [])
        if group == "Crafting":
            prof = next((p for pre, p in CRAFT_PROF if key.startswith(pre)), None)
            if inst.startswith("Crafted Sets - "):
                prof = SET_PROF.get(inst[len("Crafted Sets - "):])
            for r, (rest, spell) in zip(rows, info):
                if not r[0]:
                    continue
                p = next((v for c, v in PROF_CODE.items() if c in rest), prof)
                spell = spell or item_spell.get(r[0])
                if p:
                    add(r[0], {"crafted": {"profession": p, **({"spellId": spell} if spell else {})}})
            continue
        base = re.sub(r"\d+$", "", key)
        if group == "Factions" and base in FACTION_TABLES and key not in NO_STANDING_TABLES:
            side = (FACTION_TABLES[base], 0)
        elif group == "PvP" and key.startswith(BG_REP):
            side = None  # from the side's header
        else:
            continue
        words = [STANDING_WORD[w] for w in re.findall(r"[A-Z][a-z]+", name + " " + key) if w in STANDING_WORD]
        level = words[0] if len(set(words)) == 1 else None
        for r, (rest, _) in zip(rows, info):
            if not r[0]:
                level = next((v for c, v in STANDING_CODE.items() if c == r[2].strip()), level)
                side = next((v for c, v in BG_SIDE.items() if c in rest), side)
                continue
            if level and side:
                fid, pf = side
                add(r[0], {"rep": {"repFactionId": fid, "repLevel": level, **({"playerFaction": pf} if pf else {})}})
    return out


def merge_atlas_sources(wowhead, atlas):
    """api/sources.json's "wowhead" {item id (str): [UIItemSource]} with AtlasLoot's crafted / rep sources in place of
    Wowhead's of the same kind (its quest / drop / vendor sources stay). Returns (merged, {"<kind> replaced" / "<kind>
    added": items})."""
    merged, moved = dict(wowhead), {}
    for iid, entries in atlas.items():
        old = wowhead.get(str(iid), [])
        kinds = {k for e in entries for k in e}
        for k in kinds:
            label = f"{k} replaced" if any(k in s for s in old) else f"{k} added"
            moved[label] = moved.get(label, 0) + 1
        merged[str(iid)] = entries + [s for s in old if not set(s) & kinds]
    return merged, moved


class Sources:
    def __init__(self, items, tables, names, table_groups, loot_drops):
        """items {id: item}, tables {key: rows}, names {key: (name, instance)}, table_groups {key: group},
        loot_drops {id: [(source, drops, looted, kills, rate)]} (loot_sync.load_loot)."""
        self.items, self.tables, self.names, self.groups = items, tables, names, table_groups
        self.loot_drops = loot_drops or {}
        self.workbook = _load("workbook_sources.json", {"sources": {}, "removed": [], "unknown": []})
        pf = _load("pfquest_sources.json", {"items": {}, "npcs": {}, "zones": {}})
        self.pf, self.npcs, self.zones = pf["items"], pf["npcs"], pf["zones"]
        self.wowhead = _load("wowhead_sources.json", {})
        self.item_tables = {}
        for key, rows in tables.items():
            for r in rows:
                if r[0] and key not in self.item_tables.setdefault(r[0], []):
                    self.item_tables[r[0]].append(key)
        self.conflicts = []  # (item id, reason): marked by a rule, but the loot tracker saw it drop
        self.mark_rule = {}  # item id -> which rule marked it (shown on /removed, to trace a wrong mark)
        self.not_on_server = self._not_on_server()

    # ---- table types
    def table_type(self, key):
        name, inst = self.names.get(key, ("", "Other"))
        group = self.groups.get(key, "")
        if group == "Instances":
            return "raid" if inst in RAIDS else "dungeon" if inst not in ("Other", "Keys", "Miscellaneous") else "other"
        if group == "Sets":
            return "raid" if inst.startswith(("Tier 1", "Tier 2", "Tier 3", "Ruins", "Temple", "Zul'Gurub")) else "set"
        return GROUP_TYPE.get(group, "other")

    def cut_table(self, key):
        """Why this loot table is not on Vanilla+ (a CUT_TEXT key), or None."""
        name, inst = self.names.get(key, ("", "Other"))
        if key in AV_TABLES:
            return "AV"
        if key in BROOD_TABLES:
            return "Brood"
        if inst in CUT_INSTANCES and name != KURINNAXX:
            return CUT_INSTANCES[inst]
        return None

    # ---- not on Vanilla+
    def _not_on_server(self):
        reasons, rule = {}, self.mark_rule

        def mark(iid, text, why, force=False):
            if force or iid not in reasons:
                reasons[iid], rule[iid] = text, why

        by_name = {}
        for iid, it in self.items.items():
            by_name.setdefault(it["name"], []).append(iid)
        for iid in self.items:
            keys = self.item_tables.get(iid, [])
            cuts = [self.cut_table(k) for k in keys]
            if keys and all(cuts):
                mark(iid, CUT_TEXT[cuts[0]], "AtlasLoot table: " + ", ".join(keys), force=True)
        for name in AQ_T3_REMOVE_NAMES:
            for iid in by_name.get(name, []):
                mark(iid, CUT_TEXT["AQ/Naxx"], "Names list (workbook AQ_T3_REMOVE_NAMES)")
        for row in self.workbook["removed"]:
            name, src, note = (row + [None, None, None])[:3]
            name = re.sub(r"\s*\[\d+\]$", "", name or "")
            if name in REMOVED_SHEET_KEEP:
                continue
            ids = by_name.get(name, [])
            if len(ids) > 1:  # a classic item and the server's own of the same name: the classic one is meant
                ids = [i for i in ids if i <= CLASSIC_MAX_ITEM_ID]
            if any(i > CLASSIC_MAX_ITEM_ID for i in ids):
                continue  # only a server item has the name: the row can't mean it (reported to Lokiy)
            text = ("Alterac Valley is not on Vanilla+." if "Alterac" in (src or "") else
                    CUT_TEXT["Brood"] if "Nozdormu" in (src or "") else
                    CUT_TEXT["old SM"] if "Scarlet" in (src or "") or (src or "").startswith("SM") else
                    "Removed from Vanilla+.")
            for iid in ids:
                mark(iid, text, f"Workbook Removed Items sheet: {row[0]} ({src})")
        for iid, text in FORCE_IDS.items():
            if iid in self.items:
                mark(iid, text, "Fixed list (workbook FORCE_NOVELTY_IDS)", force=True)
        for name in LOKIY_NOT_ON_SERVER:
            for iid in by_name.get(name, []):
                mark(iid, "Not on Vanilla+ (Lokiy).", "Lokiy's answer, 2026-10-25")
        for iid in LOKIY_NOT_ON_SERVER_IDS:
            if iid in self.items:
                mark(iid, "Not on Vanilla+ (Lokiy).", "Lokiy's answer, 2026-10-28", force=True)
        for iid in AV_ONLY_IDS:
            if iid in self.items:
                mark(iid, CUT_TEXT["AV"], "Alterac Valley is its only source (Lokiy 2026-10-30)", force=True)
        for iid, it in self.items.items():
            if iid < OLD_PVP_MAX_ID and it["quality"] == 3 and OLD_PVP_SET.match(it["name"]):
                mark(iid, CUT_TEXT["old pvp"], "Old PvP rank set (name, rare, id < 18000), Lokiy 2026-10-28")
        for iid in LOKIY_UNOBTAINABLE_IDS:
            if iid in self.items:
                mark(iid, "Can't be obtained (Lokiy 2026-10-28).", "Lokiy's answer, 2026-10-28", force=True)
        for iid in LOKIY_PROBABLY_NOT_IDS:
            if iid in self.items:
                mark(iid, "Lokiy is 99% sure it isn't on Vanilla+.", "Lokiy's answer, 2026-10-28", force=True)
        # a set that exists twice: the version with 2/4/6/8-piece bonuses is Vanilla+'s, the 3/5/8 one the old one
        sets = {}
        for iid, it in self.items.items():
            name, bonuses = set_info(it)
            if name:
                sets.setdefault(name, {}).setdefault(bonuses, []).append(iid)
        # (also under another name: classic Cenarion Raiment vs Vanilla+'s Cenarion Armor, matched by the first word)
        new_words = {n.split()[-1] if n.startswith("The ") else n.split()[0]
                     for n, versions in sets.items() if any(b[:3] == (2, 4, 6) for b in versions)}
        for name, versions in sets.items():
            word = name.split()[-1] if name.startswith("The ") else name.split()[0]
            if word in new_words or (" of " in name and name.split(" of ")[-1] in new_words):
                for iid in versions.get((3, 5, 8), []):
                    mark(iid, CUT_TEXT["old set"], f"Set-bonus rule: {name} with 3/5/8-piece bonuses")
        for iid, it in self.items.items():
            level = next((int(m.group(1)) for ln in it.get("lines", []) for m in [re.match(r"Requires Level (\d+)", ln)]
                          if m), 0)
            if level >= 60 or it.get("quality", 0) < 2:
                continue
            keys = set(self.item_tables.get(iid, []))
            wh = self.wowhead.get(str(iid), [])
            if keys & OLD_SM_TABLES:
                mark(iid, CUT_TEXT["old SM"], "Old Scarlet Monastery rule (Lokiy 2026-10-25): " +
                     ", ".join(sorted(keys & OLD_SM_TABLES)))
            elif not keys and str(iid) not in self.workbook["sources"] and wh and all(x.get("drop", {}).get("zoneId") == SM_ZONE for x in wh):
                mark(iid, CUT_TEXT["old SM"], "Old Scarlet Monastery rule (Lokiy 2026-10-25): Wowhead drops only in SM")
        # Wowhead-only items whose every Wowhead source is cut content (no table, workbook or tracker source)
        for sid, srcs in self.wowhead.items():
            iid = int(sid)
            if iid in reasons or iid not in self.items or self.item_tables.get(iid) or \
                    str(iid) in self.workbook["sources"] or self.loot_drops.get(iid) or not srcs:
                continue
            why = []
            for s in srcs:
                d, r = s.get("drop"), s.get("rep")
                if d and d.get("zoneId") in WH_CUT_ZONES and d.get("npcId") != KURINNAXX_NPC:
                    why.append(WH_CUT_ZONES[d["zoneId"]])
                elif r and r.get("repFactionId") in WH_CUT_REPS:
                    why.append(WH_CUT_REPS[r["repFactionId"]])
                else:
                    why = []
                    break
            if why:
                mark(iid, CUT_TEXT[why[0]], "Wowhead sources only, all in " + ", ".join(sorted(set(why))), force=True)
        # item id -> (text, type, from, href): LOKIY_ON_SERVER / the shared world boss loot / OUTSIDE_SM
        self.lokiy_source = {}
        for name, (text, kind) in LOKIY_ON_SERVER.items():
            for iid in by_name.get(name, []):
                self.lokiy_source[iid] = (text, kind, "lokiy", None)
        for r in self.tables.get(SHARED_WB_KEY, []):
            if r[0]:
                self.lokiy_source[r[0]] = (SHARED_WB_TEXT, "world boss", "lokiy", f"/loot/{SHARED_WB_KEY}")
        for iid, text in OUTSIDE_SM.items():
            self.lokiy_source[iid] = (text, "drop", "wowhead", None)
        for iid in self.lokiy_source:
            reasons.pop(iid, None)
            rule.pop(iid, None)
        for iid in list(reasons):
            if self.loot_drops.get(iid):
                self.conflicts.append((iid, reasons.pop(iid)))
                rule.pop(iid, None)
        return reasons

    # ---- sources
    def sources(self, iid):
        """[{"text", "type", "from", "href"?}] for this item, from the first layer that has any (type: ITEM_TYPE first)."""
        out = self._sources(iid)
        for npc, mob, zone, n, log in EXTERNAL_DROPS.get(iid, []):
            out.append({"text": f"Drops from {mob} ({zone}; {log}, seen {n}x)", "type": "drop", "from": "external",
                        "href": f"/npc/{npc}"})
        if iid in ITEM_TYPE:
            out = [{**e, "type": ITEM_TYPE[iid]} for e in out]
        return out

    def _sources(self, iid):
        if iid in self.lokiy_source:
            text, kind, frm, href = self.lokiy_source[iid]
            return [{"text": text, "type": kind, "from": frm, **({"href": href} if href else {})}]
        keys = self.item_tables.get(iid, [])
        wb = self.workbook["sources"].get(str(iid))
        if keys and not (wb and all(self.groups.get(k) == "Sets" for k in keys)):
            out = []
            for k in keys:
                name, inst = self.names.get(k, (k, "Other"))
                where = f" ({inst})" if inst and inst != "Other" and inst not in name else ""
                out.append({"text": (name or k) + where, "type": self.table_type(k), "from": "atlasloot",
                            "href": f"/loot/{k}"})
            if iid in ALSO_WORLD_DROP and self.pf.get(str(iid)):
                out += [e for e in self._pfquest(self.pf[str(iid)]) if e["type"] == "world drop"]
            return out
        if wb:
            return [{"text": part.strip(), "type": wb["type"], "from": "workbook"} for part in wb["text"].split(" / ")]
        if self.loot_drops.get(iid):
            return [{"text": f"Drops from {src}", "type": "drop", "from": "tracker"}
                    for src, *_ in self.loot_drops[iid][:5]]
        pf = self.pf.get(str(iid))
        if pf:
            return self._pfquest(pf)
        wh = self.wowhead.get(str(iid))
        if wh:
            return [s for s in (self._wowhead(s) for s in wh) if s]
        return []

    def _pfquest(self, pf):
        out = []
        n, drops = pf.get("n", 0), pf.get("drops", [])
        many = n > 10
        for npc, name, zone, chance, lvl in drops[:3 if many else 5]:
            where = ", ".join(x for x in (zone, f"level {lvl}" if lvl else "") if x)
            out.append({"text": f"{name}{f' ({where})' if where else ''}: {chance:g}%", "type": "drop",
                        "from": "pfquest", "href": f"https://www.wowhead.com/classic/npc={npc}"})
        if many:
            lv = pf.get("lvl")
            out.append({"text": f"World drop: {n} kinds of mobs{f', level {lv[0]}-{lv[1]}' if lv else ''}",
                        "type": "world drop", "from": "pfquest"})
        for npc, name, zone in pf.get("vendors", [])[:3]:
            out.append({"text": f"Sold by {name}{f' ({zone})' if zone else ''}", "type": "vendor", "from": "pfquest",
                        "href": f"https://www.wowhead.com/classic/npc={npc}"})
        for obj, name, zone in pf.get("objects", [])[:3]:
            out.append({"text": f"In {name}{f' ({zone})' if zone else ''}", "type": "container", "from": "pfquest",
                        "href": f"https://www.wowhead.com/classic/object={obj}"})
        return out

    def _zone(self, zid):
        return self.zones.get(str(zid), "") if zid else ""

    def _wowhead(self, s):
        if "drop" in s:
            d = s["drop"]
            npc = self.npcs.get(str(d.get("npcId")), [None])[0] if d.get("npcId") else None
            who = npc or d.get("otherName") or ""
            zone = self._zone(d.get("zoneId"))
            if not who and not zone:
                return None  # an NPC pfQuest doesn't know (Season of Discovery) in no zone: nothing to show
            text = " - ".join(x for x in (zone, who) if x) or "Drop"
            if d.get("category") and d["category"] != "Original Loot":
                text += f" ({d['category']})"
            e = {"text": text, "type": "drop", "from": "wowhead"}
            if d.get("npcId"):
                e["href"] = f"https://www.wowhead.com/classic/npc={d['npcId']}"
            return e
        if "quest" in s:
            q = s["quest"]
            return {"text": f"Quest: {q.get('name') or q.get('id')}", "type": "quest", "from": "wowhead",
                    "href": f"https://www.wowhead.com/classic/quest={q['id']}"}
        if "crafted" in s:
            c = s["crafted"]
            e = {"text": PROFESSIONS.get(c.get("profession"), "Crafted"), "type": "crafting", "from": "wowhead"}
            if c.get("spellId"):
                e["href"] = f"https://www.wowhead.com/classic/spell={c['spellId']}"
            return e
        if "soldBy" in s:
            v = s["soldBy"]
            zone = self._zone(v.get("zoneId"))
            name = v.get("npcName") or "a vendor"
            return {"text": f"Sold by {name}{f' ({zone})' if zone else ''}", "type": "vendor", "from": "wowhead"}
        if "rep" in s:
            r = s["rep"]
            fac = FACTIONS.get(r.get("repFactionId"), f"Faction {r.get('repFactionId')}")
            return {"text": f"{fac} - {REP_LEVELS.get(r.get('repLevel'), '')}".rstrip(" -"), "type": "reputation",
                    "from": "wowhead"}
        return None
