// Dependency-free Chrome/CDP regression tests, using the site's existing approach.
import assert from 'node:assert/strict';
import { spawn } from 'node:child_process';
import { createServer } from 'node:http';
import { mkdtemp, readFile, rm, writeFile } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { once } from 'node:events';

const chrome = process.env.CHROME_BIN || (process.platform === 'darwin'
  ? '/Applications/Google Chrome.app/Contents/MacOS/Google Chrome' : '/usr/bin/google-chrome');
const profile = await mkdtemp(join(tmpdir(), 'burn-browser-'));
const html = await readFile(new URL('../burn.html', import.meta.url));
const fixture = {
  generatedAt: '2026-10-04T00:00:00Z', totalTokens: 1000000,
  byModelTop: [{short: 'demo', tokens: 1000000}],
  last30: [
    {date: '2026-10-01', tokens: 100000, commits: 21, topRepos: ['public-demo (21)'], topMsgs: ['Ship keyboard navigation']},
    {date: '2026-10-02', tokens: 800000, commits: 1, topRepos: ['public-demo (1)'], topMsgs: ['Ship refresh retention']},
    {date: '2026-10-03', tokens: 100000, commits: 25, topRepos: ['public-demo (25)'], topMsgs: ['Ship dated notes']}
  ]
};
const server = createServer((req, res) => {
  res.writeHead(200, {'Content-Type': 'text/html; charset=utf-8'});
  res.end(html);
}).listen(0, '127.0.0.1');
await once(server, 'listening');
const url = `http://127.0.0.1:${server.address().port}/burn.html`;
const child = spawn(chrome, ['--headless=new', '--remote-debugging-port=0',
  `--user-data-dir=${profile}`, '--disable-dev-shm-usage', 'about:blank'], {stdio: 'ignore'});
const sleep = ms => new Promise(resolve => setTimeout(resolve, ms));
let ws;
try {
  let port;
  for (let i = 0; i < 100; i++) {
    try { port = (await readFile(join(profile, 'DevToolsActivePort'), 'utf8')).split('\n')[0]; break; }
    catch { await sleep(100); }
  }
  assert.ok(port, 'Chrome must start');
  console.log(`Owned server PID ${process.pid} port ${server.address().port}; Chrome PID ${child.pid} CDP port ${port}`);
  const targets = await (await fetch(`http://127.0.0.1:${port}/json`)).json();
  ws = new WebSocket(targets.find(t => t.type === 'page').webSocketDebuggerUrl);
  await new Promise(resolve => ws.addEventListener('open', resolve, {once: true}));
  let id = 0;
  const pending = new Map(), errors = [];
  ws.onmessage = event => {
    const message = JSON.parse(event.data);
    if (message.id && pending.has(message.id)) {
      const {resolve, reject, timer} = pending.get(message.id);
      clearTimeout(timer); pending.delete(message.id);
      if (message.error) reject(new Error(JSON.stringify(message.error)));
      else resolve(message.result);
    }
    if (message.method === 'Runtime.exceptionThrown') errors.push(message.params.exceptionDetails);
    if (message.method === 'Runtime.consoleAPICalled' && message.params.type === 'error') errors.push(message.params.args);
  };
  const send = (method, params = {}) => new Promise((resolve, reject) => {
    const n = ++id;
    const timer = setTimeout(() => { pending.delete(n); reject(new Error(`CDP timeout: ${method}`)); }, 15000);
    pending.set(n, {resolve, reject, timer});
    ws.send(JSON.stringify({id: n, method, params}));
  });
  const evaluate = async expression => {
    const response = await send('Runtime.evaluate', {expression, returnByValue: true, awaitPromise: true});
    assert.ok(!response.exceptionDetails, JSON.stringify(response.exceptionDetails));
    return response.result.value;
  };
  const key = async (key, code, number) => {
    await send('Input.dispatchKeyEvent', {type: 'keyDown', key, code, windowsVirtualKeyCode: number, ...(key === 'Enter' ? {text:'\r'} : {})});
    await send('Input.dispatchKeyEvent', {type: 'keyUp', key, code, windowsVirtualKeyCode: number});
  };
  const refresh = async payload => {
    await evaluate(`window.__payload=${JSON.stringify(payload)}; window.__refresh()`);
  };
  const state = () => evaluate(`({date:document.activeElement.dataset.date || null,
    expanded:[...document.querySelectorAll('.ship-day[aria-expanded="true"]')].map(b=>b.dataset.date),
    note:document.getElementById('note').innerText})`);
  const load = async (width, nojs = false) => {
    await send('Emulation.setDeviceMetricsOverride', {width, height: 900, deviceScaleFactor: 1, mobile: width < 500});
    await send('Emulation.setScriptExecutionDisabled', {value: nojs});
    await send('Page.navigate', {url});
    for (let i = 0; i < 100; i++) {
      if (await evaluate(`document.readyState==='complete' && !!document.querySelector('${nojs ? 'noscript h1' : '.ship-day'}')`)) return;
      await sleep(100);
    }
    throw new Error('Burn page did not load');
  };
  await send('Runtime.enable');
  await send('Page.enable');
  // Test-only seams replace network input and capture the actual scheduled refresh.
  // No source data is written and the production five-minute delay is never awaited.
  await send('Page.addScriptToEvaluateOnNewDocument', {source: `
    window.__payload=${JSON.stringify(fixture)};
    window.fetch=async()=>({ok:true,json:async()=>window.__payload});
    const interval=window.setInterval.bind(window);
    window.setInterval=(fn,ms)=>{if(ms===300000){window.__refresh=fn;return 0;}return interval(fn,ms);};
    window.__frames=0;
    const raf=window.requestAnimationFrame.bind(window);
    window.requestAnimationFrame=fn=>raf(t=>{window.__frames++;fn(t);});
  `});
  for (const width of [1440, 390]) {
    await load(width);
    // Start at the back link; native Tab reaches the first dated note control.
    await evaluate('document.querySelector(".crumb a").focus()');
    await key('Tab', 'Tab', 9);
    assert.equal((await state()).date, '2026-10-01');
    assert.equal(await evaluate('document.activeElement.matches(":focus-visible") && getComputedStyle(document.activeElement).outlineStyle === "solid"'), true);
    await key('Enter', 'Enter', 13);
    const opened = await state();
    assert.deepEqual(opened.expanded, ['2026-10-01']);
    assert.ok(opened.note.includes('Ship keyboard navigation'));
    assert.ok(!opened.note.includes('Ship dated notes'));
    const ax = (await send('Accessibility.getFullAXTree')).nodes;
    const button = ax.find(n => n.role?.value === 'button' && n.name?.value === '2026-10-01 · 21 commits');
    assert.ok(button, 'Browser accessibility tree exposes date and count');
    assert.equal(button.properties.find(p => p.name === 'expanded').value.value, true);
    assert.ok(ax.some(n => n.role?.value === 'status'));
    assert.ok(ax.some(n => n.role?.value === 'heading' && n.name?.value === 'agents at work'));
    await evaluate('window.__noteNode=document.getElementById("note"); window.__noteChild=window.__noteNode.firstChild');
    await refresh(fixture);
    assert.deepEqual(await state(), opened, 'Unchanged refresh preserves selection, note and focus');
    assert.equal(await evaluate('window.__noteNode===document.getElementById("note") && window.__noteChild===window.__noteNode.firstChild'), true, 'Unchanged live region content is retained');
    await key(' ', 'Space', 32);
    assert.deepEqual((await state()).expanded, []);
    assert.equal((await state()).note, '', 'Space closes the open note');
    await key(' ', 'Space', 32);
    assert.deepEqual((await state()).expanded, ['2026-10-01'], 'Space opens the correct date');
    await key('Enter', 'Enter', 13);
    assert.equal((await state()).note, '', 'Enter closes the open note');
    await key('Enter', 'Enter', 13);
    const changed = structuredClone(fixture);
    changed.last30 = [changed.last30[2], changed.last30[1], {...changed.last30[0], commits: 2, topMsgs: ['Updated same dated note']}];
    await refresh(changed);
    assert.equal((await state()).date, '2026-10-01');
    assert.deepEqual((await state()).expanded, ['2026-10-01']);
    assert.ok((await state()).note.includes('Updated same dated note'));
    assert.ok((await state()).note.includes('2 commits'));
    assert.equal(await evaluate('document.activeElement.getAttribute("aria-label")'), '2026-10-01 · 2 commits');
    assert.equal(await evaluate('document.activeElement.matches(":focus-visible")'), true);
    // A refresh must not steal focus from someone who has left the controls.
    await evaluate('document.querySelector(".crumb a").focus()');
    await refresh(changed);
    assert.equal(await evaluate('document.activeElement.matches(".crumb a")'), true);
    await evaluate(`document.querySelector('.ship-day[data-date="2026-10-01"]').focus()`);
    const removed = structuredClone(changed);
    removed.last30 = removed.last30.filter(d => d.date !== '2026-10-01');
    await refresh(removed);
    assert.deepEqual((await state()).expanded, []);
    assert.equal((await state()).note, 'The note for 2026-10-01 is no longer available.');
    assert.equal(await evaluate('document.activeElement.id'), 'ship-heading');
    await refresh(removed);
    assert.equal(await evaluate('document.activeElement.id'), 'ship-heading', 'Recovery heading keeps focus on a subsequent unchanged refresh');
    await key('Tab', 'Tab', 9);
    await key('Enter', 'Enter', 13);
    assert.ok((await state()).note.includes('Ship dated notes'), 'First activation after removal opens the remaining correct note');
    // Dense 30-day data exercises actual CSS-pixel targets; SVG scaling is irrelevant.
    const dense = {...fixture, last30: Array.from({length:30}, (_, i) => ({date:`2026-09-${String(i+1).padStart(2,'0')}`, tokens:i+1, commits:20, topMsgs:[`Public demo day ${i+1}`]}))};
    await refresh(dense);
    assert.equal(await evaluate('document.querySelectorAll(".ship-day").length'), 30);
    const dates = dense.last30.map(d => d.date);
    for (const date of dates) {
      const rect = await evaluate(`(() => {const b=[...document.querySelectorAll('.ship-day')].find(b=>b.dataset.date===${JSON.stringify(date)});b.scrollIntoView({block:'center',behavior:'instant'});const r=b.getBoundingClientRect();return {x:r.x,y:r.y,width:r.width,height:r.height,hit:document.elementFromPoint(r.x+r.width/2,r.y+r.height/2)===b};})()`);
      assert.ok(rect.width >= 24 && rect.height >= 24 && rect.hit, `${width}px actual hit target ${date}`);
      await send('Input.dispatchMouseEvent', {type:'mousePressed', x:rect.x+rect.width/2, y:rect.y+rect.height/2, button:'left', clickCount:1});
      await send('Input.dispatchMouseEvent', {type:'mouseReleased', x:rect.x+rect.width/2, y:rect.y+rect.height/2, button:'left', clickCount:1});
      assert.ok((await state()).note.includes(`Public demo day ${Number(date.slice(-2))}`), 'Coordinate hit opens correct date');
      assert.deepEqual((await state()).expanded, [date]);
    }
    assert.ok(await evaluate('document.documentElement.scrollWidth<=innerWidth'), `${width}px no horizontal overflow`);
    const rects = await evaluate('[...document.querySelectorAll(".ship-day")].map(b=>{const r=b.getBoundingClientRect();return {x:r.x,y:r.y,right:r.right,bottom:r.bottom};})');
    for (let i=0;i<rects.length;i++) for (let j=i+1;j<rects.length;j++) {
      const a=rects[i], b=rects[j];
      assert.ok(a.right<=b.x || b.right<=a.x || a.bottom<=b.y || b.bottom<=a.y, 'Date controls never overlap');
    }
    await refresh({...fixture,last30:[{date:'2026-10-04',tokens:10,commits:0,topMsgs:[]}]});
    await evaluate('document.querySelector(".ship-day").focus()');
    await key('Enter', 'Enter', 13);
    assert.ok((await state()).note.includes('quiet day'));
    const empty = {...fixture,last30:[]};
    await refresh(empty);
    assert.deepEqual((await state()).expanded, []);
    assert.equal(await evaluate('document.activeElement.id'), 'ship-heading');
    await refresh(empty);
    assert.equal(await evaluate('document.activeElement.id'), 'ship-heading', 'Empty-data recovery focus survives another refresh');
    await sleep(300);
    assert.equal(await evaluate('window.__frames'), 0, 'No continuous repaint callbacks');
    assert.equal(await evaluate('document.getAnimations().some(a=>a.playState==="running")'), false, 'Only finite entrance animations');
    if (process.env.SCREENSHOT_DIR) {
      await refresh(fixture);
      await evaluate('document.querySelector(".ship-day").focus()');
      await key('Enter', 'Enter', 13);
      await sleep(300);
      const image = await send('Page.captureScreenshot', {format:'png'});
      await writeFile(join(process.env.SCREENSHOT_DIR, `burn-${width}.png`), Buffer.from(image.data,'base64'));
    }
    console.log(`PASS ${width}px keyboard toggle, accessible names/state, unchanged/changed/reordered/removed refresh, 30 non-overlapping actual pointer targets, no continuous motion`);
  }
  await load(390, true);
  assert.ok((await evaluate('document.body.innerText')).includes('The burn log and ship notes need JavaScript.'));
  assert.equal(await evaluate('document.querySelector("noscript a").getAttribute("href")'), 'burn.md');
  assert.ok(await evaluate('document.documentElement.scrollWidth<=innerWidth'));
  assert.ok((await readFile(new URL('../burn.md',import.meta.url),'utf8')).toLowerCase().includes('token'));
  console.log('PASS readable JavaScript-disabled snapshot link');
  assert.deepEqual(errors, [], 'No runtime or console errors');
} finally {
  ws?.close();
  child.kill();
  if (child.exitCode===null) await once(child,'exit');
  await new Promise(resolve=>server.close(resolve));
  await rm(profile,{recursive:true,force:true,maxRetries:3,retryDelay:100});
  console.log('Owned browser/server stopped and isolated profile removed');
}
