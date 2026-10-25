"""One-time import of the old Item Database workbook (Lokiy's Item Template.xlsx, PC: projects\\_deprecated\\Item
Database, see its HISTORY.md) into workbook_sources.json. Run once (2026-10-25); db.lokiy.dev is the source of truth
after that and the workbook is no longer maintained.

    python3 import_workbook.py "Item Template.xlsx"

Reads with the standard library only (an .xlsx is zipped XML):
  - Items: the Source column (one or more sources joined by " / ") and its source type, which the workbook keeps as
    the cell's fill colour (write_all.py COLORS). Raid and world boss share a colour: a source that starts with a
    raid name is a raid, a known world boss a world boss, else a rare mob or a chest. A few cells have another fill
    (Zandalar reputation, events, Sacks of Spoils): typed by their text.
  - Removed Items (78 rows: name, source, note) and Unknown Source (item IDs) as they are.
Writes workbook_sources.json next to this script.
"""
import json
import os
import re
import sys
import zipfile
import xml.etree.ElementTree as ET

HERE = os.path.dirname(os.path.abspath(__file__))
NS = {"m": "http://schemas.openxmlformats.org/spreadsheetml/2006/main",
      "r": "http://schemas.openxmlformats.org/officeDocument/2006/relationships"}
FILL_TYPE = {"FFA335EE": "raid", "FF627998": "dungeon", "FF5FE65D": "reputation", "FFE6CC80": "crafting",
             "FFFF8000": "pvp", "FF0070FF": "world drop", "FFFFFF00": "quest"}
RAID_PREFIXES = ("Molten Core", "Blackwing Lair", "Zul'Gurub", "Onyxia", "MC/BWL", "Ony/BWL", "Scarlet Monastery",
                 "Ruins of Ahn'Qiraj", "Temple of Ahn'Qiraj", "Naxxramas", "Tier ")
# the rest of the raid colour: world bosses, else rare spawns (Lo'Grosh, ...) and chests
WORLD_BOSSES = ("Kurinnaxx", "Azuregos", "Lord Kazzak", "Emerald Dragons", "Lady Hederine", "Taerar", "Ysondre",
                "Emeriss", "Lethon", "Prince Skaldrenox", "Lord Skwol", "Baron Kazum", "High Marshal Whirlaxis")


def col_idx(ref):
    n = 0
    for ch in re.match(r"[A-Z]+", ref).group(0):
        n = n * 26 + ord(ch) - 64
    return n - 1


def read_sheets(path):
    """{sheet name: [[(value, fill rgb), ...], ...]}"""
    z = zipfile.ZipFile(path)
    shared = [("".join(t.text or "" for t in si.iter("{%s}t" % NS["m"])))
              for si in ET.fromstring(z.read("xl/sharedStrings.xml")).findall("m:si", NS)]
    styles = ET.fromstring(z.read("xl/styles.xml"))
    fills = []
    for f in styles.find("m:fills", NS):
        fg = f.find("m:patternFill/m:fgColor", NS)
        fills.append(fg.get("rgb") if fg is not None else None)
    xf_fill = [fills[int(xf.get("fillId", 0))] for xf in styles.find("m:cellXfs", NS)]
    rels = {r.get("Id"): r.get("Target") for r in ET.fromstring(z.read("xl/_rels/workbook.xml.rels"))}
    res = {}
    for sh in ET.fromstring(z.read("xl/workbook.xml")).find("m:sheets", NS):
        target = rels[sh.get("{%s}id" % NS["r"])].lstrip("/")
        target = target if target.startswith("xl/") else "xl/" + target
        rows = []
        for row in ET.fromstring(z.read(target)).iter("{%s}row" % NS["m"]):
            vals = {}
            for c in row.findall("m:c", NS):
                v, t = c.find("m:v", NS), c.get("t")
                if t == "s" and v is not None:
                    val = shared[int(v.text)]
                elif t == "inlineStr":
                    val = "".join(x.text or "" for x in c.iter("{%s}t" % NS["m"]))
                elif v is not None:
                    val = v.text
                else:
                    continue
                vals[col_idx(c.get("r"))] = (val, xf_fill[int(c.get("s", 0))])
            if vals:
                rows.append([vals.get(i, (None, None)) for i in range(max(vals) + 1)])
        res[sh.get("name")] = rows
    return res


def main(path):
    sheets = read_sheets(path)
    items = sheets["Items"]
    head = [c[0] for c in items[0]]
    i_id, i_src = head.index("ID"), head.index("Source")
    sources, untyped = {}, 0
    for r in items[1:]:
        if len(r) <= i_src or not r[i_src][0] or not r[i_id][0]:
            continue
        text, fill = r[i_src]
        kind = FILL_TYPE.get(fill)
        if kind == "raid" and not text.startswith(RAID_PREFIXES):
            kind = ("world boss" if text.startswith(WORLD_BOSSES) else "container" if "Chest" in text
                    else "other" if text == "DEPRECATED" else "rare mob")
        elif kind is None:  # cells with another fill
            kind = ("reputation" if "reputation" in text else "container" if "of Spoils" in text
                    else "event" if re.search(r"Event|Darkmoon|Invasion", text) else "other")
        untyped += kind is None
        sources[str(int(float(r[i_id][0])))] = {"text": text.strip(), "type": kind}
    removed = [[c[0] for c in r[:3]] for r in sheets["Removed Items"][1:]]
    unknown = [int(float(r[0][0])) for r in sheets["Unknown Source"][1:] if r and r[0][0]]
    out = {"about": "One-time import of Item Template.xlsx (2026-10-25, import_workbook.py). Not maintained.",
           "sources": sources, "removed": removed, "unknown": unknown}
    with open(os.path.join(HERE, "workbook_sources.json"), "w", encoding="utf-8") as f:
        json.dump(out, f, ensure_ascii=False, indent=0, sort_keys=True)
    types = {}
    for s in sources.values():
        types[s["type"]] = types.get(s["type"], 0) + 1
    print(f"{len(sources)} sources ({untyped} without a type), {len(removed)} removed rows, {len(unknown)} unknown;",
          types)


if __name__ == "__main__":
    main(sys.argv[1])
