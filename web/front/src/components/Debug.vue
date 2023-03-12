<template>
  <div v-if="isDebug()" id="debug-layout">
    <a-button class="debug-btn" type="primary" icon="tool" @click="show = !show"></a-button>
    <a-modal :visible="show" :mask="false" width="fit-content" :footer="null" :header="null"
      wrap-class-name="debug-page debug-wrap" dialog-class="debug-dialog" @cancel="show = false">
      <a-select size="small" v-model="appearance" @select="appearance = $event" dropdownClassName="debug-item">
        <a-select-option value="auto">自动&nbsp;&nbsp;&nbsp;&nbsp;</a-select-option>
        <a-select-option value="light">🌞 浅色</a-select-option>
        <a-select-option value="dark">🌚 暗黑</a-select-option>
      </a-select>
      <a-button size="small" type="" @click="onSpeakerDetect">发现设备</a-button>

      <div v-if="speakerId >= 0">
        <a-button size="small" @click="speakerSendServerInfo">发送服务器信息</a-button>
        <a-button size="small" @click="speakerReconnect">重新连接</a-button>
      </div>
      <div v-if="lineId >= 0" style="margin: 1rem 0;width: 20rem;">
        <a-input-search size="small" v-model="audioFile" placeholder="音频文件全路径" @search="playFile">
          <a-button size="small" slot="enterButton">播放文件</a-button>
        </a-input-search>
        <label for="">{{ playing ? '正在播放' : '已暂停' }}<a-switch size="small" v-model="playing" /></label>
        <label for="">{{ localSpeaker ? '声卡已打开' : '声卡已关闭' }}<a-switch size="small" v-model="localSpeaker" /></label>
        <div>
          <label for="">{{ spectrumLog ? '频谱图对数' : '频谱图线性' }}<a-switch size="small" v-model="spectrumLog" /></label>
        </div>
      </div>
    </a-modal>
  </div>
</template>

<script>
import { mapState } from 'vuex';
import { changeAppearance } from '@/common/theme';
import { socket } from '@/common/request';
import mock from 'mockjs';
window.mock = mock;

export default {
  data() {
    return {
      show: false,
      speakerId: -1,
      lineId: -1,
      audioFile: "",
      playing: true,
      localSpeaker: false,
      spectrumLog: false,
    };
  },
  computed: {
    ...mapState(['settings']),
    appearance: {
      get() {
        if (this.settings.appearance === undefined) return 'auto';
        return this.settings.appearance;
      },
      set(value) {
        this.$store.commit('updateSettings', {
          key: 'appearance',
          value: value,
        });
        changeAppearance(value);
      },
    },
  },
  watch: {
    $route(newVal) {
      this.speakerId = -1;
      this.lineId = -1;
      if (newVal.name === 'speaker') {
        this.speakerId = parseInt(newVal.params.id);
      } else if (newVal.name === 'line') {
        this.lineId = parseInt(newVal.params.id);
      }
    },
    show(newVal) {
      if (newVal) {
        if (this.$route.name === 'line') {
          this.lineId = parseInt(this.$route.params.id);
        }
        this.loadStatus()
      }
    },
    lineId(line) {
      if (line < 0) return;
      this.loadStatus()
    },
    playing(newVal) {
      socket.send('pause', { Line: this.lineId, Pause: !newVal });
    },
    localSpeaker(newVal) {
      socket.send('localSpeaker', newVal);
    },
    spectrumLog(newVal) {
      socket.send('setLine', { id: this.lineId, sl: newVal });
    }
  },
  methods: {
    isDebug() {
      return this.settings.enableDebugTool;
    },
    loadStatus() {
      if (this.lineId < 0) return;
      socket.send('debugStatus', { line: this.lineId }).then(s => {
        this.playing = s.fplay
        this.audioFile = s.furl
        this.localSpeaker = s.local
        this.spectrumLog = s.sl
      })
    },
    onSpeakerDetect() {
      socket.send('addSpeaker', {
        Ver: 1,
        ID: mock.Random.integer(1, 99999999),
        IP: mock.Random.ip(),
        MAC: (mock.Random.hex() + mock.Random.hex())
          .replaceAll('#', '')
          .replace(/(.{2})(?=.)/g, '$1:'),
        DataPort: mock.Random.natural(1, 65535),
        BitsMask: [1, 2, 3],
        RateMask: [1, 2, 3],
        AVol: mock.Random.boolean(),
      });
    },
    speakerSendServerInfo() {
      socket.send('sendServerInfo', this.speakerId);
    },
    speakerReconnect() {
      socket.send('spReconnect', this.speakerId);
    },
    playFile() {
      if (this.audioFile.length < 4) {
        return;
      }
      let file = this.audioFile
      if (file[0] === '"') {
        file = file.substring(1, file.length - 1)
      }
      socket.send('playFile', { Line: this.lineId, File: file });
    },
  },
};
</script>

<style lang="scss">
#debug-layout {
  .debug-btn {
    position: fixed;
    bottom: 1rem;
    right: 0;
    width: 32px;
    height: 32px;
    z-index: 9999;
  }
}

.debug-item {
  z-index: 9999 !important;
}

.debug-wrap.debug-page {
  width: auto;
  height: auto;
  top: auto;
  left: auto;
  right: 32px;
  overflow: hidden;
  z-index: 8888;

  .debug-dialog {
    top: 0;
    bottom: 0;
    padding: 0;
    margin: 1rem;

    .ant-modal-close {
      .ant-modal-close-x {
        width: 32px;
        height: 32px;
        line-height: 32px;
      }
    }

    .ant-modal-body {
      padding-top: 2rem;
    }
  }
}
</style>