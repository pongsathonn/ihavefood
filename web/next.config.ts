import type { NextConfig } from 'next'

const nextConfig: NextConfig = {
  devIndicators: {
    position: 'bottom-left',
  },
  images: {
    remotePatterns: [
      new URL('https://wadpgqajugnhnkf9.public.blob.vercel-storage.com/**'),
    ],
  },
}

export default nextConfig
