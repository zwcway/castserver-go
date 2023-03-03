import Vue from 'vue';
import Router from 'vue-router';
import Antd from 'ant-design-vue';
import App from './App.vue';
import router from './router';
import store from './store';

import 'ant-design-vue/dist/antd.css';
import '@/assets/icons';
import '@/assets/css/global.scss';
import 'animate.css';
import 'bulma';
import 'ionicons';

if (process.env.NODE_ENV !== 'production') {
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
}

const originalPush = Router.prototype.push;
Router.prototype.push = function push(location) {
  return originalPush.call(this, location).catch(err => err);
};

Vue.config.productionTip = false;

Vue.use(Antd);

new Vue({
  router,
  store,
  render: h => h(App),
}).$mount('#app');
