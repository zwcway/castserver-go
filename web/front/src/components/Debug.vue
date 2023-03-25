<template>
  <div v-if="isDebug()" id="debug-layout">
    <a-button class="debug-btn" type="primary" icon="tool" @click="show = !show"></a-button>
    <a-modal :visible="show" :mask="false" width="fit-content" :footer="null" :header="null"
      wrap-class-name="debug-page debug-wrap" dialog-class="debug-dialog" @cancel="show = false">
      <p>
        <label for="">mock
          <a-switch size="small" v-model="mock" checked-children="开" un-checked-children="关" />
        </label>
      </p>
      <p>
        <a-select size="small" v-model="appearance" dropdownClassName="debug-item">
          <a-select-option value="auto">自动&nbsp;&nbsp;&nbsp;&nbsp;</a-select-option>
          <a-select-option value="light">🌞 浅色</a-select-option>
          <a-select-option value="dark">🌚 暗黑</a-select-option>
        </a-select>
      </p>
      <p>
        <a-button size="small" type="" @click="onSpeakerDetect">发现设备</a-button>
      </p>

      <div v-if="speakerId >= 0">
        <a-button size="small" @click="speakerSendServerInfo">发送服务器信息</a-button>
        <a-button size="small" @click="speakerReconnect">重新连接</a-button>
      </div>
      <div v-if="lineId >= 0">
        <p>
          <a-input-search size="small" v-model="audioFile" placeholder="音频文件全路径" @search="playFile">
            <a-button size="small" slot="enterButton">播放文件</a-button>
          </a-input-search>
        </p>
        <p>
          <label for="">播放状态<a-switch size="small" v-model="playing" checked-children="正在播放"
              un-checked-children="已暂停" /></label>
          <label for="">声卡<a-switch size="small" v-model="localSpeaker" checked-children="开"
              un-checked-children="关" /></label>
        </p>
        <p>
          <label for="">频谱图<a-switch size="small" v-model="spectrumLog" checked-children="对数"
              un-checked-children="线性" /></label>
        </p>
      </div>
      <div v-if="elements.length">
        <h4>管道元</h4>
        <a-table :data-source="elements" size="small">
          <a-table-column key="name" title="名称">
            <template slot-scope="text, record">
              {{ record.name }}
              <a-switch v-if="lineId >= 0" size="small" :checked="record.on" checked-children="开" un-checked-children="关"
                @change="onLineElementPower(record, $event)" />
              <a-switch v-else size="small" :checked="record.on" checked-children="开" un-checked-children="关"
                @change="onSpeakerElementPower(record, $event)" />
            </template>
          </a-table-column>
          <a-table-column key="cost" title="耗时">
            <template slot-scope="text, record">
              {{ record.cost }}us
            </template>
          </a-table-column>
        </a-table>
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
      playing: false,
      localSpeaker: false,
      spectrumLog: false,
      elements: [],
      elementColums: [
        { title: '名称', dataIndex: 'name' },
        { title: '耗时', dataIndex: 'cost' },
      ]
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
        this.$store.commit('updateSettings', { key: 'appearance', value });
        changeAppearance(value);
      },
    },
    mock: {
      get() {
        return this.settings.enableMock;
      },
      set(value) {
        this.$store.commit('updateSettings', { key: 'enableMock', value });
        window.location.reload()
      }
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
      this.send('pause', { Line: this.lineId, Pause: !newVal });
    },
    localSpeaker(newVal) {
      this.send('localSpeaker', newVal);
    },
    spectrumLog(newVal) {
      this.send('setLine', { id: this.lineId, sl: newVal });
    }
  },
  methods: {
    isDebug() {
      return this.settings.enableDebugTool || process.env.NODE_ENV !== 'production';
    },
    loadStatus() {
      if (this.lineId < 0) return;
      socket.send('debugStatus', { line: this.lineId }).then(s => {
        this.playing = s.fplay
        this.audioFile = s.furl
        this.localSpeaker = s.local
        this.spectrumLog = s.sl
        this.elements = s.eles
      })
    },
    send() {
      socket.send.call(socket, ...arguments).then(() => {
        this.loadStatus()
      })
    },
    onSpeakerDetect() {
      this.send('addSpeaker', {
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
      this.send('sendServerInfo', this.speakerId);
    },
    speakerReconnect() {
      this.send('spReconnect', this.speakerId);
    },
    playFile() {
      if (this.audioFile.length < 4) {
        return;
      }
      let file = this.audioFile
      if (file[0] === '"') {
        file = file.substring(1, file.length - 1)
      }
      this.send('playFile', { Line: this.lineId, File: file });
    },
    onLineElementPower(ele, on) {
      this.send('elementPower', { line: this.lineId, n: ele.name, on })
    },
    onSpeakerElementPower(ele, on) {
      this.send('elementPower', { sp: this.speakerId, n: ele.name, on })
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

    label {
      margin: 0 0.5rem;
      display: inline-block;
    }

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