package hunter

import (
	"time"

	"github.com/wowsims/classic/sim/core"
)

func (hunter *Hunter) getMongooseBiteConfig(rank int) core.SpellConfig {
	spellId := [5]int32{0, 1495, 14269, 14270, 14271}[rank]
	baseDamage := [5]float64{0, 10, 15, 25, 35}[rank] // server: weapon damage plus this
	manaCost := [5]float64{0, 60, 90, 110, 150}[rank]
	level := [5]int{0, 16, 30, 44, 58}[rank]

	spellConfig := core.SpellConfig{
		SpellCode:     SpellCode_HunterMongooseBite,
		ActionID:      core.ActionID{SpellID: spellId},
		SpellSchool:   core.SpellSchoolPhysical,
		DefenseType:   core.DefenseTypeMelee,
		ProcMask:      core.ProcMaskMeleeSpecial,
		Flags:         core.SpellFlagMeleeMetrics | core.SpellFlagAPL,
		Rank:          rank,
		RequiredLevel: level,

		ManaCost: core.ManaCostOptions{
			FlatCost: manaCost,
		},

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			CD: core.Cooldown{
				Timer:    hunter.NewTimer(),
				Duration: time.Second * 5,
			},
		},

		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return hunter.DistanceFromTarget <= core.MaxMeleeAttackDistance && hunter.DefensiveState.IsActive()
		},

		BonusCritRating:  float64(hunter.Talents.SavageStrikes) * 3 * core.CritRatingPerCritChance, // DBC: 3%/rank
		CritDamageBonus:  hunter.mortalShots(),
		DamageMultiplier: 1,
		ThreatMultiplier: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			hunter.DefensiveState.Deactivate(sim)
			damage := baseDamage + spell.Unit.MHNormalizedWeaponDamage(sim, spell.MeleeAttackPower())
			spell.CalcAndDealDamage(sim, target, damage, spell.OutcomeMeleeWeaponSpecialHitAndCrit)
		},
	}

	return spellConfig
}

func (hunter *Hunter) registerMongooseBiteSpell() {
	// Aura is only used as a pre-requisite for Mongoose Bite
	hunter.DefensiveState = hunter.RegisterAura(core.Aura{
		Label:    "Defensive State",
		ActionID: core.ActionID{SpellID: 5302},
		Duration: time.Second * 5,

		OnSpellHitTaken: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if result.DidDodge() {
				aura.Activate(sim)
			}
		},
	})

	rank := map[int32]int{
		25: 1,
		40: 2,
		50: 3,
		60: 4,
	}[hunter.Level]

	config := hunter.getMongooseBiteConfig(rank)
	hunter.MongooseBite = hunter.GetOrRegisterSpell(config)
}
