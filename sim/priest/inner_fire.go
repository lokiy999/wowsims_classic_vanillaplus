package priest

import (
	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/stats"
)

// Inner Fire (DBC ids 588/7128/602/1006/10951/10952): +armor and 20 charges, one charge is
// removed by each melee or ranged hit taken. Improved Inner Fire adds 10/20/30% (DBC 9/19/29 + 1)
// to both the armor bonus and the number of charges.
// The sim does not recast it, so it is applied as a permanent buff when the option is selected.
var innerFireRankLevels = [...]int32{12, 20, 30, 40, 50, 60}
var innerFireRankSpellIDs = [...]int32{588, 7128, 602, 1006, 10951, 10952}
var innerFireRankArmor = [...]float64{315, 495, 720, 945, 1170, 1395}

const innerFireBaseCharges = 20

func (priest *Priest) ApplyInnerFire() {
	rank := -1
	for i, level := range innerFireRankLevels {
		if priest.Level >= level {
			rank = i
		}
	}
	if rank < 0 {
		return
	}

	talentBonus := 0.10 * float64(priest.Talents.ImprovedInnerFire)
	armor := innerFireRankArmor[rank] * (1 + talentBonus)
	charges := int32(innerFireBaseCharges * (1 + talentBonus))

	aura := priest.RegisterAura(core.Aura{
		ActionID:   core.ActionID{SpellID: innerFireRankSpellIDs[rank]},
		Label:      "Inner Fire",
		BuildPhase: core.CharacterBuildPhaseBuffs,
		Duration:   core.NeverExpires,
		MaxStacks:  charges,
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			priest.AddStatDynamic(sim, stats.Armor, armor)
			aura.SetStacks(sim, charges)
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			priest.AddStatDynamic(sim, stats.Armor, -armor)
		},
		OnSpellHitTaken: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if result.Landed() && spell.ProcMask.Matches(core.ProcMaskMeleeOrRanged) {
				aura.RemoveStack(sim)
			}
		},
	})
	core.MakePermanent(aura)
}
