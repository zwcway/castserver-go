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
    return r;
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
