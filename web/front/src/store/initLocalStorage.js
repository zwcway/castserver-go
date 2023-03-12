console.debug('[debug][initLocalStorage.js]');

let localStorage = {
  settings: {
    serverHost: '',
    serverPort: '4415',
    enableDebugTool: false,
    enableMock: false,
    closeAppOption: 'ask',
    subTitleDefault: false,
    linuxEnableCustomTitlebar: false,

    showSpectrum: true,
  },
  data: {
    loginMode: null,
  },
};

export default localStorage;
