// NPC pages on db.lokiy.dev (assets/database/db_npcs.json via tools/gen_db_npcs.py, from its api/npcs.json and
// api/mobs.json), loaded together with the database. NPCs it doesn't have keep Wowhead's link.
const DB_NPCS = '/classic/assets/database/db_npcs.json';
const DB_SITE = 'https://db.lokiy.dev';

let byId: Record<string, string> = {};
let byName: Record<string, string> = {};

export async function loadDbNpcs(): Promise<void> {
	try {
		const response = await fetch(DB_NPCS);
		if (!response.ok) return;
		const data = await response.json();
		byId = data.ids ?? {};
		byName = data.names ?? {};
	} catch {
		// keep Wowhead's NPC links
	}
}

// db.lokiy.dev's page for this creature id, if it has one.
export function dbNpcUrl(npcId: number): string | undefined {
	const page = byId[String(npcId)];
	return page && (page.startsWith('/') ? DB_SITE + page : page);
}

// db.lokiy.dev's page for a source shown by name only (a boss without a creature id, the server's own bosses).
export function dbNpcUrlByName(name: string): string | undefined {
	const page = byName[name];
	return page && (page.startsWith('/') ? DB_SITE + page : page);
}
