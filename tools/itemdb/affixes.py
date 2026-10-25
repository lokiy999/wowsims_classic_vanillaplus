"""Random suffixes ("of the Eagle", ...) from the server's own data (db.lokiy.dev; sim repo Part GC, 2026-10-25).

- Names and values: the Patch 3 client DBCs in dbc/ (copied from the hub's data\\dbc): ItemRandomProperties.dbc
  (suffix id, internal name, 5 SpellItemEnchantment ids, display name) and SpellItemEnchantment.dbc (field 13: the
  text, "+12 Stamina"). They replace Wowhead's randomEnchant table for the values. 37 suffixes are new on Vanilla+
  (ids 2165-2201, no display name, internal labels like "dag", "m fire"): kept, but only shown once a drop has one.
- Which suffixes an item can roll is server-side (not in the client): Wowhead's per-item lists for classic items,
  plus every suffix Lokiy's loot tracker saw (loot.json recentDrops[].suffix), pooled per item cache group
  (~/itemcache/items.json item.randomProperty: the items of one group roll from the same list).
- Chances stay Wowhead's (affix_chances.json); our own counts are shown next to them ("seen N×").
"""
import json
import os
import struct

HERE = os.path.dirname(os.path.abspath(__file__))
DBC_DIR = os.path.join(HERE, "dbc")
ITEMCACHE = os.path.expanduser("~/itemcache/items.json")
LOOT_JSON = "/var/www/lokiy/loot/loot.json"
NEW_ON_VPLUS = range(2165, 2202)  # suffixes stock 1.12 doesn't have


def read_dbc(name):
    """[(int fields), string block] of a WDBC file."""
    with open(os.path.join(DBC_DIR, name), "rb") as f:
        b = f.read()
    magic, n, fields, size, ssize = struct.unpack("<4s4I", b[:20])
    assert magic == b"WDBC" and size == fields * 4, name
    recs = [struct.unpack_from(f"<{fields}I", b, 20 + i * size) for i in range(n)]
    return recs, b[20 + n * size:]


def _str(block, off):
    end = block.index(b"\0", off)
    return block[off:end].decode("utf-8", "replace")


def load_suffixes():
    """{suffix id: {"name": display name ("" for the V+-only ones), "label": internal name, "effects": [texts]}}."""
    ench_recs, ench_s = read_dbc("SpellItemEnchantment.dbc")
    ench = {r[0]: _str(ench_s, r[13]) for r in ench_recs}
    recs, strs = read_dbc("ItemRandomProperties.dbc")
    out = {}
    for r in recs:
        effects = [ench[e] for e in r[2:7] if e and ench.get(e)]
        out[r[0]] = {"name": _str(strs, r[7]), "label": _str(strs, r[1]), "effects": effects}
    return out


def load_groups():
    """{item id: randomProperty group} for the items the item cache saw with random suffixes."""
    try:
        cache = json.load(open(ITEMCACHE, encoding="utf-8"))
    except (OSError, ValueError):
        return {}
    return {int(i): v["item"]["randomProperty"] for i, v in cache.get("items", {}).items()
            if isinstance(v, dict) and v.get("item", {}).get("randomProperty")}


def load_seen_suffixes():
    """{item id: {suffix id: count}} from the loot tracker's drops that saved the suffix id."""
    try:
        d = json.load(open(LOOT_JSON, encoding="utf-8"))
    except (OSError, ValueError):
        return {}
    out = {}
    for e in d.get("recentDrops", []):
        if e.get("item") and e.get("suffix"):
            c = out.setdefault(int(e["item"]), {})
            c[int(e["suffix"])] = c.get(int(e["suffix"]), 0) + int(e.get("quantity") or 1)
    return out


class Affixes:
    """Every item's possible suffixes: {item id: {suffix id: (seen on this item, seen in its group)}}."""

    def __init__(self, wowhead_ids):
        """wowhead_ids: {item id: [suffix ids]} (Wowhead's per-item lists, loot_sync.load_affix_ids)."""
        self.suffixes = load_suffixes()
        self.groups = load_groups()
        self.seen = load_seen_suffixes()
        members = {}
        for iid, g in self.groups.items():
            members.setdefault(g, []).append(iid)
        group_seen = {}
        for iid, counts in self.seen.items():
            g = self.groups.get(iid)
            key = ("g", g) if g else ("i", iid)
            for sid, n in counts.items():
                group_seen.setdefault(key, {})
                group_seen[key][sid] = group_seen[key].get(sid, 0) + n
        self.items = {}
        for iid in set(wowhead_ids) | set(self.seen) | set(self.groups):
            g = self.groups.get(iid)
            pooled = group_seen.get(("g", g) if g else ("i", iid), {})
            ids = set(wowhead_ids.get(iid, [])) | set(pooled)
            if not ids:
                continue
            self.items[iid] = {sid: (self.seen.get(iid, {}).get(sid, 0), pooled.get(sid, 0)) for sid in ids
                               if sid in self.suffixes}
        self.wowhead = {iid: set(v) for iid, v in wowhead_ids.items()}

    def rows(self, iid):
        """[(suffix id, name, effects, seen on item, seen in group, from Wowhead)] to show for the item: the V+-only
        suffixes without a display name only once seen."""
        out = []
        for sid, (n_item, n_group) in sorted(self.items.get(iid, {}).items()):
            s = self.suffixes[sid]
            if not s["name"] and not n_group:
                continue
            name = s["name"] or f'"{s["label"]}" (new on Vanilla+)'
            out.append((sid, name, s["effects"], n_item, n_group, sid in self.wowhead.get(iid, ())))
        return out
