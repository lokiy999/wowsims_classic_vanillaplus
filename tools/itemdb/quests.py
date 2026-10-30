"""db.lokiy.dev's quest and zone pages (Part GO): /quest/<id>, /quests, /zone/<id>, /zones, api/quests.json and
api/zones.json, from pfQuest's classic quests (pfquest_quests.json, written by import_pfquest.py) and the NPC pages'
data. Classic 1.12 data (vmangos), not checked against Vanilla+.

Rewards come from the item sources (Wowhead's "quest" sources carry the quest id); pfQuest has none.
"""
import json
import os
import re
from html import escape as esc

HERE = os.path.dirname(os.path.abspath(__file__))
JUNK = re.compile(r"<NYI>|<TXT>|\(123\)")
WH_QUEST = re.compile(r"wowhead\.com/classic/quest=(\d+)")
# pfQuest records that mix in an old quest of the same ID, fixed to the parts that match Wowhead (merged over pfQuest's
# record; Lokiy 2026-10-28, Parts GZ / HA). 172: start / end / event are Children's Week (Horde), name / text /
# objective / level an old quest "Ambushed In The Forest" (Ambassador Berrybuck, Ambassador's Satchel). 171: all
# Children's Week (Alliance) except the old quest's Watcher Callahan as a second start / end NPC. Found by a scan for
# start / end NPCs in different zones (40 quests, the others are quests handed out in several cities).
QUEST_FIX = {
    172: {"name": "Children's Week", "start": {"U": [14451]}, "end": {"U": [14444]}, "event": 10,
          "goal": "pfQuest's data for this quest ID mixes in an old quest (Ambushed In The Forest), so its text and "
                  "level are left out here; see Wowhead for them.", "desc": "", "needs": {}, "lvl": 0, "min": 0},
    171: {"start": {"U": [14305]}, "end": {"U": [14450]}},
}
RACES = {1: "Human", 2: "Orc", 4: "Dwarf", 8: "Night Elf", 16: "Undead", 32: "Tauren", 64: "Gnome", 128: "Troll"}
ALLIANCE, HORDE = 1 | 4 | 8 | 64, 2 | 16 | 32 | 128
CLASSES = {1: "Warrior", 2: "Paladin", 4: "Hunter", 8: "Rogue", 16: "Priest", 64: "Shaman", 128: "Mage",
           256: "Warlock", 1024: "Druid"}
SKILLS = {129: "First Aid", 164: "Blacksmithing", 165: "Leatherworking", 171: "Alchemy", 182: "Herbalism",
          185: "Cooking", 186: "Mining", 197: "Tailoring", 202: "Engineering", 333: "Enchanting", 356: "Fishing",
          393: "Skinning"}
# zone names the loot tables / tracker use for instances pfQuest names differently
ZONE_ALIASES = {"Lower Blackrock Spire": "Blackrock Spire", "Upper Blackrock Spire": "Blackrock Spire",
                "The Sunken Temple": "Sunken Temple", "Temple of Ahn'Qiraj": "Ahn'Qiraj",
                "Dire Maul (North)": "Dire Maul", "Dire Maul (East)": "Dire Maul", "Dire Maul (West)": "Dire Maul"}
NOTE = ('<p class="muted small">Classic 1.12 quest data (pfQuest / vmangos), not checked against Vanilla+: '
        'quests the server changed or added may differ or be missing.</p>')


def side(race):
    if not race or race & 255 == 255:
        return ""
    if race & ALLIANCE and not race & HORDE:
        return "Alliance" if race & ALLIANCE == ALLIANCE else ", ".join(n for b, n in RACES.items() if race & b)
    if race & HORDE and not race & ALLIANCE:
        return "Horde" if race & HORDE == HORDE else ", ".join(n for b, n in RACES.items() if race & b)
    return ", ".join(n for b, n in RACES.items() if race & b)


def quest_text(t):
    """The game's quest text: $B line breaks, $N / $C / $R the player's name / class / race, $Gmale:female;."""
    t = esc(t)
    t = re.sub(r"\$[Gg]\s*([^:;]*):([^;]*);", r"\1/\2", t)
    t = re.sub(r"\$[Bb]", "<br>", t)
    for k, v in (("N", "name"), ("C", "class"), ("R", "race")):
        t = re.sub(r"\$" + k, f"&lt;{v}&gt;", t, flags=re.I)
    return t


class Quests:
    def __init__(self, npcs, skip_quests=(), skip_zones=()):
        """npcs: gen_site's {creature id: {"name", "zone", "zid"?, ...}}. skip_quests / skip_zones: not on Vanilla+
        (gen_site.py QUESTS_NOT_ON_SERVER / ZONES_NOT_ON_SERVER): no page, no link, not in the lists."""
        path = os.path.join(HERE, "pfquest_quests.json")
        raw = json.load(open(path, encoding="utf-8")) if os.path.exists(path) else {}
        self.quests = {int(k): v for k, v in raw.get("quests", {}).items() if not JUNK.search(v["name"])}
        for qid, fix in QUEST_FIX.items():
            self.quests[qid] = {**self.quests.get(qid, {}), **fix}
        self.skipped = {q: self.quests.pop(q)["name"] for q in skip_quests if q in self.quests}
        self.skip_zones = set(skip_zones)
        self.zones = {int(k): v for k, v in raw.get("zones", {}).items() if int(k) not in self.skip_zones}
        self.objects = {int(k): v for k, v in raw.get("objects", {}).items()}
        self.npcs = npcs
        # a zone name -> its area id (the one most NPCs use, for names pfQuest has twice)
        counts = {}
        for n in npcs.values():
            if n.get("zid"):
                counts[n["zid"]] = counts.get(n["zid"], 0) + 1
        self.zone_by_name = {}
        for zid, name in sorted(self.zones.items(), key=lambda kv: (-counts.get(kv[0], 0), kv[0])):
            self.zone_by_name.setdefault(name, zid)
        for alias, name in ZONE_ALIASES.items():
            if name in self.zone_by_name:
                self.zone_by_name[alias] = self.zone_by_name[name]
        self.leads_to, self.npc_quests, self.item_quests, self.rewards = {}, {}, {}, {}
        for qid, q in self.quests.items():
            for p in q.get("pre", []):
                self.leads_to.setdefault(p, []).append(qid)
            for kind, key in (("starts", "start"), ("ends", "end")):
                for nid in q.get(key, {}).get("U", []):
                    self.npc_quests.setdefault(nid, {}).setdefault(kind, []).append(qid)
            for iid in q.get("start", {}).get("I", []):
                self.item_quests.setdefault(iid, []).append(("Starts", qid))
            for iid in q.get("needs", {}).get("I", []):
                self.item_quests.setdefault(iid, []).append(("Needed for", qid))
            for iid in q.get("needs", {}).get("IR", []):
                self.item_quests.setdefault(iid, []).append(("Handed out by", qid))

    def add_rewards(self, item_srcs):
        """Rewards from the item sources: a "quest" source linking quest=<id>."""
        for iid, srcs in item_srcs.items():
            for e in srcs:
                m = WH_QUEST.search(e.get("href", ""))
                if m and int(m.group(1)) in self.quests:
                    self.rewards.setdefault(int(m.group(1)), []).append(iid)
                    self.item_quests.setdefault(iid, []).append(("Reward from", int(m.group(1))))

    def zone_of(self, qid):
        q = self.quests[qid]
        for key in ("start", "end"):
            for nid in q.get(key, {}).get("U", []):
                n = self.npcs.get(nid) or {}
                if n.get("zid"):
                    return n["zid"] if n["zid"] not in self.skip_zones else None
                if n.get("zone") in self.zone_by_name:
                    return self.zone_by_name[n["zone"]]
        return None

    def npc_zone(self, n):
        zid = n.get("zid") or self.zone_by_name.get(n.get("zone") or "")
        return zid if zid not in self.skip_zones else None

    def href(self, qid):
        return f"/quest/{qid}" if qid in self.quests else None

    def link(self, qid):
        q = self.quests.get(qid)
        if qid in self.skipped:  # not on Vanilla+: the name only
            return f'<span class="muted">{esc(self.skipped[qid])}</span>'
        if not q:
            return f'<a class="ext" href="https://www.wowhead.com/classic/quest={qid}" rel="noopener">Quest {qid}</a>'
        lv = f' <span class="muted">({q["lvl"]})</span>' if q.get("lvl") else ""
        return f'<a href="/quest/{qid}">{esc(q["name"])}</a>{lv}'

    def npc_link(self, nid):
        n = self.npcs.get(nid)
        return f'<a href="/npc/{nid}">{esc(n["name"])}</a>' if n else f"NPC {nid}"

    # ---- cards on other pages
    def npc_card(self, nid):
        d = self.npc_quests.get(nid)
        if not d:
            return ""
        rows = []
        for kind, label in (("starts", "Starts"), ("ends", "Ends")):
            qs = sorted(set(d.get(kind, [])), key=lambda q: (self.quests[q].get("lvl", 0), self.quests[q]["name"]))
            if qs:
                rows.append(f'<h3>{label}</h3><ul class="loot cols">{"".join(f"<li>{self.link(q)}</li>" for q in qs)}</ul>')
        return f'<section class="card"><h2>Quests</h2>{"".join(rows)}{NOTE}</section>'

    def item_card(self, iid):
        rows = self.item_quests.get(iid)
        if not rows:
            return ""
        lis = "".join(f'<li><span class="muted">{esc(label)}</span> {self.link(q)}</li>'
                      for label, q in sorted(set(rows), key=lambda r: (r[0], r[1])))
        return f'<section class="card"><h2>Quests</h2><ul class="loot">{lis}</ul>{NOTE}</section>'

    # ---- the pages
    def write(self, out, page, item_cell, npc_api, zone_tables):
        """Writes /quest/<id>, /quests, /zone/<id>, /zones and their APIs. npc_api: gen_site's {creature id: {"name",
        "level", "zone", "boss"}}; zone_tables: {zone name: [(loot table key, label)]}. Returns (quests, zones) counts."""
        for sub in ("quest", "zone"):
            os.makedirs(os.path.join(out, sub), exist_ok=True)
        q_api, q_index, zone_quests = {}, [], {}
        for qid, q in sorted(self.quests.items()):
            zid = self.zone_of(qid)
            zone = self.zones.get(zid, "")
            if zid:
                zone_quests.setdefault(zid, []).append(qid)
            sd = side(q.get("race", 0))
            cls = ", ".join(n for b, n in CLASSES.items() if q.get("class", 0) & b)
            info = " &middot; ".join(x for x in (
                f'Level {q["lvl"]}' if q.get("lvl") else "", f'Requires level {q["min"]}' if q.get("min") else "",
                esc(sd), esc(cls), esc(SKILLS.get(q.get("skill"), "")),
                f'<a href="/zone/{zid}">{esc(zone)}</a>' if zid else "") if x)

            def ends(d):
                out_ = [self.npc_link(n) for n in d.get("U", [])]
                out_ += [esc(self.objects.get(o, f"Object {o}")) for o in d.get("O", [])]
                out_ += [item_cell(i) for i in d.get("I", [])]
                return ", ".join(out_) or '<span class="muted">unknown</span>'

            parts = []
            needs = q.get("needs", {})
            goal = []
            if needs.get("I"):
                goal.append(f'<li><span class="muted">Collect</span> <ul class="loot">'
                            f'{"".join(f"<li>{item_cell(i)}</li>" for i in needs["I"])}</ul></li>')
            if needs.get("U"):
                goal.append(f'<li><span class="muted">Kill / talk to</span> {", ".join(self.npc_link(n) for n in needs["U"])}</li>')
            if needs.get("O"):
                goal.append(f'<li><span class="muted">Use</span> '
                            f'{", ".join(esc(self.objects.get(o, f"Object {o}")) for o in needs["O"])}</li>')
            parts.append(f'<section class="card"><h2>Objectives</h2><p>{quest_text(q.get("goal", ""))}</p>'
                         + (f'<ul class="quest-goal">{"".join(goal)}</ul>' if goal else "") + '</section>')
            parts.append(f'<section class="card"><h2>Start and end</h2><p><b>Start:</b> {ends(q.get("start", {}))}</p>'
                         f'<p><b>End:</b> {ends(q.get("end", {}))}</p>'
                         + (f'<p><b>Handed out:</b> {", ".join(item_cell(i) for i in needs["IR"])}</p>' if needs.get("IR") else "")
                         + '</section>')
            if self.rewards.get(qid):
                parts.append(f'<section class="card"><h2>Rewards</h2><ul class="loot">'
                             f'{"".join(f"<li>{item_cell(i)}</li>" for i in sorted(set(self.rewards[qid])))}</ul>'
                             f'<p class="muted small">Items whose source is this quest; money and XP aren\'t known.</p></section>')
            chain = []
            if q.get("pre"):
                chain.append(f'<p><b>Requires:</b> {", ".join(self.link(p) for p in q["pre"])}</p>')
            if self.leads_to.get(qid):
                chain.append(f'<p><b>Leads to:</b> {", ".join(self.link(p) for p in sorted(self.leads_to[qid]))}</p>')
            if q.get("close"):
                chain.append(f'<p><b>Can\'t be done with:</b> {", ".join(self.link(p) for p in q["close"])}</p>')
            if chain:
                parts.append(f'<section class="card"><h2>Quest chain</h2>{"".join(chain)}</section>')
            if q.get("desc"):
                parts.append(f'<section class="card"><h2>Description</h2><p>{quest_text(q["desc"])}</p></section>')
            body = f"""
<article>
  <div class="crumbs"><a href="/quests">Quests</a>{f' / <a href="/zone/{zid}">{esc(zone)}</a>' if zid else ''}</div>
  <h1>{esc(q['name'])}</h1>
  <p class="muted">{info}{' &middot; ' if info else ''}Quest {qid} &middot;
  <a href="https://www.wowhead.com/classic/quest={qid}" rel="noopener">Wowhead</a></p>
  {''.join(parts)}
  {NOTE}
</article>"""
            with open(os.path.join(out, "quest", f"{qid}.html"), "w", encoding="utf-8") as f:
                f.write(page(q["name"], body, path=f"/quest/{qid}",
                             description=f"{q['name']}" + (f", level {q['lvl']}" if q.get("lvl") else "")
                             + (f", {zone}" if zone else "") + " - quest on Vanilla Plus."))
            q_api[qid] = {"name": q["name"], "level": q.get("lvl", 0), "side": sd, "zone": zone,
                          "url": f"https://db.lokiy.dev/quest/{qid}"}
            q_index.append([qid, q["name"], q.get("lvl", 0), sd, zone])

        # zones: every zone with NPCs, quests or loot tables
        zone_npcs = {}  # by the NPC page's zone (instance NPCs have no map position in pfQuest)
        for nid, a in npc_api.items():
            zid = self.npc_zone(self.npcs.get(nid) or {}) or self.zone_by_name.get(a["zone"])
            if zid:
                zone_npcs.setdefault(zid, []).append(nid)
        tables_by_zid = {}
        for z, rows in zone_tables.items():
            if z in self.zone_by_name:
                tables_by_zid.setdefault(self.zone_by_name[z], []).extend(rows)
        z_api, z_rows = {}, []
        for zid in sorted(set(zone_npcs) | set(zone_quests) | set(tables_by_zid)):
            name = self.zones.get(zid, f"Zone {zid}")
            nids = zone_npcs.get(zid, [])
            bosses = sorted((n for n in nids if npc_api[n]["boss"]), key=lambda n: npc_api[n]["name"])
            others = sorted((n for n in nids if not npc_api[n]["boss"]), key=lambda n: npc_api[n]["name"])
            qs = sorted(zone_quests.get(zid, []), key=lambda q: (self.quests[q].get("lvl", 0), self.quests[q]["name"]))
            parts = []
            if tables_by_zid.get(zid):
                lis = "".join(f'<li><a href="/loot/{k}">{esc(label)}</a></li>' for k, label in tables_by_zid[zid])
                parts.append(f'<section class="card"><h2>Loot tables</h2><ul class="loot cols">{lis}'
                             f'</ul></section>')
            if bosses:
                parts.append(f'<section class="card"><h2>Bosses</h2><ul class="loot cols">'
                             f'{"".join(f"<li>{self.npc_link(n)}</li>" for n in bosses)}</ul></section>')
            if qs:
                parts.append(f'<section class="card"><h2>Quests <span class="muted">{len(qs)}</span></h2>'
                             f'<ul class="loot cols">{"".join(f"<li>{self.link(q)}</li>" for q in qs)}</ul>'
                             f'<p class="muted small">By where the quest starts (or ends).</p></section>')
            if others:
                lis = "".join(f'<li>{self.npc_link(n)}'
                              f'{" <span class=muted>" + esc(str(npc_api[n]["level"])) + "</span>" if npc_api[n]["level"] else ""}'
                              f'</li>' for n in others)
                parts.append(f'<section class="card"><h2>NPCs <span class="muted">{len(others)}</span></h2>'
                             f'<ul class="loot cols">{lis}</ul></section>')
            body = f"""
<article>
  <div class="crumbs"><a href="/zones">Zones</a></div>
  <h1>{esc(name)}</h1>
  <p class="muted">{len(nids)} NPCs &middot; {len(qs)} quests &middot; Zone {zid} &middot;
  <a href="https://www.wowhead.com/classic/zone={zid}" rel="noopener">Wowhead</a></p>
  {''.join(parts)}
  <p class="muted small">NPCs and quests from pfQuest's classic data (plus Vanilla+ NPCs from combat logs).</p>
</article>"""
            with open(os.path.join(out, "zone", f"{zid}.html"), "w", encoding="utf-8") as f:
                f.write(page(name, body, path=f"/zone/{zid}",
                             description=f"{name}: {len(nids)} NPCs and {len(qs)} quests on Vanilla Plus."))
            z_api[zid] = {"name": name, "npcs": len(nids), "quests": len(qs), "url": f"https://db.lokiy.dev/zone/{zid}"}
            z_rows.append((name, zid, len(nids), len(qs)))

        with open(os.path.join(out, "quests.json"), "w", encoding="utf-8") as f:
            json.dump(q_index, f, ensure_ascii=False, separators=(",", ":"))
        for name, data in (("quests", q_api), ("zones", z_api)):
            tmp = os.path.join(out, "api", f"{name}.json.tmp")
            with open(tmp, "w", encoding="utf-8") as f:
                json.dump(data, f, ensure_ascii=False, separators=(",", ":"))
            os.replace(tmp, os.path.join(out, "api", f"{name}.json"))

        lis = "".join(f'<li><a href="/zone/{zid}">{esc(name)}</a> <span class="muted">{n} NPCs, {nq} quests</span></li>'
                      for name, zid, n, nq in sorted(z_rows))
        body = f"""
<section class="intro"><h1>Zones</h1>
<p class="muted">{len(z_rows)} zones and instances with their NPCs, quests and loot tables.</p></section>
<section class="card"><ul class="loot cols">{lis}</ul></section>"""
        with open(os.path.join(out, "zones.html"), "w", encoding="utf-8") as f:
            f.write(page("Zones", body, path="/zones", description=f"{len(z_rows)} zones on Vanilla Plus."))

        body = f"""
<section class="intro"><h1>Quests</h1>
<p class="muted">{len(q_api)} quests from pfQuest's classic 1.12 data, not checked against Vanilla+ yet: start and end
NPCs, objectives, the quest chain and the reward items the database knows. By zone: <a href="/zones">Zones</a>.</p>
<form class="spell-search" onsubmit="return false">
  <input id="qq" type="search" placeholder="Name, zone or ID..." autocomplete="off">
  <select id="qs"><option value="">Both sides</option><option>Alliance</option><option>Horde</option></select>
  <input id="ql" type="number" min="1" max="60" placeholder="Level" style="width:5.5rem">
</form></section>
<section class="card"><p class="muted small" id="qcount"></p><ul class="loot" id="qlist"></ul></section>
<script>
(async () => {{
  const all = await (await fetch('/quests.json')).json();
  const $ = id => document.getElementById(id), MAX = 300;
  const esc = t => String(t).replace(/[&<>"]/g, c => ({{'&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;'}})[c]);
  const p = new URLSearchParams(location.search);
  $('qq').value = p.get('q') || ''; $('qs').value = p.get('side') || ''; $('ql').value = p.get('lvl') || '';
  function show() {{
    const q = $('qq').value.trim().toLowerCase(), s = $('qs').value, l = +$('ql').value;
    const hits = all.filter(x => (!s || !x[3] || x[3] === s) && (!l || Math.abs(x[2] - l) <= 2)
      && (!q || String(x[0]) === q || x[1].toLowerCase().includes(q) || x[4].toLowerCase().includes(q)));
    $('qcount').textContent = `${{hits.length}} quests` + (hits.length > MAX ? ` (first ${{MAX}} shown)` : '');
    $('qlist').innerHTML = hits.slice(0, MAX).map(x => `<li><a href="/quest/${{x[0]}}">${{esc(x[1])}}</a> <span class="muted">`
      + [x[2] ? 'Level ' + x[2] : '', x[3], x[4], '#' + x[0]].filter(Boolean).map(esc).join(' · ') + '</span></li>').join('');
    const u = new URLSearchParams(); if (q) u.set('q', q); if (s) u.set('side', s); if (l) u.set('lvl', l);
    history.replaceState(null, '', u.toString() ? '?' + u : location.pathname);
  }}
  ['qq', 'qs', 'ql'].forEach(id => $(id).addEventListener('input', show));
  show();
}})();
</script>"""
        with open(os.path.join(out, "quests.html"), "w", encoding="utf-8") as f:
            f.write(page("Quests", body, path="/quests", description=f"{len(q_api)} quests on Vanilla Plus."))
        return len(q_api), len(z_api)
