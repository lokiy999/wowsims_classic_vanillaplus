package warrior

import (
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/stats"
)

// Recklessness (server Spell.csv 1719): +100% critical strike chance and +30% damage taken for 15 sec, 30 min cooldown.
func (warrior *Warrior) RegisterRecklessnessCD() {
	if warrior.Level < 50 {
		return
	}

	actionID := core.ActionID{SpellID: 1719}

	reckAura := warrior.RegisterAura(core.Aura{
		Label:    "Recklessness",
		ActionID: actionID,
		Duration: time.Second * 15,
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			warrior.PseudoStats.DamageTakenMultiplier *= 1.3
			warrior.AddStatDynamic(sim, stats.MeleeCrit, 100*core.CritRatingPerCritChance)
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			warrior.PseudoStats.DamageTakenMultiplier /= 1.3
			warrior.AddStatDynamic(sim, stats.MeleeCrit, -100*core.CritRatingPerCritChance)

		},
	})

	Recklessness := warrior.RegisterSpell(BerserkerStance, core.SpellConfig{
		ActionID: actionID,
		Cast: core.CastConfig{
			IgnoreHaste: true,
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			CD: core.Cooldown{
				Timer:    warrior.NewTimer(),
				Duration: time.Duration(float64(time.Minute*30) * warrior.innerRageFactor()),
			},
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			reckAura.Activate(sim)
		},
	})

	warrior.AddMajorCooldown(core.MajorCooldown{
		Spell: Recklessness.Spell,
		Type:  core.CooldownTypeDPS,
	})
}
