<template>
  <div class="speaker-list">
    <div v-for="sp in speakers" :key="sp.id" :id="'speaker-' + sp.id">
      <SpeakerRow :speaker="sp" />
    </div>
    <div class="notification is-primary is-light" v-if="speakers.length === 0">
      当前还没有连接任何的扬声器。
    </div>
  </div>
</template>

<script>
import { mapState } from 'vuex';
import SpeakerRow from '@/components/SpeakerRow';
import VolumeLevel from '@/common/volumeLevel';
import {
  getSpeakerList,
  listenSpeakerChanged,
  listenSpeakerLevelMeter,
  removeListenSpeakerEvent,
  removeListenSpeakerLevelMeter,
} from '@/api/speaker';
import { socket } from '@/common/request';

let level = new VolumeLevel('200ms');
window.level = level;

let frameIndex = 0;
function renderVolumeLevel() {
  if (frameIndex > level.length) {
    frameIndex = 0;
  }

  if (level.length) {
    level.commitWidth(frameIndex);
    frameIndex++;
  }
  renderVolumeLevel.timer = window.requestAnimationFrame(renderVolumeLevel);
}

export default {
  name: 'Speakers',
  components: { SpeakerRow },
  data() {
    return {
      show: false,
      speakers: [],
      readyCount: 0,
    };
  },
  computed: {
    ...mapState(['settings']),
  },
  mounted() {
    socket.onConnected().then(() => this.loadData());
    this.$parent.$refs.scrollbar.restorePosition();

    window.requestAnimationFrame(renderVolumeLevel);
  },
  destroyed() {
    removeListenSpeakerEvent();
    removeListenSpeakerLevelMeter();
  },
  watch: {
    speakers(newVal, oldVal) {
      cancelAnimationFrame(renderVolumeLevel.timer);
      level.clear();
      this.$nextTick(() => {
        newVal.forEach(s => {
          level.push(s.id, document.getElementById('speaker-' + s.id).querySelector('.vue-slider-process'));
        })
      })
    },
  },
  methods: {
    loadData() {
      getSpeakerList().then(result => {
        this.speakers = result || [];
      });
      listenSpeakerChanged((act, speaker) => {
        let i = this.findSpeaker(speaker.id);
        switch (act) {
          case 1:
            this.speakers.unshift(speaker);
            break;
          case 2:
            if (i >= 0) this.speakers[i] = speaker;
            else this.speakers.unshift(speaker);
            break;
          case 3:
            if (i >= 0) this.speakers[i] = speaker;
            break;
          case 4:
            if (i >= 0) delete this.speakers[i];
            break;
        }
      });
      listenSpeakerLevelMeter(levels => {
        levels.forEach(s => level.setValById(s[0], s[1]));
      });
    },
    findSpeaker(id) {
      let s;
      for (let i = 0; i < this.speakers.length; i++) {
        s = this.speakers[i];
        if (s.id === id) {
          return i;
        }
      }
      return -1;
    },
  },
};
</script>

<style lang="scss">
.speaker-list {
  .notification {
    justify-content: center;
    height: 100%;
    position: fixed;
    width: 100%;
    align-items: center;
    display: flex;
    top: initial;
    z-index: -1;
  }
}

.vue-slider-process {
  width: 100% !important;
  left: 0 !important;
  transition: none !important;
  overflow: hidden;
  position: relative;
  background-color: transparent !important;
  .level-meter {
    position: absolute;
    width: 100%;
    height: 100%;
    top: 0;
    right:0;
    background-color: pink;
    transition-property: right;
    transition-duration: 200ms;
  
  }
}
</style>
