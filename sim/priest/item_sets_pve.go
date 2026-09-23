package priest

import (

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/stats"
)

///////////////////////////////////////////////////////////////////////////
//                            Classic Phase 1 Item Sets - Molten Core
///////////////////////////////////////////////////////////////////////////

var ItemSetVestmentsOfProphecy = core.NewItemSet(core.ItemSet{
	Name: "Vestments of Prophecy",
	Bonuses: map[int32]core.ApplyEffect{
		// +20 Stamina.
		2: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddStat(stats.Stamina, 20)
		},
		// Improves your chance to get a critical strike with spells by 2%.
		4: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddStat(stats.SpellCrit, 2*core.SpellCritRatingPerCritChance)
		},
		// Whenever you are struck by a Silence or spell Interruption effect you will
		// generate 10% of your total Health and Mana over 10 sec. This effect can only
		// occur once every 30 sec.
		// Nothing to do: this sim does not model enemy Silence/Interrupt effects landing
		// on the player, so there is nothing for this proc to hook into.
		6: func(agent core.Agent) {
			// Nothing to do
		},
		// Increases the damage and healing done by your spells by 5%.
		8: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.PseudoStats.DamageDealtMultiplier *= 1.05
			c.PseudoStats.HealingDealtMultiplier *= 1.05
		},
	},
})

///////////////////////////////////////////////////////////////////////////
//                            Classic Phase 3 Item Sets - BWL
///////////////////////////////////////////////////////////////////////////

var ItemSetVestmentsOfTranscendence = core.NewItemSet(core.ItemSet{
	Name: "Vestments of Transcendence",
	Bonuses: map[int32]core.ApplyEffect{
		// Allows 15% of your Mana regeneration to continue while casting.
		2: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.PseudoStats.SpiritRegenRateCasting += .15
		},
		// When struck in melee, you will Fade for 4 sec. This effect can only occur
		// once every 30 sec.
		// Nothing to do: Fade is not implemented in this sim (it is a pure
		// threat-drop utility effect with no mechanical impact on damage/healing).
		4: func(agent core.Agent) {
			// Nothing to do
		},
		// Increases your chance of a critical hit with Prayer of Healing by 25%.
		// Nothing to do: Prayer of Healing is not implemented in this sim.
		6: func(agent core.Agent) {
			// Nothing to do
		},
		// Your heals on others will also heal you for 10% of the amount healed.
		8: func(agent core.Agent) {
			c := agent.GetCharacter()
			healthMetrics := c.NewHealthMetrics(core.ActionID{SpellID: 25997})

			c.RegisterAura(core.Aura{
				Label:    "Vestments of Transcendence (8) Trigger",
				Duration: core.NeverExpires,
				OnReset: func(aura *core.Aura, sim *core.Simulation) {
					aura.Activate(sim)
				},
				OnHealDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
					if result.Target != &c.Unit && result.Damage > 0 {
						c.GainHealth(sim, result.Damage*0.10, healthMetrics)
					}
				},
			})
		},
	},
})

///////////////////////////////////////////////////////////////////////////
//                            Classic Phase 4 Item Sets - ZG and AB
///////////////////////////////////////////////////////////////////////////

var ItemSetConfessorsRaiment = core.NewItemSet(core.ItemSet{
	Name: "Confessor's Raiment",
	Bonuses: map[int32]core.ApplyEffect{
		// Increase the range of your Smite and Holy Fire spells by 5 yds.
		// Nothing to do: range has no mechanical effect in this sim.
		2: func(agent core.Agent) {
			// Nothing to do
		},
		// Reduces the casting time of your Mind Control spell by 0.5 sec.
		// Nothing to do: Mind Control is not implemented in this sim.
		3: func(agent core.Agent) {
			// Nothing to do
		},
		// Increases healing done by spells and effects by up to 44.
		5: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddStat(stats.HealingPower, 44)
		},
	},
})

///////////////////////////////////////////////////////////////////////////
//                            Classic Phase 5 Item Sets - AQ
///////////////////////////////////////////////////////////////////////////

///////////////////////////////////////////////////////////////////////////
//                            Classic Phase 6 Item Sets - Naxx
///////////////////////////////////////////////////////////////////////////
