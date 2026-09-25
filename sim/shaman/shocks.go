package shaman

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

// Shared logic for all shocks.
func (shaman *Shaman) newShockSpellConfig(actionId core.ActionID, spellSchool core.SpellSchool, baseCost float64, shockTimer *core.Timer) core.SpellConfig {
	cdDuration := time.Second*10 - time.Second*time.Duration(shaman.Talents.Reverberation) // DBC: 10s base (every shock rank), Reverberation -1s/rank

	return core.SpellConfig{
		ActionID:    actionId,
		SpellSchool: spellSchool,
		DefenseType: core.DefenseTypeMagic,
		ProcMask:    core.ProcMaskSpellDamage,
		Flags:       SpellFlagShaman | core.SpellFlagAPL,

		ManaCost: core.ManaCostOptions{
			FlatCost:   baseCost,
			Multiplier: 100 - 5*shaman.Talents.Convection, // DBC: 5%/rank
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			CD: core.Cooldown{
				Timer:    shaman.NewTimer(),
				Duration: cdDuration,
			},
			SharedCD: core.Cooldown{
				Timer:    shockTimer,
				Duration: cdDuration,
			},
		},

		DamageMultiplier: 1,
		ThreatMultiplier: 1,
	}
}

func (shaman *Shaman) registerShocks() {
	shockTimer := shaman.NewTimer()
	shaman.registerEarthShockSpell(shockTimer)
	shaman.registerFlameShockSpell(shockTimer)
	shaman.registerFrostShockSpell(shockTimer)
	shaman.registerAftershockSpell()
}
