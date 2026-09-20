package hunter

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

func (hunter *Hunter) registerRapidFire() {
	if hunter.Level < 26 {
		return
	}

	actionID := core.ActionID{SpellID: 3045}
	// +30% ranged speed for 20 sec, 3 min cooldown (confirmed in game); Stalking takes 1 min off per rank (DBC).
	cooldown := time.Minute*3 - time.Minute*time.Duration(hunter.Talents.Stalking)

	hunter.RapidFireAura = hunter.RegisterAura(core.Aura{
		Label:    "Rapid Fire",
		ActionID: actionID,
		Duration: time.Second * 20,

		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			aura.Unit.MultiplyRangedSpeed(sim, 1.3)
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			aura.Unit.MultiplyRangedSpeed(sim, 1/1.3)
		},
	})

	hunter.RapidFire = hunter.RegisterSpell(core.SpellConfig{
		ActionID: actionID,

		// No mana cost (confirmed in game).
		Cast: core.CastConfig{
			CD: core.Cooldown{
				Timer:    hunter.NewTimer(),
				Duration: cooldown,
			},
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			hunter.RapidFireAura.Activate(sim)
		},
	})

	hunter.AddMajorCooldown(core.MajorCooldown{
		Spell: hunter.RapidFire,
		Type:  core.CooldownTypeDPS,
	})
}
