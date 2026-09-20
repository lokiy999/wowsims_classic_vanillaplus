package paladin

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

// Crusader Strike (spell 33487, level 40): 151 mana, 8 sec cooldown, on the global cooldown. A melee strike for weapon damage
// (nothing extra) that can crit, and it consecrates the weapon (Consecrated Arms, 33488): +5% attack speed per stack for
// 15 sec, up to 5 stacks. All confirmed in game.
func (paladin *Paladin) registerCrusaderStrike() {
	if paladin.Level < 40 {
		return
	}

	actionID := core.ActionID{SpellID: 33487}

	arms := paladin.RegisterAura(core.Aura{
		Label:     "Consecrated Arms",
		ActionID:  core.ActionID{SpellID: 33488},
		Duration:  time.Second * 15,
		MaxStacks: 5,
		OnStacksChange: func(aura *core.Aura, sim *core.Simulation, oldStacks int32, newStacks int32) {
			paladin.MultiplyAttackSpeed(sim, (1+0.05*float64(newStacks))/(1+0.05*float64(oldStacks)))
		},
	})

	paladin.crusaderStrike = paladin.RegisterSpell(core.SpellConfig{
		SpellCode:   SpellCode_PaladinCrusaderStrike,
		ActionID:    actionID,
		SpellSchool: core.SpellSchoolPhysical,
		DefenseType: core.DefenseTypeMelee,
		ProcMask:    core.ProcMaskMeleeMHSpecial,
		Flags:       core.SpellFlagMeleeMetrics | core.SpellFlagAPL,

		ManaCost: core.ManaCostOptions{
			FlatCost: 151,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			CD: core.Cooldown{
				Timer:    paladin.NewTimer(),
				Duration: time.Second * 8,
			},
		},
		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return paladin.DistanceFromTarget <= core.MaxMeleeAttackDistance
		},

		DamageMultiplier: paladin.getWeaponSpecializationModifier(),
		ThreatMultiplier: 1,
		BonusCoefficient: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := spell.Unit.MHWeaponDamage(sim, spell.MeleeAttackPower())
			result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeWeaponSpecialHitAndCrit)
			if result.Landed() {
				arms.Activate(sim)
				arms.AddStack(sim)
			}
		},
	})

	if paladin.Options.IsUsingCrusaderStrikeStopAttack {
		paladin.crusaderStrike.Flags |= core.SpellFlagBatchStopAttackMacro
	}
}
