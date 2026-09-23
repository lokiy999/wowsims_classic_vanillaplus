package core

import (
	"testing"
	"time"
)

func TestSingleAuraExclusiveDurationNoOverwrite(t *testing.T) {
	sim := &Simulation{}

	target := Unit{
		Type:        EnemyUnit,
		Index:       0,
		Level:       63,
		auraTracker: newAuraTracker(),
	}
	mangle := bleedDamageAura(&target, Aura{Label: "Test Bleed", Duration: time.Minute}, 1.3)

	sim.CurrentTime = 1 * time.Second

	mangle.Activate(sim)
}

func TestSingleAuraExclusiveDurationOverwrite(t *testing.T) {
	sim := &Simulation{}

	target := Unit{
		Type:        EnemyUnit,
		Index:       0,
		Level:       63,
		auraTracker: newAuraTracker(),
	}
	mangle := bleedDamageAura(&target, Aura{Label: "Test Bleed", Duration: time.Minute}, 1.3)

	sim.CurrentTime = 1 * time.Second

	mangle.Activate(sim)
}
