import initLocalStorage from './initLocalStorage';
import pkg from '../../package.json';

if (localStorage.getItem('appVersion') === null) {
  localStorage.setItem('settings', JSON.stringify(initLocalStorage.settings));
  localStorage.setItem('data', JSON.stringify(initLocalStorage.data));
  localStorage.setItem('appVersion', pkg.version);
}

export default {
  wsConnected: false,
  enableScrolling: true,
  title: 'YesPlayMusic',
  toast: {
    show: false,
    text: '',
    timer: null,
  },
  settings: JSON.parse(localStorage.getItem('settings')),
  data: JSON.parse(localStorage.getItem('data')),
};
