import mock from 'mockjs';
import Socket from '@/common/ws';

function generateTPL(id, tpl) {
  let template = {};
  if (id) template['data|' + id] = [tpl];
  else template['data|0-10'] = [tpl];

  return template;
}
let funcTpl = (tpl, id) => {
  return () => {
    return mock.mock(generateTPL(id, tpl)).data;
  };
};
let nomarlTpl = (tpl, id) => {
  return mock.mock(generateTPL(id, tpl)).data;
};

mock.mock(
  /\/api\/speakers$/,
  funcTpl({
    id: '@integer(0, 999999)',
    name: '@ctitle',
    ip: '@ip',
    mac: '@mac',
    volume: '@integer(0, 99)',
  })
);
Socket.addBeforeSend('speakerList', () => {
  return nomarlTpl({
    id: '@integer(0, 999999)',
    name: '@ctitle',
    ip: '@ip',
    mac: '@mac',
    volume: '@integer(0, 99)',
  });
});

/*
mock.mock(/\/api\/speaker\/\d+$/, (options) => {
  let id = options.url.match(/\/(\d+)$/)[1]
  return mock.mock({
    id: id,
    name: '@ctitle',
    ip: '@ip',
    mac: '@mac',
    volume: '@integer(0, 99)',
  })
});
Socket.addBeforeSend('speakerInfo', (params) => {
  return mock.mock({
    id: params.id,
    name: '@ctitle',
    ip: '@ip',
    mac: '@mac',
    volume: '@integer(0, 99)',
  });
});
*/

mock.mock(
  /\/api\/lines$/,
  funcTpl({
    id: '@integer(0, 999999)',
    name: '@ctitle',
  })
);
Socket.addBeforeSend('lineList', function () {
  return nomarlTpl(
    {
      id: '@integer(0, 999999)',
      name: '@ctitle',
    },
    '1-20'
  );
});

mock.mock(/\/api\/line\/\d+$/, options => {
  let id = options.url.match(/\/(\d+)$/)[1];
  return mock.mock({
    id: id,
    name: '@ctitle',
    ip: '@ip',
    volume: '@integer(0, 99)',
  });
});
Socket.addBeforeSend('lineInfo', function (params) {
  return mock.mock({
    id: params.id,
    name: '@ctitle',
  });
});
