<template>
  <div class="container">
    <div class="card">{{ name }}</div>
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
        <div class="column">
          <label>支持的采样率</label>
          <span class="tags">
            <span class="tag" v-for="rate in rateList">{{ rate }}Hz</span>
          </span>
        </div>
        <div class="column">
          <label>支持的位宽</label>
          <span class="tags">
            <span class="tag" v-for="bit in bitsList">{{ bit }}bit</span>
          </span>
        </div>
        <div class="column">
          <label>连接状态</label>
          <span>{{ state }}</span>
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
import { getSpeakerInfo, sendServerInfo, reconnect } from '@/api/speaker';
import { socket } from '@/common/request';

export default {
  name: 'Speaker',
  data() {
    return {
      id: 0,
      speaker: {},
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
    name() {
      return this.speaker.name || '';
    },
    mac() {
      return this.speaker.mac || '';
    },
    channel() {
      return this.speaker.channel || '';
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

<style scoped></style>
