package paladin

import (
	"strconv"
	"time"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/stats"
)

// Seal of Light and Seal of Wisdom (server Spell.csv). Unlike classic, both seals have a passive effect while active:
// Seal of Light increases healing done (20165, 20347-20349: +20/40/64/102), Seal of Wisdom lowers the mana cost of all
// spells (20166, 20356, 20357: -5/10/20). Divine Grace increases both by 10% per rank. The melee procs (Seal of Light
// heals the paladin, Seal of Wisdom restores mana) are not modeled: their proc rate is not in the spell data.
// Judging the seals applies Judgement of Light / Judgement of Wisdom (sim/core/debuffs.go).
func (paladin *Paladin) divineGrace() float64 {
	return 1 + 0.1*float64(paladin.Talents.DivineGrace)
}

func (paladin *Paladin) registerSealOfLight() {
	ranks := []struct {
		level    int32
		spellID  int32
		manaCost float64
		healing  float64
		judgeID  int32
	}{
		{level: 30, spellID: 20165, manaCost: 110, healing: 20, judgeID: 20185},
		{level: 40, spellID: 20347, manaCost: 140, healing: 40, judgeID: 20344},
		{level: 50, spellID: 20348, manaCost: 180, healing: 64, judgeID: 20345},
		{level: 60, spellID: 20349, manaCost: 210, healing: 102, judgeID: 20346},
	}

	debuffs := paladin.NewEnemyAuraArray(core.JudgementOfLightAura)

	for i, rank := range ranks {
		rank := rank
		if paladin.Level < rank.level {
			break
		}

		judgeSpell := paladin.RegisterSpell(core.SpellConfig{
			ActionID:    core.ActionID{SpellID: rank.judgeID},
			SpellSchool: core.SpellSchoolHoly,
			DefenseType: core.DefenseTypeMagic,
			ProcMask:    core.ProcMaskEmpty,
			Flags:       core.SpellFlagMeleeMetrics,

			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				spell.CalcAndDealOutcome(sim, target, spell.OutcomeAlwaysHit)
				debuffs.Get(target).Activate(sim)
			},
		})

		healing := rank.healing * paladin.divineGrace()
		aura := paladin.RegisterAura(core.Aura{
			Label:    "Seal of Light" + paladin.Label + strconv.Itoa(i+1),
			ActionID: core.ActionID{SpellID: rank.spellID},
			Duration: time.Minute * 2,
		}).AttachStatBuff(stats.HealingPower, healing)

		paladin.aurasSoL = append(paladin.aurasSoL, aura)
		paladin.spellsJoL = append(paladin.spellsJoL, judgeSpell)
		paladin.registerSealSpell(aura, judgeSpell, rank.level, i+1, rank.manaCost)
	}
}

func (paladin *Paladin) registerSealOfWisdom() {
	ranks := []struct {
		level     int32
		spellID   int32
		manaCost  float64
		reduction float64
		judgeID   int32
	}{
		{level: 38, spellID: 20166, manaCost: 135, reduction: 5, judgeID: 20186},
		{level: 48, spellID: 20356, manaCost: 170, reduction: 10, judgeID: 20354},
		{level: 58, spellID: 20357, manaCost: 200, reduction: 20, judgeID: 20355},
	}

	debuffs := paladin.NewEnemyAuraArray(core.JudgementOfWisdomAura)

	for i, rank := range ranks {
		rank := rank
		if paladin.Level < rank.level {
			break
		}

		judgeSpell := paladin.RegisterSpell(core.SpellConfig{
			ActionID:    core.ActionID{SpellID: rank.judgeID},
			SpellSchool: core.SpellSchoolHoly,
			DefenseType: core.DefenseTypeMagic,
			ProcMask:    core.ProcMaskEmpty,
			Flags:       core.SpellFlagMeleeMetrics,

			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				spell.CalcAndDealOutcome(sim, target, spell.OutcomeAlwaysHit)
				debuffs.Get(target).Activate(sim)
			},
		})

		reduction := int32(rank.reduction * paladin.divineGrace())
		changeCosts := func(delta int32) {
			for _, spell := range paladin.Spellbook {
				if spell.Cost != nil && spell.Cost.CostType() == core.CostTypeMana {
					spell.Cost.FlatModifier += delta
				}
			}
		}
		aura := paladin.RegisterAura(core.Aura{
			Label:    "Seal of Wisdom" + paladin.Label + strconv.Itoa(i+1),
			ActionID: core.ActionID{SpellID: rank.spellID},
			Duration: time.Minute * 2,
			OnGain: func(_ *core.Aura, _ *core.Simulation) {
				changeCosts(-reduction)
			},
			OnExpire: func(_ *core.Aura, _ *core.Simulation) {
				changeCosts(reduction)
			},
		})

		paladin.aurasSoW = append(paladin.aurasSoW, aura)
		paladin.spellsJoW = append(paladin.spellsJoW, judgeSpell)
		paladin.registerSealSpell(aura, judgeSpell, rank.level, i+1, rank.manaCost)
	}
}

// registerSealSpell registers the castable seal (APL) that activates the seal aura.
func (paladin *Paladin) registerSealSpell(aura *core.Aura, judgeSpell *core.Spell, level int32, rank int, manaCost float64) {
	paladin.RegisterSpell(core.SpellConfig{
		ActionID:    aura.ActionID,
		SpellSchool: core.SpellSchoolHoly,
		Flags:       core.SpellFlagAPL,

		RequiredLevel: int(level),
		Rank:          rank,

		ManaCost: core.ManaCostOptions{
			FlatCost:   manaCost - paladin.getLibramSealCostReduction(),
			Multiplier: paladin.benediction(),
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			paladin.applySeal(aura, judgeSpell, sim)
		},
	})
}
