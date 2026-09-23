package core

import (
	"math"
	"testing"

	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
	"github.com/wowsims/classic/sim/core/stats"
)

// V+ server: a boss's chance to crit is only reduced by Defense; dodge, parry and block
// don't push crits off the attack table.
func Test_EnemyCritNotPushedByAvoidance(t *testing.T) {
	attacker := &Unit{Type: EnemyUnit, Level: 63, stats: stats.Stats{}}
	defender := &Unit{Type: PlayerUnit, Level: 60, stats: stats.Stats{}, PseudoStats: stats.NewPseudoStats()}
	defender.PseudoStats.CanBlock = true
	defender.PseudoStats.CanParry = true

	// Far more avoidance than there is room for on the table.
	defender.stats[stats.Dodge] = 90
	defender.stats[stats.Parry] = 30
	defender.stats[stats.Block] = 30

	attackTable := NewAttackTable(attacker, defender, nil)

	sim := NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{},
		Encounter:  &proto.Encounter{},
		Raid:       &proto.Raid{},
	}, simsignals.CreateSignals())

	spell := &Spell{Unit: attacker, SpellMetrics: make([]SpellMetrics, 1)}

	const n = 200_000
	crits := 0
	for i := 0; i < n; i++ {
		result := &SpellResult{Target: defender, Damage: 100}
		spell.OutcomeEnemyMeleeWhite(sim, result, attackTable)
		if result.Outcome.Matches(OutcomeCrit) {
			crits++
		}
	}

	missChance := max(0, attackTable.BaseMissChance)
	expected := max(0, min(attackTable.BaseCritChance, 1-missChance))
	got := float64(crits) / n
	if expected <= 0 {
		t.Fatalf("test setup: expected a positive base crit chance, got %.4f", expected)
	}
	if math.Abs(got-expected) > 0.005 {
		t.Errorf("crit rate with 150%% avoidance = %.4f, expected the base crit chance %.4f", got, expected)
	}
}
