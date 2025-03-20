'use client';

import Link from 'next/link';

export default function NotFound() {
  return (
    <div className="min-h-screen bg-gray-50 flex items-center justify-center px-4">
      <div className="max-w-lg w-full text-center">
        <div className="mb-8">
          <div className="relative mx-auto w-48 h-48">
            <div className="absolute inset-0 bg-blue-100 rounded-full opacity-50 animate-ping-slow"></div>
            <div className="relative z-10 bg-white rounded-full w-full h-full flex items-center justify-center shadow-lg">
              <p className="text-7xl font-bold text-blue-600">404</p>
            </div>
          </div>
        </div>
        
        <h2 className="text-3xl font-bold text-gray-900 mb-3">页面未找到</h2>
        <p className="text-lg text-gray-600 mb-8 max-w-md mx-auto">
          抱歉，您请求的页面不存在或已被移动。
        </p>
        
        <Link
          href="/"
          className="inline-flex items-center justify-center px-6 py-3 border border-transparent text-base font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 transition-colors shadow-md"
        >
          <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5 mr-2" viewBox="0 0 20 20" fill="currentColor">
            <path fillRule="evenodd" d="M9.707 16.707a1 1 0 01-1.414 0l-6-6a1 1 0 010-1.414l6-6a1 1 0 011.414 1.414L5.414 9H17a1 1 0 110 2H5.414l4.293 4.293a1 1 0 010 1.414z" clipRule="evenodd" />
          </svg>
          返回首页
        </Link>
      </div>
      
      <style jsx>{`
        @keyframes ping-slow {
          0% {
            transform: scale(0.8);
            opacity: 0.5;
          }
          50% {
            transform: scale(1.1);
            opacity: 0.3;
          }
          100% {
            transform: scale(0.8);
            opacity: 0.5;
          }
        }
        .animate-ping-slow {
          animation: ping-slow 3s cubic-bezier(0, 0, 0.2, 1) infinite;
        }
      `}</style>
    </div>
  );
} 