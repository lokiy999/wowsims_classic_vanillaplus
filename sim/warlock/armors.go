package warlock

import (
	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/stats"
)

func (warlock *Warlock) applyDemonArmor() {
	spellID := map[int32]int32{
		25: 706,
		40: 11733,
		50: 11734,
		60: 11735,
	}[warlock.Level]

	armor := map[int32]float64{ // server (Spell.csv 706/11733/11734/11735)
		25: 210.0,
		40: 400.0,
		50: 500.0,
		60: 600.0,
	}[warlock.Level]

	shadowRes := map[int32]float64{
		25: 5.0,
		40: 15.0,
		50: 20.0,
		60: 30.0,
	}[warlock.Level]

	// DBC: Demonic Embrace +10%/rank effectiveness.
	armor *= 1 + 0.10*float64(warlock.Talents.DemonicEmbrace)
	warlock.AddStat(stats.Armor, armor)
	warlock.AddStat(stats.ShadowResistance, shadowRes)

	warlock.GetOrRegisterAura(core.Aura{
		Label:    "Demon Armor",
		ActionID: core.ActionID{SpellID: spellID},
		Duration: core.NeverExpires,
		OnReset: func(aura *core.Aura, sim *core.Simulation) {
			aura.Activate(sim)
		},
	})
}
