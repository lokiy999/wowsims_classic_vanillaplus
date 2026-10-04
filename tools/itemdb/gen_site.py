#!/usr/bin/python3
"""Static item database site for db.lokiy.dev (Vanilla Plus custom items).

Reads the server data in the repo's CSV's/ folder:
  - VPlusItemDB.lua: in-game tooltip lines of every server item
  - AtlasLoot (the server's fork): loot tables, their boss/instance names and item icons
and writes plain HTML pages (Caddy serves them from disk):
  /index.html             search + browse by instance / boss
  /item/<id>.html         one page per item (tooltip, icon, where it drops), with Open Graph tags for link previews
  /loot/<table>.html      one page per loot table (boss, trash, set, ...) that has custom items
  /items.json             search / hover-tooltip data
  /style.css, /site.js

Scope (Lokiy 2026-10-04): custom items only = item ids that do not exist in classic 1.12 (above 24283) plus the
items AtlasLoot flags as new ("N"). All items come later: change is_custom().

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


# --------------------------------------------------------------------------------------------- html

def tooltip_html(item):
    color = Q_COLOR.get(item["quality"], "#ffffff")
    out = [f'<div class="tt-name" style="color:{color}">{esc(item["name"])}</div>']
    for ln in item["lines"]:
        cls = "tt-white"
        if re.match(r"^\(\d+\) Set:", ln) or ln.startswith("  ") or re.match(r"^.+\(\d+/\d+\)$", ln):
            cls = "tt-grey"
        elif ln.startswith(GREEN):
            cls = "tt-green"
        elif ln.startswith(('"', "Requires ", "Classes:", "Races:")):
            cls = "tt-grey" if not ln.startswith('"') else "tt-flavor"
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
  <a class="brand" href="/">{icon_img('inv_misc_book_09', 'small', 'brand-icon')}<span>{SITE_NAME}</span></a>
  <form class="search" action="/" method="get" role="search">
    <input type="search" name="q" placeholder="Search items..." aria-label="Search items" autocomplete="off">
  </form>
</header>
<main>
{body}
</main>
<footer>Vanilla Plus custom items, from the server's item data and AtlasLoot. Updated {date.today().isoformat()}.</footer>
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

    flagged_new = {r[0] for rows in tables.values() for r in rows if r[3]}

    def is_custom(iid):
        return iid in items and (iid > CLASSIC_MAX_ITEM_ID or iid in flagged_new)

    icons, sources = {}, {}
    for key, rows in tables.items():
        for iid, icon, _, _ in rows:
            if iid and icon and iid not in icons:
                icons[iid] = icon
            if is_custom(iid):
                sources.setdefault(iid, [])
                if key not in sources[iid]:
                    sources[iid].append(key)

    custom = sorted(i for i in items if is_custom(i))
    loot_keys = [k for k in tables if any(is_custom(r[0]) for r in tables[k])]

    def table_label(key):
        name, cat = names.get(key, (key, "Other"))
        return name or key, cat

    # clean output (only our generated files)
    for sub in ("item", "loot"):
        shutil.rmtree(os.path.join(out, sub), ignore_errors=True)
        os.makedirs(os.path.join(out, sub), exist_ok=True)

    # ---- item pages
    for iid in custom:
        it = items[iid]
        icon = icons.get(iid, "inv_misc_questionmark")
        color = Q_COLOR.get(it["quality"], "#ffffff")
        src_html = ""
        if sources.get(iid):
            lis = []
            for key in sources[iid]:
                name, cat = table_label(key)
                where = f' <span class="muted">({esc(cat)})</span>' if cat and cat != "Other" and cat not in name else ""
                lis.append(f'<li><a href="/loot/{key}">{esc(name)}</a>{where}</li>')
            src_html = f'<section class="card"><h2>Source</h2><ul class="sources">{"".join(lis)}</ul></section>'
        else:
            src_html = '<section class="card"><h2>Source</h2><p class="muted">No known drop source yet.</p></section>'
        desc = "\n".join(it["lines"])
        body = f"""
<article class="item-page">
  <div class="item-head">{icon_img(icon, 'large', 'icon-l')}
    <div><h1 style="color:{color}">{esc(it['name'])}</h1><div class="muted">Item {iid}</div></div>
  </div>
  <div class="tooltip">{tooltip_html(it)}</div>
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
            if is_custom(iid):
                rows.append(f'<li class="custom">{item_link(items[iid], icons.get(iid, icon))}</li>')
            else:
                q = items.get(iid, {}).get("quality", 1)
                nm = items.get(iid, {}).get("name", label)
                rows.append(f'<li><a class="item ext" href="https://www.wowhead.com/classic/item={iid}" rel="noopener">'
                            f'{icon_img(icon, "small", "icon-s")}<span style="color:{Q_COLOR.get(q, "#fff")}">{esc(nm)}</span></a>'
                            f'<span class="muted ext-note">classic</span></li>')
        n_custom = sum(1 for r in tables[key] if is_custom(r[0]))
        body = f"""
<article>
  <div class="crumbs"><a href="/">All items</a>{' / ' + esc(cat) if cat and cat != 'Other' else ''}</div>
  <h1>{esc(name)}</h1>
  <p class="muted">{n_custom} custom item{'s' if n_custom != 1 else ''}. Classic items link to Wowhead.</p>
  <ul class="loot">{''.join(rows)}</ul>
</article>"""
        with open(os.path.join(out, "loot", f"{key}.html"), "w", encoding="utf-8") as f:
            f.write(page(name, body, description=f"{name}: {n_custom} Vanilla Plus custom items.", path=f"/loot/{key}"))

    # ---- index: browse by category
    by_cat = {}
    for key in loot_keys:
        name, cat = table_label(key)
        by_cat.setdefault(cat, []).append(key)
    cats = [c for c in cat_order if c in by_cat] + [c for c in by_cat if c not in cat_order]
    blocks = []
    for cat in cats:
        tbl = []
        for key in by_cat[cat]:
            name, _ = table_label(key)
            its = [items[r[0]] for r in tables[key] if is_custom(r[0])]
            seen, links = set(), []
            for it in its:
                if it["id"] in seen:
                    continue
                seen.add(it["id"])
                links.append(f'<li>{item_link(it, icons.get(it["id"]))}</li>')
            tbl.append(f'<div class="table"><h3><a href="/loot/{key}">{esc(name)}</a></h3><ul class="loot">{"".join(links)}</ul></div>')
        blocks.append(f'<section class="cat" data-cat><h2>{esc(cat)}</h2>{"".join(tbl)}</section>')
    nosrc = [items[i] for i in custom if not sources.get(i)]
    if nosrc:
        nosrc.sort(key=lambda it: (-it["quality"], it["name"]))
        blocks.append('<section class="cat" data-cat><h2>No known drop source</h2><div class="table"><ul class="loot">'
                      + "".join(f'<li>{item_link(it, icons.get(it["id"]))}</li>' for it in nosrc) + "</ul></div></section>")
    body = f"""
<section class="intro">
  <h1>Vanilla Plus custom items</h1>
  <p class="muted">{len(custom)} items that are new on Vanilla Plus, with their in-game tooltips and where they drop.
  Link any item: <code>db.lokiy.dev/item/&lt;id&gt;</code></p>
</section>
<section id="results" class="card" hidden><h2>Search results</h2><ul class="loot" id="result-list"></ul></section>
<div id="browse">{''.join(blocks)}</div>"""
    with open(os.path.join(out, "index.html"), "w", encoding="utf-8") as f:
        f.write(page("Custom items", body, description=f"{len(custom)} Vanilla Plus custom items with tooltips and drop sources.", path="/"))

    # ---- 404
    with open(os.path.join(out, "404.html"), "w", encoding="utf-8") as f:
        f.write(page("Not found", '<article><h1>Not found</h1><p class="muted">That item or page is not in the database. '
                     'Only Vanilla Plus custom items are listed for now.</p><p><a href="/">Back to all items</a></p></article>'))

    # ---- data for search / hover tooltips
    data = [{"id": i, "n": items[i]["name"], "q": items[i]["quality"], "i": icon_url(icons.get(i), "small"),
             "s": [table_label(k)[0] for k in sources.get(i, [])], "t": tooltip_html(items[i])} for i in custom]
    with open(os.path.join(out, "items.json"), "w", encoding="utf-8") as f:
        json.dump(data, f, separators=(",", ":"))

    for asset in ("style.css", "site.js"):
        shutil.copy(os.path.join(HERE, asset), os.path.join(out, asset))
    shutil.rmtree(os.path.join(out, "icons"), ignore_errors=True)
    shutil.copytree(os.path.join(HERE, "icons"), os.path.join(out, "icons"))

    print(f"{len(custom)} items, {len(loot_keys)} loot tables, {len(nosrc)} without source -> {out}")


if __name__ == "__main__":
    main()
