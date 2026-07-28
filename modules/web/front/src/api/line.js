import { Command, Event, socket } from '@/common/request';

export function getLineList() {
  return socket.send('lineList', {});
}
export function getLineInfo(id) {
  return socket.send('lineInfo', { id });
}
export function getLinePlayer(id) {
  return socket.send('linePlayer', { id }, { log: false });
}

export function setVolume(id, vol) {
  let data = { id };
  if (typeof vol === 'boolean') data['mute'] = vol;
  else data['vol'] = vol;

  return socket.send('lineVolume', data);
}

export function setEqualizer(id, seg, freq, gain) {
  return socket.send('setLineEQ', { id, seg, freq, gain });
}
export function setEnableEqualizer(id, enable) {
  return socket.send('enableLineEQ', { id, enable });
}

export function clearEqualizer(id, seg) {
  return socket.send('clearLineEQ', { id, seg });
}

export function removelistenAll() {
  socket.removeEvent([]);
}

let listEvts = [Event.Line_Created, Event.Line_Deleted, Event.Line_Edited];
let evts = [Event.Line_Created, Event.Line_Deleted, Event.Line_Edited, Event.Line_Input];

export function listenLineListChanged(callback) {
  return socket.receiveEvent(listEvts, callback);
}
export function removelistenLineListChanged() {
  socket.removeEvent(listEvts);
}

export function listenLineChanged(id, callback) {
  return socket.receiveEvent(evts, id, callback);
}

export function removelistenLineChanged(id) {
  socket.removeEvent(evts, id);
}

export function listenLineSpeakerChanged(id, callback) {
  return socket.receiveEvent(Event.Line_Speaker, id, callback);
}

export function removeListenLineSpeakerChanged(id) {
  socket.removeEvent(Event.Line_Speaker, id);
}

export function listenLineSpectrum(id, callback) {
  return socket.receiveEvent(Event.Line_Spectrum, id, callback);
}

export function removeListenLineSpectrum(id) {
  socket.removeEvent(Event.Line_Spectrum, id);
}

export function listenLineInput(id, callback) {
  return socket.receiveEvent(Event.Line_Input, id, callback);
}

export function removeListenLineInput(id) {
  socket.removeEvent(Event.Line_Input, id);
}

export function playerSeek(id, pos) {
  return socket.send('lineSeek', { id, pos });
}

export function createLine(name) {
  return socket.send('createLine', { name });
}
export function deleteLine(id, moveTo) {
  id = parseInt(id);
  moveTo = parseInt(moveTo);
  return socket.send('deleteLine', { id, moveTo });
}
export function setLine(id, key, val) {
  let data = {}
  if (typeof key === 'object') {
    Object.assign(data, key);
  } else if (typeof key === 'string') {
    data[key] = val;
  }
  data['id'] = parseInt(id);

  return socket.send('setLine', data);

}

export function testChannel(line, ch) {
  return socket.send('soundTest', { line, ch })
}

export var channelList = {
  1: {
    id: 'front-left',
    name: '左声道',
    icon: 'speaker-front-lr',
    show: false,
  },
  2: {
    id: 'front-right',
    name: '右声道',
    icon: 'speaker-front-lr',
    show: false,
  },
  3: {
    id: 'front-center',
    name: '中置声道',
    icon: 'speaker-front-center',
    show: false,
  },
  4: {
    id: 'front-left-center',
    name: '左中置声道',
    icon: 'speaker-front-center',
    show: false,
  },
  5: {
    id: 'front-right-center',
    name: '右中置声道',
    icon: 'speaker-front-center',
    show: false,
  },
  6: {
    id: 'front-bass',
    name: '重低音声道',
    icon: 'speaker-low-frequency',
    show: false,
  },
  7: {
    id: 'back-left',
    name: '后环绕左声道',
    icon: 'speaker-back-lr',
    show: false,
  },
  8: {
    id: 'back-right',
    name: '后环绕右声道',
    icon: 'speaker-back-lr',
    show: false,
  },
  9: {
    id: 'back-center',
    name: '后环绕中置声道',
    icon: 'speaker-front-center',
    show: false,
  },
  10: {
    id: 'side-left',
    name: '侧环绕左声道',
    icon: 'speaker-side-lr',
    show: false,
  },
  11: {
    id: 'side-right',
    name: '侧环绕右声道',
    icon: 'speaker-side-lr',
    show: false,
  },
  12: {
    id: 'top-front-left',
    name: '天空前左声道',
    icon: 'speaker-back-lr',
    show: false,
  },
  13: {
    id: 'top-front-center',
    name: '天空前中置声道',
    icon: 'speaker-back-lr',
    show: false,
  },
  14: {
    id: 'top-front-right',
    name: '天空前右声道',
    icon: 'speaker-back-lr',
    show: false,
  },
  15: {
    id: 'top-back-left',
    name: '天空后左声道',
    icon: 'speaker-back-lr',
    show: false,
  },
  16: {
    id: 'top-back-center',
    name: '天空后中置声道',
    icon: 'speaker-back-lr',
    show: false,
  },
  17: {
    id: 'top-back-right',
    name: '天空后右声道',
    icon: 'speaker-back-lr',
    show: false,
  },
};
