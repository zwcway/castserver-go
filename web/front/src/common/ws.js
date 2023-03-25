import { encodeReq, decode } from '@/common/jsonpack';
import store from '@/store';
import router from '@/router';
// import { clearTimeout, setTimeout } from 'timers';
import { isIP46, isPort } from '@/common/format';

let beforeSend = {};
let receiver = {};

let sockthis = this;
let apiCallback = {};
let ws = undefined;
let retried = 0;

let wsHost = '';
let wsPort = '';

const connectRetry = 500;

let retryConnectTimeout = 0;
let connectedCallback = [];

function onConnected() {
  return new Promise((resolve, reject) => {
    if (beforeSend['wsconnect']) {
      resolve();
      return;
    }
    if (ws && ws.readyState === WebSocket.OPEN) {
      resolve();
    } else connectedCallback.push(resolve);
  });
}

function wsonopen() {
  clearTimeout(retryConnectTimeout);
  store.dispatch('wsConnected');
  retried = 0;
  sendPing();

  for (; 0 < connectedCallback.length;) {
    connectedCallback[0](sockthis);
    connectedCallback.splice(0, 1);
  }
}

function wsonerror(event) {
  if (event && event.type === 'error') {
    store.dispatch('wsDisconnected');
    nav2Settings();
  }
}
function wsonclose() {
  store.dispatch('wsDisconnected');
  if (process.env.NODE_ENV !== 'production') {
    return
  }
  if (retried++ >= connectRetry) {
    console.log('websocket reconnect failed', retried);
    return;
  }
  
  console.log('websocket closed. retrying', retried);

  if (retryConnectTimeout)
    clearTimeout(retryConnectTimeout);

  retryConnectTimeout = setTimeout(() => {
    retryConnectTimeout = 0;
    connect(true);
  }, 1000);
}

function onresponse(id, code, data) {
  if (apiCallback[id] !== undefined) {
  } else if (id === 'xxxxxxxxxxx') {
    if (code) {
      store.dispatch('showToast', `Received error:${code}`);
    }
    return;
  }

  if (apiCallback[id]) {
    const cb = apiCallback[id];
    const cmd = cb['command'];
    const resolve = cb['resolve'];
    const reject = cb['reject'];
    const st = cb['st'];
    const params = cb['params'];
    const options = cb['options'];
    clearTimeout(st);
    delete apiCallback[id];

    if (!options || options.log === undefined || options.log) {
      console.log('Received Message: ', cmd, id, params, code, data);
    }

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

let pingTimeout = 0;
function readText(data) {
  if (data === 'pong') {
    clearTimeout(pingTimeout);
    pingTimeout = setTimeout(sendPing, 30000);
    return;
  }
  if (data.length >= 11) {
    let id = data.slice(0, 11);
    onresponse(id, data.charCodeAt(11), data.slice(12));
  }
}

function connect(isRetry) {
  wsHost = store.state.settings.serverHost || '***Go-WS-IP***';
  wsPort = store.state.settings.serverPort || '***Go-WS-Port***';
  disconnect();

  let portValid = isPort(wsPort);
  let hostValid = isIP46(wsHost);

  if ((!hostValid || !portValid)) {
    retried = 0;
    clearTimeout(retryConnectTimeout);
    nav2Settings();
    return;
  }
  if (!isRetry) {
    retried = 0;
    clearTimeout(retryConnectTimeout);
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
  let opt_nocallback = (typeof options === 'boolean') ? options : (options && options.noResponse);
  let isString = (options && options.raw) || false;

  const id = Number(Math.random().toString().substring(2) + Date.now())
    .toString(36)
    .substring(0, 11);

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

  return opt_nocallback ? true : promiseCallback(id, cmd, params, options);
}

function sendPing() {
  return ws.send('ping');
}

function sendSubscribe(act, evt, sub, arg) {
  let data = { evt, act }

  if (sub) {
    data['sub'] = sub;
  }
  if (arg !== undefined && arg !== null) {
    data['arg'] = parseInt(arg)
  }

  return send('subscribe', data);
}

function receiveEvent(evt, arg, cb, sub) {
  if (!(evt instanceof Array)) evt = [evt];

  if (arg instanceof Function) {
    cb = arg;
    arg = undefined
  }
  if (!(cb instanceof Function)) return;
  if (!(arg instanceof Array))
    arg = [arg];

  arg.forEach(a => {
    evt.forEach(e => {
      if (a !== undefined) {
        a = parseInt(a)
        if (sub)
          receiver[e + '.' + sub + '-' + a] = cb;
        else
          receiver[e + '-' + a] = cb;
      } else {
        if (sub)
          receiver[e + '.' + sub] = cb;
        else
          receiver[e + ''] = cb;
      }
    });
    sendSubscribe(true, evt, sub, a)
  });
}

function removeEvent(evt, arg, sub) {
  if (!(evt instanceof Array)) {
    evt = [evt];
  }
  if (!(arg instanceof Array)) {
    arg = [arg];
  }
  arg.forEach(a => {
    evt.forEach(e => {
      delete receiver['' + e];
      delete receiver[e + '-' + a];
      delete receiver[e + '.' + sub];
      delete receiver[e + '.' + sub + '-' + a];
    });
    if (a === undefined) {
      evt.forEach(e => {
        for (let r in receiver) {
          if (r.startsWith(e + '-'))
            delete receiver[r];
          else if (r.startsWith(e + '.' + sub + '-'))
            delete receiver[r];
        }
      });
    }
    sendSubscribe(false, evt, sub, a)
  });
}

function disconnect() {
  if (ws !== undefined) {
    ws.close();
    ws = undefined;
  }
  apiCallback = {};
}

function addBeforeSend(cmd, mock) {
  if (mock instanceof Function) {
    beforeSend['' + cmd] = mock;
  }
}

function getReceiver() {
  return receiver;
}

function promiseCallback(id, command, params, options) {
  return new Promise((resolve, reject) => {
    let st = setTimeout(() => {
      console.log('websocket timeout', id, command, params);
      store.dispatch('showToast', `request timeout ${id}`);
      delete apiCallback[id];
      reject();
    }, 1000);
    apiCallback[id] = { command, resolve, reject, st, params, options };
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
  // sendSubscribe,
  receiveEvent,
  removeEvent,
  addBeforeSend,
  getReceiver,
};
