package rogue

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

func (rogue *Rogue) registerExposeArmorSpell() {
	rogue.ExposeArmorAuras = rogue.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		return core.ExposeArmorAura(target, int32(0) /*removed*/)
	})

	spellID := map[int32]int32{
		25: 8647,
		40: 8650,
		50: 11197,
		60: 11198,
	}[rogue.Level]

	arpenPerCombo := map[int32]float64{
		25: 80,
		40: 210,
		50: 275,
		60: 340,
	}[rogue.Level]

	arpenPerCombo *= []float64{1, 1.25, 1.5}[int32(0) /*removed*/]

	// share ExtraCastCondition() state with ApplyEffects()
	var arpen float64
	var eaAura *core.Aura

	rogue.ExposeArmor = rogue.RegisterSpell(core.SpellConfig{
		SpellCode:    SpellCode_RogueExposeArmor,
		ActionID:     core.ActionID{SpellID: spellID},
		SpellSchool:  core.SpellSchoolPhysical,
		DefenseType:  core.DefenseTypeMelee,
		ProcMask:     core.ProcMaskMeleeMHSpecial,
		Flags:        rogue.finisherFlags(),
		MetricSplits: 6,

		EnergyCost: core.EnergyCostOptions{
			Cost:   25,
			Refund: 0,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: time.Second,
			},
			IgnoreHaste: true,
			ModifyCast: func(sim *core.Simulation, spell *core.Spell, cast *core.Cast) {
				spell.SetMetricsSplit(spell.Unit.ComboPoints())
			},
		},
		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			if rogue.ComboPoints() == 0 {
				return false
			}

			eaAura = rogue.ExposeArmorAuras.Get(target)
			arpen = float64(rogue.ComboPoints()) * arpenPerCombo

			if curActive := eaAura.ExclusiveEffects[0].Category.GetActiveEffect(); curActive != nil {
				return arpen >= curActive.Priority
			}
			return true
		},

		ThreatMultiplier: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			rogue.BreakStealth(sim)

			result := spell.CalcOutcome(sim, target, spell.OutcomeMeleeSpecialHit)
			if result.Landed() {
				eaAura.ExclusiveEffects[0].Priority = arpen
				// Server: 9/12/15/18/21 sec by combo points (Spell.csv 11198), longer with Exhaustion.
				eaAura.Duration = time.Duration(float64(time.Second*time.Duration(6+3*rogue.ComboPoints())) * rogue.exhaustionMultiplier())
				eaAura.Activate(sim)
				rogue.SpendComboPoints(sim, spell)
			} else {
				spell.IssueRefund(sim)
			}
			spell.DealOutcome(sim, result)
		},

		RelatedAuras: []core.AuraArray{rogue.ExposeArmorAuras},
	})
	rogue.Finishers = append(rogue.Finishers, rogue.ExposeArmor)
}
