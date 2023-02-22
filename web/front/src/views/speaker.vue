<template>
  <div class="container">
    <div class="card">
      <div class="volume">
        <span>{{ name }}</span>
        <vue-slider
          v-model="volume"
          :min="0"
          :max="100"
          :process="volumeLevelProcess"
          :tooltip-placement="'bottom'"
          ref="volumeSlider"
          @change="volumeChanged"
          @drag-end="volumeChanged('finally')"
        />
        <div> </div>
      </div>
    </div>
    <div class="notification">
      <div class="columns is-mobile">
        <div class="column">
          <label>MAC</label>
          <span>{{ mac }}</span>
        </div>
        <div class="column">
          <label>声道</label>
          <span>{{ channel }}</span>
        </div>
        <div class="column">
          <label>连接状态</label>
          <span>{{ state }}</span>
        </div>
        <div class="column block">
          <label>支持的采样率</label>
          <span class="tags">
            <span class="tag" v-for="rate in rateList" :key="rate">{{
              rate
            }}</span>
          </span>
        </div>
        <div class="column block">
          <label>支持的位宽</label>
          <span class="tags">
            <span class="tag" v-for="bit in bitsList" :key="bit"
              >{{ bit }}bit</span
            >
          </span>
        </div>
        <div class="column">
          <label>队列中</label>
          <span>{{ statistic.queue }}B</span>
        </div>
        <div class="column">
          <label>已发送</label>
          <span>{{ statistic.send }}B</span>
        </div>
        <div class="column">
          <label>已丢弃</label>
          <span>{{ statistic.drop }}B</span>
        </div>
      </div>
    </div>
    <div>
      <div id="spectrum">频谱</div>
    </div>
    ip mac 连接状态 网络速率 支持所有采样率 支持所有位数 所属声道 房间 音量 频谱
    统计数据 调试工具： 重新连接 队列状态
    <div class="debugger" v-if="enableDebugTool">
      <button class="button" v-on:click="sendServerInfo">发送服务端信息</button>
      <button class="button" v-on:click="reconnect">重新连接</button>
    </div>
  </div>
</template>

<script>
import { mapState } from 'vuex';
import VueSlider from 'vue-slider-component';
import 'vue-slider-component/theme/antd.css';
import { setVolume } from '@/api/speaker';
import { throttleFunction } from '@/common/throttle';
import { channelList, getLineInfo } from '@/api/line';
import { getSpeakerInfo, sendServerInfo, reconnect } from '@/api/speaker';
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
      volumeLevelProcess(dotsPos) {
        return [[0, 0, { backgroundColor: 'pink' }]];
      },
    };
  },
  computed: {
    ...mapState(['settings']),
    enableDebugTool: {
      get() {
        return this.settings.enableDebugTool;
      },
      set(value) {
        this.$store.commit('updateSettings', {
          key: 'enableDebugTool',
          value,
        });
      },
    },
    volume: {
      get() {
        return this.speaker.volume || 0;
      },
      set(value) {
        this.speaker.volume = value;
      },
    },
    name() {
      return this.speaker.name || '';
    },
    mac() {
      return this.speaker.mac || '';
    },
    channel() {
      for (let i in channelList) {
        if (channelList[i].id == (this.speaker.channel || '')) {
          return channelList[i].name;
        }
      }

      return '';
    },
    state() {
      return this.speaker.state || '';
    },
    rateList() {
      return this.speaker.rateList || [];
    },
    bitsList() {
      return this.speaker.bitsList || [];
    },
    statistic() {
      return this.speaker.statistic || { queue: 0, send: 0, drop: 0 };
    },
  },
  mounted() {
    this.id = parseInt(this.$route.params.id || 0);
    if (!this.id) {
      this.$router.push('/speakers');
      return;
    }
    socket.onConnected().then(() => this.loadData());
  },

  methods: {
    loadData() {
      getSpeakerInfo(this.id)
        .then(data => {
          this.speaker = data;
        })
        .catch(code => {
          this.$router.push('/speakers');
        });

      throttleTimer = throttleFunction(vol => {
        setVolume(this.speaker.id, vol);
      }, 200);
    },
    volumeChanged(v) {
      if (v === 'finally') return throttleTimer.finally();
      throttleTimer(v);
    },
    sendServerInfo() {
      sendServerInfo(this.id);
    },
    reconnect() {
      reconnect(this.id);
    },
  },
};
</script>

<style lang="scss" scoped>
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
    background-color: #eee;
    margin: 0.75rem;
    padding: 0.2rem 0.5rem;
    border-radius: 2px;
    display: flex;

    &.block {
      flex: 1 1 100%;
    }

    > span {
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
