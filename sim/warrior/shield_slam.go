package warrior

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

func (warrior *Warrior) registerShieldSlamSpell() {
	// Shield Slam is no longer talent-gated in the custom tree; register it for all warriors.

	spellID := int32(23925)
	damageLow := 394.0 // server (Spell.csv): 394 to 427
	damageHigh := 427.0
	threat := 254.0

	apCoef := 0.15

	warrior.ShieldSlam = warrior.RegisterSpell(AnyStance, core.SpellConfig{
		SpellCode:   SpellCode_WarriorShieldSlam,
		ActionID:    core.ActionID{SpellID: spellID},
		SpellSchool: core.SpellSchoolPhysical,
		DefenseType: core.DefenseTypeMelee,
		ProcMask:    core.ProcMaskMeleeMHSpecial, // TODO really?
		Flags:       core.SpellFlagMeleeMetrics | core.SpellFlagAPL | SpellFlagOffensive,

		RageCost: core.RageCostOptions{
			Cost:   20,
			Refund: 0.8,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			IgnoreHaste: true,
			CD: core.Cooldown{
				Timer:    warrior.NewTimer(),
				Duration: time.Second * 6,
			},
		},
		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return warrior.PseudoStats.CanBlock
		},

		CritDamageBonus: warrior.impale(),

		DamageMultiplier: 1 + 0.05*float64(warrior.Talents.ShieldAssault), // DBC: Shield Assault +5%/rank
		ThreatMultiplier: 1,
		FlatThreatBonus:  threat * 2,
		BonusCoefficient: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			damage := sim.Roll(damageLow, damageHigh) + warrior.BlockValue()*2 + apCoef*spell.MeleeAttackPower()
			result := spell.CalcAndDealDamage(sim, target, damage, spell.OutcomeMeleeSpecialHitAndCrit)

			if !result.Landed() {
				spell.IssueRefund(sim)
			}
		},
	})
}
