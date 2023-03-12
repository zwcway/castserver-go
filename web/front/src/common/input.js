var chineseRegex = /^[\u4e00-\u9fa5]{0,}$/; //定义正则表达式

export function inputLengthLimit(obj, num) {
  var val = obj.value; //获取输入的value值
  var len = 0;
  for (var i = 0; i < val.length; i++) {
    if (chineseRegex.test(val[i])) {
      //如果是中文，则让len+2
      len += 2;
    } else {
      len++; //反之英文len+1
    }
  }
  if (len > num) {
    //如果英文和汉字加起来长度超过设定的字符数，截取字数，并提示用户
    obj.value = val.substring(0, num);
  }
}
