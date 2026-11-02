# Spell coefficients (Vanilla+)

Working document for finding the real spell power / bonus healing coefficients (Lokiy 2026-11-02, started with the
priest). Everything we find or decide in the coefficient chat goes here. Open items are at the bottom.

## What the server data has

Spell.csv / the spell db (db.lokiy.dev spell pages) have base values and cast time indexes, **no coefficient field**.
So a coefficient is either the classic rule or has to be measured in game. The sim uses the classic rule:
cast time / 3.5 (cast time capped at 3.5 sec), HoTs duration / 15, and spells learned below level 20 lose 3.75% of
their coefficient per level under 20.

## Cast times (SpellCastTimes.dbc, hub `data\dbc`)

Index 14 = 3000 ms, index 15 = **4000 ms**, index 16 = 1500 ms.

- **Greater Heal and Prayer of Healing are 4 sec baseline** (confirmed by Lokiy 2026-11-02). The sim had them at
  3 sec (`sim/priest/heals.go`, `registerHealingSpells`): see the open items.
- Heal 3 sec, Flash Heal 1.5 sec, Lesser Heal r1 1.5 / r3 2.5 sec: these match the sim.

## Priest healing: coefficients (classic rule) and expected healing with 408 bonus healing

Max ranks, before talents. Spiritual Healing (+15% at 5/5) multiplies the whole result, bonus included.

| Spell | Base | Coefficient | Bonus added | Expected |
|---|---|---|---|---|
| Greater Heal r5 (4 sec, capped at 3.5) | 2752-3071 | 1.000 | +408 | avg 3320 (3160-3479) |
| Flash Heal r7 (1.5 sec) | 812-958 | 0.429 | +174.9 | avg 1060 (987-1133) |
| Heal r4 (3 sec, level 34) | 712-804 | 0.857 | +349.7 | avg 1108 (1062-1154) |
| Prayer of Healing r5 (4 sec, capped at 3.5, / 3), per target | 1041-1099 | 0.333 | +136.0 | avg 1206 (1177-1235) |
| Renew r10 (15 sec) | 970 over 15 sec | 1.0 total (0.2 per tick) | +408 | 1378 total, 275.6 per tick |
| Power Word: Shield r10 | 942 absorb | 0.1 | +40.8 | 982.8 |
| Lesser Heal r3 (2.5 sec, level 10: x0.625) | 135-157 | 0.446 | +182.1 | avg 328 |

For comparison, with the old 3 sec assumption: Greater Heal 0.857 (avg 3261), Prayer of Healing 0.286 (avg 1187).
Not looked at yet: Holy Nova's heal, Desperate Prayer, Lightwell (spell IDs did not resolve in the spell db).

## Open items

1. **Fix the sim's Greater Heal and Prayer of Healing cast time** to 4 sec (`heals.go`; Divine Fury still takes
   0.1 sec per rank off Greater Heal). Their coefficient then becomes 1.0 and 1/3 (cap at 3.5 sec). Changes healing per
   second and mana efficiency of every priest healing result: needs a Part entry in CHANGES.md and updated test results.
2. **Measure the coefficients in game** (nothing above is verified on the server): cast each spell many times at a
   known bonus healing on a target that cannot overheal and compare the average to the table. 408 bonus healing:
   Greater Heal 3320 (1.0) vs 3261 (0.857). Renew's tick shows its coefficient directly (275.6 if 1.0).
3. **Check Divine Fury / other talents** change Greater Heal from 4 sec the way the sim assumes.
4. Holy Nova, Desperate Prayer, Lightwell coefficients.
5. Other classes (the chat started with priest).
