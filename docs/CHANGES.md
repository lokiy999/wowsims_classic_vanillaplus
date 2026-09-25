

## Part CC — Server-data tooltips for all sim spells (2026-09-25)

Local tooltips existed only for talent spells; every other spell (Fireball, Frostbolt, Innervate, procs, racials,
totems, ...) showed Wowhead's live Classic/SoD tooltip, in the results, the combat log, the rotation editor and the
buff pickers. New `tools/gen_server_spell_tooltips.py` collects the spell ids the sim (Go, lines mentioning a spell
id), the UI and the rotation APLs use, builds a tooltip from `Spell.csv` (name, rank, description, buff text) and
writes it into `assets/db_inputs/wowhead_spell_tooltips.csv` (Wowhead icon kept). The ids go to
`assets/db_inputs/server_spell_ids.txt`, which gen_db now adds to the DB. 876 spells, e.g. Fireball rank 12 now
reads "596 to 760 Fire damage and an additional 76", Frostbolt rank 10 "429 to 463" (as in game). 17 spells whose
text uses values the dump does not have (Searing Totem range, Sweeping Strikes / Blade Flurry charges, Execute's
per-rage damage, Chain Heal jumps, Thunder Clap targets) keep Wowhead's tooltip.

The description resolver (`tools/gen_custom_spell_tooltips.py`) now also knows `$a` (radius, via
`SpellRadius.csv`, column 88-90), `$n` (charges, column 26), `$x` (chain targets, column 100-102) and `$*N;`
multipliers, and reads `$/N;<spellId>..` in either order. 23 talent descriptions lost an "X" or raw `$` token that
way (Moonkin Aura 20 yards, Primal Fury 5 rage, Flurry-type "next 3 attacks").

UI: a spell renders its local tooltip only when the DB carries one (`Database.hasLocalSpellTooltip`); others keep
Wowhead's own tooltip (with its buff view) instead of Wowhead text in the local style. A late Wowhead dataset no
longer lands on an element that already has a local tooltip (results tables, timeline). Also local now: the
combat log, talents without a DBC description, and the enchant line under a gear slot (server enchant text).
Level scaling (realPointsPerLevel) is not applied in tooltips; top ranks are at or near their max level, so their
numbers match.
