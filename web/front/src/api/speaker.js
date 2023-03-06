import { Command, Event, socket } from '@/common/request';
import { formatSpeaker } from '@/common/format';

export function getSpeakerList() {
  return socket.send('speakerList', {}).then(data => {
    return data.map(sp => {
      return formatSpeaker(sp);
    });
  });
}
let evts = [
  Event.SP_Deleted,
  Event.SP_Deleted,
  Event.SP_Edited,
  Event.SP_Online,
  Event.SP_Offline,
];

export function removeListenSpeakerEvent() {
  socket.sendSubscribe(Command.Speaker, false, evts);
  socket.removeEvent(Command.Speaker, evts);
}

export function listenSpeakerChanged(callback) {
  if (!(callback instanceof Function)) return;

  socket.sendSubscribe(Command.Speaker, true, evts);

  return socket.receiveCommand(Command.Speaker, callback, evts);
}
export function removeListenSpeakerLevelMeter() {
  socket.sendSubscribe(Command.Speaker, false, Event.SP_LevelMeter);
  socket.removeEvent(Command.Speaker, Event.SP_LevelMeter);
}

export function listenSpeakerLevelMeter(callback) {
  if (!(callback instanceof Function)) return;

  socket.sendSubscribe(Command.Speaker, true, Event.SP_LevelMeter);

  return socket.receiveEvent(Event.SP_LevelMeter, callback);
}
export function getSpeakerInfo(id) {
  return socket.send('speakerInfo', { id }).then(speaker => {
    return formatSpeaker(speaker);
  });
}
export function getSpeakerInfos(ids) {
  return socket.send('speakerInfos', { ids }).then(data => {
    return data.map(sp => {
      return formatSpeaker(sp);
    });
  });
}

export function getSpeakersVolumeLevel(ids) {
  return socket.send('volumeLevel', { ids });
}

export function setSpeaker(id, key, val) {
  let data = {};
  if (typeof key === 'object') {
    Object.assign(data, key);
  } else if (typeof key === 'string') {
    data[key] = val;
  }
  data['id'] = parseInt(id);

  return socket.send('setSpeaker', data);
}
export function setVolume(id, vol) {
  let data = { id };
  if (typeof vol === 'boolean') data['mute'] = vol;
  else data['vol'] = vol;

  return socket.send('speakerVolume', data);
}

export function sendServerInfo(id) {
  return socket.send('sendServerInfo', id);
}
export function reconnect(id) {
  return socket.send('spReconnect', id);
}
