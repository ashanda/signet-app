// Port of the original welcome page's scripts (resources/site/js/main.js and
// resources/site/js/live-prices.js), generated from them unchanged except
// that everything they start — timers, animation frames, observers, window
// listeners, ScrollTriggers — goes through `track`, so initSite() can return
// a cleanup function for when the SPA navigates away from the page. The
// original loads gsap + ScrollTrigger from cdnjs as globals; initSite takes
// them as arguments and exposes them the same way before the code runs.
export function initSite({ gsap, ScrollTrigger } = {}) {
  let destroyed = false
  const stops = []
  const triggers = []
  const track = {
    interval(fn, ms) { const id = setInterval(fn, ms); stops.push(() => clearInterval(id)); return id },
    timeout(fn, ms) { const id = setTimeout(fn, ms); stops.push(() => clearTimeout(id)); return id },
    raf(fn) { return requestAnimationFrame((t) => { if (!destroyed) fn(t) }) },
    observer(o) { stops.push(() => o.disconnect()); return o },
    on(target, type, fn, opts) { target.addEventListener(type, fn, opts); stops.push(() => target.removeEventListener(type, fn, opts)) },
    onLoad(fn) { if (document.readyState === 'complete') track.raf(fn); else track.on(window, 'load', fn) },
    trigger(vars) { const t = window.ScrollTrigger.create(vars); triggers.push(t); return t },
  }
  if (gsap) window.gsap = gsap
  if (ScrollTrigger) window.ScrollTrigger = ScrollTrigger

  // ---- resources/site/js/main.js ----
  ;(() => {
  'use strict';

  document.documentElement.classList.remove('no-js');

  const ACCENT = getComputedStyle(document.documentElement).getPropertyValue('--accent').trim() || '#00F0FF';

  /* ---------- nav scroll state ---------- */
  const nav = document.getElementById('site-nav');
  function handleScroll() {
    if (!nav) return;
    nav.classList.toggle('is-scrolled', window.scrollY > 40);
  }
  track.on(window, 'scroll', handleScroll, { passive: true });
  handleScroll();

  /* ---------- mobile menu ---------- */
  const menuToggle = document.getElementById('menu-toggle');
  const menuClose = document.getElementById('menu-close');
  const mobileMenu = document.getElementById('mobile-menu');
  function openMenu() {
    mobileMenu.classList.add('is-open');
    menuToggle.setAttribute('aria-expanded', 'true');
  }
  function closeMenu() {
    mobileMenu.classList.remove('is-open');
    menuToggle.setAttribute('aria-expanded', 'false');
  }
  if (menuToggle && mobileMenu) {
    menuToggle.addEventListener('click', openMenu);
    menuClose.addEventListener('click', closeMenu);
    mobileMenu.querySelectorAll('.js-close-menu').forEach(el => el.addEventListener('click', closeMenu));
    track.on(window.matchMedia('(min-width: 1080px)'), 'change', e => { if (e.matches) closeMenu(); });
  }

  /* ---------- faq accordion ---------- */
  document.querySelectorAll('.faq-item__q').forEach(btn => {
    btn.addEventListener('click', () => {
      const item = btn.closest('.faq-item');
      const panel = item.querySelector('.faq-item__panel');
      const wasOpen = item.classList.contains('is-open');
      const list = item.closest('.faq-list');
      list.querySelectorAll('.faq-item').forEach(i => {
        i.classList.remove('is-open');
        i.querySelector('.faq-item__panel').style.maxHeight = '0px';
      });
      if (!wasOpen) {
        item.classList.add('is-open');
        panel.style.maxHeight = panel.scrollHeight + 'px';
      }
    });
  });

  /* keep an open FAQ panel tall enough when the text reflows (resize / rotate) */
  track.on(window, 'resize', () => {
    document.querySelectorAll('.faq-item.is-open .faq-item__panel').forEach(panel => {
      panel.style.maxHeight = panel.scrollHeight + 'px';
    });
  });

  /* fonts and canvases change layout height after load: re-measure scroll triggers */
  track.onLoad(() => {
    if (window.ScrollTrigger) window.ScrollTrigger.refresh();
  });

  /* ---------- newsletter (front-end only, no backend) ---------- */
  const newsletterForm = document.getElementById('newsletter-form');
  if (newsletterForm) {
    newsletterForm.addEventListener('submit', e => {
      e.preventDefault();
      const btn = newsletterForm.querySelector('button');
      const original = btn.textContent;
      btn.textContent = 'Subscribed ✓';
      btn.disabled = true;
      track.timeout(() => { btn.textContent = original; btn.disabled = false; newsletterForm.reset(); }, 2200);
    });
  }

  /* ---------- canvas helpers ---------- */
  function fitCanvas(cv) {
    const dpr = Math.min(window.devicePixelRatio || 1, 2);
    const r = cv.getBoundingClientRect();
    cv.width = Math.max(1, Math.round(r.width * dpr));
    cv.height = Math.max(1, Math.round(r.height * dpr));
    const ctx = cv.getContext('2d');
    ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
    return { ctx, w: r.width, h: r.height };
  }

  /* ---------- sparklines ---------- */
  function drawSparks() {
    document.querySelectorAll('canvas[data-spark]').forEach(cv => {
      const pts = cv.dataset.spark.split(',').map(Number);
      const up = cv.dataset.up === '1';
      const { ctx, w, h } = fitCanvas(cv);
      const min = Math.min.apply(null, pts), max = Math.max.apply(null, pts);
      const xy = pts.map((v, i) => [(i / (pts.length - 1)) * (w - 2) + 1, h - 6 - ((v - min) / (max - min || 1)) * (h - 14)]);
      const stroke = up ? '#3DE6A8' : '#FF6B81';
      const grad = ctx.createLinearGradient(0, 0, 0, h);
      grad.addColorStop(0, up ? 'rgba(61,230,168,.28)' : 'rgba(255,107,129,.28)');
      grad.addColorStop(1, 'rgba(0,0,0,0)');
      ctx.clearRect(0, 0, w, h);
      ctx.beginPath(); ctx.moveTo(xy[0][0], h);
      xy.forEach(p => ctx.lineTo(p[0], p[1]));
      ctx.lineTo(xy[xy.length - 1][0], h); ctx.closePath();
      ctx.fillStyle = grad; ctx.fill();
      ctx.beginPath(); xy.forEach((p, i) => i ? ctx.lineTo(p[0], p[1]) : ctx.moveTo(p[0], p[1]));
      ctx.strokeStyle = stroke; ctx.lineWidth = 1.6; ctx.lineJoin = 'round'; ctx.stroke();
      const last = xy[xy.length - 1];
      ctx.beginPath(); ctx.arc(last[0], last[1], 2.6, 0, Math.PI * 2);
      ctx.fillStyle = stroke; ctx.shadowColor = stroke; ctx.shadowBlur = 10; ctx.fill(); ctx.shadowBlur = 0;
    });
  }
  track.on(window, 'resize', drawSparks);

  /* ---------- hero blockchain network ---------- */
  function startNetwork() {
    const cv = document.getElementById('hero-canvas');
    if (!cv) return;
    const HEX = '0123456789abcdef';
    const rnd = n => Array.from({ length: n }, () => HEX[(Math.random() * 16) | 0]).join('');
    let dim = fitCanvas(cv);
    let chain = [], height = 128940, spawnAt = 0;
    const SPEED = 0.052;

    const mk = (p) => ({ p, h: ++height, hash: rnd(6), prev: rnd(6), tx: 3 + ((Math.random() * 9) | 0), lit: 0 });
    const build = () => { chain = [0.06, 0.3, 0.54, 0.78, 1.0].map(p => mk(p)); };
    build();
    track.observer(new ResizeObserver(() => { dim = fitCanvas(cv); })).observe(cv);

    const acc = ACCENT;
    const at = (p, w, h) => [w * 0.5 + Math.sin(p * 2.5 + 0.6) * w * 0.22, h * (1.06 - p * 1.12)];

    let prev = performance.now(), pulse = -0.15, stopped = false;
    const tick = (t) => {
      if (stopped) return;
      const dt = Math.min(64, t - prev) / 1000; prev = t;
      const { ctx, w, h } = dim;
      ctx.clearRect(0, 0, w, h);
      const bw = Math.max(58, Math.min(112, w * 0.2)), bh = bw * 0.46, dep = bw * 0.16;

      chain.forEach(b => { b.p += SPEED * dt; });
      chain = chain.filter(b => b.p < 1.22);
      spawnAt -= dt;
      if (spawnAt <= 0) { chain.push(mk(-0.18)); spawnAt = 0.24 / SPEED; }
      chain.sort((x, y) => x.p - y.p);

      for (let i = 0; i < chain.length - 1; i++) {
        const [x1, y1] = at(chain[i].p, w, h), [x2, y2] = at(chain[i + 1].p, w, h);
        const hot = Math.max(chain[i].lit, chain[i + 1].lit);
        ctx.beginPath(); ctx.moveTo(x1, y1 - bh * 0.5); ctx.lineTo(x2, y2 + bh * 0.5);
        ctx.strokeStyle = acc; ctx.globalAlpha = 0.18 + hot * 0.6; ctx.lineWidth = 1.2; ctx.stroke();
        const mx = (x1 + x2) / 2, my = (y1 + y2) / 2;
        ctx.save(); ctx.translate(mx, my); ctx.rotate(Math.PI / 4);
        ctx.globalAlpha = 0.3 + hot * 0.7; ctx.strokeStyle = acc; ctx.lineWidth = 1;
        const k = bw * 0.055; ctx.strokeRect(-k, -k, k * 2, k * 2);
        ctx.restore(); ctx.globalAlpha = 1;
      }

      pulse += dt * 0.19;
      if (pulse > 1.3) pulse = -0.2;
      chain.forEach(b => {
        b.lit = Math.max(0, b.lit - dt * 1.5);
        if (Math.abs(b.p - pulse) < 0.035) b.lit = 1;
      });

      chain.forEach(b => {
        const [x, y] = at(b.p, w, h);
        const edge = Math.min(1, Math.min(b.p + 0.18, 1.2 - b.p) * 4);
        const al = Math.max(0, edge);
        const lit = b.lit;
        ctx.save();
        ctx.globalAlpha = al;
        const L = x - bw / 2, T = y - bh / 2;
        ctx.beginPath();
        ctx.moveTo(L, T); ctx.lineTo(L + dep, T - dep); ctx.lineTo(L + bw + dep, T - dep); ctx.lineTo(L + bw, T);
        ctx.closePath();
        ctx.fillStyle = 'rgba(255,255,255,' + (0.05 + lit * 0.08) + ')'; ctx.fill();
        ctx.strokeStyle = acc; ctx.globalAlpha = al * (0.3 + lit * 0.5); ctx.lineWidth = 1; ctx.stroke();
        ctx.beginPath();
        ctx.moveTo(L + bw, T); ctx.lineTo(L + bw + dep, T - dep); ctx.lineTo(L + bw + dep, T + bh - dep); ctx.lineTo(L + bw, T + bh);
        ctx.closePath();
        ctx.globalAlpha = al; ctx.fillStyle = 'rgba(0,0,0,.35)'; ctx.fill();
        ctx.strokeStyle = acc; ctx.globalAlpha = al * (0.22 + lit * 0.45); ctx.stroke();
        ctx.globalAlpha = al;
        const g = ctx.createLinearGradient(L, T, L + bw, T + bh);
        g.addColorStop(0, 'rgba(0,240,255,' + (0.1 + lit * 0.16) + ')');
        g.addColorStop(1, 'rgba(255,255,255,.02)');
        ctx.fillStyle = g;
        ctx.fillRect(L, T, bw, bh);
        ctx.strokeStyle = acc; ctx.globalAlpha = al * (0.4 + lit * 0.55); ctx.lineWidth = 1.1;
        ctx.strokeRect(L, T, bw, bh);
        if (lit > 0.01) { ctx.shadowColor = acc; ctx.shadowBlur = 22 * lit; ctx.strokeRect(L, T, bw, bh); ctx.shadowBlur = 0; }
        ctx.textAlign = 'left';
        ctx.globalAlpha = al * 0.9;
        ctx.fillStyle = '#EAF6F8';
        ctx.font = '500 ' + Math.round(bw * 0.1) + 'px JetBrains Mono, monospace';
        ctx.fillText('#' + b.h.toLocaleString('en-US'), L + bw * 0.08, T + bh * 0.33);
        ctx.globalAlpha = al * (0.55 + lit * 0.4);
        ctx.fillStyle = lit > 0.3 ? '#8FE9F5' : '#7FA3B0';
        ctx.font = '400 ' + Math.round(bw * 0.085) + 'px JetBrains Mono, monospace';
        ctx.fillText('0x' + b.hash, L + bw * 0.08, T + bh * 0.62);
        ctx.fillText(b.tx + ' tx', L + bw * 0.08, T + bh * 0.88);
        ctx.globalAlpha = al * 0.5;
        for (let i = 0; i < 4; i++) {
          ctx.fillStyle = i <= lit * 4 ? acc : 'rgba(255,255,255,.18)';
          ctx.fillRect(L + bw - bw * 0.14, T + bh * 0.22 + i * bh * 0.17, bw * 0.07, bh * 0.08);
        }
        ctx.restore();
      });

      track.raf(tick);
    };
    track.raf(tick);
  }

  /* ---------- about ledger grid ---------- */
  function startOrbit() {
    const cv = document.getElementById('orbit-canvas');
    if (!cv) return;
    let dim = fitCanvas(cv);
    track.observer(new ResizeObserver(() => { dim = fitCanvas(cv); })).observe(cv);
    const acc = ACCENT;
    let cells = [], cols = 0, rows = 0, gs = 0, ox = 0, oy = 0, path = [];

    const build = () => {
      const { w, h } = dim;
      gs = Math.max(26, Math.min(44, w / 12));
      cols = Math.max(4, Math.floor((w - 28) / gs));
      rows = Math.max(4, Math.floor((h - 28) / gs));
      ox = (w - cols * gs) / 2; oy = (h - rows * gs) / 2;
      cells = [];
      for (let r = 0; r < rows; r++) for (let c = 0; c < cols; c++) {
        cells.push({ c, r, lit: 0, seed: Math.random() });
      }
      path = [];
      let c0 = 0, r0 = (rows / 2) | 0;
      while (c0 < cols - 1) {
        path.push({ c: c0, r: r0 });
        if (Math.random() < 0.45) r0 = Math.max(0, Math.min(rows - 1, r0 + (Math.random() < .5 ? -1 : 1)));
        else c0++;
      }
      path.push({ c: cols - 1, r: r0 });
    };
    build();
    track.observer(new ResizeObserver(() => build())).observe(cv);

    const px = (c) => ox + c * gs + gs / 2;
    const py = (r) => oy + r * gs + gs / 2;
    let prev = performance.now(), wave = 0, head = 0;

    const tick = (t) => {
      const dt = Math.min(64, t - prev) / 1000; prev = t;
      const { ctx, w, h } = dim;
      const cx = w / 2, cy = h / 2, hole = Math.min(w, h) * 0.17;
      ctx.clearRect(0, 0, w, h);

      wave = (wave + dt * 0.16) % 1.6;
      head = (head + dt * 2.6) % (path.length + 6);

      cells.forEach(cell => {
        const x = px(cell.c), y = py(cell.r);
        const diag = (cell.c / cols + cell.r / rows) / 2;
        const d = Math.abs(diag - (wave - 0.3));
        const pulseAmt = d < 0.09 ? 1 - d / 0.09 : 0;
        cell.lit = Math.max(cell.lit - dt * 1.2, pulseAmt);
        const dist = Math.hypot(x - cx, y - cy);
        const mask = Math.min(1, Math.max(0, (dist - hole) / hole));
        const s2 = gs * 0.3;
        ctx.globalAlpha = (0.07 + cell.lit * 0.75) * mask;
        ctx.strokeStyle = cell.lit > 0.25 ? acc : 'rgba(255,255,255,.5)';
        ctx.lineWidth = 1;
        ctx.strokeRect(x - s2 / 2, y - s2 / 2, s2, s2);
        if (cell.lit > 0.35) {
          ctx.globalAlpha = cell.lit * 0.5 * mask;
          ctx.fillStyle = acc;
          ctx.fillRect(x - s2 / 2, y - s2 / 2, s2, s2);
        }
      });
      ctx.globalAlpha = 1;

      ctx.beginPath();
      path.forEach((p, i) => { const x = px(p.c), y = py(p.r); i ? ctx.lineTo(x, y) : ctx.moveTo(x, y); });
      ctx.strokeStyle = acc; ctx.globalAlpha = .13; ctx.lineWidth = 1.4; ctx.stroke();

      const hi = Math.min(path.length - 1, Math.floor(head));
      const frac = head - hi;
      ctx.beginPath();
      for (let i = 0; i <= hi; i++) { const x = px(path[i].c), y = py(path[i].r); i ? ctx.lineTo(x, y) : ctx.moveTo(x, y); }
      ctx.strokeStyle = acc; ctx.globalAlpha = .55; ctx.lineWidth = 1.6;
      ctx.shadowColor = acc; ctx.shadowBlur = 10; ctx.stroke(); ctx.shadowBlur = 0;
      ctx.globalAlpha = 1;

      if (hi < path.length - 1) {
        const p0 = path[hi], p1 = path[hi + 1];
        const hx = px(p0.c) + (px(p1.c) - px(p0.c)) * frac;
        const hy = py(p0.r) + (py(p1.r) - py(p0.r)) * frac;
        const g = ctx.createRadialGradient(hx, hy, 0, hx, hy, gs * 0.9);
        g.addColorStop(0, acc); g.addColorStop(1, 'rgba(0,0,0,0)');
        ctx.fillStyle = g; ctx.globalAlpha = .85;
        ctx.beginPath(); ctx.arc(hx, hy, gs * 0.9, 0, Math.PI * 2); ctx.fill();
        ctx.globalAlpha = 1;
        ctx.save(); ctx.translate(px(p0.c), py(p0.r)); ctx.rotate(Math.PI / 4);
        ctx.strokeStyle = '#EAF6F8'; ctx.globalAlpha = .8; ctx.lineWidth = 1.2;
        const k = gs * 0.2; ctx.strokeRect(-k, -k, k * 2, k * 2);
        ctx.restore(); ctx.globalAlpha = 1;
      }

      track.raf(tick);
    };
    track.raf(tick);
  }

  /* ---------- candlesticks ---------- */
  function makeCandles(n) {
    let p = 66200; const out = [];
    for (let i = 0; i < n; i++) {
      const o = p;
      const c = o + (Math.random() - .45) * 520;
      const hi = Math.max(o, c) + Math.random() * 240;
      const lo = Math.min(o, c) - Math.random() * 240;
      out.push({ o, c, hi, lo }); p = c;
    }
    return out;
  }

  function startCandles() {
    const cv = document.getElementById('candle-canvas');
    if (!cv) return;
    let candles = makeCandles(46);

    function paint() {
      const { ctx, w, h } = fitCanvas(cv);
      const hi = Math.max.apply(null, candles.map(c => c.hi)), lo = Math.min.apply(null, candles.map(c => c.lo));
      const pad = 12, ih = h - pad * 2;
      const y = v => pad + ih - ((v - lo) / (hi - lo || 1)) * ih;
      ctx.clearRect(0, 0, w, h);
      ctx.strokeStyle = 'rgba(255,255,255,.05)'; ctx.lineWidth = 1;
      for (let i = 0; i <= 4; i++) {
        const gy = Math.round(pad + (ih / 4) * i) + .5;
        ctx.beginPath(); ctx.moveTo(0, gy); ctx.lineTo(w, gy); ctx.stroke();
      }
      const step = w / candles.length, bw = Math.max(2, Math.min(9, step * .56));
      candles.forEach((c, i) => {
        const cx = i * step + step / 2;
        const up = c.c >= c.o;
        const col = up ? '#3DE6A8' : '#FF6B81';
        ctx.strokeStyle = col; ctx.fillStyle = col; ctx.globalAlpha = i > candles.length - 4 ? 1 : .85;
        ctx.beginPath(); ctx.moveTo(cx, y(c.hi)); ctx.lineTo(cx, y(c.lo)); ctx.lineWidth = 1; ctx.stroke();
        const top = y(Math.max(c.o, c.c)), bot = y(Math.min(c.o, c.c));
        ctx.fillRect(cx - bw / 2, top, bw, Math.max(1.5, bot - top));
        ctx.globalAlpha = 1;
      });
      const lastY = y(candles[candles.length - 1].c);
      ctx.setLineDash([4, 5]); ctx.strokeStyle = ACCENT; ctx.globalAlpha = .6;
      ctx.beginPath(); ctx.moveTo(0, lastY); ctx.lineTo(w, lastY); ctx.stroke();
      ctx.setLineDash([]); ctx.globalAlpha = 1;
    }

    track.observer(new ResizeObserver(paint)).observe(cv);
    paint();
    track.interval(() => {
      if (document.hidden) return;
      const last = candles[candles.length - 1];
      const c = last.c + (Math.random() - .5) * 300;
      candles.push({ o: last.c, c, hi: Math.max(last.c, c) + Math.random() * 160, lo: Math.min(last.c, c) - Math.random() * 160 });
      candles.shift();
      paint();
    }, 2200);
  }

  /* ---------- gsap entrances / reveals / counters / lines ---------- */
  function initGsap() {
    const g = window.gsap;
    if (!g) { fallbackReveal(); return; }
    if (window.ScrollTrigger) g.registerPlugin(window.ScrollTrigger);
    const q = sel => Array.from(document.querySelectorAll(sel));

    const hero = q('[data-hero]');
    g.set(hero, { opacity: 0, y: 34 });
    g.to(hero, { opacity: 1, y: 0, duration: .9, ease: 'power3.out', stagger: .12, delay: .12 });

    if (!window.ScrollTrigger) { q('[data-reveal]').forEach(el => el.classList.add('is-visible')); return; }

    q('[data-reveal]').forEach(el => {
      g.set(el, { opacity: 0, y: 40 });
      track.trigger({
        trigger: el, start: 'top 88%', once: true,
        onEnter: () => g.to(el, { opacity: 1, y: 0, duration: .8, ease: 'power3.out', delay: (Number(el.dataset.reveal) - 1) * .09 })
      });
    });

    q('[data-parallax]').forEach(el => {
      track.trigger({
        trigger: el, start: 'top bottom', end: 'bottom top', scrub: true,
        onUpdate: s => g.set(el, { y: (s.progress - .5) * 2 * Number(el.dataset.parallax) * 2 })
      });
    });

    q('[data-count]').forEach(el => {
      const target = Number(el.dataset.count), dec = Number(el.dataset.dec || 0);
      track.trigger({
        trigger: el, start: 'top 92%', once: true,
        onEnter: () => {
          const o = { v: 0 };
          g.to(o, { v: target, duration: 1.8, ease: 'power2.out', onUpdate: () => {
            el.textContent = dec ? o.v.toFixed(dec) : Math.round(o.v).toLocaleString('en-US');
          } });
        }
      });
    });

    const stepLine = document.getElementById('step-line-fill');
    if (stepLine) {
      track.trigger({
        trigger: stepLine.parentElement, start: 'top 78%', end: 'bottom 60%', scrub: .6,
        onUpdate: s => g.set(stepLine, { scaleX: s.progress })
      });
    }
    const roadLine = document.getElementById('roadmap-line-fill');
    if (roadLine) {
      track.trigger({
        trigger: roadLine.parentElement, start: 'top 72%', end: 'bottom 70%', scrub: .6,
        onUpdate: s => g.set(roadLine, { scaleY: s.progress })
      });
    }
  }

  function fallbackReveal() {
    document.querySelectorAll('[data-hero],[data-reveal]').forEach(el => { el.style.opacity = '1'; el.style.transform = 'none'; });
    const io = track.observer(new IntersectionObserver(es => es.forEach(e => {
      if (!e.isIntersecting) return;
      const el = e.target, target = Number(el.dataset.count), dec = Number(el.dataset.dec || 0);
      el.textContent = dec ? target.toFixed(dec) : target.toLocaleString('en-US');
      io.unobserve(el);
    }), { threshold: .4 }));
    document.querySelectorAll('[data-count]').forEach(el => io.observe(el));
  }

  /* ---------- boot ---------- */
  track.raf(() => {
    drawSparks();
    startNetwork();
    startOrbit();
    startCandles();
    initGsap();
  });
})();

  // ---- resources/site/js/live-prices.js ----
  (() => {
  'use strict';

  const API = 'https://api.coingecko.com/api/v3/simple/price?ids=' +
    ['bitcoin', 'ethereum', 'solana', 'ripple', 'cardano', 'avalanche-2', 'chainlink', 'polkadot'].join(',') +
    '&vs_currencies=usd&include_24hr_change=true';

  const fmtPrice = p => {
    if (p >= 1000) return '$' + p.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
    if (p >= 1) return '$' + p.toFixed(2);
    return '$' + p.toFixed(4);
  };

  const fmtChange = (c, style) => {
    const abs = Math.abs(c).toFixed(2);
    return style === 'arrow' ? (c >= 0 ? '▲ ' : '▼ ') + abs + '%' : (c >= 0 ? '+' : '−') + abs + '%';
  };

  async function refresh() {
    if (document.hidden) return;
    try {
      const res = await fetch(API, { headers: { Accept: 'application/json' } });
      if (!res.ok) return;
      const data = await res.json();
      let updated = false;

      document.querySelectorAll('[data-coin]').forEach(el => {
        const row = data[el.dataset.coin];
        if (!row || typeof row.usd !== 'number') return;

        const price = el.querySelector('.js-price');
        if (price) price.textContent = fmtPrice(row.usd);

        const pct = el.querySelector('.js-pct');
        const change = row.usd_24h_change;
        if (pct && typeof change === 'number') {
          const up = change >= 0;
          pct.textContent = fmtChange(change, pct.dataset.fmt);
          const [upCls, downCls] = pct.classList.contains('pct') ? ['pct--up', 'pct--down'] : ['up', 'down'];
          pct.classList.toggle(upCls, up);
          pct.classList.toggle(downCls, !up);
          const spark = el.querySelector('canvas[data-spark]');
          if (spark) {
            const pts = spark.dataset.spark.split(',').map(Number);
            if ((pts[pts.length - 1] >= pts[0]) !== up) spark.dataset.spark = pts.reverse().join(',');
            spark.dataset.up = up ? '1' : '0';
          }
        }
        updated = true;
      });

      if (updated) {
        const note = document.getElementById('market-note');
        if (note) note.textContent = 'Prices and 24h change are live from CoinGecko. Sparkline shapes are illustrative.';
        window.dispatchEvent(new Event('resize'));
      }
    } catch (e) {
      /* network or rate-limit failure: keep the sample values already on the page */
    }
  }

  refresh();
  track.interval(refresh, 60000);
})();

  return function cleanup() {
    destroyed = true
    stops.forEach((f) => f())
    triggers.forEach((t) => t.kill())
    if (window.gsap) window.gsap.globalTimeline.clear()
  }
}
