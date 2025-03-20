'use client';

import { useState, useEffect } from 'react';
import Link from 'next/link';

export default function ResultsPage() {
  const [extractionResult, setExtractionResult] = useState(null);
  const [loading, setLoading] = useState(true);
  const [items, setItems] = useState([]);
  
  useEffect(() => {
    // 从localStorage获取数据
    const savedResult = localStorage.getItem('extractionResult');
    
    if (savedResult) {
      try {
        const parsedResult = JSON.parse(savedResult);
        setExtractionResult(parsedResult);
        
        if (parsedResult.extractedText) {
          setItems(parsedResult.extractedText);
        }
      } catch (err) {
        console.error('解析结果数据时出错:', err);
      }
    }
    
    setLoading(false);
  }, []);
  
  // 向上移动项目
  const moveItemUp = (index) => {
    if (index === 0) return;
    
    const newItems = [...items];
    const temp = newItems[index];
    newItems[index] = newItems[index - 1];
    newItems[index - 1] = temp;
    
    setItems(newItems);
  };
  
  // 向下移动项目
  const moveItemDown = (index) => {
    if (index === items.length - 1) return;
    
    const newItems = [...items];
    const temp = newItems[index];
    newItems[index] = newItems[index + 1];
    newItems[index + 1] = temp;
    
    setItems(newItems);
  };
  
  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-blue-500"></div>
      </div>
    );
  }
  
  if (!extractionResult || !extractionResult.extractedText) {
    return (
      <div className="min-h-screen bg-gray-50 flex flex-col items-center justify-center p-4">
        <div className="bg-white rounded-lg shadow-lg p-8 max-w-md w-full text-center">
          <svg xmlns="http://www.w3.org/2000/svg" className="h-16 w-16 text-gray-400 mx-auto mb-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9.172 16.172a4 4 0 015.656 0M9 10h.01M15 10h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          <h1 className="text-2xl font-bold text-gray-800 mb-3">没有可显示的结果</h1>
          <p className="text-gray-600 mb-6">
            您可能尚未上传作业图片，或者数据访问出现问题。
          </p>
          <Link 
            href="/upload" 
            className="bg-blue-600 text-white px-5 py-2 rounded-md hover:bg-blue-700 transition-colors inline-block"
          >
            前往上传页面
          </Link>
        </div>
      </div>
    );
  }
  
  return (
    <div className="bg-gray-50 min-h-screen">
      <div className="container-custom py-12">
        <div className="max-w-3xl mx-auto">
          <div className="text-center mb-8">
            <h1 className="text-3xl font-bold text-gray-900">提取结果</h1>
            <p className="mt-2 text-gray-600">以下是从您上传的作业图片中提取的文字内容</p>
            
            {extractionResult.layout && (
              <div className="inline-block bg-blue-50 text-blue-700 px-3 py-1 rounded-full text-sm mt-3">
                {extractionResult.layout === 'double' ? '双栏布局 (先左后右)' : '单栏布局'}
              </div>
            )}
          </div>
          
          <div className="bg-white rounded-xl shadow-lg overflow-hidden">
            {/* 提取结果 */}
            <div className="p-8">
              <div className="flex justify-between items-center mb-6 pb-2 border-b border-gray-100">
                <h2 className="text-xl font-semibold">提取内容</h2>
                <div className="text-gray-500 text-sm">
                  您可以通过左侧的按钮调整条目顺序
                </div>
              </div>
              
              <div className="space-y-6">
                {items.map((item, index) => (
                  <div key={index} className="flex items-start gap-2">
                    <div className="flex flex-col mt-2">
                      <button 
                        onClick={() => moveItemUp(index)}
                        disabled={index === 0}
                        className={`p-1 ${index === 0 ? 'text-gray-300' : 'text-gray-500 hover:text-blue-600'}`}
                        title="上移"
                      >
                        <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 15l7-7 7 7" />
                        </svg>
                      </button>
                      <button 
                        onClick={() => moveItemDown(index)}
                        disabled={index === items.length - 1}
                        className={`p-1 ${index === items.length - 1 ? 'text-gray-300' : 'text-gray-500 hover:text-blue-600'}`}
                        title="下移"
                      >
                        <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
                        </svg>
                      </button>
                    </div>
                    <div className="flex-1 bg-gray-50 p-4 rounded-lg">
                      <div className="font-medium text-gray-800 mb-2">{item.question}</div>
                      <div className="text-gray-700 whitespace-pre-wrap">{item.answer}</div>
                    </div>
                  </div>
                ))}
              </div>
              
              {extractionResult.feedback && (
                <div className="mt-8 pt-6 border-t border-gray-100">
                  <h2 className="text-xl font-semibold mb-3">
                    系统反馈
                  </h2>
                  <div className="bg-blue-50 p-4 rounded-lg text-gray-700">
                    {extractionResult.feedback}
                  </div>
                </div>
              )}
            </div>
            
            <div className="bg-gray-50 p-6 border-t border-gray-100 flex justify-between">
              <Link
                href="/upload"
                className="px-4 py-2 bg-white border border-gray-300 rounded shadow-sm hover:bg-gray-50 transition-colors"
              >
                重新上传
              </Link>
              
              <div className="flex space-x-3">
                <button
                  onClick={() => {
                    const textToCopy = items
                      .map(item => `${item.question}\n${item.answer}`)
                      .join('\n\n');
                    
                    navigator.clipboard.writeText(textToCopy)
                      .then(() => alert('内容已复制到剪贴板'))
                      .catch(err => console.error('复制失败:', err));
                  }}
                  className="px-4 py-2 bg-gray-600 text-white rounded shadow-sm hover:bg-gray-700 transition-colors"
                >
                  复制文本
                </button>
                
                <Link
                  href="/"
                  className="px-4 py-2 bg-blue-600 text-white rounded shadow-sm hover:bg-blue-700 transition-colors"
                >
                  返回首页
                </Link>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
} 