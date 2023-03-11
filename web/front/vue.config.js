const webpack = require('webpack');
const { defineConfig } = require('@vue/cli-service');
const path = require('path');

function resolve(dir) {
  return path.join(__dirname, dir);
}
const isProd = process.env.NODE_ENV === "production";

process.env.AppName = "Castspeaker Server";
process.env.AppID = "castserver";
process.env.AppTitle = "Castspeaker 管理后台";
process.env.Lang = "zh";
process.env.PLATFORM = "browser";

module.exports = defineConfig({
  outputDir: path.resolve(__dirname, "../public"),
  lintOnSave: false,
  // 生产环境打包不输出 map
  productionSourceMap: false,
  configureWebpack: {
    plugins: [
      new webpack.DefinePlugin({ // 替换变量为
        'process.env.NODE_ENV': JSON.stringify(process.env.NODE_ENV),
        'process.platform': JSON.stringify(process.env.PLATFORM || process.platform),
        'process.env.AppName': JSON.stringify(process.env.AppName),
        'process.env.AppID': JSON.stringify(process.env.AppID),
        'process.env.Lang': process.env.Lang,
        'process.env.Mock': isProd ? false : (process.argv.indexOf('--mock') > 0 ? true : false),
      }),
    ],
    devtool: isProd ? false : "source-map",
    performance: {   //  就是为了加大文件允许体积，提升报错门栏。  
      hints: "warning", // 枚举
      maxAssetSize: 50000000, // 整数类型（以字节为单位）
      maxEntrypointSize: 50000000, // 整数类型（以字节为单位）
      assetFilter: function (assetFilename) {
        // 提供资源文件名的断言函数        
        return assetFilename.endsWith('.css') || assetFilename.endsWith('.js');
      }
    },
    optimization: {
      splitChunks: false,
    },
  },
  devServer: {
    hot: true,
    port: process.env.DEV_SERVER_PORT || 8080,
    proxy: {
      '^/api': {
        target: `http://localhost:${process.env.DEV_MOCK_PORT}`,
        changeOrigin: true,
      },
    },
  },
  pwa: {
    name: process.env.AppName,
    iconPaths: {
      favicon32: 'img/icons/favicon-32x32.png',
    },
    themeColor: '#ffffff00',
    manifestOptions: {
      background_color: '#335eea',
    },
    // workboxOptions: {
    //   swSrc: "dev/sw.js",
    // },
  },
  pages: {
    index: {
      entry: 'src/main.js',
      template: 'public/index.html',
      filename: 'index.html',
      title: process.env.AppTitle,
      chunks: ['main', 'chunk-vendors', 'chunk-common', 'index'],
    },
  },
  transpileDependencies: true,
  css: {
    loaderOptions: {
      less: {
        lessOptions: {
          javascriptEnabled: true,
          math: "always",
        }
      },
      scss: {
        additionalData: `
          @import "@/assets/css/global.scss";
        `
      }
    }
  },
  chainWebpack(config) {
    config.resolve.symlinks(true)
    config.module.rules.delete('svg');
    config.module.rule('svg').exclude.add(resolve('src/assets/icons')).end();
    config.module
      .rule('icons')
      .test(/\.svg$/)
      .include.add(resolve('src/assets/icons'))
      .end()
      .use('svg-sprite-loader')
      .loader('svg-sprite-loader')
      .options({
        symbolId: 'icon-[name]',
      })
      .end();
    config.module
      .rule('napi')
      .test(/\.node$/)
      .use('node-loader')
      .loader('node-loader')
      .end();
    config.module
      .rule('less')
      .test(/\.less$/)
      .use('less-loader')
      .loader('less-loader')
      .end();

    // LimitChunkCountPlugin 可以通过合并块来对块进行后期处理。用以解决 chunk 包太多的问题
    // config.plugin('chunkPlugin')
    //   .use(webpack.optimize.LimitChunkCountPlugin, [
    //     {
    //       maxChunks: 1,
    //     },
    //   ]).use(webpack.optimize.MinChunkSizePlugin, [
    //     {
    //       minChunkSize: 4000000,
    //     }
    //   ]);
  },
  // 添加插件的配置
  pluginOptions: {
    // electron-builder的配置文件
    electronBuilder: {
      nodeIntegration: true,
      externals: [],
      builderOptions: {
        productName: process.env.AppName,
        copyright: 'Copyright © ' + process.env.AppName,
        // compression: "maximum", // 机器好的可以打开，配置压缩，开启后会让 .AppImage 格式的客户端启动缓慢
        asar: true,
        directories: {
          output: 'dist_electron',
        },
        mac: {
          target: [
            {
              target: 'dmg',
              arch: ['x64', 'arm64', 'universal'],
            },
          ],
          artifactName: '${productName}-${os}-${version}-${arch}.${ext}',
          category: 'public.app-category.music',
          darkModeSupport: true,
        },
        win: {
          target: [
            {
              target: 'portable',
              arch: ['x64'],
            },
            {
              target: 'nsis',
              arch: ['x64'],
            },
          ],
          publisherName: process.env.AppName,
          icon: 'build/icons/icon.ico',
          publish: ['github'],
        },
        linux: {
          target: [
            {
              target: 'AppImage',
              arch: ['x64'],
            },
            {
              target: 'tar.gz',
              arch: ['x64', 'arm64'],
            },
            {
              target: 'deb',
              arch: ['x64', 'armv7l', 'arm64'],
            },
            {
              target: 'rpm',
              arch: ['x64'],
            },
            {
              target: 'snap',
              arch: ['x64'],
            },
            {
              target: 'pacman',
              arch: ['x64'],
            },
          ],
          category: 'Music',
          icon: './build/icon.icns',
        },
        dmg: {
          icon: 'build/icons/icon.icns',
        },
        nsis: {
          oneClick: true,
          perMachine: true,
          deleteAppDataOnUninstall: true,
        },
      },
      // 主线程的配置文件
      chainWebpackMainProcess: config => {
        config.plugin('define').tap(args => {
          args[0]['IS_ELECTRON'] = true;
          return args;
        });
        config.resolve.alias.set(
          'jsbi',
          path.join(__dirname, 'node_modules/jsbi/dist/jsbi-cjs.js')
        );
      },
      // 渲染线程的配置文件
      chainWebpackRendererProcess: config => {
        // 渲染线程的一些其他配置
        // Chain webpack config for electron renderer process only
        // The following example will set IS_ELECTRON to true in your app
        config.plugin('define').tap(args => {
          args[0]['IS_ELECTRON'] = true;
          return args;
        });
      },
      // 主入口文件
      // mainProcessFile: 'src/main.js',
      // mainProcessArgs: []
    },
  },
});
