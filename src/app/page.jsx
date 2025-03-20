import Link from 'next/link';

export default function HomePage() {
  return (
    <div className="bg-gray-50 min-h-screen">
      {/* 主要内容区 */}
      <div className="container-custom py-16">
        <div className="text-center mb-12">
          <h1 className="text-4xl font-bold text-gray-900 mb-4">作业手写文字提取工具</h1>
          <p className="text-lg text-gray-600 max-w-3xl mx-auto">
            使用AI技术自动从作业图片中提取手写文字内容，快速获取学生作答内容
          </p>
        </div>
        
        <div className="grid grid-cols-1 md:grid-cols-3 gap-8 max-w-6xl mx-auto mb-16">
          {/* 功能卡片1 */}
          <div className="bg-white rounded-xl shadow-md overflow-hidden transition-transform hover:scale-[1.02] hover:shadow-lg">
            <div className="p-6">
              <div className="w-12 h-12 rounded-full bg-blue-100 flex items-center justify-center mb-4">
                <svg xmlns="http://www.w3.org/2000/svg" className="h-6 w-6 text-blue-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
                </svg>
              </div>
              <h2 className="text-xl font-semibold text-gray-800 mb-2">上传作业图片</h2>
              <p className="text-gray-600 mb-4">
                拍摄或选择清晰的作业照片，确保内容完整可见
              </p>
            </div>
          </div>
          
          {/* 功能卡片2 */}
          <div className="bg-white rounded-xl shadow-md overflow-hidden transition-transform hover:scale-[1.02] hover:shadow-lg">
            <div className="p-6">
              <div className="w-12 h-12 rounded-full bg-green-100 flex items-center justify-center mb-4">
                <svg xmlns="http://www.w3.org/2000/svg" className="h-6 w-6 text-green-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
                </svg>
              </div>
              <h2 className="text-xl font-semibold text-gray-800 mb-2">AI提取文字</h2>
              <p className="text-gray-600 mb-4">
                先进的AI算法自动识别图片中的手写文字内容
              </p>
            </div>
          </div>
          
          {/* 功能卡片3 */}
          <div className="bg-white rounded-xl shadow-md overflow-hidden transition-transform hover:scale-[1.02] hover:shadow-lg">
            <div className="p-6">
              <div className="w-12 h-12 rounded-full bg-purple-100 flex items-center justify-center mb-4">
                <svg xmlns="http://www.w3.org/2000/svg" className="h-6 w-6 text-purple-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                </svg>
              </div>
              <h2 className="text-xl font-semibold text-gray-800 mb-2">结果展示</h2>
              <p className="text-gray-600 mb-4">
                清晰展示提取结果，方便查看、复制和保存文字内容
              </p>
            </div>
          </div>
        </div>
        
        {/* CTA区域 */}
        <div className="text-center">
          <Link 
            href="/upload"
            className="inline-block bg-blue-600 text-white px-8 py-3 rounded-lg font-medium shadow-md hover:bg-blue-700 hover:shadow-lg transition-all transform hover:-translate-y-1"
          >
            立即开始使用
          </Link>
        </div>
      </div>
      
      {/* 功能介绍区 */}
      <div className="bg-white py-16">
        <div className="container-custom">
          <div className="max-w-4xl mx-auto">
            <h2 className="text-3xl font-bold text-center text-gray-900 mb-12">工具特点</h2>
            
            <div className="space-y-8">
              <div className="flex items-start">
                <div className="flex-shrink-0 mt-1">
                  <div className="w-8 h-8 rounded-full bg-blue-100 flex items-center justify-center">
                    <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4 text-blue-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                    </svg>
                  </div>
                </div>
                <div className="ml-4">
                  <h3 className="text-xl font-semibold text-gray-800 mb-1">高精度OCR识别</h3>
                  <p className="text-gray-600">
                    使用先进的AI技术实现对中文、英文和数学等各类手写内容的准确识别，适应不同的书写风格和习惯
                  </p>
                </div>
              </div>
              
              <div className="flex items-start">
                <div className="flex-shrink-0 mt-1">
                  <div className="w-8 h-8 rounded-full bg-blue-100 flex items-center justify-center">
                    <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4 text-blue-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                    </svg>
                  </div>
                </div>
                <div className="ml-4">
                  <h3 className="text-xl font-semibold text-gray-800 mb-1">结构化输出</h3>
                  <p className="text-gray-600">
                    自动识别题目和答案，将提取结果以结构化方式展示，便于后续处理和分析
                  </p>
                </div>
              </div>
              
              <div className="flex items-start">
                <div className="flex-shrink-0 mt-1">
                  <div className="w-8 h-8 rounded-full bg-blue-100 flex items-center justify-center">
                    <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4 text-blue-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                    </svg>
                  </div>
                </div>
                <div className="ml-4">
                  <h3 className="text-xl font-semibold text-gray-800 mb-1">简单易用</h3>
                  <p className="text-gray-600">
                    直观的用户界面，只需上传图片即可获取结果，无需复杂操作，适合各类用户使用
                  </p>
                </div>
              </div>
              
              <div className="flex items-start">
                <div className="flex-shrink-0 mt-1">
                  <div className="w-8 h-8 rounded-full bg-blue-100 flex items-center justify-center">
                    <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4 text-blue-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                    </svg>
                  </div>
                </div>
                <div className="ml-4">
                  <h3 className="text-xl font-semibold text-gray-800 mb-1">快速处理</h3>
                  <p className="text-gray-600">
                    强大的后台处理能力，迅速完成图像分析和文字提取，节省人工录入时间
                  </p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
} 