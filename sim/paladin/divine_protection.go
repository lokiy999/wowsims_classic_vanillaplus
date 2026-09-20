package paladin

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

// Divine Protection (DBC 498): reduces all damage taken by 50% but damage dealt by 30% for 10 sec, 10 min cooldown.
// Requires the Divine Protection talent.
func (paladin *Paladin) registerDivineProtection() {
	if !paladin.Talents.DivineProtection {
		return
	}

	actionID := core.ActionID{SpellID: 498}

	aura := paladin.RegisterAura(core.Aura{
		Label:    "Divine Protection",
		ActionID: actionID,
		Duration: time.Second * 10,
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			aura.Unit.PseudoStats.DamageTakenMultiplier *= 0.5
			aura.Unit.PseudoStats.DamageDealtMultiplier *= 0.7
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			aura.Unit.PseudoStats.DamageTakenMultiplier /= 0.5
			aura.Unit.PseudoStats.DamageDealtMultiplier /= 0.7
		},
	})

	spell := paladin.RegisterSpell(core.SpellConfig{
		ActionID: actionID,
		Flags:    core.SpellFlagNoOnCastComplete | core.SpellFlagAPL,

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			CD: core.Cooldown{
				Timer:    paladin.NewTimer(),
				Duration: time.Minute * 10,
			},
		},
		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			aura.Activate(sim)
		},
	})

	paladin.AddMajorCooldown(core.MajorCooldown{
		Spell: spell,
		Type:  core.CooldownTypeSurvival,
	})
}
