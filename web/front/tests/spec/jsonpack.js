import { describe, it } from 'mocha';
const assert = require("assert")
import Mock from 'mockjs'

import {encode, decode} from '../../src/common/jsonpack';

let stringEncoder = new TextEncoder()

let json = Mock.mock({
  'data|100': [{
    int: '@integer(0, 99999999)',
    str: '@ctitle',
    "bool|1-2": true,
  }]
})
console.log(json)

describe("jsonpack", () => {
  it("长度测试", function() {
    assert(encode(json).length < stringEncoder.encode(JSON.stringify(json)).length)
  })
  it('稳定性测试', function() {
    assert.deepEqual(json, decode(encode(json)))
  });
  it('encode性能测试', function() {
    let count = 1000

    while(count -- > 0) {
      encode(json)
    }
    assert(true)
  });

  let packed = encode(json)

  it('decode性能测试', function() {
    let count = 1000
    while(count -- > 0) {
      decode(packed)
    }
    assert(true)
  });
})