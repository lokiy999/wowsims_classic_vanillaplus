import {
	Consumes,
	Flask,
	Food,
} from '../core/proto/common.js';
import { SavedTalents } from '../core/proto/ui.js';

import {
	PaladinAura,
	PaladinOptions as HolyPaladinOptions,
} from '../core/proto/paladin.js';

import * as PresetUtils from '../core/preset_utils.js';

import DefaultApl from './apls/default.apl.json';
import FlashOfLightApl from './apls/flash_of_light.apl.json';
import BlankGear from './gear_sets/blank.gear.json';
import PlaceholderGear from './gear_sets/placeholder.gear.json';

// Preset options for this spec.
// Eventually we will import these values for the raid sim too, so its good to
// keep them in a separate file.

export const GearBlank = PresetUtils.makePresetGear('Blank', BlankGear);
// Placeholder: the elemental shaman phase 1 set until a healing set is made.
export const DefaultGear = PresetUtils.makePresetGear('Placeholder (caster mail)', PlaceholderGear);

export const ROTATION_PRESET_HOLY_LIGHT = PresetUtils.makePresetAPLRotation('Holy Light', DefaultApl);
export const ROTATION_PRESET_FLASH_OF_LIGHT = PresetUtils.makePresetAPLRotation('Flash of Light', FlashOfLightApl);

// Default talents. Uses the wowhead calculator format, make the talents on
// https://wowhead.com/classic/talent-calc and copy the numbers in the url.

export const StandardTalents = {
	name: 'Standard',
	data: SavedTalents.create({
		talentsString: '',
	}),
};

export const DefaultOptions = HolyPaladinOptions.create({
	aura: PaladinAura.DevotionAura,
});

export const DefaultConsumes = Consumes.create({
	flask: Flask.FlaskUnknown,
	food: Food.FoodUnknown,
});
