'use client';

import { Inter } from 'next/font/google';
import './globals.css';

const inter = Inter({ subsets: ['latin'] });

export default function GlobalError({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  return (
    <html lang="zh-CN">
      <body className={inter.className}>
        <div className="flex flex-col items-center justify-center min-h-screen text-center px-4">
          <h2 className="text-3xl font-bold mb-4 dark:text-white">系统错误</h2>
          <p className="mb-8 text-gray-600 dark:text-gray-300">
            抱歉，应用程序出现了严重错误。
          </p>
          <button
            onClick={() => reset()}
            className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 transition-colors"
          >
            重试
          </button>
        </div>
      </body>
    </html>
  );
} 