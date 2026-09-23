import { Phase } from '../core/constants/other.js';
import * as PresetUtils from '../core/preset_utils.js';
import {
	AgilityElixir,
	Alcohol,
	ArmorElixir,
	AttackPowerBuff,
	Conjured,
	Consumes,
	Debuffs,
	Explosive,
	FirePowerBuff,
	Flask,
	Food,
	HealthElixir,
	IndividualBuffs,
	Potions,
	Profession,
	RaidBuffs,
	SaygesFortune,
	SpellPowerBuff,
	StrengthBuff,
	TristateEffect,
	WeaponImbue,
	ZanzaBuff,
} from '../core/proto/common.js';
import { Blessings, PaladinAura, PaladinOptions as ProtectionPaladinOptions,PaladinSeal } from '../core/proto/paladin.js';
import { SavedTalents } from '../core/proto/ui.js';
import APLProtJson from './apls/prot.apl.json';
import BlankGear from './gear_sets/blank.gear.json';

// Preset options for this spec.
// Eventually we will import these values for the raid sim too, so its good to
// keep them in a separate file.

///////////////////////////////////////////////////////////////////////////
//                                 Gear Presets
///////////////////////////////////////////////////////////////////////////

export const GearBlank = PresetUtils.makePresetGear('Blank', BlankGear);

export const GearPresets = {};

export const DefaultGear = GearBlank;

///////////////////////////////////////////////////////////////////////////
//                                 APL Presets
///////////////////////////////////////////////////////////////////////////

export const APLProt = PresetUtils.makePresetAPLRotation('Prot', APLProtJson);

export const APLPresets = {
	[Phase.Phase1]: [APLProt],
};

export const DefaultAPL = APLPresets[Phase.Phase1][0];

///////////////////////////////////////////////////////////////////////////
//                                 Talent presets
///////////////////////////////////////////////////////////////////////////

// Default talents. Uses the wowhead calculator format, make the talents on
// https://wowhead.com/classic/talent-calc and copy the numbers in the url.

export const ProtTalents = {
	name: 'Prot',
	data: SavedTalents.create({
		talentsString: '',
	}),
};

export const TalentPresets = {
	[Phase.Phase1]: [ProtTalents],
};

export const DefaultTalents = TalentPresets[Phase.Phase1][0];

///////////////////////////////////////////////////////////////////////////
//                                 Options
///////////////////////////////////////////////////////////////////////////

export const DefaultOptions = ProtectionPaladinOptions.create({
	aura: PaladinAura.SanctityAura,
	primarySeal: PaladinSeal.Righteousness,
	personalBlessing: Blessings.BlessingOfSanctuary,
	righteousFury: true,
});

export const DefaultConsumes = Consumes.create({
	agilityElixir: AgilityElixir.ElixirOfTheMongoose,
	healthElixir: HealthElixir.ElixirOfFortitude,
	armorElixir: ArmorElixir.ElixirOfSuperiorDefense,
	defaultPotion: Potions.GreaterStoneshieldPotion,
	dragonBreathChili: true,
	food: Food.FoodTenderWolfSteak,
	flask: Flask.FlaskOfTheTitans,
	firePowerBuff: FirePowerBuff.ElixirOfGreaterFirepower,
	fillerExplosive: Explosive.ExplosiveDenseDynamite,
	//mainHandImbue: WeaponImbue.WildStrikes,
	//offHandImbue: WeaponImbue.MagnificentTrollshine,

	spellPowerBuff: SpellPowerBuff.GreaterArcaneElixir,
	strengthBuff: StrengthBuff.JujuPower,
	zanzaBuff: ZanzaBuff.ROIDS,
	attackPowerBuff: AttackPowerBuff.JujuMight,
	defaultConjured: Conjured.ConjuredDemonicRune,
	alcohol: Alcohol.AlcoholRumseyRumBlackLabel,
});

export const DefaultIndividualBuffs = IndividualBuffs.create({
	blessingOfWisdom: TristateEffect.TristateEffectImproved,
	fengusFerocity: true,
	moldarsMoxie: true,
	rallyingCryOfTheDragonslayer: true,
	saygesFortune: SaygesFortune.SaygesDamage,
	slipkiksSavvy: true,
	songflowerSerenade: true,
	spiritOfZandalar: true,
});

export const DefaultRaidBuffs = RaidBuffs.create({
	powerWordFortitude: TristateEffect.TristateEffectImproved,
	arcaneBrilliance: true,
	battleShout: TristateEffect.TristateEffectImproved,
	divineSpirit: true,
	giftOfTheWild: TristateEffect.TristateEffectImproved,
	sanctityAura: true,
});

export const DefaultDebuffs = Debuffs.create({
	curseOfRecklessness: true,
	faerieFire: true,
	giftOfArthas: true,
	exposeArmor: TristateEffect.TristateEffectImproved,
	judgementOfWisdom: true,
	judgementOfTheCrusader: TristateEffect.TristateEffectImproved,
	improvedScorch: true,
});

export const OtherDefaults = {
	distanceFromTarget: 5, // Max melee range
	profession1: Profession.Blacksmithing,
	profession2: Profession.Engineering,
};
