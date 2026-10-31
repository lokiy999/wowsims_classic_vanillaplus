#!/usr/bin/env python3
"""Item level model for db.lokiy.dev's /ilvl page (sim repo Part HY; Lokiy 2026-10-31).

    python3 ilvl_model.py --fit      # fit on classic items, 5-fold held-out check -> ilvl_fit.json (commit it)

gen_site.py only applies the saved fit (report(); the page is built in gen_site.py), so builds stay fast.

Training data: classic items, quality 2-4, required level 10+, with stats or weapon damage (random-suffix items show
none). Blizzard's own classic item data first (assets/db_inputs/wago_db2_items.csv, Classic Era: item level,
inventory slot, required level, primary stats, armor, resistances, weapon damage / speed, item set); Wowhead's
classic tooltips (wowhead_item_tooltips.csv via gen_site.load_wowhead) for what that file lacks: the Equip lines
(spell damage, crit, mp5, ...), Use / proc effects, the armor type, and items the file doesn't have.

Model: stat value = (sum over stats of (amount x weight) ^ 1.5) ^ (2/3), with bonus armor (armor above a fitted base
for armor type / slot / quality) as a stat, a set-piece factor and extra spell schools at a fraction; divided by the
slot modifier; a straight budget line per quality turns it into an item level. Weapons also get an item level from
their DPS (linear per weapon group and quality), blended 70 / 30 with the stats; a caster / healer weapon (its own
slot modifier and budget line) takes the higher of the two, since in 1.12 it spends its budget on spell power or on
DPS. Weapon DPS is
1.12's (the tooltip) where Blizzard's Classic Era file has a later, raised value (59 weapons, mostly caster ones). Items with Use / proc / unpriced
Equip effects can't be priced fully: their prediction is a lower bound.
"""
import collections
import csv
import json
import os
import random
import re
import sys

HERE = os.path.dirname(os.path.abspath(__file__))
REPO = os.path.dirname(os.path.dirname(HERE))
WAGO = os.path.join(REPO, "assets/db_inputs/wago_db2_items.csv")
FIT = os.path.join(HERE, "ilvl_fit.json")
CLASSIC_MAX_ITEM_ID = 24283
QUALS = (2, 3, 4)
# raid tiers in progression order (Lokiy 2026-10-31); an entry may list several instances ("|")
TIERS = ["Zul'Gurub", "Molten Core|Onyxia's Lair", "Blackwing Lair", "Scarlet Monastery"]
BUDGET_FLAG = 5  # levels between the real and the model item level that flag an item over / under budget
# tiers whose item levels were set by hand (Lokiy 2026-10-31): the report suggests all of them from the stats; the
# other tiers are classic-like anchors and keep theirs
CALC_TIERS = {"Scarlet Monastery"}

# ---- tooltip parsing (classic Wowhead lines and the server's lines read the same)
SLOT_GROUP = {"Head": "A", "Chest": "A", "Legs": "A", "Robe": "A", "Shoulder": "B", "Hands": "B", "Waist": "B",
              "Feet": "B", "Wrist": "C", "Neck": "C", "Back": "C", "Finger": "C", "Held In Off-hand": "C",
              "Off Hand": "S", "One-Hand": "W1", "Main Hand": "W1", "Two-Hand": "W2", "Ranged": "WR",
              "Thrown": "WR", "Relic": "R", "Trinket": "T"}
GROUP_LABEL = {"A": "Head / chest / legs", "B": "Shoulder / hands / waist / feet",
               "C": "Wrist / neck / back / finger / held", "S": "Shield", "W1": "One-hand weapon",
               "W2": "Two-hand weapon", "WR": "Ranged weapon", "WAND": "Wand"}
# wago InventoryType -> tooltip slot name
INV_SLOT = {1: "Head", 2: "Neck", 3: "Shoulder", 5: "Chest", 20: "Robe", 6: "Waist", 7: "Legs", 8: "Feet", 9: "Wrist",
            10: "Hands", 11: "Finger", 12: "Trinket", 13: "One-Hand", 14: "Off Hand", 15: "Ranged", 16: "Back",
            17: "Two-Hand", 21: "Main Hand", 22: "Off Hand", 23: "Held In Off-hand", 25: "Thrown", 26: "Ranged",
            28: "Relic"}
ARMOR_TYPES = {"Cloth", "Leather", "Mail", "Plate"}
PRIM = {"Strength", "Agility", "Stamina", "Intellect", "Spirit"}
RES = {"Fire Resistance", "Nature Resistance", "Frost Resistance", "Shadow Resistance", "Arcane Resistance"}
EQ = [  # (regex, stat key); the first that matches an Equip line wins
    (r"damage and healing done by magical spells and effects by up to (\d+)", "sd"),
    (r"healing done by spells and effects by up to (\d+)", "heal"),
    (r"critical strike with spells by (\d+)%", "scrit"),
    (r"critical strike by (\d+)%", "crit"),
    (r"hit with spells by (\d+)%", "shit"),
    (r"chance to hit by (\d+)%", "hit"),
    (r"Restores (\d+) mana per 5 sec", "mp5"),
    (r"Restores (\d+) health per 5 sec", "hp5"),
    (r"^Equip: \+(\d+) Attack Power\.$", "ap"),
    (r"^Equip: \+(\d+) ranged Attack Power\.$", "rap"),
    (r"\+(\d+) All Resistances", "allres"),
    (r"Increased Defense \+(\d+)", "def"),
    (r"chance to dodge an attack by (\d+)%", "dodge"),
    (r"chance to parry an attack by (\d+)%", "parry"),
    (r"chance to block attacks with a shield by (\d+)%", "block"),
    (r"block value of your shield by (\d+)", "bv"),
    (r"damage done by (?:Shadow|Fire|Frost|Arcane|Nature|Holy) spells and effects by up to (\d+)", "school"),
    (r"Increased (?:Daggers|Swords|Axes|Maces|Two-handed Swords|Two-handed Axes|Two-handed Maces|Bows|Guns|"
     r"Crossbows|Polearms|Staves|Fist Weapons) \+(\d+)", "wskill"),
    (r"Decreases the magical resistances of your spell targets by (\d+)", "spen"),
]
FERAL = re.compile(r"\+(\d+) Attack Power in Cat")
TEST_ITEM = re.compile(r"\bTest\b|^\d+ (?:Green|Blue|Epic|Purple)", re.I)  # Blizzard's test / placeholder items


def parse(lines):
    """-> dict(slot, group, atype, armor, dps, stats [(key, n)], schools [n], set, effect, ilvl, req) or None."""
    r = {"stats": [], "schools": [], "set": False, "effect": False, "armor": 0, "dps": 0, "slot": None,
         "atype": None, "ilvl": 0, "req": 0, "sblock": 0}
    for ln in lines:
        ln = ln.replace("\xa0", " ").strip()
        m = re.match(r"^Item Level (\d+)$", ln)
        if m:
            r["ilvl"] = int(m.group(1))
            continue
        m = re.match(r"^Requires Level (\d+)$", ln)
        if m:
            r["req"] = int(m.group(1))
            continue
        head = ln.split(",")[0].strip()
        if head in SLOT_GROUP and r["slot"] is None:
            r["slot"] = head
            r["atype"] = (ln.split(",", 1)[1].strip() if "," in ln else "") or None
            continue
        m = re.match(r"^(\d+) Armor$", ln)
        if m:
            r["armor"] = int(m.group(1))
            continue
        m = re.match(r"^(\d+) Block$", ln)
        if m:  # a shield's own block value: base like armor, only what is above the base counts (Model.bonus_block)
            r["sblock"] = int(m.group(1))
            continue
        m = re.match(r"^\(([\d.]+) damage per second\)$", ln)
        if m:
            r["dps"] = float(m.group(1))
            continue
        m = re.match(r"^\+ ?(\d+) (.+)$", ln)
        if m and (m.group(2) in PRIM or m.group(2) in RES):
            r["stats"].append(("prim" if m.group(2) in PRIM else "res", int(m.group(1))))
            continue
        if ln.startswith("Equip:"):
            fm = FERAL.search(ln)
            if fm:
                r["stats"].append(("feral", int(fm.group(1))))
                continue
            for rx, key in EQ:
                mm = re.search(rx, ln)
                if mm:
                    if key == "school":
                        r["schools"].append(int(mm.group(1)))
                    elif key == "allres":  # the server writes one line; classic lists the five resistances
                        r["stats"] += [("res", int(mm.group(1)))] * 5
                    else:
                        r["stats"].append((key, float(mm.group(1))))
                    break
            else:
                r["effect"] = True
        elif ln.startswith(("Use:", "Chance on hit:")):
            r["effect"] = True
        elif re.match(r"^\(\d+\) Set ?:", ln):
            r["set"] = True
    if not r["slot"]:
        return None
    return finish(r)


def finish(r):
    r["group"] = SLOT_GROUP[r["slot"]]
    if r["group"] == "S" and r["atype"] != "Shield":
        r["group"] = "W1"  # an off-hand weapon
    if r["group"] == "WR" and r["atype"] in ("Wand", "Idol", "Libram", "Totem"):
        r["group"] = "WAND" if r["atype"] == "Wand" else "R"
    return r


def role(r):
    """Who the item is for, from its stats: healer / caster / tank / melee / ranged / any."""
    keys = {k for k, _ in r["stats"]}
    if "heal" in keys:
        return "healer"
    if keys & {"sd", "scrit", "shit", "spen"} or r["schools"]:
        return "caster"
    if keys & {"def", "dodge", "parry", "block"} or r["group"] == "S":
        return "tank"
    if "rap" in keys or r["group"] == "WR":
        return "ranged"
    if keys & {"ap", "crit", "hit", "wskill", "feral"} or r["group"] in ("W1", "W2"):
        return "melee"
    return "any"


def caster(r):
    return role(r) in ("caster", "healer") and r["group"] != "WAND"


def cweapon(r):
    return r["group"] in ("W1", "W2") and caster(r)


def slot_mod(p, r):
    """The slot modifier; caster / healer weapons have their own (their budget is spell power, not DPS)."""
    return p["m_" + r["group"] + ("C" if r["group"] in ("W1", "W2") and caster(r) else "")]


# ---- the model
P0 = {"w_prim": 1.0, "w_res": 1.0, "w_sd": .86, "w_heal": .45, "w_scrit": 14, "w_crit": 14, "w_shit": 8,
      "w_hit": 10, "w_mp5": 2.5, "w_hp5": 2.5, "w_ap": .5, "w_rap": .4, "w_feral": .25, "w_def": 1.0, "w_dodge": 12,
      "w_parry": 20, "w_block": 5, "w_bv": .65, "w_school": .72, "w_wskill": 2.5, "w_spen": .8, "w_barmor": .07,
      "m_A": 1.0, "m_B": .75, "m_C": .5625, "m_S": .5625, "m_W1": .42, "m_W2": 1.0, "m_WR": .32, "m_WAND": .32,
      "m_W1C": .7, "m_W2C": 1.4, "f_set": .1, "f_school2": .3}
# the scale: a primary stat point on head / chest / legs, and a caster one-hand (caster weapons have their own budget)
FIXED = ("w_prim", "m_A", "m_W1C")


def armor_key(r):
    return "|".join(map(str, (r["atype"] if r["slot"] != "Back" else "Cloak", r["slot"], r["q"])))


def linfit(pts):
    n = len(pts)
    sx = sum(x for x, _ in pts); sy = sum(y for _, y in pts)
    sxx = sum(x * x for x, _ in pts); sxy = sum(x * y for x, y in pts)
    d = n * sxx - sx * sx
    if not n or not d:
        return 0.0, (sy / n if n else 0.0)
    a = (n * sxy - sx * sy) / d
    return a, (sy - a * sx) / n


def median(xs):
    xs = sorted(xs)
    return xs[len(xs) // 2] if xs else 0


class Model:
    """p: weights / modifiers; armor {key: (a, b)}; dps {"group|q": (a, b)}; budget {q: (a, b)}; budget_c {q: (a, b)}:
    caster / healer weapons' own budget line (in 1.12 their DPS is flat and their spell power grows faster with the
    item level than other gear's stats)."""

    def __init__(self, p=None, armor=None, dps=None, budget=None, budget_c=None, block=None):
        self.p, self.armor, self.dps, self.budget = dict(p or P0), armor or {}, dps or {}, budget or {}
        self.budget_c, self.block = budget_c or {}, block or {}

    @classmethod
    def load(cls, path=FIT):
        d = json.load(open(path, encoding="utf-8"))
        return cls(d["params"], d["armor"], d["dps"], {int(q): ab for q, ab in d["budget"].items()},
                   {int(q): ab for q, ab in d.get("budget_c", {}).items()},
                   {int(q): ab for q, ab in d.get("block", {}).items()})

    def fit_armor(self, train):
        rows = collections.defaultdict(list)
        for r in train:
            if r["armor"] and (r["atype"] in ARMOR_TYPES or r["atype"] == "Shield" or r["slot"] == "Back"):
                rows[armor_key(r)].append((r["ilvl"], r["armor"]))
        self.armor = {}
        for key, pts in rows.items():
            if len(pts) >= 6:  # robust: refit without the 20% farthest (most items carry no bonus armor)
                a, b = linfit(pts)
                res = [y - (a * x + b) for x, y in pts]
                cut = sorted(abs(e) for e in res)[int(len(res) * .8)]
                self.armor[key] = linfit([pt for pt, e in zip(pts, res) if abs(e) <= cut] or pts)

    def fit_block(self, train):
        """A shield's base block value per quality, linear in ilvl (robust, like the base armor)."""
        self.block = {}
        for q in QUALS:
            pts = [(r["ilvl"], r["sblock"]) for r in train if r["q"] == q and r.get("sblock")]
            if len(pts) >= 6:
                a, b = linfit(pts)
                res = [y - (a * x + b) for x, y in pts]
                cut = sorted(abs(e) for e in res)[int(len(res) * .8)]
                self.block[q] = linfit([pt for pt, e in zip(pts, res) if abs(e) <= cut] or pts)

    def bonus_block(self, r, ilvl):
        ab = self.block.get(r["q"]) if r.get("sblock") else None
        if not ab:
            return 0
        base = ab[0] * ilvl + ab[1]
        return r["sblock"] - base if r["sblock"] - base > .15 * base else 0

    def bonus_armor(self, r, ilvl):
        ab = self.armor.get(armor_key(r)) if r["armor"] else None
        if not ab:
            return 0
        base = ab[0] * ilvl + ab[1]
        return r["armor"] - base if r["armor"] - base > .15 * base else 0

    def fit_dps(self, train):
        rows = collections.defaultdict(list)
        for r in train:
            if r["dps"] and r["group"] in ("W1", "W2", "WR", "WAND"):
                rows[f'{r["group"]}|{r["q"]}'].append((r["ilvl"], r["dps"]))
        self.dps = {k: linfit(v) for k, v in rows.items() if len(v) >= 8}

    def value(self, r, barmor, bblock=0):
        p = self.p
        parts = [n * p["w_" + k] for k, n in r["stats"]]
        if bblock > 0:
            parts.append(bblock * p["w_bv"])
        sch = sorted(r["schools"], reverse=True)
        if sch:
            parts.append(sch[0] * p["w_school"] + sum(s * p["w_school"] * p["f_school2"] for s in sch[1:]))
        if barmor > 0:
            parts.append(barmor * p["w_barmor"])
        v = sum(x ** 1.5 for x in parts if x > 0) ** (1 / 1.5) if parts else 0.0
        return v * (1 + p["f_set"]) if r["set"] else v

    def fit_budget(self, train):
        self.budget, self.budget_c = {}, {}
        for q in QUALS:
            pts = [(r["ilvl"], self.value(r, self.bonus_armor(r, r["ilvl"]), self.bonus_block(r, r["ilvl"]))
                    / slot_mod(self.p, r))
                   for r in train if r["q"] == q and not r["effect"] and (r["stats"] or r["schools"])
                   and not cweapon(r)]
            self.budget[q] = linfit(pts)
            pts = [(r["ilvl"], self.value(r, 0) / slot_mod(self.p, r))
                   for r in train if r["q"] == q and not r["effect"] and cweapon(r)]
            if len(pts) >= 8:
                self.budget_c[q] = linfit(pts)

    def predict(self, r):
        """(ilvl from stats, ilvl from dps, combined); None where there is nothing to go on."""
        a, b = self.budget_c[r["q"]] if cweapon(r) and r["q"] in self.budget_c else self.budget[r["q"]]
        st = None
        if r["stats"] or r["schools"]:
            # bonus armor / block depend on the ilvl: the item level whose budget its value meets. Iterating diverged
            # on shields (base armor rises ~35 a level, so a weak-stat shield priced from a low guess got all its
            # armor as bonus: Aegis of the Scarlet Commander 295). st(x) - x only falls as x rises: bisect.
            mod = slot_mod(self.p, r)

            def st_at(x):
                return (self.value(r, self.bonus_armor(r, x), self.bonus_block(r, x)) / mod - b) / a
            lo, hi = -50.0, 250.0
            for _ in range(40):
                mid = (lo + hi) / 2
                lo, hi = (mid, hi) if st_at(mid) > mid else (lo, mid)
            st = (lo + hi) / 2
        ab = self.dps.get(f'{r["group"]}|{r["q"]}')
        dp = (r["dps"] - ab[1]) / ab[0] if ab and r["dps"] and ab[0] else None
        if dp is not None and st is not None:
            # a 1.12 caster / healer weapon spends its budget on spell power (low DPS) or on DPS (a little spell
            # power): whichever gives more
            return st, dp, max(st, dp) if caster(r) else .7 * dp + .3 * st
        return st, dp, dp if dp is not None else st

    def to_json(self):
        return {"params": self.p, "armor": self.armor, "dps": self.dps, "budget": self.budget, "budget_c": self.budget_c,
                "block": self.block}


# ---- fitting
def objective(model, rows):
    model.fit_budget(rows)
    tot = n = 0
    for r in rows:
        comb = model.predict(r)[2]
        if comb is not None:
            tot += min((comb - r["ilvl"]) ** 2, 100)  # capped: outliers don't drive the fit
            n += 1
    return tot / max(n, 1)


def fit(train, rounds=6, start=None, quiet=False):
    model = Model(start)
    model.fit_armor(train)
    model.fit_block(train)
    model.fit_dps(train)
    rows = [r for r in train if not r["effect"]]
    keys = [k for k in model.p if k not in FIXED]
    best = objective(model, rows)
    for rnd in range(rounds):
        for k in keys:
            for f in (1.25, .8, 1.1, .91, 1.04, .96):
                old = model.p[k]
                model.p[k] = old * f
                o = objective(model, rows)
                if o < best - 1e-6:
                    best = o
                else:
                    model.p[k] = old
        if not quiet:
            print(f"  round {rnd + 1}: mean capped squared miss {best:.2f}", flush=True)
    model.fit_budget(rows)
    return model


def wago_rows():
    """{item id: csv row dict} from Blizzard's classic item data."""
    with open(WAGO, encoding="utf-8") as f:
        return {int(r["ID"]): r for r in csv.DictReader(f)}


WAGO_STAT = {3: "Agility", 4: "Strength", 5: "Intellect", 6: "Spirit", 7: "Stamina"}


def from_wago(r, w):
    """A parsed tooltip record with Blizzard's numbers put in (item level, slot, required level, primary stats, armor,
    resistances, DPS, set); its Equip / Use lines and the armor type stay from the tooltip."""
    inv = int(w["InventoryType"])
    if inv in INV_SLOT:
        r["slot"] = INV_SLOT[inv]
    r["ilvl"] = int(w["ItemLevel"])
    r["req"] = int(w["RequiredLevel"])
    r["armor"] = int(w["Resistances_0"])
    prim = [("prim", int(w[f"StatModifier_bonusAmount_{i}"])) for i in range(10)
            if int(w[f"StatModifier_bonusStat_{i}"]) in WAGO_STAT and int(w[f"StatModifier_bonusAmount_{i}"]) > 0]
    res = [("res", int(w[f"Resistances_{i}"])) for i in range(2, 7) if int(w[f"Resistances_{i}"]) > 0]
    # Equip lines from the tooltip; an "All Resistances" line is in the file's resistances already
    r["stats"] = prim + res + [s for s in r["stats"] if s[0] not in ("prim", "res")]
    delay = int(w["ItemDelay"] or 0)
    if delay:
        dmg = sum(float(w[f"MinDamage_{i}"]) + float(w[f"MaxDamage_{i}"]) for i in range(5)) / 2
        era = round(dmg / (delay / 1000), 2) if dmg else 0
        # the file is Classic Era: Blizzard raised some weapons' DPS after 1.12 (mostly caster weapons, e.g.
        # Azuresong Mageblade 42.5 -> 52.5); Vanilla+ runs 1.12, so the tooltip's 1.12 DPS stays where they differ
        if r["dps"] and era and abs(era - r["dps"]) > .05 * r["dps"]:
            r["dps_1_12"] = True
        else:
            r["dps"] = era
    r["set"] = r["set"] or int(w["ItemSet"] or 0) > 0
    return finish(r)


def training_items():
    sys.path.insert(0, HERE)
    import gen_site as G  # noqa: E402 (the Wowhead classic tooltips)
    wh = G.load_wowhead()
    wago = wago_rows()
    train = []
    for iid, w in wh.items():
        if iid > CLASSIC_MAX_ITEM_ID or w.get("quality") not in QUALS or not w.get("lines"):
            continue
        r = parse(w["lines"])
        if not r:
            continue
        r.update(id=iid, name=w["name"], q=w["quality"], src="wowhead")
        if iid in wago and int(wago[iid]["OverallQualityID"]) == w["quality"]:
            r = from_wago(r, wago[iid])
            r["src"] = "wago"
        if not r["ilvl"] or r["group"] in ("T", "R") or r["req"] < 10 or not (r["stats"] or r["schools"] or r["dps"]):
            continue
        if G.DEV_ITEM.search(w["name"]) or TEST_ITEM.search(w["name"]):
            continue
        train.append(r)
    return train


def miss_table(model, rows):
    """{(quality, group or 'all'): [abs misses]} for rows without effects."""
    out = collections.defaultdict(list)
    for r in rows:
        comb = model.predict(r)[2]
        if comb is None or r["effect"]:
            continue
        e = abs(comb - r["ilvl"])
        out[(r["q"], r["group"])].append(e)
        out[(r["q"], "all")].append(e)
    return out


def main_fit():
    train = training_items()
    n_wago = sum(1 for r in train if r["src"] == "wago")
    n_dps = sum(1 for r in train if r.get("dps_1_12"))
    print(f"{len(train)} classic items: {n_wago} from Blizzard's item data (+ Wowhead's Equip lines), "
          f"{len(train) - n_wago} from Wowhead's tooltips only; {n_dps} weapons keep 1.12's DPS", flush=True)
    model = fit(train)
    # 5-fold held-out check: fit on four fifths (from the full fit's weights), measure the fifth
    random.seed(1)
    order = list(range(len(train)))
    random.shuffle(order)
    held = collections.defaultdict(list)
    for k in range(5):
        test = [train[i] for i in order[k::5]]
        rest = [train[i] for n, i in enumerate(order) if n % 5 != k]
        m = fit(rest, rounds=2, start=model.p, quiet=True)
        for key, e in miss_table(m, test).items():
            held[key] += e
        print(f"  fold {k + 1} done", flush=True)
    report_rows = []
    for (q, g), e in sorted(held.items(), key=lambda kv: (kv[0][0], kv[0][1] != "all", kv[0][1])):
        e.sort()
        report_rows.append({"q": q, "group": g, "n": len(e), "median": round(e[len(e) // 2], 1),
                            "p90": round(e[int(len(e) * .9)], 1)})
        print(f"  q{q} {GROUP_LABEL.get(g, 'all'):38s} n={len(e):4d}  median {e[len(e) // 2]:.1f}  "
              f"90% {e[int(len(e) * .9)]:.1f}")
    eff = collections.defaultdict(list)
    for r in train:
        comb = model.predict(r)[2]
        if r["effect"] and comb is not None:
            eff[r["q"]].append(abs(comb - r["ilvl"]))
    out = model.to_json()
    out.update(fitted="2026-10-31", items=len(train), from_wago=n_wago, from_wowhead=len(train) - n_wago,
               dps_1_12=n_dps,
               heldout=report_rows, effects={q: {"n": len(e), "median": round(median(e), 1)} for q, e in eff.items()})
    with open(FIT, "w", encoding="utf-8") as f:
        json.dump(out, f, indent=1)
    print("weights:", {k: round(v, 3) for k, v in model.p.items()})
    print("budget:", model.budget)
    print("wrote", FIT)


# ---- applying the model to Vanilla+ items (gen_site.py)
def report(items, status, instances, model=None):
    """Rows for the /ilvl page: Vanilla+ changed (status 2) and new (3) items; every raid tier item is the comparison
    pool. items: {id: {"name", "q", "ilvl", "lines": [[class, text]]}} (api/items.json shape); status {id: search
    status}; instances {id: set of instance names}. Returns (rows, tier stats)."""
    model = model or Model.load()
    tier_of = {name: n for n, t in enumerate(TIERS) for name in t.split("|")}
    recs = {}
    for iid, v in items.items():
        # weapons and armor only (item class 2 / 4): a recipe's tooltip carries the made item's stats
        if v.get("q") not in QUALS or not v.get("ilvl") or v.get("cls") not in (2, 4):
            continue
        r = parse([ln for _, ln in v["lines"]])
        if not r or r["group"] in ("T", "R"):
            continue
        r.update(id=iid, name=v["name"], q=v["q"], ilvl=v["ilvl"])
        comb = model.predict(r)[2]
        tiers = sorted({tier_of[i] for i in instances.get(iid, ()) if i in tier_of})
        recs[iid] = dict(id=iid, name=v["name"], q=v["q"], ilvl=v["ilvl"], req=r["req"], group=r["group"],
                         role=role(r), tier=tiers[0] if tiers else None, status=status.get(iid),
                         model=round(comb, 1) if comb is not None else None, effect=r["effect"], set=r["set"],
                         value=round(model.value(r, model.bonus_armor(r, r["ilvl"]), model.bonus_block(r, r["ilvl"]))
                                     / slot_mod(model.p, r), 1))
    # tier medians (real item level of the raid's blue / epic level-60 items) and the order flags
    tstat = []
    for n in range(len(TIERS)):
        il = sorted(x["ilvl"] for x in recs.values() if x["tier"] == n and x["q"] >= 3 and x["req"] >= 60)
        tstat.append({"tier": TIERS[n], "n": len(il), "median": il[len(il) // 2] if il else None,
                      "p10": il[len(il) // 10] if il else None, "p90": il[int(len(il) * .9)] if il else None})
    for n in range(1, len(tstat)):  # a whole tier below the one before breaks the order by itself
        a, b = tstat[n - 1]["median"], tstat[n]["median"]
        tstat[n]["below_prev"] = a is not None and b is not None and b < a
    # tiers whose item levels were set by hand get them from their stats, priced like the tier before them
    # (Lokiy 2026-10-31: "I guessed SM ... adjust it based on calculations"; BWL stays as is)
    for n, t in enumerate(TIERS):
        if t in CALC_TIERS and n > 0:
            prev = [x for x in recs.values() if x["tier"] == n - 1 and x["q"] >= 3 and x["req"] >= 60
                    and x["model"] is not None and not x["effect"]]
            off = median([x["ilvl"] - x["model"] for x in prev])
            recalc(recs, n, off, TIERS[n - 1].split("|")[0], tstat[n - 1]["p10"], tstat[n - 1]["p90"] + 5)
            tstat[n]["offset"] = round(off, 1)
            new = sorted(x["suggest"] for x in recs.values() if x["tier"] == n and "suggest" in x)
            if new:
                tstat[n]["suggested_median"] = new[len(new) // 2]
                tstat[n]["suggested_p10"], tstat[n]["suggested_p90"] = new[len(new) // 10], new[int(len(new) * .9)]
    rows = []
    for x in recs.values():
        if x["status"] not in (2, 3):
            continue
        flags = []
        if x["effect"]:
            flags.append("can't price")
        if x["model"] is not None and not x["effect"] and abs(x["model"] - x["ilvl"]) > BUDGET_FLAG:
            flags.append("over budget" if x["model"] > x["ilvl"] else "under budget")
        t = x["tier"]
        if "suggest" in x and TIERS[t] in CALC_TIERS:  # a recalculated tier
            if x["suggest"] != x["ilvl"]:
                flags.append("recalculated")
            if x.get("far"):
                flags.append("model far off")
                x["why"] += "; " + x["far"]
        elif t is not None and x["q"] >= 3 and x["req"] >= 60 and TIERS[t] not in CALC_TIERS:
            # outliers allowed: only below the tier before's 10th percentile or above the tier after's 90th; a
            # classic-like tier is never flagged for being above a later one (Lokiy: BWL stays as is)
            low = tstat[t - 1]["p10"] if t > 0 else None
            high = (tstat[t + 1]["p90"] if t + 1 < len(tstat) and TIERS[t + 1] not in CALC_TIERS
                    and not tstat[t + 1]["below_prev"] and TIERS[t] in CALC_TIERS else None)
            why = None
            if low is not None and x["ilvl"] < low:
                why = f"below most of {TIERS[t - 1].split('|')[0]} (10% are under {low})"
            elif high is not None and x["ilvl"] > high:
                why = f"above most of {TIERS[t + 1].split('|')[0]} (90% are at most {high})"
            if why:
                flags.append("breaks the order")
                x["why"] = why
                suggest(x, recs, low if why.startswith("below") else None, high if why.startswith("above") else None)
        x["flags"] = flags
        rows.append(x)
    return sorted(rows, key=lambda x: (x["tier"] if x["tier"] is not None else 9, x["group"], -x["ilvl"])), tstat


def recalc(recs, n, off, prev_name, lo, hi):
    """Suggested item levels for tier n from the stats: the model value + the tier before's offset (how far above
    the classic budget it is priced), kept within lo..hi (the tier before's 10th percentile to its 90th + 5: the
    raid stays in the progression; an item held at the edge is flagged, its stats are far off the classic budget).
    An item with an effect (no full price) takes its comparable items' value, at least its own lower bound."""
    tier = [x for x in recs.values() if x["tier"] == n and x["q"] >= 3 and x["req"] >= 60]
    for x in tier:
        if x["model"] is not None and not x["effect"]:
            v = round(x["model"] + off)
            x["suggest"] = int(min(max(v, lo), hi))
            x["why"] = f"stats priced like {prev_name} ({off:+.1f} levels over the classic budget)"
            if v != x["suggest"]:
                x["far"] = f"stats worth {v}: held at {x['suggest']}"
    for x in tier:
        if "suggest" in x:
            continue
        pool = [y for y in tier if "suggest" in y and y["group"] == x["group"] and y["q"] == x["q"]]
        same = [y for y in pool if y["role"] == x["role"]] or pool
        comp = sorted(same, key=lambda y: abs((y["value"] or 0) - (x["value"] or 0)))[:5]
        if len(comp) < 2:
            continue  # too little to go on: it keeps its item level
        sug = median([y["suggest"] for y in comp])
        if x["model"] is not None:  # its own (lower bound) value
            sug = max(sug, min(round(x["model"] + off), hi))
        x["suggest"] = int(sug)
        x["why"] = "has an effect the model can't price: comparable items' recalculated level (at least its own)"
        x["comp"] = [[y["id"], y["name"], y["suggest"]] for y in comp]


def suggest(x, recs, low=None, high=None):
    """Lokiy's rule: the item level of comparable items in the same raid (slot group, quality, role, similar stat
    value), a few levels higher only where its stats need it; kept inside the tier order (at least low, at most
    high), and never above the item's level when it is too high already."""
    pool = [y for y in recs.values() if y["id"] != x["id"] and y["tier"] == x["tier"] and y["group"] == x["group"]
            and y["q"] == x["q"] and y["req"] >= 60]
    same_role = [y for y in pool if y["role"] == x["role"]] or pool
    close = [y for y in same_role if x["value"] and y["value"] and abs(y["value"] / x["value"] - 1) <= .2]
    comp = sorted(close if len(close) >= 2 else same_role,
                  key=lambda y: abs((y["value"] or 0) - (x["value"] or 0)))[:5]
    if not comp:
        return
    base = median([y["ilvl"] for y in comp])
    sug = base
    if x["model"] is not None and not x["effect"] and x["model"] > base + 2:  # its stats need more than theirs
        sug = min(round(x["model"]), base + 4)
    if low is not None:  # too low for its tier: up to the tier before's range at least
        sug = max(sug, low)
    if high is not None:  # too high: down into the tier after's range, never above its level now
        sug = min(sug, high, x["ilvl"])
    x["suggest"] = int(sug)
    x["comp"] = [[y["id"], y["name"], y["ilvl"]] for y in comp]


if __name__ == "__main__":
    if "--fit" in sys.argv:
        main_fit()
    else:
        print(__doc__)
