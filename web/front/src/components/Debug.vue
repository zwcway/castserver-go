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
        <a-button size="small" type="" @click="onSpeakerDetect">发现虚拟设备</a-button>
        <a-button size="small" type="" @click="onSpeakerCreate">创建虚拟设备...</a-button>
        <a-modal v-model:open="isShowSpeakerCreate" width="fit-content" title="创建设备"
          @cancel="isShowSpeakerCreate = false" @ok="onSpeakerCreateOk" okText="创建" okType="primary">
          <a-form :model="speakerCreateForm" layout="horizontal" :label-col="{ span: 3 }" :wrapper-col="{ span: 16 }"
            :rules="speakerCreateFormRules">
            <a-form-item label="名称">
              <a-input v-model:value="speakerCreateForm.name" />
            </a-form-item>
            <a-form-item label="ID">
              <a-input v-model:value="speakerCreateForm.id" />
            </a-form-item>
            <a-form-item label="IP">
              <a-input v-model:value="speakerCreateForm.ip" />
            </a-form-item>
            <a-form-item label="MAC">
              <a-input v-model:value="speakerCreateForm.mac" />
            </a-form-item>
            <a-form-item label="绝对音量">
              <a-switch v-model:checked="speakerCreateForm.avol" />
            </a-form-item>
            <a-form-item label="采样率">
              <a-checkbox-group v-model:value="speakerCreateForm.rate">
                <a-checkbox value="44100" name="rate">44.1KHz</a-checkbox>
                <a-checkbox value="48000" name="rate">48KHz</a-checkbox>
                <a-checkbox value="96000" name="rate">96KHz</a-checkbox>
                <a-checkbox value="192000" name="rate">192KHz</a-checkbox>
                <a-checkbox value="384000" name="rate">384KHz</a-checkbox>
              </a-checkbox-group>
            </a-form-item>
            <a-form-item label="位宽">
              <a-checkbox-group v-model:value="speakerCreateForm.bits">
                <a-checkbox value="s16le" name="bits">16位整数</a-checkbox>
                <a-checkbox value="s24le" name="bits">24位整数</a-checkbox>
                <a-checkbox value="s32le" name="bits">32位整数</a-checkbox>
                <a-checkbox value="f32le" name="bits">32位浮点</a-checkbox>
                <a-checkbox value="f64le" name="bits">64位浮点</a-checkbox>
              </a-checkbox-group>
            </a-form-item>
          </a-form>
        </a-modal>
      </p>

      <div v-if="speakerId >= 0">
        <a-button size="small" @click="speakerSendServerInfo">发送服务器信息</a-button>
        <a-button size="small" @click="speakerReconnect">重新连接</a-button>
        <a-button size="small" @click="speakerControlSample">设置音频格式</a-button>
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
        <a-table :data-source="elements" size="small" rowKey="name">
          <a-table-column key="name" title="名称">
            <template slot-scope="text, record">
              {{ record.name }}
              <a-switch v-if="lineId > 0 && record.on >= 0" size="small" :checked="!!record.on" checked-children="开"
                un-checked-children="关" @change="onLineElementPower(record, $event)" />
              <a-switch v-else-if="record.on >= 0" size="small" :checked="!!record.on" checked-children="开"
                un-checked-children="关" @change="onSpeakerElementPower(record, $event)" />
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

function generateMAC() {
  return (mock.Random.hex() + mock.Random.hex())
          .replaceAll('#', '')
          .replace(/(.{2})(?=.)/g, '$1:')
}
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
      ],
      isShowSpeakerCreate: false,
      speakerCreateForm: {
        name: '',
        v: 1,
        id: mock.Random.integer(1, 99999999),
        ip: mock.Random.ip(),
        mac: generateMAC(),
        port: mock.Random.natural(1, 65535),
        bits: ["s16le"],
        rate: ["44100"],
        avol: false,
      },
      speakerCreateFormRules: {
        name: [
          { required: true, trigger: 'change' },
          { min: 1, max: 5, trigger: 'blur' },
        ],
        id: [
          { required: true, trigger: 'change' },
          { min: 1, max: 5, trigger: 'blur' },
        ],
        bits: [
          { required: true, type: 'array', trigger: 'change' },
        ],
        rate: [
          { required: true, type: 'array', trigger: 'change' },
        ],
      }
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
      this.elements = [];
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
      setTimeout(() => {
        this.loadStatus()
      }, 1000)
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
        MAC: generateMAC(),
        DataPort: mock.Random.natural(1, 65535),
        BitsMask: ["s16le", "s24le", "s32le"],
        RateMask: ["44100"].map(r => parseInt(r)),
        AVol: mock.Random.boolean(),
      });
    },
    onSpeakerCreate() {
      this.isShowSpeakerCreate = true
    },
    onSpeakerCreateOk() {
      this.send('addSpeaker', {
        Ver: 1,
        ID: this.speakerCreateForm.id,
        IP: this.speakerCreateForm.id,
        MAC: this.speakerCreateForm.mac,
        DataPort: mock.Random.natural(1, 65535),
        BitsMask: this.speakerCreateForm.bits,
        RateMask: this.speakerCreateForm.rate.map(r => parseInt(r)),
        AVol: this.speakerCreateForm.avol,
      });
    },
    speakerSendServerInfo() {
      this.send('sendServerInfo', this.speakerId);
    },
    speakerReconnect() {
      this.send('spReconnect', this.speakerId);
    },
    speakerControlSample() {
      this.send('spSample', this.speakerId);
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
    bottom: 4rem;
    right: 0;
    width: 32px;
    height: 32px;
    z-index: 9999;
  }
}

.ant-modal-body {
  .ant-form-item {
    margin: 0;
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