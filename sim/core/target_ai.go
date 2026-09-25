package core

import (
	"log"

	"github.com/wowsims/classic/sim/core/proto"
)

type TargetAI interface {
	Initialize(*Target, *proto.Target)
	Reset(*Simulation)
	ExecuteCustomRotation(*Simulation)
}

func (target *Target) initialize(config *proto.Target) {
	if config == nil {
		return
	}

	if target.CurrentTarget != nil {
		if config.SwingSpeed > 0 {
			aaOptions := AutoAttackOptions{
				MainHand: Weapon{
					BaseDamageMin: config.MinBaseDamage,
					SwingSpeed:    config.SwingSpeed,
					SpellSchool:   SpellSchoolFromProto(config.SpellSchool),
				},
				AutoSwingMelee: true,
			}
			if config.DualWield {
				aaOptions.OffHand = aaOptions.MainHand
				if !config.DualWieldPenalty {
					target.PseudoStats.DisableDWMissPenalty = true
				}
			}
			target.EnableAutoAttacks(target, aaOptions)
		}
		if config.SpellDamageInterval > 0 && config.SpellDamageMin > 0 {
			target.registerPeriodicSpellDamage(config)
		}
	}

	if target.AI != nil {
		target.AI.Initialize(target, config)

		target.gcdAction = &PendingAction{
			Priority: ActionPriorityGCD,
			OnAction: func(sim *Simulation) {
				target.Rotation.DoNextAction(sim)
			},
		}
	}
}

// Empty Agent interface functions.
func (target *Target) AddRaidBuffs(_ *proto.RaidBuffs)   {}
func (target *Target) AddPartyBuffs(_ *proto.PartyBuffs) {}
func (target *Target) ApplyTalents()                     {}
func (target *Target) GetCharacter() *Character          { return nil }
func (target *Target) Initialize()                       {}

func (target *Target) ExecuteCustomRotation(sim *Simulation) {
	if target.AI != nil {
		target.AI.ExecuteCustomRotation(sim)
	}
}

type AIFactory func() TargetAI

type PresetTarget struct {
	// String in folder-structure format identifying a category for this unit, e.g. "Black Temple/Bosses".
	PathPrefix string

	Config *proto.Target

	AI AIFactory
}

func (pt PresetTarget) Path() string {
	return pt.PathPrefix + "/" + pt.Config.Name
}
func (pt PresetTarget) ToProto() *proto.PresetTarget {
	// CHECKME might need cloning
	return &proto.PresetTarget{
		Path:   pt.Path(),
		Target: pt.Config,
	}
}

var presetTargets []*PresetTarget
var PresetEncounters []*proto.PresetEncounter

func AddPresetTarget(newPreset *PresetTarget) {
	for _, preset := range presetTargets {
		if preset.Path() == newPreset.Path() {
			log.Fatalf("Preset Target with path %s already added!", newPreset.Path())
		}
	}
	presetTargets = append(presetTargets, newPreset)
}

func GetPresetTargetWithPath(path string) *PresetTarget {
	for _, preset := range presetTargets {
		if preset.Path() == path {
			return preset
		}
	}
	return nil
}

func GetPresetTargetWithID(id int32) *PresetTarget {
	for _, preset := range presetTargets {
		if preset.Config.Id == id {
			return preset
		}
	}
	return nil
}

func AddPresetEncounter(name string, targetPaths []string) {
	if len(targetPaths) == 0 {
		log.Fatalf("Encounter must have targets!")
	}

	var path string
	targetProtos := make([]*proto.PresetTarget, len(targetPaths))

	for i, targetPath := range targetPaths {
		presetTarget := GetPresetTargetWithPath(targetPath)
		if presetTarget == nil {
			log.Fatalf("No preset target with path: %s", targetPath)
		}
		targetProtos[i] = presetTarget.ToProto()

		if i == 0 {
			path = presetTarget.PathPrefix + "/" + name
		}
	}

	for _, preset := range PresetEncounters {
		if preset.Path == path {
			log.Fatalf("Preset Encounter with path %s already added!", path)
		}
	}

	PresetEncounters = append(PresetEncounters, &proto.PresetEncounter{
		Path:    path,
		Targets: targetProtos,
	})
}

// Spell ids used only for the name and icon of the periodic boss spell, by school.
var bossSpellIDs = map[proto.SpellSchool]int32{
	proto.SpellSchool_SpellSchoolArcane: 25345, // Arcane Missile
	proto.SpellSchool_SpellSchoolFire:   25306, // Fireball
	proto.SpellSchool_SpellSchoolFrost:  25304, // Frostbolt
	proto.SpellSchool_SpellSchoolHoly:   10934, // Smite
	proto.SpellSchool_SpellSchoolNature: 15208, // Lightning Bolt
	proto.SpellSchool_SpellSchoolShadow: 25307, // Shadow Bolt
}

// registerPeriodicSpellDamage makes the target cast a damaging spell at its current target every
// SpellDamageInterval seconds (target options in the encounter settings). It lets effects that react to spell damage
// taken (resistances, Eye for an Eye, Shield of Faith, ...) matter in a tank sim.
func (target *Target) registerPeriodicSpellDamage(config *proto.Target) {
	school := config.SpellDamageSchool
	if school == proto.SpellSchool_SpellSchoolPhysical {
		school = proto.SpellSchool_SpellSchoolShadow
	}
	minDamage := config.SpellDamageMin
	maxDamage := minDamage * (1 + config.SpellDamageSpread)
	interval := DurationFromSeconds(config.SpellDamageInterval)

	spell := target.RegisterSpell(SpellConfig{
		ActionID:    ActionID{SpellID: bossSpellIDs[school]},
		SpellSchool: SpellSchoolFromProto(school),
		DefenseType: DefenseTypeMagic,
		ProcMask:    ProcMaskSpellDamage,
		Flags:       SpellFlagNoOnCastComplete,

		DamageMultiplier: 1,
		ThreatMultiplier: 1,

		ApplyEffects: func(sim *Simulation, unit *Unit, spell *Spell) {
			spell.CalcAndDealDamage(sim, unit, sim.Roll(minDamage, maxDamage), spell.OutcomeMagicHit)
		},
	})

	target.RegisterResetEffect(func(sim *Simulation) {
		StartPeriodicAction(sim, PeriodicActionOptions{
			Period: interval,
			OnAction: func(sim *Simulation) {
				if target.CurrentTarget != nil {
					spell.Cast(sim, target.CurrentTarget)
				}
			},
		})
	})
}

