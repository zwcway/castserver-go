import { encodeReq, decode } from '@/common/jsonpack';
import store from '@/store';
import router from '@/router';
// import { clearTimeout, setTimeout } from 'timers';
import { isIP46, isPort } from '@/common/format';

let beforeSend = {};
let receiver = {};

let sockthis = this;
let callback = {};
let ws = undefined;
let retried = 0;

let wsHost = '';
let wsPort = '';

const connectRetry = 500;

let retryConnectTimeout = 0;
let connected = [];

function onConnected() {
  return new Promise((resolve, reject) => {
    if (beforeSend['wsconnect']) {
      resolve();
      return;
    }
    if (ws && ws.readyState === WebSocket.OPEN) {
      resolve();
    } else connected.push(resolve);
  });
}

function wsonopen() {
  store.dispatch('wsConnected');
  retried = 0;
  sendPing();

  for (; 0 < connected.length;) {
    connected[0](sockthis);
    delete connected[0];
    connected = connected.slice(1);
  }
}

function wsonerror(event) {
  if (event.type === 'error') {
    store.dispatch('wsDisconnected');
    nav2Settings();
  }
}
function wsonclose() {
  store.dispatch('wsDisconnected');
  if (retried++ >= connectRetry) {
    retried = 0;
    console.log('websocket reconnect failed', retried);
    return;
  }
  console.log('websocket closing. retrying', retried);

  if (retryConnectTimeout) clearTimeout(retryConnectTimeout);

  retryConnectTimeout = setTimeout(() => {
    retryConnectTimeout = 0;
    connect(true);
  }, 1000);
}

function onresponse(id, code, data) {
  if (callback[id] !== undefined) {
  } else if (id === 'xxxxxxxxxxx') {
    if (code) {
      store.dispatch('showToast', `Received error:${code}`);
    }
    return;
  }

  if (callback[id]) {
    let cb = callback[id];
    let cmd = cb['command'];
    let resolve = cb['resolve'];
    let reject = cb['reject'];
    let st = cb['st'];
    let params = cb['params'];
    clearTimeout(st);
    delete callback[id];

    console.log('Received Message: ', cmd, id, params, code, data);

    if (code === 0) {
      resolve(data);
    } else {
      store.dispatch('showToast', `Received error:${code} (${id})`);
      reject(code, data);
    }
  } else {
    // 未配置回调
    console.log('Received Message: ', id, code, data);
  }
}
function wsonmessage(evt) {
  if (evt.data instanceof Blob) {
    readBlob(evt.data);
    return;
  } else if (typeof evt.data === 'string') {
    readText(evt.data);
    return;
  }
  console.log('Received invalid', evt);
}

function readBlob(data) {
  let fileReader = new FileReader();
  fileReader.onload = function () {
    let id = '';
    let result = new Uint8Array(fileReader.result);
    for (let i = 0; i < 5 && result[i]; i++) {
      id += String.fromCharCode(result[i]);
    }
    // 事件格式： event+evt+sub+data
    if (
      result[0] === 101 &&
      result[1] === 118 &&
      result[2] === 101 &&
      result[3] === 110 &&
      result[4] === 116
    ) {
      let evt = result[5];
      let sub = result[6];
      let arg = result[7];
      let data = decode(result.slice(8));

      if (receiver[evt + '.' + sub + '-' + arg] instanceof Function) {
        receiver[evt + '.' + sub + '-' + arg].call(undefined, data, evt, sub);
      }
      if (receiver[evt + '-' + arg] instanceof Function) {
        receiver[evt + '-' + arg].call(undefined, data, evt, sub);
      }
      if (receiver['' + evt] instanceof Function) {
        receiver['' + evt].call(undefined, data, evt, sub);
      }
      return;
    }
    // 普通格式： id+code+data
    for (let i = 5; i < 11 && result[i]; i++) {
      id += String.fromCharCode(result[i]);
    }
    onresponse(id, result[11], decode(result.slice(12)));
  };
  fileReader.readAsArrayBuffer(data);
}

function readText(data) {
  if (data === 'pong') {
    setTimeout(() => {
      sendPing();
    }, 30000);
    return;
  }
  if (data.length >= 11) {
    let id = data.slice(0, 11);
    onresponse(id, data.charCodeAt(11), data.slice(12));
  }
}

let GoWSIP = '***Go-WS-IP***';
let GoWSPort = '***Go-WS-Port***';

function connect(force) {
  wsHost = store.state.settings.serverHost || GoWSIP;
  wsPort = store.state.settings.serverPort || GoWSPort;

  disconnect();

  let portValid = isPort(wsPort);
  let hostValid = isIP46(wsHost);

  if ((!hostValid || !portValid) && !force) {
    nav2Settings();
    return;
  }

  if (beforeSend['wsconnect']) {
    ws = beforeSend['wsconnect']();
    return;
  }

  ws = new WebSocket(`ws://${wsHost}:${wsPort}/api`);

  ws.onopen = wsonopen;
  ws.onmessage = wsonmessage;
  ws.onclose = wsonclose;
  ws.onerror = wsonerror;
}

function send(cmd, params, options) {
  let opt_nocallback = options && options.noResponse;
  let isString = (options && options.raw) || false;

  let id = Number(Math.random().toString().substr(2) + Date.now())
    .toString(36)
    .substr(0, 11);

  for (let i in beforeSend) {
    if (beforeSend.hasOwnProperty(i) && i === cmd) {
      return new Promise((resolve, reject) => {
        let ret = beforeSend[i](params);
        console.log('mock', i, params, ret);
        resolve(ret);
      });
    }
  }

  if (!ws || ws.readyState !== WebSocket.OPEN) {
    console.error('websocket not connect');
    return opt_nocallback ? newPromiseResolve() : newPromiseReject(id);
  }

  if (isString) {
    ws.send(encodeRaw(id, cmd, params));
  } else {
    ws.send(encodeReq(id, cmd, params));
  }

  return opt_nocallback ? true : promiseCallback(id, cmd, params);
}

function sendPing() {
  return ws.send('ping');
}

function sendSubscribe(evt, act, sub, arg) {
  let data = { evt, act }

  if (sub) {
    data['sub'] = sub;
  }
  if (arg !== undefined) {
    data['arg'] = parseInt(arg)
  }

  return send('subscribe', { evt, act, sub, arg }, { noResponse: true });
}

function receiveEvent(evt, arg, cb, sub) {
  if (!(evt instanceof Array)) evt = [evt];

  if (arg instanceof Function) {
    cb = arg;
    evt.forEach(e => {
      receiver['' + e] = cb;
    });
    sendSubscribe(true, evt, undefined, sub)
    return
  }
  if (!(cb instanceof Function)) return;
  if (!(arg instanceof Array))
    arg = [arg];

  evt.forEach(e => {
    arg.forEach(a => {
      a = parseInt(a)
      if (sub)
        receiver[e + '.' + sub + '-' + a] = cb;
      else
        receiver[e + '-' + a] = cb;
      sendSubscribe(true, evt, a, sub)
    });
  });
}

function removeEvent(evt, sub, arg) {
  if (!(evt instanceof Array)) {
    evt = [evt];
  }
  if (!(arg instanceof Array)) {
    arg = [arg];
  }
  evt.forEach(e => {
    delete receiver['' + e];
    arg.forEach(a => {
      delete receiver[e + '-' + a];
      sendSubscribe(false, evt, arg, sub)
    });
  });
}

function disconnect() {
  if (ws != undefined) {
    ws.close();
    ws = undefined;
  }
  retried = 0;
  callback = {};
}

function addBeforeSend(cmd, mock) {
  if (mock instanceof Function) {
    beforeSend['' + cmd] = mock;
  }
}

function getReceiver() {
  return receiver;
}

function promiseCallback(id, command, params) {
  return new Promise((resolve, reject) => {
    let st = setTimeout(() => {
      console.log('websocket timeout', id, command, params);
      store.dispatch('showToast', `request timeout ${id}`);
      delete callback[id];
      reject();
    }, 1000);
    callback[id] = { command, resolve, reject, st, params };
  });
}

function newPromiseResolve() {
  return new Promise((resolve, reject) => {
    resolve();
  });
}

function newPromiseReject(id) {
  return new Promise((resolve, reject) => {
    reject();
  });
}

function encodeRaw(id, command, params) {
  if (typeof command !== 'string') {
    console.error(`Command not string`, command);
    return;
  }
  if (
    params !== undefined &&
    typeof params !== 'string' &&
    typeof params !== 'number'
  ) {
    console.error(`Parameter not string/number`, params);
    return;
  }
  let arr = [command, '' + params];

  return arr.join('\x00');
}

function nav2Settings() {
  if (router.currentRoute.name === 'settings') return;
  router
    .push({
      name: 'settings',
      params: {
        forceTo: 'server',
      },
    })
    .catch(err => err);
}

export default {
  onConnected,
  connect,
  send,
  sendSubscribe,
  receiveCommand,
  receiveEvent,
  removeEvent,
  addBeforeSend,
  getReceiver,
};
