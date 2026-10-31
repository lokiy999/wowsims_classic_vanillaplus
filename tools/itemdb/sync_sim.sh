#!/bin/bash
# Rebuild and redeploy sim.lokiy.dev when db.lokiy.dev's item data (/api/items.json), source data
# (/api/sources.json) or spell data (/api/spells.json, since Part GH) or NPC pages (/api/npcs.json + mobs.json, since
# Part GL; quests.json since Part GO) changed since the sim was last built from them (Lokiy 2026-10-24, CHANGES.md Part FL).
# Started in the background by sync_db.sh and itemcache_sync.py after they rebuild db.lokiy.dev.
#
# Runs the item pipeline (docs/private-server-item-rules.md), checks the result (sim_check.py), builds the sim,
# restarts it, and commits + pushes the generated files. A failed step or a too-big change deploys nothing, resets
# the generated files and is not retried until the data changes again. If the restarted sim doesn't answer, the
# previous server binary and data are put back.
#
#   DRY_RUN=1          stop after the check (nothing deployed, files reset)
#   FORCE=1            run even if the data is unchanged
#   SIM_HEALTH_URL=... URL the restarted sim must answer (default http://127.0.0.1:3333/classic/)
REPO=/opt/wowsims-classic
DB_API=/var/www/db/api
STATE=/home/lokiy/loot-sync
LOG=$STATE/sim-sync.log
BUILT=$STATE/sim-built.sha    # data the deployed sim was built from
FAILED=$STATE/sim-failed.sha  # data a run failed on (not retried)
HEALTH=${SIM_HEALTH_URL:-http://127.0.0.1:3333/classic/}
GENERATED="assets/database/db.bin assets/database/db.json assets/database/leftover_db.bin assets/database/leftover_db.json
 assets/db_inputs/custom_items.json assets/db_inputs/renumber.json assets/db_inputs/included_items.json
 assets/db_inputs/add_items.json assets/db_inputs/item_phases.json assets/db_inputs/item_sources.json
 assets/database/db_icons.json assets/database/db_spell_ids.json assets/spells
 assets/database/db_npcs.json"
export PATH=$HOME/.nvm/versions/node/v20.20.2/bin:$PATH:/usr/local/go/bin:$HOME/go/bin

exec 8>/tmp/sim-sync.lock
flock -n 8 || exit 0
cd "$REPO" || exit 1
log() { echo "$(date -u '+%F %T') $*" >>"$LOG"; }

SHA=$(cat "$DB_API/items.json" "$DB_API/sources.json" "$DB_API/spells.json" "$DB_API/npcs.json" "$DB_API/mobs.json" \
	"$DB_API/quests.json" 2>/dev/null | sha256sum | cut -c1-16)
if [ -z "$FORCE" ]; then
	[ "$SHA" = "$(cat "$BUILT" 2>/dev/null)" ] && exit 0
	[ "$SHA" = "$(cat "$FAILED" 2>/dev/null)" ] && exit 0
fi

# Never build in or over someone's unfinished work (a hand-run pipeline included): every tracked file must match git.
BUSY=$(git status --porcelain --untracked-files=no | awk '{print $2}')
if [ -n "$BUSY" ]; then
	log "data $SHA changed, but the repo has uncommitted work ($(echo $BUSY | head -c 200)); not built"
	exit 0
fi

reset_generated() { git checkout -q -- $GENERATED; }
fail() {
	log "FAILED ($1): nothing deployed, not retried until the data changes"
	reset_generated
	echo "$SHA" >"$FAILED"
	exit 1
}

log "data $SHA: building the sim"
{
	python3 docs/parse_vplus.py && python3 docs/gen_include.py && python3 docs/parse_vplus.py &&
		python3 docs/gen_phases.py && python3 docs/gen_sources.py &&
		python3 tools/gen_db_spells.py && go run ./tools/database/gen_db -outDir=./assets -gen=db &&
		python3 tools/import_icons.py && python3 tools/gen_db_icons.py && python3 tools/gen_db_npcs.py
} >/tmp/sim-sync-pipeline.log 2>&1 || fail "pipeline, see /tmp/sim-sync-pipeline.log"

SUMMARY=$(python3 tools/itemdb/sim_check.py)
RC=$?
echo "$SUMMARY" >>"$LOG"
case $RC in
	1) fail "check" ;;
	2)
		# no item changed; spell data (db_spell_ids.json, assets/spells, spell icons in db.json) may have
		if [ -z "$(git status --porcelain -- $GENERATED)" ]; then
			log "no changes, nothing to deploy"; reset_generated; echo "$SHA" >"$BUILT"; exit 0
		fi
		log "no item changes, but spell / icon data changed" ;;
	0) ;;
	*) fail "check crashed" ;;
esac
if [ -n "$DRY_RUN" ]; then
	log "dry run: not deployed"
	reset_generated
	exit 0
fi

restart_sim() {
	local old new i
	old=$(systemctl show -p MainPID --value wowsims)
	kill "$old"
	for i in $(seq 30); do
		sleep 2
		new=$(systemctl show -p MainPID --value wowsims)
		[ "$new" != 0 ] && [ "$new" != "$old" ] && curl -sf -o /dev/null --max-time 5 "$HEALTH" && return 0
	done
	return 1
}

cp -p wowsimclassic /tmp/wowsimclassic.prev
(make dist/classic/.dirstamp && make devserver) >/tmp/sim-sync-build.log 2>&1 || fail "build, see /tmp/sim-sync-build.log"
if ! restart_sim; then
	log "restarted sim doesn't answer $HEALTH: putting the previous version back"
	reset_generated
	make dist/classic/.dirstamp >/dev/null 2>&1
	cp -p /tmp/wowsimclassic.prev wowsimclassic.new && mv -f wowsimclassic.new wowsimclassic  # a running binary can't be overwritten
	restart_sim && log "previous version running again" || log "PREVIOUS VERSION DOESN'T ANSWER EITHER"
	echo "$SHA" >"$FAILED"
	exit 1
fi

git add -- $GENERATED
git commit -q -m "Item data from db.lokiy.dev (auto, sync_sim.sh)" -m "$SUMMARY" &&
	log "deployed and committed $(git rev-parse --short HEAD)"
git push -q >>"$LOG" 2>&1 || log "push failed (committed on the server only)"
echo "$SHA" >"$BUILT"
