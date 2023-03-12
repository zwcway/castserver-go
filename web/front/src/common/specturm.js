
let SP, ctx;
let spdata = [];
let slow = [];
let title = [];
let spRequestId;
let speakerIndex = 0;
let level = new VolumeLevel('200ms');
let playerDurationTimeout;

export function start(sp) {
    SP
}
function drawSpectrum() {
    ctx.clearRect(0, 0, SP.width, SP.height);
    const spc = Math.max(16, spdata.length);
    let w = SP.width / spc;
    let space = w > 1 ? w * 0.1 : 0;
    let left = 0;
    let spd = 0;
    w -= space;
    for (var i = 0; i < spc; i++) {
        left = i * (w + space);
        spd = spdata[i]
        if (spd >= SP.height) {
            spd = SP.height - 1;
        }

        ctx.beginPath();
        ctx.lineWidth = w;
        ctx.strokeStyle = 'hsl(171deg 100% 41% / 20%)';

        if (slow[i] > spd) {
            slow[i] -= 2;
        } else if (spd - slow[i] > 8) {
            slow[i] += 5;
        } else {
            slow[i] = spd;
        }

        ctx.moveTo(left, SP.height);
        ctx.lineTo(left, SP.height - slow[i]);
        ctx.stroke();

        if (title[i] > slow[i]) {
            title[i] -= 0.5;
        } else {
            title[i] = slow[i];
        }

        ctx.beginPath();
        ctx.lineWidth = w;
        ctx.strokeStyle = '#bbb';
        ctx.moveTo(left, SP.height - title[i]);
        ctx.lineTo(left, SP.height - title[i] + 1);
        ctx.stroke();
    }

    if (speakerIndex > level.length) {
        speakerIndex = 0;
    }

    if (level.length) {
        level.commitWidth(speakerIndex);
        speakerIndex++;
    }

    spRequestId = requestAnimationFrame(drawSpectrum);
}