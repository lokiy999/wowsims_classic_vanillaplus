"""Object pages for db.lokiy.dev (sim repo Part HU): /object/<id> for every chest, ore vein, herb, ... that holds an
item, and /objects (a list by name).

Data: pfquest_objects.json (import_pfquest.py; classic vmangos data from pfQuest, not checked against Vanilla+):
{object id: {"name", "zones": [area ids by spawn count], "spawns", "items": [[item id, chance %]]}}. Plus, as its
own layer, what Lokiy's loot tracker saw in opened objects (lokiy.dev/loot loot.json "chests", kind "object"): the
addon only knows the object's name, so it is matched by name, and by zone when several objects share the name.

Left out: items that aren't in the database (not on Vanilla+), objects that then hold nothing, and objects that only
spawn in zones that aren't on Vanilla+.
"""
import html
import json
import os

HERE = os.path.dirname(os.path.abspath(__file__))
SHOWN = 80  # item rows per object page; the rest is counted
NOTE = ('<p class="muted small">Classic 1.12 data (pfQuest / vmangos), not checked against Vanilla+: the chance is per '
        'opening; an object that shares a loot pool (chests) lists the whole pool.</p>')


def esc(s):
    return html.escape(str(s), quote=True)


class Objects:
    def __init__(self, items, zone_names, skip_zones, chests=()):
        """items: the database's items; zone_names {area id: name} (zones with a page or not); skip_zones: area ids
        that aren't on Vanilla+; chests: loot.json "chests"."""
        path = os.path.join(HERE, "pfquest_objects.json")
        raw = json.load(open(path, encoding="utf-8")) if os.path.exists(path) else {}
        self.zone_names = zone_names
        self.objs = {}
        for k, v in raw.items():
            held = [[i, c] for i, c in v["items"] if i in items]
            zones = [z for z in v.get("zones", []) if z not in skip_zones]
            if not held or (v.get("zones") and not zones):
                continue
            self.objs[int(k)] = {"name": v["name"], "zones": zones, "items": held}
        # the loot tracker's opened objects, by name (and zone when several objects have that name)
        by_name = {}
        for oid, o in self.objs.items():
            by_name.setdefault(o["name"], []).append(oid)
        self.seen, self.unmatched = {}, []
        for c in chests:
            if c.get("kind") != "object" or not c.get("items"):
                continue
            ids = by_name.get(c["name"], [])
            if len(ids) > 1:
                in_zone = [i for i in ids if {zone_names.get(z) for z in self.objs[i]["zones"]} & set(c.get("zones", {}))]
                ids = in_zone or ids
            if not ids:
                self.unmatched.append(c["name"])
            for i in ids:
                self.seen[i] = c

    def href(self, oid):
        return f"/object/{oid}" if oid in self.objs else None

    def write(self, out, page, item_cell, zone_pages):
        """Writes /object/<id>.html and /objects.html. zone_pages: area ids that have a /zone page. Returns the count."""
        os.makedirs(os.path.join(out, "object"), exist_ok=True)
        for f in os.listdir(os.path.join(out, "object")):
            os.remove(os.path.join(out, "object", f))

        def zone_link(z):
            name = self.zone_names.get(z, f"Zone {z}")
            return f'<a href="/zone/{z}">{esc(name)}</a>' if z in zone_pages else esc(name)

        for oid, o in self.objs.items():
            zones = ", ".join(zone_link(z) for z in o["zones"][:6]) + (" ..." if len(o["zones"]) > 6 else "")
            parts = []
            c = self.seen.get(oid)
            if c:
                rows = "".join(f'<tr><td>{item_cell(it["id"])}</td><td class="rate">{it["drops"]}</td>'
                               f'<td class="rate">{it["drops"] / c["opened"]:.0%}</td></tr>'
                               for it in sorted(c["items"], key=lambda it: -it["drops"]) if it.get("id"))
                parts.append(f'<section class="card"><h2>Seen by Lokiy\'s loot tracker</h2><p class="muted">Opened '
                             f'{c["opened"]} times ({esc(", ".join(c.get("zones", {})))}).</p><table class="obs-table">'
                             f'<thead><tr><th>Item</th><th>Drops</th><th>Per opening</th></tr></thead><tbody>{rows}'
                             f'</tbody></table><p class="muted small">From <a href="https://lokiy.dev/loot/">Lokiy\'s '
                             f'loot tracker</a>, matched by the object\'s name.</p></section>')
            rows = "".join(f'<tr><td>{item_cell(i)}</td><td class="rate">{f"{ch:g}%" if ch else "?"}</td></tr>'
                           for i, ch in o["items"][:SHOWN])
            more = f'<p class="muted small">{len(o["items"]) - SHOWN} more with lower chances.</p>' \
                if len(o["items"]) > SHOWN else ""
            parts.append(f'<section class="card"><h2>Contents</h2><table class="obs-table"><thead><tr><th>Item</th>'
                         f'<th>Chance</th></tr></thead><tbody>{rows}</tbody></table>{more}{NOTE}</section>')
            body = f"""
<article>
  <div class="crumbs"><a href="/objects">Objects</a></div>
  <h1>{esc(o['name'])}</h1>
  <p class="muted">{f'{zones} &middot; ' if zones else ''}Object {oid} &middot;
  <a href="https://www.wowhead.com/classic/object={oid}" rel="noopener">Wowhead</a></p>
  {''.join(parts)}
</article>"""
            with open(os.path.join(out, "object", f"{oid}.html"), "w", encoding="utf-8") as f:
                f.write(page(o["name"], body, path=f"/object/{oid}",
                             description=f"{o['name']}: what it holds on Vanilla Plus."))
        # the list, by name
        rows = []
        tag = ' <span class="tag">loot tracker</span>'
        for oid, o in sorted(self.objs.items(), key=lambda kv: (kv[1]["name"], kv[0])):
            zone = zone_link(o["zones"][0]) if o["zones"] else '<span class="muted">-</span>'
            rows.append(f'<tr><td><a href="/object/{oid}">{esc(o["name"])}</a>'
                        f'{tag if oid in self.seen else ""}</td>'
                        f'<td>{oid}</td><td>{zone}</td><td class="rate">{len(o["items"])}</td></tr>')
        body = f"""
<section class="intro"><h1>Objects</h1>
<p class="muted">{len(self.objs)} chests, ore veins, herbs and other objects that hold items, from pfQuest's classic data
({len(self.seen)} also seen by Lokiy's loot tracker). Objects only in content that isn't on Vanilla+ are left out.</p></section>
<section class="card"><table class="obs-table removed-table"><thead><tr><th>Object</th><th>ID</th><th>Zone</th>
<th>Items</th></tr></thead><tbody>{''.join(rows)}</tbody></table></section>"""
        with open(os.path.join(out, "objects.html"), "w", encoding="utf-8") as f:
            f.write(page("Objects", body, path="/objects", description=f"{len(self.objs)} objects on Vanilla Plus."))
        return len(self.objs)
