<template>
  <div>
    <nav :class="{ 'has-custom-titlebar': hasCustomTitlebar }">
      <Win32Titlebar v-if="enableWin32Titlebar" />
      <LinuxTitlebar v-if="enableLinuxTitlebar" />
      <div class="border"></div>
      <div class="tabs is-large is-boxed">
        <div class="navbar-list navbar-start">
          <div class="navbar-item">
            <div class="buttons">
              <a class="button is-normal is-none is-inverted">
                <svg-icon icon-class="logo" :size="24" />
              </a>
              <a class="link-state button is-normal is-none is-inverted" :class="{
                'is-danger': !wsConnected,
                'is-primary': wsConnected,
              }" :title="wsConnected ? $t('server connected') : $t('server disconnected')">
                <svg-icon :icon-class="wsConnected ? 'link' : 'unlink'" :size="0" />
              </a>
            </div>
          </div>
        </div>
        <ul class="navbar-list">
          <li class="navbar-item" :class="{ 'is-active': $route.name === 'speakers' }">
            <a :href="$router.resolve({ name: 'speakers' }).href" @click="reload">
              <a-badge count="5" show-zero>
                <svg-icon icon-class="speakers" class="icon" :size="0" />
              </a-badge>
              <span class="name">{{ $t('speaker') }}</span>
            </a>
          </li>
          <li class="navbar-item" v-for="(line, i) in lines" :key="line.id"
            :class="{ 'is-active': isLineRoute(line.id) }">
            <a @click="toRouteLine(i, line)">
              <svg-icon icon-class="music" class="icon" :size="0" />
              <span class="name" :id="'nav-' + line.id">{{ line.name }}</span>
              <svg-icon icon-class="x" class="icon delete-line" :size="0" v-show="!line.def && isLineRoute(line.id)"
                v-on:click.native.stop.prevent="
                  // 需要使用 native ，否则组件无法监听事件
                  deleteLine(line.id);
                $event.stopPropagation();                                                                                                                                                                                                                                               " />
            </a>
          </li>
          <li class="navbar-item">
            <a v-on:click="newLineClick" class="newline">
              <svg-icon icon-class="speakers" class="icon" :size="0" />
              <span class="name" v-show="!newLine">{{ $t('new') }}</span>
              <input type="text" id="newline-input-name" class="line-name" maxlength="10" v-show="newLine"
                v-model="newLineName" :class="{
                  'animate__animated animate__headShake': newLineNameError,
                }" placeholder="名称" @keyup="inputLengthLimit($event, 10)" @keyup.enter="submitNewLine"
                @blur="submitNewLine" />
            </a>
          </li>
        </ul>
        <div class="navbar-list navbar-end">
          <div class="navbar-item" :class="{ 'is-active': $route.name === 'settings' }">
            <div class="buttons">
              <router-link to="/settings" class="button is-primary is-normal"
                :class="{ 'is-outlined': $route.name !== 'settings' }" :title="$t('settings')">
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
  listenLineListChanged,
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
    this.$nextTick(function () {
      document.addEventListener(
        'keyup',
        (this.onKeyUp = e => {
          if (e.key === 'Escape') {
            this.newLine = false;
          }
        })
      );
    });
  },
  methods: {
    loadData() {
      getLineList().then(data => {
        this.lines = data;
      });
      listenLineListChanged(() => {
        getLineList().then(data => {
          this.lines = data;
        });
      });
    },
    go(where) {
      if (where === 'back') this.$router.go(-1);
      else this.$router.go(1);
    },
    toRouteLine(fi, line) {
      const id = line.id;
      if (this.$route.name === 'line') {
        const fid = parseInt(this.$route.params.id)
        let fi = -1, ti;
        let dir = '';
        for (let i = 0; i < this.lines.length; i++) {
          if (this.lines[i].id === fid)
            fi = i
          else if (this.lines[i].id === line.id)
            ti = i
        }
        if (fi < ti) dir = 'slide-to-left';
        else if (fi > ti) dir = 'slide-to-right';

        this.$router.push({ name: 'line', params: { id, dir } });
        return
      }
      this.$router.push({ path: `/line/${line.id}` });
    },
    toRouteLeftOrRight(dir) {
      if (this.$route.name !== 'line') return
      const fid = parseInt(this.$route.params.id)
      let fi = -1
      for (let i = 0; i < this.lines.length; i++) {
        if (this.lines[i].id === fid) {
          fi = i
          break
        }
      }
      if (fi < 0) return

      if (dir === 'left' && fi < this.lines.length - 1) {
        this.toRouteLine(fi, this.lines[fi + 1])
      } else if (dir === 'right' && fi > 0) {
        this.toRouteLine(fi, this.lines[fi - 1])
      }
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
      if (!id) {
        this.$alert();
      }
      let that = this
      this.$confirm({
        title: '确定要移除该线路吗？',
        content: h =>
          h(
            'div',
            { style: 'color:red;' },
            '移除线路后，该线路下的所有扬声器将自动移动至默认线路中。'
          ),
        okText: '是',
        okType: 'danger',
        cancelText: '否',
        onOk() {
          deleteLine(id, 0).then(() => {
            let i = 0;
            for (i = 0; i < that.lines.length; i++) {
              if (that.lines[i].id == id) {
                that.lines.splice(i, 1);
                break;
              }
            }
            that.lines = that.lines
            if (that.lines.length === 0) {
              that.$router.replace('/speakers');
              return;
            }
            if (i > 0) i--;

            that.$router.replace('/line/' + that.lines[i].id);
          });
        },
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
<style lang="scss">
@import '@/assets/css/function.scss';

nav {
  position: relative;
  display: flex;
  width: 100%;

  .border {
    height: 1px;
    flex: none;
    position: absolute;
    left: 0;
    width: 100%;
    bottom: 0;
    background: var(--color-border);
  }
}

.tabs {
  display: flex;
  height: var(--size-navbar);
  overflow: hidden;
  width: 100%;

  .navbar-list {
    display: flex;
    overflow: hidden;
    overflow-x: auto;
    justify-content: flex-start;
    flex-grow: 1;
    flex-shrink: unset;
    height: var(--size-navbar);
    border: none;

    &::-webkit-scrollbar {
      display: none;
    }

    .navbar-item {
      flex-basis: var(--size-navbar-item);
      flex-shrink: 0;
      display: flex;
      flex-direction: column;
      width: var(--size-navbar-item);
      height: var(--size-navbar);
      background: var(--color-navbar-bg);
      border: 1px solid var(--color-border);
      border-right-color: var(--color-navbar-bg);
      border-top-color: var(--color-navbar-bg);

      &:last-child {
        border-right-color: var(--color-border);
      }

      &:first-child {
        border-left-color: var(--color-navbar-bg);
      }

      a {
        height: auto;
        text-decoration: none;
        flex-direction: column;
        position: relative;
        display: flex;
        width: 100%;
        justify-items: center;
        align-items: center;
        padding: 1rem 0;
        text-align: center;
        color: var(--color-text);

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
          top: 3px;
          right: 5px;
          width: 0.8rem;
          height: 1rem;
          color: var(--color-secondary);

          &:hover {
            color: var(--color-danger);
          }
        }

        .ant-badge-count {
          background-color: var(--color-body-bg);
          color: var(--color-text);
          box-shadow: 0 0 0 1px var(--color-border-hover) inset;
        }
        .name {
          font-size: 0.7rem;
          margin-top: 0.5rem;
          max-width: calc(var(--size-navbar-item) - 15px);
          overflow: hidden;
          text-overflow: ellipsis;
          white-space: nowrap;
          width: 100%;
        }
      }

      &:hover {
        a {
          color: var(--color-border-hover);
        }
      }

      &.is-active {
        position: relative;
        border-top-left-radius: 10px;
        border-top-right-radius: 10px;
        border-top-color: var(--color-border);
        border-bottom-color: var(--color-navbar-bg);
        border-right-color: var(--color-border);

        a {
          color: var(--color-border-hover);
          .name {
            color: var(--color-border-hover);
          }
        }

        &+.navbar-item {
          border-left-color: var(--color-navbar-bg);
        }
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

  .navbar-list {
    .navbar-item {
      .buttons {
        position: relative;
        display: flex;
        justify-items: center;
        align-items: center;
        width: 100%;
      }

      .link-state {
        width: 1rem;
        height: 1rem;
        padding: 0;
        position: absolute;
        bottom: 0.5rem;
        right: 0.5rem;

        .svg-icon {
          width: 100% !important;
          height: 100% !important;
        }
      }
    }

    &.navbar-start,
    &.navbar-end {
      border: none;
      display: flex;
      align-items: center;
      flex: none;
      width: 60px;
      z-index: 1;

      .navbar-item {
        width: 100%;
        flex: none;
        flex-direction: row;
        justify-content: center;
      }
    }

    &.navbar-start {
      box-shadow: 0.2rem 0 1rem -0.5rem var(--color-primary-bg-for-transparent);

      .navbar-item {
        border-left: none;
        border-top-left-radius: 0;

        &:last-child {
          border-right-color: var(--color-border);
        }
      }
    }

    &.navbar-end {
      box-shadow: -0.2rem 0 1rem -0.5rem var(--color-primary-bg-for-transparent);

      .navbar-item {
        border-right: none;
        border-top-right-radius: 0;

        &:first-child {
          border-left-color: var(--color-border);
        }
      }
    }

  }

}

@include for_breakpoint(mobile) {
  #app {
    --size-navbar: 60px;
    --size-navbar-item: 80px;
  }

  .navbar {
    .navbar-list {
      .navbar-item {
        padding: 0;

        a {
          padding: 10px 0;
        }

        .svg-icon {
          width: 16px;
          height: 16px;
        }
      }
    }
  }
}
</style>
