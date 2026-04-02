'use strict';

const fs = require('node:fs');
const path = require('node:path');

const distFiles = ['layout.js', 'analysis.js', 'bidi.js', 'line-break.js', 'measurement.js'];
const packageRoot = path.dirname(require.resolve('@chenglou/pretext/package.json'));

hexo.extend.generator.register('happyladysacue-pretext-vendor', function () {
  return distFiles.map((file) => ({
    path: `vendor/pretext/${file}`,
    data: () => fs.createReadStream(path.join(packageRoot, 'dist', file)),
  }));
});
