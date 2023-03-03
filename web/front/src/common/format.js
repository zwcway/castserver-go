function mask2arr(mask) {
  let arr = [];
  for (let i = 0; i < 64; i++) {
    if ((mask & 0x01) > 0) arr.push(i);

    mask >>= 1;
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
    return r;
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

let ipv46Regex =
  /(?:^(?:25[0-5]|2[0-4]\d|1\d\d|[1-9]\d|\d)(?:\.(?:25[0-5]|2[0-4]\d|1\d\d|[1-9]\d|\d)){3}$)|(?:^(?:(?:[a-fA-F\d]{1,4}:){7}(?:[a-fA-F\d]{1,4}|:)|(?:[a-fA-F\d]{1,4}:){6}(?:(?:25[0-5]|2[0-4]\d|1\d\d|[1-9]\d|\d)(?:\\.(?:25[0-5]|2[0-4]\d|1\d\d|[1-9]\d|\d)){3}|:[a-fA-F\d]{1,4}|:)|(?:[a-fA-F\d]{1,4}:){5}(?::(?:25[0-5]|2[0-4]\d|1\d\d|[1-9]\d|\d)(?:\\.(?:25[0-5]|2[0-4]\d|1\d\d|[1-9]\d|\d)){3}|(?::[a-fA-F\d]{1,4}){1,2}|:)|(?:[a-fA-F\d]{1,4}:){4}(?:(?::[a-fA-F\d]{1,4}){0,1}:(?:25[0-5]|2[0-4]\d|1\d\d|[1-9]\d|\d)(?:\\.(?:25[0-5]|2[0-4]\d|1\d\d|[1-9]\d|\d)){3}|(?::[a-fA-F\d]{1,4}){1,3}|:)|(?:[a-fA-F\d]{1,4}:){3}(?:(?::[a-fA-F\d]{1,4}){0,2}:(?:25[0-5]|2[0-4]\d|1\d\d|[1-9]\d|\d)(?:\\.(?:25[0-5]|2[0-4]\d|1\d\d|[1-9]\d|\d)){3}|(?::[a-fA-F\d]{1,4}){1,4}|:)|(?:[a-fA-F\d]{1,4}:){2}(?:(?::[a-fA-F\d]{1,4}){0,3}:(?:25[0-5]|2[0-4]\d|1\d\d|[1-9]\d|\d)(?:\\.(?:25[0-5]|2[0-4]\d|1\d\d|[1-9]\d|\d)){3}|(?::[a-fA-F\d]{1,4}){1,5}|:)|(?:[a-fA-F\d]{1,4}:){1}(?:(?::[a-fA-F\d]{1,4}){0,4}:(?:25[0-5]|2[0-4]\d|1\d\d|[1-9]\d|\d)(?:\\.(?:25[0-5]|2[0-4]\d|1\d\d|[1-9]\d|\d)){3}|(?::[a-fA-F\d]{1,4}){1,6}|:)|(?::(?:(?::[a-fA-F\d]{1,4}){0,5}:(?:25[0-5]|2[0-4]\d|1\d\d|[1-9]\d|\d)(?:\\.(?:25[0-5]|2[0-4]\d|1\d\d|[1-9]\d|\d)){3}|(?::[a-fA-F\d]{1,4}){1,7}|:)))(?:%[0-9a-zA-Z]{1,})?$)/gm;

export function isIP46(ip) {
  if (typeof ip !== 'string' || ip.length == 0) {
    return false;
  }
  return ipv46Regex.test(ip);
}

export function isPort(port) {
  if (typeof port !== 'string' || port.length == 0) {
    return false;
  }
  if (!port.match(/^\d+$/)) {
    return false;
  }
  let i = parseInt(port);
  return i > 0 && i < 65535;
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
  if (typeof bit !== 'number') {
    bit = parseInt(bit);
  }
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
