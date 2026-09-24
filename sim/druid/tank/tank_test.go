package tank

import (
	"testing"

	_ "github.com/wowsims/classic/sim/common"
	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
)

func init() {
	RegisterFeralTankDruid()
}

func TestFeralTank(t *testing.T) {
	core.RunTestSuite(t, t.Name(), core.FullCharacterTestSuiteGenerator([]core.CharacterSuiteConfig{
		{
			Class: proto.Class_ClassDruid,
			Phase: 1,
			Race:  proto.Race_RaceTauren,

			Talents:     StandardTalents,
			GearSet:     core.GetGearSet("../../../ui/feral_tank_druid/gear_sets", "placeholder"),
			Rotation:    core.GetAplRotation("../../../ui/feral_tank_druid/apls", "default"),
			Buffs:       core.FullBuffs,
			Consumes:    FullConsumes,
			SpecOptions: core.SpecOptionsCombo{Label: "Default", SpecOptions: PlayerOptionsDefault},

			IsTank:          true,
			InFrontOfTarget: true,

			ItemFilter: core.ItemFilter{
				WeaponTypes: []proto.WeaponType{
					proto.WeaponType_WeaponTypeDagger,
					proto.WeaponType_WeaponTypeMace,
					proto.WeaponType_WeaponTypeOffHand,
					proto.WeaponType_WeaponTypeStaff,
				},
				ArmorType: proto.ArmorType_ArmorTypeLeather,
				RangedWeaponTypes: []proto.RangedWeaponType{
					proto.RangedWeaponType_RangedWeaponTypeIdol,
				},
			},
			EPReferenceStat: proto.Stat_StatAttackPower,
			StatsToWeigh: []proto.Stat{
				proto.Stat_StatStrength,
				proto.Stat_StatAgility,
				proto.Stat_StatStamina,
				proto.Stat_StatAttackPower,
				proto.Stat_StatArmor,
				proto.Stat_StatDefense,
			},
		},
	}))
}

var StandardTalents = ""

var PlayerOptionsDefault = &proto.Player_FeralTankDruid{
	FeralTankDruid: &proto.FeralTankDruid{
		Options: &proto.FeralTankDruid_Options{
			InnervateTarget: &proto.UnitReference{}, // no Innervate
			StartingRage:    20,
		},
	},
}

var FullConsumes = core.ConsumesCombo{
	Label: "Consumes",
	Consumes: &proto.Consumes{
		Flask:         proto.Flask_FlaskOfTheTitans,
		Food:          proto.Food_FoodSmokedDesertDumpling,
		DefaultPotion: proto.Potions_MightyRagePotion,
	},
}
