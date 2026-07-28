export default {
  updateWsConnected(state, connect) {
    state.wsConnected = connect;
  },
  updateSettings(state, { key, value }) {
    state.settings[key] = value;
  },
  updateData(state, { key, value }) {
    state.data[key] = value;
  },
  updateToast(state, toast) {
    state.toast = toast;
  },
  updateModal(state, { modalName, key, value }) {
    state.modals[modalName][key] = value;
    if (key === 'show') {
      // 100ms的延迟是为等待右键菜单blur之后再disableScrolling
      value === true
        ? setTimeout(() => (state.enableScrolling = false), 100)
        : (state.enableScrolling = true);
    }
  },
  enableScrolling(state, status = null) {
    state.enableScrolling = status ? status : !state.enableScrolling;
  },
  updateTitle(state, title) {
    state.title = title;
  },
};
