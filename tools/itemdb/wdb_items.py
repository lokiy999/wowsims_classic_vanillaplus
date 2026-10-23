#!/usr/bin/env python3
"""Read a WoW 1.12 itemcache.wdb (the client's cache of every item it has seen).

    python3 wdb_items.py itemcache.wdb [more.wdb ...] [-o items.json]

Each record is the server's full item_template row as the client got it (SMSG_ITEM_QUERY_SINGLE_RESPONSE),
so it holds what the item really is on the server now: stats, damage, spells, set, required level...
Prints a short summary, or writes {id: record} as JSON with -o.
"""

import argparse
import json
import struct
import sys

HEADER = 20  # "BDIW", build, locale, record version, unknown


class Reader:
    def __init__(self, buf, pos, end):
        self.buf, self.pos, self.end = buf, pos, end

    def i(self):
        v, = struct.unpack_from("<i", self.buf, self.pos)
        self.pos += 4
        return v

    def u(self):
        v, = struct.unpack_from("<I", self.buf, self.pos)
        self.pos += 4
        return v

    def f(self):
        v, = struct.unpack_from("<f", self.buf, self.pos)
        self.pos += 4
        return round(v, 4)

    def s(self):
        n = self.buf.index(b"\0", self.pos, self.end)
        v = self.buf[self.pos:n].decode("utf-8", errors="replace")
        self.pos = n + 1
        return v


def parse_record(r):
    it = {"class": r.u(), "subclass": r.u()}
    names = [r.s() for _ in range(4)]
    it["name"] = names[0]
    for k in ("displayId", "quality", "flags", "buyPrice", "sellPrice", "inventoryType", "allowableClass",
              "allowableRace", "itemLevel", "requiredLevel", "requiredSkill", "requiredSkillRank", "requiredSpell",
              "requiredHonorRank", "requiredCityRank", "requiredRepFaction", "requiredRepValue", "maxCount",
              "stackable", "containerSlots"):
        it[k] = r.i()
    stats = [(r.u(), r.i()) for _ in range(10)]
    it["stats"] = [[t, v] for t, v in stats if v]
    dmg = [(r.f(), r.f(), r.u()) for _ in range(5)]
    it["damage"] = [[lo, hi, school] for lo, hi, school in dmg if lo or hi]
    it["armor"] = r.i()
    it["resist"] = [r.i() for _ in range(6)]  # holy, fire, nature, frost, shadow, arcane
    it["delay"] = r.u()
    it["ammoType"] = r.u()
    it["rangedModRange"] = r.f()
    spells = []
    for _ in range(5):
        sid, trigger, charges, cd, cat, catcd = r.u(), r.u(), r.i(), r.i(), r.u(), r.i()
        if sid:
            spells.append({"id": sid, "trigger": trigger, "charges": charges, "cooldown": cd,
                           "category": cat, "categoryCooldown": catcd})
    it["spells"] = spells
    it["bonding"] = r.u()
    it["description"] = r.s()
    for k in ("pageText", "language", "pageMaterial", "startQuest", "lockId", "material", "sheath",
              "randomProperty", "block", "itemSet", "maxDurability", "area", "map", "bagFamily"):
        it[k] = r.i()
    return it


def read_wdb(path):
    """{item id: record}. Records the parser can't place exactly are skipped and counted."""
    with open(path, "rb") as fh:
        buf = fh.read()
    if buf[:4] != b"BDIW":
        raise ValueError(f"{path}: not an itemcache.wdb (magic {buf[:4]!r})")
    build, = struct.unpack_from("<I", buf, 4)
    items, bad, pos = {}, [], HEADER
    while pos + 8 <= len(buf):
        iid, size = struct.unpack_from("<II", buf, pos)
        pos += 8
        if iid == 0 and size == 0:
            break  # end marker
        end = pos + size
        try:
            r = Reader(buf, pos, end)
            it = parse_record(r)
            if r.pos != end:
                raise ValueError(f"read {r.pos - pos} of {size} bytes")
            items[iid] = it
        except (ValueError, struct.error) as e:
            bad.append((iid, str(e)))
        pos = end
    return {"build": build, "items": items, "bad": bad}


def main():
    ap = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    ap.add_argument("wdb", nargs="+")
    ap.add_argument("-o", "--output")
    args = ap.parse_args()
    items = {}
    for path in args.wdb:
        res = read_wdb(path)
        print(f"{path}: build {res['build']}, {len(res['items'])} items, {len(res['bad'])} unreadable")
        for iid, err in res["bad"][:5]:
            print(f"  item {iid}: {err}")
        items.update(res["items"])
    if args.output:
        with open(args.output, "w", encoding="utf-8") as fh:
            json.dump({str(k): v for k, v in sorted(items.items())}, fh, ensure_ascii=False, separators=(",", ":"))
        print(f"Wrote {args.output}: {len(items)} items")


if __name__ == "__main__":
    sys.exit(main())
