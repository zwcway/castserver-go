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
  socket.removeEvent(Command.Speaker, evts);
}

export function listenSpeakerChanged(callback) {
  if (!(callback instanceof Function)) return;

  return socket.receiveCommand(Command.Speaker, evts, callback);
}
export function removeListenSpeakerLevelMeter() {
  socket.removeEvent(Command.Speaker, Event.SP_LevelMeter);
}

export function listenSpeakerLevelMeter(callback) {
  if (!(callback instanceof Function)) return;
  return socket.receiveCommand(Command.Speaker, Event.SP_LevelMeter, callback);
}

export function getSpeakerInfo(id) {
  return socket.send('speakerInfo', id).then(speaker => {
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
