import { Command, Event, socket } from '@/common/request';

export function getLineList() {
  return socket.send('lineList', {});
}
export function getLineInfo(id) {
  return socket.send('lineInfo', { id });
}

export function setVolume(id, vol) {
  let data = { id };
  if (typeof vol === 'boolean') data['mute'] = vol;
  else data['vol'] = vol;

  return socket.send('lineVolume', data, { noResponse: true });
}

export function setEqualizer(id, freq, gain, q) {
  return socket.send('setLineEQ', { id, freq, gain, q }, { noResponse: true });
}
export function clearEqualizer(id) {
  return socket.send('clearLineEQ', { id }, { noResponse: true });
}

let evts = [Event.Line_Created, Event.Line_Deleted, Event.Line_Edited];

export function listenLineChanged(callback) {
  socket.sendSubscribe(Command.Line, true, evts);
  return socket.receiveCommand(Command.Line, callback, evts);
}
export function removelistenLineChanged() {
  socket.sendSubscribe(Command.Line, false, evts);
  socket.removeEvent(Command.Line, evts);
}

export function listenLineSpectrum(id, callback) {
  socket.sendSubscribe(Command.Line, true, Event.Line_Spectrum, id);
  return socket.receiveEvent(Event.Line_Spectrum, callback, id);
}

export function removeListenLineSpectrum(id) {
  socket.sendSubscribe(Command.Line, false, Event.Line_Spectrum, id);
  socket.removeEvent(Command.Line, Event.Line_Spectrum, id);
}

export function createLine(name) {
  return socket.send('createLine', { name });
}
export function deleteLine(id, moveTo) {
  id = parseInt(id);
  moveTo = parseInt(moveTo);
  return socket.send('deleteLine', { id, moveTo });
}
export function setLine(id, key, val)
{
  let data = {}
  if (typeof key === 'object') {
    Object.assign(data, key);
  } else if (typeof key === 'string') {
    data[key] = val;
  }
  data['id'] = parseInt(id);

  return socket.send('setLine', data);

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
  6: {
    id: 'front-bass',
    name: '重低音声道',
    icon: 'speaker-low-frequency',
    show: false,
  },
  7: {
    id: 'side-left',
    name: '侧环绕左声道',
    icon: 'speaker-side-lr',
    show: false,
  },
  8: {
    id: 'side-right',
    name: '侧环绕右声道',
    icon: 'speaker-side-lr',
    show: false,
  },
  10: {
    id: 'back-left',
    name: '后置环绕左声道',
    icon: 'speaker-back-lr',
    show: false,
  },
  11: {
    id: 'back-right',
    name: '后置环绕右声道',
    icon: 'speaker-back-lr',
    show: false,
  },
};
