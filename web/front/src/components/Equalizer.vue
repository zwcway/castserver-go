<template>
  <div>
    <div class="equalizer">
      <div class="name"></div>
      <div class="body">
        <div class="axis-db"></div>
        <div class="sliders" :class="`band-${eqbands}`">
          <div class="slider" v-for="(eq, i) in equalizes" :key="i">
            <vue-slider
              v-model="eq[2]"
              :direction="`btt`"
              :min="-15"
              :max="15"
              :process="gainProcess"
              @change="gainChanged(i, $event)"
              @drag-end="gainChanged(i, 'finally')"
            />
            <label for="">{{ label(eq[0]) }}</label>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import VueSlider from 'vue-slider-component';
import 'vue-slider-component/theme/antd.css';
import { throttleFunction } from '@/common/throttle';

let throttleTimer;

export default {
  name: 'Equalizer',
  components: {
    VueSlider,
  },
  emits: ['change'],
  props: {
    eq: { type: Array, required: false },
    bands: { type: Number, required: false },
  },
  data() {
    return {
      show: false,
      eqbands: 15,
      defaultEQs: {
        15: [
          [30, 1, 0],
          [60, 1, 0],
          [200, 1, 0],
          [300, 1, 0],
          [360, 1, 0],
          [470, 1, 0],
          [730, 1, 0],
          [1200, 1, 0],
          [1800, 1, 0],
          [2900, 1, 0],
          [4800, 1, 0],
          [7000, 1, 0],
          [11000, 1, 0],
          [15000, 1, 0],
          [19000, 1, 0],
        ],
        31: [
          [40, 1, 0],
          [100, 1, 0],
          [200, 1, 0],
          [260, 1, 0],
          [360, 1, 0],
          [460, 1, 0],
          [570, 1, 0],
          [640, 1, 0],
          [750, 1, 0],
          [840, 1, 0],
          [940, 1, 0],
          [1000, 1, 0],
          [1100, 1, 0],
          [1200, 1, 0],
          [1300, 1, 0],
          [1400, 1, 0],
          [1500, 1, 0],
          [2000, 1, 0],
          [3000, 1, 0],
          [4100, 1, 0],
          [5100, 1, 0],
          [6100, 1, 0],
          [7000, 1, 0],
          [8000, 1, 0],
          [9000, 1, 0],
          [10000, 1, 0],
          [12000, 1, 0],
          [14000, 1, 0],
          [16000, 1, 0],
          [18000, 1, 0],
          [20000, 1, 0],
        ],
      },
      equalizes:[],
      gainProcess(dotsPos) {
        return [[0, dotsPos, { backgroundColor: 'pink' }]];
      },
    };
  },
  watch: {
    eq(newVal, oldVal) {
      this.setNewEQ(newVal)
    },
    bands: {
      handler() {
        if (this.bands in this.defaultEQs) {
          this.eqbands = this.bands;
          this.setNewEQ(this.eq)
        }
      },
    },
  },
  mounted() {
    this.setNewEQ(this.eq)
    throttleTimer = throttleFunction((i, gain) => {
      this.setGain(i, gain);
    }, 200);
  },
  methods: {
    setNewEQ(newVal) {
      this.equalizes = Array.from(this.defaultEQs[this.eqbands]);
      if (newVal === undefined) return;
      newVal.forEach(e => {
        let freq = 0;
        if (!(e instanceof Array)) {
          freq = e;
        } else if (e.length != 3) {
            return;
        }
        freq = e[0];

        for(let i = 0; i < this.equalizes.length; i ++) {
          if (this.equalizes[i][0] == freq) {
            this.equalizes[i][1] = e[2];
            this.equalizes[i][2] = e[1];
            this.equalizes = this.equalizes;
            return;
          }
        }
      })
    },
    bandList() {
      let keys = [];
      for (let k in this.defaultEQs) {
        keys.push(k);
      }
      return keys;
    },
    gainChanged(i, v) {
      if (v === 'finally') return throttleTimer.finally(this.equalizes[i][2]);
      this.equalizes[i][2] = v;
      this.equalizes = this.equalizes;
      throttleTimer(i, v);
    },
    setGain(i, gain) {
      let eq = this.equalizes[i];
      if (eq === undefined) return;
      this.$emit('change', eq[0], gain);
    },
    label(freq) {
      return freq >= 1000
        ? ((freq / 1000).toFixed(1) + 'k').replace('.0', '')
        : '' + freq;
    },
  },
};
</script>

<style lang="scss" scoped>
@import '../../node_modules/bulma/sass/utilities/_all.sass';

.equalizer {
  display: flex;
  margin: auto;
  background: $white;
  .body {
    display: flex;
  }
  .sliders {
    display: flex;
    flex-direction: row;
    justify-content: center;
    align-content: flex-start;
    > .slider {
      display: flex;
      flex-direction: column;
      flex: 0 0 2rem;
      width: 2rem;
      &.band-21 {
        flex-basis: 1.8rem;
        width: 1.8rem;
      }
      &.band-31 {
        flex-basis: 1.5rem;
        width: 1.5rem;
      }
      .vue-slider {
        flex-basis: 150px;
      }
      > label {
        font-size: 0.5rem;
      }
    }
  }
}
</style>
