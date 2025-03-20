/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  swcMinify: true,
  
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
};

module.exports = nextConfig 