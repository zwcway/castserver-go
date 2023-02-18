const path = require('path')

const { VueLoaderPlugin } = require('vue-loader')
const HtmlWebpackPlugin = require('html-webpack-plugin')

module.exports = {
  entry: './src/main.js', //入口文件的地址
  output: {
    path: path.resolve(__dirname + '/dist'), //必须提供一个绝对路径
    filename: 'bundle.js', //打包后主文件的名称,输出文件的名称
  }, //输出文件夹的配置
  module: {
    rules: [
      { test: /\.vue$/, use: 'vue-loader' }, //以什么模块去处理某些类型的文件
      { test: /\.s[ca]ss$/, use: ['style-loader', 'css-loader', 'scss-loader'] },

      {
        test: /\.m?js$/,
        use: { // 将es6转换成es5
          loader: 'babel-loader',
          options: {
            presets: ['@babel/preset-env']
          }
        }
      },

      { test: /\.(png|jpe?g|gif|svg|webp)$/, type: 'asset/resource' }

    ]
  }, //使用哪些模块去处理文件
  plugins: [
    new VueLoaderPlugin(),
    new HtmlWebpackPlugin({
      title: 'Webpack Vue',
      template: './public/index.html'
    }),
  ], //项目中使用了哪些插件
}