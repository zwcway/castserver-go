<template>
  <div class="line-page" :class="{ 'select-channel': specifyChannel >= 0 }">
    <div class="player" v-if="isPlayer">
      <span>{{ currentDuration }}</span>
      <VueSlider v-model="line.source.cur" :min="0" :max="line.source.dur" tooltipPlacement="bottom"
        :tooltipFormatter="playerSliderFormater" @drag-start="onPlayerSeekStart" @drag-end="onPlayerSeekStop" />
      <span>{{ totalDuration }}</span>
    </div>
    <div class="volume">
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
          <span class="ratebits" v-show="line.source">{{
            showSourceFormat(line.source)
          }}</span>
        </span>
      </div>
      <div :id="'line-' + line.id" class="level-meter-slider">
        <Volume :volume="volume" :mute="line.mute || false" @change="onVolumeChanged" tooltip-placement="bottom" />
        <a-select :value="layout" class="layout-select" :options="layoutList" @select="layout = $event"></a-select>
        <a-button type="link" @click.stop.prevent="isShowEqualizer = true">
          <i class="codicon codicon-settings"></i>
        </a-button>
      </div>
    </div>
    <div class="container" @click="onShowChannelInfo(-1)">
      <canvas id="spectrum" class="spectrum" height="0" width="0"></canvas>
      <div class="background" rel="background"></div>
      <div class="room">
        <div class="line"></div>
        <div class="line"></div>
        <div class="line"></div>
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
        <span>{{ infomation.name }}</span>
      </div>
      <ul class="speaker-list" v-show="infomation.speakers && infomation.speakers.length">
        <li class="speaker level-meter-slider" v-for="sp in infomation.speakers" :key="sp.id" :id="'speaker-' + sp.id">
          <div>
            <svg-icon icon-class="speaker" :size="0"></svg-icon>
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
      <equalizer id="equalizer" :bands="equalizerBands" :eq="line.eq" ref="equalizer" @change="onEQChange" />
      <template slot="title">
        <div class="eq-toolbar">
          <span>均衡器</span>
          <a-switch v-model="line.eqenable" checked-children="开" un-checked-children="关" @change="onEqEnable" />
          <a-select class="eq-band-select" :options="equalizerBandsList" :value="eqBandsSelected"
            @select="eqBandsSelected = $event">
          </a-select>
        </div>
      </template>
    </a-modal>
    <div class="select-channel-mask" v-show="specifyChannel >= 0">
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

var SP, ctx;
let spdata = [];
let slow = [];
let title = [];
let spRequestId;
let speakerIndex = 0;
let level = new VolumeLevel('200ms');
let playerDurationTimeout;

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

  if (speakerIndex > level.length) {
    speakerIndex = 0;
  }

  if (level.length) {
    level.commitWidth(speakerIndex);
    speakerIndex++;
  }

  spRequestId = requestAnimationFrame(drawSpectrum);
}

export default {
  name: 'Speaker',
  components: {
    VueSlider,
    Equalizer,
    Volume,
  },
  data() {
    return {
      line: { id: 0, name: '', newName: '' },
      isLineNameEdit: false,
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
      this.init();
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
    specifyChannel(newVal) {
      // 选择声道的时候禁止滚动
      this.$root.$emit('scrollTo', 0);

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
    ...mapState(['enableScrolling']),
    scrolling: {
      get() {
        return this.enableScrolling || false;
      },
      set(value) {
        this.$store.commit('enableScrolling', value);
      },
    },
    volume: {
      get() {
        return this.line.vol || 0;
      },
      set(value) {
        this.line.vol = value;
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
  mounted() {
    socket.onConnected().then(() => this.init());
  },
  destroyed() {
    this.deinit();
  },
  activated() {
    // keep-alived 开启后生效
    this.init();
  },
  methods: {
    deinit() {
      cancelAnimationFrame(spRequestId);
      ApiLine.removelistenLineChanged();
      ApiLine.removeListenLineSpectrum();
      ApiLine.removeListenLineInput();
      ApiSpeaker.removeListenSpeakerSpectrum();
      ApiLine.removeListenLineSpeakerChanged();
      level.clear();
      this.$destroyAll();
      document.removeEventListener('keyup', this.onDocumentKeyUp);
      clearTimeout(playerDurationTimeout);
    },
    init() {
      this.deinit();
      this.line.id = parseInt(this.$route.params.id);
      this.$set(this.line, 'eq', []);
      this.specifyChannel = -1;
      this.isLineNameEdit = false;
      this.infomation = {};
      this.initSpectrum();
      this.$destroyAll();
      this.$nextTick(function () {
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

      ApiLine.listenLineSpectrum(this.line.id, this.onSpectrumChange);
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
            level.clear();
            level.push(
              'line-' + this.line.id,
              document.querySelector(
                '#line-' + this.line.id + ' .vue-slider-process'
              )
            );
            this.line.speakers.forEach((sp, i) => {
              level.push(
                i,
                document.querySelector(
                  '#speaker-' + sp.id + ' .vue-slider-process'
                )
              );
            });
          });
        })
        .catch(e => {
          console.log(e);
          this.$router.push('/speakers');
        });
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
      clearTimeout(playerDurationTimeout)
      if (this.line.source.cur >= this.line.source.dur) {
        return
      }
      playerDurationTimeout = setInterval(() => {
        ApiLine.getLineInfo(this.line.id).then(data => {
          if (data.source.cur > data.source.dur) {
            data.source.cur = data.source.dur;
          }
          if (data.source.cur >= data.source.dur) {
            clearTimeout(playerDurationTimeout)
          }
          this.$set(this.line, 'source', data.source)
        })
      }, 1000)
    },
    onPlayerSeekStart() {
      clearTimeout(playerDurationTimeout)
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
      spdata = data['s'] || [];
      level.setValById('line-' + lm[0], lm[1]);
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
          continue;
        }
        this.layoutList[i].disabled = false;
      }
    },
    onVolumeChanged(v) {
      ApiLine.setVolume(this.line.id, v);
    },
    onSpeakerVolumeChanged(v, id) {
      ApiSpeaker.setVolume(id, v);
    },
    onSpeakerVolumeMute(v, id) {
      ApiSpeaker.setSpeaker(id, 'mute', v);
    },
    initSpectrum() {
      cancelAnimationFrame(spRequestId);

      SP = document.getElementById('spectrum');
      let bg = SP.nextSibling;
      SP.width = bg.offsetWidth;
      SP.height = bg.offsetHeight;
      ctx = SP.getContext('2d');
      drawSpectrum();
    },
    onEQChange(freq, gain) {
      ApiLine.setEqualizer(this.line.id, freq, gain);
    },
    onEqEnable(enable) {
      ApiLine.setEnableEqualizer(this.line.id, enable);
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
        formatRate(src.rate) + '/' + src.bits + '/' + formatLayout(src.channels)
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

      .volume-controller {
        flex: 1 0 8rem;

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

    .layout-select {
      margin-left: 2rem;
    }
  }

  .container {
    position: relative;
    display: flex;
    justify-content: center;
    align-items: center;
    width: 480px;
    height: 320px;
    margin: 0 auto;
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

  }

  .infomation {
    display: block;
    position: relative;
    margin: 3rem 5rem;
    padding: 10px;
    background-color: var(--color-secondary-bg);
    border-radius: 5px;

    >.channel-name {
      text-align: center;
      font-weight: bold;
      position: relative;
      display: block;
      height: 2rem;
      margin-bottom: 1rem;
      margin-top: -0.5rem;

      >span {
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

  @media only screen and (max-width: 479px) {
    .container {
      width: 320px;
      height: 240px;

      .room {}

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

    .player,
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
