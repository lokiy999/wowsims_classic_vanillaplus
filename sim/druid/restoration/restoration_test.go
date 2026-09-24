package restoration

import (
	"testing"

	_ "github.com/wowsims/classic/sim/common"
	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
)

func init() {
	RegisterRestorationDruid()
}

func TestRestorationDruid(t *testing.T) {
	core.RunTestSuite(t, t.Name(), core.FullCharacterTestSuiteGenerator([]core.CharacterSuiteConfig{
		{
			Class:    proto.Class_ClassDruid,
			Phase:    1,
			Race:     proto.Race_RaceTauren,
			IsHealer: true,

			Talents:     StandardTalents,
			GearSet:     core.GetGearSet("../../../ui/restoration_druid/gear_sets", "placeholder"),
			Rotation:    core.GetAplRotation("../../../ui/restoration_druid/apls", "default"),
			Buffs:       core.FullBuffs,
			Consumes:    FullConsumes,
			SpecOptions: core.SpecOptionsCombo{Label: "Standard", SpecOptions: PlayerOptionsStandard},

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
			EPReferenceStat: proto.Stat_StatSpellPower,
			StatsToWeigh: []proto.Stat{
				proto.Stat_StatIntellect,
				proto.Stat_StatSpirit,
				proto.Stat_StatSpellPower,
				proto.Stat_StatSpellCrit,
				proto.Stat_StatMP5,
			},
		},
	}))
}

var StandardTalents = ""

var PlayerOptionsStandard = &proto.Player_RestorationDruid{
	RestorationDruid: &proto.RestorationDruid{
		Options: &proto.RestorationDruid_Options{
			InnervateTarget: &proto.UnitReference{},
		},
	},
}

var FullConsumes = core.ConsumesCombo{
	Label: "Consumes",
	Consumes: &proto.Consumes{
		Flask:         proto.Flask_FlaskOfDistilledWisdom,
		DefaultPotion: proto.Potions_MajorManaPotion,
	},
}
