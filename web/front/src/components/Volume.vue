<template>
  <div class="volume-controller">
    <div
      class="mute"
      :class="{ 'is-muted': isMute }"
      @click.stop.prevent="onVolumeMute()"
    >
      <svg-icon
        :icon-class="isMute ? 'volume-mute' : 'volume'"
        :size="16"
      ></svg-icon>
    </div>
    <vue-slider
      v-model="curVolume"
      :min="0"
      :max="100"
      :process="volumeLevelProcess"
      :disabled="isMute"
      :tooltip-placement="this.tooltipPlacement"
      ref="volumeSlider"
      @change="throttleTimer"
      @drag-end="throttleTimer(curVolume)"
    />
  </div>
</template>

<script>
import VueSlider from 'vue-slider-component';
import 'vue-slider-component/theme/antd.css';
import { throttleFunction } from '@/common/throttle';

let throttleTimer;
export default {
  components: {
    VueSlider,
  },
  props: {
    volume: { type: Number, required: true },
    mute: { type: Boolean, required: true },
    tooltipPlacement: { type: String, required: false },
  },
  emits: ['mute', 'change'],
  watch: {
    volume(newVal, oldVal) {
      this.curVolume = newVal;
    },
    mute(newVal, oldVal) {
      this.isMute = newVal;
    }
  },
  data() {
    return {
      isMute: false,
      curVolume: 0,
      throttleTimer: () => {}
    };
  },
  mounted() {
    this.curVolume = this.volume;
    this.isMute = this.mute;

    this.throttleTimer = throttleFunction(vol => {
      this.$emit('change', vol);
    }, 200);
  },
  methods: {
    volumeLevelProcess(dotsPos) {
      return [[0, 0, { backgroundColor: 'pink' }]];
    },
    onVolumeMute() {
      this.isMute = !this.isMute;
      this.$emit('mute', this.isMute);
    },
  },
};
</script>

<style lang="scss" scoped>
.volume-controller {
  display: flex;
  flex-direction: row;
  flex-wrap: nowrap;
  justify-items: start;
  .vue-slider {
    flex: 1 0 auto;
  }
}

.mute {
      margin-left: 2rem;
      margin-right: 1rem;
      cursor: pointer;
    }

</style>