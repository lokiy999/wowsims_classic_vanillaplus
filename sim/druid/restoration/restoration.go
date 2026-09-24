package restoration

import (
	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/druid"
)

func RegisterRestorationDruid() {
	core.RegisterAgentFactory(
		proto.Player_RestorationDruid{},
		proto.Spec_SpecRestorationDruid,
		func(character *core.Character, options *proto.Player) core.Agent {
			return NewRestorationDruid(character, options)
		},
		func(player *proto.Player, spec interface{}) {
			playerSpec, ok := spec.(*proto.Player_RestorationDruid)
			if !ok {
				panic("Invalid spec value for Restoration Druid!")
			}
			player.Spec = playerSpec
		},
	)
}

func NewRestorationDruid(character *core.Character, options *proto.Player) *RestorationDruid {
	restoOptions := options.GetRestorationDruid()
	selfBuffs := druid.SelfBuffs{}

	resto := &RestorationDruid{
		Druid: druid.New(character, druid.Humanoid, selfBuffs, options.TalentsString),
	}

	resto.SelfBuffs.InnervateTarget = &proto.UnitReference{}
	if restoOptions.Options.InnervateTarget != nil {
		resto.SelfBuffs.InnervateTarget = restoOptions.Options.InnervateTarget
	}

	return resto
}

type RestorationDruid struct {
	*druid.Druid
}

func (resto *RestorationDruid) GetDruid() *druid.Druid {
	return resto.Druid
}

// Heals go to the first target dummy (the healing sim's stand-in for a raid member), or to the druid.
func (resto *RestorationDruid) GetMainTarget() *core.Unit {
	target := resto.Env.Raid.GetFirstTargetDummy()
	if target == nil {
		return &resto.Unit
	}
	return &target.Unit
}

func (resto *RestorationDruid) Initialize() {
	resto.CurrentTarget = resto.GetMainTarget()
	resto.Druid.Initialize()
	resto.Druid.RegisterHealingSpells()
}

func (resto *RestorationDruid) Reset(sim *core.Simulation) {
	resto.Druid.Reset(sim)
}
