import mock from 'mockjs';
import Socket from '@/common/ws';
import { Command, Event } from '@/common/request';

function generateTPL(id, tpl) {
  let template = {};
  if (id) template['data|' + id] = [tpl];
  else template['data|0-10'] = [tpl];

  return template;
}
let funcTpl = (tpl, id) => {
  return () => {
    return mock.mock(generateTPL(id, tpl)).data;
  };
};
let nomarlTpl = (tpl, id) => {
  return mock.mock(generateTPL(id, tpl)).data;
};

Socket.addBeforeSend('speakerList', () => {
  return nomarlTpl({
    'id|+1': 1,
    name: '@ctitle',
    ip: '@ip',
    mac: '@mac',
    ch: '@integer(0,8)',
    rate: '@integer(8000, 384000)',
    bits: '@integer(8, 32)',
    vol: '@integer(0, 99)',
    mute: '@boolean',
    line: {
      id: "@integer(0,10)",
      name: "@ctitle",
    },
    'rateList|0-5': [44100, 48000, 96000],
    'bitsList|0-5': [8, 16, 24, 32],

  });
});
Socket.addBeforeSend('wsconnect', () => {
  return {
    readyState: WebSocket.OPEN,
    send: () => {},
    close: () => {},
  };
});
Socket.addBeforeSend('subscribe', () => {
  return;
});
Socket.addBeforeSend('speakerInfo', params => {
  return mock.mock({
    id: params.id,
    name: '@ctitle',
    ch: '@integer(0,10)',
    ip: '@ip',
    mac: '@mac',
    rate: '@integer(8000, 384000)',
    bits: '@integer(8, 32)',
    vol: '@integer(0, 99)',
    mute: '@boolean',
    line: {
      id: "@integer(0,10)",
      name: "@ctitle",
    },
    'rateList|0-5': [44100, 48000, 96000],
    'bitsList|0-5': [8, 16, 24, 32],
  });
});
Socket.addBeforeSend('speakerVolume', params => {});
Socket.addBeforeSend('setChannel', params => {});
Socket.addBeforeSend('setSpeaker', params => {});
Socket.addBeforeSend('lineVolume', params => {});
Socket.addBeforeSend('setLineEQ', params => {});
Socket.addBeforeSend('clearLineEQ', params => {});

Socket.addBeforeSend('lineList', function () {
  return nomarlTpl(
    {
      'id|+1': 0,
      name: '@ctitle',
      vol: '@integer(0,100)',
      mute: '@boolean',
    },
    '0-20'
  );
});
Socket.addBeforeSend('lineInfo', function (params) {
  return mock.mock({
    id: params.id,
    name: '@ctitle',
    vol: '@integer(0,100)',
    mute: false,
    source: {
      rate: '@integer(44100, 384000)',
      bits: '@integer(8,64)',
      channels: '@integer(1, 16)',
      type: '@integer(1,4)',
      cur:'@integer(0,100)',
      dur:'@integer(1,1000)',
    },
    eq: [['@integer(20,20000)','@float(-15,15)', '@float(0,1)']],
    'speakers|0-20': [
      {
        'id|+1': 1,
        name: '@ctitle',
        ch: '@integer(1,11)',
        vol: '@integer(0, 100)',
        mute: '@boolean',
        ip: '@ip',
        mac: '@mac',
        rate: '@integer(8000, 384000)',
        bits: '@integer(8, 32)',
      },
    ],
  });
});

function splevelmeter(receiver) {
  let data = [];
  if (
    receiver[Event.SP_LevelMeter] instanceof Function
  ) {
    for(let i = 0; i < 10; i ++) {
      data.push([i, Math.random() * 100]);
    }
    receiver[ Event.SP_LevelMeter].call(
      undefined,
      data,
      Command.Speaker,
      Event.SP_LevelMeter
    );
  }
}
function linespectrum(receiver) {
  let data = [];
  let found = false;
  for (let k in receiver) {
    if (k.startsWith(Event.Line_Spectrum + '-')) {
      found = receiver[k];
      break;
    }
  }
  if (!found) {
    return;
  }
  for (let i = 0; i < 48; i++) {
    data[i] = Math.random() * 128;
  }
  found.call(undefined, data, Command.Line, Event.Line_Spectrum);
}
setInterval(() => {
  let receiver = Socket.getReceiver();
  splevelmeter(receiver);
  linespectrum(receiver);
}, 100);
