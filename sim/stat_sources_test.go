package sim

import (
	"math"
	"testing"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

func computePlayerStats(t *testing.T, player *proto.Player) *proto.PlayerStats {
	t.Helper()
	player.Equipment = &proto.EquipmentSpec{}
	player.Rotation = &proto.APLRotation{}
	res := core.ComputeStats(&proto.ComputeStatsRequest{
		Raid: &proto.Raid{Parties: []*proto.Party{{Players: []*proto.Player{player}}}},
	})
	return res.RaidStats.Parties[0].Players[0]
}

// Multipliers and conversions count toward the source that adds them in the stat tooltip (Part IL): talents under
// Talents, not "From other stats".
func TestStatDependencySource(t *testing.T) {
	near := func(a, b float64) bool { return math.Abs(a-b) < 1e-6 }

	// Rogue: Physical Prowess 2/2 (+100% Strength), Deadliness 5/5 (+20% Attack Power).
	ps := computePlayerStats(t, &proto.Player{
		Name:          "Rogue",
		Race:          proto.Race_RaceHuman,
		Class:         proto.Class_ClassRogue,
		TalentsString: "-0000000000000002-000000000000000005",
		Spec:          &proto.Player_Rogue{Rogue: &proto.Rogue{Options: &proto.RogueOptions{}}},
	})
	gearStr := ps.GearStatsRaw.Stats[stats.Strength]
	talentsStr := ps.TalentsStatsRaw.Stats[stats.Strength] - gearStr
	if gearStr <= 0 || talentsStr != gearStr || ps.ConsumesStats.Stats[stats.Strength] != 2*gearStr {
		t.Errorf("Strength: gear %v, talents %v (want %v), final %v", gearStr, talentsStr, gearStr, ps.ConsumesStats.Stats[stats.Strength])
	}
	finalAP := ps.ConsumesStats.Stats[stats.AttackPower]
	talentsAP := ps.TalentsStatsRaw.Stats[stats.AttackPower] - ps.GearStatsRaw.Stats[stats.AttackPower]
	if !near(talentsAP, finalAP-finalAP/1.2) {
		t.Errorf("Attack Power from Deadliness: %v, want %v (final %v)", talentsAP, finalAP-finalAP/1.2, finalAP)
	}

	// Priest: Spiritual Guidance 2/2 (spell damage +20% of Spirit).
	ps = computePlayerStats(t, &proto.Player{
		Name:          "Priest",
		Race:          proto.Race_RaceUndead,
		Class:         proto.Class_ClassPriest,
		TalentsString: "-00000000000002",
		Spec:          &proto.Player_ShadowPriest{ShadowPriest: &proto.ShadowPriest{Options: &proto.ShadowPriest_Options{}}},
	})
	spirit := ps.ConsumesStats.Stats[stats.Spirit]
	talentsSP := ps.TalentsStatsRaw.Stats[stats.SpellPower] - ps.GearStatsRaw.Stats[stats.SpellPower]
	if spirit <= 0 || !near(talentsSP, 0.2*spirit) {
		t.Errorf("Spell power from Spiritual Guidance: %v, want %v", talentsSP, 0.2*spirit)
	}

	// Every phase's attributed stats plus "From other stats" add up to the final stats.
	for s := range ps.ConsumesStats.Stats {
		if ps.ConsumesStatsRaw.Stats[s] > ps.ConsumesStats.Stats[s]+1e-6 && !stats.IsPrimaryStat(stats.Stat(s)) {
			t.Errorf("%s: attributed %v is more than the final %v", stats.Stat(s).StatName(), ps.ConsumesStatsRaw.Stats[s], ps.ConsumesStats.Stats[s])
		}
	}
	for _, p := range stats.PrimaryStats {
		if ps.ConsumesStatsRaw.Stats[p] != ps.ConsumesStats.Stats[p] {
			t.Errorf("%s: attributed %v, final %v", p.StatName(), ps.ConsumesStatsRaw.Stats[p], ps.ConsumesStats.Stats[p])
		}
	}
}
