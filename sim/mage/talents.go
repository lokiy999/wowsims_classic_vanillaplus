package mage

import (
	"slices"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/stats"
)

func (mage *Mage) ApplyTalents() {
	mage.applyArcaneTalents()
	mage.applyFireTalents()
	mage.applyFrostTalents()
	mage.applyPyromania()
	mage.applyHotStreak()
	mage.applyArcaneResilience()
	mage.applySpellTwisting()
}

func (mage *Mage) applyArcaneTalents() {
	mage.applyArcaneConcentration()
	mage.registerPresenceOfMindCD()
	mage.registerArcanePowerCD()

	// Arcane Subtlety
	if mage.Talents.ArcaneSubtlety > 0 {
		// DBC: -25%/rank Arcane threat and -5/rank target resistance to all your spells.
		mage.AddStat(stats.SpellPenetration, 5*float64(mage.Talents.ArcaneSubtlety))
		threatMultiplier := 1 - .25*float64(mage.Talents.ArcaneSubtlety)
		mage.OnSpellRegistered(func(spell *core.Spell) {
			if spell.SpellSchool.Matches(core.SpellSchoolArcane) && spell.Flags.Matches(SpellFlagMage) {
				spell.ThreatMultiplier *= threatMultiplier
			}
		})
	}

	// Arcane Focus
	if mage.Talents.ArcaneFocus > 0 {
		bonusHit := 1 * float64(mage.Talents.ArcaneFocus) * core.SpellHitRatingPerHitChance // DBC: 1%/rank
		mage.OnSpellRegistered(func(spell *core.Spell) {
			if spell.SpellSchool.Matches(core.SpellSchoolArcane) && spell.Flags.Matches(SpellFlagMage) {
				spell.BonusHitRating += bonusHit
			}
		})
	}

	// Magic Absorption
	if mage.Talents.MagicAbsorption > 0 {
		magicAbsorptionBonus := 5 * float64(mage.Talents.MagicAbsorption) // DBC: +5 all resistances per rank
		mage.AddResistances(magicAbsorptionBonus)
	}

	// Arcane Meditation
	mage.PseudoStats.SpiritRegenRateCasting += 0.10 * float64(mage.Talents.ArcaneMeditation) // DBC: 10%/rank

	if mage.Talents.ArcaneMind > 0 {
		// DBC: +3%/rank total Intellect (not Mana).
		mage.MultiplyStat(stats.Intellect, 1.0+0.03*float64(mage.Talents.ArcaneMind))
	}

	// Arcane Instability
	if mage.Talents.ArcaneInstability > 0 {
		bonusDamageMultiplierAdditive := .01 * float64(mage.Talents.ArcaneInstability)
		bonusCritRating := 1 * float64(mage.Talents.ArcaneInstability) * core.SpellCritRatingPerCritChance

		mage.OnSpellRegistered(func(spell *core.Spell) {
			if spell.Flags.Matches(SpellFlagMage) {
				spell.DamageMultiplierAdditive += bonusDamageMultiplierAdditive
				spell.BonusCritRating += bonusCritRating
			}
		})
	}

	// Time Pressure (DBC): +4%/rank casting speed.
	if mage.Talents.TimePressure > 0 {
		mage.MultiplyCastSpeed(1.0 + 0.04*float64(mage.Talents.TimePressure))
	}

	// Overheat (DBC): +2%/rank spell crit (all schools).
	if mage.Talents.Overheat > 0 {
		bonusCrit := 2 * float64(mage.Talents.Overheat) * core.SpellCritRatingPerCritChance
		mage.OnSpellRegistered(func(spell *core.Spell) {
			if spell.Flags.Matches(SpellFlagMage) {
				spell.BonusCritRating += bonusCrit
			}
		})
	}

	// Arcane Wrath (DBC): +50%/rank Arcane crit strike damage bonus.
	if mage.Talents.ArcaneWrath > 0 {
		critBonus := 0.5 * float64(mage.Talents.ArcaneWrath)
		mage.OnSpellRegistered(func(spell *core.Spell) {
			if spell.SpellSchool.Matches(core.SpellSchoolArcane) && spell.Flags.Matches(SpellFlagMage) {
				spell.CritDamageBonus += critBonus
			}
		})
	}

	// Mind Mastery (DBC): spell damage up to 5%/rank of total Intellect.
	if mage.Talents.MindMastery > 0 {
		mage.AddStatDependency(stats.Intellect, stats.SpellPower, 0.05*float64(mage.Talents.MindMastery))
	}
}

func (mage *Mage) applyFireTalents() {
	mage.applyIgnite()
	mage.applyImprovedScorch()
	mage.applyMasterOfElements()

	mage.registerCombustionCD()

	// Incinerate (DBC): +3%/rank crit for Fire Blast, Scorch, Flamestrike and Blast Wave.
	if mage.Talents.Incinerate > 0 {
		bonusCrit := 3 * float64(mage.Talents.Incinerate) * core.SpellCritRatingPerCritChance
		mage.OnSpellRegistered(func(spell *core.Spell) {
			switch spell.SpellCode {
			case SpellCode_MageFireBlast, SpellCode_MageScorch, SpellCode_MageFlamestrike, SpellCode_MageBlastWave:
				spell.BonusCritRating += bonusCrit
			}
		})
	}

	// Burning Soul
	if mage.Talents.BurningSoul > 0 {
		threatMultiplier := 1 - .05*float64(mage.Talents.BurningSoul) // DBC: 5%/rank
		mage.OnSpellRegistered(func(spell *core.Spell) {
			if spell.SpellSchool.Matches(core.SpellSchoolFire) && spell.Flags.Matches(SpellFlagMage) {
				spell.ThreatMultiplier *= threatMultiplier
			}
		})
	}

	// Critical Mass
	if mage.Talents.CriticalMass > 0 {
		bonusCrit := 1 * float64(mage.Talents.CriticalMass) * core.SpellCritRatingPerCritChance // DBC: 1%/rank
		mage.OnSpellRegistered(func(spell *core.Spell) {
			if spell.SpellSchool.Matches(core.SpellSchoolFire) && spell.Flags.Matches(SpellFlagMage) {
				spell.BonusCritRating += bonusCrit
			}
		})
	}

	// Fire Power
	if mage.Talents.FirePower > 0 {
		bonusDamageMultiplierAdditive := 0.02 * float64(mage.Talents.FirePower)
		mage.OnSpellRegistered(func(spell *core.Spell) {
			// Fire Power buffs pretty much all mage fire spells EXCEPT ignite
			if spell.SpellSchool.Matches(core.SpellSchoolFire) && spell.Flags.Matches(SpellFlagMage) && spell != mage.Ignite {
				spell.DamageMultiplierAdditive += bonusDamageMultiplierAdditive
			}
		})
	}
}

func (mage *Mage) applyFrostTalents() {
	mage.registerColdSnapCD()
	mage.registerIceBarrierSpell()
	mage.applyWintersChill()

	// Elemental Precision
	if mage.Talents.ElementalPrecision > 0 {
		bonusHit := 1 * float64(mage.Talents.ElementalPrecision) * core.SpellHitRatingPerHitChance // DBC: 1%/rank

		mage.OnSpellRegistered(func(spell *core.Spell) {
			if spell.Flags.Matches(SpellFlagMage) && (spell.SpellSchool.Matches(core.SpellSchoolFire) || spell.SpellSchool.Matches(core.SpellSchoolFrost)) {
				spell.BonusHitRating += bonusHit
			}
		})
	}

	// Frost Shards (DBC): +20%/rank Frost crit strike damage bonus.
	if mage.Talents.FrostShards > 0 {
		critBonus := .20 * float64(mage.Talents.FrostShards)
		mage.OnSpellRegistered(func(spell *core.Spell) {
			if spell.SpellSchool.Matches(core.SpellSchoolFrost) && spell.Flags.Matches(SpellFlagMage) {
				spell.CritDamageBonus += critBonus
			}
		})
	}

	// Rimebound (DBC): +1%/rank Frost damage.
	if mage.Talents.Rimebound > 0 {
		bonus := 0.01 * float64(mage.Talents.Rimebound)
		mage.OnSpellRegistered(func(spell *core.Spell) {
			if spell.SpellSchool.Matches(core.SpellSchoolFrost) && spell.Flags.Matches(SpellFlagMage) {
				spell.DamageMultiplierAdditive += bonus
			}
		})
	}

	// Arctic Gale (DBC): +2%/rank Frost crit chance.
	if mage.Talents.ArcticGale > 0 {
		bonusCrit := 2 * float64(mage.Talents.ArcticGale) * core.SpellCritRatingPerCritChance
		mage.OnSpellRegistered(func(spell *core.Spell) {
			if spell.SpellSchool.Matches(core.SpellSchoolFrost) && spell.Flags.Matches(SpellFlagMage) {
				spell.BonusCritRating += bonusCrit
			}
		})
	}

	// Frost Channeling
	if mage.Talents.FrostChanneling > 0 {
		manaCostMultiplier := 3 * mage.Talents.FrostChanneling            // DBC: 3%/rank
		threatMultiplier := 1 - .06*float64(mage.Talents.FrostChanneling) // DBC: 6%/rank
		mage.OnSpellRegistered(func(spell *core.Spell) {
			if spell.SpellSchool.Matches(core.SpellSchoolFrost) && spell.Flags.Matches(SpellFlagMage) {
				spell.Cost.Multiplier -= manaCostMultiplier
				spell.ThreatMultiplier *= threatMultiplier
			}
		})
	}
}

func (mage *Mage) applyArcaneConcentration() {
	if mage.Talents.ArcaneConcentration == 0 {
		return
	}

	procChance := 0.02 * float64(mage.Talents.ArcaneConcentration)

	mage.ClearcastingAura = mage.RegisterAura(core.Aura{
		Label:    "Clearcasting",
		ActionID: core.ActionID{SpellID: 12577},
		Duration: time.Second * 15,
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			aura.Unit.PseudoStats.SchoolCostMultiplier.AddToMagicSchools(-100)
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			aura.Unit.PseudoStats.SchoolCostMultiplier.AddToMagicSchools(100)
		},
		OnCastComplete: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
			// OnCastComplete is called after OnSpellHitDealt / etc, so don't deactivate if it was just activated.
			if aura.RemainingDuration(sim) == aura.Duration {
				return
			}
			if !spell.Flags.Matches(SpellFlagMage) {
				return
			}
			if spell.Cost != nil && spell.Cost.GetCurrentCost() == 0 {
				return
			}
			aura.Deactivate(sim)
		},
	})

	mage.RegisterAura(core.Aura{
		Label:    "Arcane Concentration",
		Duration: core.NeverExpires,
		OnReset: func(aura *core.Aura, sim *core.Simulation) {
			aura.Activate(sim)
		},
		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if !result.Landed() || !spell.Flags.Matches(SpellFlagMage) || spell.SpellCode == SpellCode_MageArcaneMissiles {
				return
			}

			// TODO: Classic verify arcane missile proc chance
			// Arcane Missile ticks can proc CC, just at a low rate of about 1.5% with 5/5 Arcane Concentration
			// if spell == mage.ArcaneMissilesTickSpell {
			// 	procChance *= 0.15
			// }

			if sim.Proc(procChance, "Arcane Concentration") {
				mage.ClearcastingAura.Activate(sim)
			}
		},
	})
}

func (mage *Mage) registerPresenceOfMindCD() {
	if true { // TODO: Presence of Mind removed from custom tree
		return
	}

	actionID := core.ActionID{SpellID: 12043}
	cooldown := time.Second * 180

	affectedSpells := []*core.Spell{}
	pomAura := mage.RegisterAura(core.Aura{
		Label:    "Presence of Mind",
		ActionID: actionID,
		Duration: time.Second * 15,
		OnInit: func(aura *core.Aura, sim *core.Simulation) {
			for spellIdx := range mage.Spellbook {
				if spell := mage.Spellbook[spellIdx]; spell.DefaultCast.CastTime > 0 {
					affectedSpells = append(affectedSpells, spell)
				}
			}
		},
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			core.Each(affectedSpells, func(spell *core.Spell) {
				spell.CastTimeMultiplier -= 1
			})
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			core.Each(affectedSpells, func(spell *core.Spell) {
				spell.CastTimeMultiplier += 1
			})
			mage.PresenceOfMind.CD.Use(sim)
		},
		OnCastComplete: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
			if !slices.Contains(affectedSpells, spell) {
				return
			}

			aura.Deactivate(sim)
		},
	})

	mage.PresenceOfMind = mage.RegisterSpell(core.SpellConfig{
		ActionID: actionID,
		Flags:    core.SpellFlagNoOnCastComplete,
		Cast: core.CastConfig{
			CD: core.Cooldown{
				Timer:    mage.NewTimer(),
				Duration: cooldown,
			},
		},
		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return mage.GCD.IsReady(sim)
		},
		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			pomAura.Activate(sim)
		},
	})

	mage.AddMajorCooldown(core.MajorCooldown{
		Spell: mage.PresenceOfMind,
		Type:  core.CooldownTypeDPS,
	})
}

func (mage *Mage) registerArcanePowerCD() {
	if !mage.Talents.ArcanePower {
		return
	}

	actionID := core.ActionID{SpellID: 12042}

	affectedSpells := []*core.Spell{}

	mage.OnSpellRegistered(func(spell *core.Spell) {
		if spell.Flags.Matches(SpellFlagMage) {
			affectedSpells = append(affectedSpells, spell)
		}
	})

	mage.ArcanePowerAura = mage.RegisterAura(core.Aura{
		Label:    "Arcane Power",
		ActionID: actionID,
		// DBC: Arcane Power lasts 20s (calculator), +5s per rank.
		Duration: time.Second * time.Duration(20+5*mage.Talents.ImprovedArcanePower),
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			for _, spell := range affectedSpells {
				spell.DamageMultiplierAdditive += 0.3
				if spell.Cost != nil {
					spell.Cost.Multiplier += 30
				}
			}
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			for _, spell := range affectedSpells {
				spell.DamageMultiplierAdditive -= 0.3
				if spell.Cost != nil {
					spell.Cost.Multiplier -= 30
				}
			}
		},
	})
	core.RegisterPercentDamageModifierEffect(mage.ArcanePowerAura, 1.3)

	spell := mage.RegisterSpell(core.SpellConfig{
		ActionID: actionID,
		Flags:    core.SpellFlagNoOnCastComplete,
		Cast: core.CastConfig{
			CD: core.Cooldown{
				Timer: mage.NewTimer(),
				// DBC: Improved Arcane Power -60s cooldown per rank.
				Duration: time.Second * time.Duration(180-60*mage.Talents.ImprovedArcanePower),
			},
		},
		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			mage.ArcanePowerAura.Activate(sim)
		},
	})

	mage.AddMajorCooldown(core.MajorCooldown{
		Spell: spell,
		Type:  core.CooldownTypeDPS,
	})
}

func (mage *Mage) applyImprovedScorch() {
	if mage.Talents.ImprovedScorch == 0 {
		return
	}

	mage.ImprovedScorchAuras = mage.NewEnemyAuraArray(func(unit *core.Unit) *core.Aura {
		return core.ImprovedScorchAura(unit)
	})
}

func (mage *Mage) applyMasterOfElements() {
	if mage.Talents.MasterOfElements == 0 {
		return
	}

	refundCoeff := 0.1 * float64(mage.Talents.MasterOfElements)
	manaMetrics := mage.NewManaMetrics(core.ActionID{SpellID: 29076})

	mage.RegisterAura(core.Aura{
		Label:    "Master of Elements",
		Duration: core.NeverExpires,
		OnReset: func(aura *core.Aura, sim *core.Simulation) {
			aura.Activate(sim)
		},
		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if spell.ProcMask.Matches(core.ProcMaskMeleeOrRanged) {
				return
			}
			if spell.CurCast.Cost == 0 {
				return
			}
			if result.DidCrit() {
				mage.AddMana(sim, spell.Cost.BaseCost*refundCoeff, manaMetrics)
			}
		},
	})
}

func (mage *Mage) registerCombustionCD() {
	if !mage.Talents.Combustion {
		return
	}

	actionID := core.ActionID{SpellID: 11129}
	cd := core.Cooldown{
		Timer:    mage.NewTimer(),
		Duration: time.Minute * 3,
	}

	var fireSpells []*core.Spell
	mage.OnSpellRegistered(func(spell *core.Spell) {
		if spell.SpellSchool.Matches(core.SpellSchoolFire) && spell.Flags.Matches(SpellFlagMage) {
			fireSpells = append(fireSpells, spell)
		}
	})

	numCrits := 0
	critPerStack := 10.0 * core.SpellCritRatingPerCritChance

	mage.CombustionAura = mage.RegisterAura(core.Aura{
		Label:     "Combustion",
		ActionID:  actionID,
		Duration:  core.NeverExpires,
		MaxStacks: 20,
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			numCrits = 0
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			cd.Use(sim)
			mage.UpdateMajorCooldowns()
		},
		OnStacksChange: func(aura *core.Aura, sim *core.Simulation, oldStacks int32, newStacks int32) {
			bonusCrit := critPerStack * float64(newStacks-oldStacks)
			for _, spell := range fireSpells {
				spell.BonusCritRating += bonusCrit
			}
		},
		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if !result.Landed() || numCrits >= 3 || !spell.SpellSchool.Matches(core.SpellSchoolFire) || !spell.Flags.Matches(SpellFlagMage) {
				return
			}

			// Ignite, Living Bomb explosions, and Fire Blast with Overheart don't consume crit stacks
			// To Do: Classic - I don't believe ignite can crit so can probably remove this check?
			if spell.SpellCode == SpellCode_MageIgnite {
				return
			}

			// TODO: This wont work properly with flamestrike
			aura.AddStack(sim)

			if result.DidCrit() {
				numCrits++
				if numCrits == 3 {
					aura.Deactivate(sim)
				}
			}
		},
	})

	spell := mage.RegisterSpell(core.SpellConfig{
		ActionID: actionID,
		Flags:    core.SpellFlagNoOnCastComplete,
		Cast: core.CastConfig{
			CD: cd,
		},
		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return !mage.CombustionAura.IsActive()
		},
		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			mage.CombustionAura.Activate(sim)
			mage.CombustionAura.AddStack(sim)
		},
	})

	mage.AddMajorCooldown(core.MajorCooldown{
		Spell: spell,
		Type:  core.CooldownTypeDPS,
	})
}

func (mage *Mage) registerColdSnapCD() {
	if true { // TODO: Cold Snap removed from custom tree
		return
	}

	// Grab all frost spells with a CD > 0
	var affectedSpells = []*core.Spell{}
	mage.OnSpellRegistered(func(spell *core.Spell) {
		if spell.SpellSchool.Matches(core.SpellSchoolFrost) && spell.CD.Duration > 0 {
			affectedSpells = append(affectedSpells, spell)
		}
	})

	spell := mage.RegisterSpell(core.SpellConfig{
		ActionID: core.ActionID{SpellID: 12472},
		Flags:    core.SpellFlagNoOnCastComplete,

		Cast: core.CastConfig{
			CD: core.Cooldown{
				Timer:    mage.NewTimer(),
				Duration: time.Duration(time.Minute * 10),
			},
		},
		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			for _, spell := range affectedSpells {
				spell.CD.Reset()
			}
		},
	})

	mage.AddMajorCooldown(core.MajorCooldown{
		Spell: spell,
		Type:  core.CooldownTypeDPS,
	})
}

func (mage *Mage) applyWintersChill() {
	if mage.Talents.WintersChill == 0 {
		return
	}

	procChance := float64(mage.Talents.WintersChill) * 0.2

	wcAuras := mage.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		return core.WintersChillAura(target)
	})
	mage.Env.RegisterPreFinalizeEffect(func() {
		for _, spell := range mage.GetSpellsMatchingSchool(core.SpellSchoolFrost) {
			spell.RelatedAuras = append(spell.RelatedAuras, wcAuras)
		}
	})

	mage.RegisterAura(core.Aura{
		Label:    "Winters Chill Talent",
		Duration: core.NeverExpires,
		OnReset: func(aura *core.Aura, sim *core.Simulation) {
			aura.Activate(sim)
		},
		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if !result.Landed() || !spell.SpellSchool.Matches(core.SpellSchoolFrost) {
				return
			}

			if sim.Proc(procChance, "Winters Chill") {
				aura := wcAuras.Get(result.Target)
				aura.Activate(sim)
				if aura.IsActive() {
					aura.AddStack(sim)
				}
			}
		},
	})
}

// Pyromania (DBC): -3/-5% cast time on Fireball and Pyroblast, -50/-100% on Flamestrike,
// and -15/-30s Blast Wave cooldown.
func (mage *Mage) applyPyromania() {
	if mage.Talents.Pyromania == 0 {
		return
	}
	rank := float64(mage.Talents.Pyromania)
	mage.OnSpellRegistered(func(spell *core.Spell) {
		switch spell.SpellCode {
		case SpellCode_MageFireball, SpellCode_MagePyroblast:
			spell.CastTimeMultiplier -= []float64{0, .03, .05}[mage.Talents.Pyromania]
		case SpellCode_MageFlamestrike:
			spell.CastTimeMultiplier -= 0.50 * rank
		case SpellCode_MageBlastWave:
			spell.CD.Duration -= time.Second * 15 * time.Duration(mage.Talents.Pyromania)
		}
	})
}

// Hot Streak (DBC): 5%/rank chance on Fire spell crit to make the next Scorch or
// Pyroblast instant and free.
func (mage *Mage) applyHotStreak() {
	if mage.Talents.HotStreak == 0 {
		return
	}
	procChance := 0.05 * float64(mage.Talents.HotStreak)
	var affected []*core.Spell
	mage.OnSpellRegistered(func(spell *core.Spell) {
		if spell.SpellCode == SpellCode_MageScorch || spell.SpellCode == SpellCode_MagePyroblast {
			affected = append(affected, spell)
		}
	})
	hotStreak := mage.RegisterAura(core.Aura{
		Label:    "Hot Streak",
		ActionID: core.ActionID{SpellID: 33905},
		Duration: time.Second * 15,
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			for _, spell := range affected {
				spell.CastTimeMultiplier -= 1
				spell.Cost.Multiplier -= 100
			}
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			for _, spell := range affected {
				spell.CastTimeMultiplier += 1
				spell.Cost.Multiplier += 100
			}
		},
		OnCastComplete: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
			if aura.RemainingDuration(sim) == aura.Duration {
				return
			}
			if spell.SpellCode == SpellCode_MageScorch || spell.SpellCode == SpellCode_MagePyroblast {
				aura.Deactivate(sim)
			}
		},
	})
	mage.RegisterAura(core.Aura{
		Label:    "Hot Streak Talent",
		Duration: core.NeverExpires,
		OnReset: func(aura *core.Aura, sim *core.Simulation) {
			aura.Activate(sim)
		},
		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if !result.DidCrit() || !spell.SpellSchool.Matches(core.SpellSchoolFire) || !spell.Flags.Matches(SpellFlagMage) || spell.SpellCode == SpellCode_MageIgnite {
				return
			}
			if !hotStreak.IsActive() && sim.Proc(procChance, "Hot Streak") {
				hotStreak.Activate(sim)
			}
		},
	})
}

// Arcane Resilience (DBC): armor equal to 50%/rank of Intellect.
func (mage *Mage) applyArcaneResilience() {
	if mage.Talents.ArcaneResilience == 0 {
		return
	}
	mage.AddStatDependency(stats.Intellect, stats.Armor, 0.5*float64(mage.Talents.ArcaneResilience))
}

// Spell Twisting (Vanilla+ calculator): Fire spells give +15% crit on the next Frost spell,
// Frost spells give +15% crit on the next Fire spell, Arcane spells give it to both.
func (mage *Mage) applySpellTwisting() {
	if !mage.Talents.SpellTwisting {
		return
	}
	crit := 15.0 * core.SpellCritRatingPerCritChance
	var fireSpells, frostSpells []*core.Spell
	mage.OnSpellRegistered(func(spell *core.Spell) {
		if !spell.Flags.Matches(SpellFlagMage) {
			return
		}
		if spell.SpellSchool.Matches(core.SpellSchoolFire) {
			fireSpells = append(fireSpells, spell)
		} else if spell.SpellSchool.Matches(core.SpellSchoolFrost) {
			frostSpells = append(frostSpells, spell)
		}
	})
	makeAura := func(label string, id int32, targets func() []*core.Spell) *core.Aura {
		return mage.RegisterAura(core.Aura{
			Label:    label,
			ActionID: core.ActionID{SpellID: id},
			Duration: time.Second * 15,
			OnGain: func(aura *core.Aura, sim *core.Simulation) {
				for _, s := range targets() {
					s.BonusCritRating += crit
				}
			},
			OnExpire: func(aura *core.Aura, sim *core.Simulation) {
				for _, s := range targets() {
					s.BonusCritRating -= crit
				}
			},
			OnCastComplete: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
				if aura.RemainingDuration(sim) == aura.Duration {
					return
				}
				for _, s := range targets() {
					if s == spell {
						aura.Deactivate(sim)
						return
					}
				}
			},
		})
	}
	frostBuff := makeAura("Spell Twisting: Ice", 33876, func() []*core.Spell { return frostSpells })
	fireBuff := makeAura("Spell Twisting: Fire", 33877, func() []*core.Spell { return fireSpells })
	core.MakePermanent(mage.RegisterAura(core.Aura{
		Label: "Spell Twisting",
		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if !spell.Flags.Matches(SpellFlagMage) || !spell.ProcMask.Matches(core.ProcMaskSpellDamage) || spell.SpellCode == SpellCode_MageIgnite {
				return
			}
			switch {
			case spell.SpellSchool.Matches(core.SpellSchoolFire):
				frostBuff.Activate(sim)
			case spell.SpellSchool.Matches(core.SpellSchoolFrost):
				fireBuff.Activate(sim)
			case spell.SpellSchool.Matches(core.SpellSchoolArcane):
				frostBuff.Activate(sim)
				fireBuff.Activate(sim)
			}
		},
	}))
}
