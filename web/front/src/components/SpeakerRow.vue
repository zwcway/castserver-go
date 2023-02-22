<template>
  <div class="card" @click.stop="show = !show">
    <div class="card-content speaker">
      <div class="media">
        <div class="media-left">
          <figure class="image is-48x48">
            <svg-icon icon-class="speaker" />
          </figure>
        </div>
        <div class="media-content">
          <div class="columns is-vcentered">
            <div class="column">
              <p class="title is-5">
                <router-link :to="`/speaker/${speaker.id}`">
                  {{ speaker.name }}
                </router-link>
              </p>
              <p class="subtitle is-6">
                <span>{{ speaker.ip }}</span>
                <span class="ratebits">{{ showRateBits(speaker) }}</span>
              </p>
            </div>
            <div class="column" v-on:click.stop="">
              <vue-slider
                v-model="volume"
                :min="0"
                :max="100"
                :process="volumeLevelProcess"
                ref="volumeSlider"
                @change="volumeChanged"
                @drag-end="volumeChanged('finally')"
              />
            </div>
          </div>
          <div class="content" v-if="show">
            <p>content</p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import VueSlider from 'vue-slider-component';
import 'vue-slider-component/theme/antd.css';
import { setVolume as setSpeakerVolume } from '@/api/speaker';
import { throttleFunction } from '@/common/throttle';
import { formatRate, formatBits } from '@/common/format';
import '@/assets/css/speaker.scss';

let throttleTimer;

export default {
  components: {
    VueSlider,
  },
  props: {
    speaker: { type: Object, required: true },
  },
  emits: ['domReady'],
  data() {
    let that = this;
    return {
      show: false,
      volumeLevelProcess(dotsPos) {
        return [[0, 0, { backgroundColor: 'pink' }]];
      },
    };
  },
  watch: {},
  mounted() {
    let that = this;
    let si = setInterval(() => {
      if (that.$refs.volumeSlider) {
        that.$emit(
          'domReady',
          that.$refs.volumeSlider.$el.querySelector('.vue-slider-process')
        );
        clearInterval(si);
      }
    }, 100);

    throttleTimer = throttleFunction(vol => {
      setSpeakerVolume(this.speaker.id, vol);
    }, 200);
  },
  computed: {
    volume: {
      get() {
        return this.speaker.volume || 0;
      },
      set(value) {
        this.speaker.volume = value;
      },
    },
  },
  methods: {
    volumeChanged(v) {
      if (v === 'finally') return throttleTimer.finally();
      throttleTimer(v);
    },
    getTitleLink(item) {
      return `/${this.type}/${item.id}`;
    },
    getLineLink(item) {
      return `/line/${item.line.id}`;
    },
    getChannelLink(item) {
      return `/channel/${item.channel.id}`;
    },
    showRateBits(sp) {
      return formatRate(sp.rate) + '/' + formatBits(sp.bits)
    }
  },
};
</script>

<style lang="scss">
@import '../../node_modules/bulma/sass/utilities/derived-variables.sass';

.media-content {
  overflow: visible;
}
.vue-slider-dot-handle {
  border-color: $primary;
}
.vue-slider-dot-handle:hover {
  border-color: $scheme-main;
  background-color: $primary;
}

</style>
