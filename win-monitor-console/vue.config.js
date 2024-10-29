module.exports = {
    publicPath: '/dist/',
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
    devServer: {
        proxy: {
            '/api': {
                target: 'http://localhost/', // 本地API服务的地址
                changeOrigin: true,
            }
        }
    }
};
