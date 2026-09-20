package warlock

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

// Mayhem (spell 34020): when activated, +100% crit chance on the next Destruction spell. 2 min cooldown.
func (warlock *Warlock) registerMayhemCD() {
	if !warlock.Talents.Mayhem {
		return
	}

	actionID := core.ActionID{SpellID: 34020}
	critBonus := 100.0 * core.SpellCritRatingPerCritChance

	isDestruction := func(spell *core.Spell) bool {
		switch spell.SpellCode {
		case SpellCode_WarlockShadowBolt, SpellCode_WarlockImmolate, SpellCode_WarlockConflagrate,
			SpellCode_WarlockSearingPain, SpellCode_WarlockSoulFire, SpellCode_WarlockShadowburn:
			return true
		}
		return false
	}

	aura := warlock.RegisterAura(core.Aura{
		Label:    "Mayhem",
		ActionID: actionID,
		Duration: core.NeverExpires,
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			for _, spell := range warlock.Spellbook {
				if isDestruction(spell) {
					spell.BonusCritRating += critBonus
				}
			}
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			for _, spell := range warlock.Spellbook {
				if isDestruction(spell) {
					spell.BonusCritRating -= critBonus
				}
			}
		},
		OnCastComplete: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
			if isDestruction(spell) {
				aura.Deactivate(sim)
			}
		},
	})

	spell := warlock.RegisterSpell(core.SpellConfig{
		ActionID: actionID,
		Flags:    core.SpellFlagNoOnCastComplete | core.SpellFlagAPL,
		Cast: core.CastConfig{
			CD: core.Cooldown{
				Timer:    warlock.NewTimer(),
				Duration: time.Minute * 2,
			},
		},
		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			aura.Activate(sim)
		},
	})

	warlock.AddMajorCooldown(core.MajorCooldown{
		Spell: spell,
		Type:  core.CooldownTypeDPS,
	})
}
