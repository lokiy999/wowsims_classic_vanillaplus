import { UIItem } from '../proto/ui';

// EP for trinkets (Part IK): a trinket gets an EP value when all its effects are plain stats that are in the item's
// stats (e.g. Neltharion's Tear: spell damage and spell hit). One with a Use: effect, a proc, or any other special
// effect (Battle Shout, Corruption, threat, ...) is N/A, since stat weights can't value it.

// The green tooltip lines that are only stats (docs/parse_vplus.py turns them into item stats), or that do nothing
// in a fight (stealth detection, mount / indoor speed).
const STAT_ONLY_LINES: Array<RegExp> = [
	/^Equip: \+\d+ (Ranged )?Attack Power\.$/,
	/^Equip: Restores \d+ (mana|health) per 5 sec\.$/,
	/^Equip: Increases healing done by spells and effects by up to \d+\.$/,
	/^Equip: Increases damage and healing done by magical spells and effects by up to \d+\.$/,
	/^Equip: Increases damage done by (Arcane|Fire|Frost|Holy|Nature|Shadow) spells and effects by up to \d+\.$/,
	/^Equip: Increased Defense \+\d+\.$/,
	/^Equip: Increases your chance to (dodge|parry) an attack by \d+%\.$/,
	/^Equip: Increases your chance to block attacks with a shield by \d+%\.$/,
	/^Equip: Increases the block value of your shield by \d+\.$/,
	/^Equip: Improves your chance to hit( with spells)? by \d+%\.$/,
	/^Equip: Improves your chance to get a critical strike( with spells)? by \d+%\.$/,
	/^Equip: Improves your critical strike chance for all attacks and spells by \d+%\.$/,
	/^Equip: Decreases the magical resistances of your spell targets by \d+\.$/,
	/^Equip: \+\d+ All Resistances\.$/,
	/^Equip: \+\d+ to all attributes\.$/,
	/^Equip: (Slightly )?[Ii]ncreases your stealth detection\.$/,
	/^Equip: Increases mount speed by \d+%\.$/,
	/^Equip: Indoor movement speed increased by \d+%\.$/,
];

const GREEN_LINE = /<div style="color:#1eff00">(.*?)<\/div>/g;

const cache = new Map<number, boolean>();

// Whether stat weights can value this trinket: every green (effect) line in its tooltip is a plain stat.
export function trinketHasOnlyStats(item: UIItem): boolean {
	const cached = cache.get(item.id);
	if (cached !== undefined) return cached;

	let statsOnly = true;
	for (const match of (item.tooltip || '').matchAll(GREEN_LINE)) {
		const text = match[1].replace(/<[^>]*>/g, '').trim();
		if (!STAT_ONLY_LINES.some(re => re.test(text))) {
			statsOnly = false;
			break;
		}
	}
	cache.set(item.id, statsOnly);
	return statsOnly;
}

export const TRINKET_EP_NA_TOOLTIP = 'This trinket has a Use: or proc effect, which stat weights can’t value. Sim it to compare.';
