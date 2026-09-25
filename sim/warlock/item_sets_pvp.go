package warlock

import (
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/stats"
)

var bluePvPBonuses = map[int32]core.ApplyEffect{
	// +20 Stamina.
	2: func(agent core.Agent) {
		c := agent.GetCharacter()
		c.AddStat(stats.Stamina, 20)
	},
	// Reduces the casting time of your Immolate spell by 0.5 sec. (server)
	4: func(agent core.Agent) {
		warlock := agent.(WarlockAgent).GetWarlock()
		warlock.GetOrRegisterAura(core.Aura{
			Label: "Immolate Cast Time Reduction",
			OnInit: func(aura *core.Aura, sim *core.Simulation) {
				for _, spell := range warlock.Immolate {
					spell.DefaultCast.CastTime -= time.Millisecond * 500
				}
			},
		})
	},
	// Increases damage and healing done by magical spells and effects by up to 23.
	6: func(agent core.Agent) {
		c := agent.GetCharacter()
		c.AddStat(stats.SpellPower, 23)
	},
}

var ItemSetChampionsThreads = core.NewItemSet(core.ItemSet{
	Name:    "Champion's Dreadgear",
	Bonuses: bluePvPBonuses,
})

var ItemSetLieutenantCommandersThreads = core.NewItemSet(core.ItemSet{
	Name:    "Lieutenant Commander's Dreadgear",
	Bonuses: bluePvPBonuses,
})

var epicPvpBonuses = map[int32]core.ApplyEffect{
	// Increases damage and healing done by magical spells and effects by up to 23.
	2: func(agent core.Agent) {
		c := agent.GetCharacter()
		c.AddStat(stats.SpellPower, 23)
	},
	// Reduces the casting time of your Immolate spell by 0.5 sec. (server)
	3: func(agent core.Agent) {
		warlock := agent.(WarlockAgent).GetWarlock()
		warlock.GetOrRegisterAura(core.Aura{
			Label: "Immolate Cast Time Reduction",
			OnInit: func(aura *core.Aura, sim *core.Simulation) {
				for _, spell := range warlock.Immolate {
					spell.DefaultCast.CastTime -= time.Millisecond * 500
				}
			},
		})
	},
	// +20 Stamina.
	6: func(agent core.Agent) {
		c := agent.GetCharacter()
		c.AddStat(stats.Stamina, 20)
	},
}

var ItemSetWarlordsThreads = core.NewItemSet(core.ItemSet{
	Name:    "Warlord's Threads",
	Bonuses: epicPvpBonuses,
})

var ItemSetFieldMarshalsThreads = core.NewItemSet(core.ItemSet{
	Name:    "Field Marshal's Threads",
	Bonuses: epicPvpBonuses,
})
