#!/usr/bin/env python3
"""Read WoW 1.12's WDB/questcache.wdb (the server's answers to the client's quest queries) into plain records.

    python3 wdb_quests.py <questcache.wdb>        # print a summary and a few records

File: header "WQST" + build, locale, two more uint32 (20 bytes), then records: quest id (uint32), record size
(uint32), record. A record is SMSG_QUEST_QUERY_RESPONSE as the client keeps it (checked against Wowhead's classic
rewards, 2026-10-30): 39 uint32s, title, objectives, details, end text (0-terminated), 4 x (creature or object id,
count, item id, item count), 4 objective texts. A creature id with the high bit set is a game object
(-id). Only quests the client has queried are in the file: quest givers, the quest log, quest links.
"""
import struct
import sys

MAGIC = b"TSQW"  # "WQST" reversed
HEAD = 20
N_INTS = 39
# the uint32s at the start of a record that we use (the rest: unknown or 0)
F_ID, F_METHOD, F_LEVEL, F_ZONE, F_TYPE, F_REP_FACTION, F_REP_VALUE = 0, 1, 2, 3, 4, 5, 6
F_NEXT, F_MONEY, F_MONEY_MAX, F_SPELL, F_SRC_ITEM, F_FLAGS = 9, 10, 11, 12, 13, 14
F_ITEMS, F_CHOICES = 15, 23  # 4 x (item, count), 6 x (item, count)


def _cstr(data, pos):
    end = data.index(b"\x00", pos)
    return data[pos:end].decode("utf-8", errors="replace"), end + 1


def parse_record(data):
    """One record -> {"id", "level", "zone", "type", "flags", "money", "moneyMax", "spell", "srcItem", "next",
    "repFaction", "repValue", "items": [[id, count]], "choices": [[id, count]], "title", "objectives", "details",
    "end", "needs": [[creature (or -object), count, item, count]], "goals": [texts]} (empty fields left out)."""
    v = struct.unpack_from(f"<{N_INTS}I", data, 0)
    signed = struct.unpack_from(f"<{N_INTS}i", data, 0)
    rec = {"id": v[F_ID], "method": v[F_METHOD], "level": signed[F_LEVEL], "zone": signed[F_ZONE], "type": v[F_TYPE],
           "repFaction": v[F_REP_FACTION], "repValue": signed[F_REP_VALUE], "next": v[F_NEXT],
           "money": signed[F_MONEY], "moneyMax": v[F_MONEY_MAX], "spell": v[F_SPELL], "srcItem": v[F_SRC_ITEM],
           "flags": v[F_FLAGS],
           "items": [[v[F_ITEMS + 2 * i], v[F_ITEMS + 2 * i + 1]] for i in range(4) if v[F_ITEMS + 2 * i]],
           "choices": [[v[F_CHOICES + 2 * i], v[F_CHOICES + 2 * i + 1]] for i in range(6) if v[F_CHOICES + 2 * i]]}
    pos = N_INTS * 4
    for key in ("title", "objectives", "details", "end"):
        rec[key], pos = _cstr(data, pos)
    needs = []
    for _ in range(4):
        cre, ccount, item, icount = struct.unpack_from("<iIII", data, pos)
        pos += 16
        if cre or item:
            needs.append([cre if cre >= 0 else -(cre & 0x7FFFFFFF), ccount, item, icount])
    rec["needs"] = needs
    goals = []
    for _ in range(4):
        if pos >= len(data):
            break
        t, pos = _cstr(data, pos)
        goals.append(t)
    rec["goals"] = [g for g in goals if g]
    return {k: x for k, x in rec.items() if x not in (0, "", [], None) or k in ("id", "level")}


def parse(path):
    """{quest id: record} from a questcache.wdb; raises ValueError for anything else."""
    data = open(path, "rb").read()
    if data[:4] != MAGIC or len(data) < HEAD:
        raise ValueError(f"{path}: not a questcache.wdb")
    out, pos = {}, HEAD
    while pos + 8 <= len(data):
        qid, size = struct.unpack_from("<II", data, pos)
        if size == 0:
            break
        body = data[pos + 8:pos + 8 + size]
        pos += 8 + size
        if len(body) < N_INTS * 4:
            continue
        rec = parse_record(body)
        if rec["id"] == qid:
            out[qid] = rec
    return out


if __name__ == "__main__":
    qs = parse(sys.argv[1])
    print(len(qs), "quests;", sum(1 for q in qs.values() if q.get("items") or q.get("choices")), "with reward items")
    for q in list(qs.values())[:3]:
        print(q)
