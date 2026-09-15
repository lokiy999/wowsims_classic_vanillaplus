package priest

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

func (priest *Priest) registerVampiricEmbraceSpell() {
	if !priest.Talents.VampiricEmbrace {
		return
	}

	actionID := core.ActionID{SpellID: 15286}
	manaCost := 40.0
	duration := time.Minute * 1
	cdTimer := priest.NewTimer()

	partyPlayers := priest.Env.Raid.GetPlayerParty(&priest.Unit).Players
	// DBC: Vampiric Embrace heals for 10% of Shadow damage; Improved Vampiric Embrace adds 10% per rank.
	healthReturnedMultuplier := 0.10 + 0.10*float64(priest.Talents.ImprovedVampiricEmbrace)

	// Routed through a real healing spell (instead of a direct GainHealth call) so it
	// populates TotalHealing and shows up in the HPS metric - including overhealing,
	// same as any other heal (dealHealingInternal records the pre-clamp amount).
	healSpell := priest.RegisterSpell(core.SpellConfig{
		ActionID:    actionID.WithTag(1),
		SpellSchool: core.SpellSchoolShadow,
		ProcMask:    core.ProcMaskSpellHealing,
		Flags:       core.SpellFlagHelpful | core.SpellFlagNoOnCastComplete | core.SpellFlagPassiveSpell,

		DamageMultiplier: 1,
		ThreatMultiplier: 1,
	})

	// DBC: "heals party members for X% of any Shadow spell damage YOU deal" -
	// only the priest who cast Vampiric Embrace, not any Shadow damage source.
	// Must be hooked on BOTH OnSpellHitTaken (direct hits, e.g. Mind Blast) and
	// OnPeriodicDamageTaken (DoT ticks, e.g. Shadow Word: Pain/Devouring
	// Plague/Mind Flay) - dealDamageInternal only calls OnSpellHitTaken for
	// non-periodic damage (sim/core/spell_result.go), so a Shadow Priest's
	// mostly-DoT damage would otherwise never trigger this heal at all.
	onShadowDamageTaken := func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
		if result.Landed() && spell.Unit == &priest.Unit && spell.SpellSchool.Matches(core.SpellSchoolShadow) {
			healthGained := result.Damage * healthReturnedMultuplier
			if healthGained <= 0 {
				return
			}
			for _, player := range partyPlayers {
				healSpell.CalcAndDealHealing(sim, &player.GetCharacter().Unit, healthGained, healSpell.OutcomeHealing)
			}
		}
	}

	priest.VampiricEmbraceAuras = priest.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		return target.GetOrRegisterAura(core.Aura{
			ActionID:              actionID,
			Label:                 "Vampiric Embrace (Health) - " + target.Label,
			Duration:              duration,
			OnSpellHitTaken:       onShadowDamageTaken,
			OnPeriodicDamageTaken: onShadowDamageTaken,
		})
	})

	priest.VampiricEmbrace = priest.RegisterSpell(core.SpellConfig{
		ActionID:    actionID,
		SpellSchool: core.SpellSchoolShadow,
		DefenseType: core.DefenseTypeMagic,
		ProcMask:    core.ProcMaskEmpty,
		Flags:       SpellFlagPriest | core.SpellFlagAPL,

		ManaCost: core.ManaCostOptions{
			FlatCost: manaCost,
		},

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			CD: core.Cooldown{
				Timer:    cdTimer,
				Duration: time.Second * 15,
			},
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			result := spell.CalcOutcome(sim, target, spell.OutcomeMagicHit)
			if result.Landed() {
				priest.VampiricEmbraceAuras.Get(target).Activate(sim)
			}
			spell.DealOutcome(sim, result)
		},
	})
}
