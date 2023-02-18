<template>
  <div>
    <nav :class="{ 'has-custom-titlebar': hasCustomTitlebar }">
      <Win32Titlebar v-if="enableWin32Titlebar" />
      <LinuxTitlebar v-if="enableLinuxTitlebar" />
      <div class="tabs is-large is-boxed">
        <div class="navbar-start">
          <div class="navbar-item">
            <div class="buttons">
              <a class="button is-normal is-none is-inverted">
                <svg-icon icon-class="logo" :size="24" />
              </a>
              <a
                class="link-state button is-normal is-none is-inverted"
                :class="{
                  'is-danger': !wsConnected,
                  'is-primary': wsConnected,
                }"
                :title="wsConnected ? '服务器已连接' : '服务器已断开'"
              >
                <svg-icon
                  :icon-class="wsConnected ? 'link' : 'unlink'"
                  :size="0"
                />
              </a>
            </div>
          </div>
        </div>
        <ul>
          <li :class="{ 'is-active': $route.name === 'speakers' }">
            <a
              :href="$router.resolve({ name: 'speakers' }).href"
              @click="reload"
            >
              <svg-icon icon-class="speakers" class="icon" :size="0" />
              <span>扬声器列表</span>
            </a>
          </li>
          <li
            v-for="line in lines"
            :class="{ 'is-active': isLineRoute(line.id) }"
          >
            <router-link :to="`/line/${line.id}`">
              <svg-icon icon-class="music" class="icon" :size="0" />
              <span>{{ line.name }}</span>
            </router-link>
          </li>
        </ul>
        <div class="navbar-end">
          <div class="navbar-item">
            <div class="buttons">
              <router-link
                to="/settings"
                class="button is-primary is-normal"
                :class="{ 'is-outlined': $route.name !== 'settings' }"
                title="设置"
              >
                <svg-icon :icon-class="'settings'" :size="32" />
              </router-link>
            </div>
          </div>
        </div>
      </div>
    </nav>
  </div>
</template>

<script>
import { mapState } from 'vuex';
// import icons for win32 title bar
// icons by https://github.com/microsoft/vscode-codicons
import '@vscode/codicons/dist/codicon.css';
import Win32Titlebar from '@/components/Win32Titlebar.vue';
import LinuxTitlebar from '@/components/LinuxTitlebar.vue';
import { socket } from '@/common/request';
import { getLineList, listenLineChanged } from '@/api/line';

export default {
  name: 'Navbar',
  components: {
    Win32Titlebar,
    LinuxTitlebar,
  },
  inject: ['reload'],
  data() {
    return {
      inputFocus: false,
      langs: ['zh-CN', 'zh-TW', 'en', 'tr'],
      keywords: '',
      enableWin32Titlebar: false,
      enableLinuxTitlebar: false,
      lines: [
        { id: 0, name: '主卧' },
        { id: 1, name: '次卧' },
      ],
    };
  },
  computed: {
    ...mapState(['settings', 'data', 'wsConnected']),
    hasCustomTitlebar() {
      return this.enableWin32Titlebar || this.enableLinuxTitlebar;
    },
  },
  created() {
    if (process.platform === 'win32') {
      this.enableWin32Titlebar = true;
    } else if (
      process.platform === 'linux' &&
      this.settings.linuxEnableCustomTitlebar
    ) {
      this.enableLinuxTitlebar = true;
    }
  },
  mounted() {
    socket.onConnected().then(() => this.loadData());
  },
  methods: {
    loadData() {
      getLineList().then(data => {
        this.lines = data;
      });
      listenLineChanged(data => {});
    },
    go(where) {
      if (where === 'back') this.$router.go(-1);
      else this.$router.go(1);
    },
    toRouteLine(line) {
      this.$router.push({ path: `/line/${line.id}` });
    },
    toRoute(name) {
      this.$router.push({ name });
    },
    isLineRoute(id) {
      return (
        this.$route.name === 'line' &&
        Number(this.$route.params.id) === Number(id)
      );
    },
  },
};
</script>
<style lang="scss" scoped>
@import '../../node_modules/bulma/bulma.sass';

.tabs .buttons .button {
  border: 0;
  border-radius: 100%;
  padding: 8px;
  width: 3rem;
  height: 3rem;
}

.button.is-none:hover {
  background: none !important;
  cursor: auto;
}

.tabs .navbar-start,
.tabs .navbar-end {
  border-bottom-color: hsl(0deg, 0%, 86%);
  border-bottom-style: solid;
  border-bottom-width: 1px;
}

.tabs a {
  text-decoration: none;
  flex-direction: column;
}

.tabs {
  overflow: hidden;
}

.tabs ul {
  overflow: hidden;
  overflow-x: auto;
  justify-content: flex-start;
  flex-grow: 1;
  flex-shrink: unset;
}

.tabs ul::-webkit-scrollbar {
  display: none;
}

.tabs ul li {
  flex-basis: 88px;
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
}

.tabs .svg-icon {
  width: 3rem;
  height: 3rem;
}

.tabs span {
  font-size: 0.5rem;
  margin-top: 0.5rem;
}

.navbar-start {
  display: flex;
  align-items: center;
  box-shadow: 0.2rem 0 0.5em -0.5rem $box-color;
}

.navbar-end {
  display: flex;
  align-items: center;
  box-shadow: -0.2rem 0 0.5em -0.5rem $box-color;
}

.navbar-start {
  .navbar-item {
    height: 100%;
    position: relative;

    .link-state {
      width: 1rem;
      height: 1rem;
      padding: 0;
      position: absolute;
      bottom: 0.1rem;
      right: 0.5rem;

      .svg-icon {
        width: 100% !important;
        height: 100% !important;
      }
    }
  }
}
</style>
