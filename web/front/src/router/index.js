import Vue from 'vue';
import VueRouter from 'vue-router';

Vue.use(VueRouter);

const routes = [
  {
    path: '/speakers',
    name: 'speakers',
    component: () => import('@/views/speakers.vue'),
    meta: {
      title: '扬声器列表',
      keepAlive: false,
      savePosition: true,
    },
  },
  {
    path: '/speaker/:id',
    name: 'speaker',
    component: () => import('@/views/speaker.vue'),
    meta: {
      keepAlive: false,
      savePosition: true,
    },
  },
  {
    path: '/line/create',
    name: 'newline',
    component: () => import('@/views/newline.vue'),
    props: true,
    meta: {
      keepAlive: false,
      savePosition: true,
    },
  },
  {
    path: '/line/:id/:action?/:spid?',
    name: 'line',
    component: () => import('@/views/line.vue'),
    props: true,
    meta: {
      keepAlive: true,
      savePosition: true,
    },
  },
  {
    path: '/css3gen',
    name: 'css3gen',
    component: () => import('@/views/css3gen.vue'),
  },
  {
    path: '/settings',
    name: 'settings',
    component: () => import('@/views/settings.vue'),
    meta: {
      keepAlive: true,
      savePosition: true,
    },
    props: {
      forceTo: '',
    },
  },
  {
    path: '*',
    redirect: '/speakers',
    name: 'notFound',
    hidden: true,
  },
];

const router = new VueRouter({
  routes,
});

export default router;
