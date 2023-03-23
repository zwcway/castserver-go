import Vue from 'vue';
import Router from 'vue-router';
import Antd from 'ant-design-vue';
import Vue2TouchEvents from 'vue2-touch-events'

import App from './App.vue';
import router from './router';
import store from './store';

import 'ant-design-vue/dist/antd.css';
import '@/assets/icons';
import '@/assets/css/layout.scss';
import '@/assets/css/slider.scss';
import '@/assets/css/ant-design-vue.scss';
import 'animate.css';
import 'ionicons';

window.store = store;

if (process.env.NODE_ENV !== 'production') {
  store.commit('updateSettings', { key: 'enableDebugTool', value: true });
  window.resetApp = () => {
    localStorage.clear();
    indexedDB.deleteDatabase(process.env.AppID);
    document.cookie.split(';').forEach(function (c) {
      document.cookie = c
        .replace(/^ +/, '')
        .replace(/=.*/, '=;expires=' + new Date().toUTCString() + ';path=/');
    });
    return '已重置应用，请刷新页面（按Ctrl/Command + R）';
  };
  console.log(
    '如出现问题，可尝试在本页输入 %cresetApp()%c 然后按回车重置应用。',
    'background: #eaeffd;color:#335eea;padding: 4px 6px;border-radius:3px;',
    'background:unset;color:unset;'
  );

  if (store.state.settings.enableMock) {
    require('./mock');
  }
  window.socket = require('@/common/ws').default;
  window.socket.mock = (b) => {
    store.commit('updateSettings', { key: 'enableMock', value: b === undefined || b })
    window.location.reload()
  }
}

const originalPush = Router.prototype.push;
Router.prototype.push = function push(location) {
  return originalPush.call(this, location).catch(err => err);
};

Vue.config.productionTip = false;

Vue.use(Antd);
Vue.use(Vue2TouchEvents, {
  disableClick: false,
  touchClass: '',
  tapTolerance: 10,
  touchHoldTolerance: 400,
  swipeTolerance: 30,
  longTapTimeInterval: 400,
  namespace: 'touch'
})

new Vue({
  router,
  store,
  render: h => h(App),
}).$mount('#app');
