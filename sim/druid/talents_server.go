package druid

// Talents from the server's custom trees (text from ui/core/talents/trees/druid.json).
func (druid *Druid) applyServerTalents() {
	// Subtlety: threat -10% per rank.
	druid.PseudoStats.ThreatMultiplier *= 1 - 0.1*float64(druid.Talents.Subtlety)

	// Survival Instincts (DBC 33772-33776): chance to be critically hit by melee attacks -1/2/3/4/6%.
	// The health restored on your own critical strikes is not modeled.
	druid.PseudoStats.ReducedCritTakenChance += []float64{0, 0.01, 0.02, 0.03, 0.04, 0.06}[min(druid.Talents.SurvivalInstincts, 5)]

	// Swiftbloom: casting speed +20% (the 0.5 sec GCD reduction is not modeled).
	if druid.Talents.Swiftbloom {
		druid.MultiplyCastSpeed(1.2)
	}
}
