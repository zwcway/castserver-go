<template>
  <div>
    <nav :class="{ 'has-custom-titlebar': hasCustomTitlebar }">
      <Win32Titlebar v-if="enableWin32Titlebar" />
      <LinuxTitlebar v-if="enableLinuxTitlebar" />
      <div class="border"></div>
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
              <span>扬声器</span>
            </a>
          </li>
          <li
            v-for="line in lines"
            :key="line.id"
            :class="{ 'is-active': isLineRoute(line.id) }"
          >
            <router-link :to="`/line/${line.id}`">
              <svg-icon icon-class="music" class="icon" :size="0" />
              <span :id="'nav-' + line.id">{{ line.name }}</span>
              <svg-icon
                icon-class="x"
                class="icon delete-line"
                :size="0"
                v-show="isLineRoute(line.id)"
                v-on:click.native.stop.prevent="
                  // 需要使用 native ，否则组件无法监听事件
                  deleteLine(line.id);
                  $event.stopPropagation();
                "
              />
            </router-link>
          </li>
          <li>
            <a v-on:click="newLineClick" class="newline">
              <svg-icon icon-class="speakers" class="icon" :size="0" />
              <span v-show="!newLine">新增</span>
              <input
                type="text"
                id="newline-input-name"
                class="line-name"
                maxlength="10"
                v-show="newLine"
                v-model="newLineName"
                :class="{
                  'animate__animated animate__headShake': newLineNameError,
                }"
                placeholder="名称"
                @keyup="inputLengthLimit($event, 10)"
                @keyup.enter="submitNewLine"
                @blur="submitNewLine"
              />
            </a>
          </li>
        </ul>
        <div class="navbar-end">
          <div
            class="navbar-item"
            :class="{ 'is-active': $route.name === 'settings' }"
          >
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
import {
  getLineList,
  listenLineChanged,
  createLine,
  deleteLine,
} from '@/api/line';
import { inputLengthLimit } from '@/common/input';

let errorAnimateTimeout = 0;

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
      lines: [],
      newLine: false,
      newLineName: '',
      newLineNameError: false,
    };
  },
  watch: {
    $route(to, from) {
      this.newLine = false;
      this.newLineNameError = false;
    },
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
    inputLengthLimit(evt, len) {
      return inputLengthLimit(evt.target, len);
    },
    newLineClick() {
      if (this.newLine) {
        document.getElementById('newline-input-name').focus();
        return;
      }
      this.newLine = true;
      if (errorAnimateTimeout) clearTimeout(errorAnimateTimeout);
      errorAnimateTimeout = setTimeout(() => {
        document.getElementById('newline-input-name').focus();
      }, 500);
    },
    deleteLine(id) {
      if (id === undefined) return;
      deleteLine(id, 0).then(() => {
        let i = 0;
        for (i = 0; i < this.lines.length; i++) {
          if (this.lines[i].id == id) {
            this.lines.splice(i, 1);
            break;
          }
        }
        if (this.lines.length === 0) {
          this.$router.replace('/speakers');
          return;
        }
        if (i > 0) i--;

        this.$router.replace('/line/' + this.lines[i].id);
      });
    },
    submitNewLine() {
      if (this.newLineName.length == 0) {
        this.newLineNameError = true;
        if (errorAnimateTimeout) clearTimeout(errorAnimateTimeout);
        errorAnimateTimeout = setTimeout(() => {
          this.newLineNameError = false;
        }, 500);
        return;
      }
      errorAnimateTimeout = 0;
      this.newLineNameError = false;
      this.newLine = false;
      createLine(this.newLineName).then(line => {
        this.lines.push(line);
      });
      this.newLineName = '';
    },
  },
};
</script>
<style lang="scss" scoped>
@import '../../node_modules/bulma/sass/utilities/_all.sass';

$nav-height: 90px !default;
$nav-item-width: 98px !default;

nav {
  position: relative;
  .border {
    height: 1px;
    flex: none;
    position: absolute;
    left: 0;
    width: 100%;
    bottom: 0;
    background: $border;
  }
}

.tabs {
  height: $nav-height;
  overflow: hidden;
  ul {
    overflow: hidden;
    overflow-x: auto;
    justify-content: flex-start;
    flex-grow: 1;
    flex-shrink: unset;
    height: $nav-height + 1;
    border: none;
    &::-webkit-scrollbar {
      display: none;
    }
    li {
      flex-basis: $nav-item-width;
      flex-shrink: 0;
      display: flex;
      flex-direction: column;
      width: $nav-item-width;
      height: 100%;
      background: #fafafa;
      a {
        height: 100%;
        text-decoration: none;
        flex-direction: column;
        position: relative;
        &:hover {
          border-color: $border;
        }

        &.newline {
          .line-name {
            font-size: 0.7rem;
            margin-top: 0.5rem;
            max-width: 80px;
            overflow: hidden;
            text-overflow: ellipsis;
          }
        }

        .delete-line {
          position: absolute;
          top: -3px;
          right: 9px;
          width: 0.8rem;
          color: $border;
          &:hover {
            color: $danger;
          }
        }
      }

      &.is-active {
        position: relative;
      }
      .svg-icon {
        width: 2rem;
        height: 2rem;
        margin: 0;
      }
    }
  }

  .buttons .button {
    border: 0;
    border-radius: 100%;
    padding: 8px;
    width: 3rem;
    height: 3rem;
    &.is-none:hover {
      background: none !important;
      cursor: auto;
    }
  }
  .navbar-item {
    height: 100%;
    position: relative;
    display: flex;

    &.is-active {
      border: none;
      position: relative;
      background: #fff;
      a:hover {
        background: $link-hover-border;
      }
    }
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
  .navbar-start,
  .navbar-end {
    border: none;
    display: flex;
    align-items: center;
    flex: none;
  }
  .navbar-start {
    box-shadow: 0.2rem 0 0.5em -0.5rem hsl(0deg 0% 29%);
  }
  .navbar-end {
    box-shadow: -0.2rem 0 0.5em -0.5rem hsl(0deg 0% 29%);
  }

  span {
    font-size: 0.7rem;
    margin-top: 0.5rem;
    max-width: 80px;
    overflow: hidden;
    text-overflow: ellipsis;
  }
}

@media screen and (max-width: 819px) {
  .tabs {
    .navbar-item {
      padding: 0;
    }
  }
}
</style>
