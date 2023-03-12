function mask2arr(mask) {
  let arr = [];
  if (typeof mask === 'number') {
    for (let i = 0; i < 64; i++) {
      if ((mask & 0x01) > 0) arr.push(i);

      mask >>= 1;
    }
  }
  return arr;
}
function formatRateMask(rateMask) {
  return mask2arr(rateMask).map(r => {
    return formatRate(r);
  });
}
function formatBitMask(bitMask) {
  return mask2arr(bitMask).map(r => {
    return r + '';
  });
}
function formatMAC(mac) {
  if (mac instanceof Array) {
    let arr = mac.map(m => {
      return '' + m;
    });
    return arr.join(':');
  }
  return '';
}
function formatIP(ip) {
  if (ip instanceof Array) {
    let arr = ip.map(m => {
      return '' + m;
    });
    return arr.join('.');
  }
  return '';
}

const ipv6Regex = /(([0-9a-fA-F]{1,4}:){7,7}[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,7}:|([0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,5}(:[0-9a-fA-F]{1,4}){1,2}|([0-9a-fA-F]{1,4}:){1,4}(:[0-9a-fA-F]{1,4}){1,3}|([0-9a-fA-F]{1,4}:){1,3}(:[0-9a-fA-F]{1,4}){1,4}|([0-9a-fA-F]{1,4}:){1,2}(:[0-9a-fA-F]{1,4}){1,5}|[0-9a-fA-F]{1,4}:((:[0-9a-fA-F]{1,4}){1,6})|:((:[0-9a-fA-F]{1,4}){1,7}|:)|fe80:(:[0-9a-fA-F]{0,4}){0,4}%[0-9a-zA-Z]{1,}|::(ffff(:0{1,4}){0,1}:){0,1}((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])|([0-9a-fA-F]{1,4}:){1,4}:((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9]))/gi;
const ipv4Regex = /^(25[0-5]|2[0-4][0-9]|1?[0-9][0-9]{1,2})(\.(25[0-5]|2[0-4][0-9]|1?[0-9]{1,2})){3}$/;
const numRegex = /^\d+$/;

export function isIPv4(ip) {
  if (typeof ip !== 'string')
    return false;
  const ips = ip.split('.')
    .map(s => {
      if (!numRegex.test(s))
        return -1;
      const i = parseInt(s)
      if (i > 255 || i < 0) // 允许尾部255或0的非广播地址
        return -1;
      return i;
    }).filter(i => {
      return i > 0;
    })
  if (ips.length !== 4) {
    return false;
  }
  if (ips[0] === 0 || ips[0] === 255) {
    return false;
  }
  return true;
}

const hexRegex = /^[0-9a-f]+$/i;
export function isIPv6(ip) {
  if (typeof ip !== 'string')
    return false;
  let sample = 0;
  const ips = ip.split(':')
    .map(s => {
      if (s.length === 0) {
        sample++; // 简写只允许出现一次
        return 0;
      }
      if (!hexRegex.test(s)) {
        return -1;
      }
      const i = parseInt(s, 16);
      return i;
    })

  if (ips.length < 3 || ips.length > 8 || sample > 1) {
    return false
  }

  for (let i = 0; i < ips.length; i++) {
    if (i > (1 << 15) - 1 || i < 0) {
      return false;
    }
  }

  return true;
}

export function isIP46(ip) {
  if (typeof ip !== 'string' || ip.length == 0) {
    return false;
  }
  return isIPv4(ip) || isIPv6(ip);
}

const domainRegex = /^[a-zA-Z0-9][a-zA-Z0-9-]{1,61}[a-zA-Z0-9]\.[a-zA-Z]{2,}$/i;
export function isHost(host) {
  if (typeof host !== 'string' || host.length == 0) {
    return false;
  }
  if (isIP46(host))
    return true;
  return domainRegex.test(host);
}

export function isPort(port) {
  if (typeof port === 'string') {
    if (port.length == 0)
      return false;
    if (!numRegex.test(port))
      return false;
    port = parseInt(port);
  }
  if (typeof port !== 'number')
    return false;

  return port > 0 && port < 65535;
}

export function formatSpeaker(speaker) {
  if (speaker.mac && typeof speaker.mac !== 'string')
    speaker.mac = formatMAC(speaker.mac);
  else if (speaker.mac === undefined) speaker.mac = '00:00:00:00:00:00';

  if (speaker.ip && typeof speaker.ip !== 'string')
    speaker.ip = formatIP(speaker.ip);

  if (!speaker.rateList) {
    if (speaker.rateMask && typeof speaker.rateMask === 'number')
      speaker.rateList = formatRateMask(speaker.rateMask);
    else speaker.rateList = [];
  }
  if (!speaker.bitList) {
    if (speaker.bitsMask && typeof speaker.bitsMask === 'number')
      speaker.bitList = formatBitMask(speaker.bitsMask);
    else speaker.bitList = [];
  }

  return speaker;
}

export function formatRate(rate) {
  if (typeof rate !== 'number') {
    rate = parseInt(rate);
  }

  return rate / 1000 + 'Hz';
}

export function formatBits(bit) {
  if (typeof bit === 'string') {
    return bit;
  }
  bit = parseInt(bit);
  return bit + '';
}

export function formatLayout(channels) {
  switch (channels) {
    case 1:
      return 'mono';
    case 2:
      return 'stereo';
    case 3:
      return '2.1';
    case 5:
      return '5.0';
    case 6:
      return '5.1';
    case 7:
      return '7.0';
    case 8:
      return '7.1';
    default:
      return channels === undefined ? 'N' : channels + '';
  }
}

export function formatDuration(seconds) {
  seconds = parseInt(seconds)
  if (!(seconds >= 0)) 
    return '00:00:00'

  const hour = (seconds / 3600).toFixed(0).padStart(2, '0')
  const min = (seconds / 60).toFixed(0).padStart(2, '0')
  const sec = (seconds % 60).toFixed(0).padStart(2, '0')
  return hour + ':' + min + ':' + sec
}