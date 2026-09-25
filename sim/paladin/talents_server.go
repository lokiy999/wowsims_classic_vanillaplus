package paladin

import (
	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/stats"
)

// Talents from the server's custom trees (text from ui/core/talents/trees/paladin.json).
func (paladin *Paladin) applyServerTalents() {
	// The Revenant: reduces the chance you are critically hit by 2% per rank.
	paladin.PseudoStats.ReducedCritTakenChance += 0.02 * float64(paladin.Talents.TheRevenant)

	// Shield of Faith: all spell damage taken -5% per rank (-15% at 3/3).
	if paladin.Talents.ShieldOfFaith > 0 {
		for school := stats.SchoolIndexArcane; school < stats.SchoolLen; school++ {
			paladin.PseudoStats.SchoolDamageTakenMultiplier[school] *= 1 - 0.05*float64(paladin.Talents.ShieldOfFaith)
		}
	}

	paladin.applyMorale()
	paladin.applyIllumination()
}

// Morale: "your melee attacks have a 4% chance (per rank) to restore 125 Mana". The fear part is not modeled.
func (paladin *Paladin) applyMorale() {
	if paladin.Talents.Morale == 0 {
		return
	}
	chance := 0.04 * float64(paladin.Talents.Morale)
	metrics := paladin.NewManaMetrics(core.ActionID{SpellID: []int32{0, 33462, 33463, 33464, 33465, 33466}[paladin.Talents.Morale]})

	core.MakePermanent(paladin.RegisterAura(core.Aura{
		Label: "Morale",
		OnSpellHitDealt: func(_ *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if result.Landed() && spell.ProcMask.Matches(core.ProcMaskMelee) && sim.RandomFloat("Morale") < chance {
				paladin.AddMana(sim, 125, metrics)
			}
		},
	}))
}

// Illumination: crits from Flash of Light, Holy Light, Holy Shock, Exorcism, Holy Wrath and Hammer of Wrath return part of the spell's base mana
// cost. Rank 1 says 20%, rank 5 says 50%; the server tooltips for ranks 2-4 repeat the rank 5 text, so the middle
// ranks are interpolated.
func (paladin *Paladin) applyIllumination() {
	if paladin.Talents.Illumination == 0 {
		return
	}
	pct := []float64{0, 0.2, 0.275, 0.35, 0.425, 0.5}[paladin.Talents.Illumination]
	metrics := paladin.NewManaMetrics(core.ActionID{SpellID: []int32{0, 20210, 20212, 20213, 20214, 20215}[paladin.Talents.Illumination]})

	core.MakePermanent(paladin.RegisterAura(core.Aura{
		Label: "Illumination",
		OnSpellHitDealt: func(_ *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if !result.DidCrit() {
				return
			}
			switch spell.SpellCode {
			case SpellCode_PaladinHolyShock, SpellCode_PaladinExorcism, SpellCode_PaladinHolyWrath, SpellCode_PaladinHammerOfWrath:
				paladin.AddMana(sim, pct*spell.DefaultCast.Cost, metrics)
			}
		},
		OnHealDealt: func(_ *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if !result.DidCrit() {
				return
			}
			switch spell.SpellCode {
			case SpellCode_PaladinHolyShockHeal, SpellCode_PaladinHolyLight, SpellCode_PaladinFlashOfLight:
				paladin.AddMana(sim, pct*spell.DefaultCast.Cost, metrics)
			}
		},
	}))
}
