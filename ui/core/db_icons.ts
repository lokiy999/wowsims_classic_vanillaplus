// Icons come from db.lokiy.dev (extracted from the game client: <name>.jpg 56 px, <name>_s.jpg 18 px); Wowhead's CDN
// only for names it doesn't have (later-expansion / Season of Discovery icons). The name list is
// assets/database/db_icons.json (tools/gen_db_icons.py), loaded together with the database. Until it is loaded, icons
// come from Wowhead.
const DB_ICON_URL = 'https://db.lokiy.dev/icons/';
const DB_ICON_LIST = '/classic/assets/database/db_icons.json';

let dbIcons: Set<string> | null = null;

export async function loadDbIcons(): Promise<void> {
	try {
		const response = await fetch(DB_ICON_LIST);
		if (response.ok) dbIcons = new Set(await response.json());
	} catch {
		// keep the Wowhead icons
	}
}

export function iconUrl(name: string, size: 'small' | 'medium' | 'large' = 'large'): string {
	const icon = name.toLowerCase().replace(/\.(jpg|png)$/, '');
	if (dbIcons?.has(icon)) return `${DB_ICON_URL}${icon}${size === 'small' ? '_s' : ''}.jpg`;
	return `https://wow.zamimg.com/images/wow/icons/${size}/${icon}.jpg`;
}
