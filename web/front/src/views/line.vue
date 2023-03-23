<template>
  <div class="line-page" :class="{ 'select-channel': specifyChannel >= 0 }">
    <div class="player" v-if="isPlayer" @touchstart.stop @mousedown.stop>
      <span>{{ currentDuration }}</span>
      <VueSlider v-model="line.source.cur" :min="0" :max="line.source.dur" tooltipPlacement="bottom"
        :tooltipFormatter="playerSliderFormater" @drag-start="onPlayerSeekStart" @drag-end="onPlayerSeekStop" />
      <span>{{ totalDuration }}</span>
    </div>
    <div class="volume" @touchstart.stop @mousedown.stop>
      <div>
        <div class="line-name">
          <span v-if="!isLineNameEdit">{{ line.name }}</span>
          <a-input v-else :value="line.name" placeholder="名称" :max-width="6" @change="line.newName = $event.target.value"
            @blur="onNameChange" @keyup.enter="onNameChange" />
          <a-button v-if="!isLineNameEdit" type="link" @click.stop.prevent="isLineNameEdit = true">
            <a-icon type="edit" />
          </a-button>
          <a-button v-else type="link" @click.stop.prevent="onNameChange">
            <i class="codicon codicon-check"></i>
          </a-button>
        </div>
        <span class="source">
          <span class="ratebits" v-show="line.source && line.source.rate">
            {{ showSourceFormat(line.source) }}
          </span>
        </span>
      </div>
      <div :id="'line-' + line.id" class="level-meter-slider">
        <Volume :volume="line.vol" :mute="line.mute || false" @change="onVolumeChanged" tooltip-placement="bottom"
          @mute="onVolumeChanged" />
        <a-button type="link" @click="isLayout3d = !isLayout3d" icon="code-sandbox"
          :class="{ active: isLayout3d }"></a-button>
        <a-select :value="layout" class="layout-select" :options="layoutList" @select="onLayoutChanged"
          :dropdownMatchSelectWidth="false"></a-select>
        <a-button type="link" @click.stop.prevent="isShowEqualizer = true">
          <i class="codicon codicon-settings"></i>
        </a-button>
      </div>
    </div>
    <div class="container room room-2d" :class="{ 'room-3d': isLayout3d }" @click="onShowChannelInfo(-1)">
      <canvas :id="`spectrum-${line.id}`" class="spectrum" height="0" width="0"></canvas>
      <div class="background" rel="background"></div>
      <div class="wall">
        <div class="cube">
          <div class="cube-face wall-floor">
            <div class="line"></div>
            <div class="line"></div>
            <div class="line"></div>
          </div>
          <div class="cube-face wall-background"></div>
          <div class="cube-face wall-left"></div>
          <div class="cube-face wall-right"></div>
        </div>
      </div>
      <div class="channels channels-layout" :class="layoutClass()">
        <div class="speaker" v-for="(ch, id) in channelAttr" :key="ch.id" v-bind:id="ch.id"
          @click.stop.prevent="onSelectChannel(id)" @mouseover="onChannelMouseHover(id, true)"
          @mouseleave="onChannelMouseHover(id, false)" @touchend="onChannelMouseHover(id, false)" :class="{
            enabled: channelSpeakers[id] && channelSpeakers[id].length > 0,
            active: ch.show,
          }">
          <svg-icon v-bind:iconClass="ch.icon" :size="0" />
        </div>
      </div>
    </div>
    <div id="popper" class="popper" v-show="popper.length > 0" data-popper-placement="top">
      <div>
        <div class="channel-name">{{ popper }}</div>
        <div data-popper-arrow="true" class="arrow">
          <div></div>
        </div>
      </div>
    </div>
    <div class="infomation">
      <div class="channel-name" v-show="infomation.channelId">
        <div>
          <svg-icon icon-class="speaker" :size="0" @click.native="onChannelTest(infomation.channelId)"></svg-icon>
          <span>{{ infomation.name }}</span>
        </div>
      </div>
      <ul class="speaker-list" v-show="infomation.speakers && infomation.speakers.length">
        <li class="speaker level-meter-slider" v-for="sp in infomation.speakers" :key="sp.id" :id="'speaker-' + sp.id">
          <div>
            <svg-icon icon-class="speaker" :size="0" @click.native="onIamHere(sp.id)"></svg-icon>
            <span class="name" @click.stop.prevent="gotoSpeaker(sp.id)">
              {{ sp.name }}
            </span>
            <span class="channel-name user-select-none" v-if="sp.ch && channelAttr[sp.ch]"
              @click.stop.prevent="onShowChannelInfo(sp.ch)">
              {{ showChannelName(sp.ch) }}
              <a-button icon="close" @click.stop.prevent="onRemoveSPChannel(sp)"></a-button>
            </span>
            <a-button type="link" v-if="!sp.ch || (!infomation.channelId && !channelAttr[sp.ch])"
              @click.stop.prevent="onSpecifyChannel(sp.id)">
              指定声道
            </a-button>
          </div>
          <span class="ip">{{ sp.ip }}</span>
          <span class="ratebits">{{ showRatebits(sp) }}</span>
          <Volume :volume="sp.vol" :mute="sp.mute" tooltip-placement="top" @change="onSpeakerVolumeChanged($event, sp.id)"
            @mute="onSpeakerVolumeMute($event, sp.id)" />
        </li>
      </ul>
      <input v-show="!infomation.speakers || infomation.speakers.length === 0" type="button" value="选择扬声器" />
    </div>
    <a-modal :visible="isShowEqualizer" width="fit-content" :footer="null" @cancel="isShowEqualizer = false">
      <equalizer id="equalizer" :bands="equalizerBands" :eq="line.eq.eqs" ref="equalizer" @change="onEQChange" />
      <template slot="title">
        <div class="eq-toolbar">
          <span>均衡器</span>
          <a-button type="danger" shape="circle" icon="rest" @click="onEQClear()"></a-button>
          <a-switch v-model="line.eq.enable" checked-children="开" un-checked-children="关" @change="onEqEnable" />
          <a-select class="eq-band-select" :options="equalizerBandsList" :value="eqBandsSelected + ''"
            @select="eqBandsSelected = $event">
          </a-select>
        </div>
      </template>
    </a-modal>
    <div class="select-channel-mask" v-show="specifyChannel >= 0"
      :class="{ 'animate__animated animate__fadeIn': specifyChannel >= 0 }">
      <a-button icon="close" @click.stop.prevent="specifyChannel = -1"></a-button>
    </div>
  </div>
</template>

<script>
import { mapState } from 'vuex';
import VueSlider from 'vue-slider-component';
import 'vue-slider-component/theme/antd.css';
import Equalizer from '@/components/Equalizer';
import Volume from '@/components/Volume';
// import icons for win32 title bar
// icons by https://github.com/microsoft/vscode-codicons
import '@vscode/codicons/dist/codicon.css';
import * as ApiLine from '@/api/line';
import { socket, Event } from '@/common/request';
import VolumeLevel from '@/common/volumeLevel';
import * as ApiSpeaker from '@/api/speaker';
import { createPopper } from '@popperjs/core';
import { formatRate, formatBits, formatLayout, formatDuration } from '@/common/format';
import '@/assets/css/popper.scss';
import '@/assets/css/speaker.scss';

const layoutDefault = '2-0';

export default {
  name: 'Speaker',
  components: {
    VueSlider,
    Equalizer,
    Volume,
  },
  data() {
    return {
      line: { id: 0, name: '', newName: '', eq: { enable: false, eqs: [] }, vol: 0, mute: false },
      isLineNameEdit: false,
      isLayout3d: false,
      layout: layoutDefault,
      layoutList: [
        { key: '1-0', label: '单声道  ', disabled: false },
        { key: '2-0', label: '立体声  ', disabled: false },
        { key: '2-1', label: '2.1 声道', disabled: false },
        { key: '5-0', label: '5.0 声道', disabled: false },
        { key: '5-0-back-', label: '5.0(后) 声道', disabled: false },
        { key: '5-1', label: '5.1 声道', disabled: false },
        { key: '5-1-back-', label: '5.1(后) 声道', disabled: false },
        { key: '7-0', label: '7.0 声道', disabled: false },
        { key: '7-1', label: '7.1 声道', disabled: false },
        { key: '7-1-2', label: '7.1.2 声道', disabled: false },
        { key: '7-1-4', label: '7.1.4 声道', disabled: false },
      ],
      channels: 0,
      channelLayout: 'none',
      channelSpeakers: {},
      channelAttr: ApiLine.channelList,
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
      specifyChannel: -1, // speakerid
      fromAction: '',
      shownChannelId: 0,
      specturmCtx: {},
    };
  },
  watch: {
    $route(to, from) {
      if (to.params.action === 'specifychannel') {
        this.specifyChannel = parseInt(to.params.spid);
        return;
      } else if (this.fromAction === 'specifychannel') {
        this.specifyChannel = -1;
        return;
      }
    },
    isShowEqualizer(newVal) {
      if (!newVal) return;
      this.$nextTick(() => {
        // 等待页面加载完成
        let bands = [];
        this.$refs.equalizer.bandList().forEach(band => {
          bands.push({
            label: band + ' 段',
            value: band,
            key: band + ' 段',
          });
        });

        this.equalizerBandsList = bands;
        this.eqBandsSelected = bands[0].key;
      });
    },
    specifyChannel(newVal) {
      this.$root.$emit('scrollTo', 0);

      // 选择声道的时候禁止滚动
      this.scrolling = newVal < 0;
    },
  },
  beforeRouteEnter(to, from, next) {
    next(vm => {
      // 通过 `vm` 访问组件实例,将值传入fromPath
      vm.fromAction = from.params ? from.params.action : '';
    });
  },
  computed: {
    ...mapState(['enableScrolling', 'settings']),
    scrolling: {
      get() {
        return this.enableScrolling || false;
      },
      set(value) {
        this.$store.commit('enableScrolling', value);
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
        ApiLine.clearEqualizer(this.line.id).then(() => {
          this.line.eq.eqs = []
        });
      },
    },
    isPlayer() {
      return this.line.source && this.line.source.type === 1;
    },
    currentDuration() {
      return formatDuration(this.line.source.cur)
    },
    totalDuration() {
      return formatDuration(this.line.source.dur)
    },
    playerPosition: {
      get() {
        return this.line.source && this.line.source.cur
      },
      set(val) {
        this.startPlayerDuration()
      }
    }
  },
  activated() {
    this.$root.$emit('scrollTo', 0);
    // keep-alived 开启后生效
    socket.onConnected().then(() => this.init());
  },
  deactivated() {
    // keep-alived 开启后生效
    this.deinit();
  },
  methods: {
    deinit() {
      ApiLine.removelistenLineChanged();
      ApiLine.removeListenLineSpectrum();
      ApiLine.removeListenLineInput();
      ApiSpeaker.removeListenSpeakerSpectrum();
      ApiLine.removeListenLineSpeakerChanged();

      cancelAnimationFrame(this.specturmCtx.spRequestId);
      this.specturmCtx.ctx = null;
      this.specturmCtx.SP = null;
      this.specturmCtx.level.clear();

      this.$destroyAll();
      document.removeEventListener('keyup', this.onDocumentKeyUp);
      clearTimeout(this.playerDurationTimeout);
    },
    init() {
      this.line.id = parseInt(this.$route.params.id);
      this.$set(this.line, 'eq', []);
      this.specifyChannel = -1;
      this.isLineNameEdit = false;
      this.infomation = {};
      this.$destroyAll();
      this.$nextTick(function () {
        if (this.settings.showSpectrum) {
          this.initSpectrum();
        }
        document.addEventListener('keyup', this.onDocumentKeyUp);
      });

      ApiLine.listenLineChanged(this.line.id, line => {
        this.line = line;
      });
      ApiLine.listenLineSpeakerChanged(this.line.id, (speaker, evt, sub) => {
        let found = -1;
        for (let i = 0; i < this.line.speakers.length; i++) {
          let sp = this.line.speakers[i];
          if (sp.id === speaker.id) {
            found = i;
            break;
          }
        }
        switch (sub) {
          case Event.SP_Detected:
            speaker.__class = 'animate__bounceIn';
            this.line.speakers.unshift(speaker);
            break;
          case Event.SP_Deleted:
            if (found >= 0) {
              this.speakers[found].__class = 'animate__bounceOut';
              setTimeout(() => {
                this.speakers.splice(found, 1);
              }, 750);
            }
            break;
          case Event.SP_Online:
          case Event.SP_Offline:
          case Event.SP_Edited:
          case Event.SP_Edited:
            if (found < 0) {
              speaker.__class = 'animate__bounceIn';
              this.line.speakers.unshift(speaker);
            } else {
              this.speakers[found] = speaker;
            }
            break;
        }
      });

      if (this.settings.showSpectrum) {
        ApiLine.listenLineSpectrum(this.line.id, this.onSpectrumChange);
      } else {
        // this.onSpectrumChange({ l: [this.line.id, 0], s: [] })
        cancelAnimationFrame(this.specturmCtx.spRequestId)
        // this.specturmCtx.level.clear();
      }

      ApiLine.listenLineInput(this.line.id, s => {
        this.$set(this.line, 'source', s)
        this.startPlayerDuration()
      })
      ApiSpeaker.listenSpeakerSpectrum(levels => {
        levels.forEach(s => level.setValById(s[0], s[1]));
      });

      ApiLine.getLineInfo(this.line.id)
        .then(data => {
          data = data || {};
          if (data.id === undefined) {
            console.log('line info error', data);
            this.$router.push('/speakers');
            return;
          }

          this.line = data;
          if (data.source) {
            if (data.source.cur > data.source.dur) {
              data.source.cur = data.source.dur;
            }
            this.startPlayerDuration()
          }

          this.initChannelSpeakers();
          this.onShowChannelInfo(-1);

          this.$nextTick(() => {
            this.specturmCtx.level.clear();
            this.specturmCtx.level.push(
              'line-' + this.line.id,
              document.querySelector(
                '#line-' + this.line.id + ' .vue-slider-process'
              )
            );
            this.line.speakers.forEach((sp, i) => {
              this.specturmCtx.level.push(i,
                document.querySelector('#speaker-' + sp.id + ' .vue-slider-process')
              );
            });
          });
        })
        .catch(e => {
          console.log(e);
          this.$router.push('/speakers');
        });
    },
    initSpectrum() {
      cancelAnimationFrame(this.specturmCtx.spRequestId);

      this.specturmCtx.SP = document.getElementById('spectrum-' + this.line.id);
      let bg = this.specturmCtx.SP.nextSibling;
      this.specturmCtx.SP.width = bg.offsetWidth;
      this.specturmCtx.SP.height = bg.offsetHeight;
      this.specturmCtx.ctx = this.specturmCtx.SP.getContext('2d');
      this.specturmCtx.slow = [];
      this.specturmCtx.title = [];
      this.specturmCtx.spdata = [];
      this.specturmCtx.speakerIndex = 0;
      this.specturmCtx.level = new VolumeLevel('200ms');

      this.drawSpectrum();
    },
    onDocumentKeyUp(e) {
      if (e.key === 'Escape') {
        this.isLineNameEdit = false;
        this.specifyChannel = -1;
      }
    },
    initChannelSpeakers() {
      this.line.speakers = this.line.speakers || [];
      this.channelSpeakers = {};
      let speakers = {};
      this.line.speakers.forEach((sp, i) => {
        if (!sp.ch) return;
        if (speakers[sp.ch] === undefined) {
          speakers[sp.ch] = [];
        }
        speakers[sp.ch].push(sp);
      });
      this.channelSpeakers = speakers;

      for (var i in this.channelAttr) {
        if (this.channelAttr[i].show) {
          this.$set(this.channelAttr[i], 'show', false);
        }
      }
      this.computeLayout();
    },
    startPlayerDuration() {
      clearTimeout(this.playerDurationTimeout)
      if (!this.isPlayer || this.line.source.cur >= this.line.source.dur) {
        return
      }
      this.playerDurationTimeout = setInterval(() => {
        ApiLine.getLinePlayer(this.line.id).then(data => {
          if (data.cur > data.dur) {
            data.cur = data.dur;
          }
          if (data.cur >= data.dur) {
            clearTimeout(this.playerDurationTimeout)
          }
          this.$set(this.line, 'source', data)
        })
      }, 1000)
    },
    onPlayerSeekStart() {
      clearTimeout(this.playerDurationTimeout)
    },
    onPlayerSeekStop() {
      ApiLine.playerSeek(this.line.id, this.line.source.cur)
      this.startPlayerDuration()
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
      const lm = data['l'] || [];
      this.specturmCtx.spdata = data['s'] || [];
      this.specturmCtx.level.setValById('line-' + lm[0], lm[1]);
    },
    computeLayout() {
      let known = false;
      let layout = this.line.layout.split(/[\.|\(|\)| ]/).join('-');
      if (layout === 'mono') layout = '1-0';
      else if (layout === 'stereo') layout = '2-0';
      for (let i = 0; i < this.layoutList.length; i++) {
        if (layout === this.layoutList[i].key) {
          known = true;
          break;
        }
      }
      if (known)
        this.layout = layout;
      else
        this.layout = layoutDefault;

      this.disableLayoutSelect();
    },
    disableLayoutSelect() {
      for (let i = this.layoutList.length - 1, s = -1; i >= 0; i--) {
        if (this.layoutList[i].key === this.layout) {
          s = i;
        } else if (s > 0) {
          this.layoutList[i].disabled = true;
          continue;
        }
        this.layoutList[i].disabled = false;
      }
    },
    onLayoutChanged(layout) {
      this.layout = layout
    },
    onVolumeChanged(v) {
      ApiLine.setVolume(this.line.id, v).then(() => {
        if (typeof v === 'boolean') {
          this.line.mute = v
        } else {
          this.line.vol = v
        }
      });
    },
    onSpeakerVolumeChanged(v, id) {
      ApiSpeaker.setVolume(id, v);
    },
    onSpeakerVolumeMute(v, id) {
      ApiSpeaker.setSpeaker(id, 'mute', v);
    },
    onEQChange(freq, gain) {
      ApiLine.setEqualizer(this.line.id, freq, gain).then((s) => {
        this.line.eq = s
      }).catch(() => {
        this.line.eq.eqs = this.line.eq.eqs
      });
    },
    onEqEnable(enable) {
      ApiLine.setEnableEqualizer(this.line.id, enable);
    },
    onEQClear() {
      ApiLine.clearEqualizer(this.line.id).then(() => {
        this.line.eq.eqs = [];
      });
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
    },
    onRemoveSPChannel(speaker) {
      let that = this;
      this.$confirm({
        title: '确定要移除该扬声器所关联的声道吗？',
        content: h =>
          h(
            'div',
            { style: 'color:red;' },
            '移除关联的声道之后，该扬声器将处于空闲状态。可再次点击“指定声道”按钮以重新指定。'
          ),
        okText: '是',
        okType: 'danger',
        cancelText: '否',
        onOk() {
          ApiSpeaker.setSpeaker(speaker.id, 'ch', -1).then(() => {
            that.changeSpeakerAttrById(speaker.id, 'ch', 0);
            that.onShowChannelInfo(that.shownChannelId);
          });
        },
      });
    },
    onSelectChannel(ch) {
      let spid = this.specifyChannel;
      ch = parseInt(ch);
      if (spid >= 0) {
        ApiSpeaker.setSpeaker(spid, 'ch', ch)
          .then(() => {
            this.changeSpeakerAttrById(spid, 'ch', ch);
          })
          .finally(() => {
            // this.$router.go(-1);
            this.specifyChannel = -1;
          });
        return;
      }
      this.onShowChannelInfo(ch);
    },
    showChannelName(ch) {
      if (ch in this.channelAttr) return this.channelAttr[ch].name;
      return '';
    },
    onShowChannelInfo(chid) {
      this.shownChannelId = chid;
      if (!(chid in this.channelAttr)) {
        this.infomation = {
          speakers: this.line.speakers || [],
        };
        for (var k in this.channelAttr) {
          if (this.channelAttr[k].show) {
            this.$set(this.channelAttr[k], 'show', false);
            this.onChannelMouseHover(k, false);
          }
        }
        return;
      }

      let ch = this.channelAttr[chid];
      this.infomation = {
        channelId: chid,
        name: ch.name,
        speakers: this.channelSpeakers[chid],
      };

      for (var k in this.channelAttr) {
        if (k == chid) {
          this.$set(this.channelAttr[k], 'show', true);
          this.onChannelMouseHover(k, true);
        } else if (this.channelAttr[k].show) {
          this.$set(this.channelAttr[k], 'show', false);
        }
      }
    },
    showSourceFormat(src) {
      if (src === undefined) return '';
      return (
        formatRate(src.rate) + '/' + src.bits + '/' + src.layout
      );
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
    speakerIds() {
      return (this.line.speakers || []).map(s => { return s.id })
    },
    gotoSpeaker(id) {
      this.$router.push('/speaker/' + id);
    },
    onChannelTest(chid) {
      ApiLine.testChannel(this.line.id, chid)
    },
    onIamHere(spid) {
      ApiSpeaker.test(spid)
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
    onVolumeMute() {
      this.$set(this.line, 'mute', !this.line.mute);
    },
    onNameChange() {
      if (this.line.name === this.line.newName) {
        this.isLineNameEdit = false;
        return;
      }
      ApiLine.setLine(this.line.id, 'name', this.line.newName)
        .then(() => {
          this.line.name = this.line.newName;
        })
        .catch(() => {
          this.line.newName = this.line.name;
        })
        .finally(() => {
          let nav = document.getElementById('nav-' + this.line.id);
          nav.innerText = this.line.name;
          this.isLineNameEdit = false;
        });
    },
    playerSliderFormater(val) {
      return formatDuration(val)
    },
    drawSpectrum() {
      const ctx = this.specturmCtx.ctx;
      const SP = this.specturmCtx.SP;
      const spdata = this.specturmCtx.spdata;
      const hd = this.specturmCtx.hd;
      let slow = this.specturmCtx.slow;
      let title = this.specturmCtx.title;

      ctx.clearRect(0, 0, SP.width, SP.height);
      const spc = Math.max(16, spdata.length);
      let w = SP.width / spc;
      let space = w > 1 ? w * 0.1 : 0;
      let left = 0;
      let spd = 0;
      let sum = 0;
      w -= space;
      for (var i = 0; i < spc; i++) {
        left = i * (w + space);
        spd = spdata[i]
        // spd *= 0.5 * (1 - Math.cos(2*Math.PI*i/spc-1))
        if (spd >= SP.height) {
          spd = SP.height - 1;
        }
        sum += spd;
        if (hd && spd < 1) {
          spd = 1
        }

        ctx.beginPath();
        ctx.lineWidth = w;
        ctx.strokeStyle = 'hsl(171deg 100% 41% / 20%)';

        if (slow[i] > spd) {
          slow[i] -= 2;
        } else if (spd - slow[i] > 8) {
          slow[i] += 5;
        } else {
          slow[i] = spd;
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
        ctx.lineWidth = w;
        ctx.strokeStyle = '#bbb';
        ctx.moveTo(left, SP.height - title[i]);
        ctx.lineTo(left, SP.height - title[i] + 1);
        ctx.stroke();
      }
      this.specturmCtx.hd = sum > 0

      if (this.specturmCtx.speakerIndex > this.specturmCtx.level.length) {
        this.specturmCtx.speakerIndex = 0;
      }

      if (this.specturmCtx.level.length) {
        this.specturmCtx.level.commitWidth(this.specturmCtx.speakerIndex);
        this.specturmCtx.speakerIndex++;
      }

      this.specturmCtx.spRequestId = requestAnimationFrame(this.drawSpectrum);
    }
  },
};
</script>

<style lang="scss">
@import '@/assets/css/channels-layout.scss';

.line-page {
  position: relative;

  .player {
    padding: 1rem 3rem;
    display: flex;
    flex-direction: row;
    align-items: center;
    justify-items: center;

    .vue-slider {
      flex-grow: 1;

      .vue-slider-dot {
        height: 24px !important;
        width: 8px !important;

        .vue-slider-dot-handle {
          border-radius: 0;
        }
      }
    }

    >span {
      margin: 0 1rem;
    }
  }

  .volume {
    padding: 1rem 3rem;
    padding-top: 2rem;
    display: flex;
    flex-wrap: nowrap;
    z-index: 1;
    position: relative;

    >div:first-child {
      flex: 0 0 auto;
      margin-right: 1rem;
    }

    >div:last-child {
      flex: 1 1 auto;
      display: flex;
      flex-wrap: wrap;
      align-items: center;

      .volume-controller {
        flex: 1 0 8rem;
        margin-right: 1rem;

        .vue-slider {
          min-width: 5rem;
        }
      }
    }

    .line-name {
      width: auto;
      display: flex;
      flex-direction: row;
      justify-content: flex-start;
      flex-wrap: nowrap;

      >input,
      >span {
        max-width: 8rem;
        padding: 0 4px;
        height: 1.5rem;
        line-height: 1.5rem;
        width: auto;
      }

      >button {
        width: 1rem;
        padding-left: 0;
      }
    }

    .source {
      display: block;
    }

    .layout-select {}
  }

  .container {
    position: relative;
    display: flex;
    justify-content: center;
    align-items: center;
    // width: 480px;
    // height: 320px;
    margin: 0 auto;
    padding-bottom: 4rem;
    overflow: hidden;

    .svg-icon {
      display: block;
    }

  }

  .infomation {
    display: block;
    position: relative;
    margin: 3rem 5rem;
    padding: 10px;
    background-color: var(--color-secondary-bg);
    border-radius: 5px;

    .svg-icon {
      cursor: pointer;
    }

    >.channel-name {
      text-align: center;
      font-weight: bold;
      position: relative;
      display: block;
      height: 2rem;
      margin-bottom: 1rem;
      margin-top: -0.5rem;

      >div {
        display: inline-block;
        z-index: 1;
        position: relative;
        color: var(--color-text);
        top: 0;
        display: flex;
        align-items: center;
        justify-items: center;
        width: fit-content;
        margin: 0 auto;

        >span {
          margin-left: 5px;
        }

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

        .svg-icon {
          width: 1rem;
          height: 1rem;
        }
      }
    }

  }

  .select-channel-mask {
    --animate-duration: 0.5s;

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
          color: var(--color-body-bg);
        }

        &.enabled {
          .svg-icon {
            color: var(--color-border);
          }
        }
      }
    }
  }

  #popper {
    z-index: 6;
  }

  @include for_breakpoint(mobile) {
    .infomation {
      margin-left: 1rem;
      margin-right: 1rem;
    }

    .player,
    .volume {
      padding: 1rem 1rem;
    }
  }

  @include for_breakpoint(small) {
    .infomation {
      margin-left: 0;
      margin-right: 0;
      padding: 0;
    }
  }

  @include for_breakpoint(tablet) {
    .infomation {
      margin-left: 3rem;
      margin-right: 3rem;
    }
  }

  @include for_breakpoint(desktop) {
    // .container {
    //   width: 640px;
    //   height: 320px;
    //   margin-top: 0rem;
    // }

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

    .player,
    .volume {
      padding: 1rem 6rem;
    }
  }
}

.eq-toolbar {
  display: flex;
  flex-direction: row;
  position: relative;
  width: calc(100% - 5rem);

  >span {
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
