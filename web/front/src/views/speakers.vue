<template>
  <div class="speaker-list">
    <div v-for="(sp, i) in speakers" :key="i" :id="'speaker-' + sp.id" class="speaker">
      <SpeakerRow :speaker="sp" :class="sp.__class ? sp.__class : ''" />
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
import * as ApiSpeaker from '@/api/speaker';
import { socket, Event as SrvEvent } from '@/common/request';

let level = new VolumeLevel('200ms');

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
    ApiSpeaker.removeListenSpeakerEvent(undefined);
    ApiSpeaker.removeListenSpeakerSpectrum(undefined);
    level.clear();
  },
  watch: {
    speakers(newVal, oldVal) {
      cancelAnimationFrame(renderVolumeLevel.timer);
      level.clear();
      this.$nextTick(() => {
        newVal.forEach(s => {
          level.push(
            s.id,
            document
              .getElementById('speaker-' + s.id)
              .querySelector('.vue-slider-process')
          );
        });
      });
    },
  },
  methods: {
    loadData() {
      ApiSpeaker.getSpeakerList().then(result => {
        this.speakers = result || [];
      });
      ApiSpeaker.listenSpeakerChanged(undefined, (speaker, evt) => {
        let i = this.findSpeaker(speaker.id);
        switch (evt) {
          case SrvEvent.SP_Detected:
            speaker.__class = 'animate__bounceIn';
            this.speakers.unshift(speaker);
            break;
          case SrvEvent.SP_Edited:
          case SrvEvent.SP_Online:
          case SrvEvent.SP_Offline:
          case SrvEvent.SP_Moved:
            if (i >= 0) this.speakers[i] = speaker;
            break;
          case SrvEvent.SP_Deleted:
            if (i >= 0) {
              this.speakers[i].__class = 'animate__bounceOut';
              setTimeout(() => {
                this.speakers.splice(i, 1);
              }, 750);
            }
            break;
        }
      });
      ApiSpeaker.listenSpeakerSpectrum(undefined, levels => {
        levels.forEach(s => level.setValById(s[0], s[1]));
      });
    },
    speakerIds() {
      return this.speakers.map(s => {return s.id})
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
  .speaker {
    width: 100%;
    height: auto;
  }

  .notification {
    justify-content: center;
    height: 100%;
    position: fixed;
    width: 100%;
    align-items: center;
    display: flex;
    top: initial;
    background-color: var(--color-secondary-bg);
    // z-index: -1;
  }
}
</style>
