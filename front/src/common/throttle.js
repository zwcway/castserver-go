/**
 * 限制函数调用频率
 * @param fn 想要调用的函数
 * @param delay 频率 ms
 * @returns {(function(*): void)|*}
 */
export function throttleFunction(fn, delay) {
  let timer,
    ret = undefined;
  return function (v) {
    let _this = this;
    let args = arguments;
    if (timer) {
      return ret;
    }
    timer = setTimeout(function () {
      ret = fn.apply(_this, args);
      timer = null;
    }, delay);
  };
}
