package sim

import (
	"math"
	"testing"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
	"github.com/wowsims/classic/sim/core/stats"
)

// Primary stats are rounded down after percentage bonuses, like the server (Part II). Checked against Lokiy's Undead
// shadow priest in game (2026-11-01, BetterCharacterStats): Intellect 297, Mana 5551, Health 4117, Spell Crit 5.79%.
func TestPrimaryStatsRoundedDown(t *testing.T) {
	gear := []*proto.ItemSpec{{Id: 26103, Enchant: 2544}, {Id: 18814}, {Id: 17622, Enchant: 2726}, {Id: 26019, Enchant: 2621},
		{Id: 19682, Enchant: 1891}, {Id: 21186, Enchant: 2775}, {Id: 17620, Enchant: 2769}, {Id: 19388, Enchant: 2772},
		{Id: 19133, Enchant: 2544}, {Id: 19131, Enchant: 929}, {Id: 19434}, {Id: 19147}, {Id: 26175}, {Id: 26059},
		{Id: 19360, Enchant: 2504}, {Id: 19891}, {Id: 19861}}
	for _, it := range gear {
		if _, ok := core.ItemsByID[it.Id]; !ok {
			t.Skipf("item %d is not in the sim DB", it.Id)
		}
	}
	player := &proto.Player{
		Name:          "Lokiy",
		Race:          proto.Race_RaceUndead,
		Class:         proto.Class_ClassPriest,
		Equipment:     &proto.EquipmentSpec{Items: gear},
		TalentsString: "320230005002332005-23020500500202",
		Spec:          &proto.Player_ShadowPriest{ShadowPriest: &proto.ShadowPriest{Options: &proto.ShadowPriest_Options{}}},
		Rotation:      &proto.APLRotation{},
	}
	rsr := &proto.RaidSimRequest{
		Raid:       &proto.Raid{Parties: []*proto.Party{{Players: []*proto.Player{player}}}},
		Encounter:  &proto.Encounter{Duration: 60, Targets: []*proto.Target{{Level: 63}}},
		SimOptions: &proto.SimOptions{Iterations: 1, IsTest: true, RandomSeed: 1},
	}
	sim := core.NewSim(rsr, simsignals.CreateSignals())
	sim.Reset()
	char := sim.Raid.Parties[0].Players[0].GetCharacter()
	s := char.GetStats()

	for _, c := range []struct {
		stat stats.Stat
		want float64
	}{
		{stats.Intellect, 297},
		{stats.Stamina, 290},
		{stats.Spirit, 186},
		{stats.Mana, 5551},
		{stats.Health, 4117},
	} {
		if s[c.stat] != c.want {
			t.Errorf("%s: %v, want %v (in game)", c.stat.StatName(), s[c.stat], c.want)
		}
	}
	if crit := s[stats.SpellCrit] / core.SpellCritRatingPerCritChance; math.Abs(crit-5.79) > 0.005 {
		t.Errorf("spell crit %.3f%%, want 5.79%% (in game)", crit)
	}
	// A buff gained in combat rounds the same way: +31 Intellect (Arcane Intellect) gives floor((259 + 31) * 1.15) =
	// 333, and taking it off again gives 297 back.
	char.AddStatsDynamic(sim, stats.Stats{stats.Intellect: 31})
	if got := char.GetStat(stats.Intellect); got != 333 {
		t.Errorf("Intellect with +31: %v, want 333", got)
	}
	char.AddStatsDynamic(sim, stats.Stats{stats.Intellect: -31})
	if got := char.GetStat(stats.Intellect); got != 297 {
		t.Errorf("Intellect after the buff: %v, want 297", got)
	}
	if got := char.GetStat(stats.Mana); got != 5551 {
		t.Errorf("Mana after the buff: %v, want 5551", got)
	}

	for _, p := range stats.PrimaryStats {
		if s[p] != math.Floor(s[p]) {
			t.Errorf("%s is not a whole number: %v", p.StatName(), s[p])
		}
	}
}
