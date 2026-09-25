package item_sets

import (
	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/stats"
)

///////////////////////////////////////////////////////////////////////////
//                                 Cloth
///////////////////////////////////////////////////////////////////////////

var ItemSetTheHighlandersIntent = core.NewItemSet(core.ItemSet{
	Name: "The Highlander's Intent",
	Bonuses: map[int32]core.ApplyEffect{ // server bonus texts (VPlusItemDB.lua)
		// Improves your chance to get a critical strike with spells by 1%.
		2: func(agent core.Agent) {
			character := agent.GetCharacter()
			character.AddStat(stats.SpellCrit, 1*core.SpellCritRatingPerCritChance)
		},
		// +18 Stamina.
		3: func(agent core.Agent) {
			character := agent.GetCharacter()
			character.AddStat(stats.Stamina, 18)
		},
	},
})

var ItemSetTheDefilersIntent = core.NewItemSet(core.ItemSet{
	Name: "The Defiler's Intent",
	Bonuses: map[int32]core.ApplyEffect{ // server bonus texts (VPlusItemDB.lua)
		// Improves your chance to get a critical strike with spells by 1%.
		2: func(agent core.Agent) {
			character := agent.GetCharacter()
			character.AddStat(stats.SpellCrit, 1*core.SpellCritRatingPerCritChance)
		},
		// +18 Stamina.
		3: func(agent core.Agent) {
			character := agent.GetCharacter()
			character.AddStat(stats.Stamina, 18)
		},
	},
})

///////////////////////////////////////////////////////////////////////////
//                                 Leather
///////////////////////////////////////////////////////////////////////////

var ItemSetTheHighlandersPurpose = core.NewItemSet(core.ItemSet{
	Name: "The Highlander's Purpose",
	Bonuses: map[int32]core.ApplyEffect{ // server bonus texts (VPlusItemDB.lua)
		// Improves your chance to get a critical strike by 1%.
		2: func(agent core.Agent) {
			character := agent.GetCharacter()
			character.AddStat(stats.MeleeCrit, 1*core.CritRatingPerCritChance)
		},
		// +18 Stamina.
		3: func(agent core.Agent) {
			character := agent.GetCharacter()
			character.AddStat(stats.Stamina, 18)
		},
	},
})

var ItemSetTheHighlandersWill = core.NewItemSet(core.ItemSet{
	Name: "The Highlander's Will",
	Bonuses: map[int32]core.ApplyEffect{ // server bonus texts (VPlusItemDB.lua)
		// Improves your chance to get a critical strike with spells by 1%.
		2: func(agent core.Agent) {
			character := agent.GetCharacter()
			character.AddStat(stats.SpellCrit, 1*core.SpellCritRatingPerCritChance)
		},
		// +18 Stamina.
		3: func(agent core.Agent) {
			character := agent.GetCharacter()
			character.AddStat(stats.Stamina, 18)
		},
	},
})

var ItemSetTheDefilersPurpose = core.NewItemSet(core.ItemSet{
	Name: "The Defiler's Purpose",
	Bonuses: map[int32]core.ApplyEffect{ // server bonus texts (VPlusItemDB.lua)
		// Improves your chance to get a critical strike by 1%.
		2: func(agent core.Agent) {
			character := agent.GetCharacter()
			character.AddStat(stats.MeleeCrit, 1*core.CritRatingPerCritChance)
		},
		// +18 Stamina.
		3: func(agent core.Agent) {
			character := agent.GetCharacter()
			character.AddStat(stats.Stamina, 18)
		},
	},
})

var ItemSetTheDefilersWill = core.NewItemSet(core.ItemSet{
	Name: "The Defiler's Will",
	Bonuses: map[int32]core.ApplyEffect{ // server bonus texts (VPlusItemDB.lua)
		// Improves your chance to get a critical strike with spells by 1%.
		2: func(agent core.Agent) {
			character := agent.GetCharacter()
			character.AddStat(stats.SpellCrit, 1*core.SpellCritRatingPerCritChance)
		},
		// +18 Stamina.
		3: func(agent core.Agent) {
			character := agent.GetCharacter()
			character.AddStat(stats.Stamina, 18)
		},
	},
})

///////////////////////////////////////////////////////////////////////////
//                                 Mail
///////////////////////////////////////////////////////////////////////////

var ItemSetTheHighlandersDetermination = core.NewItemSet(core.ItemSet{
	Name: "The Highlander's Determination",
	Bonuses: map[int32]core.ApplyEffect{ // server bonus texts (VPlusItemDB.lua)
		// Improves your chance to get a critical strike by 1%.
		2: func(agent core.Agent) {
			character := agent.GetCharacter()
			character.AddStat(stats.MeleeCrit, 1*core.CritRatingPerCritChance)
		},
		// +18 Stamina.
		3: func(agent core.Agent) {
			character := agent.GetCharacter()
			character.AddStat(stats.Stamina, 18)
		},
	},
})

var ItemSetTheDefilersFortitude = core.NewItemSet(core.ItemSet{
	Name: "The Defiler's Fortitude",
	Bonuses: map[int32]core.ApplyEffect{ // server bonus texts (VPlusItemDB.lua)
		// Improves your chance to get a critical strike by 1%.
		2: func(agent core.Agent) {
			character := agent.GetCharacter()
			character.AddStat(stats.MeleeCrit, 1*core.CritRatingPerCritChance)
		},
		// +18 Stamina.
		3: func(agent core.Agent) {
			character := agent.GetCharacter()
			character.AddStat(stats.Stamina, 18)
		},
	},
})

var ItemSetTheDefilersDetermination = core.NewItemSet(core.ItemSet{
	Name: "The Defiler's Determination",
	Bonuses: map[int32]core.ApplyEffect{ // server bonus texts (VPlusItemDB.lua)
		// Improves your chance to get a critical strike by 1%.
		2: func(agent core.Agent) {
			character := agent.GetCharacter()
			character.AddStat(stats.MeleeCrit, 1*core.CritRatingPerCritChance)
		},
		// +18 Stamina.
		3: func(agent core.Agent) {
			character := agent.GetCharacter()
			character.AddStat(stats.Stamina, 18)
		},
	},
})

///////////////////////////////////////////////////////////////////////////
//                                 Plate
///////////////////////////////////////////////////////////////////////////

var ItemSetTheHighlandersResolve = core.NewItemSet(core.ItemSet{
	Name: "The Highlander's Resolve",
	Bonuses: map[int32]core.ApplyEffect{ // server bonus texts (VPlusItemDB.lua)
		// Improves your chance to get a critical strike by 1%.
		2: func(agent core.Agent) {
			character := agent.GetCharacter()
			character.AddStat(stats.MeleeCrit, 1*core.CritRatingPerCritChance)
		},
		// +18 Stamina.
		3: func(agent core.Agent) {
			character := agent.GetCharacter()
			character.AddStat(stats.Stamina, 18)
		},
	},
})

var ItemSetTheHighlandersResolution = core.NewItemSet(core.ItemSet{
	Name: "The Highlander's Resolution",
	Bonuses: map[int32]core.ApplyEffect{ // server bonus texts (VPlusItemDB.lua)
		// Improves your chance to get a critical strike by 1%.
		2: func(agent core.Agent) {
			character := agent.GetCharacter()
			character.AddStat(stats.MeleeCrit, 1*core.CritRatingPerCritChance)
		},
		// +18 Stamina.
		3: func(agent core.Agent) {
			character := agent.GetCharacter()
			character.AddStat(stats.Stamina, 18)
		},
	},
})

var ItemSetTheDefilersResolution = core.NewItemSet(core.ItemSet{
	Name: "The Defiler's Resolution",
	Bonuses: map[int32]core.ApplyEffect{ // server bonus texts (VPlusItemDB.lua)
		// Improves your chance to get a critical strike by 1%.
		2: func(agent core.Agent) {
			character := agent.GetCharacter()
			character.AddStat(stats.MeleeCrit, 1*core.CritRatingPerCritChance)
		},
		// +18 Stamina.
		3: func(agent core.Agent) {
			character := agent.GetCharacter()
			character.AddStat(stats.Stamina, 18)
		},
	},
})
