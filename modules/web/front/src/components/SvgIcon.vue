<template>
  <svg
    :class="svgClass"
    aria-hidden="true"
    :style="{
      width: size > 0 ? `${size}px` : false,
      height: size > 0 ? `${size}px` : false,
    }"
  >
    <use :xlink:href="iconName" />
  </svg>
</template>

<script>
export default {
  name: 'SvgIcon',
  props: {
    iconClass: {
      type: String,
      required: true,
    },
    className: {
      type: String,
      default: '',
    },
    size: {
      type: Number,
      default: 32,
    },
  },
  computed: {
    iconName() {
      return `#icon-${this.iconClass}`;
    },
    svgClass() {
      if (this.className) {
        return 'svg-icon ' + this.className;
      } else {
        return 'svg-icon';
      }
    },
  },
};
</script>

<style scoped>
.svg-icon {
  fill: currentColor;
}
</style>
