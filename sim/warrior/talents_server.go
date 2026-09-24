package warrior

import (
	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/stats"
)

// Talents from the server's custom trees (text from ui/core/talents/trees/warrior.json).
func (warrior *Warrior) applyServerTalents() {
	warrior.applyBladeMail()
}

// Blade Mail: "When struck in combat, deals damage to the attacker equal to 0.5% of your Armor" per rank.
func (warrior *Warrior) applyBladeMail() {
	if warrior.Talents.BladeMail == 0 {
		return
	}
	pct := 0.005 * float64(warrior.Talents.BladeMail)

	bladeMail := warrior.Unit.RegisterSpell(core.SpellConfig{
		ActionID:    core.ActionID{SpellID: []int32{0, 33419, 33420, 33421}[warrior.Talents.BladeMail]},
		SpellSchool: core.SpellSchoolPhysical,
		ProcMask:    core.ProcMaskEmpty,
		Flags:       core.SpellFlagNoOnCastComplete | core.SpellFlagPassiveSpell,

		DamageMultiplier: 1,
		ThreatMultiplier: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.CalcAndDealDamage(sim, target, pct*warrior.GetStat(stats.Armor), spell.OutcomeAlwaysHit)
		},
	})

	core.MakePermanent(warrior.RegisterAura(core.Aura{
		Label: "Blade Mail",
		OnSpellHitTaken: func(_ *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if result.Landed() && spell.ProcMask.Matches(core.ProcMaskMelee) {
				bladeMail.Cast(sim, spell.Unit)
			}
		},
	}))
}
