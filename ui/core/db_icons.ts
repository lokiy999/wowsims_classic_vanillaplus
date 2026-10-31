// Icons come from db.lokiy.dev (<name>.jpg 56 px, <name>_s.jpg 18 px): the game client's icons plus every other icon
// name the sim's DB and UI use (tools/import_icons.py adds those to its set). The name list
// (assets/database/db_icons.json, tools/gen_db_icons.py) is built into the page, so it's known from the first icon on.
// Only a name it doesn't list (an item or spell the DB doesn't have, looked up on Wowhead while the page runs) still
// shows Wowhead's icon.
import DB_ICONS from '../../assets/database/db_icons.json';

const DB_ICON_URL = 'https://db.lokiy.dev/icons/';
const dbIcons = new Set<string>(DB_ICONS);

export function iconUrl(name: string, size: 'small' | 'medium' | 'large' = 'large'): string {
	const icon = name.toLowerCase().replace(/\.(jpg|png)$/, '');
	if (dbIcons.has(icon)) return `${DB_ICON_URL}${icon}${size === 'small' ? '_s' : ''}.jpg`;
	return `https://wow.zamimg.com/images/wow/icons/${size}/${icon}.jpg`;
}
