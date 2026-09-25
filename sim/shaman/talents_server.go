package shaman

import (
	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/stats"
)

// Talents from the server's custom trees (text from ui/core/talents/trees/shaman.json).
func (shaman *Shaman) applyServerTalents() {
	// Rockhide: all damage taken -2% per rank (the chance to hit and stun nearby enemies is not modeled).
	shaman.PseudoStats.DamageTakenMultiplier *= 1 - 0.02*float64(shaman.Talents.Rockhide)

	// Primal Endurance: +2% total Health per rank (the fatal damage reduction is not modeled).
	if shaman.Talents.PrimalEndurance > 0 {
		shaman.MultiplyStat(stats.Health, 1+0.02*float64(shaman.Talents.PrimalEndurance))
	}

	// Nature's Grace: threat -10% per rank.
	shaman.PseudoStats.ThreatMultiplier *= 1 - 0.1*float64(shaman.Talents.NatureGrace)

	// Nature's Guardian: -2% chance to be critically hit per rank.
	shaman.PseudoStats.ReducedCritTakenChance += 0.02 * float64(shaman.Talents.NaturesGuardian)

	// Meditation: 10% of mana regeneration per rank continues while casting.
	shaman.PseudoStats.SpiritRegenRateCasting += 0.1 * float64(shaman.Talents.Meditation)

	shaman.applyArmamentsOfStorm()
	shaman.applyBloodlust()
}

// Armaments of Storm (Spell.csv 33636-33640): melee auto attacks have a 5% chance per rank to trigger 33641,
// Nature damage equal to 5 x level (300 at 60, "up to 300 ... scales with your level").
func (shaman *Shaman) applyArmamentsOfStorm() {
	if shaman.Talents.ArmamentsOfStorm == 0 {
		return
	}

	damage := 5 * float64(shaman.Level)
	procSpell := shaman.RegisterSpell(core.SpellConfig{
		ActionID:    core.ActionID{SpellID: 33641},
		SpellSchool: core.SpellSchoolNature,
		DefenseType: core.DefenseTypeMagic,
		ProcMask:    core.ProcMaskSpellDamageProc,
		Flags:       core.SpellFlagNoOnCastComplete | core.SpellFlagPassiveSpell,

		DamageMultiplier: 1,
		ThreatMultiplier: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.CalcAndDealDamage(sim, target, damage, spell.OutcomeMagicHitAndCrit)
		},
	})

	core.MakeProcTriggerAura(&shaman.Unit, core.ProcTrigger{
		Name:       "Armaments of Storm",
		ActionID:   core.ActionID{SpellID: []int32{0, 33636, 33637, 33638, 33639, 33640}[shaman.Talents.ArmamentsOfStorm]},
		Callback:   core.CallbackOnSpellHitDealt,
		ProcMask:   core.ProcMaskMeleeWhiteHit,
		Outcome:    core.OutcomeLanded,
		ProcChance: 0.05 * float64(shaman.Talents.ArmamentsOfStorm),
		Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			procSpell.Cast(sim, result.Target)
		},
	})
}

// Bloodlust (Spell.csv 33630): attack and casting speed +10% for 3 min, 1 min cooldown, so it can be kept up all
// the time. The sim assumes the shaman keeps it on itself (TODO question 34: can it target the caster?).
func (shaman *Shaman) applyBloodlust() {
	if !shaman.Talents.Bloodlust {
		return
	}

	core.MakePermanent(shaman.RegisterAura(core.Aura{
		Label:    "Bloodlust",
		ActionID: core.ActionID{SpellID: 33630},
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			shaman.MultiplyAttackSpeed(sim, 1.1)
			shaman.MultiplyCastSpeed(1.1)
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			shaman.MultiplyAttackSpeed(sim, 1/1.1)
			shaman.MultiplyCastSpeed(1 / 1.1)
		},
	}))
}
