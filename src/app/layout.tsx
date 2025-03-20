import { Inter } from 'next/font/google';
import './globals.css';

export const metadata = {
  title: '作业批改助手',
  description: '使用AI智能批改学生作业',
  icons: {
    icon: '/icon.svg',
    shortcut: '/icon.svg',
    apple: '/icon.svg',
  },
};

const inter = Inter({ subsets: ['latin'] });

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="zh-CN">
      <head>
        <link rel="icon" href="/icon.svg" />
        <link rel="shortcut icon" href="/favicon.ico" />
      </head>
      <body className={inter.className}>
        <header className="bg-white dark:bg-gray-900 shadow">
          <div className="container mx-auto p-4 flex justify-between items-center">
            <h1 className="text-2xl font-bold text-blue-600 dark:text-blue-400">作业批改助手</h1>
            <nav>
              <a href="/" className="text-gray-700 dark:text-gray-200 hover:text-blue-600 dark:hover:text-blue-400 mr-4">首页</a>
              <a href="/homework" className="text-gray-700 dark:text-gray-200 hover:text-blue-600 dark:hover:text-blue-400">批改</a>
            </nav>
          </div>
        </header>
        <main>
          {children}
        </main>
        <footer className="bg-gray-800 dark:bg-gray-900 text-white py-8">
          <div className="container mx-auto px-4">
            <div className="flex flex-col md:flex-row justify-between">
              <div className="mb-4 md:mb-0">
                <p className="text-sm dark:text-gray-200">&copy; {new Date().getFullYear()} 作业批改助手 - 由Gemini 2.0 Thinking模型提供支持</p>
              </div>
              <div className="flex">
                <a href="#" className="text-gray-300 hover:text-white dark:text-gray-300 dark:hover:text-white mr-4">
                  关于我们
                </a>
                <a href="#" className="text-gray-300 hover:text-white dark:text-gray-300 dark:hover:text-white mr-4">
                  使用条款
                </a>
                <a href="#" className="text-gray-300 hover:text-white dark:text-gray-300 dark:hover:text-white">
                  隐私政策
                </a>
              </div>
            </div>
          </div>
        </footer>
      </body>
    </html>
  );
}
