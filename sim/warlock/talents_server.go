package warlock

import (
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/stats"
)

// Talents from the server's custom trees (text from ui/core/talents/trees/warlock.json).
func (warlock *Warlock) applyServerTalents() {
	warlock.applyFelPact()
	warlock.applyFeedingDemons()

	// Demonic Onslaught: +4% melee and spell crit per rank for the Imp, Voidwalker, Succubus and Felhunter.
	if warlock.Talents.DemonicOnslaught > 0 {
		crit := 4 * float64(warlock.Talents.DemonicOnslaught)
		for _, pet := range warlock.BasePets {
			pet.AddStats(stats.Stats{
				stats.MeleeCrit: crit * core.CritRatingPerCritChance,
				stats.SpellCrit: crit * core.SpellCritRatingPerCritChance,
			})
		}
	}

	// Destructive Reach: threat of Destruction spells -10% per rank.
	if warlock.Talents.DestructiveReach > 0 {
		mult := 1 - 0.1*float64(warlock.Talents.DestructiveReach)
		warlock.OnSpellRegistered(func(spell *core.Spell) {
			switch spell.SpellCode {
			case SpellCode_WarlockShadowBolt, SpellCode_WarlockImmolate, SpellCode_WarlockSearingPain,
				SpellCode_WarlockSoulFire, SpellCode_WarlockConflagrate, SpellCode_WarlockShadowburn:
				spell.ThreatMultiplier *= mult
			}
		})
	}

	warlock.applyDefiler()
}

// Defiler (DBC 34305): time between periodic ticks and duration -20%, global cooldown -0.3 sec on the
// warlock's own spells (not the pet's). The number of ticks stays the same, so DoTs and channels deal the
// same damage in less time. The extra Drain target is not modeled (single target sim).
func (warlock *Warlock) applyDefiler() {
	if !warlock.Talents.Defiler {
		return
	}
	warlock.GetOrRegisterAura(core.Aura{
		Label: "Defiler",
		OnInit: func(aura *core.Aura, sim *core.Simulation) {
			for _, spell := range warlock.Spellbook {
				if spell.DefaultCast.GCD > 0 {
					spell.DefaultCast.GCD = max(core.GCDMin, spell.DefaultCast.GCD-time.Millisecond*300)
				}
				dots := append([]*core.Dot{spell.AOEDot()}, spell.Dots()...)
				for _, dot := range dots {
					if dot != nil {
						dot.TickLength = time.Duration(float64(dot.TickLength) * 0.8)
						dot.RecomputeAuraDuration()
					}
				}
			}
		},
	})
}

// Prolonged Misery: Corruption, Curse of Agony, Curse of Exhaustion and Immolate last 2 sec longer per rank.
// Returns the extra ticks for a DoT with the given tick length.
func (warlock *Warlock) prolongedMiseryTicks(tickLength time.Duration) int32 {
	return int32(time.Duration(warlock.Talents.ProlongedMisery) * 2 * time.Second / tickLength)
}

// Fel Pact (Spell.csv 33952-33956): the demon's critical hits with spells or abilities have an 8% chance per rank
// (33957-33961) to give the warlock 15% of total mana over 15 sec (33962: 1% every sec). The other half (the
// warlock's casts heal the demon) does nothing in a DPS sim.
func (warlock *Warlock) applyFelPact() {
	if warlock.Talents.FelPact == 0 {
		return
	}

	actionID := core.ActionID{SpellID: 33962}
	manaMetrics := warlock.NewManaMetrics(actionID)
	var regen *core.PendingAction
	regenAura := warlock.RegisterAura(core.Aura{
		Label:    "Fel Pact",
		ActionID: actionID,
		Duration: time.Second * 15,
		OnReset: func(aura *core.Aura, sim *core.Simulation) {
			regen = nil
		},
	})

	procChance := 0.08 * float64(warlock.Talents.FelPact)
	for _, pet := range warlock.BasePets {
		core.MakeProcTriggerAura(&pet.Unit, core.ProcTrigger{
			Name:       "Fel Pact Demon",
			ActionID:   core.ActionID{SpellID: []int32{0, 33957, 33958, 33959, 33960, 33961}[warlock.Talents.FelPact]},
			Callback:   core.CallbackOnSpellHitDealt,
			ProcMask:   core.ProcMaskSpellDamage | core.ProcMaskMeleeSpecial | core.ProcMaskRangedSpecial,
			Outcome:    core.OutcomeCrit,
			ProcChance: procChance,
			Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				// A new proc restarts the 15 sec regeneration.
				if regen != nil {
					regen.Cancel(sim)
				}
				regen = core.StartPeriodicAction(sim, core.PeriodicActionOptions{
					Period:   time.Second,
					NumTicks: 15,
					OnAction: func(sim *core.Simulation) {
						warlock.AddMana(sim, 0.01*warlock.MaxMana(), manaMetrics)
					},
				})
				regenAura.Activate(sim)
			},
		})
	}
}

// Feeding Demons (Spell.csv 33969-33970): the warlock's spell critical strikes have a 50% chance per rank to
// restore 6 mana per level to the demon (33971).
func (warlock *Warlock) applyFeedingDemons() {
	if warlock.Talents.FeedingDemons == 0 || len(warlock.BasePets) == 0 {
		return
	}

	actionID := core.ActionID{SpellID: 33971}
	petMetrics := make(map[*WarlockPet]*core.ResourceMetrics, len(warlock.BasePets))
	for _, pet := range warlock.BasePets {
		petMetrics[pet] = pet.NewManaMetrics(actionID)
	}
	mana := 6 * float64(warlock.Level)

	core.MakeProcTriggerAura(&warlock.Unit, core.ProcTrigger{
		Name:       "Feeding Demons",
		ActionID:   core.ActionID{SpellID: []int32{0, 33969, 33970}[warlock.Talents.FeedingDemons]},
		Callback:   core.CallbackOnSpellHitDealt,
		ProcMask:   core.ProcMaskSpellDamage,
		Outcome:    core.OutcomeCrit,
		ProcChance: 0.5 * float64(warlock.Talents.FeedingDemons),
		Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if pet := warlock.ActivePet; pet != nil && pet.IsEnabled() {
				pet.AddMana(sim, mana, petMetrics[pet])
			}
		},
	})
}

// Herald of Woe (Spell.csv 33931, 33932, 34006): cooldown of Amplify Curse, Death and Decay and Death Coil -20% per rank.
func (warlock *Warlock) heraldOfWoe(cd time.Duration) time.Duration {
	return time.Duration(float64(cd) * (1 - 0.2*float64(warlock.Talents.HeraldOfWoe)))
}
