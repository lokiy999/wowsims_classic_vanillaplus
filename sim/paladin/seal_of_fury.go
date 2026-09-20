package paladin

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

// Seal of Fury (talent, spell 20423): 8% of base mana, lasts 2 min. Melee attacks deal an additional 30% of normal weapon
// damage as physical damage to up to 3 enemies in front of the paladin. Its Judgement forces the target to attack the
// paladin, which has no damage and is not modeled.
func (paladin *Paladin) registerSealOfFury() {
	if !paladin.Talents.SealOfFury {
		return
	}

	actionID := core.ActionID{SpellID: 20423}

	judgeSpell := paladin.RegisterSpell(core.SpellConfig{
		SpellCode:   SpellCode_PaladinJudgementOfFury,
		ActionID:    actionID.WithTag(1),
		SpellSchool: core.SpellSchoolHoly,
		ProcMask:    core.ProcMaskEmpty,
		Flags:       core.SpellFlagNoOnCastComplete | core.SpellFlagPassiveSpell,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.CalcAndDealOutcome(sim, target, spell.OutcomeAlwaysHit)
		},
	})

	procSpell := paladin.RegisterSpell(core.SpellConfig{
		ActionID:    core.ActionID{SpellID: 34092},
		SpellSchool: core.SpellSchoolPhysical,
		DefenseType: core.DefenseTypeMelee,
		ProcMask:    core.ProcMaskEmpty, // no proc mask, so it cannot proc itself or the seal
		Flags:       core.SpellFlagMeleeMetrics | core.SpellFlagNoOnCastComplete | core.SpellFlagSuppressEquipProcs,

		DamageMultiplier: paladin.getWeaponSpecializationModifier(),
		ThreatMultiplier: 1,
		BonusCoefficient: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := 0.3 * paladin.MHWeaponDamage(sim, spell.MeleeAttackPower())
			spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeSpecialHitAndCrit)
		},
	})

	aura := paladin.RegisterAura(core.Aura{
		Label:    "Seal of Fury",
		ActionID: actionID,
		Duration: time.Minute * 2,

		OnSpellHitDealt: func(_ *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if !result.Landed() || !spell.ProcMask.Matches(core.ProcMaskMeleeWhiteHit|core.ProcMaskMeleeSpecial) {
				return
			}
			// Up to 3 enemies: the target that was hit and the next two.
			target := result.Target
			for i := 0; i < 3 && i < int(sim.GetNumTargets()); i++ {
				procSpell.Cast(sim, target)
				target = sim.Environment.NextTargetUnit(target)
			}
		},
	})
	paladin.aurasSoF = append(paladin.aurasSoF, aura)

	paladin.sealOfFury = paladin.RegisterSpell(core.SpellConfig{
		ActionID:    actionID,
		SpellSchool: core.SpellSchoolHoly,
		Flags:       core.SpellFlagAPL,

		ManaCost: core.ManaCostOptions{
			BaseCost:   0.08,
			Multiplier: paladin.benediction(),
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			paladin.applySeal(aura, judgeSpell, sim)
		},
	})

	paladin.spellsJoF = append(paladin.spellsJoF, judgeSpell)
}
