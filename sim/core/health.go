package core

import (
	"time"

	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/stats"
)

type healthBar struct {
	unit *Unit

	currentHealth float64

	DamageTakenHealthMetrics *ResourceMetrics

	// Death delay (paladin Stoicism): a killing blow starts a window in which all damage and healing is held back,
	// then applied at once; the unit only dies if its health is still 0 after that.
	deathDelay        time.Duration
	deathDelayActive  bool
	deathDelayNet     float64
	deathDelayMetrics *ResourceMetrics
}

// SetDeathDelay defers death by the given duration after a killing blow (see healthBar.deathDelay).
func (unit *Unit) SetDeathDelay(delay time.Duration, actionID ActionID) {
	unit.healthBar.deathDelay = delay
	unit.healthBar.deathDelayMetrics = unit.NewHealthMetrics(actionID)
}

func (unit *Unit) EnableHealthBar() {
	unit.healthBar = healthBar{
		unit:                     unit,
		DamageTakenHealthMetrics: unit.NewHealthMetrics(ActionID{OtherID: proto.OtherAction_OtherActionDamageTaken}),
	}
}

func (unit *Unit) HasHealthBar() bool {
	return unit.healthBar.unit != nil
}

func (hb *healthBar) reset(_ *Simulation) {
	if hb.unit == nil {
		return
	}
	hb.currentHealth = hb.MaxHealth()
	hb.deathDelayActive = false
	hb.deathDelayNet = 0
}

func (hb *healthBar) MaxHealth() float64 {
	return hb.unit.stats[stats.Health]
}

func (hb *healthBar) CurrentHealth() float64 {
	return hb.currentHealth
}

func (hb *healthBar) CurrentHealthPercent() float64 {
	return hb.currentHealth / hb.unit.stats[stats.Health]
}

func (hb *healthBar) GainHealth(sim *Simulation, amount float64, metrics *ResourceMetrics) {
	if amount < 0 {
		panic("Trying to gain negative health!")
	}
	if hb.deathDelayActive {
		hb.deathDelayNet += amount
		return
	}

	oldHealth := hb.currentHealth
	newHealth := min(oldHealth+amount, hb.unit.MaxHealth())
	metrics.AddEvent(amount, newHealth-oldHealth)

	if sim.Log != nil {
		hb.unit.Log(sim, "Gained %0.3f health from %s (%0.3f --> %0.3f).", amount, metrics.ActionID, oldHealth, newHealth)
	}

	hb.currentHealth = newHealth
}

func (hb *healthBar) RemoveHealth(sim *Simulation, amount float64) {
	if amount < 0 {
		panic("Trying to remove negative health!")
	}

	oldHealth := hb.currentHealth
	newHealth := max(oldHealth-amount, 0)
	metrics := hb.DamageTakenHealthMetrics
	metrics.AddEvent(-amount, newHealth-oldHealth)

	// TMI calculations need timestamps and Max HP information for each damage taken event
	if hb.unit.Metrics.isTanking {
		entry := tmiListItem{
			Timestamp:      sim.CurrentTime,
			WeightedDamage: amount / hb.MaxHealth(),
		}
		hb.unit.Metrics.tmiList = append(hb.unit.Metrics.tmiList, entry)
	}

	if sim.Log != nil {
		hb.unit.Log(sim, "Spent %0.3f health from %s (%0.3f --> %0.3f).", amount, metrics.ActionID, oldHealth, newHealth)
	}

	hb.currentHealth = newHealth
}

var ChanceOfDeathAuraLabel = "Chance of Death"

func (character *Character) trackChanceOfDeath(healingModel *proto.HealingModel) {
	character.Unit.Metrics.isTanking = false
	for _, target := range character.Env.Encounter.TargetUnits {
		if target.CurrentTarget == &character.Unit {
			character.Unit.Metrics.isTanking = true
		}
	}
	if !character.Unit.Metrics.isTanking {
		return
	}

	if healingModel == nil {
		return
	}

	character.Unit.Metrics.tmiBin = healingModel.BurstWindow

	character.RegisterAura(Aura{
		Label:    ChanceOfDeathAuraLabel,
		Duration: NeverExpires,
		OnReset: func(aura *Aura, sim *Simulation) {
			aura.Activate(sim)
		},
		OnSpellHitTaken: func(aura *Aura, sim *Simulation, spell *Spell, result *SpellResult) {
			if result.Damage > 0 {
				character.takeDamageForDeathTracking(sim, result.Damage)
			}
		},
		OnPeriodicDamageTaken: func(aura *Aura, sim *Simulation, spell *Spell, result *SpellResult) {
			if result.Damage > 0 {
				character.takeDamageForDeathTracking(sim, result.Damage)
			}
		},
	})

	if healingModel.Hps != 0 {
		character.applyHealingModel(healingModel)
	}
}

func (character *Character) takeDamageForDeathTracking(sim *Simulation, damage float64) {
	unit := &character.Unit
	hb := &unit.healthBar
	if hb.deathDelayActive {
		hb.deathDelayNet -= damage
		return
	}
	if hb.deathDelay > 0 && damage >= hb.currentHealth && !unit.Metrics.Died {
		hb.deathDelayActive = true
		hb.deathDelayNet = -damage
		if sim.Log != nil {
			character.Log(sim, "Killing blow of %0.1f deferred for %s", damage, hb.deathDelay)
		}
		StartDelayedAction(sim, DelayedActionOptions{
			DoAt: sim.CurrentTime + hb.deathDelay,
			OnAction: func(sim *Simulation) {
				hb.deathDelayActive = false
				net := hb.deathDelayNet
				hb.deathDelayNet = 0
				if net < 0 {
					unit.RemoveHealth(sim, -net)
				} else if net > 0 {
					hb.GainHealth(sim, net, hb.deathDelayMetrics)
				}
				character.checkDeath(sim)
			},
		})
		return
	}
	unit.RemoveHealth(sim, damage)
	character.checkDeath(sim)
}

func (character *Character) checkDeath(sim *Simulation) {
	if character.CurrentHealth() <= 0 && !character.Metrics.Died {
		character.Metrics.Died = true
		if sim.Log != nil {
			character.Log(sim, "Dead")
		}
	}
}

func (character *Character) applyHealingModel(healingModel *proto.HealingModel) {
	// Store variance parameters for healing cadence. Note that low rolls on
	// cadence are special cased here so that the model is still well-behaved
	// when CadenceVariation exceeds CadenceSeconds.
	medianCadence := healingModel.CadenceSeconds
	if medianCadence == 0 {
		medianCadence = 2.0
	}
	minCadence := max(0.0, medianCadence-healingModel.CadenceVariation)
	cadenceVariationLow := medianCadence - minCadence

	healthMetrics := character.NewHealthMetrics(ActionID{OtherID: proto.OtherAction_OtherActionHealingModel})

	// Dummy spell for healing callback
	healingModelSpell := character.RegisterSpell(SpellConfig{
		ActionID: ActionID{OtherID: proto.OtherAction_OtherActionHealingModel},
	})

	character.RegisterResetEffect(func(sim *Simulation) {

		// Initialize randomized cadence model
		timeToNextHeal := DurationFromSeconds(0.0)
		healPerTick := 0.0
		pa := &PendingAction{
			NextActionAt: timeToNextHeal,
		}

		pa.OnAction = func(sim *Simulation) {
			// Use modeled HPS to scale heal per tick based on random cadence
			healPerTick = healingModel.Hps * (float64(timeToNextHeal) / float64(time.Second))
			totalHeal := healPerTick * character.PseudoStats.HealingTakenMultiplier
			// Execute the heal
			character.GainHealth(sim, totalHeal, healthMetrics)

			// Callback that can be used by tank specs
			result := healingModelSpell.NewResult(&character.Unit)
			result.Damage = totalHeal
			character.OnHealTaken(sim, healingModelSpell, result)
			healingModelSpell.DisposeResult(result)

			// Random roll for time to next heal. In the case where CadenceVariation exceeds CadenceSeconds, then
			// CadenceSeconds is treated as the median, with two separate uniform distributions to the left and right
			// of it.
			signRoll := sim.RandomFloat("Healing Cadence Variation Sign")
			magnitudeRoll := sim.RandomFloat("Healing Cadence Variation Magnitude")

			if signRoll < 0.5 {
				timeToNextHeal = DurationFromSeconds(minCadence + magnitudeRoll*cadenceVariationLow)
			} else {
				timeToNextHeal = DurationFromSeconds(medianCadence + magnitudeRoll*healingModel.CadenceVariation)
			}

			// Refresh action
			pa.NextActionAt = sim.CurrentTime + timeToNextHeal
			sim.AddPendingAction(pa)
		}

		sim.AddPendingAction(pa)
	})
}

func (character *Character) GetPresimOptions(playerConfig *proto.Player) *PresimOptions {
	healingModel := playerConfig.HealingModel
	if healingModel == nil || healingModel.Hps != 0 || healingModel.CadenceSeconds == 0 {
		// If Hps is not 0, then we don't need to run the presim.
		// Tank sims should always have nonzero Cadence set, even if disabled
		return nil
	}
	return &PresimOptions{
		SetPresimPlayerOptions: func(player *proto.Player) {
			player.HealingModel = nil
		},
		OnPresimResult: func(presimResult *proto.UnitMetrics, iterations int32, duration time.Duration) bool {
			character.applyHealingModel(&proto.HealingModel{
				Hps:            presimResult.Dtps.Avg * 1.50,
				CadenceSeconds: healingModel.CadenceSeconds,
			})
			return true
		},
	}
}
