<template>
  <div id="app" :class="{ 'user-select-none': userSelectNone }">
    <Navbar v-show="showNavbar" ref="navbar" class="navbar" />
    <Scrollbar ref="scrollbar" class="scrollbar" />
    <main
      ref="main"
      :style="{ overflow: enableScrolling ? 'auto' : 'hidden' }"
      class="main-body"
      @scroll="handleScroll"
    >
      <router-view
        v-if="!$route.meta.keepAlive && isRouteViewAlive"
      ></router-view>
      <keep-alive>
        <router-view
          v-if="$route.meta.keepAlive && isRouteViewAlive"
        ></router-view>
      </keep-alive>
    </main>
    <Debug />
    <Toast />
  </div>
</template>

<script>
import Scrollbar from '@/components/Scrollbar.vue';
import Navbar from '@/components/Navbar.vue';
import Debug from '@/components/Debug.vue';
import Toast from '@/components/Toast.vue';
import { mapState } from 'vuex';
import { socket } from '@/common/request';
import { changeAppearance } from '@/common/theme';

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
      isRouteViewAlive: true,
      userSelectNone: false,
    };
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
    this.fetchData();
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
    fetchData() {
      // this.$store.dispatch('fetchLikedSongs');
      // this.$store.dispatch('fetchLikedSongsWithDetails');
      // this.$store.dispatch('fetchLikedPlaylist');
    },
    handleScroll() {
      this.$refs.scrollbar.handleScroll();
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
    height: $nav-height;
    left: 0;
    display: flex;
  }
  .main-body {
    top: $nav-height;
    bottom: 0;
    width: 100%;
    position: fixed;
    // margin-top: $nav-height;
    scrollbar-width: none; /* firefox */
    -ms-overflow-style: none; /* IE 10+ */
    &::-webkit-scrollbar {
      display: none; /* Chrome Safari */
    }
  }
}
</style>
