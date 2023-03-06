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
                <router-link class="speaker-name" :to="`/speaker/${spInfo.id}`">
                  {{ spInfo.name }}
                </router-link>
                <router-link
                  v-if="spInfo.line"
                  class="line-name"
                  :to="`/line/${spInfo.line.id}`"
                >
                  <a-button type="link">{{ spInfo.line.name }}</a-button>
                </router-link>
              </p>
              <p class="subtitle is-6">
                <span>{{ spInfo.ip }}</span>
                <span class="ratebits">{{ showRateBits(speaker) }}</span>
              </p>
            </div>
            <div class="column" v-on:click.stop="">
              <Volume
                :volume="volume"
                :mute="mute"
                @change="setSpeakerVolume"
                @mute="setSpeakerMute"
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
import Volume from '@/components/Volume';
import { setVolume as setSpeakerVolume, setSpeaker } from '@/api/speaker';
import { formatRate, formatBits } from '@/common/format';
import '@/assets/css/speaker.scss';

export default {
  components: {
    VueSlider,
    Volume,
  },
  props: {
    speaker: { type: Object, required: true },
  },
  data() {
    let that = this;
    return {
      show: false,
      spInfo: {},
      volumeLevelProcess(dotsPos) {
        return [[0, 0, { backgroundColor: 'pink' }]];
      },
    };
  },
  watch: {
    speaker(newVal, oldVal) {
      this.spInfo = newVal;
    },
  },
  mounted() {
    this.spInfo = this.speaker;
  },
  computed: {
    volume: {
      get() {
        return this.spInfo.vol || 0;
      },
      set(value) {
        this.spInfo.vol = value;
      },
    },
    mute: {
      get() {
        return this.spInfo.mute || false;
      },
      set(value) {
        this.$set(this.spInfo, 'mute', value);
      },
    },
  },
  methods: {
    setSpeakerVolume(v) {
      setSpeakerVolume(this.spInfo.id, v)
        .then(() => {
          this.$set(this.spInfo, 'vol', v);
        })
        .catch(() => {
          this.$set(this.spInfo, 'vol', this.volume);
        });
    },
    setSpeakerMute(v) {
      setSpeaker(this.spInfo.id, 'mute', v)
        .then(() => {
          this.$set(this.spInfo, 'mute', v);
        })
        .catch(() => {
          this.$set(this.spInfo, 'mute', this.mute);
        });
    },
    getTitleLink(item) {
      return `/${this.type}/${item.id}`;
    },
    getLineLink(item) {
      return `/line/${item.line.id}`;
    },
    getChannelLink(item) {
      if (!item.channel) {
        return '';
      }
      return `/channel/${item.channel.id}`;
    },
    showRateBits(sp) {
      return formatRate(sp.rate) + '/' + formatBits(sp.bits);
    },
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
.line-name {
  align-self: flex-end;
}
</style>
