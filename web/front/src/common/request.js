import router from '@/router';
import socket from '@/common/ws';
import store from '@/store';

if (process.env.NODE_ENV !== 'production' && process.env.Mock) {
  require('../mock');
}
console.log(process.env.Mock)
const Command = Object.freeze({
  Server: 1,
  Speaker: 2,
  Line: 3,
});
const Event = Object.freeze({
  SP_Detected: 0x01,
  SP_Online: 0x02,
  SP_Offline: 0x03,
  SP_Deleted: 0x04,
  SP_Moved: 0x05,
  SP_Edited: 0x06,
  SP_LevelMeter: 7,
  Line_Created: 8,
	Line_Deleted:9,
	Line_Edited:10,
	Line_LevelMeter:11,
	Line_Spectrum:12,
});

export { socket, Command, Event};
