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
    id: '@integer(0, 2)',
    name: '@ctitle',
    ip: '@ip',
    mac: '@mac',
    rate: '@integer(8000, 384000)',
    bits: '@integer(8, 32)',
    volume: '@integer(0, 99)',
  });
});
Socket.addBeforeSend('wsconnect', () => {
  return {
    readyState: WebSocket.OPEN,
    send: () => { },
    close: () => { },
  }
})
Socket.addBeforeSend("subscribe", () => {
  return
})
Socket.addBeforeSend('speakerInfo', (params) => {
  return mock.mock({
    id: params.id,
    name: '@ctitle',
    channel: "@integer(0,10)",
    ip: '@ip',
    mac: '@mac',
    rate: '@integer(8000, 384000)',
    bits: '@integer(8, 32)',
    volume: '@integer(0, 99)',
    'rateList|0-5': [44100, 48000, 96000],
    'bitsList|0-5': [8, 16, 24, 32],
  });
});
Socket.addBeforeSend('speakerVolume', params => {})
Socket.addBeforeSend('setChannel', params => {})
Socket.addBeforeSend('lineVolume', params => {})
Socket.addBeforeSend('setLineEQ', params => {})
Socket.addBeforeSend('clearLineEQ', params => {})

Socket.addBeforeSend('lineList', function () {
  return nomarlTpl(
    {
      id: '@integer(0, 999999)',
      name: '@ctitle',
    },
    '1-20'
  );
});
Socket.addBeforeSend('lineInfo', function (params) {
  return mock.mock({
    id: params.id,
    name: '@ctitle',
    'speakers|0-20': [{
      id: "@integer(1,1000)",
      name: '@ctitle',
      channel: "@integer(1,11)",
      volume: "@integer(0, 100)",
      ip: '@ip',
      mac: '@mac',
      rate: '@integer(8000, 384000)',
      bits: '@integer(8, 32)',
    }]
  });
});

function splevelmeter(receiver) {
  let data = [1, Math.random() * 100]
  if (receiver[Command.Speaker + '.' + Event.SP_LevelMeter] instanceof Function) {
    receiver[Command.Speaker + '.' + Event.SP_LevelMeter].call(undefined, data, Command.Speaker, Event.SP_LevelMeter);
  }
}
function linespectrum(receiver) {
  let data = []
  if (!(receiver[Event.Line_Spectrum + '-' + 1] instanceof Function)) {
    return
  }
  for (let i = 0; i < 48; i++) {
    data[i] = Math.random() * 128
  }
  receiver[Event.Line_Spectrum + '-' + 1].call(undefined, data, Command.Line, Event.Line_Spectrum);
}
setInterval(() => {
  let receiver = Socket.getReceiver()
  splevelmeter(receiver)
  linespectrum(receiver)
}, 1000)