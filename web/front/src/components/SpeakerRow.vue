<template>
  <div class="speaker" :class="spInfo.cTime > 0 ? 'connected' : 'disconnected'">
    <div class="speaker-name">
      <svg-icon icon-class="speaker" :size="0" />
      <router-link class="name" :to="`/speaker/${spInfo.id}`">
        {{ spInfo.name }}
      </router-link>
      <a-button type="link" v-if="spInfo.line" class="line-name" @click.stop="$router.push(`/line/${spInfo.line.id}`)">
        {{ spInfo.line.name }}
      </a-button>
    </div>
    <div class="speaker-info">
      <svg-icon :icon-class="spInfo.cTime > 0 ? 'link' : 'unlink'" :size="0"
        :class="spInfo.cTime > 0 ? 'is-primary' : 'is-danger'" />
      <span class="ip">{{ spInfo.ip }}</span>
      <span class="ratebits">{{ showRateBits(speaker) }}</span>
    </div>
    <div class="speaker-volume level-meter-slider" v-on:click.stop="">
      <Volume :volume="volume" :mute="mute" @change="setSpeakerVolume" @mute="setSpeakerMute" />
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
        return [[0, 0, {}]];
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
    gotoSpeaker(id) {
      this.$router.push('/speaker/' + id);
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
    showRateBits(spInfo) {
      return formatRate(spInfo.rate) + '/' + formatBits(spInfo.bits);
    },
  },
};
</script>

<style lang="scss" scoped>
@import "@/assets/css/function.scss";

.speaker {
  border: 0;

  .speaker-volume.level-meter-slider {
    width: auto;
    flex: unset;
    flex-wrap: nowrap;
    flex: 1 0 12rem;
    // max-width: 20rem;
    display: flex;
    align-items: center;

    .volume-controller {
      width: 100%;
    }
  }


  .vue-slider-dot-handle {
    border-color: var(--color-border);
  }

  .vue-slider-dot-handle:hover {
    border-color: var(--color-primary);
    background-color: var(--color-primary-bg);
  }

  .line-name {
    align-self: flex-end;
  }

  .subtitle {
    span {
      padding: 0 3px;
    }
  }

  .not-connect {
    .speaker-icon {
      .svg-icon {
        color: grey;
      }

      .codicon {
        color: red;
        position: absolute;
        top: 0rem;
        left: 0rem;
        font-size: 2rem;
      }
    }

    .connect-info {
      color: red;
    }

    .speaker {
      background-color: #efefef;
    }
  }

  @include for_breakpoint(mobile) {

    .not-connect {
      .speaker-icon {
        .codicon {
          font-size: 2rem;
        }
      }
    }
  }

}
</style>
