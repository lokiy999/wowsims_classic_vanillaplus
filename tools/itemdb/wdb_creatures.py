#!/usr/bin/env python3
"""Read WoW 1.12's WDB/creaturecache.wdb (the server's answers to the client's creature queries) into plain records
(sim repo Part HX).

    python3 wdb_creatures.py <creaturecache.wdb>     # print a summary and a few records

File: header "BOMW" ("WMOB" reversed) + build, locale, two more uint32 (20 bytes), then records: creature id (uint32),
record size (uint32), record. A record is SMSG_CREATURE_QUERY_RESPONSE as the client keeps it: name and three more
name forms (0-terminated), subname ("<Accessories Quartermaster>"), then type flags, creature type, family, rank,
an unknown uint32, pet spell data, display id (uint32 each) and two bytes (civilian, racial leader). No level and
no zone: the client never gets those from this query. Only creatures the client has met are in the file.
"""
import struct
import sys

MAGIC = b"BOMW"
HEAD = 20
TYPES = {1: "Beast", 2: "Dragonkin", 3: "Demon", 4: "Elemental", 5: "Giant", 6: "Undead", 7: "Humanoid",
         8: "Critter", 9: "Mechanical", 10: "Not specified", 11: "Totem"}
RANKS = {0: "Normal", 1: "Elite", 2: "Rare elite", 3: "Boss", 4: "Rare"}


def parse_record(data):
    """One record -> {"name", "subname"?, "type", "family"?, "rank", "flags"?, "display"} (empty fields left out)."""
    pos, names = 0, []
    for _ in range(5):
        end = data.index(b"\x00", pos)
        names.append(data[pos:end].decode("utf-8", errors="replace"))
        pos = end + 1
    flags, ctype, family, rank, _unk, _pet, display = struct.unpack_from("<7I", data, pos)
    rec = {"name": names[0], "subname": names[4], "type": ctype, "family": family, "rank": rank, "flags": flags,
           "display": display}
    return {k: v for k, v in rec.items() if v not in (0, "") or k in ("name", "type", "rank")}


def parse(path):
    """{creature id: record} from a creaturecache.wdb; raises ValueError for anything else."""
    data = open(path, "rb").read()
    if data[:4] != MAGIC or len(data) < HEAD:
        raise ValueError(f"{path}: not a creaturecache.wdb")
    out, pos = {}, HEAD
    while pos + 8 <= len(data):
        cid, size = struct.unpack_from("<II", data, pos)
        if size == 0:
            break
        body = data[pos + 8:pos + 8 + size]
        pos += 8 + size
        if len(body) < size:
            break  # a file WoW was still writing: the cut-off last record is left out
        try:
            out[cid] = parse_record(body)
        except (ValueError, struct.error):
            continue
    return out


if __name__ == "__main__":
    cs = parse(sys.argv[1])
    print(len(cs), "creatures;", sum(1 for c in cs if c >= 20000), "with Vanilla+ ids (>= 20000)")
    for cid in list(cs)[:3]:
        print(cid, cs[cid])
