console.debug('[debug][initLocalStorage.js]');

let localStorage = {
  settings: {
    serverHost: '',
    serverPort: '4415',
    enableDebugTool: false,
    closeAppOption: 'ask',
    subTitleDefault: false,
    linuxEnableCustomTitlebar: false,
  },
  data: {
    loginMode: null,
  },
};

export default localStorage;
