"""db.lokiy.dev/ilvl: the item level model's calculator and the report on Vanilla+ items (sim repo Part HY).

/ilvl.html (calculator + report; ilvl.js does both in the browser) and /api/ilvl.json: {"fit": the saved fit
(ilvl_fit.json: weights, slot modifiers, budget lines, base armor, DPS lines, held-out misses, data counts), "tiers":
[{tier, n, median, p10, p90}], "items": [report rows: id, name, q, ilvl, model, group, role, tier, status, flags,
why?, suggest?, comp? [[id, name, ilvl]]]}. Model values are labeled as such and never replace an item's real level.
"""
import html
import json
import os

from ilvl_model import FIT, GROUP_LABEL, TIERS, BUDGET_FLAG, Model, report


def esc(s):
    return html.escape(str(s), quote=True)


def write(out, page, items, status, instances):
    """items: {id: {"name", "q", "ilvl", "lines"}} (api/items.json); status {id: search status}; instances {id: set of
    instance names}. Returns (rows, number of suggestions per tier)."""
    fit = json.load(open(FIT, encoding="utf-8"))
    rows, tstat = report(items, status, instances, Model.load())
    with open(os.path.join(out, "api", "ilvl.json"), "w", encoding="utf-8") as f:
        json.dump({"fit": fit, "tiers": tstat, "items": rows}, f, ensure_ascii=False, separators=(",", ":"))
    held = "".join(f'<tr><td>{"Green" if h["q"] == 2 else "Blue" if h["q"] == 3 else "Epic"}</td>'
                   f'<td>{esc(GROUP_LABEL.get(h["group"], "All") if h["group"] != "all" else "All")}</td>'
                   f'<td class="rate">{h["n"]}</td><td class="rate">{h["median"]}</td><td class="rate">{h["p90"]}</td></tr>'
                   for h in fit["heldout"] if h["group"] == "all" or h["n"] >= 10)
    tiers = "".join(f'<tr><td>{n + 1}. {esc(t["tier"].replace("|", " / "))}</td><td class="rate">{t["n"]}</td>'
                    f'<td class="rate">{t["median"]}</td><td class="rate">{t["p10"]}-{t["p90"]}</td>'
                    f'<td class="rate">{sum(1 for r in rows if r["tier"] == n and "suggest" in r)}</td>'
                    f'<td class="muted small">{"its median is below the tier before&#39;s: the tier breaks the order" if t.get("below_prev") else ""}</td></tr>'
                    for n, t in enumerate(tstat))
    body = f"""
<section class="intro"><h1>Item level calculator</h1>
<p class="muted">A model of classic item budgets: what item level an item's stats are worth. It is fitted on
{fit["items"]:,} classic items ({fit["from_wago"]:,} from Blizzard's classic item data, {fit["from_wowhead"]:,} from
Wowhead's classic tooltips only; their Equip lines always from Wowhead) and applied to the items Vanilla+ changed or added.
<b>Model and suggested item levels are estimates</b>; the real item level of every item is the server's and is not
changed by this page.</p></section>

<section class="card"><h2>Calculator</h2>
<form id="calc" class="filters calc" autocomplete="off">
  <label>Slot<select name="slot"></select></label>
  <label>Armor type<select name="atype"><option value="">-</option><option>Cloth</option><option>Leather</option>
    <option>Mail</option><option>Plate</option><option>Shield</option></select></label>
  <label>Quality<select name="q"><option value="2">Uncommon (green)</option><option value="3">Rare (blue)</option>
    <option value="4" selected>Epic</option></select></label>
  <label>Armor<input name="armor" type="number" min="0"></label>
  <label>Weapon DPS<input name="dps" type="number" min="0" step="0.1"></label>
  <label class="chk"><input name="set" type="checkbox"> Set piece</label>
  <label class="chk"><input name="effect" type="checkbox"> Has a Use / proc effect</label>
  <fieldset class="stats-in"><legend>Stats (Equip bonuses as the tooltip shows them)</legend><div id="stat-inputs"></div></fieldset>
</form>
<div id="calc-out" class="calc-out"></div>
<p class="muted small">Stat value = (&Sigma; (amount &times; weight)<sup>1.5</sup>)<sup>2/3</sup> &divide; slot modifier, read
off a straight budget line per quality; armor above the normal armor for the slot counts as bonus armor; weapons also
get an item level from their DPS (70 / 30 with the stats). Use / proc effects aren't priced: with one, the model item
level is a lower bound.</p></section>

<section class="card"><h2>How well it fits</h2>
<p class="muted">5-fold held-out check on classic items without Use / proc effects: the model fitted on four fifths
predicts the fifth. Misses in item levels.</p>
<table class="obs-table"><thead><tr><th>Quality</th><th>Slot group</th><th>Items</th><th>Median miss</th>
<th>90% within</th></tr></thead><tbody>{held}</tbody></table></section>

<section class="card"><h2>Raid tiers</h2>
<p class="muted">Progression order (Lokiy): item levels should go up from tier to tier, outliers allowed. Real item
levels of each raid's blue / epic level-60 items. An item Vanilla+ changed or added that is below the tier before or above
the tier after is flagged "breaks the order" and gets a suggestion: the item level of comparable items in the same
raid (same slot group, quality and role, similar stat value), a few levels higher only where its stats need it.
"Outliers allowed": only items below most of the tier before (its 10th percentile) or above most of the tier after (its
90th) are flagged; a tier whose median is below the tier before is marked as a whole instead.</p>
<table class="obs-table"><thead><tr><th>Tier</th><th>Items</th><th>Median ilvl</th><th>10-90%</th><th>Suggestions</th>
<th></th></tr></thead><tbody>{tiers}</tbody></table></section>

<section class="card"><h2>Vanilla+ changed and new items</h2>
<form id="rep-filters" class="filters" autocomplete="off">
  <label>Raid<select name="tier"><option value="">Any</option>{''.join(f'<option value="{n}">{esc(t.replace("|", " / "))}</option>' for n, t in enumerate(TIERS))}<option value="none">Not in these raids</option></select></label>
  <label>Quality<select name="q"><option value="">Any</option><option value="2">Green</option><option value="3">Blue</option><option value="4">Epic</option></select></label>
  <label>Slot group<select name="group"><option value="">Any</option>{''.join(f'<option value="{k}">{esc(v)}</option>' for k, v in GROUP_LABEL.items())}</select></label>
  <label>Flag<select name="flag"><option value="">Any</option><option>breaks the order</option><option>over budget</option>
    <option>under budget</option><option>can't price</option></select></label>
  <label>Name<input name="name" type="search"></label>
</form>
<p class="muted small" id="rep-count"></p>
<div class="tbl"><table class="obs-table" id="rep"></table></div>
<p class="muted small">Model = the item level its stats are worth (over / under budget: more than {BUDGET_FLAG} levels from
the real one). Suggested = only for items that break the tier order, with the comparable items it is based on.</p></section>
<script src="/ilvl.js" defer></script>"""
    with open(os.path.join(out, "ilvl.html"), "w", encoding="utf-8") as f:
        f.write(page("Item level calculator", body, path="/ilvl",
                     description="What item level an item's stats are worth, and Vanilla+'s changed items against it."))
    per_tier = {t: sum(1 for r in rows if r["tier"] == n and "suggest" in r) for n, t in enumerate(TIERS)}
    return rows, per_tier
