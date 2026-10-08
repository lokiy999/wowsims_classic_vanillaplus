// db.lokiy.dev/search: filter all items by name, quality, type, slot, level, class, binding, status, source and stats.
// Data: /search.json {types, sections, cats, classes, stats, items: [[id, name, quality, type, section, level, slot,
// bind, classMask, status, stats, cats, icon]]}. The filter state lives in the URL, so a search can be linked.
(() => {
	const QC = ['#9d9d9d', '#ffffff', '#1eff00', '#0070dd', '#a335ee', '#ff8000', '#e6cc80'];
	const QN = ['Poor', 'Common', 'Uncommon', 'Rare', 'Epic', 'Legendary'];
	const STATUS = { 0: 'same as classic', 1: 'classic', 2: 'changed', 3: 'new' };
	const BIND = { bop: 'Binds when picked up', boe: 'Binds when equipped', bou: 'Binds when used', quest: 'Quest item' };
	const PER_PAGE = 50;
	const ZAM = n => `https://wow.zamimg.com/images/wow/icons/small/${n}.jpg`;
	const $ = id => document.getElementById(id);
	const esc = s => String(s).replace(/[&<>"]/g, c => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;' })[c]);
	const form = $('filters');
	let D = null;
	let rows = [];
	let sortKey = 'q';
	let sortDir = -1;
	let pageNo = 0;

	const opt = (v, label, sel) => `<option value="${esc(v)}"${sel ? ' selected' : ''}>${esc(label)}</option>`;

	const statRow = (k = '', v = '') => {
		const div = document.createElement('div');
		div.className = 'stat-row';
		div.innerHTML =
			`<select name="sk" aria-label="Stat">${opt('', 'Stat...')}${D.stats.map(([key, label]) => opt(key, label, key === k)).join('')}</select>` +
			`<span class="muted">&ge;</span><input name="sv" type="number" min="0" step="any" value="${esc(v)}" placeholder="any" aria-label="Minimum">` +
			`<button type="button" class="x" aria-label="Remove stat filter">&times;</button>`;
		div.querySelector('.x').addEventListener('click', () => {
			div.remove();
			update();
		});
		$('stat-rows').appendChild(div);
	};

	const fillSections = () => {
		const t = form.type.value;
		const secs = t === '' ? [] : [...new Set(D.items.filter(it => D.types[it[3]][0] === t).map(it => it[4]))];
		secs.sort((a, b) => D.sections[a].localeCompare(D.sections[b]));
		const cur = new URLSearchParams(location.search).get('sec') || '';
		form.sec.innerHTML = opt('', 'All') + secs.map(s => opt(D.sections[s], D.sections[s], D.sections[s] === cur)).join('');
		form.sec.disabled = !secs.length;
	};

	const readState = () => {
		const p = new URLSearchParams(location.search);
		form.q.value = p.get('q') || '';
		for (const cb of form.querySelectorAll('[name=ql]')) cb.checked = (p.get('ql') || '').split(',').includes(cb.value);
		for (const n of ['type', 'slot', 'cls', 'bind', 'st', 'src', 'lmin', 'lmax']) form[n].value = p.get(n) || '';
		fillSections();
		form.sec.value = p.get('sec') || '';
		$('stat-rows').innerHTML = '';
		for (const s of (p.get('stats') || '').split(',').filter(Boolean)) {
			const [k, v] = s.split(':');
			statRow(k, v || '');
		}
		if (p.get('sort')) {
			sortKey = p.get('sort').replace(/^-/, '');
			sortDir = p.get('sort').startsWith('-') ? -1 : 1;
		}
		pageNo = Math.max(0, (parseInt(p.get('p'), 10) || 1) - 1);
	};

	const statFilters = () =>
		[...$('stat-rows').querySelectorAll('.stat-row')]
			.map(r => [r.querySelector('[name=sk]').value, r.querySelector('[name=sv]').value])
			.filter(([k]) => k);

	const writeState = () => {
		const p = new URLSearchParams();
		if (form.q.value.trim()) p.set('q', form.q.value.trim());
		const ql = [...form.querySelectorAll('[name=ql]:checked')].map(cb => cb.value);
		if (ql.length) p.set('ql', ql.join(','));
		for (const n of ['type', 'sec', 'slot', 'cls', 'bind', 'st', 'src', 'lmin', 'lmax']) if (form[n].value) p.set(n, form[n].value);
		const sf = statFilters();
		if (sf.length) p.set('stats', sf.map(([k, v]) => (v ? `${k}:${v}` : k)).join(','));
		if (!(sortKey === 'q' && sortDir === -1)) p.set('sort', (sortDir < 0 ? '-' : '') + sortKey);
		if (pageNo) p.set('p', pageNo + 1);
		const qs = p.toString();
		history.replaceState(null, '', qs ? `?${qs}` : location.pathname);
	};

	const filter = () => {
		const q = form.q.value.trim().toLowerCase();
		const ql = new Set([...form.querySelectorAll('[name=ql]:checked')].map(cb => +cb.value));
		const type = form.type.value;
		const sec = form.sec.value;
		const slot = form.slot.value;
		const cls = form.cls.value === '' ? -1 : +form.cls.value;
		const bind = form.bind.value;
		const st = form.st.value;
		const src = form.src.value === '' ? -1 : +form.src.value;
		const lmin = form.lmin.value === '' ? -1 : +form.lmin.value;
		const lmax = form.lmax.value === '' ? 999 : +form.lmax.value;
		const sf = statFilters().map(([k, v]) => [k, v === '' ? null : +v]);
		return D.items.filter(it => {
			const [id, name, quality, t, s, lv, sl, b, cm, status, stats, cats] = it;
			if (q && !name.toLowerCase().includes(q) && String(id) !== q) return false;
			if (ql.size && !ql.has(quality)) return false;
			if (type && D.types[t][0] !== type) return false;
			if (sec && D.sections[s] !== sec) return false;
			if (slot && sl !== slot) return false;
			if (cls >= 0 && !(cm & (1 << cls))) return false;
			if (bind && (bind === 'none' ? b : b !== bind)) return false;
			if (st === 'new' && status !== 3) return false;
			if (st === 'changed' && status !== 2) return false;
			if (st === 'vp' && status < 2) return false;
			if (st === 'same' && status !== 0) return false;
			if (src >= 0 && !(cats && cats.includes(src))) return false;
			if (lv < lmin || lv > lmax) return false;
			for (const [k, min] of sf) {
				const v = stats ? stats[k] : undefined;
				if (v === undefined || (min !== null && v < min)) return false;
			}
			return true;
		});
	};

	const columns = () => {
		const cols = [];
		const type = form.type.value;
		const sf = statFilters().map(([k]) => k);
		if (type === 'weapons' || type === 'ranged') cols.push('dps', 'speed');
		if (['cloth', 'leather', 'mail', 'plate'].includes(type) || form.sec.value === 'Shields') cols.push('armor');
		for (const k of sf) if (!cols.includes(k)) cols.push(k);
		return cols;
	};
	const statLabel = k => (D.stats.find(s => s[0] === k) || [k, k])[1];
	const shortLabel = k => ({ dps: 'DPS', speed: 'Speed', armor: 'Armor' })[k] || statLabel(k).replace(/ %$/, '%');

	const sortVal = (it, key) => {
		if (key === 'q') return it[2];
		if (key === 'name') return it[1].toLowerCase();
		if (key === 'lv') return it[5];
		if (key === 'type') return D.sections[it[4]];
		if (key === 'slot') return it[6] || '';
		return it[10] && it[10][key] !== undefined ? it[10][key] : -1;
	};

	const render = () => {
		const cols = columns();
		const list = rows.slice().sort((a, b) => {
			const x = sortVal(a, sortKey);
			const y = sortVal(b, sortKey);
			const c = typeof x === 'string' ? x.localeCompare(y) : x - y;
			return c * sortDir || b[2] - a[2] || a[1].localeCompare(b[1]);
		});
		const pages = Math.max(1, Math.ceil(list.length / PER_PAGE));
		pageNo = Math.min(pageNo, pages - 1);
		const shown = list.slice(pageNo * PER_PAGE, (pageNo + 1) * PER_PAGE);
		const th = (key, label, cls = '') =>
			`<th class="${cls}"><button type="button" data-sort="${key}">${esc(label)}${sortKey === key ? (sortDir < 0 ? ' &#9662;' : ' &#9652;') : ''}</button></th>`;
		const head =
			'<tr>' +
			th('name', 'Item', 'l') +
			th('lv', 'Req') +
			th('slot', 'Slot', 'l hide-s') +
			th('type', 'Type', 'l hide-s') +
			cols.map(k => th(k, shortLabel(k))).join('') +
			'<th class="hide-s"></th></tr>';
		const body = shown
			.map(it => {
				const [id, name, quality, , s, lv, sl, , , status, stats, , icon] = it;
				const src = icon ? (icon.startsWith('/') ? icon : ZAM(icon)) : ZAM('inv_misc_questionmark');
				const tag = status >= 2 ? `<span class="tag">${STATUS[status]}</span>` : '';
				return (
					`<tr><td class="l"><a class="item" href="/item/${id}" data-item="${id}"><img class="icon-s" src="${esc(src)}" alt="" loading="lazy" onerror="this.onerror=null;this.src='${ZAM('inv_misc_questionmark')}'">` +
					`<span style="color:${QC[quality] || '#fff'}">${esc(name)}</span></a></td>` +
					`<td>${lv || ''}</td><td class="l hide-s">${esc(sl || '')}</td><td class="l hide-s muted">${esc(D.sections[s])}</td>` +
					cols.map(k => `<td>${stats && stats[k] !== undefined ? stats[k] : ''}</td>`).join('') +
					`<td class="hide-s">${tag}</td></tr>`
				);
			})
			.join('');
		$('count').textContent = `${list.length.toLocaleString()} item${list.length === 1 ? '' : 's'}`;
		$('results-table').innerHTML = list.length
			? `<thead>${head}</thead><tbody>${body}</tbody>`
			: '<tbody><tr><td class="muted">No items match these filters.</td></tr></tbody>';
		const pager = pages > 1
			? `<button type="button" data-page="${pageNo - 1}"${pageNo ? '' : ' disabled'}>&larr; Prev</button>` +
				`<span class="muted">Page ${pageNo + 1} of ${pages}</span>` +
				`<button type="button" data-page="${pageNo + 1}"${pageNo < pages - 1 ? '' : ' disabled'}>Next &rarr;</button>`
			: '';
		for (const el of document.querySelectorAll('.pager')) el.innerHTML = pager;
	};

	const update = () => {
		pageNo = 0;
		if (!['q', 'name', 'lv', 'type', 'slot'].includes(sortKey) && !columns().includes(sortKey)) {
			sortKey = 'q'; // the sorted stat's column is gone
			sortDir = -1;
		}
		rows = filter();
		writeState();
		render();
	};

	document.addEventListener('click', e => {
		const s = e.target.closest('[data-sort]');
		if (s) {
			const k = s.dataset.sort;
			if (sortKey === k) sortDir = -sortDir;
			else {
				sortKey = k;
				sortDir = k === 'name' || k === 'type' || k === 'slot' ? 1 : -1;
			}
			writeState();
			render();
			return;
		}
		const p = e.target.closest('[data-page]');
		if (p && !p.disabled) {
			pageNo = +p.dataset.page;
			writeState();
			render();
			$('results-top').scrollIntoView({ block: 'start' });
		}
	});

	fetch('/search.json')
		.then(r => r.json())
		.then(data => {
			D = data;
			form.type.innerHTML = opt('', 'All') + D.types.map(([k, n]) => opt(k, n)).join('');
			const slots = [...new Set(D.items.map(it => it[6]).filter(Boolean))].sort();
			form.slot.innerHTML = opt('', 'All') + slots.map(s => opt(s, s)).join('');
			form.cls.innerHTML = opt('', 'Any') + D.classes.map((c, i) => opt(i, c)).join('');
			form.bind.innerHTML = opt('', 'Any') + Object.entries(BIND).map(([k, v]) => opt(k, v)).join('') + opt('none', 'Not bound');
			form.src.innerHTML = opt('', 'Any') + D.cats.map((c, i) => opt(i, c)).join('');
			$('qualities').innerHTML = QN.map(
				(n, i) => `<label class="chip" style="--qc:${QC[i]}"><input type="checkbox" name="ql" value="${i}"><span>${n}</span></label>`,
			).join('');
			readState();
			rows = filter();
			render();
			form.addEventListener('input', e => {
				if (e.target.name === 'type') {
					fillSections();
					form.sec.value = '';
				}
				clearTimeout(update.t);
				update.t = setTimeout(update, e.target.name === 'q' || e.target.type === 'number' ? 200 : 0);
			});
			form.addEventListener('submit', e => e.preventDefault());
			$('add-stat').addEventListener('click', () => statRow());
			$('reset').addEventListener('click', () => {
				history.replaceState(null, '', location.pathname);
				sortKey = 'q';
				sortDir = -1;
				readState();
				update();
			});
			form.classList.remove('loading');
		})
		.catch(() => {
			$('count').textContent = 'Could not load the item data.';
		});
})();
