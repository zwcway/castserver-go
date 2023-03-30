<template>
  <div class="container">
    <div class="card">
      <div class="volume">
        <div class="speaker-name">
          <span v-if="!isSpeakerNameEdit">{{ speaker.name }}</span>
          <a-input v-else :value="speaker.name" placeholder="名称" :max-width="6"
            @change="speaker.newName = $event.target.value" @blur="onNameChange" @keyup.enter="onNameChange" />
          <a-button v-if="!isSpeakerNameEdit" type="link" @click.stop.prevent="isSpeakerNameEdit = true">
            <a-icon type="edit" />
          </a-button>
          <a-button v-else type="link" @click.stop.prevent="onNameChange">
            <i class="codicon codicon-check"></i>
          </a-button>
        </div>
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
          <label>{{ $t('position') }}</label>
          <span>
            <a-select size="small" v-model="line" :options="lineListOptions" :placeholder="$t('select line')"
              :dropdownMatchSelectWidth="false" />
            <a-select size="small" v-model="channel" :options="channelListOptions" :placeholder="$t('select channel')"
              :dropdownMatchSelectWidth="false" />
          </span>
        </div>
        <div class="column">
          <label>{{ $t('connection state')}}</label>
          <span :class="speaker.cTime ? 'success' : 'warn'">{{ speaker.cTime ? $t('connected') : $t('disconnected') }}</span>
        </div>
        <div class="column block">
          <label>{{ $t('sample rates supported') }}</label>
          <span class="tags">
            <a-tag class="tag" v-for="(rate, i) in speaker.rateList" :key="i"> {{ rate | num }}Hz </a-tag>
          </span>
        </div>
        <div class="column block">
          <label>{{ $t('sample bits supported') }}</label>
          <span class="tags">
            <a-tag class="tag" v-for="(bit, i) in speaker.bitList" :key="i"> {{ bit }} </a-tag>
          </span>
        </div>
        <div class="column">
          <label>{{ $t('queued size') }}</label>
          <span>{{ speaker.statistic ? speaker.statistic.q : 0 | bytes }}</span>
        </div>
        <div class="column">
          <label>{{ $t('sended size') }}</label>
          <span>{{ speaker.statistic ? speaker.statistic.s : 0 | bytes }}</span>
        </div>
        <div class="column">
          <label>{{ $t('droped size') }}</label>
          <span>{{ speaker.statistic ? speaker.statistic.d : 0 | bytes }}</span>
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
      isSpeakerNameEdit: false,
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
      let opts = [{ key: '-1', label: '选择声道' }]
      for (let i in ApiLine.channelList) {
        opts.push({
          key: i + '',
          label: ApiLine.channelList[i].name
        })
      }
      return opts
    },
    lineListOptions() {
      let opts = [{ key: '-1', label: '选择线路' }]
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
    onNameChange() {
      if (!this.speaker.newName || this.speaker.name === this.speaker.newName || this.speaker.newName.length === 0) {
        this.isSpeakerNameEdit = false;
        return;
      }
      ApiSpeaker.setSpeaker(this.speaker.id, 'name', this.speaker.newName)
        .then(() => {
          this.speaker.name = this.speaker.newName;
        })
        .catch(() => {
          this.speaker.newName = this.speaker.name;
        })
        .finally(() => {
          this.isSpeakerNameEdit = false;
        });
    },
  },
};
</script>

<style lang="scss" scoped>
.container {
  margin: 1rem 2rem;
}

.speaker-name {
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
    color: var(--color-text);
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
        // padding: 1px 3px;
        // margin: 0 0.2rem;
        // line-height: 1.1rem;
        // height: 1rem;
        border-radius: 0.2rem;
        background-color: var(--color-secondary-bg-for-transparent);
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
