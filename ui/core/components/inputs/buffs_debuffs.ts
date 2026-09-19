import { BlessingOfKingsType, Class, Faction, SaygesFortune, Stat, TristateEffect } from '../../proto/common';
import { ActionId } from '../../proto_utils/action_id';
import { IconEnumPicker } from '../icon_enum_picker';
import {
	makeBooleanDebuffInput,
	makeBooleanIndividualBuffInput,
	makeBooleanRaidBuffInput,
	makeEnumIndividualBuffInput,
	makeEnumRaidBuffInput,
	makeMultistateIndividualBuffInput,
	makeMultistateRaidBuffInput,
	makeTristateDebuffInput,
	makeTristateIndividualBuffInput,
	makeTristateRaidBuffInput,
	withLabel,
	withOptionLabels,
} from '../icon_inputs';
import { IconPicker, IconPickerDirection } from '../icon_picker';
import * as InputHelpers from '../input_helpers';
import { MultiIconPicker } from '../multi_icon_picker';
import { ItemStatOption, PickerStatOptions } from './stat_options';

///////////////////////////////////////////////////////////////////////////
//                                 RAID BUFFS
///////////////////////////////////////////////////////////////////////////

export const AllStatsBuff = withLabel(
	makeTristateRaidBuffInput({
		actionId: () => ActionId.fromSpellId(9885),
		impId: ActionId.fromSpellId(17055),
		fieldName: 'giftOfTheWild',
	}),
	'Mark of the Wild',
);

// Enum picker so the extra stat variants (ZG set bonus / talented, which stack
// with each other) can be selected from one icon instead of needing separate
// toggles.
export const BlessingOfKings = withLabel(
	withOptionLabels(
		makeEnumIndividualBuffInput({
			fieldName: 'blessingOfKingsType',
			direction: IconPickerDirection.Horizontal,
			values: [
				{
					// No actionId/iconUrl, so this renders as a flat grey square
					// (like the off state of the Intellect/Spirit/Stamina pickers)
					// instead of a desaturated copy of the Kings icon.
					color: 'grey',
					value: BlessingOfKingsType.BlessingOfKingsNone,
					tooltip: 'Disabled',
				},
				{
					actionId: () => ActionId.fromSpellId(20217),
					value: BlessingOfKingsType.BlessingOfKingsNormal,
					text: '(10%)',
					tooltip: 'Blessing of Kings (10%)',
				},
				{
					actionId: () => ActionId.fromSpellId(20217),
					value: BlessingOfKingsType.BlessingOfKingsZgSet,
					text: '(11%)',
					tooltip: "Blessing of Kings + Zandalar Vindicator's Regalia 3pc (11%)",
				},
				{
					actionId: () => ActionId.fromSpellId(20217),
					value: BlessingOfKingsType.BlessingOfKingsTalented,
					text: '(12%)',
					tooltip: 'Blessing of Kings + Improved Blessing of Kings talent (12%)',
				},
				{
					actionId: () => ActionId.fromSpellId(20217),
					value: BlessingOfKingsType.BlessingOfKingsZgSetTalented,
					text: '(13%)',
					tooltip: "Blessing of Kings + Zandalar Vindicator's Regalia 3pc + Improved Blessing of Kings talent (13%)",
				},
			],
		}),
	),
	'Blessing of Kings',
);

export const ArmorBuff = withLabel(
	makeTristateRaidBuffInput({
		actionId: () => ActionId.fromSpellId(10293),
		impId: ActionId.fromSpellId(20142),
		showWhen: player => player.getFaction() === Faction.Alliance,
		fieldName: 'devotionAura',
	}),
	'Devotion Aura',
);

export const PhysDamReductionBuff = withLabel(
	makeTristateRaidBuffInput({
		actionId: () => ActionId.fromSpellId(10408),
		impId: ActionId.fromSpellId(16293),
		showWhen: player => player.getFaction() === Faction.Horde,
		fieldName: 'stoneskinTotem',
	}),
	'Stoneskin',
);

//export const DamageReductionPercentBuff = withLabel(
//	makeBooleanIndividualBuffInput({
//		actionId: player =>
//			player.getMatchingSpellActionId([
//				{ id: 20911, minLevel: 30, maxLevel: 39 },
//				{ id: 20912, minLevel: 40, maxLevel: 49 },
//				{ id: 20913, minLevel: 50, maxLevel: 59 },
//				{ id: 20914, minLevel: 60 },
//			]),
//		showWhen: player => player.getFaction() === Faction.Alliance,
//		fieldName: 'blessingOfSanctuary',
//	}),
//	'Blessing of Sanctuary',
//);

export const ResistanceBuff = InputHelpers.makeMultiIconInput({
	values: [
		// Shadow
		// Priests always have their own Shadow Protection (sim/priest/priest.go's
		// AddRaidBuffs forces raidBuffs.shadowProtection = true unconditionally).
		makeBooleanRaidBuffInput({
			actionId: () => ActionId.fromSpellId(10958),
			fieldName: 'shadowProtection',
			showWhen: player => player.getClass() !== Class.ClassPriest,
		}),
		makeBooleanRaidBuffInput({
			actionId: () => ActionId.fromSpellId(19896),
			fieldName: 'shadowResistanceAura',
		}),
		// Nature
		makeBooleanRaidBuffInput({
			actionId: () => ActionId.fromSpellId(10601),
			fieldName: 'natureResistanceTotem',
			showWhen: player => player.getFaction() === Faction.Horde,
		}),
		makeBooleanRaidBuffInput({
			actionId: () => ActionId.fromSpellId(20190),
			fieldName: 'aspectOfTheWild',
		}),
		// Fire
		makeBooleanRaidBuffInput({
			actionId: () => ActionId.fromSpellId(19900),
			fieldName: 'fireResistanceAura',
			showWhen: player => player.getFaction() === Faction.Alliance,
		}),
		makeBooleanRaidBuffInput({
			actionId: () => ActionId.fromSpellId(10538),
			fieldName: 'fireResistanceTotem',
			showWhen: player => player.getFaction() === Faction.Horde,
		}),
		// Frost
		makeBooleanRaidBuffInput({
			actionId: () => ActionId.fromSpellId(19898),
			fieldName: 'frostResistanceAura',
			showWhen: player => player.getFaction() === Faction.Alliance,
		}),
		makeBooleanRaidBuffInput({
			actionId: () => ActionId.fromSpellId(10479),
			fieldName: 'frostResistanceTotem',
			showWhen: player => player.getFaction() === Faction.Horde,
		}),
	],
	label: 'Resistances',
});

// A Priest's own Power Word: Fortitude is only forced to at least Regular
// (sim/priest/priest.go's AddRaidBuffs: max(uiValue, ownTalentValue)) — Improved
// requires the talent. So unlike Divine Spirit/Shadow Protection below, this toggle
// stays visible and selectable even for a Priest: setting it to Improved still means
// something (another priest in the raid providing the improved version) if their own
// talent doesn't reach it.
export const StaminaBuff = withLabel(
	makeTristateRaidBuffInput({
		actionId: () => ActionId.fromSpellId(10938),
		impId: ActionId.fromSpellId(14767),
		fieldName: 'powerWordFortitude',
	}),
	'Stamina',
);

export const BloodPactBuff = withLabel(
	makeTristateRaidBuffInput({
		actionId: () => ActionId.fromSpellId(11767),
		impId: ActionId.fromSpellId(18696),
		fieldName: 'bloodPact',
	}),
	'Blood Pact',
);

export const BlessingOfMight = withLabel(
	makeTristateIndividualBuffInput({
		actionId: () => ActionId.fromSpellId(25291),
		impId: ActionId.fromSpellId(20048),
		fieldName: 'blessingOfMight',
		showWhen: player => player.getFaction() === Faction.Alliance,
	}),
	'Blessing of Might',
);

export const StrengthBuffHorde = withLabel(
	makeTristateRaidBuffInput({
		actionId: () => ActionId.fromSpellId(25361),
		impId: ActionId.fromSpellId(16295),
		fieldName: 'strengthOfEarthTotem',
		showWhen: player => player.getFaction() === Faction.Horde,
	}),
	'Strength',
);

export const GraceOfAir = withLabel(
	makeTristateRaidBuffInput({
		actionId: () => ActionId.fromSpellId(25359),
		impId: ActionId.fromSpellId(16295),
		fieldName: 'graceOfAirTotem',
		showWhen: player => player.getFaction() === Faction.Horde,
	}),
	'Agility',
);

export const IntellectBuff = withLabel(
	makeBooleanRaidBuffInput({
		actionId: () => ActionId.fromSpellId(10157),
		fieldName: 'arcaneBrilliance',
	}),
	'Intellect',
);

// Priests always have their own Divine Spirit (sim/priest/priest.go's AddRaidBuffs
// forces raidBuffs.divineSpirit = true unconditionally) — same reasoning as
// StaminaBuff above.
export const SpiritBuff = withLabel(
	makeBooleanRaidBuffInput({
		actionId: () => ActionId.fromSpellId(27841),
		fieldName: 'divineSpirit',
		showWhen: player => player.getClass() !== Class.ClassPriest,
	}),
	'Spirit',
);

// Scrolls no longer share a cell with their matching raid buff (Stamina/Intellect/Spirit
// above) — on this server they stack with them instead of being a fallback for when the
// raid buff is missing (see sim/core/buffs.go), so they're rendered in the Consumables
// tab's Scrolls row (consumes_picker.ts) instead of here. Each is now a tristate
// (none/rank IV/rank V) since the rank V scrolls (81013/81012/81014) don't stack with
// their own rank IV, same reasoning as the Agility/Strength/Protection scrolls' V option.
// Rendered as a dropdown (not a click-to-cycle icon) so rank IV and rank V -
// which have different icons - are both visible and directly selectable
// instead of only distinguishable by a small corner badge.
export const ScrollOfStamina = makeEnumRaidBuffInput({
	fieldName: 'scrollOfStamina',
	values: [
		{ value: TristateEffect.TristateEffectMissing },
		{ value: TristateEffect.TristateEffectRegular, actionId: () => ActionId.fromItemId(10307) },
		{ value: TristateEffect.TristateEffectImproved, actionId: () => ActionId.fromItemId(81013) },
	],
});
export const ScrollOfIntellect = makeEnumRaidBuffInput({
	fieldName: 'scrollOfIntellect',
	values: [
		{ value: TristateEffect.TristateEffectMissing },
		{ value: TristateEffect.TristateEffectRegular, actionId: () => ActionId.fromItemId(10308) },
		{ value: TristateEffect.TristateEffectImproved, actionId: () => ActionId.fromItemId(81012) },
	],
});
export const ScrollOfSpirit = makeEnumRaidBuffInput({
	fieldName: 'scrollOfSpirit',
	values: [
		{ value: TristateEffect.TristateEffectMissing },
		{ value: TristateEffect.TristateEffectRegular, actionId: () => ActionId.fromItemId(10306) },
		{ value: TristateEffect.TristateEffectImproved, actionId: () => ActionId.fromItemId(81014) },
	],
});

export const BattleShoutBuff = withLabel(
	makeTristateRaidBuffInput({
		actionId: () => ActionId.fromSpellId(25289),
		impId: ActionId.fromSpellId(12861),
		fieldName: 'battleShout',
	}),
	'Battle Shout',
);

export const TrueshotAuraBuff = withLabel(
	makeBooleanRaidBuffInput({
		actionId: () => ActionId.fromSpellId(20906),
		fieldName: 'trueshotAura',
	}),
	'Trueshot Aura',
);

export const BlessingOfWisdom = withLabel(
	makeTristateIndividualBuffInput({
		actionId: () => ActionId.fromSpellId(25290),
		impId: ActionId.fromSpellId(20245),
		fieldName: 'blessingOfWisdom',
	}),
	'Blessing of Wisdom',
);
export const ManaSpringTotem = withLabel(
	makeTristateRaidBuffInput({
		actionId: () => ActionId.fromSpellId(10497),
		impId: ActionId.fromSpellId(16208),
		fieldName: 'manaSpringTotem',
	}),
	'Mana Spring Totem',
);

export const MeleeCritBuff = withLabel(
	makeBooleanRaidBuffInput({ actionId: () => ActionId.fromSpellId(24932), fieldName: 'leaderOfThePack' }),
	'Leader of the Pack',
);

export const SpellCritBuff = withLabel(makeBooleanRaidBuffInput({ actionId: () => ActionId.fromSpellId(24907), fieldName: 'moonkinAura' }), 'Moonkin Aura');

// Misc Buffs
export const RetributionAura = makeTristateRaidBuffInput({
	actionId: () => ActionId.fromSpellId(10301),
	impId: ActionId.fromSpellId(20092),
	fieldName: 'retributionAura',
	showWhen: player => player.getFaction() === Faction.Alliance,
});

export const SanctityAura = makeBooleanRaidBuffInput({
	actionId: () => ActionId.fromSpellId(20218),
	fieldName: 'sanctityAura',
	showWhen: player => player.getFaction() === Faction.Alliance,
});

export const Thorns = makeTristateRaidBuffInput({
	actionId: () => ActionId.fromSpellId(9910),
	impId: ActionId.fromSpellId(16840),
	fieldName: 'thorns',
});

export const Innervate = makeMultistateIndividualBuffInput({
	actionId: () => ActionId.fromSpellId(29166),
	numStates: 11,
	fieldName: 'innervates',
});

export const PowerInfusion = makeMultistateIndividualBuffInput({
	actionId: () => ActionId.fromSpellId(10060),
	numStates: 11,
	fieldName: 'powerInfusions',
});

export const BattleSquawkBuff = makeMultistateRaidBuffInput({
	actionId: () => ActionId.fromSpellId(23060),
	numStates: 6,
	fieldName: 'battleSquawk',
});

///////////////////////////////////////////////////////////////////////////
//                                 WORLD BUFFS
///////////////////////////////////////////////////////////////////////////

export const RallyingCryOfTheDragonslayer = withLabel(
	makeBooleanIndividualBuffInput({
		actionId: () => ActionId.fromSpellId(22888),
		fieldName: 'rallyingCryOfTheDragonslayer',
	}),
	'Rallying Cry Of The Dragonslayer',
);
export const SpiritOfZandalar = withLabel(
	makeBooleanIndividualBuffInput({
		actionId: () => ActionId.fromSpellId(24425),
		fieldName: 'spiritOfZandalar',
	}),
	'Spirit of Zandalar',
);
export const SongflowerSerenade = withLabel(
	makeBooleanIndividualBuffInput({
		actionId: () => ActionId.fromSpellId(15366),
		fieldName: 'songflowerSerenade',
	}),
	'Songflower Serenade',
);
export const WarchiefsBlessing = withLabel(
	makeBooleanIndividualBuffInput({
		actionId: () => ActionId.fromSpellId(16609),
		fieldName: 'warchiefsBlessing',
		// showWhen: player => player.getFaction() === Faction.Horde,
	}),
	`Warchief's Blessing`,
);

export const SaygesDarkFortune = (inputs: ItemStatOption<SaygesFortune>[]) =>
	makeEnumIndividualBuffInput({
		direction: IconPickerDirection.Horizontal,
		values: [
			{ iconUrl: 'https://wow.zamimg.com/images/wow/icons/large/inv_misc_orb_02.jpg', value: SaygesFortune.SaygesUnknown, text: `Sayge's Dark Fortune` },
			...inputs.map(input => input.config),
		],
		fieldName: 'saygesFortune',
	});

export const SaygesDamage = { actionId: () => ActionId.fromSpellId(23768), value: SaygesFortune.SaygesDamage, text: `Sayge's Damage` };
export const SaygesAgility = { actionId: () => ActionId.fromSpellId(23736), value: SaygesFortune.SaygesAgility, text: `Sayge's Agility` };
export const SaygesIntellect = { actionId: () => ActionId.fromSpellId(23766), value: SaygesFortune.SaygesIntellect, text: `Sayge's Intellect` };
export const SaygesSpirit = { actionId: () => ActionId.fromSpellId(23738), value: SaygesFortune.SaygesSpirit, text: `Sayge's Spirit` };
export const SaygesStamina = { actionId: () => ActionId.fromSpellId(23737), value: SaygesFortune.SaygesStamina, text: `Sayge's Stamina` };

// Dire Maul Buffs
export const FengusFerocity = withLabel(
	makeBooleanIndividualBuffInput({
		actionId: () => ActionId.fromSpellId(22817),
		fieldName: 'fengusFerocity',
	}),
	`Fengus' Ferocity`,
);
export const MoldarsMoxie = withLabel(
	makeBooleanIndividualBuffInput({
		actionId: () => ActionId.fromSpellId(22818),
		fieldName: 'moldarsMoxie',
	}),
	`Moldar's Moxie`,
);
export const SlipKiksSavvy = withLabel(
	makeBooleanIndividualBuffInput({
		actionId: () => ActionId.fromSpellId(22820),
		fieldName: 'slipkiksSavvy',
	}),
	`Slip'kik's Savvy`,
);

///////////////////////////////////////////////////////////////////////////
//                                 DEBUFFS
///////////////////////////////////////////////////////////////////////////

export const MajorArmorDebuff = InputHelpers.makeMultiIconInput({
	values: [
		makeBooleanDebuffInput({
			actionId: () => ActionId.fromSpellId(11597),
			fieldName: 'sunderArmor',
		}),
		makeTristateDebuffInput({
			actionId: () => ActionId.fromSpellId(11198),
			impId: ActionId.fromSpellId(14169),
			fieldName: 'exposeArmor',
		}),
	],
	label: 'Major Armor Penetration',
});

export const CurseOfRecklessness = withLabel(
	makeBooleanDebuffInput({
		actionId: () => ActionId.fromSpellId(11717),
		fieldName: 'curseOfRecklessness',
	}),
	'Curse of Recklessness',
);

export const FaerieFire = withLabel(
	makeBooleanDebuffInput({
		actionId: () => ActionId.fromSpellId(9907),
		fieldName: 'faerieFire',
	}),
	'Faerie Fire',
);

export const curseOfWeaknessDebuff = withLabel(
	makeTristateDebuffInput({
		actionId: () => ActionId.fromSpellId(11708),
		impId: ActionId.fromSpellId(18181),
		fieldName: 'curseOfWeakness',
	}),
	'Curse of Weakness',
);

export const AttackPowerDebuff = InputHelpers.makeMultiIconInput({
	values: [
		makeTristateDebuffInput({
			actionId: () => ActionId.fromSpellId(11556),
			impId: ActionId.fromSpellId(12879),
			fieldName: 'demoralizingShout',
		}),
		makeTristateDebuffInput({
			actionId: () => ActionId.fromSpellId(9898),
			impId: ActionId.fromSpellId(16862),
			fieldName: 'demoralizingRoar',
		}),
	],
	label: 'Attack Power',
});

// TODO: SoD Mangle
//export const BleedDebuff = withLabel(makeBooleanDebuffInput({ actionId: () => ActionId.fromSpellId(409828), fieldName: 'mangle' }), 'Bleed');

export const MeleeAttackSpeedDebuff = InputHelpers.makeMultiIconInput({
	values: [
		makeTristateDebuffInput({
			actionId: () => ActionId.fromSpellId(6343),
			impId: ActionId.fromSpellId(26110),
			fieldName: 'thunderClap',
		}),
		makeBooleanDebuffInput({
			actionId: () => ActionId.fromSpellId(21992),
			fieldName: 'thunderfury',
		}),
	],
	label: 'Attack Speed',
});

export const MeleeHitDebuff = withLabel(
	makeBooleanDebuffInput({
		actionId: () => ActionId.fromSpellId(24977),
		fieldName: 'insectSwarm',
	}),
	'Insect Swarm',
);

export const SpellISBDebuff = withLabel(
	makeBooleanDebuffInput({
		actionId: () => ActionId.fromSpellId(17803),
		fieldName: 'improvedShadowBolt',
	}),
	'Improved Shadow Bolt',
);

export const SpellScorchDebuff = withLabel(
	makeBooleanDebuffInput({
		actionId: () => ActionId.fromSpellId(12873),
		fieldName: 'improvedScorch',
	}),
	'Scorch',
);

export const SpellWintersChillDebuff = withLabel(
	makeBooleanDebuffInput({
		actionId: () => ActionId.fromSpellId(28595),
		fieldName: 'wintersChill',
	}),
	"Winter's Chill",
);

export const SpellStormstrikeDebuff = withLabel(
	makeBooleanDebuffInput({
		actionId: () => ActionId.fromSpellId(17364),
		fieldName: 'stormstrike',
	}),
	'Stormstrike',
);

export const SpellShadowWeavingDebuff = withLabel(
	makeBooleanDebuffInput({
		actionId: () => ActionId.fromSpellId(15334),
		fieldName: 'shadowWeaving',
	}),
	'Shadow Weaving',
);

export const CurseOfElements = makeBooleanDebuffInput({
	actionId: () => ActionId.fromSpellId(11722),
	fieldName: 'curseOfElements',
});

export const CurseOfShadow = makeBooleanDebuffInput({
	actionId: () => ActionId.fromSpellId(17937),
	fieldName: 'curseOfShadow',
});

export const WarlockCursesConfig = InputHelpers.makeMultiIconInput({ values: [CurseOfElements, CurseOfShadow], label: 'Warlock Curses' });

export const HuntersMark = withLabel(
	makeTristateDebuffInput({
		actionId: () => ActionId.fromSpellId(14325),
		impId: ActionId.fromSpellId(19425),
		fieldName: 'huntersMark',
	}),
	`Hunter's Mark`,
);
export const JudgementOfWisdom = withLabel(
	makeBooleanDebuffInput({
		actionId: () => ActionId.fromSpellId(20355),
		fieldName: 'judgementOfWisdom',
		showWhen: player => player.getFaction() === Faction.Alliance,
	}),
	'Judgement of Wisdom',
);
export const JudgementOfTheCrusader = withLabel(
	makeTristateDebuffInput({
		actionId: () => ActionId.fromSpellId(20303),
		impId: ActionId.fromSpellId(20337),
		fieldName: 'judgementOfTheCrusader',
		showWhen: player => player.getFaction() === Faction.Alliance,
	}),
	'Judgement of the Crusader',
);

// Misc Debuffs
export const JudgementOfLight = makeBooleanDebuffInput({
	actionId: () => ActionId.fromSpellId(20346),
	fieldName: 'judgementOfLight',
	showWhen: player => player.getFaction() === Faction.Alliance,
});
export const GiftOfArthas = makeBooleanDebuffInput({
	actionId: () => ActionId.fromSpellId(11374),
	fieldName: 'giftOfArthas',
});
export const CrystalYield = makeBooleanDebuffInput({
	actionId: () => ActionId.fromSpellId(15235),
	fieldName: 'crystalYield',
});

///////////////////////////////////////////////////////////////////////////
//                                 CONFIGS
///////////////////////////////////////////////////////////////////////////

export const RAID_BUFFS_CONFIG = [
	// Core Stat Buffs
	{
		config: AllStatsBuff,
		picker: IconPicker,
		stats: [],
	},
	{
		config: BlessingOfKings,
		picker: IconEnumPicker,
		stats: [],
	},
	{
		config: StaminaBuff,
		picker: IconPicker,
		stats: [],
	},
	{
		config: BloodPactBuff,
		picker: IconPicker,
		stats: [],
	},
	{
		config: IntellectBuff,
		picker: IconPicker,
		stats: [Stat.StatIntellect],
	},
	{
		config: SpiritBuff,
		picker: IconPicker,
		stats: [Stat.StatSpirit],
	},

	// Tank-related Buffs
	{
		config: ArmorBuff,
		picker: IconPicker,
		stats: [Stat.StatArmor],
	},
	{
		config: PhysDamReductionBuff,
		picker: IconPicker,
		stats: [Stat.StatArmor],
	},
	// {
	// 	config: DamageReductionPercentBuff,
	// 	picker: IconPicker,
	// 	stats: [Stat.StatArmor],
	// },
	{
		config: ResistanceBuff,
		picker: MultiIconPicker,
		stats: [Stat.StatNatureResistance, Stat.StatShadowResistance, Stat.StatFireResistance, Stat.StatFrostResistance],
	},

	// Physical Damage Buffs
	{
		config: BlessingOfMight,
		picker: IconPicker,
		stats: [Stat.StatAttackPower, Stat.StatStrength, Stat.StatAgility],
	},
	{
		config: StrengthBuffHorde,
		picker: IconPicker,
		stats: [Stat.StatStrength],
	},
	{
		config: BattleShoutBuff,
		picker: IconPicker,
		stats: [Stat.StatAttackPower],
	},
	{
		config: GraceOfAir,
		picker: IconPicker,
		stats: [Stat.StatAgility],
	},
	{
		config: TrueshotAuraBuff,
		picker: IconPicker,
		stats: [Stat.StatRangedAttackPower, Stat.StatAttackPower],
	},
	{
		config: MeleeCritBuff,
		picker: IconPicker,
		stats: [Stat.StatMeleeCrit],
	},
	// Threat Buffs

	// Spell Damage Buffs
	{
		config: SpellCritBuff,
		picker: IconPicker,
		stats: [Stat.StatSpellCrit],
	},
	{
		config: BlessingOfWisdom,
		picker: IconPicker,
		stats: [Stat.StatMP5],
	},
	{
		config: ManaSpringTotem,
		picker: IconPicker,
		stats: [Stat.StatMP5],
	},
] as PickerStatOptions[];

export const MISC_BUFFS_CONFIG = [
	{
		config: Thorns,
		picker: IconPicker,
		stats: [Stat.StatArmor],
	},
	{
		config: RetributionAura,
		picker: IconPicker,
		stats: [Stat.StatArmor],
	},
	{
		config: SanctityAura,
		picker: IconPicker,
		stats: [Stat.StatHolyPower],
	},
	{
		config: Innervate,
		picker: IconPicker,
		stats: [Stat.StatMP5],
	},
	{
		config: PowerInfusion,
		picker: IconPicker,
		stats: [Stat.StatMP5, Stat.StatSpellPower],
	},
	{
		config: BattleSquawkBuff,
		picker: IconPicker,
		stats: [Stat.StatMeleeHit],
	},
] as PickerStatOptions[];

export const WORLD_BUFFS_CONFIG = [
	{
		config: RallyingCryOfTheDragonslayer,
		picker: IconPicker,
		stats: [Stat.StatMeleeCrit, Stat.StatSpellCrit, Stat.StatAttackPower],
	},
	{
		config: SongflowerSerenade,
		picker: IconPicker,
		stats: [],
	},
	{
		config: SpiritOfZandalar,
		picker: IconPicker,
		stats: [],
	},
	{
		config: WarchiefsBlessing,
		picker: IconPicker,
		stats: [],
	},
	{
		config: FengusFerocity,
		picker: IconPicker,
		stats: [Stat.StatAttackPower],
	},
	{
		config: MoldarsMoxie,
		picker: IconPicker,
		stats: [Stat.StatHealth],
	},
	{
		config: SlipKiksSavvy,
		picker: IconPicker,
		stats: [Stat.StatSpellCrit],
	},
] as PickerStatOptions[];

export const SAYGES_CONFIG = [
	{
		config: SaygesDamage,
		stats: [],
	},
	{
		config: SaygesAgility,
		stats: [Stat.StatAgility],
	},
	{
		config: SaygesIntellect,
		stats: [Stat.StatIntellect],
	},
	{
		config: SaygesSpirit,
		stats: [Stat.StatSpirit, Stat.StatMP5],
	},
	{
		config: SaygesStamina,
		stats: [Stat.StatStamina],
	},
] as ItemStatOption<SaygesFortune>[];

export const DEBUFFS_CONFIG = [
	// Standard Debuffs
	{
		config: MajorArmorDebuff,
		stats: [Stat.StatAttackPower],
		picker: MultiIconPicker,
	},
	{
		config: CurseOfRecklessness,
		picker: IconPicker,
		stats: [Stat.StatAttackPower],
	},
	{
		config: FaerieFire,
		picker: IconPicker,
		stats: [Stat.StatAttackPower],
	},
	/* {
		config: BleedDebuff,
		picker: IconPicker,
		stats: [Stat.StatAttackPower, Stat.StatRangedAttackPower],
	}, */

	// Magic
	{
		config: JudgementOfTheCrusader,
		picker: IconPicker,
		stats: [Stat.StatHolyPower],
	},
	{
		config: SpellISBDebuff,
		picker: IconPicker,
		stats: [Stat.StatShadowPower],
	},
	{
		config: SpellScorchDebuff,
		picker: IconPicker,
		stats: [Stat.StatFirePower],
	},
	{
		config: SpellWintersChillDebuff,
		picker: IconPicker,
		stats: [Stat.StatFrostPower],
	},
	{
		config: SpellStormstrikeDebuff,
		picker: IconPicker,
		stats: [Stat.StatNaturePower],
	},
	{
		config: SpellShadowWeavingDebuff,
		picker: IconPicker,
		stats: [Stat.StatShadowPower],
	},
	{
		config: WarlockCursesConfig,
		picker: MultiIconPicker,
		stats: [Stat.StatSpellPower],
	},

	// Defensive
	{
		config: AttackPowerDebuff,
		picker: MultiIconPicker,
		stats: [Stat.StatArmor],
	},
	{
		config: MeleeAttackSpeedDebuff,
		picker: MultiIconPicker,
		stats: [Stat.StatArmor],
	},
	{
		config: curseOfWeaknessDebuff,
		picker: IconPicker,
		stats: [Stat.StatArmor],
	},
	{
		config: MeleeHitDebuff,
		picker: IconPicker,
		stats: [Stat.StatDodge],
	},

	// Other Debuffs
	{
		config: HuntersMark,
		picker: IconPicker,
		stats: [Stat.StatRangedAttackPower],
	},
	{
		config: JudgementOfWisdom,
		picker: IconPicker,
		stats: [Stat.StatMP5, Stat.StatIntellect],
	},
] as PickerStatOptions[];

export const MISC_DEBUFFS_CONFIG = [
	// Misc Debuffs
	{
		config: GiftOfArthas,
		picker: IconPicker,
		stats: [Stat.StatAttackPower, Stat.StatRangedAttackPower],
	},
	{
		config: CrystalYield,
		picker: IconPicker,
		stats: [Stat.StatAttackPower, Stat.StatRangedAttackPower],
	},
	{
		config: JudgementOfLight,
		picker: IconPicker,
		stats: [Stat.StatStamina],
	},
] as PickerStatOptions[];
