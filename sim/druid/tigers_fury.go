package druid

import (
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/stats"
)

func (druid *Druid) registerTigersFurySpell() {
	actionID := core.ActionID{SpellID: map[int32]int32{
		25: 5217,
		40: 6793,
		50: 9845,
		60: 9846,
	}[druid.Level]}

	// Server: +140/280/420/560 attack power for 10 sec (Spell.csv).
	apBonus := map[int32]float64{
		25: 140.0,
		40: 280.0,
		50: 420.0,
		60: 560.0,
	}[druid.Level]

	druid.TigersFuryAura = druid.RegisterAura(core.Aura{
		Label:    "Tiger's Fury Aura",
		ActionID: actionID,
		Duration: time.Duration(float64(10*time.Second) * (1 + 0.2*float64(druid.Talents.FeralInstinct))), // Feral Instinct: +20%/rank (DBC 16947-16949)
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			druid.AddStatDynamic(sim, stats.AttackPower, apBonus)
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			druid.AddStatDynamic(sim, stats.AttackPower, -apBonus)
		},
	})

	spell := druid.RegisterSpell(Cat, core.SpellConfig{
		ActionID: actionID,
		Flags:    core.SpellFlagAPL,

		EnergyCost: core.EnergyCostOptions{
			Cost: 20, // server (Spell.csv)
		},
		Cast: core.CastConfig{
			CD: core.Cooldown{
				Timer:    druid.NewTimer(),
				Duration: time.Second * 10, // server (Spell.csv)
			},
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			druid.TigersFuryAura.Activate(sim)
		},
	})

	druid.TigersFury = spell
}
