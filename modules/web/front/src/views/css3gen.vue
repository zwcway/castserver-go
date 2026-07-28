<template>
    <div class="page">
        <div class="container">
            <div :style="style()" class="cube">
                <div class="cube__face cube__face--front">front</div>
                <div class="cube__face cube__face--back">back</div>
                <div class="cube__face cube__face--right">right</div>
                <div class="cube__face cube__face--left">left</div>
                <div class="cube__face cube__face--top">top</div>
                <div class="cube__face cube__face--bottom">bottom</div>
            </div>
        </div>
        scaleX<a-slider v-model="scaleX" :min="-10" :max="10" :step="0.1" :tooltip-visible="true" />
        b1<a-slider v-model="b1" :min="-10" :max="10" :step="0.1" :tooltip-visible="true" />
        c1<a-slider v-model="c1" :min="-10" :max="10" :step="0.1" :tooltip-visible="true" />
        d1<a-slider v-model="d1" :min="-1" :max="1" :step="0.001" :tooltip-visible="true" />
        skewX<a-slider v-model="skewX" :min="-10" :max="10" :step="0.1" :tooltip-visible="true" />
        scaleY<a-slider v-model="scaleY" :min="-10" :max="10" :step="0.1" :tooltip-visible="true" />
        c2<a-slider v-model="c2" :min="-10" :max="10" :step="0.1" :tooltip-visible="true" />
        a<a-slider v-model="d2" :min="-1" :max="1" :step="0.001" :tooltip-visible="true" />
        a<a-slider v-model="a3" :min="-10" :max="10" :step="0.1" :tooltip-visible="true" />
        a<a-slider v-model="b3" :min="-10" :max="10" :step="0.1" :tooltip-visible="true" />
        a<a-slider v-model="c3" :min="-10" :max="10" :step="0.1" :tooltip-visible="true" />
        a<a-slider v-model="d3" :min="-10" :max="10" :step="0.1" :tooltip-visible="true" />
        a<a-slider v-model="a4" :min="-10" :max="10" :step="0.1" :tooltip-visible="true" />
        a<a-slider v-model="b4" :min="-10" :max="10" :step="0.1" :tooltip-visible="true" />
        a <a-slider v-model="c4" :min="-10" :max="10" :step="0.1" :tooltip-visible="true" />
        scaleAll<a-slider v-model="scaleAll" :min="-10" :max="10" :step="0.1" :tooltip-visible="true" />
    </div>
</template>

<script>
export default {
    data() {
        return {
            scaleX: 1,
            b1: 0,
            c1: 0,
            d1: 0,
            skewX: 0,
            scaleY: 1,
            c2: 0,
            d2: 0,
            a3: 0,
            b3: 0,
            c3: 1,
            d3: 0,
            a4: 0,
            b4: 0,
            c4: 0,
            scaleAll: 1
        }
    },
    methods: {
        style() {
            return `transform: matrix3d(${this.scaleX}, ${this.b1}, ${this.c1}, ${this.d1}, ${this.skewX}, ${this.scaleY}, ${this.c2}, ${this.d2}, ${this.a3}, ${this.b3}, ${this.c3}, ${this.d3}, ${this.a4}, ${this.b4}, ${this.c4}, ${this.scaleAll})`
        }
    }
}
</script>

<style lang="scss" scoped>
.page {
    width: 100%;
    height: 100%;
}

.container {
    display: flex;
    align-items: center;
    justify-items: center;
    width: 480px;
    height: 320px;
    margin: 0 auto;
    
  perspective: 600px;

    .cube {
        width: 480px;
        height: 320px;

        position: relative;
        // border: 1px solid #000;
        // background-color: blue;
        transform-style: preserve-3d;

        .cube__face {
            position: absolute;
            width: 480px;
            height: 320px;
            text-align: center;
            line-height: 320px;
        }

        .cube__face--front {
            width: 480px;
            height: 320px;
            transform: rotateY(0deg) translateZ(160px);
            background-color: rgba(255, 0, 0, 0.5);
        }

        .cube__face--right {
            width: 320px;
            height: 320px;

            transform: rotateY(90deg) translateZ(160px);
            background-color: rgba(255, 255, 0, 0.5);
     }

        .cube__face--back {
            width: 480px;
            height: 320px;
            transform: rotateY(180deg) translateZ(160px);
            background-color: rgba(153, 255, 0, 0.5);
    }

        .cube__face--left {
            width: 320px;
            height: 320px;
            transform: rotateY(-90deg) translateZ(160px);
            background-color: rgba(127, 255, 212, 0.5);
       }

        .cube__face--top {
            width: 480px;
            height: 320px;
            transform: rotateX(90deg) translateZ(160px);
            background-color: rgba(0, 0, 255, 0.5);
      }

        .cube__face--bottom {
            width: 480px;
            height: 320px;
           transform: rotateX(-90deg) translateZ(160px);
            background-color: rgba(218, 112, 214, 0.5);
     }
    }
}</style>