package hunter

import (
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/stats"
)

const (
	KnightLieutenantsChainGauntlets = 16403
	BloodGuardsChainGauntlets       = 16530
	MarshalsChainGrips              = 16463
	GeneralsChainGloves             = 16571
	RenatakisCharmofBeasts          = 19953
	DevilsaurEye                    = 19991
	DevilsaurTooth                  = 19992
	KnightLieutenantsChainVices     = 23279
	BloodGuardsChainVices           = 22862
	MarkOfTheVeteranHunterA         = 26162
	MarkOfTheVeteranHunterB         = 26171
)

func init() {
	// Mark of the Veteran (hunter): Equip: Reduces the cooldown of your Multi-Shot by 2 sec.
	markOfTheVeteranHunter := func(agent core.Agent) {
		hunter := agent.(HunterAgent).GetHunter()
		core.MakePermanent(hunter.RegisterAura(core.Aura{
			Label: "Mark of the Veteran (Multi-Shot)",
			OnInit: func(aura *core.Aura, sim *core.Simulation) {
				if hunter.MultiShot != nil {
					hunter.MultiShot.CD.Duration -= time.Second * 2
				}
			},
		}))
	}
	core.NewItemEffect(MarkOfTheVeteranHunterA, markOfTheVeteranHunter)
	core.NewItemEffect(MarkOfTheVeteranHunterB, markOfTheVeteranHunter)

	// Equip: Reduces the mana cost of your Arcane Shot by 15.
	core.NewItemEffect(KnightLieutenantsChainGauntlets, func(agent core.Agent) {
		hunter := agent.(HunterAgent).GetHunter()
		core.MakePermanent(hunter.RegisterAura(core.Aura{
			Label: "Arcane Shot Mana Reduction",
			OnInit: func(aura *core.Aura, sim *core.Simulation) {
				if hunter.ArcaneShot != nil {
					hunter.ArcaneShot.Cost.FlatModifier -= 15.0
				}
			},
		}))
	})
	// Equip: Reduces the mana cost of your Arcane Shot by 15.
	core.NewItemEffect(BloodGuardsChainGauntlets, func(agent core.Agent) {
		hunter := agent.(HunterAgent).GetHunter()
		core.MakePermanent(hunter.RegisterAura(core.Aura{
			Label: "Arcane Shot Mana Reduction",
			OnInit: func(aura *core.Aura, sim *core.Simulation) {
				if hunter.ArcaneShot != nil {
					hunter.ArcaneShot.Cost.FlatModifier -= 15.0
				}
			},
		}))
	})

	// Equip: Increases the damage done by your Multi-Shot by 5% (server)
	core.NewItemEffect(MarshalsChainGrips, func(agent core.Agent) {
		hunter := agent.(HunterAgent).GetHunter()
		core.MakePermanent(hunter.RegisterAura(core.Aura{
			Label: "Multi-Shot Damage Increase",
			OnInit: func(aura *core.Aura, sim *core.Simulation) {
				hunter.MultiShot.BaseDamageMultiplierAdditive += 0.05
			},
		}))
	})
	// Equip: Increases the damage done by your Multi-Shot by 5% (server)
	core.NewItemEffect(GeneralsChainGloves, func(agent core.Agent) {
		hunter := agent.(HunterAgent).GetHunter()
		core.MakePermanent(hunter.RegisterAura(core.Aura{
			Label: "Multi-Shot Damage Increase",
			OnInit: func(aura *core.Aura, sim *core.Simulation) {
				hunter.MultiShot.BaseDamageMultiplierAdditive += 0.05
			},
		}))
	})
	// Equip: Increases the damage done by your Multi-Shot by 5% (server)
	core.NewItemEffect(KnightLieutenantsChainVices, func(agent core.Agent) {
		hunter := agent.(HunterAgent).GetHunter()
		core.MakePermanent(hunter.RegisterAura(core.Aura{
			Label: "Multi-Shot Damage Increase",
			OnInit: func(aura *core.Aura, sim *core.Simulation) {
				hunter.MultiShot.BaseDamageMultiplierAdditive += 0.05
			},
		}))
	})
	// Equip: Increases the damage done by your Multi-Shot by 5% (server)
	core.NewItemEffect(BloodGuardsChainVices, func(agent core.Agent) {
		hunter := agent.(HunterAgent).GetHunter()
		core.MakePermanent(hunter.RegisterAura(core.Aura{
			Label: "Multi-Shot Damage Increase",
			OnInit: func(aura *core.Aura, sim *core.Simulation) {
				hunter.MultiShot.BaseDamageMultiplierAdditive += 0.05
			},
		}))
	})
	// Use: Instantly clears the cooldowns of Aimed Shot, Multishot, Volley, and Arcane Shot. (cooldown 3 min)
	core.NewItemEffect(RenatakisCharmofBeasts, func(agent core.Agent) {
		hunter := agent.(HunterAgent).GetHunter()

		spell := hunter.RegisterSpell(core.SpellConfig{
			ActionID: core.ActionID{ItemID: RenatakisCharmofBeasts},
			ProcMask: core.ProcMaskEmpty,
			Flags:    core.SpellFlagNoOnCastComplete | core.SpellFlagOffensiveEquipment,

			Cast: core.CastConfig{
				CD: core.Cooldown{
					Timer:    hunter.NewTimer(),
					Duration: time.Second * 180,
				},
				SharedCD: core.Cooldown{
					Timer:    hunter.GetOffensiveTrinketCD(),
					Duration: time.Second * 10,
				},
			},

			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				for _, s := range []*core.Spell{hunter.AimedShot, hunter.MultiShot, hunter.Volley, hunter.ArcaneShot} {
					if s != nil {
						s.CD.Reset()
					}
				}
			},
		})

		hunter.AddMajorCooldown(core.MajorCooldown{
			Type:  core.CooldownTypeDPS,
			Spell: spell,
		})
	})

	core.NewItemEffect(DevilsaurEye, func(agent core.Agent) {
		hunter := agent.(HunterAgent).GetHunter()

		procBonus := stats.Stats{
			stats.AttackPower:       150,
			stats.RangedAttackPower: 150,
			stats.MeleeHit:          3, // server: +3% hit for 30 sec
		}
		aura := hunter.GetOrRegisterAura(core.Aura{
			Label:    "Devilsaur Fury",
			ActionID: core.ActionID{SpellID: 24352},
			Duration: time.Second * 30,

			OnGain: func(aura *core.Aura, sim *core.Simulation) {
				aura.Unit.AddStatsDynamic(sim, procBonus)
			},
			OnExpire: func(aura *core.Aura, sim *core.Simulation) {
				aura.Unit.AddStatsDynamic(sim, procBonus.Invert())
			},
		})

		spell := hunter.GetOrRegisterSpell(core.SpellConfig{
			ActionID: core.ActionID{SpellID: 24352},
			Flags:    core.SpellFlagNoOnCastComplete | core.SpellFlagOffensiveEquipment,

			Cast: core.CastConfig{
				CD: core.Cooldown{
					Timer:    hunter.NewTimer(),
					Duration: time.Minute * 2,
				},
			},
			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				aura.Activate(sim)
			},
		})

		hunter.AddMajorCooldown(core.MajorCooldown{
			Spell: spell,
			Type:  core.CooldownTypeDPS,
		})
	})

	core.NewItemEffect(DevilsaurTooth, func(agent core.Agent) {
		hunter := agent.(HunterAgent).GetHunter()
		if hunter.pet == nil {
			return
		}

		// Hunter aura so its visible in the timeline
		// TODO: Probably should add pet auras in the timeline at some point
		trackingAura := hunter.GetOrRegisterAura(core.Aura{
			Label:    "Primal Instinct Hunter",
			ActionID: core.ActionID{SpellID: 24353},
			Duration: core.NeverExpires,
		})

		aura := hunter.pet.GetOrRegisterAura(core.Aura{
			Label:    "Primal Instinct",
			ActionID: core.ActionID{SpellID: 24353},
			Duration: core.NeverExpires,

			OnGain: func(aura *core.Aura, sim *core.Simulation) {
				if hunter.pet.focusDump != nil {
					hunter.pet.focusDump.BonusCritRating += 100
				}
				if hunter.pet.specialAbility != nil {
					hunter.pet.specialAbility.BonusCritRating += 100
				}
				trackingAura.Activate(sim)
			},
			OnExpire: func(aura *core.Aura, sim *core.Simulation) {
				if hunter.pet.focusDump != nil {
					hunter.pet.focusDump.BonusCritRating -= 100
				}
				if hunter.pet.specialAbility != nil {
					hunter.pet.specialAbility.BonusCritRating -= 100
				}
				trackingAura.Deactivate(sim)
			},
			OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				if spell == hunter.pet.focusDump || spell == hunter.pet.specialAbility {
					aura.Deactivate(sim)
				}
			},
		})

		spell := hunter.GetOrRegisterSpell(core.SpellConfig{
			ActionID: core.ActionID{SpellID: 24353},
			Flags:    core.SpellFlagNoOnCastComplete | core.SpellFlagOffensiveEquipment,

			Cast: core.CastConfig{
				DefaultCast: core.Cast{
					GCD: core.GCDDefault,
				},
				CD: core.Cooldown{
					Timer:    hunter.NewTimer(),
					Duration: time.Minute * 2,
				},
			},
			ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
				return hunter.pet.IsEnabled()
			},
			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				aura.Activate(sim)
			},
		})

		hunter.AddMajorCooldown(core.MajorCooldown{
			Spell: spell,
			Type:  core.CooldownTypeDPS,
			ShouldActivate: func(sim *core.Simulation, character *core.Character) bool {
				return hunter.pet != nil && hunter.pet.IsEnabled()
			},
		})
	})
}
