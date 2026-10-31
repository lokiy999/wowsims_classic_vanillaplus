#!/bin/sh
# Rebuild db.lokiy.dev when the loot data (lokiy.dev/loot/loot.json, built by the LootTracker addon's
# build_loot.py) is newer than the last build. Run from lokiy's crontab every 5 minutes:
#   */5 * * * * /opt/wowsims-classic/tools/itemdb/sync_db.sh
# First, when a player's upload was approved since loot.json was built (lokiy.dev/admin/loot writes MARK; the
# upload service can't write the web root itself), build loot.json again with it.
LOOT=/var/www/lokiy/loot/loot.json
STAMP=/var/www/db/.loot-built
MARK=/home/lokiy/loot-upload/data/rebuild-requested
LOG=/home/lokiy/loot-sync/db-sync.log
# Item cache uploaded on lokiy.dev/admin/upload: merged and rebuilt by itemcache_sync.py (takes the build lock itself)
python3 /opt/wowsims-classic/tools/itemdb/itemcache_sync.py --pickup >>"$LOG" 2>&1
# Quest cache uploaded there too: merged into ~/questcache/quests.json and rebuilt by questcache_sync.py (Part HK)
python3 /opt/wowsims-classic/tools/itemdb/questcache_sync.py --pickup >>"$LOG" 2>&1
# and the creature cache (creaturecache_sync.py, Part HX)
python3 /opt/wowsims-classic/tools/itemdb/creaturecache_sync.py --pickup >>"$LOG" 2>&1
NEED_LOOT=
[ -f "$MARK" ] && { [ ! -f "$LOOT" ] || [ "$MARK" -nt "$LOOT" ]; } && NEED_LOOT=1
if [ -z "$NEED_LOOT" ]; then
	[ -f "$LOOT" ] || exit 0
	[ -f "$STAMP" ] && [ ! "$LOOT" -nt "$STAMP" ] && exit 0
fi
exec 9>/tmp/db-sync.lock
flock -n 9 || exit 0
if [ -n "$NEED_LOOT" ]; then
	if python3 /home/lokiy/loot-sync/build_loot.py >>"$LOG" 2>&1; then
		echo "$(date -u '+%F %T') loot.json rebuilt after an approved upload" >>"$LOG"
	else
		echo "$(date -u '+%F %T') loot.json rebuild FAILED (tried again in 5 minutes)" >>"$LOG"
		exit 0
	fi
fi
touch "$STAMP.new"
if python3 /opt/wowsims-classic/tools/itemdb/gen_site.py >>"$LOG" 2>&1; then
	mv "$STAMP.new" "$STAMP"
	python3 /home/lokiy/loot-sync/enrich.py >>"$LOG" 2>&1  # lokiy.dev/loot: links to new mob / item pages
	# sim.lokiy.dev: rebuilt in the background when the item or source data changed (without this lock)
	nohup /opt/wowsims-classic/tools/itemdb/sync_sim.sh 9>&- >/dev/null 2>&1 &
	echo "$(date -u '+%F %T') rebuilt after loot.json change" >>"$LOG"
else
	rm -f "$STAMP.new"
	echo "$(date -u '+%F %T') rebuild FAILED" >>"$LOG"
fi
