import { Player } from '../core/player.js';
import { Spec } from '../core/proto/common.js';
import { Sim } from '../core/sim.js';
import { TypedEvent } from '../core/typed_event.js';
import { SmitePriestSimUI } from './sim.js';

// Smite Priest (Part IJ): the Shadow Priest sim with Holy defaults; see ui/core/sim_pages.ts.
const sim = new Sim();
const player = new Player<Spec.SpecShadowPriest>(Spec.SpecShadowPriest, sim);
sim.raid.setPlayer(TypedEvent.nextEventID(), 0, player);

new SmitePriestSimUI(document.body, player);
