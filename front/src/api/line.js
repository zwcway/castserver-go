import { socket } from '@/common/request';

export function getLineList() {
  return socket
    .send('lineList', {})
    .then(data => {
      return data;
    })
    .catch((err, code, res) => {});
}
export function getLineInfo(id) {
  return socket
    .send('lineInfo', { id })
    .then(data => {
      return data;
    })
    .catch((err, code, res) => {});
}
export function listenLineChanged(callback) {
  return socket.receive('lineEvent', callback);
}
