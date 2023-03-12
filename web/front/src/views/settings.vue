<template>
  <div class="settings-page">
    <a-anchor wrapperClass="anchors" :getContainer="() => this.$parent.$refs.main" showInkInFixed @click.prevent=""
      :offsetTop="20">
      <a-anchor-link href="#basic" title="基本" />
      <a-anchor-link href="#config" title="配置" />
    </a-anchor>
    <div class="container is-max-desktop">
      <div id="basic" class="hr">基本</div>
      <div class="field is-horizontal" :class="{ 'highlight animate__animated animate__headShake': !wsConnected, }">
        <div class="field-label">
          <label class="label" for="server-host">服务器</label>
        </div>
        <div class="field-body">
          <div class="control">
            <div class="input-ip-group">
              <a-input v-model="serverHost" id="server-host" class="input"
                :class="{ 'is-danger': !wsConnected || hostError, }" placeholder="服务器地址" @change="" />
              <a-input-number v-model="serverPort" class="input" :min="1" :max="65535"
                :class="{ 'is-danger': !wsConnected || portError, }" placeholder="端口" type="number"
                @change="serverPort = $event" />
            </div>
          </div>
        </div>
      </div>

      <div class="field is-horizontal">
        <div class="field-label">
          <label class="label" for="appearance">主题</label>
        </div>
        <div class="field-body">
          <div class="control">
            <div class="select">
              <a-select v-model="appearance" id="appearance">
                <a-select-option value="auto">自动&ensp;&ensp;&ensp;&ensp;</a-select-option>
                <a-select-option value="light">🌞 浅色</a-select-option>
                <a-select-option value="dark">🌚 暗黑</a-select-option>
              </a-select>
            </div>
          </div>
        </div>
      </div>

      <div v-if="isElectron && !isMac" class="field is-horizontal">
        <div class="field-label">
          <label class="label" for="close-app-option">关闭主面板时...</label>
        </div>
        <div class="field-body">
          <div class="control">
            <div class="select">
              <select v-model="closeAppOption" id="close-app-option">
                <option value="ask">询问</option>
                <option value="exit">直接退出</option>
                <option value="minimizeToTray">最小化到托盘</option>
              </select>
            </div>
          </div>
        </div>
      </div>

      <div v-if="isElectron && isLinux" class="field is-horizontal">
        <div class="field-label">
          <label class="label" for="enable-custom-titlebar">启用自定义标题栏 (重启后生效)</label>
        </div>
        <div class="field-body">
          <div class="control">
            <div class="toggle">
              <input id="enable-custom-titlebar" v-model="enableCustomTitlebar" class="input" type="checkbox"
                name="enable-custom-titlebar" />
              <label for="enable-custom-titlebar" />
            </div>
          </div>
        </div>
      </div>

      <div id="line" class="hr">线路</div>
      <div class="field is-horizontal">
        <div class="field-label">
          <label class="label" for="spectrum">音阶、频谱图</label>
        </div>
        <div class="field-body">
          <div class="control">
            <div class="select">
              <a-switch id="spectrum" v-model="spectrum" checked-children="开" un-checked-children="关" />
            </div>
          </div>
        </div>
      </div>

      <div id="config" class="hr">配置</div>
      <div class="table">
        <a-table :columns="configsColume" :data-source="configsData" size="small" :pagination="false" />
      </div>
    </div>
  </div>
</template>

<script>
import { mapState, mapActions } from 'vuex';
import { socket } from '@/common/request';
import { isIP46, isPort } from '@/common/format';
import { changeAppearance } from '@/common/theme';
import * as ApiSystem from '@/api/system';
import pkg from '../../package.json';

const electron =
  process.env.IS_ELECTRON === true ? window.require('electron') : null;

export default {
  name: 'Settings',
  data() {
    return {
      isConnectErr: false,
      hostError: false,
      portError: false,
      configsData: [],
    };
  },
  watch: {
    wsConnected(newVal, oldVal) {
      if (newVal && !this.isConnectErr) {
        this.$router.go(-1);
      }
    },
  },
  computed: {
    ...mapState(['settings', 'data', 'wsConnected']),
    isElectron() {
      return process.env.IS_ELECTRON;
    },
    isMac() {
      return /macintosh|mac os x/i.test(navigator.userAgent);
    },
    isLinux() {
      return process.platform === 'linux';
    },
    version() {
      return pkg.version;
    },
    appearance: {
      get() {
        if (this.settings.appearance === undefined) return 'auto';
        return this.settings.appearance;
      },
      set(value) {
        this.$store.commit('updateSettings', { key: 'appearance', value, });
        changeAppearance(value);
      },
    },
    closeAppOption: {
      get() {
        return this.settings.closeAppOption;
      },
      set(value) {
        this.$store.commit('updateSettings', { key: 'closeAppOption', value, });
      },
    },
    serverHost: {
      get() {
        return this.settings.serverHost || '';
      },
      set(value) {
        this.hostError = !isIP46(value);
        this.$store.commit('updateSettings', { key: 'serverHost', value });
        if (!this.hostError && !this.portError) {
          socket.connect();
        }
      },
    },
    serverPort: {
      get() {
        return this.settings.serverPort || '';
      },
      set(value) {
        this.portError = !isPort(value);
        this.$store.commit('updateSettings', { key: 'serverPort', value });
        if (!this.hostError && !this.portError) {
          socket.connect();
        }
      },
    },
    enableCustomTitlebar: {
      get() {
        return this.settings.linuxEnableCustomTitlebar;
      },
      set(value) {
        this.$store.commit('updateSettings', { key: 'linuxEnableCustomTitlebar', value });
      },
    },
    spectrum: {
      get() {
        return this.settings.showSpectrum;
      },
      set(value) {
        this.$store.commit('updateSettings', { key: 'showSpectrum', value });
      },
    },
    configsColume() {
      let secion = '___'
      return [
        {
          title: '', dataIndex: 'sec', customRender: (value, row, index) => {
            const obj = {
              children: value,
              attrs: {},
            };
            if (secion !== value) {
              let span = 0
              for (let i = index; i < this.configsData.length; i++) {
                if (this.configsData[i].sec === value)
                  span++
                else
                  break;
              }
              obj.attrs.rowSpan = span;
              secion = value
            } else {
              obj.attrs.rowSpan = 0;
            }
            return obj;
          },
        },
        { title: '名称', dataIndex: 'kn' },
        { title: '值', dataIndex: 'val' },
      ]
    },

  },
  mounted() {
  },
  destroyed() { },

  created() { },
  activated() {
    ApiSystem.config().then(c => {
      this.configsData = c.map((d, i) => {
        const vs = d.name.split('.')
        d.sec = vs[0]
        d.kn = vs[1]
        d.key = i
        return d
      }).sort((d1, d2) => {
        return d1.name > d2.name ? 1 : (d1.name === d2.name ? 0 : -1)
      })
    })
  },
  methods: {
    ...mapActions(['showToast']),
    isHighlight(name) {
      return this.$route.params.forceTo === name;
    },
  },
};
</script>

<style lang="scss">
.settings-page {
  padding: 1rem;
  display: flex;
  justify-content: center;

  .anchors {
    position: fixed;
    top: var(--size-navbar) + 50px;
    right: 1rem;
  }

  .hr {
    margin: 2rem 0 1rem 0;
    padding: 0 0 0.5rem 0;
    border-bottom: 1px solid var(--color-border);
    font-size: 1rem;
    font-weight: bold;
    line-height: 1rem;
  }

  .container {
    width: 720px;
    flex-grow: 0;
  }

  .field {
    display: flex;
    justify-content: flex-start;
    align-items: center;
    margin: 5px 0;

    .field-label {
      flex-shrink: 1;
      text-align: right;
      margin-right: 1rem;
      font-size: 1rem;
      font-weight: bold;
      min-width: 5rem;
    }

    .field-body {
      flex-grow: 1;
      margin-left: 1rem;
      display: flex;


      .control {
        text-align: left;
      }
    }

    &.is-horizontal {
      overflow: hidden;
    }
  }

  .table {
    margin: 5px 0;

  }


  .input-ip-group {
    display: flex;
    justify-content: right;
  }

  .input-ip-group .input:first-child {
    border-top-right-radius: 0;
    border-bottom-right-radius: 0;
    max-width: 15rem;
  }

  .input-ip-group .input:last-child {
    border-top-left-radius: 0;
    border-bottom-left-radius: 0;
    border-left: 0;
    max-width: 8rem;
    width: 8rem;
  }

  .is-danger {
    color: var(--color-danger);
    border-color: var(--color-danger);

    input,
    .ant-input-number-input {
      color: var(--color-danger);
    }
  }

  .beforeAnimation {
    -webkit-transition: 0.2s cubic-bezier(0.24, 0, 0.5, 1);
    transition: 0.2s cubic-bezier(0.24, 0, 0.5, 1);
  }

  .afterAnimation {
    box-shadow: 0 0 0 1px hsla(0, 0%, 0%, 0.1), 0 4px 0px 0 hsla(0, 0%, 0%, 0.04),
      0 4px 9px hsla(0, 0%, 0%, 0.13), 0 3px 3px hsla(0, 0%, 0%, 0.05);
    -webkit-transition: 0.35s cubic-bezier(0.54, 1.6, 0.5, 1);
    transition: 0.35s cubic-bezier(0.54, 1.6, 0.5, 1);
  }



  @media screen and (min-width: 768px),
  print {
    .field-label {
      margin-left: 3em;
    }

    .field-body {
      margin-right: 3em;
    }
  }

  @media screen and (max-width: 820px) {
    .anchors {
      display: none;
    }
  }
}
</style>
