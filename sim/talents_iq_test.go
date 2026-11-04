package sim

import (
	"strconv"
	"testing"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
	"google.golang.org/protobuf/encoding/protojson"
)

// castOnly runs one iteration of a player who only casts spellID, and returns that spell's casts, landed hits and
// damage.
func castOnly(t *testing.T, player *proto.Player, spellID int32, duration float64) (casts, hits int32, damage float64) {
	t.Helper()
	rotation := &proto.APLRotation{}
	apl := `{"type":"TypeAPL","priorityList":[{"action":{"castSpell":{"spellId":{"spellId":` + strconv.Itoa(int(spellID)) + `}}}}]}`
	if err := protojson.Unmarshal([]byte(apl), rotation); err != nil {
		t.Fatal(err)
	}
	player.Equipment = &proto.EquipmentSpec{}
	player.Rotation = rotation
	rsr := &proto.RaidSimRequest{
		Raid:       &proto.Raid{Parties: []*proto.Party{{Players: []*proto.Player{player}}}},
		Encounter:  &proto.Encounter{Duration: duration, Targets: []*proto.Target{{Level: 63}}},
		SimOptions: &proto.SimOptions{Iterations: 1, IsTest: true, RandomSeed: 7},
	}
	res := core.RunRaidSim(rsr)
	if res.Error != nil {
		t.Fatal(res.Error.Message)
	}
	for _, action := range res.RaidMetrics.Parties[0].Players[0].Actions {
		if action.Id.GetSpellId() != spellID {
			continue
		}
		for _, target := range action.Targets {
			casts += target.Casts
			hits += target.Hits + target.Crits + target.Ticks + target.CritTicks
			damage += target.Damage
		}
	}
	return
}

// Part IN: talents with server data that the sim didn't read.
func TestTalentsPartIN(t *testing.T) {
	// Ice Shards with Cold Grip 2/2: 5 sec cooldown, four shards of 60 (level) Frost damage, +30%.
	casts, hits, damage := castOnly(t, &proto.Player{
		Name: "Mage", Race: proto.Race_RaceHuman, Class: proto.Class_ClassMage,
		TalentsString: "--000000000000201",
		Spec:          &proto.Player_Mage{Mage: &proto.Mage{Options: &proto.Mage_Options{}}},
	}, 35714, 20)
	if casts != 5 || hits == 0 || damage/float64(hits) < 60*1.3*0.7 || damage/float64(hits) > 60*1.3*2 {
		t.Errorf("Ice Shards: %d casts (want 5: at 0, 5, 10, 15 and 20 sec), %d hits, %.0f damage", casts, hits, damage)
	}

	// Death and Decay: 15 ticks of 60 (level) Shadow damage every 2 sec; 40 sec fight, so all of them.
	casts, hits, damage = castOnly(t, &proto.Player{
		Name: "Warlock", Race: proto.Race_RaceHuman, Class: proto.Class_ClassWarlock,
		TalentsString: "0000000000000001",
		Spec:          &proto.Player_Warlock{Warlock: &proto.Warlock{Options: &proto.WarlockOptions{}}},
	}, 33926, 40)
	if casts != 1 || hits != 15 || damage < 15*60*0.7 || damage > 15*60 {
		t.Errorf("Death and Decay: %d casts (want 1), %d ticks (want 15), %.0f damage (at most 900)", casts, hits, damage)
	}

	// Hurricane (Moonkin): a 2 sec cast, then 10 ticks of 134 Nature damage (+ spell damage), once per 2 min.
	casts, hits, damage = castOnly(t, &proto.Player{
		Name: "Druid", Race: proto.Race_RaceTauren, Class: proto.Class_ClassDruid,
		TalentsString: "00000001000001",
		Spec:          &proto.Player_BalanceDruid{BalanceDruid: &proto.BalanceDruid{Options: &proto.BalanceDruid_Options{}}},
	}, 17402, 30)
	if casts != 1 || hits != 10 || damage < 10*134 {
		t.Errorf("Hurricane: %d casts (want 1), %d ticks (want 10), %.0f damage (at least 1340)", casts, hits, damage)
	}
}

// Master Conjuror 2/2: Spellstone's +1% spell crit becomes +1.5%.
func TestMasterConjurorSpellstone(t *testing.T) {
	spellCrit := func(talents string) float64 {
		ps := computePlayerStats(t, &proto.Player{
			Name: "Warlock", Race: proto.Race_RaceHuman, Class: proto.Class_ClassWarlock,
			TalentsString: talents,
			Spec: &proto.Player_Warlock{Warlock: &proto.Warlock{Options: &proto.WarlockOptions{
				WeaponImbue: proto.WarlockOptions_Spellstone,
			}}},
		})
		return ps.FinalStats.Stats[stats.SpellCrit] / core.SpellCritRatingPerCritChance
	}
	if diff := spellCrit("-000002") - spellCrit(""); diff < 0.499 || diff > 0.501 {
		t.Errorf("Master Conjuror 2/2 adds %.3f%% spell crit to Spellstone, want 0.5%%", diff)
	}
}
