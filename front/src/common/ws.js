import { encode, decode } from '@/common/jsonpack';
import store from '@/store';
import router from '@/router';
import { clearTimeout, setTimeout } from 'timers';

let beforeSend = {};

function Socket() {
  this.callback = {};
  this.ws = undefined;
  this.retried = 0;

  let host = '';
  let port = '';

  const connectRetry = 500;

  let timeout = 0;
  let receiver = {};
  let pingpongLock = true;
  let connected = [];

  this.onConnected = () => {
    return new Promise((resolve, reject) => {
      if (this.ws && this.ws.readyState === WebSocket.OPEN) {
        resolve(this);
      } else connected.push(resolve);
    });
  };
  this.connect = force => {
    if (
      !force &&
      host === store.state.settings.serverHost &&
      port === store.state.settings.serverPort
    ) {
      return;
    }
    host = store.state.settings.serverHost;
    port = store.state.settings.serverPort;
    this.disconnect();
    if (!host.length || !port.length) {
      nav2Settings();

      return;
    }
    this.ws = new WebSocket(`ws://${host}:${port}/api`);

    this.ws.onopen = () => {
      store.dispatch('wsConnected');
      this.retried = 0;
      this.sendString('ping');
      for (; 0 < connected.length; ) {
        connected[0](this);
        delete connected[0];
        connected = connected.slice(1);
      }
    };
    this.ws.onerror = event => {
      if (event.type === 'error') {
        store.dispatch('wsDisconnected');
        nav2Settings();
      }
    };
    this.ws.onclose = event => {
      store.dispatch('wsDisconnected');
      if (this.retried++ >= connectRetry) {
        this.retried = 0;
        console.log('websocket reconnect failed', this.retried);
        return;
      }
      console.log('websocket closing. retrying', this.retried);
      if (timeout) clearTimeout(timeout);
      timeout = setTimeout(() => {
        timeout = 0;
        this.connect(true);
      }, 1000);
    };

    function onresponse(id, code, data) {
      if (id === 'pon pon pon') {
        pingpongLock = true;
        setTimeout(() => {
          if (pingpongLock) {
            pingpongLock = false;
            this.sendString('ping');
          }
        }, 10000);
        return;
      }

      if (receiver[id] instanceof Function && data instanceof Array) {
        receiver[id].apply(undefined, data);
      } else if (this.callback[id] !== undefined) {
      } else if (id === 'xxxxxxxxxxx') {
        if (code) {
          store.dispatch('showToast', `Received error:${code}`);
        }
        return;
      }

      console.log('Received Message: ', id, code, data);

      if (this.callback[id]) {
        let resolve = this.callback[id][0];
        let reject = this.callback[id][1];
        let st = this.callback[id][2];
        clearTimeout(st);
        delete this.callback[id];

        if (code === 0) {
          resolve(data);
        } else {
          store.dispatch('showToast', `Received error:${code} (${id})`);
          reject(code, data);
        }
      }
      // 未配置回调
    }

    this.ws.onmessage = evt => {
      if (evt.data instanceof Blob) {
        let that = this;
        let fileReader = new FileReader();
        fileReader.onload = function () {
          let id = '';
          let result = new Uint8Array(this.result);
          for (let i = 0; i < 11 && result[i]; i++) {
            id += String.fromCharCode(result[i]);
          }

          onresponse.call(that, id, result[11], decode(result.slice(12)));
        };
        fileReader.readAsArrayBuffer(evt.data);
        return;
      } else if (typeof evt.data === 'string' && evt.data.length >= 11) {
        let id = evt.data.slice(0, 11);
        onresponse.call(this, id, evt.data.charCodeAt(11), evt.data.slice(12));
        return;
      }
      console.log('Received invalid', evt);
    };
  };

  this.send = (cmd, params, options) => {
    let opt_nocallback = options && options.noResponse;
    let isString = (options && options.raw) || false;

    let id = Number(Math.random().toString().substr(2) + Date.now())
      .toString(36)
      .substr(0, 11);

    for (let i in beforeSend) {
      if (beforeSend.hasOwnProperty(i) && i === cmd) {
        return new Promise((resolve, reject) => {
          resolve(beforeSend[i](params));
        });
      }
    }

    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
      console.error('websocket not connect');
      return opt_nocallback ? true : promiseReject(id);
    }

    if (isString) {
      this.ws.send(encodeRaw(id, cmd, params));
    } else {
      this.ws.send(encode({ id, cmd, params }));
    }

    return opt_nocallback
      ? true
      : promiseCallback.call(this, id, cmd, isString);
  };

  this.sendString = (command, params) => {
    return this.send(command, params, { raw: true, noResponse: true });
  };
  this.receive = (command, callback) => {
    if (callback instanceof Function) receiver['' + command] = callback;
  };

  this.disconnect = () => {
    if (this.ws) {
      this.ws.close();
      delete this.ws;
    }
    this.retried = 0;
    this.callback = {};
  };

  this.connect();
}

Socket.addBeforeSend = (cmd, mock) => {
  if (mock instanceof Function) {
    beforeSend['' + cmd] = mock;
  }
};

function promiseCallback(id, command, isRaw) {
  return new Promise((resolve, reject) => {
    let st = setTimeout(() => {
      console.log('websocket timeout', id, command);
      store.dispatch('showToast', `request timeout ${id}`);
      delete this.callback[id];
      reject();
    }, 1000);
    this.callback[id] = [resolve, reject, st];
    if (isRaw) {
      this.callback[id].push(true);
    }
  });
}
function promiseResolve(id) {
  return new Promise((resolve, reject) => {
    resolve();
  });
}
function promiseReject(id) {
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
  router.push({
    name: 'settings',
    params: {
      forceTo: 'server',
    },
  });
}

export default Socket;
