// db.lokiy.dev/ilvl (sim repo Part HY): the item level calculator and the report table, from /api/ilvl.json
// (ilvl_page.py). The calculator repeats ilvl_model.py's Model.predict() with the saved fit.
(() => {
	const $ = id => document.getElementById(id);
	const esc = s => String(s == null ? '' : s).replace(/[&<>"']/g, c => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' })[c]);
	const QC = { 2: '#1eff00', 3: '#0070dd', 4: '#a335ee' };
	const QN = { 2: 'Green', 3: 'Blue', 4: 'Epic' };
	const SLOTS = [['Head', 'A'], ['Chest', 'A'], ['Legs', 'A'], ['Shoulder', 'B'], ['Hands', 'B'], ['Waist', 'B'],
		['Feet', 'B'], ['Wrist', 'C'], ['Neck', 'C'], ['Back', 'C'], ['Finger', 'C'], ['Held In Off-hand', 'C'],
		['Off Hand', 'S'], ['One-Hand', 'W1'], ['Two-Hand', 'W2'], ['Ranged', 'WR'], ['Wand', 'WAND']];
	const STATS = [['Strength', 'prim'], ['Agility', 'prim'], ['Stamina', 'prim'], ['Intellect', 'prim'], ['Spirit', 'prim'],
		['Resistance (each school)', 'res', 5], ['Spell damage and healing', 'sd'], ['Healing', 'heal'],
		['Damage, one school', 'school'], ['Second school', 'school2'], ['Crit %', 'crit'], ['Hit %', 'hit'],
		['Spell crit %', 'scrit'], ['Spell hit %', 'shit'], ['Attack power', 'ap'], ['Ranged attack power', 'rap'],
		['Mana per 5 sec', 'mp5'], ['Health per 5 sec', 'hp5'], ['Defense', 'def'], ['Dodge %', 'dodge'],
		['Parry %', 'parry'], ['Block %', 'block'], ['Block value', 'bv'], ['Weapon skill', 'wskill'],
		['Spell penetration', 'spen']];
	let D;

	// ---- the model (ilvl_model.py Model)
	const armorKey = (atype, slot, q) => `${slot === 'Back' ? 'Cloak' : atype}|${slot}|${q}`;
	function value(r, barmor, bblock) {
		const p = D.fit.params;
		const parts = r.stats.map(([k, n]) => n * p['w_' + k]);
		if (bblock > 0) parts.push(bblock * p.w_bv);
		const sch = r.schools.slice().sort((a, b) => b - a);
		if (sch.length) parts.push(sch[0] * p.w_school + sch.slice(1).reduce((s, x) => s + x * p.w_school * p.f_school2, 0));
		if (barmor > 0) parts.push(barmor * p.w_barmor);
		let v = parts.filter(x => x > 0).reduce((s, x) => s + x ** 1.5, 0) ** (1 / 1.5);
		return r.set ? v * (1 + p.f_set) : v;
	}
	function bonusArmor(r, ilvl) {
		const ab = r.armor && D.fit.armor[armorKey(r.atype, r.slot, r.q)];
		if (!ab) return 0;
		const base = ab[0] * ilvl + ab[1];
		return r.armor - base > 0.15 * base ? r.armor - base : 0;
	}
	// a shield's own block value is base like armor: only what is above the fitted base counts (ilvl_model.py)
	function bonusBlock(r, ilvl) {
		const ab = r.sblock && D.fit.block && D.fit.block[r.q];
		if (!ab) return 0;
		const base = ab[0] * ilvl + ab[1];
		return r.sblock - base > 0.15 * base ? r.sblock - base : 0;
	}
	// a caster / healer weapon: its own slot modifier, and the higher of the DPS and the stats item level (in 1.12
	// it spends its budget on spell power or on DPS; ilvl_model.py caster / slot_mod)
	const caster = r => (r.group === 'W1' || r.group === 'W2') &&
		(r.schools.length > 0 || r.stats.some(([k]) => ['heal', 'sd', 'scrit', 'shit', 'spen'].includes(k)));
	function predict(r) {
		const [a, b] = (caster(r) && D.fit.budget_c && D.fit.budget_c[r.q]) || D.fit.budget[r.q];
		const mod = D.fit.params['m_' + r.group + (caster(r) ? 'C' : '')];
		let st = null;
		if (r.stats.length || r.schools.length) {
			// bonus armor / block depend on the ilvl: bisect for the level whose budget the value meets (ilvl_model.py)
			const stAt = x => (value(r, bonusArmor(r, x), bonusBlock(r, x)) / mod - b) / a;
			let lo = -50, hi = 250;
			for (let i = 0; i < 40; i++) {
				const mid = (lo + hi) / 2;
				if (stAt(mid) > mid) lo = mid; else hi = mid;
			}
			st = (lo + hi) / 2;
		}
		const ab = D.fit.dps[`${r.group}|${r.q}`];
		const dp = ab && r.dps && ab[0] ? (r.dps - ab[1]) / ab[0] : null;
		const comb = dp != null && st != null ? (caster(r) ? Math.max(st, dp) : 0.7 * dp + 0.3 * st) : dp != null ? dp : st;
		return [st, dp, comb];
	}

	// ---- calculator
	function calcInit() {
		const f = $('calc');
		f.slot.innerHTML = SLOTS.map(([s, g]) => `<option value="${esc(s)}" data-g="${g}">${esc(s)}</option>`).join('');
		$('stat-inputs').innerHTML = STATS.map(([label, k], i) =>
			`<label>${esc(label)}<input data-i="${i}" type="number" min="0" step="any"></label>`).join('');
		f.addEventListener('input', calc);
		calc();
	}
	function calc() {
		const f = $('calc');
		const opt = f.slot.selectedOptions[0];
		const q = +f.q.value;
		const r = { slot: opt.value, group: opt.dataset.g, atype: f.atype.value || null, q, armor: +f.armor.value || 0,
			dps: +f.dps.value || 0, sblock: +f.sblock.value || 0, set: f.set.checked, stats: [], schools: [] };
		if (r.slot === 'Wand') r.slot = 'Ranged';
		for (const inp of $('stat-inputs').querySelectorAll('input')) {
			const n = +inp.value;
			if (!n) continue;
			const [, k, times] = STATS[+inp.dataset.i];
			if (k === 'school' || k === 'school2') r.schools.push(n);
			else for (let t = 0; t < (times || 1); t++) r.stats.push([k, n]);
		}
		const [st, dp, comb] = predict(r);
		const h = D.fit.heldout.find(x => x.q === q && x.group === r.group) || D.fit.heldout.find(x => x.q === q && x.group === 'all');
		if (comb == null) {
			$('calc-out').innerHTML = '<p class="muted">Enter stats, or a weapon\'s DPS.</p>';
			return;
		}
		$('calc-out').innerHTML =
			`<p class="big">Model item level <b>${comb.toFixed(1)}</b>${f.effect.checked ? ' <span class="tag">at least: effect not priced</span>' : ''}</p>` +
			`<p class="muted">From the stats ${st == null ? '-' : st.toFixed(1)}${dp == null ? '' : ` &middot; from the DPS ${dp.toFixed(1)}`}` +
			(h ? ` &middot; typical miss for ${esc(QN[q].toLowerCase())} items here: ${h.median} (90% within ${h.p90})` : '') + '</p>';
	}

	// ---- report; a click on a column head sorts by it (again: the other way), empty values last
	let sortKey = null, sortDir = 1;
	const SORT = {
		name: r => r.name, tier: r => r.tier, group: r => D.groupLabel[r.group] || r.group, ilvl: r => r.ilvl,
		model: r => r.model, diff: r => r.model == null ? null : r.model - r.ilvl, flags: r => r.flags.join(', ') || null,
		suggest: r => r.suggest, comp: r => r.comp ? r.comp.length : null,
	};
	const TEXT_KEYS = ['name', 'group', 'flags'];
	function repRender() {
		const f = $('rep-filters');
		const name = f.name.value.trim().toLowerCase();
		const list = D.items.filter(r =>
			(f.tier.value === '' || (f.tier.value === 'none' ? r.tier == null : r.tier === +f.tier.value)) &&
			(f.q.value === '' || r.q === +f.q.value) &&
			(f.group.value === '' || r.group === f.group.value) &&
			(f.flag.value === '' || r.flags.includes(f.flag.value)) &&
			(!name || r.name.toLowerCase().includes(name)));
		if (sortKey) {
			const f = SORT[sortKey];
			list.sort((a, b) => {
				const x = f(a), y = f(b);
				if (x == null || y == null) return (x == null) - (y == null);
				return (typeof x === 'string' ? x.localeCompare(y) : x - y) * sortDir;
			});
		}
		$('rep-count').textContent = `${list.length.toLocaleString()} item${list.length === 1 ? '' : 's'}`;
		const tierName = t => t == null ? '' : D.tiers[t].tier.split('|')[0];
		const rows = list.slice(0, 500).map(r => {
			const diff = r.model == null ? '' : (r.model - r.ilvl).toFixed(1);
			const comp = r.comp ? `<details><summary>${r.comp.length} comparable</summary>${r.comp.map(([id, nm, il]) =>
				`<a href="/item/${id}" data-item="${id}">${esc(nm)}</a> ${il}`).join('<br>')}</details>` : '';
			return `<tr><td class="l"><a href="/item/${r.id}" data-item="${r.id}" style="color:${QC[r.q]}">${esc(r.name)}</a>` +
				`${r.status === 3 ? ' <span class="tag">new</span>' : ''}</td><td class="l muted">${esc(tierName(r.tier))}</td>` +
				`<td class="l muted">${esc(D.groupLabel[r.group] || r.group)}</td><td>${r.ilvl}</td>` +
				`<td>${r.model == null ? '' : r.model.toFixed(1)}</td><td>${diff > 0 ? '+' : ''}${diff}</td>` +
				`<td class="l">${r.flags.map(x => `<span class="tag">${esc(x)}</span>`).join(' ')}${r.why ? `<br><span class="muted small">${esc(r.why)}</span>` : ''}</td>` +
				`<td>${r.suggest == null ? '' : `<b>${r.suggest}</b>`}</td><td class="l small">${comp}</td></tr>`;
		}).join('');
		const th = (key, label, cls = '') => `<th class="${cls}"><button type="button" data-sort="${key}">${label}` +
			`${sortKey === key ? (sortDir < 0 ? ' &#9662;' : ' &#9652;') : ''}</button></th>`;
		$('rep').innerHTML = '<thead><tr>' + th('name', 'Item', 'l') + th('tier', 'Raid', 'l') + th('group', 'Slot group', 'l') +
			th('ilvl', 'Real ilvl') + th('model', 'Model') + th('diff', 'Model &minus; real') + th('flags', 'Flags', 'l') +
			th('suggest', 'Suggested') + th('comp', 'Based on', 'l') + '</tr></thead>' +
			`<tbody>${rows}</tbody>` + (list.length > 500 ? `<tfoot><tr><td colspan="9" class="muted">First 500 shown; filter to see the rest.</td></tr></tfoot>` : '');
	}

	fetch('/api/ilvl.json').then(r => r.json()).then(d => {
		D = d;
		D.groupLabel = { A: 'Head / chest / legs', B: 'Shoulder / hands / waist / feet', C: 'Wrist / neck / back / finger',
			S: 'Shield', W1: 'One-hand weapon', W2: 'Two-hand weapon', WR: 'Ranged weapon', WAND: 'Wand' };
		calcInit();
		$('rep-filters').addEventListener('input', repRender);
		$('rep').addEventListener('click', e => {
			const b = e.target.closest('[data-sort]');
			if (!b) return;
			const k = b.dataset.sort;
			if (sortKey === k) sortDir = -sortDir;
			else {
				sortKey = k;
				sortDir = TEXT_KEYS.includes(k) || k === 'tier' ? 1 : -1;
			}
			repRender();
		});
		repRender();
	});
})();
