<template>
  <div v-if="isDebug()" id="debug-layout">
    <a-button class="debug-btn" type="primary" icon="tool" @click="show = !show"></a-button>
    <a-modal :visible="show" :mask="false" width="fit-content" :footer="null" :header="null"
      wrap-class-name="debug-page debug-wrap" dialog-class="debug-dialog" @cancel="show = false">
      <a-select :value="appearance" @select="appearance = $event" dropdownClassName="debug-item">
        <a-select-option value="auto">自动&nbsp;&nbsp;&nbsp;&nbsp;</a-select-option>
        <a-select-option value="light">🌞 浅色</a-select-option>
        <a-select-option value="dark">🌚 暗黑</a-select-option>
      </a-select>
      <a-button type="" @click="onSpeakerDetect">发现设备</a-button>

      <div v-if="speakerId >= 0">
        <a-button @click="speakerSendServerInfo">发送服务器信息</a-button>
        <a-button @click="speakerReconnect">重新连接</a-button>
      </div>
      <div v-if="lineId >= 0" style="margin: 1rem 0;width: 20rem;">
        <a-input v-model="audioFile" placeholder="音频文件全路径" />
        <a-button @click="playFile">设置文件</a-button>
        <a-button @click="playPause">{{ pause ? '开始' : "播放" }}</a-button>
        <a-button @click="localSpeaker">{{ lsp ? '关闭声卡' : '打开声卡' }}</a-button>
      </div>
    </a-modal>
  </div>
</template>

<script>
import { mapState } from 'vuex';
import { changeAppearance } from '@/common/theme';
import { socket } from '@/common/request';
import * as ApiSpeaker from '@/api/speaker';
import mock from 'mockjs';
window.mock = mock;

export default {
  data() {
    return {
      show: false,
      speakerId: -1,
      lineId: -1,
      audioFile: "",
      pause: true,
      lsp: false, 
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
          value,
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
    }
  },
  methods: {
    isDebug() {
      return (
        process.env.NODE_ENV !== 'production' && this.settings.enableDebugTool
      );
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
      ApiSpeaker.sendServerInfo(this.speakerId);
    },
    speakerReconnect() {
      ApiSpeaker.reconnect(this.speakerId);
    },
    playFile() {
      if (this.audioFile.length < 12) {
        return;
      }
      let file = this.audioFile
      if (file[0] === '"') {
        file = file.substring(1, file.length - 1)
      }
      socket.send('playFile', {Line: this.lineId, File: file});
    },
    playPause() {
      socket.send('pause', {Line: this.lineId, Pause: !this.pause}).then(()=> {
        this.pause = !this.pause
      });
    },
    localSpeaker() {
      socket.send('localSpeaker', !this.lsp).then(() => {
        this.lsp = !this.lsp
      });
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