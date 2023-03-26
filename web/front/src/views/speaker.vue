<template>
  <div class="container">
    <div class="card">
      <div class="volume">
        <span>{{ speaker.name }}</span>
        <vue-slider v-model="volume" :min="0" :max="100" :process="volumeLevelProcess" :tooltip-placement="'bottom'"
          ref="volumeSlider" @change="volumeChanged" @drag-end="volumeChanged('finally')" />
        <div> </div>
      </div>
    </div>
    <div class="notification">
      <div class="columns is-mobile">
        <div class="column">
          <label>MAC</label>
          <span>{{ speaker.mac }}</span>
        </div>
        <div class="column">
          <label>位置</label>
          <span>
            <a-select size="small" v-model="line" :options="lineListOptions" placeholder="选择线路"
              :dropdownMatchSelectWidth="false" />
            <a-select size="small" v-model="channel" :options="channelListOptions" placeholder="选择声道"
              :dropdownMatchSelectWidth="false" />
          </span>
        </div>
        <div class="column">
          <label>连接状态</label>
          <span>{{ speaker.cTime ? '已连接' : '未连接' }}</span>
        </div>
        <div class="column block">
          <label>支持的采样率</label>
          <span class="tags">
            <span class="tag" v-for="rate in speaker.rateList" :key="rate">
              {{ rate }}
            </span>
          </span>
        </div>
        <div class="column block">
          <label>支持的位宽</label>
          <span class="tags">
            <span class="tag" v-for="bit in speaker.bitsList" :key="bit">
              {{ bit }}bit
            </span>
          </span>
        </div>
        <div class="column">
          <label>队列中</label>
          <span>{{ speaker.statistic ? speaker.statistic.queue : 0 }}B</span>
        </div>
        <div class="column">
          <label>已发送</label>
          <span>{{ speaker.statistic ? speaker.statistic.send : 0 }}B</span>
        </div>
        <div class="column">
          <label>已丢弃</label>
          <span>{{ speaker.statistic ? speaker.statistic.drop : 0 }}B</span>
        </div>
      </div>
    </div>
    <div>
      <div id="spectrum">频谱</div>
    </div>
  </div>
</template>

<script>
import VueSlider from 'vue-slider-component';
import 'vue-slider-component/theme/antd.css';
import { throttleFunction } from '@/common/throttle';
import * as ApiLine from '@/api/line';
import * as ApiSpeaker from '@/api/speaker';
import { socket } from '@/common/request';

let throttleTimer;

export default {
  name: 'Speaker',
  components: {
    VueSlider,
  },
  data() {
    return {
      id: 0,
      speaker: {},
      lineList: [],
      volumeLevelProcess(dotsPos) {
        return [[0, 0, { backgroundColor: 'pink' }]];
      },
    };
  },
  computed: {
    volume: {
      get() {
        return this.speaker.vol;
      },
      set(value) {
        this.speaker.volume = value;
      },
    },
    channel: {
      get() {
        return this.speaker.ch > 0 ? this.speaker.ch + '' : '-1';
      },
      set(ch) {
        ApiSpeaker.setSpeaker(this.speaker.id, 'ch', parseInt(ch)).then(() => {
          this.speaker.ch = ch
        })
      }
    },
    channelListOptions() {
      let opts = [{key: '-1', label: '选择声道'}]
      for (let i in ApiLine.channelList) {
        opts.push({
          key: i + '',
          label: ApiLine.channelList[i].name
        })
      }
      return opts
    },
    lineListOptions() {
      let opts = [{key: '-1', label: '选择线路'}]
      for (let i in this.lineList) {
        opts.push({
          key: this.lineList[i].id + '',
          label: this.lineList[i].name
        })
      }
      return opts
    },
    line: {
      get() {
        return '' + (this.speaker.line ? this.speaker.line.id : -1);
      },
      set(nl) {
        ApiSpeaker.setSpeaker(this.speaker.id, 'line', parseInt(nl)).then(() => {
          this.loadData()
        })
      }
    },
  },
  mounted() {
    if (this.$route.params.id === undefined) {
      this.$router.replace('/speakers');
      return;
    }
    this.id = parseInt(this.$route.params.id || 0);
    socket.onConnected().then(() => this.loadData());
  },

  methods: {
    loadData() {
      ApiLine.getLineList().then(l => {
        this.lineList = l
      })
      ApiSpeaker.getSpeakerInfo(this.id)
        .then(data => {
          this.speaker = data;
        })
        .catch(code => {
          this.$router.replace('/speakers');
        });

      throttleTimer = throttleFunction(vol => {
        ApiSpeaker.setVolume(this.speaker.id, vol);
      }, 200);
    },
    volumeChanged(v) {
      if (v === 'finally') return throttleTimer.finally();
      throttleTimer(v);
    },
  },
};
</script>

<style lang="scss" scoped>
.container {
  margin: 1rem 2rem;
}

.volume {
  padding: 1rem 6rem;
  margin-top: 0rem;
  display: flex;

  .vue-slider {
    flex: 2 1 auto;
    margin: 0 2rem;
  }
}

.columns {
  display: flex;
  justify-content: left;
  flex-wrap: wrap;

  .column {
    flex: 2 1 auto;
    background-color: var(--color-secondary-bg);
    color: var(--color-secondary);
    margin: 0.75rem;
    padding: 0.2rem 0.5rem;
    border-radius: 2px;
    display: flex;

    &.block {
      flex: 1 1 100%;
    }

    >span {
      font-size: 0.5rem;
      margin-left: 1rem;
      align-items: center;
      line-height: 1rem;
      display: flex;

      .tag {
        padding: 0 0.5rem;
        margin: 0 0.2rem;
        line-height: 0.5rem;
        height: 1rem;
      }
    }
  }
}

@media only screen and (max-width: 479px) {
  .volume {
    padding: 1rem 1rem;
  }
}
</style>
