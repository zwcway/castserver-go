function VolumeLevel(duration) {
  let vals, eles, maps;
  let dur = duration;

  this.clear = () => {
    this.length = 0;
    vals = [];
    eles = [];
    maps = {};
  };

  this.remove = id => {
    let i = maps['' + id];
    if (i === undefined) {
      return;
    }
    delete maps['' + id];
    delete vals[i];
    delete eles[i];
  };

  this.push = (id, ele) => {
    if (maps['' + id] !== undefined) return;
    ele.innerHTML='<div class="level-meter"></div>'
    ele = ele.querySelector('.level-meter')
    maps['' + id] = eles.length;
    vals.push('');
    eles.push(ele);
    this.length++;
  };

  this.setEle = (id, el) => {
    let i = maps['' + id];
    if (i !== undefined) {
      eles[i] = el;
    } else {
      this.push(id);
    }
  };

  this.eleSize = () => {
    let c = 0,
      i;
    for (i = 0; i < eles.length; i++) if (eles[i] !== undefined) c++;
    return c;
  };

  this.commitWidth = i => {
    if (eles[i] === undefined) return;
    let e = eles[i].style;
    if (e.right !== vals[i]) {
      e.right = vals[i];
    }
  };

  this.setVal = (i, val) => {
    if (i < 0 || i >= eles.length) {
      return;
    }
    if (typeof val === 'number') val = val + '%';
    else if ( typeof val === 'string' && !val.endsWith('%'))
      val = val + '%';

    vals[i] = val;
  };

  this.setValById = (id, val) => {
    let i = maps['' + id];
    if (i === undefined) {
      return;
    }
    this.setVal(i, val);
  };

  this.getValById = id => {
    let i = maps['' + id];
    if (i !== undefined) {
      return vals[i];
    }
    return '';
  };

  this.clear();
}

VolumeLevel.prototype.length = 0;

export default VolumeLevel;
