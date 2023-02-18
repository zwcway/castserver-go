import { socket } from '@/common/request';
import { formatSpeaker } from '@/common/format';

window.socket = socket;

export function getSpeakerList() {
  return socket.send('speakerList', {}).then(data => {
    return data.map(sp => {
      return formatSpeaker(sp);
    });
  });
}
export function removeListenSpeakerEvent() {
  socket.send('subscribe', ['speakerEvt', false], { noResponse: true });
}

export function listenSpeakerChanged(callback) {
  if (!callback instanceof Function) return;

  socket.send('subscribe', ['speakerEvt', true], { noResponse: true });

  return socket.receive('speakerEvt', (act, speaker) => {
    callback(act, formatSpeaker(speaker));
  });
}

export function listenSpeakerLevelMeter(callback) {
  if (!callback instanceof Function) return;

  socket.send('subscribe', ['spLvMeteEvt', true], { noResponse: true });

  return socket.receive('spLvMeteEvt', callback);
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
    .catch((err, code, res) => {});
}

export function setVolume(id, vol) {
  return socket.send('volume', { id, vol }, { noResponse: true });
}

export function sendServerInfo(id) {
  return socket.sendString('sendServerInfo', id);
}
export function reconnect(id) {
  return socket.sendString('spReconnect', id);
}
