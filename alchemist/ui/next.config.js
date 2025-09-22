/** @type {import('next').NextConfig} */
const nextConfig = {
  swcMinify: true,
  output: 'export',
  trailingSlash: true,
  distDir: 'dist',
};

module.exports = nextConfig;
