package druid

// Talents from the server's custom trees (text from ui/core/talents/trees/druid.json).
func (druid *Druid) applyServerTalents() {
	// Subtlety: threat -10% per rank.
	druid.PseudoStats.ThreatMultiplier *= 1 - 0.1*float64(druid.Talents.Subtlety)

	// Swiftbloom: casting speed +20% (the 0.5 sec GCD reduction is not modeled).
	if druid.Talents.Swiftbloom {
		druid.MultiplyCastSpeed(1.2)
	}
}
