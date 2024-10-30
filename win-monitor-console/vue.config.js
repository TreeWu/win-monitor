const NodePolyfillPlugin = require('node-polyfill-webpack-plugin');

module.exports = {
    configureWebpack: {
        plugins: [
            new NodePolyfillPlugin()
        ],
        resolve: {
            fallback: {
                "url": require.resolve("url/")
            }
        }
    },
    publicPath: '/console/',
    css: {
        loaderOptions: {
            css: {}, less: {
                lessOptions: {
                    modifyVars: {
                        'primary-color': '#1DA57A', 'link-color': '#1DA57A', 'border-radius-base': '2px',
                    },
                    javascriptEnabled: true,
                },
            },
        },
    },
    chainWebpack: config => {
        config.module
            .rule('ts')
            .test(/\.ts$/)
            .use('ts-loader')
            .loader('ts-loader')
            .end();
    },
    devServer: {
        proxy: {
            '/api/console': {
                target: 'http://localhost/', // 本地API服务的地址
                changeOrigin: true,
            }
        }
    }
};
