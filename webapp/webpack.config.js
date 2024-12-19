const exec = require('child_process').exec

const path = require('path')

const ModuleFederationPlugin = require('webpack').container.ModuleFederationPlugin

const PLUGIN_ID = require('../plugin.json').id

const NPM_TARGET = process.env.npm_lifecycle_event //eslint-disable-line no-process-env
const isDev = NPM_TARGET === 'debug' || NPM_TARGET === 'debug:watch'

const plugins = [
  new ModuleFederationPlugin({
    name: 'parabol',
    shared: {
      react: {
        import: 'react', // the "react" package will be used a provided and fallback module
        shareKey: 'react', // under this name the shared module will be placed in the share scope
        shareScope: 'default', // share scope with this name will be used
        singleton: true, // only a single version of the shared module is allowed
      },
      'react-dom': {
        singleton: true, // only a single version of the shared module is allowed
      },
    },
  }),
]
if (NPM_TARGET === 'build:watch' || NPM_TARGET === 'debug:watch') {
  plugins.push({
    apply: (compiler) => {
      compiler.hooks.watchRun.tap('WatchStartPlugin', () => {
        // eslint-disable-next-line no-console
        console.log('Change detected. Rebuilding webapp.')
      })
      compiler.hooks.afterEmit.tap('AfterEmitPlugin', () => {
        exec('cd .. && make deploy-from-watch', (err, stdout, stderr) => {
          if (stdout) {
            process.stdout.write(stdout)
          }
          if (stderr) {
            process.stderr.write(stderr)
          }
        })
      })
    },
  })
}

const config = {
  entry: [
    './src/index.tsx',
  ],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, 'src'),
      '@mui/styled-engine': '@mui/styled-engine-sc',
    },
    modules: [
      'src',
      'node_modules',
      path.resolve(__dirname),
    ],
    extensions: ['*', '.js', '.jsx', '.ts', '.tsx'],
  },
  module: {
    rules: [
      {
        test: /\.(js|jsx|ts|tsx)$/,
        exclude: /node_modules/,
        use: {
          loader: 'babel-loader',
          options: {
            cacheDirectory: true,

            // Babel configuration is in babel.config.js because jest requires it to be there.
          },
        },
      },
      {
        test: /\.(scss|css)$/,
        use: [
          'style-loader',
          {
            loader: 'css-loader',
          },
          {
            loader: 'sass-loader',
            options: {
              sassOptions: {
                includePaths: ['node_modules/compass-mixins/lib', 'sass'],
              },
            },
          },
        ],
      },
    ],
  },
  optimization: {
    minimize: false,
  },
  externals: {
    react: 'React',
    'react-dom': 'ReactDOM',
    redux: 'Redux',
    'react-redux': 'ReactRedux',
    'prop-types': 'PropTypes',
    'react-bootstrap': 'ReactBootstrap',
    'react-router-dom': 'ReactRouterDom',
  },
  output: {
    devtoolNamespace: PLUGIN_ID,
    path: path.join(__dirname, '/dist'),
    publicPath: '/static/plugins/' + PLUGIN_ID + '/',
    filename: 'main.js',
  },
  mode: (isDev) ? 'eval-source-map' : 'production',
  plugins,
}

if (isDev) {
  Object.assign(config, {devtool: 'eval-source-map'})
}

module.exports = config
