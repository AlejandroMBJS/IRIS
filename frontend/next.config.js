/** @type {import('next').NextConfig} */
const nextConfig = {
  typescript: {
    // ignoreBuildErrors: true, // Removed for production readiness
  },
  // Enable standalone output for Docker deployment
  output: 'standalone',
}

module.exports = nextConfig

