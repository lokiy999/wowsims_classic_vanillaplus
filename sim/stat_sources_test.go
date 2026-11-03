package sim

import (
	"testing"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

// A stat's own multiplier counts toward its source in the stat tooltip (Part IL): Physical Prowess (2/2, +100%
// Strength) shows as Talents, not "From other stats".
func TestStatMultiplierSource(t *testing.T) {
	player := &proto.Player{
		Name:          "Rogue",
		Race:          proto.Race_RaceHuman,
		Class:         proto.Class_ClassRogue,
		Equipment:     &proto.EquipmentSpec{},
		TalentsString: "-0000000000000002",
		Spec:          &proto.Player_Rogue{Rogue: &proto.Rogue{Options: &proto.RogueOptions{}}},
		Rotation:      &proto.APLRotation{},
	}
	res := core.ComputeStats(&proto.ComputeStatsRequest{
		Raid: &proto.Raid{Parties: []*proto.Party{{Players: []*proto.Player{player}}}},
	})
	ps := res.RaidStats.Parties[0].Players[0]
	gear := ps.GearStatsRaw.Stats[stats.Strength]
	talents := ps.TalentsStatsRaw.Stats[stats.Strength]
	final := ps.ConsumesStats.Stats[stats.Strength]
	if gear <= 0 || talents != 2*gear || final != talents {
		t.Errorf("Strength: gear %v, talents %v (want %v), final %v", gear, talents, 2*gear, final)
	}
}
