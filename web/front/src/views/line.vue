<template>
  <div
    class="line-page"
    @click="showChannelInfo(-1)"
    :class="{ 'select-channel': specifyChannel }"
  >
    <div class="volume">
      <span>{{ line.name }}</span>
      <vue-slider
        v-model="volume"
        :min="0"
        :max="100"
        :process="volumeLevelProcess"
        :tooltip-placement="'bottom'"
        ref="volumeSlider"
        @change="onVolumeChanged"
        @drag-end="onVolumeChanged('finally')"
      />
      <div>
        <Select
          :value="layout"
          class="layout-select"
          :options="layoutList"
          @select="layout = $event"
        ></Select>
        <Button type="link" @click.stop.prevent="isShowEqualizer = true">
          <i class="codicon codicon-settings"></i>
        </Button>
      </div>
    </div>
    <div class="container">
      <canvas id="spectrum" class="spectrum" height="0" width="0"></canvas>
      <div class="background" rel="background"></div>
      <div class="room">
        <div class="line"></div>
        <div class="line"></div>
        <div class="line"></div>
      </div>
      <div class="channels channels-layout" :class="layoutClass()">
        <div
          class="speaker"
          v-for="(ch, id) in channelAttr"
          :key="ch.id"
          v-bind:id="ch.id"
          @click.stop.prevent="onSelectChannel(id)"
          @mouseover="onChannelMouseHover(id, true)"
          @mouseleave="onChannelMouseHover(id, false)"
          @touchend="onChannelMouseHover(id, false)"
          :class="{
            enabled: channelSpeakers[id] && channelSpeakers[id].length > 0,
            active: ch.show,
          }"
        >
          <svg-icon v-bind:iconClass="ch.icon" :size="0" />
        </div>
      </div>
    </div>
    <div
      id="popper"
      class="popper"
      v-show="popper.length > 0"
      data-popper-placement="top"
    >
      <div>
        <div class="channel-name">{{ popper }}</div>
        <div data-popper-arrow="true" class="arrow">
          <div></div>
        </div>
      </div>
    </div>
    <div class="infomation">
      <div class="channel-name" v-show="infomation.channelId">
        <span>{{ infomation.name }}</span>
      </div>
      <ul
        class="speaker-list"
        v-show="infomation.speakers && infomation.speakers.length"
      >
        <li class="speaker" v-for="sp in infomation.speakers" :key="sp.ip">
          <div>
            <svg-icon icon-class="speaker" :size="0"></svg-icon>
            <span class="name" @click.stop.prevent="gotoSpeaker(sp.id)">
              {{ sp.name }}
            </span>
            <span
              class="channel-name user-select-none"
              @click.stop.prevent="showChannelInfo(sp.channel)"
              v-show="channelAttr[sp.channel]"
            >
              {{ showChannelName(sp.channel) }}
              <Button
                icon="close"
                @click.stop.prevent="onRemoveSPChannel(sp)"
              ></Button>
            </span>
            <Button
              type="link"
              v-show="!infomation.channelId && !channelAttr[sp.channel]"
              @click.stop.prevent="onSpecifyChannel(sp.id)"
            >
              指定声道
            </Button>
          </div>
          <span class="ip">{{ sp.ip }}</span>
          <span class="ratebits">{{ showRatebits(sp) }}</span>
          <vue-slider
            v-model="sp.volume"
            :min="0"
            :max="100"
            :process="speakerVolumeLevelProcess"
            tooltip-placement="top"
            @change="onSpeakerVolumeChanged($event, sp.id)"
            @drag-end="onSpeakerVolumeChanged('finally', sp.id)"
          />
        </li>
      </ul>
      <input
        v-show="!infomation.speakers || infomation.speakers.length === 0"
        type="button"
        value="选择扬声器"
      />
    </div>
    <modal
      :visible="isShowEqualizer"
      width="fit-content"
      :footer="null"
      @cancel="isShowEqualizer = false"
    >
      <equalizer
        id="equalizer"
        :bands="equalizerBands"
        ref="equalizer"
        @change="onEQChange"
      />
      <template slot="title">
        <div class="eq-toolbar">
          <span>均衡器</span>
          <Select
            class="eq-band-select"
            :options="equalizerBandsList"
            :value="eqBandsSelected"
            @select="eqBandsSelected = $event"
          >
          </Select>
        </div>
      </template>
    </modal>
    <div class="select-channel-mask" v-show="specifyChannel">
      <Button icon="close" @click.stop.prevent="specifyChannel = 0"></Button>
    </div>
  </div>
</template>

<script>
import Vue from 'vue';
import VueSlider from 'vue-slider-component';
import 'vue-slider-component/theme/antd.css';
import Equalizer from '@/components/Equalizer';
import { Modal, Button, Select } from 'ant-design-vue';
import 'ant-design-vue/lib/modal/style';
import 'ant-design-vue/lib/button/style';
// import icons for win32 title bar
// icons by https://github.com/microsoft/vscode-codicons
import '@vscode/codicons/dist/codicon.css';
import {
  channelList,
  clearEqualizer,
  getLineInfo,
  listenLineSpectrum,
  removeListenLineSpectrum,
  setEqualizer,
  setVolume as setLineVolume,
} from '@/api/line';
import { socket } from '@/common/request';
import { setVolume as setSpeakerVolume, setChannel } from '@/api/speaker';
import { throttleFunction } from '@/common/throttle';
import { createPopper } from '@popperjs/core';
import { formatRate, formatBits } from '@/common/format';
import '@/assets/css/popper.scss';
import '@/assets/css/speaker.scss';

Vue.use(Modal, Button);

let throttleTimer, speakerThrottleTimer;

var SP, ctx;
let spdata = [];
let slow = [];
let title = [];
let spRequestId;
function drawSpectrum() {
  ctx.clearRect(0, 0, SP.width, SP.height);
  let spc = Math.min(128, spdata.length);
  let w = SP.width / (spc + 1);
  let left = 0;
  for (var i = 0; i < spc; i++) {
    left = i * w + 4;

    ctx.beginPath();
    ctx.lineWidth = 4;
    ctx.strokeStyle = 'hsl(171deg 100% 41% / 20%)';

    if (slow[i] > spdata[i]) {
      slow[i] -= 2;
    } else if (spdata[i] - slow[i] > 8) {
      slow[i] += 5;
    } else {
      slow[i] = spdata[i];
    }

    ctx.moveTo(left, SP.height);
    ctx.lineTo(left, SP.height - slow[i]);
    ctx.stroke();

    if (title[i] > slow[i]) {
      title[i] -= 0.5;
    } else {
      title[i] = slow[i];
    }

    ctx.beginPath();
    ctx.lineWidth = 4;
    ctx.strokeStyle = '#bbb';
    ctx.moveTo(left, SP.height - title[i]);
    ctx.lineTo(left, SP.height - title[i] + 1);
    ctx.stroke();
  }

  spRequestId = requestAnimationFrame(drawSpectrum);
}

export default {
  name: 'Speaker',
  components: {
    VueSlider,
    Equalizer,
    Modal,
    Button,
    Select,
  },
  data() {
    return {
      line: {},
      layout: '2-0',
      layoutList: [
        { key: '2-0', label: '2.0 声道', disabled: false },
        { key: '2-1', label: '2.1 声道', disabled: false },
        { key: '5-1', label: '5.1 声道', disabled: false },
        { key: '7-1', label: '7.1 声道', disabled: false },
      ],
      channels: 0,
      channelLayout: 'none',
      channelSpeakers: {},
      channelAttr: channelList,
      popper: '',
      infomation: {},
      volumeLevelProcess(dotsPos) {
        return [[0, 0, { backgroundColor: 'pink' }]];
      },
      speakerVolumeLevelProcess(dotsPos) {
        return [[0, 0, { backgroundColor: 'pink' }]];
      },
      isShowEqualizer: false,
      equalizerBands: 15,
      equalizerBandsList: [],
      specifyChannel: 0, // speakerid
      fromAction: '',
      shownChannelId: 0,
    };
  },
  props: ['id'],
  watch: {
    $route(to, from) {
      if (to.params.action === 'specifychannel') {
        this.specifyChannel = parseInt(to.params.spid);
        return;
      } else if (this.fromAction === 'specifychannel') {
        this.specifyChannel = 0;
        return;
      }
      this.specifyChannel = 0;
      this.infomation = {};
      this.loadData();
      this.initSpectrum();
      Modal.destroyAll();
    },
    isShowEqualizer(newVal) {
      if (!newVal) return;
      this.$nextTick(() => {
        // 等待页面加载完成
        let bands = [];
        this.$refs.equalizer.bandList().forEach(band => {
          bands.push({
            key: band,
            label: band + ' 段',
          });
        });

        this.equalizerBandsList = bands;
        this.eqBandsSelected = bands[0].key;
      });
    },
  },
  beforeRouteEnter(to, from, next) {
    next(vm => {
      // 通过 `vm` 访问组件实例,将值传入fromPath
      vm.fromAction = from.params ? from.params.action : '';
    });
  },
  computed: {
    volume: {
      get() {
        return this.line.volume || 0;
      },
      set(value) {
        this.line.volume = value;
      },
    },
    eqBandsSelected: {
      get() {
        return this.equalizerBands;
      },
      set(value) {
        let band = parseInt(value);
        if (band == this.equalizerBands) return;
        this.equalizerBands = band;
        clearEqualizer(this.line.id);
      },
    },
  },
  mounted() {
    this.initSpectrum();
    socket.onConnected().then(() => this.loadData());
  },
  destroyed() {
    cancelAnimationFrame(spRequestId);
    removeListenLineSpectrum(this.line.id);
    Modal.destroyAll();
  },
  activated() {
    // keep-alived 开启后生效
    this.loadData();
  },
  methods: {
    loadData() {
      if (!this.id) {
        this.$router.push('/speakers');
        return;
      }
      let id = parseInt(this.id || '0');

      getLineInfo(id).then(data => {
        data = data || {};
        if (data.id === undefined) {
          this.$router.push('/speakers');
          return;
        }

        this.line = data;

        this.initChannelSpeakers();
        this.showChannelInfo(-1);

        listenLineSpectrum(this.line.id, this.onSpectrumChange);
      });
    },
    initChannelSpeakers() {
      this.channelSpeakers = {};
      let speakers = {};
      (this.line.speakers || []).forEach(sp => {
        if (speakers[sp.channel] === undefined) {
          speakers[sp.channel] = [];
        }
        speakers[sp.channel].push(sp);
      });
      this.channelSpeakers = speakers;

      for (var i in this.channelAttr) {
        if (this.channelAttr[i].show) {
          this.$set(this.channelAttr[i], 'show', false);
        }
      }
      this.computeLayout();
    },
    changeSpeakerAttrById(spid, att, val) {
      let speakers = this.line.speakers || [];
      let len = speakers.length;
      let sp;
      for (let i = 0; i < len; i++) {
        sp = speakers[i];
        if (sp.id == spid) {
          sp[att] = val;
          this.$set(this.line, 'speakers', speakers);
          break;
        }
      }
      this.initChannelSpeakers();
    },
    onSpectrumChange(data) {
      spdata = data;
    },
    computeLayout() {
      let chs = [];
      for (let ch in this.channelSpeakers) {
        if (this.channelAttr[ch] === undefined) continue;
        if (this.channelSpeakers[ch].length) chs.push(ch);
      }
      let l = '-' + chs.sort().join('-') + '-';
      if (l.length === 2) {
        this.layout = '2-0';
      } else if (l.indexOf('-7-') >= 0 || l.indexOf('-8-') >= 0) {
        this.layout = '7-1';
      } else if (
        l.indexOf('-3-') >= 0 ||
        l.indexOf('-10-') >= 0 ||
        l.indexOf('-11-') >= 0
      ) {
        this.layout = '5-1';
      } else if (l.indexOf('-6-') >= 0) {
        this.layout = '2-1';
      } else if (l.indexOf('-1-') >= 0 || l.indexOf('-2-') >= 0) {
        this.layout = '2-0';
      } else {
        this.layout = '7-1';
      }
      this.disableLayoutSelect();
    },
    disableLayoutSelect() {
      for (let i = this.layoutList.length - 1, s = -1; i >= 0; i--) {
        if (this.layoutList[i].key === this.layout) {
          s = i;
        } else if (s > 0) {
          this.layoutList[i].disabled = true;
        } else {
          this.layoutList[i].disabled = false;
        }
      }
    },
    onVolumeChanged(v) {
      if (v === 'finally') return throttleTimer.finally();
      throttleTimer(v);
    },
    onSpeakerVolumeChanged(v, id) {
      if (v === 'finally') return speakerThrottleTimer.finally();
      speakerThrottleTimer(id, v);
    },
    initSpectrum() {
      throttleTimer = throttleFunction(vol => {
        setLineVolume(this.line.id, vol);
      }, 200);
      speakerThrottleTimer = throttleFunction((id, vol) => {
        setSpeakerVolume(id, vol);
      }, 200);

      cancelAnimationFrame(spRequestId);

      SP = document.getElementById('spectrum');
      let bg = SP.nextSibling;
      SP.width = bg.offsetWidth;
      SP.height = bg.offsetHeight;
      ctx = SP.getContext('2d');
      drawSpectrum();
    },
    onEQChange(freq, gain) {
      setEqualizer(this.line.id, freq, gain);
    },
    onSpecifyChannel(spid) {
      // this.$router.push({
      //   name: this.$route.name,
      //   params: {
      //     id: this.$route.params.id,
      //     action: 'specifychannel',
      //     spid: spid,
      //   },
      // });

      this.specifyChannel = spid;
      window.scrollTo({
        top: 0,
        left: 0,
        behavior: 'smooth',
      });
    },
    onRemoveSPChannel(speaker) {
      let that = this;
      Modal.confirm({
        title: '确定要移除该扬声器所关联的声道吗？',
        content:
          '移除关联的声道之后，该扬声器将处于空闲状态。可再次点击“指定声道”按钮以重新指定。',
        okText: '是',
        okType: 'danger',
        cancelText: '否',
        onOk() {
          setChannel(speaker.id, 0).then(() => {
            that.changeSpeakerAttrById(speaker.id, 'channel', 0);
            that.showChannelInfo(that.shownChannelId);
          });
        },
      });
    },
    onSelectChannel(ch) {
      let spid = this.specifyChannel;
      ch = parseInt(ch);
      if (spid) {
        setChannel(spid, ch)
          .then(() => {
            this.changeSpeakerAttrById(spid, 'channel', ch);
          })
          .finally(() => {
            // this.$router.go(-1);
            this.specifyChannel = 0;
          });
        return;
      }
      this.showChannelInfo(ch);
    },
    showChannelName(ch) {
      if (ch in this.channelAttr) return this.channelAttr[ch].name;
      return '';
    },
    showChannelInfo(chid) {
      this.shownChannelId = chid;
      if (!(chid in this.channelAttr)) {
        this.infomation = {
          speakers: this.line.speakers || [],
        };
        return;
      }

      let ch = this.channelAttr[chid];
      this.infomation = {
        channelId: chid,
        name: ch.name,
        speakers: this.channelSpeakers[chid],
      };

      for (var k in this.channelAttr) {
        if (k === chid) {
          this.$set(this.channelAttr[k], 'show', true);
        } else {
          this.$set(this.channelAttr[k], 'show', false);
        }
      }
    },
    onChannelMouseHover(id, shown) {
      if (this.channelAttr[id] !== undefined && shown) {
        let ch = this.channelAttr[id];
        this.popper = ch.name;
        let ref = document.querySelector('#popper');
        let target = document.querySelector('#' + ch.id);

        createPopper(target, ref, {
          placement:
            ref.attributes.getNamedItem('data-popper-placement').value || 'top',
          strategy: 'fixed',
        });
      } else {
        this.popper = '';
      }
    },
    gotoSpeaker(id) {
      this.$router.push('/speaker/' + id);
    },
    showRatebits(sp) {
      return formatRate(sp.rate) + '/' + formatBits(sp.bits);
    },
    layoutClass() {
      let a = {};
      a['layout-' + this.layout] = true;
      return a;
    },
    modalClose() {
      this.isShowEqualizer = false;
    },
  },
};
</script>

<style lang="scss">
@import 'bulma/sass/utilities/_all.sass';
@import 'ant-design-vue/lib/select/style/index.css';

$light-color: hsl(171, 100%, 41%);

.line-page {
  position: relative;
  .container {
    position: relative;
    display: flex;
    justify-content: center;
    align-items: center;
    width: 480px;
    height: 320px;
    margin-top: -3rem;
    padding-bottom: 4rem;
    overflow: hidden;

    .svg-icon {
      display: block;
    }

    .background {
      width: 100%;
      height: 100%;
      position: relative;
    }

    .spectrum {
      position: absolute;
      bottom: 1rem;
    }

    .room {
      width: 100%;
      top: 0;
      bottom: 2rem;
      position: absolute;

      .line {
        border: 1px solid;
        position: absolute;
        border-radius: 100%;
      }

      .line:nth-child(1) {
        right: 10%;
        bottom: 0;
        top: 50%;
        left: 10%;
      }

      .line:nth-child(2) {
        bottom: 5%;
        top: 51%;
        left: 14%;
        right: 14%;
      }

      .line:nth-child(3) {
        bottom: 10%;
        top: 52%;
        left: 18%;
        right: 18%;
      }
    }

    #front-left {
      top: 43%;
      left: 30%;
    }

    #front-right {
      top: 43%;
      right: 30%;
    }

    #front-center {
      top: 44%;
      left: calc(50% - 1.5rem);

      .svg-icon {
        width: 3rem;
      }
    }

    #front-bass {
      top: 47%;
      left: 23%;

      .svg-icon {
        width: 2.5rem;
      }
    }

    #side-left {
      top: 59%;
      left: 10%;

      .svg-icon {
        width: 3rem;
      }
    }

    #side-right {
      top: 59%;
      right: 10%;

      .svg-icon {
        width: 3rem;
      }
    }

    #back-left {
      top: 78%;
      left: 27%;

      .svg-icon {
        width: 3rem;
      }
    }

    #back-right {
      top: 78%;
      right: 27%;

      .svg-icon {
        width: 3rem;
      }
    }

    .channels {
      width: 100%;
      height: 100%;
      position: absolute;

      .speaker {
        position: absolute;
        cursor: pointer;

        &.enabled {
          .svg-icon {
            color: #000000;
          }
        }

        &.active {
          .svg-icon {
            color: hsl(229, 29%, 68%);
          }
        }

        &.active.enabled {
          .svg-icon {
            color: hsl(229, 53%, 53%);
          }
        }

        .svg-icon {
          width: 4rem;
          height: 48px;
          color: #bbbbbb;
        }
      }
    }
  }

  .layout-2-0 {
    .speaker {
      display: none;
    }

    #front-left,
    #front-right {
      display: block;
    }
  }

  .layout-2-1 {
    .speaker {
      display: none;
    }

    #front-left,
    #front-right,
    #front-bass {
      display: block;
    }
  }

  .layout-5-1 {
    .speaker {
      display: none;
    }

    #front-left,
    #front-right,
    #front-center,
    #front-bass,
    #back-left,
    #back-right {
      display: block;
    }
  }

  .infomation {
    display: block;
    position: relative;
    margin: 0 5rem;
    padding: 5px 10px 2rem 10px;
    background-color: $white-ter;
    border-radius: 5px;

    > .channel-name {
      text-align: center;
      font-weight: bold;
      position: relative;
      display: block;
      height: 2rem;

      > span {
        display: inline-block;
        z-index: 1;
        position: relative;
        color: #fff;
        top: 0;

        &::after {
          content: '';
          left: -10px;
          right: -10px;
          top: -10px;
          bottom: -5px;
          position: absolute;
          border-radius: 0 0 10px 10px;
          background-color: #22d0b2;
          z-index: -1;
        }
      }
    }

    .speaker-list {
      display: flex;
      justify-content: center;
      flex-wrap: wrap;

      .speaker {
        cursor: pointer;
        line-height: 2rem;
        background-color: #fff;
        border-radius: 5px;
        border-bottom: solid 1px $border;
        position: relative;
        margin: 0;
        padding: 10px;
        width: 100%;
        min-width: 330px;

        &:hover {
          border-color: $light-color;

          .svg-icon {
            color: $light-color;
          }
        }
        > div {
          display: flex;
          flex-direction: row;
          flex-wrap: nowrap;
          line-height: 2rem;
          .svg-icon {
            flex: 0 0 16px;
            width: 16px;
            height: 16px;
            margin: 0 3px;
          }

          .name {
            flex: 1 0 9rem;
            width: 9rem;
            overflow: hidden;
            text-overflow: ellipsis;
          }
          .channel-name {
            flex: 0 0 auto;
            width: auto;
            white-space: nowrap;
            background: $light-color;
            color: $white;
            padding: 0 4px;
            border-radius: 2px;
            button {
              width: 20px;
              height: 21px;
              font-size: 14px;
              margin-left: 3px;
            }
          }
        }
        .ip {
          font-size: 0.5rem;
          margin: 0 4px;
        }
      }
    }
  }

  .volume {
    padding: 1rem 6rem;
    padding-top: 2rem;
    display: flex;
    z-index: 1;
    position: relative;
    .vue-slider {
      flex: 2 1 auto;
      margin: 0 2rem;
    }
    .layout-select {
    }
  }
  .select-channel-mask {
    position: absolute;
    top: 0;
    bottom: 0;
    left: 0;
    right: 0;
    background: #00000066;
    z-index: 5;
    backdrop-filter: blur(6px);
    display: flex;
    justify-content: right;
    padding: 1rem;
    button {
      position: relative;
      right: 0;
    }
  }
  &.select-channel {
    .volume {
      z-index: unset;
    }
    .layout-select {
      z-index: 6;
      position: relative;
    }
    .container .channels {
      z-index: 6;
      .speaker {
        .svg-icon {
          color: $white;
        }
        &.enabled {
          .svg-icon {
            color: $light-color;
          }
        }
      }
    }
  }

  #popper {
    z-index: 6;
  }
  @media only screen and (max-width: 479px) {
    .container {
      width: 320px;
      height: 240px;
      .room {
      }
      .channels {
        .svg-icon {
          width: 3rem;
        }

        #front-bass .svg-icon {
          width: 2em !important;
        }

        #side-left .svg-icon,
        #side-right .svg-icon {
          width: 2em;
        }

        #front-center {
          left: calc(50% - 1rem);

          .svg-icon {
            width: 2em;
          }
        }
      }
    }

    .infomation {
      margin-left: 1rem;
      margin-right: 1rem;
    }
    .volume {
      padding: 1rem 1rem;
    }
  }

  @media only screen and (min-width: 480px) {
    .container {
      width: 480px;
      height: 320px;

      .channels {
        .speaker {
          .svg-icon {
            width: 3rem !important;
          }

          &#front-bass .svg-icon {
            width: 2em !important;
          }
        }
      }
    }

    .infomation {
      margin-left: 3rem;
      margin-right: 3rem;
    }
  }

  @media only screen and (min-width: 820px) {
    .container {
      width: 640px;
      height: 320px;

      .channels {
        .svg-icon {
          width: 6rem !important;
        }
      }
    }

    .infomation {
      margin-left: 5rem;
      margin-right: 5rem;
      .speaker-list {
        .speaker {
          width: auto;
          margin: 1rem;
        }
      }
    }
  }
}

.eq-toolbar {
  display: flex;
  flex-direction: row;
  position: relative;
  width: calc(100% - 5rem);
  > span {
    flex-grow: 1;
  }
  .eq-band-select {
    align-self: flex-end;
    right: 0;
    top: 0;
  }
}
.equalizer {
  overflow-x: auto;
}
</style>
