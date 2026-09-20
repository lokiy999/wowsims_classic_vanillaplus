package core

import (
	"math"

	"github.com/wowsims/classic/sim/core/stats"
)

// Flametongue Totem buff (rank 4, level 58), given to non-shaman classes through the weapon imbue option like the
// Windfury Totem. From the server data and confirmed in game: +40 spell damage, and every main-hand hit deals
// 1217 * weapon speed / 100 additional Fire damage (15.8 to 48.7 across weapon speeds, 12.17 per second of speed).
// Improved Weapon Totems adds +25%/50% to the effect (both parts).
const flametongueTotemSpellPower = 40.0
const flametongueTotemDamagePerSecond = 1217.0 / 100
const flametongueTotemProcSpellID = 16389

func ApplyFlametongueTotem(character *Character) {
	weapon := character.MainHand()
	if weapon == nil {
		return
	}

	effect := 1 + 0.25*float64(character.ImprovedWeaponTotems)
	character.AddStats(stats.Stats{stats.SpellPower: math.Floor(flametongueTotemSpellPower * effect)})

	procSpell := character.RegisterSpell(SpellConfig{
		ActionID:    ActionID{SpellID: flametongueTotemProcSpellID},
		SpellSchool: SpellSchoolFire,
		DefenseType: DefenseTypeMagic,
		ProcMask:    ProcMaskSpellDamageProc,
		Flags:       SpellFlagNoOnCastComplete | SpellFlagPassiveSpell,

		DamageMultiplier: effect,
		ThreatMultiplier: 1,
		BonusCoefficient: .1,

		ApplyEffects: func(sim *Simulation, target *Unit, spell *Spell) {
			if weapon.SwingSpeed != 0 {
				damage := flametongueTotemDamagePerSecond * weapon.SwingSpeed
				spell.CalcAndDealDamage(sim, target, damage, spell.OutcomeMagicHitAndCrit)
			}
		},
	})

	MakePermanent(character.RegisterAura(Aura{
		Label: "Flametongue Totem",
		OnSpellHitDealt: func(aura *Aura, sim *Simulation, spell *Spell, result *SpellResult) {
			if result.Landed() && spell.ProcMask.Matches(ProcMaskMeleeMH) {
				procSpell.Cast(sim, result.Target)
			}
		},
	}))
}
