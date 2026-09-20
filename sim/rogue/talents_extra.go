package rogue

import (
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/stats"
)

// Talents added in the second audit round (see docs/CHANGES.md Part BA).
func (rogue *Rogue) applyAuditTalents() {
	rogue.applyGainingAnAdvantage()
	rogue.applyBrigandage()
}

// Gaining an Advantage (DBC 33703-33709): auto attacks have a 20%/rank chance to give +1% Agility, stacking up to 20 times,
// for 15 sec (the target stat reduction only matters against players, so it is not modeled).
func (rogue *Rogue) applyGainingAnAdvantage() {
	if rogue.Talents.GainingAnAdvantage == 0 {
		return
	}

	procChance := 0.2 * float64(rogue.Talents.GainingAnAdvantage)

	const maxStacks = 20
	var agilityDeps [maxStacks + 1]*stats.StatDependency
	for i := range agilityDeps {
		agilityDeps[i] = rogue.NewDynamicMultiplyStat(stats.Agility, 1+0.01*float64(i))
	}

	buff := rogue.RegisterAura(core.Aura{
		Label:     "Gaining an Advantage",
		ActionID:  core.ActionID{SpellID: 33709},
		Duration:  time.Second * 15,
		MaxStacks: maxStacks,
		OnStacksChange: func(aura *core.Aura, sim *core.Simulation, oldStacks int32, newStacks int32) {
			rogue.DisableDynamicStatDep(sim, agilityDeps[oldStacks])
			rogue.EnableDynamicStatDep(sim, agilityDeps[newStacks])
		},
	})

	core.MakePermanent(rogue.RegisterAura(core.Aura{
		Label: "Gaining an Advantage Trigger",
		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if !result.Landed() || !spell.ProcMask.Matches(core.ProcMaskMeleeWhiteHit) {
				return
			}
			if sim.Proc(procChance, "Gaining an Advantage") {
				buff.Activate(sim)
				buff.AddStack(sim)
			}
		},
	}))
}

// Brigandage (DBC 33726-33730): auto attacks have a 20%/rank chance to deal 0.5 per level Shadow damage (30 at level 60)
// and drain mana, energy or rage from the target (the drain does nothing to the rogue, so it is not modeled).
func (rogue *Rogue) applyBrigandage() {
	if rogue.Talents.Brigandage == 0 {
		return
	}

	procChance := 0.2 * float64(rogue.Talents.Brigandage)
	damage := 0.5 * float64(rogue.Level)

	proc := rogue.RegisterSpell(core.SpellConfig{
		ActionID:    core.ActionID{SpellID: 33726},
		SpellSchool: core.SpellSchoolShadow,
		DefenseType: core.DefenseTypeMagic,
		ProcMask:    core.ProcMaskEmpty,
		Flags:       core.SpellFlagNoOnCastComplete,

		DamageMultiplier: 1,
		ThreatMultiplier: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.CalcAndDealDamage(sim, target, damage, spell.OutcomeMagicHit)
		},
	})

	core.MakePermanent(rogue.RegisterAura(core.Aura{
		Label: "Brigandage",
		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if !result.Landed() || !spell.ProcMask.Matches(core.ProcMaskMeleeWhiteHit) {
				return
			}
			if sim.Proc(procChance, "Brigandage") {
				proc.Cast(sim, result.Target)
			}
		},
	}))
}
