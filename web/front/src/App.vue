<template>
  <div id="app" :class="{ 'user-select-none': userSelectNone }">
    <Scrollbar ref="scrollbar" />
    <Navbar v-show="showNavbar" ref="navbar" />
    <main
      ref="main"
      :style="{ overflow: enableScrolling ? 'auto' : 'hidden' }"
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
    <Toast />
  </div>
</template>

<script>
import Scrollbar from './components/Scrollbar.vue';
import Navbar from './components/Navbar.vue';
import Toast from './components/Toast.vue';
import { mapState } from 'vuex';
import { socket } from '@/common/request';

socket.connect();

export default {
  name: 'App',
  components: {
    Navbar,
    Toast,
    Scrollbar,
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
    ...mapState(['enableScrolling']),
    showNavbar() {
      return true;
    },
  },
  created() {
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

<style lang="scss"></style>
