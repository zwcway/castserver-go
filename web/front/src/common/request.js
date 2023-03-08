import socket from '@/common/ws';

if (process.env.NODE_ENV !== 'production') {
  window.socket = socket
  if ( process.env.Mock) {
    require('../mock');
  }
}
console.log(process.env.Mock);
const Command = Object.freeze({
  Server: 1,
  Speaker: 2,
  Line: 3,
});
const Event = Object.freeze({
  SP_Detected: 11,
  SP_Online: 12,
  SP_Offline: 13,
  SP_Deleted: 14,
  SP_Moved: 15,
  SP_Edited: 16,
  SP_LevelMeter: 17,
  Line_Created: 18,
  Line_Deleted: 19,
  Line_Edited: 20,
  Line_Speaker: 21,
  Line_LevelMeter: 22,
  Line_Spectrum: 23,
  Line_Input: 24,
});

export { socket, Command, Event };
