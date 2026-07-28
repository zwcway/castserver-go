import Vue from 'vue';
import Vuex from 'vuex';
import state from './state';
import mutations from './mutations';
import actions from './actions';
import saveToLocalStorage from './plugins/localStorage';

Vue.use(Vuex);

const options = {
  state,
  mutations, // setter
  actions, //
  plugins: [saveToLocalStorage],
};

const store = new Vuex.Store(options);

window
  .matchMedia('(prefers-color-scheme: dark)')
  .addEventListener('change', () => {
    if (store.state.settings.appearance === 'auto') {
      console.log('auto');
    }
  });

export default store;
