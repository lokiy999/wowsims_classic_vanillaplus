package priest

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

// Power Word: Requital (Vanilla+ custom spell, granted by the talent of the same name).
// Instant, 20s cooldown. Damage ranges are from the server tooltip; the "feared/stunned/
// incapacitated" clause deals the same damage on every rank, so it is not modeled separately.
const RequitalRanks = 4

var RequitalSpellId = [RequitalRanks + 1]int32{0, 33808, 33809, 33810, 33811}
var RequitalBaseDamage = [RequitalRanks + 1][]float64{{0}, {395, 419}, {539, 573}, {661, 705}, {739, 793}}
var RequitalManaCost = [RequitalRanks + 1]float64{0, 150, 200, 260, 340}
var RequitalLevel = [RequitalRanks + 1]int{0, 30, 40, 50, 60}

// TODO: spell power coefficient is unknown (asked the user); 1.5/3.5 is a placeholder for an instant spell.
const RequitalSpellCoef = 1.5 / 3.5

func (priest *Priest) registerRequitalSpell() {
	if !priest.Talents.PowerWordRequital {
		return
	}

	cdTimer := priest.NewTimer()
	for rank := 1; rank <= RequitalRanks; rank++ {
		if RequitalLevel[rank] > int(priest.Level) {
			continue
		}
		low := RequitalBaseDamage[rank][0]
		high := RequitalBaseDamage[rank][1]

		priest.GetOrRegisterSpell(core.SpellConfig{
			ActionID:    core.ActionID{SpellID: RequitalSpellId[rank]},
			SpellSchool: core.SpellSchoolHoly,
			DefenseType: core.DefenseTypeMagic,
			ProcMask:    core.ProcMaskSpellDamage,
			Flags:       SpellFlagPriest | core.SpellFlagAPL,

			RequiredLevel: RequitalLevel[rank],
			Rank:          rank,

			ManaCost: core.ManaCostOptions{
				FlatCost: RequitalManaCost[rank],
			},
			Cast: core.CastConfig{
				DefaultCast: core.Cast{
					GCD: core.GCDDefault,
				},
				CD: core.Cooldown{
					Timer:    cdTimer,
					Duration: time.Second * 20,
				},
			},

			BonusCoefficient: RequitalSpellCoef,

			DamageMultiplier: 1,
			ThreatMultiplier: 1,

			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				baseDamage := sim.Roll(low, high)
				spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)
			},
		})
	}
}
