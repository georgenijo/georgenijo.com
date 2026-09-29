// Dependency-free Chrome/CDP smoke tests; Chrome is available on GitHub runners.
import assert from 'node:assert/strict';
import { spawn } from 'node:child_process';
import { createServer } from 'node:http';
import { mkdtemp, readFile, rm, writeFile } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { once } from 'node:events';

const chrome = process.env.CHROME_BIN || (process.platform === 'darwin'
  ? '/Applications/Google Chrome.app/Contents/MacOS/Google Chrome' : '/usr/bin/google-chrome');
const profile = await mkdtemp(join(tmpdir(), 'homepage-browser-'));
const html = await readFile(new URL('../index.html', import.meta.url));
const server = createServer((req, res) => {
  res.writeHead(200, {'Content-Type': 'text/html; charset=utf-8'});
  res.end(html);
}).listen(0, '127.0.0.1');
await once(server, 'listening');
const url = process.env.HOMEPAGE_URL || `http://127.0.0.1:${server.address().port}/`;
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
  const load = async (width, mobile, reduce, nojs = false) => {
    await send('Emulation.setDeviceMetricsOverride', {width, height: 900, deviceScaleFactor: 1, mobile});
    await send('Emulation.setEmulatedMedia', {features: [{name: 'prefers-reduced-motion', value: reduce ? 'reduce' : 'no-preference'}]});
    await send('Emulation.setScriptExecutionDisabled', {value: nojs});
    await send('Page.navigate', {url});
    for (let i = 0; i < 100; i++) {
      if (await evaluate('document.readyState === "complete" && !!document.querySelector("#motion-toggle")')) return;
      await sleep(100);
    }
    throw new Error('Homepage did not finish loading');
  };
  await send('Runtime.enable');
  await send('Page.enable');
  // Instrument executed native animation callbacks, not merely scheduled callbacks.
  await send('Page.addScriptToEvaluateOnNewDocument', {source: `
    window.__frames = 0;
    const nativeRAF = window.requestAnimationFrame.bind(window);
    window.requestAnimationFrame = fn => nativeRAF(t => { window.__frames++; fn(t); });
  `});
  for (const width of [1440, 390]) {
    await load(width, width < 500, false);
    assert.equal(await evaluate('document.querySelectorAll("#bars .b").length'), 21);
    assert.equal(await evaluate('document.querySelectorAll("#pipe .node").length'), 9);
    for (const section of ['top', 'inflection', 'wall', 'factory', 'routing', 'numbers', 'stack', 'timeline']) {
      await evaluate(`document.getElementById('${section}').scrollIntoView({behavior:'instant'})`);
      await sleep(100);
      assert.ok(await evaluate('document.documentElement.scrollWidth <= innerWidth'), `${width}px overflow at ${section}`);
    }
    // Keyboard tabs change both focus and associated panel.
    await evaluate('document.querySelector("#layer-2").focus()');
    await send('Input.dispatchKeyEvent', {type: 'keyDown', key: 'ArrowDown', code: 'ArrowDown'});
    await send('Input.dispatchKeyEvent', {type: 'keyUp', key: 'ArrowDown', code: 'ArrowDown'});
    assert.equal(await evaluate('document.activeElement.id'), 'layer-1');
    assert.equal(await evaluate('document.querySelector("#detail").getAttribute("aria-labelledby")'), 'layer-1');
    await evaluate('document.querySelector("#layer-4").click()');
    assert.ok((await evaluate('document.querySelector("#detail").innerText')).includes('murmur-app'));
    await evaluate('document.querySelector("#top").scrollIntoView({behavior:"instant"}); document.querySelector("#motion-toggle").click()');
    await sleep(100);
    const paused = await evaluate('[window.__frames, document.querySelector("#ticker").innerHTML, document.querySelector("#flow").toDataURL()]');
    await sleep(1700);
    assert.deepEqual(await evaluate('[window.__frames, document.querySelector("#ticker").innerHTML, document.querySelector("#flow").toDataURL()]'), paused, 'Pause stops frames, hero and ticker');
    await evaluate('document.querySelector("#motion-toggle").click()');
    await sleep(200);
    assert.ok(await evaluate(`window.__frames > ${paused[0]}`), 'Resume restarts animation');
    if (process.env.SCREENSHOT_DIR) {
      const image = await send('Page.captureScreenshot', {format: 'png'});
      await writeFile(join(process.env.SCREENSHOT_DIR, `homepage-${width}.png`), Buffer.from(image.data, 'base64'));
    }
    console.log(`PASS ${width}px layout, controls, pause/resume`);
  }
  await load(390, true, true);
  await sleep(300);
  assert.equal(await evaluate('window.__frames'), 0, 'Reduced motion must not run a hidden animation loop');
  assert.equal(await evaluate('document.querySelector("#motion-toggle").getAttribute("aria-pressed")'), 'true');
  await evaluate('document.getElementById("factory").scrollIntoView({behavior:"instant"})');
  await sleep(300);
  assert.equal(await evaluate('window.__frames'), 0);
  assert.equal(await evaluate('document.querySelectorAll("#bars-table tbody tr").length'), 21);
  console.log('PASS reduced motion without native animation callbacks');
  await load(390, true, false, true);
  assert.ok((await evaluate('document.body.innerText')).includes('6,300'));
  assert.equal(await evaluate('document.querySelectorAll("#bars-table tbody tr").length'), 21);
  assert.equal(await evaluate('document.querySelectorAll("#tokens-table tbody tr").length'), 5);
  assert.equal(await evaluate('document.querySelectorAll("#rules article").length'), 6);
  assert.equal(await evaluate('getComputedStyle(document.querySelector("#factory .shead")).opacity'), '1');
  console.log('PASS JavaScript-disabled content and exact data tables');
  assert.deepEqual(errors, [], 'No runtime or console errors');
} finally {
  ws?.close();
  child.kill();
  if (child.exitCode === null) await once(child, 'exit');
  server.close();
  await rm(profile, {recursive: true, force: true, maxRetries: 3, retryDelay: 100});
}
