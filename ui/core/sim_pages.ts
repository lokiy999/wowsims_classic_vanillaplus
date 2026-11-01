import { Spec } from './proto/common';
import { getSpecSiteUrl } from './proto_utils/utils';

// Extra pages that run another spec's sim with their own defaults (Part IJ): the same simulation, but their own
// address, title, spec-switcher entry and saved settings.
export interface SimPageVariant {
	spec: Spec;
	// Folder under /classic/ (ui/<path>/index.html).
	path: string;
	// Name under the class in the spec switcher.
	label: string;
	// Name in the page header.
	title: string;
	icon: string;
	// Browser storage prefix. Never change it once released, or people lose their saved settings.
	storageKey: string;
}

export const simPageVariants: Array<SimPageVariant> = [
	{
		spec: Spec.SpecShadowPriest,
		path: 'smite_priest',
		label: 'Smite',
		title: 'Smite Priest',
		icon: 'https://db.lokiy.dev/icons/spell_holy_holysmite.jpg',
		storageKey: '__classic_smite_priest',
	},
];

// The variant this page is, if it is one.
export function currentSimPageVariant(): SimPageVariant | undefined {
	const parts = window.location.pathname.split('/');
	return simPageVariants.find(variant => parts.includes(variant.path));
}

export function getSimPageVariantUrl(variant: SimPageVariant): string {
	return getSpecSiteUrl(variant.spec).replace(/[^/]+\/$/, `${variant.path}/`);
}
