const enumValue = name => Object.freeze({ toString: () => name });
const JsonPackType = Object.freeze({
  INTEGER: 0x01,
  BOOLEAN: 0x02,
  STRING: 0x03,
  ARRAY: 0x04,
  MAP: 0x05,
  FLOAT: 0x06,
});

let stringEncoder = new TextEncoder();
let stringDecoder = new TextDecoder();

let float32 = new DataView(new ArrayBuffer(4));
class PackArray extends Array {
  numberSize(code) {
    if (code < 0xff) {
      return 1;
    } else if (code <= 0xffff) {
      return 2;
    }
    // else if (code <= 0xffffff) {
    //   return 3
    // }
    else if (code <= 0xffffffff) {
      return 4;
    }
    return 0;
  }
  pushString(str) {
    for (var i = 0; i < str.length; i ++) {
      this.push(str.charCodeAt(i))
    }
  }
  pushNumber(code) {
    if (code < 0xff) {
      this.push(code);
      return 1;
    } else if (code <= 0xffff) {
      this.push(code & 0xff);
      this.push(code >> 8);
      return 2;
    }
    // else if (code <= 0xffffff) {
    //   this.push(code & 0xff)
    //   this.push((code >> 8) & 0xff)
    //   this.push((code >> 16) & 0xff)
    //   return 3
    // }
    else if (code <= 0xffffffff) {
      this.push(code & 0xff);
      this.push((code >> 8) & 0xff);
      this.push((code >> 16) & 0xff);
      this.push((code >> 24) & 0xff);
      return 4;
    }
    return 0;
  }

  setFlag(i, flag) {
    this[i] |= flag & 0x0f;
  }
  pushType(t) {
    this.push((t << 4) & 0xf0);
  }
  pushTypeFlag(t, flag) {
    this.push(((t << 4) | (flag & 0x0f)) & 0xff);
  }

  concat(arr) {
    arr.forEach(item => this.push(item));
  }

  encodeInteger(int) {
    let size = this.numberSize(int);
    this.pushTypeFlag(JsonPackType.INTEGER, size);
    this.pushNumber(int);
  }
  encodeFloat32(float) {
    float32.setFloat32(float)
    int = float32.getUint32();
    let size = this.numberSize(int);
    this.pushTypeFlag(JsonPackType.FLOAT, size);
    this.pushNumber(int);
  }
  encodeBoolean(bool) {
    this.pushType(JsonPackType.BOOLEAN);
    this.setFlag(this.length - 1, bool);
  }
  /*
  encodeString(str) {
    str = encodeURIComponent(str)
    let c = ''
    let arr = [], i = 0
    for(i = 0; i < str.length; i++) {
      c = str.charAt(i)
      if (c === '%') {
        arr.push(parseInt(str.charAt(i + 1) + str.charAt(i + 2), 16));
        i += 2;
      } else
        arr.push(c.charCodeAt(0));
    }
    this.pushTypeFlag(JsonPackType.STRING, arr.length > 0xff)
    this.pushLength(arr.length)
    this.concat(arr)
  }
*/
  encodeString(str) {
    let arr = stringEncoder.encode(str);
    this.pushTypeFlag(JsonPackType.STRING, this.numberSize(arr.length));
    this.pushNumber(arr.length);
    this.concat(arr);
  }

  encodeArray(array) {
    this.pushTypeFlag(JsonPackType.ARRAY, this.numberSize(array.length));
    this.pushNumber(array.length);
    for (let i = 0; i < array.length; i++) {
      this.encodeObject(array[i]);
    }
  }

  encodeMap(map) {
    let unpacker = new PackArray();
    let len = 0;
    for (let k in map) {
      if (!map.hasOwnProperty(k)) continue;
      if (k.length > 0xff) {
        throw 'key size too big: ' + k;
      }
      unpacker.encodeString('' + k);
      unpacker.encodeObject(map[k]);
      len++;
    }
    this.pushTypeFlag(JsonPackType.MAP, this.numberSize(len));
    this.pushNumber(len);
    this.concat(unpacker);
  }
  encodeObject(obj) {
    if (typeof obj === 'string') {
      this.encodeString(obj);
    } else if (typeof obj === 'number') {
      if (obj % 1 === 0) {
        this.encodeInteger(obj);
      } else {
        this.encodeFloat32(obj);
      }
    } else if (typeof obj === 'boolean') {
      this.encodeBoolean(obj);
    } else if (obj instanceof Array) {
      this.encodeArray(obj);
    } else if (obj instanceof Object) {
      this.encodeMap(obj);
    }
  }
}

class UnPackArray {
  constructor(arr) {
    this.pos = 0;
    this.array = arr;
  }
  seek(i) {
    this.pos += i;
  }
  at(i) {
    return this.array[i + this.pos];
  }
  r(i) {
    return this.array[i + this.pos++];
  }
  r8() {
    return this.r(0);
  }
  r16() {
    return this.r(0) | (this.r(0) << 8);
  }
  r24() {
    return this.r(0) | (this.r(0) << 8) | (this.r(0) << 16);
  }
  r32() {
    return this.r(0) | (this.r(0) << 8) | (this.r(0) << 16) | (this.r(0) << 24);
  }
  rNumber(flag) {
    switch (flag) {
      case 1:
        return this.r8();
      case 2:
        return this.r16();
      // case 3: return this.r24()
      case 4:
        return this.r32();
    }
    throw 'unknown flag';
  }
  slice(len) {
    return this.array.slice(this.pos, this.pos + len);
  }

  type(i) {
    return i >> 4;
  }
  flag(i) {
    return i & 0x0f;
  }

  decodeInteger(flag) {
    switch (flag) {
      case 1:
        return this.r8();
      case 2:
        return this.r16();
      case 3:
        return this.r24();
      case 4:
        return this.r32();
    }
    throw 'decode failed';
  }
  decodeFloat32(flag) {
    let i = this.decodeInteger(flag)
    float32.setInt32(i)
    return float32.getFloat32()
  }

  decodeBoolean(flag) {
    return flag > 0;
  }

  /*
  decodeString(flag) {
    let len = flag ? this.r16() : this.r8()
    let s = ''
    while (len -- > 0) {
      s += '%' + this.r8().toString(16)
    }
    return decodeURIComponent(s)
  }
*/
  decodeString(flag) {
    let len = this.rNumber(flag);
    let str = stringDecoder.decode(this.slice(len));
    this.seek(len);
    return str;
  }

  decodeArray(flag) {
    let ret = [];
    let len = this.rNumber(flag);

    for (let i = 0; i < len; i++) {
      ret.push(this.decodeObject());
    }
    return ret;
  }
  decodeMap(flag) {
    let ret = {};
    let len = this.rNumber(flag);
    let key = '';
    for (let i = 0; i < len; i++) {
      key = this.decodeObject(JsonPackType.STRING);
      if (!key.length) {
        throw 'decode error. no key in map';
      }
      ret[key] = this.decodeObject();
    }
    return ret;
  }
  decodeObject(mustType = null) {
    let type = this.r8(),
      flag = this.flag(type);
    type = this.type(type);

    if (mustType && mustType !== type) throw 'type not expected';

    switch (type) {
      case JsonPackType.INTEGER:
        return this.decodeInteger(flag);
      case JsonPackType.BOOLEAN:
        return this.decodeBoolean(flag);
      case JsonPackType.STRING:
        return this.decodeString(flag);
      case JsonPackType.ARRAY:
        return this.decodeArray(flag);
      case JsonPackType.MAP:
        return this.decodeMap(flag);
      case JsonPackType.FLOAT:
        return this.decodeFloat32(flag);
      default:
        throw 'decode error. unknown type ' + type;
    }
  }
}

export function encode(obj) {
  let arr = new PackArray();
  arr.encodeObject(obj);
  return new Uint8Array(arr);
}

export function encodeReq(id, cmd, obj) {
  let arr = new PackArray();
  arr.pushString(id);
  arr.push(0);
  arr.pushString(cmd);
  arr.push(0);
  arr.encodeObject(obj);
  return new Uint8Array(arr);
}

export function decode(bytes) {
  if (bytes.length === 0) return {};
  let arr = new UnPackArray(bytes);
  return arr.decodeObject();
}
