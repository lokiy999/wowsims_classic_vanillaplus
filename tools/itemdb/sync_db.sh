#!/bin/sh
# Rebuild db.lokiy.dev when Lokiy's loot data (lokiy.dev/loot/loot.json, uploaded by the LootTracker addon's
# publish_site.py) is newer than the last build. Run from lokiy's crontab every 5 minutes:
#   */5 * * * * /opt/wowsims-classic/tools/itemdb/sync_db.sh
LOOT=/var/www/lokiy/loot/loot.json
STAMP=/var/www/db/.loot-built
LOG=/home/lokiy/loot-sync/db-sync.log
[ -f "$LOOT" ] || exit 0
[ -f "$STAMP" ] && [ ! "$LOOT" -nt "$STAMP" ] && exit 0
exec 9>/tmp/db-sync.lock
flock -n 9 || exit 0
touch "$STAMP.new"
if python3 /opt/wowsims-classic/tools/itemdb/gen_site.py >>"$LOG" 2>&1; then
	mv "$STAMP.new" "$STAMP"
	echo "$(date -u '+%F %T') rebuilt after loot.json change" >>"$LOG"
else
	rm -f "$STAMP.new"
	echo "$(date -u '+%F %T') rebuild FAILED" >>"$LOG"
fi
