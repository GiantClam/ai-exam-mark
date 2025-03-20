import './globals.css';

export const metadata = {
  title: '作业手写文字提取工具',
  description: '使用AI技术从作业图片中快速提取手写文字内容，方便教师查看学生作答',
};

export default function RootLayout({ children }) {
  return (
    <html lang="zh-CN">
      <body>
        <div className="flex flex-col min-h-screen">
          {/* 导航栏 */}
          <header className="bg-white shadow-sm sticky top-0 z-10">
            <div className="container-custom py-4">
              <div className="flex justify-between items-center">
                <a href="/" className="text-xl font-bold text-blue-600 flex items-center">
                  <svg xmlns="http://www.w3.org/2000/svg" className="h-6 w-6 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z" />
                  </svg>
                  作业文字提取
                </a>
                
                <nav className="hidden md:flex items-center space-x-8">
                  <a href="/" className="text-gray-700 hover:text-blue-600 font-medium">首页</a>
                  <a href="/upload" className="text-gray-700 hover:text-blue-600 font-medium">上传作业</a>
                  <a href="/results" className="text-gray-700 hover:text-blue-600 font-medium">提取结果</a>
                </nav>
                
                <div className="md:hidden">
                  <button className="text-gray-700">
                    <svg xmlns="http://www.w3.org/2000/svg" className="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6h16M4 12h16M4 18h16" />
                    </svg>
                  </button>
                </div>
              </div>
            </div>
          </header>
          
          {/* 主内容区 */}
          <main className="flex-grow">
            {children}
          </main>
          
          {/* 页脚 */}
          <footer className="bg-gray-800 text-gray-300 py-8">
            <div className="container-custom">
              <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
                <div>
                  <h3 className="text-xl font-semibold text-white mb-4">作业手写文字提取工具</h3>
                  <p className="mb-4">
                    使用先进的AI技术从作业图片中提取手写文字内容，
                    帮助教师快速获取学生答案，提高工作效率。
                  </p>
                </div>
                
                <div>
                  <h3 className="text-xl font-semibold text-white mb-4">快速链接</h3>
                  <ul className="space-y-2">
                    <li><a href="/" className="hover:text-white">首页</a></li>
                    <li><a href="/upload" className="hover:text-white">上传作业</a></li>
                    <li><a href="/results" className="hover:text-white">提取结果</a></li>
                  </ul>
                </div>
              </div>
              
              <div className="border-t border-gray-700 mt-8 pt-8 text-sm text-gray-400 flex flex-col md:flex-row justify-between">
                <div>
                  &copy; {new Date().getFullYear()} 作业文字提取工具 - 保留所有权利
                </div>
                <div className="mt-4 md:mt-0">
                  <a href="#" className="hover:text-gray-300 mr-4">隐私政策</a>
                  <a href="#" className="hover:text-gray-300">使用条款</a>
                </div>
              </div>
            </div>
          </footer>
        </div>
      </body>
    </html>
  );
} 