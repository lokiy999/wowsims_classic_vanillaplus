import * as PresetUtils from '../core/preset_utils.js';
import { Consumes, Flask, Food, WeaponImbue } from '../core/proto/common.js';
import { RestorationShaman_Options as RestorationShamanOptions } from '../core/proto/shaman.js';
import { SavedTalents } from '../core/proto/ui.js';
import ChainHealApl from './apls/chain_heal.apl.json';
import DefaultApl from './apls/default.apl.json';
import BlankGear from './gear_sets/blank.gear.json';
import PlaceholderGear from './gear_sets/placeholder.gear.json';

// Preset options for this spec.
// Eventually we will import these values for the raid sim too, so its good to
// keep them in a separate file.

export const GearBlank = PresetUtils.makePresetGear('Blank', BlankGear);
// Placeholder: the elemental phase 1 set until a healing set is made.
export const DefaultGear = PresetUtils.makePresetGear('Placeholder (elemental P1)', PlaceholderGear);

export const ROTATION_PRESET_HEALING_WAVE = PresetUtils.makePresetAPLRotation('Healing Wave', DefaultApl);
export const ROTATION_PRESET_CHAIN_HEAL = PresetUtils.makePresetAPLRotation('Chain Heal', ChainHealApl);

// Default talents. Uses the wowhead calculator format, make the talents on
// https://wowhead.com/classic/talent-calc and copy the numbers in the url.
export const TankHealingTalents = {
	name: 'Tank Healing',
	data: SavedTalents.create({
		talentsString: '',
	}),
};
export const RaidHealingTalents = {
	name: 'Raid Healing',
	data: SavedTalents.create({
		talentsString: '',
	}),
};

export const DefaultOptions = RestorationShamanOptions.create({
	earthShieldPPM: 0,
});

export const DefaultConsumes = Consumes.create({
	flask: Flask.FlaskUnknown,
	food: Food.FoodUnknown,
	mainHandImbue: WeaponImbue.RockbiterWeapon,
	offHandImbue: WeaponImbue.RockbiterWeapon,
});
