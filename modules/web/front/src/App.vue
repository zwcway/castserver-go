<template>
  <div id="app" :class="{ 'user-select-none': userSelectNone }">
    <Navbar v-show="showNavbar" ref="navbar" class="navbar" />
    <Scrollbar ref="scrollbar" class="scrollbar" />
    <main ref="main" :style="{ overflow: enableScrolling ? 'auto' : 'hidden' }" class="main-body" @scroll="handleScroll"
      v-touch:swipe.left='onSwipeLeft' v-touch:swipe.right="onSwipeRight">
      <Transition :name="transitionName">
        <router-view v-if="!$route.meta.keepAlive && isRouteViewAlive" :key="$route.path" />
        <keep-alive v-if="$route.meta.keepAlive && isRouteViewAlive">
          <router-view :key="$route.path" />
        </keep-alive>
      </Transition>
    </main>
    <Debug />
    <Toast />
  </div>
</template>

<script>
import { mapState } from 'vuex';
import Scrollbar from '@/components/Scrollbar.vue';
import Navbar from '@/components/Navbar.vue';
import Debug from '@/components/Debug.vue';
import Toast from '@/components/Toast.vue';
import { socket } from '@/common/request';
import { changeAppearance } from '@/common/theme';

document.addEventListener('gesturestart', (event) => {
  // 禁止 ios safari 缩放
  event.preventDefault();
});
export default {
  name: 'App',
  components: {
    Navbar,
    Toast,
    Scrollbar,
    Debug,
  },
  provide() {
    return {
      reload: this.reload,
    };
  },
  data() {
    return {
      transitionName: '',
      isRouteViewAlive: true,
      userSelectNone: false,

    };
  },
  watch: {
    $route(to) {
      const dir = to.params.dir;
      if (dir !== undefined) {
        this.transitionName = dir;
      } else {
        this.transitionName = '';
      }
    }
  },
  computed: {
    ...mapState(['settings', 'enableScrolling']),
    showNavbar() {
      return true;
    },
  },
  created() {
    socket.connect();
    changeAppearance(this.settings.appearance || 'auto');
    window.addEventListener('keydown', this.handleKeydown);
  },
  methods: {
    reload() {
      this.isRouteViewAlive = false;
      this.$nextTick(() => {
        this.isRouteViewAlive = true;
      });
    },
    handleKeydown(e) {
      if (e.code === 'Space') {
        if (e.target.tagName === 'INPUT') return false;
        if (this.$route.name === 'mv') return false;
        e.preventDefault();
      }
    },
    handleScroll() {
      this.$refs.scrollbar.handleScroll();
    },
    onEnter(el, done) {
      done()
    },
    onLeave(el) {

    },
    onSwipe(e) {
      console.log(e)
    },
    onSwipeLeft() {
      this.$refs.navbar.toRouteLeftOrRight('left')
    },
    onSwipeRight() {
      this.$refs.navbar.toRouteLeftOrRight('right')
    },
  },
};
</script>

<style lang="scss">
#app {
  height: 100%;
  position: relative;

  .navbar {
    top: 0;
    width: 100%;
    height: var(--size-navbar);
    left: 0;
    display: flex;
  }

  .main-body {
    top: var(--size-navbar);
    bottom: 0;
    width: 100%;
    position: fixed;
    // margin-top: var(--size-navbar);
    scrollbar-width: none;
    /* firefox */
    -ms-overflow-style: none;

    /* IE 10+ */
    &::-webkit-scrollbar {
      display: none;
      /* Chrome Safari */
    }

    >div {
      width: 100%;
      height: fit-content;
      position: absolute;
      max-width: 1024px;
      left: 0;
      right: 0;
      margin: auto;

    }
  }

  .slide-to-left-enter-active {
    animation: 500ms slideInRight;
  }

  .slide-to-left-leave-active {
    animation: 500ms slideOutLeft;
  }

  .slide-to-right-enter-active {
    animation: 500ms slideInLeft;
  }

  .slide-to-right-leave-active {
    animation: 500ms slideOutRight;
  }

}
</style>
