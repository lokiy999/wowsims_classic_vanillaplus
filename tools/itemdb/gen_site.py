#!/usr/bin/python3
"""Static item database site for db.lokiy.dev (all Vanilla Plus items).

Reads the server data in the repo's CSV's/ folder:
  - VPlusItemDB.lua: in-game tooltip lines of every server item
  - AtlasLoot (the server's fork): loot tables, their boss/instance names and item icons
and writes plain HTML pages (Caddy serves them from disk):
  /index.html             search + browse by instance / boss
  /item/<id>.html         one page per item (tooltip, icon, where it drops), with Open Graph tags for link previews
  /loot/<table>.html      one page per loot table (boss, trash, set, ...) that has custom or reworked items
  /reworked.html          reworked classic items by instance / boss
  /instances.html         every loot table by instance
  /type/<type>.html       every item by type (item_types.py)
  /search.html            filter all items (search.js over /search.json; stats from item_stats.py)
  /items.json             search index; /tt/<id // 500>.json hover tooltips
  /style.css, /site.js

Scope (Lokiy 2026-10-04): custom items = item ids that do not exist in classic 1.12 (above 24283) plus the items
AtlasLoot flags as new ("N"); reworked items = classic ids whose server tooltip differs from Wowhead's classic one
(reworked.py). Every other server item gets a page too (Lokiy 2026-10-04: "all other items as well").

Usage:
  python3 tools/itemdb/gen_site.py [--out /var/www/db]
"""
import argparse
import glob
import html
import json
import os
import re
import shutil
import sys
from datetime import date

REPO = os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
sys.path.insert(0, os.path.join(REPO, "docs"))
from parse_vplus import parse_lua  # noqa: E402
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from reworked import RECIPE, find_reworked, load_wowhead, norm  # noqa: E402
from item_types import ARMOR_SLOTS, PAGES, classify  # noqa: E402
from item_stats import ALL_CLASSES, CLASSES, STAT_LABELS, parse as parse_stats  # noqa: E402
from loot_sync import load_loot, load_random_affixes, seen_affixes  # noqa: E402

CSV = os.path.join(REPO, "CSV's")
ATLAS = os.path.join(CSV, "AtlasLoot")
CLASSIC_MAX_ITEM_ID = 24283
SITE_NAME = "Vanilla Plus Database"
SITE_URL = "https://db.lokiy.dev"
ICON_URL = "https://wow.zamimg.com/images/wow/icons/{size}/{icon}.jpg"
QUALITY = {"Poor": 0, "Common": 1, "Uncommon": 2, "Rare": 3, "Epic": 4, "Legendary": 5, "Artifact": 6}
Q_COLOR = {0: "#9d9d9d", 1: "#ffffff", 2: "#1eff00", 3: "#0070dd", 4: "#a335ee", 5: "#ff8000", 6: "#e6cc80"}
GREEN = ("Equip:", "Use:", "Chance on hit:", "(", "Set:")
HERE = os.path.dirname(os.path.abspath(__file__))
# Server-made icons extracted from the client's patch MPQs (tools/itemdb/icons/<name>.jpg 56px, <name>_s.jpg 18px);
# everything else comes from the public icon CDN.
LOCAL_ICONS = {f[:-4] for f in os.listdir(os.path.join(HERE, "icons")) if f.endswith(".jpg") and not f.endswith("_s.jpg")}
# Typos in the server's AtlasLoot icon names.
ICON_ALIAS = {
    "inv_brancer_11": "inv_bracer_11",
    "inv_misc bomb_03": "inv_misc_bomb_03",
    "inv_misc_armorkit 27": "inv_misc_armorkit_27",
    "inv_throwing knife_01": "inv_throwingknife_01",
}

esc = html.escape
# Blizzard / server test and placeholder items: they keep their page but stay out of search and browse lists.
DEV_ITEM = re.compile(r"^QA|\(Test\)|\bTest (Item|Bow|Crit|Defense|Stationery)|TEST ITEM|Test Beatdown|"
                      r"Properties Test|^Deprecated |^Monster - |\[PH\]|\[DEP\]|\bUNUSED\b|^Unused |^DEBUG ")
# type page sections: armor slots in equipment order, the rest alphabetical
SECTION_ORDER = {s: n for n, s in enumerate(ARMOR_SLOTS)}


def slug(s):
    return re.sub(r"[^a-z0-9]+", "-", s.lower()).strip("-")


def icon_url(icon, size="large", absolute=False):
    icon = ICON_ALIAS.get(icon, icon or "inv_misc_questionmark")
    local = icon.replace(" ", "_")
    if local in LOCAL_ICONS:
        return f"{SITE_URL if absolute else ''}/icons/{local}{'_s' if size == 'small' else ''}.jpg"
    return ICON_URL.format(size=size, icon=icon)


# --------------------------------------------------------------------------------------------- data

def load_items():
    """{id: {name, quality, lines (tooltip lines without the Rarity line)}}"""
    raw = parse_lua(os.path.join(CSV, "VPlusItemDB.lua"))
    items = {}
    for iid, lines in raw.items():
        quality = 1
        keep = []
        for ln in lines:
            m = re.match(r"^Rarity: (\w+)$", ln)
            if m:
                quality = QUALITY.get(m.group(1), 1)
            else:
                keep.append(ln)
        if keep:
            items[iid] = {"id": iid, "name": keep[0], "quality": quality, "lines": keep[1:]}
    return items


def lua_string_expr(expr):
    """AL["Trash Mobs"].." ("..BZ["Molten Core"]..")"  ->  Trash Mobs (Molten Core)"""
    parts = re.findall(r'\w+\["((?:[^"\\]|\\.)*)"\]|"((?:[^"\\]|\\.)*)"', expr)
    return "".join(a or b for a, b in parts).strip()


def load_table_names():
    """{table_key: (display name, category)}; category is the instance / group name in AtlasLoot's boss list."""
    txt = open(os.path.join(ATLAS, "TableRegister", "loottables.en.lua"), encoding="utf-8", errors="replace").read()
    names = {}
    head, _, boss = txt.partition("AtlasLoot_TableNamesBoss")
    for key, expr in re.findall(r'^\s*\["(\w+)"\]\s*=\s*\{\s*(.*?),\s*"[^"]*"\s*\},?\s*$', head, re.M):
        names[key] = (lua_string_expr(expr), "Other")
    category = "Other"
    order = []
    for ln in boss.split("\n"):
        m = re.match(r'^\t\[(.*)\]\s*=\s*\{\s*$', ln)
        if m:
            category = lua_string_expr(m.group(1))
            order.append(category)
            continue
        m = re.match(r'^\s*\["(\w+)"\]\s*=\s*\{\s*(.*?),\s*"[^"]*"\s*\},?\s*$', ln)
        if m:
            names[m.group(1)] = (lua_string_expr(m.group(2)) or names.get(m.group(1), ("", ""))[0], category)
    return names, order


def load_text_codes():
    """AtlasLoot's #codes# (e.g. #r3# -> Honored, #a2# -> Leather) from Core/TextParsing.lua."""
    txt = open(os.path.join(ATLAS, "Core", "TextParsing.lua"), encoding="utf-8", errors="replace").read()
    codes = {}
    for code, expr in re.findall(r'gsub\(\s*text\s*,\s*"(#[^"]+#)"\s*,\s*(.+?)\);?\s*$', txt, re.M):
        val = lua_string_expr(expr)
        if val and code not in codes:
            codes[code] = val
    return codes


def clean_label(s, codes):
    s = re.sub(r"=[a-z]+\d*=", "", s)
    s = re.sub(r"#[^#\s]+#", lambda m: codes.get(m.group(0), ""), s)
    return re.sub(r"\s+", " ", s).strip(" ,")


def load_loot_tables():
    """{table_key: [(item_id, icon, display_name, is_new)]} in AtlasLoot order (id 0 rows are spacers / headers)."""
    tables = {}
    for path in sorted(glob.glob(os.path.join(ATLAS, "**", "*.lua"), recursive=True)):
        rel = os.path.relpath(path, ATLAS)
        if rel.startswith(("Locale", "TableRegister", "Core", "Libs")):
            continue
        txt = open(path, encoding="utf-8", errors="replace").read()
        for m in re.finditer(r'^\t(\w+) = \{\n(.*?)^\t\},?', txt, re.M | re.S):
            rows = []
            for e in re.finditer(r'\{\s*(\d+),\s*"([^"]*)",\s*"([^"]*)"(.*?)\},?\s*$', m.group(2), re.M):
                rows.append((int(e.group(1)), e.group(2).lower(), re.sub(r"^=q\d=", "", e.group(3)),
                             bool(re.search(r'"N"\s*,?\s*$', e.group(4).strip()))))
            if any(r[0] for r in rows):
                tables.setdefault(m.group(1), rows)
    return tables


# --------------------------------------------------------------------------------------------- observed drops

def norm_zone(z):
    z = re.sub(r"\s*\(.*?\)\s*", "", z or "").strip().lower()
    return re.sub(r"^(lower|upper) ", "", z)


def observed_drops(iid, keys, names, tables_by_cat, obs, boss_names, guess=True):
    """Credit this item's logged loot events to its AtlasLoot sources.

    Returns [(label, drops, kills, estimated)]. A loot is credited to an AtlasLoot boss of the item if that boss died
    in the 10 min before it; to a trash table if a non-boss creature of that zone died before it; items AtlasLoot
    has no source for are credited to the creature that died last (within 2 min), always as an estimate (only with
    guess: classic items without a source are often vendor, crafted or world drops, where the last kill means nothing).
    """
    events = obs["loots"].get(str(iid), [])
    if not events:
        return []
    allowed, trash_zones = set(), {}
    for key in keys:
        name, cat = names.get(key, (key, "Other"))
        m = re.match(r"^Trash Mobs \((.*)\)$", name)
        if m:
            # The logs only know the zone ("Dire Maul", "Blackrock Spire"), not the wing.
            zone = re.sub(r"\s*\(.*?\)\s*", "", m.group(1)).strip()
            zone = re.sub(r"^(Lower|Upper) ", "", zone)
            trash_zones[norm_zone(m.group(1))] = f"Trash Mobs ({zone})"
        elif "random" in name.lower() or "RANDOM" in key:
            allowed |= {names[k][0] for k in tables_by_cat.get(cat, []) if not names[k][0].startswith("Trash")}
        else:
            allowed.add(name)
    credit, guessed = {}, set()
    for cands in events:
        target = None
        for name, _, _ in cands:
            if name in allowed:
                target = name
                break
        if not target and trash_zones:
            for name, _, zone in cands:
                if name not in boss_names and norm_zone(zone) in trash_zones:
                    target = "trash:" + norm_zone(zone)
                    break
        if guess and not target and not keys and cands and cands[0][1] <= 120:
            target = cands[0][0]
            guessed.add(target)
        if target:
            credit[target] = credit.get(target, 0) + 1
    out = []
    for target, n in sorted(credit.items(), key=lambda kv: -kv[1]):
        if target.startswith("trash:"):
            zone = target[6:]
            label = trash_zones[zone]
            kills = sum(v for z, v in obs["zone_kills"].items() if norm_zone(z) == zone)
            estimated = True  # all trash of the zone counted as kills
        else:
            label, kills = target, obs["kills"].get(target, 0)
            estimated = target in guessed
        rate = n / kills if kills else 1
        if kills < 10 or rate > 0.5:
            estimated = True  # few kills or a suspiciously high rate
        out.append((label, n, kills, estimated))
    return out


def rate_text(drops, kills, estimated):
    if not kills:
        return "~?"
    r = drops / kills * 100
    if r > 100:
        return "~100%+"  # more drops than logged kills: a logger missed kills, or it can drop more than once
    s = f"{r:.0f}%" if r >= 10 else f"{r:.1f}%"
    return ("~" if estimated else "") + s


# --------------------------------------------------------------------------------------------- html

def tooltip_html(item, mark=(), mark_cls="tt-new"):
    """mark: normalized lines to highlight (the changes against the other version of a reworked item)."""
    color = Q_COLOR.get(item["quality"], "#ffffff")
    name_cls = f' {mark_cls}' if norm(item["name"]) in mark else ""
    out = [f'<div class="tt-name{name_cls}" style="color:{color}">{esc(item["name"])}</div>']
    for ln in item["lines"]:
        cls = "tt-white"
        if re.match(r"^\(\d+\) Set:", ln) or ln.startswith("  ") or re.match(r"^.+\(\d+/\d+\)$", ln):
            cls = "tt-grey"
        elif ln.startswith(GREEN):
            cls = "tt-green"
        elif ln.startswith(('"', "Requires ", "Classes:", "Races:")):
            cls = "tt-grey" if not ln.startswith('"') else "tt-flavor"
        if mark and norm(ln) in mark:
            cls += " " + mark_cls
        out.append(f'<div class="{cls}">{esc(ln)}</div>')
    return "".join(out)


def icon_img(icon, size="medium", cls="icon"):
    # An icon that is neither local nor on the CDN falls back to the question mark.
    fallback = ICON_URL.format(size=size, icon="inv_misc_questionmark")
    return (f'<img class="{cls}" src="{esc(icon_url(icon, size))}" alt="" loading="lazy" '
            f'width="36" height="36" onerror="this.onerror=null;this.src=\'{fallback}\'">')


def page(title, body, *, description="", image="", color="#a335ee", path="", extra_head=""):
    og = [
        f'<meta property="og:site_name" content="{SITE_NAME}">',
        f'<meta property="og:title" content="{esc(title)}">',
        f'<meta property="og:type" content="website">',
        f'<meta property="og:url" content="{SITE_URL}{esc(path)}">',
        f'<meta name="theme-color" content="{color}">',
    ]
    if description:
        og.append(f'<meta property="og:description" content="{esc(description)}">')
        og.append(f'<meta name="description" content="{esc(description[:300])}">')
    if image:
        og.append(f'<meta property="og:image" content="{esc(image)}">')
    return f"""<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>{esc(title)} - {SITE_NAME}</title>
{chr(10).join(og)}
<link rel="icon" href="{ICON_URL.format(size='small', icon='inv_misc_book_09')}">
<link rel="stylesheet" href="/style.css">
{extra_head}
</head>
<body>
<header class="top">
  <a class="home" href="https://lokiy.dev" title="Back to lokiy.dev">&larr; lokiy.dev</a>
  <a class="brand" href="/">{icon_img('inv_misc_book_09', 'small', 'brand-icon')}<span>{SITE_NAME}</span></a>
  <nav class="nav"><a href="/">Database</a><a href="/reworked">Reworked</a><a href="/type/weapons">Types</a><a href="/instances">Instances</a><a href="/search">Search</a></nav>
  <form class="search" action="/" method="get" role="search">
    <input type="search" name="q" placeholder="Search items..." aria-label="Search items" autocomplete="off">
  </form>
</header>
<main>
{body}
</main>
<footer>Vanilla Plus items, from the server's item data and AtlasLoot. Updated {date.today().isoformat()}.
&middot; Part of <a href="https://lokiy.dev">lokiy.dev</a>: <a href="https://sim.lokiy.dev">Sim</a>,
<a href="https://calc.lokiy.dev">Calculators</a>, <a href="https://addons.lokiy.dev">Addons</a>,
<a href="https://collection.lokiy.dev">Collection</a></footer>
<script src="/site.js" defer></script>
</body>
</html>
"""


def item_link(item, icon):
    color = Q_COLOR.get(item["quality"], "#ffffff")
    return (f'<a class="item" href="/item/{item["id"]}" data-item="{item["id"]}">{icon_img(icon, "small", "icon-s")}'
            f'<span style="color:{color}">{esc(item["name"])}</span></a>')


# --------------------------------------------------------------------------------------------- build

def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--out", default="/var/www/db")
    args = ap.parse_args()
    out = args.out

    items = load_items()
    names, cat_order = load_table_names()
    tables = load_loot_tables()
    codes = load_text_codes()
    wowhead = load_wowhead()

    flagged_new = {r[0] for rows in tables.values() for r in rows if r[3]}

    def is_custom(iid):
        return iid in items and (iid > CLASSIC_MAX_ITEM_ID or iid in flagged_new)

    reworked = find_reworked(items, CLASSIC_MAX_ITEM_ID, skip=flagged_new, wh=wowhead)
    all_items = sorted(items)
    custom = [i for i in all_items if is_custom(i)]
    browsable = [i for i in all_items if not DEV_ITEM.search(items[i]["name"])]
    types = {i: classify(items[i]) for i in all_items}

    icons, sources = {}, {}
    for key, rows in tables.items():
        for iid, icon, _, _ in rows:
            if iid and icon and iid not in icons:
                icons[iid] = icon
            if iid in items:
                sources.setdefault(iid, [])
                if key not in sources[iid]:
                    sources[iid].append(key)
    for iid in all_items:
        if iid not in icons and wowhead.get(iid, {}).get("icon"):
            icons[iid] = wowhead[iid]["icon"]
    loot_keys = [k for k in tables if any(r[0] in items for r in tables[k])]

    def table_label(key):
        name, cat = names.get(key, (key, "Other"))
        return name or key, cat

    def kind(iid):
        if DEV_ITEM.search(items[iid]["name"]):
            return "Test / unused item"
        if is_custom(iid):
            return "Custom item"
        if iid in reworked:
            return "Reworked classic item"
        if iid in wowhead and not items[iid]["name"].startswith(RECIPE):
            return "Classic item, same as classic"
        return "Classic item"

    # observed drops from players' combat logs (tools/itemdb/observed_loot.json, from loot_from_logs.py)
    obs_path = os.path.join(HERE, "observed_loot.json")
    obs = json.load(open(obs_path)) if os.path.exists(obs_path) else {"loots": {}, "kills": {}, "zone_kills": {}, "logs": 0}
    tables_by_cat = {}
    for key, (nm, cat) in names.items():
        tables_by_cat.setdefault(cat, []).append(key)
    boss_names = {nm for nm, cat in names.values() if nm and not nm.startswith("Trash")}
    observed = {i: observed_drops(i, sources.get(i, []), names, tables_by_cat, obs, boss_names, guess=is_custom(i))
                for i in all_items}
    # Lokiy's loot tracker (lokiy.dev/loot) and the possible random affixes per item
    loot = load_loot()
    loot_names, loot_drops, loot_time = loot if loot else ({}, {}, None)
    affixes = load_random_affixes()
    loot_note = ('<p class="muted small">From <a href="https://lokiy.dev/loot/">Lokiy\'s loot tracker</a>'
                 + (f', updated {date.fromtimestamp(loot_time).isoformat()}' if loot_time else '') + '.</p>')
    obs_note = (f'<p class="muted small">From {obs["logs"]} players\' combat logs. Rate = drops / kills seen in those logs. '
                f'<b>~</b> = estimate (few kills, a guessed source, or a rate that looks too high).</p>')

    # clean output (only our generated files)
    for sub in ("item", "loot", "type", "tt"):
        shutil.rmtree(os.path.join(out, sub), ignore_errors=True)
        os.makedirs(os.path.join(out, sub), exist_ok=True)

    # ---- item pages
    for iid in all_items:
        it = items[iid]
        rw = reworked.get(iid)
        icon = icons.get(iid, "inv_misc_questionmark")
        color = Q_COLOR.get(it["quality"], "#ffffff")
        page_key, section = types[iid]
        if sources.get(iid):
            lis = []
            for key in sources[iid]:
                name, cat = table_label(key)
                where = f' <span class="muted">({esc(cat)})</span>' if cat and cat != "Other" and cat not in name else ""
                lis.append(f'<li><a href="/loot/{key}">{esc(name)}</a>{where}</li>')
            src_html = f'<section class="card"><h2>Source</h2><ul class="sources">{"".join(lis)}</ul></section>'
        else:
            src_html = '<section class="card"><h2>Source</h2><p class="muted">No known drop source yet.</p></section>'
        if observed[iid]:
            trs = "".join(f'<tr><td>{esc(label)}</td><td>{n}</td><td>{k}</td><td class="rate">{rate_text(n, k, est)}</td></tr>'
                          for label, n, k, est in observed[iid])
            src_html += (f'<section class="card obs"><h2>Observed drops</h2><table class="obs-table"><thead><tr><th>Source</th>'
                         f'<th>Drops</th><th>Kills</th><th>Rate</th></tr></thead><tbody>{trs}</tbody></table>{obs_note}</section>')
        if loot_drops.get(iid):
            trs = "".join(f'<tr><td>{esc(src)}</td><td>{n}</td><td>{k}</td><td class="rate">'
                          f'{(f"{r * 100:.0f}%" if r >= 0.1 else f"{r * 100:.1f}%") if r is not None else ""}</td></tr>'
                          for src, n, k, r in loot_drops[iid])
            src_html += (f'<section class="card obs"><h2>Lokiy\'s loot</h2><table class="obs-table"><thead><tr>'
                         f'<th>Source</th><th>Drops</th><th>Kills</th><th>Rate</th></tr></thead><tbody>{trs}</tbody></table>'
                         f'{loot_note}</section>')
        seen = seen_affixes(it["name"], loot_names.get(iid, {}))
        if affixes.get(iid) or seen:
            groups = {}
            for name, eff in affixes.get(iid, []):
                groups.setdefault(name, []).append(eff)
            for name in seen:
                groups.setdefault(name, [])
            lis = []
            for name in sorted(groups, key=lambda n: (-seen.get(n, 0), n)):
                effs = " / ".join(esc(e) for e in groups[name]) or '<span class="muted">stats unknown</span>'
                tag = f'<span class="tag">seen {seen[name]}&times;</span>' if name in seen else ""
                lis.append(f'<li class="{"seen" if name in seen else ""}"><span class="affix" style="color:{color}">{esc(it["name"])} '
                           f'<b>{esc(name)}</b></span>{tag}<span class="muted affix-stats">{effs}</span></li>')
            src_html += (f'<section class="card affixes"><h2>Random affixes</h2><ul class="affix-list">{"".join(lis)}</ul>'
                         f'<p class="muted small">Possible affixes from WoW Classic data (Wowhead); a name with several '
                         f'values has tiers. "Seen" = looted in <a href="https://lokiy.dev/loot/">Lokiy\'s loot tracker</a>'
                         f', matched by name.</p></section>')
        if rw:
            classic_item = {"name": rw["classic_name"], "quality": it["quality"], "lines": rw["classic"]}
            top_html = f"""<section class="card changes">
  <h2>Changed from classic</h2>
  <div class="compare">
    <div><h3>Vanilla Plus</h3><div class="tooltip">{tooltip_html(it, set(rw["added"]), "tt-new")}</div></div>
    <div><h3>Classic</h3><div class="tooltip">{tooltip_html(classic_item, set(rw["removed"]), "tt-old")}</div></div>
  </div>
  <p class="muted small">Highlighted lines differ. Classic = WoW Classic data from
  <a href="https://www.wowhead.com/classic/item={iid}" rel="noopener">Wowhead</a>.</p>
</section>"""
        else:
            top_html = f'<div class="tooltip">{tooltip_html(it)}</div>'
        wh_link = (f' &middot; <a href="https://www.wowhead.com/classic/item={iid}" rel="noopener">Wowhead</a>'
                   if iid in wowhead and not is_custom(iid) else "")
        desc = "\n".join(it["lines"])
        body = f"""
<article class="item-page">
  <div class="item-head">{icon_img(icon, 'large', 'icon-l')}
    <div><h1 style="color:{color}">{esc(it['name'])}</h1>
    <div class="muted">{kind(iid)} {iid} &middot; <a href="/type/{page_key}#{slug(section)}">{esc(section)}</a>{wh_link}</div></div>
  </div>
  {top_html}
  {src_html}
</article>"""
        with open(os.path.join(out, "item", f"{iid}.html"), "w", encoding="utf-8") as f:
            f.write(page(it["name"], body, description=desc, image=icon_url(icon, "large", absolute=True),
                         color=color, path=f"/item/{iid}"))

    # ---- loot table pages
    for key in loot_keys:
        name, cat = table_label(key)
        rows = []
        for iid, icon, label, _ in tables[key]:
            if not iid:
                header = clean_label(label, codes)
                if header:
                    rows.append(f'<li class="sub">{esc(header)}</li>')
                continue
            if iid in items:
                rate = next((rate_text(n, k, est) for label, n, k, est in observed[iid] if label == name), "")
                rate_html = f'<span class="rate" title="Observed drop rate in players\' combat logs">{rate}</span>' if rate else ""
                tag, cls = "", ""
                if is_custom(iid):
                    tag, cls = '<span class="tag">new</span>', ' class="custom"'
                elif iid in reworked:
                    tag, cls = '<span class="tag">changed</span>', ' class="reworked"'
                rows.append(f'<li{cls}>{item_link(items[iid], icons.get(iid, icon))}{tag}{rate_html}</li>')
            else:
                rows.append(f'<li><a class="item ext" href="https://www.wowhead.com/classic/item={iid}" rel="noopener">'
                            f'{icon_img(icon, "small", "icon-s")}<span>{esc(label)}</span></a>'
                            f'<span class="muted ext-note">Wowhead</span></li>')
        n_custom = sum(1 for r in tables[key] if is_custom(r[0]))
        n_rw = sum(1 for r in tables[key] if r[0] in reworked)
        counts = ", ".join(x for x in (f"{n_custom} new" if n_custom else "", f"{n_rw} changed" if n_rw else "") if x)
        counts = f"{counts} item{'s' if n_custom + n_rw != 1 else ''} on Vanilla Plus." if counts else "No changes on Vanilla Plus."
        nk = obs["kills"].get(name, 0)
        kills_note = (f"Killed {nk} times in players' combat logs; percentages are the observed drop rates "
                      f"(~ = estimate).") if nk else ""
        body = f"""
<article>
  <div class="crumbs"><a href="/instances">Instances</a>{' / ' + esc(cat) if cat and cat != 'Other' else ''}</div>
  <h1>{esc(name)}</h1>
  <p class="muted">{counts} {kills_note}</p>
  <ul class="loot">{''.join(rows)}</ul>
</article>"""
        with open(os.path.join(out, "loot", f"{key}.html"), "w", encoding="utf-8") as f:
            f.write(page(name, body, description=f"{name} loot on Vanilla Plus. {counts}", path=f"/loot/{key}"))

    # ---- browse pages: custom items (index) and reworked items, by instance / boss
    by_cat_all = {}
    for key in loot_keys:
        by_cat_all.setdefault(table_label(key)[1], []).append(key)
    cats_all = [c for c in cat_order if c in by_cat_all] + [c for c in by_cat_all if c not in cat_order]

    def browse_blocks(wanted):
        blocks = []
        for cat in cats_all:
            tbl = []
            for key in by_cat_all[cat]:
                name, _ = table_label(key)
                seen, links = set(), []
                for r in tables[key]:
                    if r[0] in items and wanted(r[0]) and r[0] not in seen:
                        seen.add(r[0])
                        links.append(f'<li>{item_link(items[r[0]], icons.get(r[0]))}</li>')
                if links:
                    tbl.append(f'<div class="table"><h3><a href="/loot/{key}">{esc(name)}</a></h3><ul class="loot">{"".join(links)}</ul></div>')
            if tbl:
                blocks.append(f'<section class="cat" data-cat><h2>{esc(cat)}</h2>{"".join(tbl)}</section>')
        nosrc = [items[i] for i in browsable if wanted(i) and not sources.get(i)]
        if nosrc:
            nosrc.sort(key=lambda it: (-it["quality"], it["name"]))
            blocks.append('<section class="cat" data-cat><h2>No known drop source</h2><div class="table wide"><ul class="loot cols">'
                          + "".join(f'<li>{item_link(it, icons.get(it["id"]))}</li>' for it in nosrc) + "</ul></div></section>")
        return "".join(blocks), len(nosrc)

    type_counts = {}
    for iid in browsable:
        type_counts[types[iid][0]] = type_counts.get(types[iid][0], 0) + 1
    type_nav = ('<section class="card"><h2>Browse by type</h2><ul class="type-grid">'
                + "".join(f'<li><a href="/type/{k}">{esc(nm)}</a> <span class="muted">{type_counts[k]}</span></li>'
                          for k, nm in PAGES if type_counts.get(k)) + "</ul></section>")

    search_box = '<section id="results" class="card" hidden><h2>Search results</h2><ul class="loot" id="result-list"></ul></section>'
    blocks, n_nosrc = browse_blocks(is_custom)
    body = f"""
<section class="intro">
  <h1>Vanilla Plus item database</h1>
  <p class="muted">All {len(all_items)} items on the server with their in-game tooltips and where they drop:
  {len(custom)} new items (below), <a href="/reworked">{len(reworked)} reworked classic items</a>, and every other item
  <a href="/type/consumables">by type</a> or <a href="/instances">by instance</a>.
  Link any item: <code>db.lokiy.dev/item/&lt;id&gt;</code></p>
</section>
{search_box}
<div id="browse">{type_nav}<h2 class="section-title">New on Vanilla Plus</h2>{blocks}</div>"""
    with open(os.path.join(out, "index.html"), "w", encoding="utf-8") as f:
        f.write(page("Item database", body, path="/",
                     description=f"All {len(all_items)} Vanilla Plus items: {len(custom)} new, {len(reworked)} reworked, with tooltips and drop sources."))

    blocks, n_rw_nosrc = browse_blocks(lambda i: i in reworked)
    body = f"""
<section class="intro">
  <h1>Reworked classic items</h1>
  <p class="muted">{len(reworked)} classic items whose stats or effects are different on Vanilla Plus. Each item page
  shows the Vanilla Plus and the classic tooltip side by side. Compared against WoW Classic data from Wowhead, so a few
  differences can be Classic-only changes rather than Vanilla Plus ones.</p>
</section>
{search_box}
<div id="browse">{blocks}</div>"""
    with open(os.path.join(out, "reworked.html"), "w", encoding="utf-8") as f:
        f.write(page("Reworked classic items", body, path="/reworked",
                     description=f"{len(reworked)} classic items with changed stats on Vanilla Plus, compared with classic."))

    # ---- instances: every loot table by instance / group
    blocks = []
    for cat in cats_all:
        lis = []
        for key in by_cat_all[cat]:
            name, _ = table_label(key)
            n_new = sum(1 for r in tables[key] if is_custom(r[0]) or r[0] in reworked)
            badge = f' <span class="tag">{n_new} new/changed</span>' if n_new else ""
            lis.append(f'<li><a href="/loot/{key}">{esc(name)}</a>{badge}</li>')
        blocks.append(f'<section class="cat"><h2>{esc(cat)}</h2><ul class="loot cols wide-list">{"".join(lis)}</ul></section>')
    body = f"""
<section class="intro"><h1>Instances &amp; loot tables</h1>
<p class="muted">Every AtlasLoot table of the server. The tag counts the new and changed items in it.</p></section>
{''.join(blocks)}"""
    with open(os.path.join(out, "instances.html"), "w", encoding="utf-8") as f:
        f.write(page("Instances", body, path="/instances", description="Vanilla Plus loot tables by instance and boss."))

    # ---- type pages: every item, by type and section
    for page_key, page_name in PAGES:
        ids = [i for i in browsable if types[i][0] == page_key]
        if not ids:
            continue
        secs = {}
        for i in ids:
            secs.setdefault(types[i][1], []).append(i)
        order = sorted(secs, key=lambda s: (SECTION_ORDER.get(s, 50), s))
        toc = " &middot; ".join(f'<a href="#{slug(s)}">{esc(s)}</a> <span class="muted">{len(secs[s])}</span>' for s in order)
        blocks = []
        for s in order:
            lst = sorted(secs[s], key=lambda i: (-items[i]["quality"], items[i]["name"]))
            lis = []
            for i in lst:
                tag = '<span class="tag">new</span>' if is_custom(i) else ('<span class="tag">changed</span>' if i in reworked else "")
                lis.append(f'<li>{item_link(items[i], icons.get(i))}{tag}</li>')
            blocks.append(f'<section class="cat" id="{slug(s)}"><h2>{esc(s)}</h2><div class="table wide"><ul class="loot cols">'
                          f'{"".join(lis)}</ul></div></section>')
        tabs = " ".join(f'<a class="{"on" if k == page_key else ""}" href="/type/{k}">{esc(nm)}</a>'
                        for k, nm in PAGES if type_counts.get(k))
        body = f"""
<nav class="type-tabs">{tabs}</nav>
<section class="intro"><h1>{esc(page_name)}</h1>
<p class="muted">{len(ids)} items, best quality first. Tags: <span class="tag">new</span> on Vanilla Plus,
<span class="tag">changed</span> from classic.</p><p class="toc">{toc}</p></section>
{''.join(blocks)}"""
        with open(os.path.join(out, "type", f"{page_key}.html"), "w", encoding="utf-8") as f:
            f.write(page(page_name, body, path=f"/type/{page_key}", description=f"{len(ids)} Vanilla Plus {page_name.lower()}."))

    # ---- advanced search: /search.html + /search.json (search.js filters it in the browser)
    sec_names = sorted({types[i][1] for i in browsable})
    sec_idx = {s: n for n, s in enumerate(sec_names)}
    type_idx = {k: n for n, (k, _) in enumerate(PAGES)}
    cat_idx = {c: n for n, c in enumerate(cats_all)}
    srows = []
    for i in browsable:
        it = items[i]
        pk, sec = types[i]
        info = parse_stats(it, pk, sec)
        status = 3 if is_custom(i) else 2 if i in reworked else 0 if kind(i).endswith("same as classic") else 1
        cats = sorted({cat_idx[table_label(k)[1]] for k in sources.get(i, []) if table_label(k)[1] in cat_idx})
        ic = icon_url(icons.get(i), "small")  # CDN icons as their bare name, local ones as /icons/<name>_s.jpg
        cdn = ICON_URL.split("{icon}")[0].format(size="small")
        ic = ic[len(cdn):-4] if ic.startswith(cdn) else ic
        srows.append([i, it["name"], it["quality"], type_idx[pk], sec_idx[sec], info.get("lv", 0), info.get("sl", ""),
                      info.get("b", ""), info.get("cl", ALL_CLASSES), status, info.get("st", 0), cats or 0, ic,
                      1 if affixes.get(i) or seen_affixes(it["name"], loot_names.get(i, {})) else 0])
    with open(os.path.join(out, "search.json"), "w", encoding="utf-8") as f:
        json.dump({"types": PAGES, "sections": sec_names, "cats": cats_all, "classes": CLASSES,
                   "stats": STAT_LABELS, "items": srows}, f, separators=(",", ":"))
    body = """
<section class="intro"><h1>Item search</h1>
<p class="muted">Filter all items. Stats are read from the in-game tooltips; set bonuses are not counted. The URL keeps
your filters, so you can share a search.</p></section>
<form id="filters" class="card filters loading" autocomplete="off">
  <label class="f-name">Name<input name="q" type="search" placeholder="Item name or id"></label>
  <fieldset class="f-quality"><legend>Quality</legend><div id="qualities" class="chips"></div></fieldset>
  <label>Type<select name="type"></select></label>
  <label>Subtype<select name="sec"></select></label>
  <label>Slot<select name="slot"></select></label>
  <label>Usable by<select name="cls"></select></label>
  <label>Required level<span class="range"><input name="lmin" type="number" min="0" max="60" placeholder="min" aria-label="Minimum level"><span class="muted">-</span><input name="lmax" type="number" min="0" max="60" placeholder="max" aria-label="Maximum level"></span></label>
  <label>Binds<select name="bind"></select></label>
  <label>Vanilla Plus<select name="st"><option value="">Any</option><option value="vp">New or changed</option><option value="new">New items</option><option value="changed">Changed from classic</option><option value="same">Same as classic</option></select></label>
  <label>Drops in<select name="src"></select></label>
  <label>Random affixes<select name="ra"><option value="">Any</option><option value="1">Has random affixes</option><option value="0">No random affixes</option></select></label>
  <fieldset class="f-stats"><legend>Stats (item has the stat, at least the value)</legend><div id="stat-rows"></div>
    <div class="f-actions"><button type="button" id="add-stat">+ Add stat filter</button><button type="button" id="reset">Reset all</button></div></fieldset>
</form>
<div id="results-top" class="results-head"><strong id="count">Loading...</strong><div class="pager"></div></div>
<div class="table-wrap"><table id="results-table" class="results"></table></div>
<div class="results-foot"><div class="pager"></div></div>"""
    with open(os.path.join(out, "search.html"), "w", encoding="utf-8") as f:
        f.write(page("Item search", body, path="/search",
                     description="Search all Vanilla Plus items by quality, type, slot, level, class and stats.",
                     extra_head='<script src="/search.js" defer></script>'))

    # ---- 404
    with open(os.path.join(out, "404.html"), "w", encoding="utf-8") as f:
        f.write(page("Not found", '<article><h1>Not found</h1><p class="muted">That item or page is not in the database.</p>'
                     '<p><a href="/">Back to all items</a></p></article>'))

    # ---- data: search index (small, all items) and hover tooltips in chunks of 500 ids (/tt/<id // 500>.json)
    data, chunks = [], {}
    for i in browsable:
        d = {"id": i, "n": items[i]["name"], "q": items[i]["quality"], "i": icon_url(icons.get(i), "small")}
        src = [table_label(k)[0] for k in sources.get(i, [])]
        if src:
            d["s"] = src
        if is_custom(i):
            d["c"] = 1
        elif i in reworked:
            d["r"] = 1
        data.append(d)
        chunks.setdefault(i // 500, {})[i] = tooltip_html(items[i])
    with open(os.path.join(out, "items.json"), "w", encoding="utf-8") as f:
        json.dump(data, f, separators=(",", ":"))
    for n, chunk in chunks.items():
        with open(os.path.join(out, "tt", f"{n}.json"), "w", encoding="utf-8") as f:
            json.dump(chunk, f, separators=(",", ":"))

    for asset in ("style.css", "site.js", "search.js"):
        shutil.copy(os.path.join(HERE, asset), os.path.join(out, asset))
    shutil.rmtree(os.path.join(out, "icons"), ignore_errors=True)
    shutil.copytree(os.path.join(HERE, "icons"), os.path.join(out, "icons"))

    print(f"{len(all_items)} items ({len(custom)} custom, {len(reworked)} reworked), {len(loot_keys)} loot tables, "
          f"{n_nosrc} custom / {n_rw_nosrc} reworked without source -> {out}")


if __name__ == "__main__":
    main()
