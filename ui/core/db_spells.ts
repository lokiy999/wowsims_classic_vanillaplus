import { IconData } from './proto/ui.js';

// Spells come from db.lokiy.dev (api/spells.json via tools/gen_db_spells.py): their pages at db.lokiy.dev/spell/<id>,
// name, icon and tooltip from assets/spells/<id // 1000>.json. Wowhead only for spells it doesn't have (Season of
// Discovery / later ones). The id list is loaded together with the database; until then spells count as missing.
export const DB_SPELL_URL = 'https://db.lokiy.dev/spell/';
const DB_SPELL_IDS = '/classic/assets/database/db_spell_ids.json';
const DB_SPELL_CHUNK = '/classic/assets/spells/';

let dbSpellIds: Set<number> | null = null;
const chunks = new Map<number, Promise<Record<string, IconData>>>();

export async function loadDbSpellIds(): Promise<void> {
	try {
		const response = await fetch(DB_SPELL_IDS);
		if (response.ok) dbSpellIds = new Set(await response.json());
	} catch {
		// keep Wowhead's spell data
	}
}

export function isDbSpell(spellId: number): boolean {
	return !!dbSpellIds?.has(spellId);
}

export async function getDbSpell(spellId: number): Promise<IconData | null> {
	if (!isDbSpell(spellId)) return null;
	const n = Math.floor(spellId / 1000);
	if (!chunks.has(n)) {
		chunks.set(
			n,
			fetch(`${DB_SPELL_CHUNK}${n}.json`)
				.then(response => (response.ok ? response.json() : {}))
				.catch(() => ({})),
		);
	}
	const data = (await chunks.get(n)!)[String(spellId)];
	return data ? IconData.create({ ...data, id: spellId }) : null;
}
