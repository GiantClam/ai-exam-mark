import Link from 'next/link';

export default function Home() {
  return (
    <div className="flex flex-col items-center py-12 px-4">
      {/* Hero Section */}
      <div className="text-center max-w-4xl mb-16">
        <h1 className="text-4xl font-bold mb-4 text-gray-900">作业手写文字提取工具</h1>
        <p className="text-lg text-gray-800">
          使用AI技术自动从作业图片中提取手写文字内容，快速获取学生作答内容
        </p>
      </div>

      {/* 3-Card Feature Section */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-8 max-w-6xl w-full mb-16">
        {/* Card 1 */}
        <div className="bg-white p-8 rounded-lg shadow-sm border border-gray-100">
          <div className="bg-blue-50 w-16 h-16 rounded-full flex items-center justify-center mb-6">
            <svg xmlns="http://www.w3.org/2000/svg" width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="text-blue-500">
              <rect x="3" y="3" width="18" height="18" rx="2" ry="2"></rect>
              <circle cx="8.5" cy="8.5" r="1.5"></circle>
              <polyline points="21 15 16 10 5 21"></polyline>
            </svg>
          </div>
          <h2 className="text-xl font-bold mb-3 text-gray-900">上传作业图片</h2>
          <p className="text-gray-800">
            拍摄或选择清晰的作业照片，确保内容完整可见
          </p>
        </div>
        
        {/* Card 2 */}
        <div className="bg-white p-8 rounded-lg shadow-sm border border-gray-100">
          <div className="bg-green-50 w-16 h-16 rounded-full flex items-center justify-center mb-6">
            <svg xmlns="http://www.w3.org/2000/svg" width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="text-green-500">
              <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"></path>
              <polyline points="14 2 14 8 20 8"></polyline>
              <line x1="16" y1="13" x2="8" y2="13"></line>
              <line x1="16" y1="17" x2="8" y2="17"></line>
              <polyline points="10 9 9 9 8 9"></polyline>
            </svg>
          </div>
          <h2 className="text-xl font-bold mb-3 text-gray-900">AI提取文字</h2>
          <p className="text-gray-800">
            先进的AI算法自动识别图片中的手写文字内容
          </p>
        </div>
        
        {/* Card 3 */}
        <div className="bg-white p-8 rounded-lg shadow-sm border border-gray-100">
          <div className="bg-purple-50 w-16 h-16 rounded-full flex items-center justify-center mb-6">
            <svg xmlns="http://www.w3.org/2000/svg" width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="text-purple-500">
              <path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"></path>
              <line x1="12" y1="11" x2="12" y2="17"></line>
              <line x1="9" y1="14" x2="15" y2="14"></line>
            </svg>
          </div>
          <h2 className="text-xl font-bold mb-3 text-gray-900">结果展示</h2>
          <p className="text-gray-800">
            清晰展示提取结果，方便查看、复制和保存文字内容
          </p>
        </div>
      </div>

      {/* CTA Button */}
      <Link 
        href="/homework"
        className="bg-blue-600 hover:bg-blue-700 text-white font-medium px-8 py-3 rounded-md transition-colors shadow-sm"
      >
        立即开始使用
      </Link>
    </div>
  );
} 