package shaman

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

// Mana Tide Totem is a trained spell on the server (not a talent). Top rank the shaman knows:
// rank 1 (40) 50 mana/sec, rank 2 (48) 80, rank 3 (58) 100, for 15 sec, 10 min cooldown (Spell.csv).
func (shaman *Shaman) registerManaTideTotemCD() {
	if shaman.Level < 40 {
		return
	}
	rank := 1
	if shaman.Level >= 58 {
		rank = 3
	} else if shaman.Level >= 48 {
		rank = 2
	}
	spellID := []int32{0, 16190, 17354, 17359}[rank]
	manaCost := []float64{0, 50, 200, 400}[rank]
	manaPerTick := []float64{0, 50, 80, 100}[rank]

	actionID := core.ActionID{SpellID: spellID}
	metrics := make([]*core.ResourceMetrics, len(shaman.Party.Players))
	for i, player := range shaman.Party.Players {
		if char := player.GetCharacter(); char.HasManaBar() {
			metrics[i] = char.NewManaMetrics(actionID)
		}
	}

	mttAura := shaman.RegisterAura(core.Aura{
		Label:    "Mana Tide Totem (Shaman)",
		ActionID: actionID,
		Duration: time.Second * 15,
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			core.StartPeriodicAction(sim, core.PeriodicActionOptions{
				Period:   time.Second,
				NumTicks: 15,
				OnAction: func(sim *core.Simulation) {
					for i, player := range shaman.Party.Players {
						if metrics[i] != nil {
							player.GetCharacter().AddMana(sim, manaPerTick, metrics[i])
						}
					}
				},
			})
		},
	})

	mttSpell := shaman.RegisterSpell(core.SpellConfig{
		ActionID: actionID,
		Flags:    SpellFlagTotem | core.SpellFlagAPL | core.SpellFlagNoOnCastComplete,

		ManaCost: core.ManaCostOptions{
			FlatCost:   manaCost,
			Multiplier: shaman.totemManaMultiplier(),
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			IgnoreHaste: true,
			CD: core.Cooldown{
				Timer:    shaman.NewTimer(),
				Duration: time.Minute * 10,
			},
		},
		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			// Replaces the current water totem for its duration.
			shaman.TotemExpirations[WaterTotem] = sim.CurrentTime + time.Second*15
			mttAura.Activate(sim)
		},
	})

	shaman.AddMajorCooldown(core.MajorCooldown{
		Spell: mttSpell,
		Type:  core.CooldownTypeMana,
		ShouldActivate: func(sim *core.Simulation, character *core.Character) bool {
			// Use it once the full 15 x mana per tick fits under max mana.
			return character.MaxMana()-character.CurrentMana() >= 15*manaPerTick
		},
	})
}
