// db.lokiy.dev: search and hover tooltips. Data: /items.json [{id, n: name, q: quality, i: small icon URL, s?: sources,
// c?: 1 if new, r?: 1 if reworked}], hover tooltips in /tt/<id // 500>.json {id: html}
(() => {
	const QC = ['#9d9d9d', '#ffffff', '#1eff00', '#0070dd', '#a335ee', '#ff8000', '#e6cc80'];
	const ICON = icon => `https://wow.zamimg.com/images/wow/icons/small/${icon}.jpg`;
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

	// ---- search (index page shows results inline; other pages submit to /?q=)
	const input = document.querySelector('.search input');
	const results = document.getElementById('results');
	const list = document.getElementById('result-list');
	const browse = document.getElementById('browse');
	const runSearch = async q => {
		if (!results) return;
		q = q.trim().toLowerCase();
		if (!q) {
			results.hidden = true;
			browse.hidden = false;
			return;
		}
		const data = await loadData();
		const hits = [...data.values()]
			.filter(it => it.n.toLowerCase().includes(q) || String(it.id) === q || (it.s || []).some(s => s.toLowerCase().includes(q)))
			.sort((a, b) => (b.n.toLowerCase().startsWith(q) - a.n.toLowerCase().startsWith(q)) || b.q - a.q || a.n.localeCompare(b.n))
			.slice(0, 200);
		list.innerHTML = hits.length
			? hits
					.map(
						it =>
							`<li><a class="item" href="/item/${it.id}" data-item="${it.id}"><img class="icon-s" src="${it.i}" alt="" onerror="this.onerror=null;this.src='${ICON('inv_misc_questionmark')}'"><span style="color:${QC[it.q] || '#fff'}">${escapeHtml(it.n)}</span></a>` +
							(it.c ? '<span class="tag">new</span>' : it.r ? '<span class="tag">changed</span>' : '') +
								(it.s ? `<span class="muted">${escapeHtml(it.s.slice(0, 3).join(', '))}</span>` : '') +
							'</li>',
					)
					.join('')
			: '<li class="muted">No items found.</li>';
		results.hidden = false;
		browse.hidden = true;
	};
	if (input && results) {
		const q = new URLSearchParams(location.search).get('q') || '';
		input.value = q;
		if (q) runSearch(q);
		input.addEventListener('input', () => {
			runSearch(input.value);
			const url = new URL(location.href);
			input.value ? url.searchParams.set('q', input.value) : url.searchParams.delete('q');
			history.replaceState(null, '', url);
		});
		input.form.addEventListener('submit', e => e.preventDefault());
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
