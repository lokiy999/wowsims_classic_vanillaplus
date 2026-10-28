// db.lokiy.dev: search (everything, grouped), the Search menu and hover tooltips. Data: /items.json [{id, n: name, q: quality, i: small icon URL, s?: sources,
// c?: 1 if new, r?: 1 if reworked}], hover tooltips in /tt/<id // 500>.json {id: html}
(() => {
	const QC = ['#9d9d9d', '#ffffff', '#1eff00', '#0070dd', '#a335ee', '#ff8000', '#e6cc80'];
	const ICON = icon => `/icons/${icon}_s.jpg`;
	let dataPromise = null;
	const loadData = () => {
		if (!dataPromise) {
			dataPromise = fetch('/items.json')
				.then(r => r.json())
				.then(list => new Map(list.map(it => [String(it.id), it])))
				.catch(() => new Map());
		}
		return dataPromise;
	};
	const ttChunks = new Map();
	const loadTooltip = id => {
		const n = Math.floor(Number(id) / 500);
		if (!ttChunks.has(n)) {
			ttChunks.set(
				n,
				fetch(`/tt/${n}.json`)
					.then(r => r.json())
					.catch(() => ({})),
			);
		}
		return ttChunks.get(n).then(chunk => chunk[id]);
	};
	const escapeHtml = s => s.replace(/[&<>"]/g, c => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;' })[c]);

	// ---- search (Part GQ): the header box searches everything. The index and reworked pages show the hits inline,
	// grouped by kind with tabs (All shows the best few of each kind); other pages submit to /?q=.
	// Data: /items.json; /npcs.json [[id, name, level, zone, rank, has loot]]; /api/mobs.json {name: url} (loot tracker
	// mobs, the ones without a creature id keep a /mob page); /spells.json [[id, name, rank, icon, status, class]];
	// /quests.json [[id, name, level, side, zone]]; /api/zones.json {id: {name, npcs, quests}}; /tables.json
	// [[key, name, group, not on V+]]. Each kind loads on the first search.
	const input = document.querySelector('.search input');
	const results = document.getElementById('results');
	const browse = document.getElementById('browse');
	const getJson = url => fetch(url).then(r => r.json()).catch(() => null);
	const meta = parts => parts.filter(Boolean).join(' · ');
	const KINDS = [
		{
			key: 'items', label: 'Items', page: q => `/search?q=${q}`,
			load: () => loadData().then(m => [...m.values()].map(it => ({
				h: `/item/${it.id}`, id: it.id, item: 1, n: it.n, i: it.i, c: QC[it.q], r: it.q, also: it.s || [],
				m: (it.s || []).slice(0, 3).join(', '), t: it.c ? 'new' : it.r ? 'changed' : '',
			}))),
		},
		{
			key: 'npcs', label: 'NPCs', page: q => `/npcs?q=${q}`,
			load: () => Promise.all([getJson('/npcs.json'), getJson('/api/mobs.json')]).then(([npcs, mobs]) => [
				...(npcs || []).map(n => ({
					h: `/npc/${n[0]}`, id: n[0], n: n[1], also: [n[3]], r: n[4] ? 1 : 0,
					m: meta([n[2] && `Level ${n[2]}`, n[4], n[3], `#${n[0]}`]), t: n[5] ? 'loot' : '',
				})),
				...Object.entries(mobs || {}).filter(([, u]) => u.includes('/mob/'))
					.map(([name, u]) => ({ h: new URL(u).pathname, n: name, m: 'loot tracker mob, no creature ID yet', also: [] })),
			]),
		},
		{
			key: 'spells', label: 'Spells', page: q => `/spells?q=${q}`,
			load: () => getJson('/spells.json').then(list => (list || []).map(s => ({
				h: `/spell/${s[0]}`, id: s[0], n: s[1] || '(no name)', i: `/icons/${s[3] || 'inv_misc_questionmark'}_s.jpg`,
				m: meta([s[2], s[5], `#${s[0]}`]), t: s[4] === 'new' || s[4] === 'changed' ? s[4] : '', also: [],
			}))),
		},
		{
			key: 'quests', label: 'Quests', page: q => `/quests?q=${q}`,
			load: () => getJson('/quests.json').then(list => (list || []).map(x => ({
				h: `/quest/${x[0]}`, id: x[0], n: x[1], also: [x[4]], m: meta([x[2] && `Level ${x[2]}`, x[3], x[4], `#${x[0]}`]),
			}))),
		},
		{
			key: 'zones', label: 'Zones', page: () => '',
			load: () => getJson('/api/zones.json').then(z => Object.entries(z || {}).map(([id, v]) => ({
				h: `/zone/${id}`, id, n: v.name, also: [], m: `${v.npcs} NPCs, ${v.quests} quests`,
			}))),
		},
		{
			key: 'tables', label: 'Loot tables', page: () => '',
			load: () => getJson('/tables.json').then(list => (list || []).map(([k, name, group, cut]) => ({
				h: `/loot/${k}`, n: name, also: [group], m: group, t: cut ? 'not on Vanilla+' : '',
			}))),
		},
	];
	const loaded = {};
	const loadKind = k => (loaded[k.key] = loaded[k.key] || k.load().catch(() => []));
	// 0 = exact name or id, 1 = name starts with it, 2 = a word starts with it, 3 = in the name, 4 = only in zone / source
	const score = (e, q) => {
		const n = e.n.toLowerCase();
		if (n === q || String(e.id) === q) return 0;
		if (n.startsWith(q)) return 1;
		if (n.includes(` ${q}`)) return 2;
		if (n.includes(q)) return 3;
		return e.also.some(a => a && a.toLowerCase().includes(q)) ? 4 : -1;
	};
	const row = e =>
		`<li><a${e.item ? ` class="item" data-item="${e.id}"` : e.i ? ' class="item"' : ''} href="${e.h}">` +
		(e.i ? `<img class="icon-s" src="${e.i}" alt="" loading="lazy" onerror="this.onerror=null;this.src='${ICON('inv_misc_questionmark')}'">` : '') +
		`<span${e.c ? ` style="color:${e.c}"` : ''}>${escapeHtml(e.n)}</span></a>` +
		(e.t ? `<span class="tag">${escapeHtml(e.t)}</span>` : '') + (e.m ? ` <span class="muted">${escapeHtml(e.m)}</span>` : '') + '</li>';
	const PER_KIND = 8;
	const MAX = 200;
	let tab = new URLSearchParams(location.search).get('t') || '';
	let lastQ = '';
	let lastHits = null;
	let seq = 0;
	const render = () => {
		const q = encodeURIComponent(lastQ);
		const found = KINDS.filter(k => lastHits[k.key].length);
		if (tab && !lastHits[tab].length) tab = '';
		const total = found.reduce((s, k) => s + lastHits[k.key].length, 0);
		const tabLink = (key, label, n) => `<a href="#" data-tab="${key}" class="${tab === key ? 'on' : ''}">${label} <span class="muted">${n.toLocaleString()}</span></a>`;
		document.getElementById('result-tabs').innerHTML = found.length > 1
			? tabLink('', 'All', total) + found.map(k => tabLink(k.key, k.label, lastHits[k.key].length)).join('')
			: '';
		const group = (k, n) => {
			const hits = lastHits[k.key];
			const more = hits.length > n
				? (tab ? k.page(q) && `<a href="${k.page(q)}">All ${hits.length.toLocaleString()} on the ${k.label} page &rarr;</a>`
					: `<a href="#" data-tab="${k.key}">Show all ${hits.length.toLocaleString()} ${k.label.toLowerCase()} &rarr;</a>`)
				: tab && k.page(q) ? `<a href="${k.page(q)}">${k.key === 'items' ? 'Open these in the item filters (slot, class, stats...)' : `Open these on the ${k.label} page`} &rarr;</a>` : '';
			return `<div class="result-group"><h3>${k.label} <span class="muted">${hits.length.toLocaleString()}</span></h3>` +
				`<ul class="loot">${hits.slice(0, n).map(row).join('')}</ul>${more ? `<p class="result-more">${more}</p>` : ''}</div>`;
		};
		document.getElementById('result-groups').innerHTML = !found.length
			? '<p class="muted">Nothing found.</p>'
			: tab ? group(KINDS.find(k => k.key === tab), MAX) : found.map(k => group(k, PER_KIND)).join('');
	};
	const setUrl = () => {
		const url = new URL(location.href);
		lastQ ? url.searchParams.set('q', lastQ) : url.searchParams.delete('q');
		tab ? url.searchParams.set('t', tab) : url.searchParams.delete('t');
		history.replaceState(null, '', url);
	};
	const runSearch = async raw => {
		if (!results) return;
		const q = raw.trim().toLowerCase();
		const my = ++seq;
		lastQ = raw.trim();
		if (!q) {
			results.hidden = true;
			browse.hidden = false;
			return;
		}
		results.hidden = false;
		browse.hidden = true;
		if (!lastHits) document.getElementById('result-groups').innerHTML = '<p class="muted">Searching...</p>';
		const lists = await Promise.all(KINDS.map(loadKind));
		if (my !== seq) return;
		lastHits = {};
		KINDS.forEach((k, n) => {
			lastHits[k.key] = lists[n]
				.map(e => [score(e, q), e])
				.filter(([s]) => s >= 0)
				.sort((a, b) => a[0] - b[0] || (b[1].r || 0) - (a[1].r || 0) || a[1].n.length - b[1].n.length || a[1].n.localeCompare(b[1].n))
				.map(([, e]) => e);
		});
		render();
	};
	if (input && results) {
		const q = new URLSearchParams(location.search).get('q') || '';
		input.value = q;
		if (q) runSearch(q);
		input.addEventListener('input', () => {
			runSearch(input.value);
			setUrl();
		});
		input.form.addEventListener('submit', e => e.preventDefault());
		results.addEventListener('click', e => {
			const a = e.target.closest('[data-tab]');
			if (!a) return;
			e.preventDefault();
			tab = a.dataset.tab;
			render();
			setUrl();
			results.scrollIntoView({ block: 'start' });
		});
	}

	// ---- the nav's Search menu: opens on hover / focus (CSS) and on click (touch)
	for (const drop of document.querySelectorAll('.nav-drop')) {
		const btn = drop.querySelector('button');
		btn.addEventListener('click', () => btn.setAttribute('aria-expanded', String(drop.classList.toggle('open'))));
		document.addEventListener('click', e => {
			if (!drop.contains(e.target)) {
				drop.classList.remove('open');
				btn.setAttribute('aria-expanded', 'false');
			}
		});
	}

	// ---- hover tooltips on item links (mouse only)
	if (!matchMedia('(hover: hover)').matches) return;
	const tt = document.createElement('div');
	tt.id = 'hover-tt';
	tt.hidden = true;
	document.body.appendChild(tt);
	let current = null;
	const place = e => {
		const pad = 14;
		let x = e.clientX + pad;
		let y = e.clientY + pad;
		const r = tt.getBoundingClientRect();
		if (x + r.width > innerWidth - 8) x = e.clientX - r.width - pad;
		if (y + r.height > innerHeight - 8) y = innerHeight - r.height - 8;
		tt.style.left = `${Math.max(8, x)}px`;
		tt.style.top = `${Math.max(8, y)}px`;
	};
	document.addEventListener('mouseover', async e => {
		const a = e.target.closest && e.target.closest('a[data-item]');
		if (!a || a === current) return;
		current = a;
		const html = await loadTooltip(a.dataset.item);
		if (!html || current !== a) return;
		tt.innerHTML = html;
		tt.hidden = false;
		place(e);
	});
	document.addEventListener('mousemove', e => {
		if (!tt.hidden) place(e);
	});
	document.addEventListener('mouseout', e => {
		const a = e.target.closest && e.target.closest('a[data-item]');
		if (a && !a.contains(e.relatedTarget)) {
			current = null;
			tt.hidden = true;
		}
	});
})();
