/** @type {import('next').NextConfig} */
const nextConfig = {
  // 基本配置
  reactStrictMode: true,
  swcMinify: true,
  
  // 禁用错误页面的自动静态优化
  experimental: {
    serverComponentsExternalPackages: ['next'],
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
};

export default nextConfig; 