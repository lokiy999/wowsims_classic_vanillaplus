"""Find classic items (same id as 1.12) whose Vanilla Plus tooltip differs from the classic one on Wowhead.

Classic tooltips: assets/db_inputs/wowhead_item_tooltips.csv (Wowhead classic tooltip HTML per item id).
Server tooltips: CSV's/VPlusItemDB.lua (in-game tooltip lines).
Both are turned into comparable text lines; lines only one side shows (item level, durability, sell price, dps,
rarity) are dropped, and number formatting is normalized.
"""
import html
import json
import os
import re
import sys

REPO = os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
WOWHEAD = os.path.join(REPO, "assets", "db_inputs", "wowhead_item_tooltips.csv")

SKIP = re.compile(r"^(Item Level \d+|Durability \d+ / \d+|Sell Price:.*|\(.*damage per second\)|Rarity: \w+|"
                  r"Max Stack: \d+|Phase \d+|Disenchants into:.*|<Random enchantment>|<Right Click to \w+>|Already known|"
                  r"Duration: .*|Mount|Unique \(\d+\)|Dropped by: .*|Drop Chance: .*|Requires .*Riding \(\d+\)|\d+ Charges?|Quest Item)$")


def wowhead_lines(tooltip):
    """Wowhead tooltip HTML -> text lines in the in-game tooltip format."""
    s = re.sub(r"\s*\n\s*", "", tooltip)
    s = re.sub(r'<div class="q0 indent">.*?</div>', "", s)  # set piece names (the game tooltip has none)
    s = re.sub(r"<!--.*?-->", "", s, flags=re.S)
    # two-column rows: "Two-Hand | Mace" -> "Two-Hand, Mace"; "1 - 3 Damage | Speed 1.90" -> "1 - 3 Damage, Speed 1.90"
    s = re.sub(r"</td>\s*<th[^>]*>", ", ", s)
    s = re.sub(r"<br\s*/?>|</?table[^>]*>|</tr>|<div[^>]*>|</div>", "\n", s)
    s = re.sub(r"<[^>]+>", "", s)
    s = html.unescape(s)
    return [ln.strip() for ln in s.split("\n") if ln.strip()]


def norm(ln):
    ln = ln.replace("\u00a0", " ").strip()
    ln = re.sub(r"\s+", " ", ln)
    ln = re.sub(r"\s*\((\d+ (Sec|Min|Hour|Day)s? Cooldown|Proc chance: [\d.]+%)\)\.?$", "", ln)  # Wowhead-only suffixes
    ln = re.sub(r"(\d+)\.\d+ (health|mana)\b", r"\1 \2", ln)  # food: "61.2 health" -> "61 health"
    ln = re.sub(r"(\d)\.0+\b", r"\1", ln)  # 2.00 -> 2
    ln = re.sub(r"(\d\.\d*?)0+\b", r"\1", ln)  # 1.90 -> 1.9
    ln = ln.rstrip(".").rstrip(",").strip()
    ln = re.sub(r"^\((\d+)\) Set ?: ", r"(\1) Set: ", ln)
    ln = re.sub(r"^Ranged, (Wand|Gun|Bow|Crossbow)$", r"\1", ln)
    return ln


def comparable(lines):
    return [norm(ln) for ln in lines if not SKIP.match(norm(ln))]


def load_wowhead():
    out = {}
    with open(WOWHEAD, encoding="utf-8") as f:
        for raw in f:
            head, _, body = raw.partition(",")
            try:
                iid = int(head)
                d = json.loads(body)
            except (ValueError, json.JSONDecodeError):
                continue
            out[iid] = {"name": d.get("name", ""), "quality": d.get("quality"), "icon": d.get("icon", ""),
                        "lines": wowhead_lines(d.get("tooltip", ""))}
    return out


RECIPE = ("Pattern:", "Plans:", "Recipe:", "Formula:", "Schematic:", "Manual:")


def comparable_set(lines, names, own):
    """Lines that matter for "was this item changed": no set piece names (the game lists the pieces you lack, Wowhead
    all of them), no class / race / rank / reputation requirements (worded differently on each side)."""
    out = []
    for ln in comparable(lines):
        if (ln in names and ln not in own) or ln.startswith(("Classes:", "Races:")):
            continue
        if ln.startswith("Requires ") and not ln.startswith("Requires Level"):
            continue
        out.append(ln)
    return out


def find_reworked(items, max_classic_id, skip=(), wh=None):
    """{id: {"classic": Wowhead classic tooltip lines, "removed": [...], "added": [...]}} for classic item ids whose
    server tooltip differs from the classic one. Recipes are left out (they show the crafted item, listed itself).
    wh: load_wowhead() result, loaded here when not given."""
    wh = wh if wh is not None else load_wowhead()
    names = {it["name"] for it in items.values()} | {w["name"] for w in wh.values()}
    out = {}
    for iid, it in items.items():
        if iid > max_classic_id or iid in skip or iid not in wh or it["name"].startswith(RECIPE):
            continue
        classic = [wh[iid]["name"]] + wh[iid]["lines"][1:]
        server = [it["name"]] + it["lines"]
        own = {it["name"], wh[iid]["name"]}
        a = comparable_set(classic, names, own)
        b = comparable_set(server, names, own)
        if "Quest Item" in wh[iid]["lines"]:
            b = [x for x in b if x != "Binds when picked up"]  # the game shows quest items as soulbound
        removed = [x for x in a if x not in b]
        added = [x for x in b if x not in a]
        if removed or added:
            keep = [ln for ln in classic[1:] if not SKIP.match(norm(ln))]
            out[iid] = {"classic_name": wh[iid]["name"], "classic": keep, "removed": removed, "added": added}
    return out


if __name__ == "__main__":
    import collections
    sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
    from gen_site import load_items
    items = load_items()
    rw = find_reworked(items, 24283)
    print(f"{len(rw)} reworked classic items")
    c = collections.Counter()
    for v in rw.values():
        for x in v["removed"]:
            c["-" + x[:70]] += 1
        for x in v["added"]:
            c["+" + x[:70]] += 1
    for x, k in c.most_common(int(sys.argv[1]) if len(sys.argv) > 1 else 20):
        print(k, x)
    json.dump(rw, open("/tmp/reworked_diff.json", "w"), indent=0)
