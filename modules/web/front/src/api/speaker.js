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
  Event.SP_Detected,
  Event.SP_Deleted,
  Event.SP_Edited,
  Event.SP_Online,
  Event.SP_Offline,
];

export function removeListenSpeakerEvent(ids) {
  socket.removeEvent(evts, ids);
}

export function listenSpeakerChanged(ids, callback) {
  if (!(callback instanceof Function)) return;

  return socket.receiveEvent(evts, ids, callback);
}
export function removeListenSpeakerSpectrum(ids) {
  socket.removeEvent(Event.SP_Spectrum, ids);
}

export function listenSpeakerSpectrum(ids, callback) {
  if (!(callback instanceof Function)) return;
  return socket.receiveEvent(Event.SP_Spectrum, ids, callback);
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

export function test(sp) {
  return socket.send('soundTest', { sp });
}
