package mage

import (
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/stats"
)

// Talents from the server's custom trees (text from ui/core/talents/trees/mage.json).
func (mage *Mage) applyServerTalents() {
	// Fire Warding: +10 Fire resistance per rank (the Fire Ward reflect is not modeled).
	mage.AddStat(stats.FireResistance, 10*float64(mage.Talents.FireWarding))

	// Thermal Expansion: "Generates mana equal to your level every 2 sec."
	if mage.Talents.ThermalExpansion {
		metrics := mage.NewManaMetrics(core.ActionID{SpellID: 34125})
		core.MakePermanent(mage.RegisterAura(core.Aura{
			Label:    "Thermal Expansion",
			ActionID: core.ActionID{SpellID: 34125},
			OnGain: func(aura *core.Aura, sim *core.Simulation) {
				core.StartPeriodicAction(sim, core.PeriodicActionOptions{
					Period: time.Second * 2,
					OnAction: func(sim *core.Simulation) {
						mage.AddMana(sim, float64(mage.Level), metrics)
					},
				})
			},
		}))
	}

	// Brilliance Aura: "Generates 1% of total Mana every 5 sec to all party members". Only the mage's own mana is
	// modeled.
	if mage.Talents.BrillianceAura {
		metrics := mage.NewManaMetrics(core.ActionID{SpellID: 34091})
		core.MakePermanent(mage.RegisterAura(core.Aura{
			Label:    "Brilliance Aura",
			ActionID: core.ActionID{SpellID: 34091},
			OnGain: func(aura *core.Aura, sim *core.Simulation) {
				core.StartPeriodicAction(sim, core.PeriodicActionOptions{
					Period: time.Second * 5,
					OnAction: func(sim *core.Simulation) {
						mage.AddMana(sim, 0.01*mage.MaxMana(), metrics)
					},
				})
			},
		}))
	}
}
