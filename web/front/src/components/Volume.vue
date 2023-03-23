<template>
  <div class="volume-controller"  @touchstart.stop @mousedown.stop>
    <div class="mute" :class="{ 'is-muted': isMute }" @click.stop.prevent="onVolumeMute()">
      <svg-icon :icon-class="isMute ? 'volume-mute' : 'volume'" :size="24"></svg-icon>
    </div>
    <vue-slider v-model="curVolume" :min="0" :max="100" :process="volumeLevelProcess"
      :tooltip-placement="this.tooltipPlacement" ref="volumeSlider" @change="throttleTimer"
      @drag-end="throttleTimer(curVolume)" />
  </div>
</template>

<script>
import VueSlider from 'vue-slider-component';
import 'vue-slider-component/theme/antd.css';
import { throttleFunction } from '@/common/throttle';

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
      throttleTimer: () => { }
    };
  },
  mounted() {
    this.curVolume = this.volume;
    this.isMute = this.mute;

    this.throttleTimer = throttleFunction(vol => {
      if (this.isMute) {
        this.isMute = false;
        this.$emit('mute', this.isMute);
      }
      this.$emit('change', vol);
    }, 200);
  },
  methods: {
    volumeLevelProcess(dotsPos) {
      return [[0, 0, {}]];
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
  align-items: center;

  .vue-slider {
    flex: 1 0 auto;
  }
}

.mute {
  margin-left: 0rem;
  margin-right: 1rem;
  cursor: pointer;
  line-height: 1rem;
}
</style>