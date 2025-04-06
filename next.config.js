/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  swcMinify: true,
  transpilePackages: ['@ant-design/icons', 'antd'],
  webpack: (config, { isServer }) => {
    config.module.rules.push({
      test: /\.(js|jsx|ts|tsx)$/,
      exclude: /node_modules/,
      use: {
        loader: 'babel-loader',
        options: {
          presets: ['next/babel'],
          plugins: [
            ['@babel/plugin-proposal-decorators', { legacy: true }],
            ['@babel/plugin-proposal-class-properties', { loose: true }],
            ['@babel/plugin-proposal-private-methods', { loose: true }],
            ['@babel/plugin-proposal-private-property-in-object', { loose: true }]
          ]
        }
      }
    });
    return config;
  },
  
  // 重定向配置
  async redirects() {
    return [
      {
        source: '/404',
        destination: '/not-found',
        permanent: true,
      },
    ];
  },

  // API 代理配置
  async rewrites() {
    return [
      {
        source: '/api/:path*',
        destination: 'http://localhost:8280/api/:path*',
      },
    ];
  },
};

module.exports = nextConfig 