import { spawn } from 'node:child_process';
import { once } from 'node:events';
import { readFile } from 'node:fs/promises';
import { join } from 'node:path';
import { setTimeout as delay } from 'node:timers/promises';

const base = process.env.CYPRESS_BASE_URL;
const profile = process.argv[2];
const repetitions = Number(process.env.BASELINE_REPETITIONS ?? 3);
if (!Number.isInteger(repetitions) || repetitions < 1) throw new Error('Expected a positive BASELINE_REPETITIONS');

const chrome = spawn(process.env.CHROME_BIN || 'google-chrome', [
  '--headless=new', '--disable-background-networking', '--disable-dev-shm-usage',
  '--disable-component-update', '--disable-sync',
  '--no-first-run', '--window-size=1920,1080', '--remote-debugging-port=0',
  `--user-data-dir=${profile}`, 'about:blank',
], { stdio: ['ignore', 'ignore', 'inherit'] });
let launchError;
chrome.on('error', error => { launchError = error; });

async function until(read, label) {
  for (let attempt = 0; attempt < 200; attempt++) {
    const value = await read();
    if (value) return value;
    await delay(100);
  }
  throw new Error(`Timed out waiting for ${label}`);
}

async function connect() {
  const port = await until(async () => {
    if (launchError) throw launchError;
    if (chrome.exitCode !== null) throw new Error(`Chrome exited with code ${chrome.exitCode}`);
    try {
      return (await readFile(join(profile, 'DevToolsActivePort'), 'utf8')).split('\n')[0];
    } catch (error) {
      if (error.code !== 'ENOENT') throw error;
    }
  }, 'Chrome DevTools');
  const targets = await (await fetch(`http://127.0.0.1:${port}/json/list`)).json();
  const socket = new WebSocket(targets.find(target => target.type === 'page').webSocketDebuggerUrl);
  await once(socket, 'open');
  return socket;
}

function browserCommands(socket) {
  let id = 0;
  const pending = new Map();
  const listeners = new Map();
  socket.addEventListener('message', event => {
    const message = JSON.parse(event.data);
    pending.get(message.id)?.(message);
    for (const listener of listeners.get(message.method) ?? []) listener(message.params);
  });
  function call(method, params = {}) {
    const requestId = ++id;
    return new Promise((resolve, reject) => {
      const timeout = setTimeout(() => {
        pending.delete(requestId);
        reject(new Error(`Timed out running ${method}`));
      }, 60000);
      pending.set(requestId, message => {
        clearTimeout(timeout);
        pending.delete(requestId);
        if (message.error) reject(new Error(JSON.stringify(message.error)));
        else resolve(message.result);
      });
      socket.send(JSON.stringify({ id: requestId, method, params }));
    });
  }
  const on = (method, listener) => {
    if (!listeners.has(method)) listeners.set(method, new Set());
    listeners.get(method).add(listener);
    return () => listeners.get(method).delete(listener);
  };
  return { call, on };
}

async function whenSettledInBrowser(path, subset) {
  const waitForSwap = target => new Promise(resolve => {
    target.addEventListener('htmx:afterSettle', resolve, { once: true });
  });
  // A double frame gives the browser an opportunity to paint the updated DOM.
  const frame = () => new Promise(resolve => requestAnimationFrame(() => requestAnimationFrame(resolve)));
  const tree = document.getElementById('file-tree');
  const directory = () => [...tree.querySelectorAll('button[hx-vals]')].find(button =>
    JSON.parse(button.getAttribute('hx-vals')).path === path);
  if (!directory() || tree.classList.contains('htmx-settling')) await waitForSwap(tree);
  const treeLoaded = waitForSwap(tree);
  const button = directory();
  if (!button) throw new Error(`Missing fixture directory ${path}`);
  button.click();
  await treeLoaded;
  if (document.getElementById('tree-current-path').textContent.trim() !== path) {
    throw new Error(`Directory did not open: ${path}`);
  }
  performance.clearResourceTimings();

  const available = document.getElementById('file-list');
  const selected = document.getElementById('selected-files');
  const availability = waitForSwap(available);
  const selection = subset && waitForSwap(selected);
  const start = performance.now();

  if (subset) {
    document.getElementById('path-random').value = path;
    document.getElementById('runCount').value = 6;
    document.getElementById('filesPerRun').value = 1000;
    document.getElementById('minFileSizeMB').value = 0;
    document.getElementById('random-subset-form').requestSubmit();
  } else {
    document.getElementById('switch-to-find-button').click();
  }

  const afterPaint = async swapped => {
    if (!swapped) return null;
    await swapped;
    await frame();
    return performance.now() - start;
  };
  const [firstAvailableFrameMs, firstSelectedFrameMs] = await Promise.all([afterPaint(availability), afterPaint(selection)]);
  const rows = available.querySelectorAll('li').length;
  const selectedRows = selected.querySelectorAll('li').length;
  const documentElements = document.querySelectorAll('*').length;
  const file = available.querySelector('li');
  const selectStart = performance.now();
  file.click();
  await frame();
  const selectFrameMs = performance.now() - selectStart;
  const requests = performance.getEntriesByType('resource')
    .filter(entry => entry.initiatorType === 'xmlhttprequest')
    .map(entry => ({ path: new URL(entry.name).pathname, bodyBytes: entry.encodedBodySize }));
  return { rows, selectedRows, documentElements, firstAvailableFrameMs, firstSelectedFrameMs, selectFrameMs, requests };
}

// Each navigation starts counting before the parser requests eager gallery images.
const galleryObserver = `performance.setResourceTimingBufferSize(2000);
new MutationObserver((_, observer) => {
  const image = document.querySelector('#carousel .carousel-image:not(.hidden) img');
  if (!image) return;
  observer.disconnect();
  image.decode().then(() => requestAnimationFrame(() => requestAnimationFrame(() => {
    window.firstPlotFrameMs = performance.now();
  })));
}).observe(document, {childList: true, subtree: true});`;

async function galleryInBrowser() {
  if (!window.firstPlotFrameMs) {
    await new Promise(resolve => {
      const timer = setInterval(() => {
        if (window.firstPlotFrameMs) {
          clearInterval(timer);
          resolve();
        }
      }, 20);
    });
  }
  const nav = performance.getEntriesByType('navigation')[0];
  const plots = performance.getEntriesByType('resource').filter(entry => new URL(entry.name).pathname.startsWith('/data/'));
  const start = performance.now();
  document.querySelector('button[onclick="nextImage()"]').click();
  await document.querySelector('#carousel .carousel-image:not(.hidden) img').decode();
  await new Promise(resolve => requestAnimationFrame(() => requestAnimationFrame(resolve)));
  const nextPlotFrameMs = performance.now() - start;
  return {
    plotResponses: plots.length,
    plotBodyBytes: plots.reduce((total, entry) => total + entry.encodedBodySize, 0),
    plotElements: document.querySelectorAll('img[src^="/data/"]').length,
    documentElements: document.querySelectorAll('*').length,
    pageBodyBytes: nav.encodedBodySize,
    firstPlotFrameMs: window.firstPlotFrameMs,
    loadEventMs: nav.loadEventEnd,
    nextPlotFrameMs,
  };
}

let socket;
try {
  socket = await connect();
  const { call, on } = browserCommands(socket);
  await call('Page.enable');
  await call('Network.enable');
  await call('Network.setCacheDisabled', { cacheDisabled: true });
  await call('Page.addScriptToEvaluateOnNewDocument', { source: galleryObserver });

  async function navigate(path) {
    const loaded = Promise.withResolvers();
    const stop = on('Page.loadEventFired', loaded.resolve);
    try {
      const result = await call('Page.navigate', { url: base + path });
      if (result.errorText) throw new Error(result.errorText);
      const timer = setTimeout(() => loaded.reject(new Error(`Timed out loading ${path}`)), 60000);
      try {
        await loaded.promise;
      } finally {
        clearTimeout(timer);
      }
    } finally {
      stop();
    }
  }
  async function evaluate(expression) {
    const result = await call('Runtime.evaluate', { expression, awaitPromise: true, returnByValue: true });
    if (result.exceptionDetails) throw new Error(result.result.description || result.exceptionDetails.text);
    return result.result.value;
  }

  await navigate('/login');
  await navigate('/training-tasks');
  const galleries = await until(async () => {
    const links = await evaluate(`Object.fromEntries([...document.querySelectorAll('a')]
      .filter(a => a.textContent.startsWith('Fixture gallery ')).map(a => [a.textContent.trim(), a.pathname]))`);
    return Object.keys(links).length === 2 && links;
  }, 'gallery task links');
  const samples = [];
  for (let rep = 1; rep <= repetitions; rep++) {
    for (const [available, path, subset] of [
      [10581, '/alice/sim/2024/LHC24f3/0', false],
      [13355, '/alice/sim/2025/LHC25h3c/0', false],
      [13355, '/alice/sim/2025/LHC25h3c/0', true],
    ]) {
      await navigate('/training-datasets/new');
      const metrics = await evaluate(`(${whenSettledInBrowser})(${JSON.stringify(path)}, ${subset})`);
      if (metrics.rows + metrics.selectedRows !== available || metrics.selectedRows !== (subset ? 6000 : 0)) {
        throw new Error(`Unexpected fixture size: ${JSON.stringify(metrics)}`);
      }
      samples.push({ rep, available, subset, ...metrics });
    }
    for (const plots of [180, 738]) {
      const path = galleries[`Fixture gallery ${plots}`];
      if (!path) throw new Error(`Gallery fixture ${plots} missing from task list`);
      await navigate(path);
      const metrics = await evaluate(`(${galleryInBrowser})()`);
      if (!metrics.plotResponses || !metrics.plotBodyBytes) {
        throw new Error(`Unexpected gallery fixture: ${JSON.stringify(metrics)}`);
      }
      samples.push({ rep, plots, ...metrics });
    }
  }
  const { product } = await call('Browser.getVersion');
  process.stdout.write(JSON.stringify({ browser: product, repetitions, samples }, null, 2) + '\n');
} finally {
  socket?.close();
  if (chrome.pid && chrome.exitCode === null) {
    chrome.kill();
    await once(chrome, 'exit');
  }
}
