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

## Measurements (Lokiy, in game)

Bonus healing 749 (all non-crit unless noted; Lokiy 2026-11-02, corrected from a first, wrong note).

| Spell | Seen | Classic rule predicts (749 bonus, no talents) |
|---|---|---|
| Heal r2 (level 22) | usually 950-1010 | 1071-1133 (coefficient 0.857) |
| Heal r4 (level 34) | 1461 once (not a crit) | 1354-1446 (0.857) |

What it implies (749 bonus healing, no talent multiplier):
- Rank 2: base roll 429-491 is a 62 spread and the observed 950-1010 is a 60 spread, so the bonus part is steady at
  about 520 (950-429 = 521, 1010-491 = 519). Coefficient = 520 / 749 = **0.69**, not 0.857. The matching spread also
  suggests no Spiritual Healing multiplier (1.15 would widen it to about 71).
- Rank 4: 1461 is above the classic maximum of 1446, so its coefficient is at least 0.877. 1461 is exactly
  712 + 749, the minimum base roll with coefficient 1.0 (then rank 4 would be 1461-1553).
- So Heal's coefficient differs by rank by more than the level-below-20 rule explains (r2 about 0.69, r4 about
  0.88-1.0). Could be a Vanilla+ change per rank, or a downranking rule.

Still needed: Heal r4 min / max / typical (does it never go below 1461?), r1 and r3 the same way, the bonus healing
shown for each, and the Spiritual Healing rank.

### Heal rank 2 test: 50-100 casts (Lokiy 2026-11-02)

Lokiy cast Heal rank 2 about 50-100 times (non-crit, bonus healing 749) and saw **958-1019**.

- Spread 61: the base roll is 429-491 (62), so the bonus part is constant and no healing multiplier applies. This fits
  the talents below: Spiritual Healing 0, Improved Healing 0.
- Bonus part = 958 - 429 = 529 and 1019 - 491 = 528, about **528.5**. Coefficient = 528.5 / 749 = **0.706**
  (classic rule 0.857 would give 1071-1133).
- Nearest neat value is 0.714 (2.5 / 3.5), which would need 740 bonus healing instead of 749 (or 535 at 749, giving
  964-1026). Divine Fury 5/5 makes Heal 2.5 sec, so one idea is that the coefficient uses the shortened cast time, but
  rank 4's 1461 (above 1446 at 0.857, so 0.88 or more) does not fit 0.714 unless it came with different bonus healing.
  Open: rank 4 and rank 1 / 3 data, and whether 749 was exact for these casts.

Talents `320230005002332005-23020500500202` decoded (sim order): Holy Focus 2, Improved Renew 3, Spell Focus 2,
Divine Fury 5, Holy Specialization 5, Purifying Light 2, Spiritual Guidance 2, Spiritual Healing 0, Improved
Healing 0; Discipline part as in the SimExporter string.

SimExporter export of the character used for these tests:

```json
{"exporter":"SimExporter 1.0","name":"Lokiy","realm":"Hogger","level":60,"class":"Priest","race":"Undead","talents":"320230005002332005-23020500500202","professions":[{"name":"Enchanting","level":300},{"name":"Tailoring","level":300}],"gear":{"version":"1.0","items":[{"id":25183,"enchant":2544},{"id":1315},{"id":22082,"enchant":2483},{"id":18510,"enchant":1888},{"id":22083,"enchant":1893},{"id":25176,"enchant":1886},{"id":19369,"enchant":931},{"id":25177,"enchant":2772},{"id":25181,"enchant":2590},{"id":25179,"enchant":929},{"id":19138},{"id":17982,"randomSuffix":409},{"id":26104},{"id":26208},{"id":19360,"enchant":2504},{"id":20582},{"id":19435}]}}
```

### Heal rank 4 test: about 50 casts (Lokiy 2026-11-02)

Heal rank 4, about 50 casts (non-crit, bonus healing 749, same gear and talents as above): **1376-1467**.

- Spread 91; the base roll is 712-804 (92). Again a constant bonus part and no multiplier.
- Bonus part = 1376 - 712 = 664 and 1467 - 804 = 663, about **663.5**. Coefficient = 663.5 / 749 = **0.886**
  (classic 0.857 would give 1354-1446). The earlier single 1461 sits inside this range.
- Heal r2 gave 0.706 and r4 gives 0.886 from the same 749, so the two ranks differ by a factor of 0.80 (r2 / r4).
  The classic below-level-20 rule does not explain that (both ranks are above level 20).
- Lead: Spiritual Guidance is 2/5 in these talents. The sim (`sim/priest/talents.go`) takes it as **10% of Spirit as
  spell power per rank** (20% here, from the DBC). If the 749 shown in game does not include that, the real bonus
  healing is higher than 749 and every coefficient above is a bit lower (r4 would be exactly 0.857 at 774 bonus
  healing, which needs Spirit of about 125 at 20%; the real Spirit is higher, so it would sit below 0.857). Needs the
  Spirit shown in game during the tests to settle.

### Update (Lokiy 2026-11-02): Spiritual Guidance is already in the 749

So the Spiritual Guidance lead above is closed; the measured coefficients stand: **Heal r2 0.706, Heal r4 0.886**.

No known formula gives them (classic: cast time / 3.5, penalty only under level 20, would be 0.857 for both). Simplest
pattern: linear in the rank's level, +0.015 per level (0.706 at level 22, 0.886 at level 34). Predictions to test at
749 bonus healing: **Heal r3** (level 28) about 0.80 -> 1162-1238; **Heal r1** (level 16) about 0.62 -> 756-802
(lower if the below-level-20 penalty applies on top). Matches -> look for a level-based rule; misses -> the
coefficients are per spell and each rank is measured.

### Heal rank 3 test (Lokiy 2026-11-02)

Heal rank 3, non-crit, bonus healing 749 (Spiritual Guidance included): **1190-1265** (so far).

- Spread 75; the base roll is 566-642 (76). Constant bonus part again.
- Bonus part = 1190 - 566 = 624 and 1265 - 642 = 623, about **623.5**. Coefficient = 623.5 / 749 = **0.833**
  (classic 0.857 would give 1208-1284; the linear guess 0.80 gave 1162-1238, so the linear pattern is wrong).

Heal so far (749 bonus healing):

| Rank | Level | Base roll | Seen | Bonus part | Coefficient |
|---|---|---|---|---|---|
| 2 | 22 | 429-491 | 958-1019 | 528.5 | 0.706 |
| 3 | 28 | 566-642 | 1190-1265 | 623.5 | 0.833 |
| 4 | 34 | 712-804 | 1376-1467 | 663.5 | 0.886 |

The coefficient rises with rank but flattens: +0.127 from r2 to r3, +0.053 from r3 to r4. No formula yet. Rank 1
(level 16) would be the fourth point; a rough curve through the three points suggests about 0.5 (673-719).

### Heal rank 1 test (Lokiy 2026-11-02)

Heal rank 1, non-crit, bonus healing 749: **719-765** (so far).

- Spread 46; the base roll is 295-341 (46). Bonus part = 719 - 295 = 424 and 765 - 341 = 424: **424**.
  Coefficient = 424 / 749 = **0.566** (my rough curve guessed 0.5; classic with the level-16 penalty would be 0.729).

Heal, all four ranks (749 bonus healing, Spiritual Guidance included, no healing multipliers):

| Rank | Level | Base roll | Seen | Bonus part | Coefficient | Classic rule |
|---|---|---|---|---|---|---|
| 1 | 16 | 295-341 | 719-765 | 424 | 0.566 | 0.729 |
| 2 | 22 | 429-491 | 958-1019 | 528.5 | 0.706 | 0.857 |
| 3 | 28 | 566-642 | 1190-1265 | 623.5 | 0.833 | 0.857 |
| 4 | 34 | 712-804 | 1376-1467 | 663.5 | 0.886 | 0.857 |

Step between ranks: +0.140, +0.127, +0.053. Close to +0.022 per level for ranks 1-3 (about (level + 9.5) / 45), then
rank 4 falls well short of that line (it would give 0.97) and looks capped near 0.886. Rank 1 / 0.85 (the level-16
penalty) = 0.666, which may be a coincidence. No clean formula yet; the coefficients may be set per rank by the server.
Better source if available: the server's own spell bonus table (a spell_bonus_data-style table in its world database).

### Greater Heal expectations at 749 bonus healing (not measured yet; 2026-11-02)

Greater Heal ranks are levels 40-60, above Heal r4 (level 34, which looked capped at 0.886), so all five ranks are
expected at the cap. Classic rule: 4 sec capped at 3.5 / 3.5 = **1.0**. If the server's cap sits 3.4% above classic like
Heal's did (0.886 / 0.857 = 1.034), the bonus is 774 instead of 749.

| Rank | Level | Base roll | Coefficient 1.0 (+749) | Coefficient 1.034 (+774) |
|---|---|---|---|---|
| 1 | 40 | 1263-1423 | 2012-2172 | 2037-2197 |
| 2 | 46 | 1611-1807 | 2360-2556 | 2385-2581 |
| 3 | 52 | 2011-2252 | 2760-3001 | 2785-3026 |
| 4 | 58 | 2516-2808 | 3265-3557 | 3290-3582 |
| 5 | 60 | 2752-3071 | 3501-3820 | 3526-3845 |

If a rank lands below its row, Greater Heal has a lower coefficient or a rank rule like Heal's low ranks. Test rank 5
first (the one in use), then rank 1.

### Greater Heal rank 5 quick sample: 7 casts (Lokiy 2026-11-02)

Greater Heal r5, 7 casts, bonus healing 749: **3628-3913**. Base roll 2752-3071.

- With a constant bonus b, the real min is 2752 + b and the real max 3071 + b, and a sample cannot go outside them:
  min seen 3628 means b <= 876, max seen 3913 means b >= 842. So the coefficient is between **1.124 and 1.170**.
- That rules out the capped classic value 1.0 (3501-3820) and 1.034 (3526-3845).
- **4 / 3.5 = 1.143** fits: bonus 856, range 3608-3927 (the seen 3628-3913 sits inside it). That is the cast time
  without the 3.5 sec cap, and without Divine Fury shortening it (3.5 / 3.5 would be 1.0).
- Caveat: 7 casts. More casts narrow the range. Greater Heal r4 / r1 not measured yet.
- Open question: Heal r4 came out at 0.886, above the uncapped cast-time value 3 / 3.5 = 0.857, so Heal and Greater Heal
  do not follow one simple rule yet.

### Caveat (Lokiy 2026-11-02): some tests may have run at 741 bonus healing, not 749

A Spirit buff fell off during testing; Spiritual Guidance (20% of Spirit as spell power at 2/5) then gave 8 less, so
some casts had **741** instead of 749. Which tests is not known yet. The bonus part (measured from the spread) stays
the same; only the coefficient (bonus part / bonus healing) changes, by about 1%:

| Spell | Bonus part | Coefficient at 749 | Coefficient at 741 |
|---|---|---|---|
| Heal r1 | 424 | 0.566 | 0.572 |
| Heal r2 | 528.5 | 0.706 | 0.713 (2.5 / 3.5 = 0.714) |
| Heal r3 | 623.5 | 0.833 | 0.841 |
| Heal r4 | 663.5 | 0.886 | 0.895 |
| Greater Heal r5 (7 casts) | 842-876 | 1.124-1.170 | 1.136-1.182 |

Greater Heal r5 at 4 / 3.5 = 1.143 and 741 bonus healing: bonus 847, range 3599-3918, still contains the seen
3628-3913. For future tests: note the bonus healing shown, and test with the Spirit buff up.

### Greater Heal rank 5: 15 casts (Lokiy 2026-11-02; replaces the 7-cast sample above, which is to be ignored)

Greater Heal r5, 15 casts, bonus healing taken as 749 (Spirit buff up): **3614-3919**. Base roll 2752-3071.

- Min seen 3614 means bonus <= 3614 - 2752 = 862; max seen 3919 means bonus >= 3919 - 3071 = 848.
  So the bonus is **848-862** and the coefficient **1.132-1.151** (at 741 bonus healing: 1.144-1.163).
- **4 / 3.5 = 1.143** sits in the middle of that range: bonus 856, range 3608-3927 (seen 3614-3919 is inside it).
  The capped classic 1.0 and the 1.034 guess are ruled out.
- Best single estimate: bonus 855 (middle of 848-862), coefficient about 1.14.

Confirmed by Lokiy: the 15-cast Greater Heal r5 sample (3614-3919) ran at **749** bonus healing, so its coefficient is
**1.132-1.151** (the 741 column above does not apply to it). Which earlier tests (if any) ran at 741 is still unknown.

### Cast times during the tests: Divine Fury 5/5 + Improved Memory 5/5 (Lokiy 2026-11-02)

All Heal tests (r1-r4) ran with Heal at **1.96 sec**; the Greater Heal tests with Greater Heal at **2.75 sec**.
Talents: Divine Fury 5/5 (-0.5 sec on Heal / Greater Heal), Improved Memory 5/5 (25% casting speed in the DBC).

| Spell | DBC cast | - Divine Fury | / 1.25 (sim's model) | Seen in game |
|---|---|---|---|---|
| Heal | 3.0 | 2.5 | 2.00 | 1.96 |
| Greater Heal | 4.0 | 3.5 | 2.80 | 2.75 |

- Both fit "(base - 0.5) x 0.785", i.e. a total casting speed of about **1.274 (27.4%)**, not 1.25. About 2.4% haste
  is unexplained (another haste source on the gear / buffs, or the server rounds differently). The sim has 25%, so its
  Heal / Greater Heal casts are about 2% slow. Open: find the source.
- **The coefficient does not follow the shortened cast time.** Greater Heal r5 fits 4.0 / 3.5 = 1.143 (the DBC cast
  time); the Divine Fury time (3.5 / 3.5 = 1.0) and the hasted time (2.75 / 3.5 = 0.79) are both ruled out. So talents
  and haste do not change a spell's coefficient, which matches the classic rule that it comes from the base cast time.
- Heal does not fit its base cast time (3.0 / 3.5 = 0.857) rank by rank: 0.566 / 0.706 / 0.833 / 0.886. Not explained by
  haste either: 1.96 / 3.5 = 0.56 would fit rank 1 only.

**Correction (Lokiy 2026-11-02): the extra ~2.4% haste is the gloves enchant.** Lokiy has a 2% haste enchant
(Enchant Gloves - Lesser Haste, enchant 931 on the gloves in the export above; the sim models it as 2% spell haste).
Casting speed 1.25 x 1.02 = **1.275**: Heal (3.0 - 0.5) / 1.275 = 1.96 sec and Greater Heal (4.0 - 0.5) / 1.275 =
2.745 sec, which is what the game shows. The sim multiplies enchant haste into the cast speed
(`sim/core/character.go`), so its Heal / Greater Heal cast times already match; the "sim is about 2% slow" note
above does not apply, and the haste gap is not an open item. Coefficient conclusions are unchanged.

### Greater Heal rank 5: updated sample 3614-3926 (Lokiy 2026-11-02, 749 bonus healing)

Min seen 3614 means bonus <= 3614 - 2752 = 862; max seen 3926 means bonus >= 3926 - 3071 = 855. So the bonus is
**855-862** and the coefficient **1.1415-1.1509**. 4 / 3.5 = 1.1429 (bonus 856, range 3608-3927) is inside it; the
seen max is 1 under that range's top.

Next: Greater Heal rank 1 (level 40, base 1263-1423). If the coefficient is also 1.143 (bonus 856) it should show
about **2119-2279**; clearly lower means Greater Heal's low ranks have a rank rule like Heal's.

### Greater Heal rank 1: 20 casts (Lokiy 2026-11-02)

Greater Heal r1 (level 40, base roll 1263-1423), 20 casts, non-crit: **2166-2311**. Bonus healing assumed 749 (to confirm).

- Min seen 2166 means bonus <= 2166 - 1263 = 903; max seen 2311 means bonus >= 2311 - 1423 = 888. So the bonus is
  **888-903** and the coefficient **1.186-1.206** at 749 bonus healing.
- That is above Greater Heal r5 (1.1415-1.151) and above 4 / 3.5 = 1.143 (which would give 2119-2279; seen is
  32-47 higher). Greater Heal r1 and r5 cannot share a coefficient at the same bonus healing.
- If the coefficient were 1.143, the bonus healing during this test would have been **777-790**, not 749 (28-41 more).
  Either the bonus healing was higher during the GH1 test (buff / proc / trinket), or Greater Heal's coefficient falls
  with rank (the opposite of Heal, whose coefficient rises with rank).
- Need: the bonus healing shown during the GH1 test, then a repeat of GH5 or GH1 with the same buffs to compare.

### Greater Heal rank 1: updated sample 2162-2314 (Lokiy 2026-11-02)

Min seen 2162 means bonus <= 2162 - 1263 = 899; max seen 2314 means bonus >= 2314 - 1423 = 891. So the bonus is
**891-899** and the coefficient **1.190-1.200** at 749 bonus healing (1.2 would be bonus 898.8, range 2162-2322).
Still above Greater Heal r5 (1.1415-1.151), so the question stands: higher bonus healing during the rank 1 test
(at 1.143 it would be 780-787) or a coefficient that falls with rank.

### Greater Heal rank 1: updated sample 2158-2314, bonus healing 749 confirmed (Lokiy 2026-11-02)

Min seen 2158 means bonus <= 2158 - 1263 = 895; max seen 2314 means bonus >= 2314 - 1423 = 891. So the bonus is
**891-895** and the coefficient **1.190-1.195** at 749 bonus healing (confirmed 749 for the rank 1 casts; the earlier
"2162-2314" and "2166-2311" samples above are superseded). 1.2 is ruled out (bonus 898.8 > 895).

So Greater Heal's coefficient is about **1.19 at rank 1 and 1.146 at rank 5**: it falls slightly with rank (-4%), the
opposite of Heal.

Every measured rank against its base cast time rule (cast time / 3.5, uncapped: Heal 0.857, Greater Heal 1.143),
at 749 bonus healing (the Heal rows carry the 741 uncertainty, about +1%):

| Spell | Level | Coefficient | / base rule |
|---|---|---|---|
| Heal r1 | 16 | 0.566 | 0.66 |
| Heal r2 | 22 | 0.706 | 0.82 |
| Heal r3 | 28 | 0.833 | 0.97 |
| Heal r4 | 34 | 0.886 | 1.03 |
| Greater Heal r1 | 40 | 1.192 | 1.04 |
| Greater Heal r5 | 60 | 1.146 | 1.00 |

Read against spell level, the ratio rises steeply through the low levels, peaks around 1.04 near levels 34-40, then
settles back to 1.00 at level 60. One smooth level curve times the base cast time fits all six points, but that is
a reading of six points, not a formula. Greater Heal ranks 2-4 (levels 46, 52, 58) and Flash Heal would test it.
