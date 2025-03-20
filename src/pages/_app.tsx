import '../app/globals.css';
import type { AppProps } from 'next/app';
import Link from 'next/link';

export default function MyApp({ Component, pageProps }: AppProps) {
  return (
    <>
      <header className="bg-white border-b border-gray-100">
        <div className="container mx-auto p-4 flex justify-between items-center">
          <Link href="/" className="flex items-center text-blue-600">
            <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" fill="currentColor" className="mr-2" viewBox="0 0 16 16">
              <path d="M15.502 1.94a.5.5 0 0 1 0 .706L14.459 3.69l-2-2L13.502.646a.5.5 0 0 1 .707 0l1.293 1.293zm-1.75 2.456-2-2L4.939 9.21a.5.5 0 0 0-.121.196l-.805 2.414a.25.25 0 0 0 .316.316l2.414-.805a.5.5 0 0 0 .196-.12l6.813-6.814z"/>
              <path fillRule="evenodd" d="M1 13.5A1.5 1.5 0 0 0 2.5 15h11a1.5 1.5 0 0 0 1.5-1.5v-6a.5.5 0 0 0-1 0v6a.5.5 0 0 1-.5.5h-11a.5.5 0 0 1-.5-.5v-11a.5.5 0 0 1 .5-.5H9a.5.5 0 0 0 0-1H2.5A1.5 1.5 0 0 0 1 2.5v11z"/>
            </svg>
            <span className="text-xl font-medium">作业文字提取</span>
          </Link>
          <nav className="space-x-8">
            <Link href="/" className="text-gray-700 hover:text-blue-600">首页</Link>
            <Link href="/homework" className="text-gray-700 hover:text-blue-600">上传作业</Link>
            <Link href="/results" className="text-gray-700 hover:text-blue-600">提取结果</Link>
          </nav>
        </div>
      </header>
      <main>
        <Component {...pageProps} />
      </main>
      <footer className="bg-gray-50 py-8 border-t border-gray-100">
        <div className="container mx-auto px-4">
          <div className="flex flex-col md:flex-row justify-between">
            <div className="mb-4 md:mb-0">
              <p className="text-sm text-gray-700">&copy; {new Date().getFullYear()} 作业批改助手 - 由Gemini 2.0 Thinking模型提供支持</p>
            </div>
            <div className="flex space-x-6">
              <a href="#" className="text-gray-700 hover:text-blue-600">关于我们</a>
              <a href="#" className="text-gray-700 hover:text-blue-600">使用条款</a>
              <a href="#" className="text-gray-700 hover:text-blue-600">隐私政策</a>
            </div>
          </div>
        </div>
      </footer>
    </>
  );
} 