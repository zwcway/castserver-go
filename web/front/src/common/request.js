import socket from '@/common/ws';

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
  SP_Spectrum: 18,
  Line_Created: 19,
  Line_Deleted: 20,
  Line_Edited: 21,
  Line_Speaker: 22,
  Line_Spectrum: 23,
  Line_LevelMeter: 24,
  Line_Input: 25,
});

export { socket, Command, Event };
