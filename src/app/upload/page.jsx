'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';

export default function UploadPage() {
  const [file, setFile] = useState(null);
  const [preview, setPreview] = useState(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [layout, setLayout] = useState('single'); // 默认单栏布局
  const router = useRouter();

  const handleFileChange = (e) => {
    setError('');
    const selectedFile = e.target.files[0];
    if (selectedFile) {
      if (!selectedFile.type.includes('image/')) {
        setError('请上传图片文件');
        setFile(null);
        setPreview(null);
        return;
      }
      
      setFile(selectedFile);
      
      // 创建预览
      const reader = new FileReader();
      reader.onload = () => {
        setPreview(reader.result);
      };
      reader.readAsDataURL(selectedFile);
    }
  };

  const handleDragOver = (e) => {
    e.preventDefault();
    e.stopPropagation();
  };

  const handleDrop = (e) => {
    e.preventDefault();
    e.stopPropagation();
    
    setError('');
    const droppedFile = e.dataTransfer.files[0];
    
    if (droppedFile) {
      if (!droppedFile.type.includes('image/')) {
        setError('请上传图片文件');
        return;
      }
      
      setFile(droppedFile);
      
      // 创建预览
      const reader = new FileReader();
      reader.onload = () => {
        setPreview(reader.result);
      };
      reader.readAsDataURL(droppedFile);
    }
  };

  const handleLayoutChange = (e) => {
    setLayout(e.target.value);
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    
    if (!file) {
      setError('请先选择要上传的图片');
      return;
    }
    
    setLoading(true);
    setError('');
    
    try {
      // 创建FormData对象
      const formData = new FormData();
      formData.append('homework', file);
      
      // 添加作业类型参数，这里默认使用general类型
      // 如果您的应用需要指定不同类型的作业（如数学、英文等），可以添加选择类型的UI
      formData.append('type', 'general');
      
      // 添加布局参数
      formData.append('layout', layout);
      
      // 发送到Golang后端的API
      const response = await fetch('/api/analyze', {
        method: 'POST',
        body: formData,
      });
      
      // 检查响应状态
      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || '服务器处理失败');
      }
      
      // 解析响应数据
      const result = await response.json();
      
      if (!result.success) {
        throw new Error(result.error || '处理失败');
      }
      
      // 构建提取结果对象
      const extractedText = [];
      
      // 检查result.result.answers是否存在并处理
      if (result.result && result.result.answers && Array.isArray(result.result.answers)) {
        // 将后端返回的结果转换为前端期望的格式
        result.result.answers.forEach(answer => {
          extractedText.push({
            question: answer.questionNumber || "未知题号",
            answer: answer.studentAnswer || "未能提取答案"
          });
        });
      } else if (typeof result.result === 'string') {
        // 如果result.result是字符串，尝试解析JSON
        try {
          const parsedResult = JSON.parse(result.result);
          if (parsedResult.answers && Array.isArray(parsedResult.answers)) {
            parsedResult.answers.forEach(answer => {
              extractedText.push({
                question: answer.questionNumber || "未知题号",
                answer: answer.studentAnswer || "未能提取答案"
              });
            });
          }
        } catch (error) {
          // 解析失败时，将整个内容作为单个结果返回
          extractedText.push({
            question: "提取内容",
            answer: result.result
          });
        }
      } else {
        // 无法识别的结果格式，返回一个通用条目
        extractedText.push({
          question: "提取内容",
          answer: "无法解析结果，请查看原始图片"
        });
      }
      
      // 保存提取结果并导航到结果页面
      localStorage.setItem('extractionResult', JSON.stringify({ 
        extractedText,
        feedback: result.result && result.result.feedback ? result.result.feedback : ''
      }));
      
      router.push('/results');
    } catch (err) {
      console.error('上传错误:', err);
      setError(err.message || '上传图片时发生错误，请重试');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="bg-gray-50 min-h-screen">
      <div className="container-custom py-12">
        <div className="max-w-3xl mx-auto">
          <div className="text-center mb-8">
            <h1 className="text-3xl font-bold text-gray-900">上传作业图片</h1>
            <p className="mt-2 text-gray-600">上传清晰的作业照片，自动提取手写文字内容</p>
          </div>
          
          <div className="bg-white rounded-xl shadow-lg overflow-hidden">
            <div className="p-8">
              <form onSubmit={handleSubmit}>
                <div 
                  className="mb-6 border-2 border-dashed border-gray-300 rounded-lg p-6 text-center cursor-pointer hover:border-blue-500 transition-colors"
                  onDragOver={handleDragOver}
                  onDrop={handleDrop}
                  onClick={() => document.getElementById('file-upload').click()}
                >
                  <input
                    id="file-upload"
                    type="file"
                    accept="image/*"
                    onChange={handleFileChange}
                    className="hidden"
                  />
                  
                  {!preview ? (
                    <div className="py-8">
                      <svg xmlns="http://www.w3.org/2000/svg" className="mx-auto h-12 w-12 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
                      </svg>
                      <p className="mt-2 text-sm text-gray-600">
                        点击上传或拖放图片到此处
                      </p>
                      <p className="mt-1 text-xs text-gray-500">
                        支持 JPG、PNG、GIF 等常见图片格式
                      </p>
                    </div>
                  ) : (
                    <div className="relative group">
                      <img 
                        src={preview} 
                        alt="预览"
                        className="max-h-80 mx-auto object-contain rounded"
                      />
                      <div className="absolute inset-0 bg-black bg-opacity-50 flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity rounded">
                        <p className="text-white">点击更换图片</p>
                      </div>
                    </div>
                  )}
                </div>
                
                {/* 布局选择 */}
                <div className="mb-6">
                  <label className="block text-gray-700 font-medium mb-2">
                    试卷布局
                  </label>
                  <div className="flex space-x-4">
                    <label className="flex items-center">
                      <input
                        type="radio"
                        name="layout"
                        value="single"
                        checked={layout === 'single'}
                        onChange={handleLayoutChange}
                        className="mr-2 h-4 w-4 text-blue-600 focus:ring-blue-500"
                      />
                      <span>单栏布局</span>
                    </label>
                    <label className="flex items-center">
                      <input
                        type="radio"
                        name="layout"
                        value="double"
                        checked={layout === 'double'}
                        onChange={handleLayoutChange}
                        className="mr-2 h-4 w-4 text-blue-600 focus:ring-blue-500"
                      />
                      <span>双栏布局 (先左后右)</span>
                    </label>
                  </div>
                  <p className="mt-1 text-sm text-gray-600">
                    双栏布局选项可改善左右分栏试卷的识别顺序
                  </p>
                </div>
                
                {error && (
                  <div className="mb-6 p-4 bg-red-50 border-l-4 border-red-500 text-red-700 rounded">
                    <div className="flex">
                      <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5 mr-2" viewBox="0 0 20 20" fill="currentColor">
                        <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7 4a1 1 0 11-2 0 1 1 0 012 0zm-1-9a1 1 0 00-1 1v4a1 1 0 102 0V6a1 1 0 00-1-1z" clipRule="evenodd" />
                      </svg>
                      {error}
                    </div>
                  </div>
                )}
                
                <div className="flex justify-between">
                  <Link
                    href="/"
                    className="px-4 py-2 text-gray-700 bg-gray-100 rounded-md hover:bg-gray-200 transition-colors"
                  >
                    返回首页
                  </Link>
                  
                  <button
                    type="submit"
                    disabled={loading || !file}
                    className={`px-6 py-2 rounded-md font-medium ${
                      loading || !file
                        ? 'bg-gray-300 cursor-not-allowed'
                        : 'bg-blue-600 text-white hover:bg-blue-700 shadow-md hover:shadow-lg'
                    } transition-all`}
                  >
                    {loading ? (
                      <span className="flex items-center">
                        <svg className="animate-spin -ml-1 mr-2 h-4 w-4 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                          <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                          <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                        </svg>
                        处理中...
                      </span>
                    ) : '提取文字'}
                  </button>
                </div>
              </form>
            </div>
            
            <div className="bg-gray-50 p-6 border-t border-gray-100">
              <h2 className="text-lg font-semibold mb-3">使用说明</h2>
              <ol className="list-decimal pl-5 space-y-2 text-gray-600">
                <li>拍摄或选择一张清晰的作业图片</li>
                <li>上传图片并等待系统处理（可能需要10-20秒）</li>
                <li>查看自动提取的手写文字内容</li>
              </ol>
              <div className="mt-4 text-sm text-gray-500">
                <p>* 为获得最佳提取效果，请确保照片清晰、光线充足、作业内容完整可见</p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
} 