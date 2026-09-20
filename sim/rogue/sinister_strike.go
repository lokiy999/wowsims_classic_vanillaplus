package rogue

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

func (rogue *Rogue) registerSinisterStrikeSpell() {

	flatDamageBonus := map[int32]float64{
		25: 15,
		40: 33,
		50: 52,
		60: 68,
	}[rogue.Level]

	spellID := map[int32]int32{
		25: 1759,
		40: 8621,
		50: 11293,
		60: 11294,
	}[rogue.Level]

	// Improved Sinister Strike (DBC 13732/13863): 3%/5% chance to land an extra Sinister Strike on the same target.
	var extraStrike *core.Spell
	if rogue.Talents.ImprovedSinisterStrike > 0 {
		extraStrike = rogue.RegisterSpell(core.SpellConfig{
			ActionID:    core.ActionID{SpellID: spellID}.WithTag(2),
			SpellSchool: core.SpellSchoolPhysical,
			DefenseType: core.DefenseTypeMelee,
			ProcMask:    core.ProcMaskMeleeMHSpecial,
			Flags:       core.SpellFlagMeleeMetrics | core.SpellFlagNoOnCastComplete,

			CritDamageBonus: rogue.lethality(),

			DamageMultiplier: []float64{1, 1.05, 1.10, 1.15, 1.20, 1.25}[rogue.Talents.Aggression],
			ThreatMultiplier: 1,
			BonusCoefficient: 1,

			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				baseDamage := flatDamageBonus + spell.Unit.MHNormalizedWeaponDamage(sim, spell.MeleeAttackPower())
				spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeWeaponSpecialHitAndCrit)
			},
		})
	}
	extraStrikeChance := []float64{0, 0.03, 0.05}[rogue.Talents.ImprovedSinisterStrike]

	rogue.SinisterStrike = rogue.RegisterSpell(core.SpellConfig{
		SpellCode:   SpellCode_RogueSinisterStrike,
		ActionID:    core.ActionID{SpellID: spellID},
		SpellSchool: core.SpellSchoolPhysical,
		DefenseType: core.DefenseTypeMelee,
		ProcMask:    core.ProcMaskMeleeMHSpecial,
		Flags:       rogue.builderFlags(),

		EnergyCost: core.EnergyCostOptions{
			Cost:   45, // DBC: Improved Sinister Strike is an extra-hit proc, not a cost reduction
			Refund: 0.8,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: time.Second,
			},
			IgnoreHaste: true,
		},

		CritDamageBonus: rogue.lethality(),

		DamageMultiplier: []float64{1, 1.05, 1.10, 1.15, 1.20, 1.25}[rogue.Talents.Aggression], // DBC: 5%/rank
		ThreatMultiplier: 1,
		BonusCoefficient: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			rogue.BreakStealth(sim)

			baseDamage := flatDamageBonus + spell.Unit.MHNormalizedWeaponDamage(sim, spell.MeleeAttackPower())
			result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeWeaponSpecialHitAndCrit)

			if result.Landed() {
				rogue.AddComboPoints(sim, 1, target, spell.ComboPointMetrics())
				if extraStrike != nil && sim.Proc(extraStrikeChance, "Improved Sinister Strike") {
					extraStrike.Cast(sim, target)
				}
			} else {
				spell.IssueRefund(sim)
			}
		},
	})
}
