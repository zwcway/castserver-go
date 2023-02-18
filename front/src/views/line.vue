<template>
  <div class="container">
    <div class="room">
      <div class="line"></div>
      <div class="line"></div>
      <div class="line"></div>
    </div>
    <div class="channels channels-layout layout-2">
      <div class="speaker front-left">
        <svg-icon icon-class="speaker-front-lr" :size="0" />
      </div>
      <div class="speaker front-right">
        <svg-icon icon-class="speaker-front-lr" :size="0" />
      </div>
      <div class="speaker front-center">
        <svg-icon icon-class="speaker-front-center" :size="0" />
      </div>
      <div class="speaker front-bass">
        <svg-icon icon-class="speaker-low-frequency" :size="0" />
      </div>
      <div class="speaker side-left">
        <svg-icon icon-class="speaker-side-lr" :size="0" />
      </div>
      <div class="speaker side-right">
        <svg-icon icon-class="speaker-side-lr" :size="0" />
      </div>
      <div class="speaker back-left">
        <svg-icon icon-class="speaker-back-lr" :size="0" />
      </div>
      <div class="speaker back-right">
        <svg-icon icon-class="speaker-back-lr" :size="0" />
      </div>
    </div>
  </div>
</template>

<script>
import { getLineInfo } from '@/api/line';
import { socket } from '@/common/request';
export default {
  name: 'Speaker',
  line: { type: Object, required: true },
  data() {
    return {
      channels: 0,
      channelLayout: 'none',
    };
  },
  mounted() {
    socket.onConnected(() => this.loadData());
  },

  methods: {
    loadData() {
      if (!this.line) {
        this.$router.push('/speakers');
        return;
      }
      getLineInfo(this.line.id).then(data => {
        this.line = data;
      });
    },
  },
};
</script>

<style lang="scss" scoped>
.container {
  position: relative;
  display: flex;
  justify-content: center;
  align-items: center;

  .room {
    width: 480px;
    height: 320px;
    position: relative;

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
  .channels {
    .speaker {
      position: absolute;

      .svg-icon {
        width: 4rem;
        height: auto;
      }
    }

    .front-left {
      top: 31%;
      left: 40%;
    }
    .front-right {
      top: 31%;
      right: 40%;
    }
    .front-center {
      top: 31%;
      left: calc(50% - 1.5rem);
      .svg-icon {
        width: 3rem;
      }
    }
    .front-bass {
      top: 35%;
      left: 35%;
    }

    .side-left {
      top: 44%;
      left: 27%;
    }
    .side-right {
      top: 44%;
      right: 27%;
    }

    .back-left {
      top: 65%;
      left: 36%;
    }
    .back-right {
      top: 65%;
      right: 36%;
    }
  }
}
</style>
