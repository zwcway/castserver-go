import router from '@/router';
import axios from 'axios';
import Socket from '@/common/ws';
import store from '@/store';

if (process.env.NODE_ENV !== 'production') {
  require('../mock');
}

function serverBase() {
  return `${store.state.settings.serverHost}:${store.state.settings.serverPort}`;
}

let baseURL = serverBase();

const socket = new Socket();

const service = axios.create({
  baseURL,
  withCredentials: true,
  timeout: 15000,
});

service.interceptors.request.use(function (config) {
  if (!config.params) config.params = {};
  if (
    store.state.settings.serverHost.length === 0 ||
    store.state.settings.serverPort.length === 0
  ) {
    router.push({
      name: 'settings',
      params: {
        forceTo: 'server',
      },
    });
  }
  config.url = `http://${store.state.settings.serverHost}:${store.state.settings.serverPort}/api/${config.url}`;

  return config;
});

service.interceptors.response.use(
  response => {
    const res = response.data;
    console.log(res);
    return res;
  },
  async error => {
    /** @type {import('axios').AxiosResponse | null} */
    const response = error.response;
    console.log(response);
  }
);

export { socket, service as axios };
