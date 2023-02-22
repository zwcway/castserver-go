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
  Event.SP_Deleted, Event.SP_Deleted, Event.SP_Edited, Event.SP_Online, Event.SP_Offline
];

export function removeListenSpeakerEvent() {
  socket.sendSubscribe(Command.Speaker, false, evts);
  socket.removeEvent(Command.Speaker, evts)
}

export function listenSpeakerChanged(callback) {
  if (!(callback instanceof Function)) return;

  socket.sendSubscribe(Command.Speaker, true, evts);

  return socket.receiveEvent(Command.Speaker, (act, speaker) => {
    callback(act, formatSpeaker(speaker));
  }, evts);
}
export function removeListenSpeakerLevelMeter() {
  socket.sendSubscribe(Command.Speaker, false, Event.SP_LevelMeter);
  socket.removeEvent(Command.Speaker, Event.SP_LevelMeter)
}

export function listenSpeakerLevelMeter(callback) {
  if (!(callback instanceof Function)) return;

  socket.sendSubscribe(Command.Speaker, true, Event.SP_LevelMeter);

  return socket.receiveEvent(Command.Speaker, callback, Event.SP_LevelMeter);
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
  return socket
    .send('volumeLevel', { ids })
    .then(data => {
      return data;
    })
    .catch((err, code, res) => { });
}

export function setChannel(id, ch) {
  id = parseInt(id)
  ch = parseInt(ch)
  return socket.send('setChannel', { id, ch })
}
export function setVolume(id, vol) {
  return socket.send('speakerVolume', { id, vol }, { noResponse: true });
}

export function sendServerInfo(id) {
  return socket.send('sendServerInfo', id);
}
export function reconnect(id) {
  return socket.send('spReconnect', id);
}   
