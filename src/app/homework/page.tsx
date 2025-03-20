'use client';

import { useState, useRef } from 'react';
import Image from 'next/image';
import Link from 'next/link';

export default function HomeworkPage() {
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [previewUrl, setPreviewUrl] = useState<string | null>(null);
  const [homeworkType, setHomeworkType] = useState<string>('english');
  const [isLoading, setIsLoading] = useState<boolean>(false);
  const [result, setResult] = useState<any>(null);
  const [error, setError] = useState<string | null>(null);
  
  const fileInputRef = useRef<HTMLInputElement>(null);
  
  // 获取API URL
  const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api';
  
  // 为每个结果卡片添加折叠/展开功能
  const [expandedCards, setExpandedCards] = useState<{[key: number]: boolean}>({});
  
  // 切换卡片的折叠/展开状态
  const toggleCardExpand = (index: number) => {
    setExpandedCards(prev => ({
      ...prev,
      [index]: !prev[index]
    }));
  };
  
  // 触发文件选择对话框
  const handleSelectClick = () => {
    if (fileInputRef.current) {
      fileInputRef.current.click();
    }
  };
  
  // 处理文件选择
  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const files = e.target.files;
    if (files && files.length > 0) {
      const file = files[0];
      
      // 验证文件类型
      if (!file.type.startsWith('image/')) {
        setError('请选择图片文件（JPG, JPEG, PNG）');
        return;
      }
      
      setSelectedFile(file);
      setError(null);
      
      // 创建预览URL
      const objectUrl = URL.createObjectURL(file);
      setPreviewUrl(objectUrl);
      
      // 清除先前的结果
      setResult(null);
    }
  };
  
  // 处理作业类型变更
  const handleTypeChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    setHomeworkType(e.target.value);
  };
  
  // 处理拍照上传（仅在移动设备上支持）
  const handleCaptureClick = () => {
    if (fileInputRef.current) {
      fileInputRef.current.capture = 'environment';
      fileInputRef.current.click();
    }
  };
  
  // 提交作业图片进行分析
  const handleSubmit = async () => {
    if (!selectedFile) {
      setError('请先选择作业图片');
      return;
    }
    
    setIsLoading(true);
    setError(null);
    
    const formData = new FormData();
    formData.append('homework', selectedFile);
    formData.append('type', homeworkType);
    
    try {
      const response = await fetch(`${API_URL}/homework/mark`, {
        method: 'POST',
        body: formData,
      });
      
      if (!response.ok) {
        throw new Error(`服务器错误: ${response.status}`);
      }
      
      const data = await response.json();
      setResult(data);
    } catch (err: any) {
      setError(`处理失败: ${err.message}`);
    } finally {
      setIsLoading(false);
    }
  };
  
  const renderResultCard = (answer: any, index: number) => {
    const isExpanded = expandedCards[index] || false;
    const isCorrect = answer.isCorrect === true; // 明确检查布尔值
    
    // 简化版的卡片，只展示题号、学生答案和是否正确
    return (
      <div key={index} className="bg-white dark:bg-gray-800 rounded-lg shadow-sm p-3 mb-2">
        <div className="flex justify-between items-center">
          <h3 className="font-semibold text-md">题目 {answer.questionNumber}</h3>
          <div className="flex items-center space-x-2">
            {answer.isCorrect !== undefined && (
              <span className={isCorrect ? 'text-green-600 dark:text-green-400' : 'text-red-600 dark:text-red-400'}>
                {isCorrect ? '✓' : '✗'}
              </span>
            )}
            <button 
              onClick={() => toggleCardExpand(index)}
              className="text-blue-600 dark:text-blue-400 hover:text-blue-800 text-sm"
            >
              {isExpanded ? '收起' : '详情'}
            </button>
          </div>
        </div>
        
        {/* 简化的内容，始终显示 */}
        <div className="text-sm mt-1 text-gray-700 dark:text-gray-300 truncate">
          <span className="font-medium">答案:</span> {answer.studentAnswer}
        </div>
        
        {/* 详细内容，仅在展开时显示 */}
        {isExpanded && (
          <div className="mt-2 pt-2 border-t border-gray-200 dark:border-gray-700 text-sm">
            {answer.correctAnswer && (
              <div className="mb-1">
                <span className="font-medium dark:text-white">正确答案:</span> <span className="dark:text-gray-200">{answer.correctAnswer}</span>
              </div>
            )}
            
            {answer.correctSteps && (
              <div className="mb-1">
                <span className="font-medium dark:text-white">正确步骤:</span> <span className="dark:text-gray-200">{answer.correctSteps}</span>
              </div>
            )}
            
            {answer.evaluation && (
              <div className="mb-1">
                <span className="font-medium dark:text-white">评价:</span> <span className="dark:text-gray-200">{answer.evaluation}</span>
              </div>
            )}
            
            {answer.suggestion && (
              <div className="mb-1">
                <span className="font-medium dark:text-white">建议:</span> <span className="dark:text-gray-200">{answer.suggestion}</span>
              </div>
            )}
            
            {answer.explanation && (
              <div className="mb-1">
                <span className="font-medium dark:text-white">解释:</span> <span className="dark:text-gray-200">{answer.explanation}</span>
              </div>
            )}
          </div>
        )}
      </div>
    );
  };

  return (
    <div className="py-8 px-4 md:px-8">
      <div className="max-w-4xl mx-auto">
        <header className="mb-8">
          <Link href="/" className="text-blue-600 hover:text-blue-800 dark:text-blue-400 dark:hover:text-blue-300 mb-4 inline-flex items-center">
            <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5 mr-1" viewBox="0 0 20 20" fill="currentColor">
              <path fillRule="evenodd" d="M9.707 16.707a1 1 0 01-1.414 0l-6-6a1 1 0 010-1.414l6-6a1 1 0 011.414 1.414L5.414 9H17a1 1 0 110 2H5.414l4.293 4.293a1 1 0 010 1.414z" clipRule="evenodd" />
            </svg>
            返回首页
          </Link>
          <h1 className="text-3xl font-bold text-center mt-4 dark:text-white">作业批改</h1>
        </header>
        
        <div className="bg-white rounded-xl shadow-md overflow-hidden mb-8">
          <div className="p-6">
            <div className="mb-6">
              <label className="block text-gray-700 dark:text-gray-200 text-sm font-bold mb-2" htmlFor="homeworkType">
                选择作业类型
              </label>
              <select
                id="homeworkType"
                className="shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 dark:text-white dark:bg-gray-700 leading-tight focus:outline-none focus:shadow-outline"
                value={homeworkType}
                onChange={handleTypeChange}
              >
                <option value="english">英语</option>
                <option value="chinese">语文</option>
                <option value="math">数学</option>
                <option value="other">其他</option>
              </select>
            </div>
            
            <div className="mb-6">
              <div className="flex justify-center items-center flex-col">
                <input
                  type="file"
                  ref={fileInputRef}
                  onChange={handleFileChange}
                  accept="image/*"
                  className="hidden"
                />
                
                {!previewUrl ? (
                  <div 
                    className="border-2 border-dashed border-gray-300 dark:border-gray-600 rounded-lg p-8 text-center cursor-pointer bg-gray-50 dark:bg-gray-800 hover:bg-gray-100 dark:hover:bg-gray-700"
                    onClick={handleSelectClick}
                  >
                    <svg xmlns="http://www.w3.org/2000/svg" className="h-12 w-12 mx-auto text-gray-400 dark:text-gray-300" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1} d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
                    </svg>
                    <p className="mt-4 text-gray-600 dark:text-gray-300">点击选择图片或拖放图片</p>
                  </div>
                ) : (
                  <div className="relative w-full h-64 mb-4">
                    <Image
                      src={previewUrl}
                      alt="作业预览"
                      fill
                      className="object-contain rounded-lg"
                    />
                    <button
                      onClick={() => {
                        setPreviewUrl(null);
                        setSelectedFile(null);
                      }}
                      className="absolute top-2 right-2 bg-red-600 text-white rounded-full p-1 hover:bg-red-700"
                    >
                      <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
                        <path fillRule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clipRule="evenodd" />
                      </svg>
                    </button>
                  </div>
                )}
                
                <div className="flex gap-4 mt-4">
                  <button
                    onClick={handleSelectClick}
                    className="bg-blue-600 hover:bg-blue-700 text-white font-medium py-2 px-4 rounded focus:outline-none focus:shadow-outline"
                  >
                    选择图片
                  </button>
                  <button
                    onClick={handleCaptureClick}
                    className="bg-green-600 hover:bg-green-700 text-white font-medium py-2 px-4 rounded focus:outline-none focus:shadow-outline"
                  >
                    拍照上传
                  </button>
                </div>
              </div>
            </div>
            
            {error && (
              <div className="bg-red-100 border-l-4 border-red-500 text-red-700 p-4 mb-6" role="alert">
                <p>{error}</p>
              </div>
            )}
            
            <div className="flex justify-center">
              <button
                onClick={handleSubmit}
                disabled={!selectedFile || isLoading}
                className={`${
                  !selectedFile || isLoading
                    ? 'bg-gray-400 cursor-not-allowed'
                    : 'bg-blue-600 hover:bg-blue-700'
                } text-white font-bold py-3 px-6 rounded-lg shadow-lg transition-all duration-200 w-full md:w-auto`}
              >
                {isLoading ? (
                  <span className="flex items-center justify-center">
                    <svg className="animate-spin -ml-1 mr-3 h-5 w-5 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                      <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                      <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                    </svg>
                    处理中...
                  </span>
                ) : (
                  '批改作业'
                )}
              </button>
            </div>
          </div>
        </div>
        
        {result && result.success && (
          <div className="mt-8 p-4 bg-white dark:bg-gray-800 rounded-lg shadow-md">
            <div className="flex justify-between items-center mb-4">
              <h2 className="text-xl font-bold dark:text-white">批改结果</h2>
              {result.result.overallScore && (
                <div className="text-blue-700 dark:text-blue-300 font-bold">
                  总分: {result.result.overallScore}
                </div>
              )}
            </div>
            
            {result.result.feedback && (
              <p className="p-3 bg-gray-50 dark:bg-gray-700 rounded-lg dark:text-gray-200 mb-4 text-sm">
                {result.result.feedback}
              </p>
            )}
            
            {result.result.answers && (
              <div>
                <div className="flex justify-between items-center mb-2">
                  <h3 className="text-md font-semibold dark:text-white">题目分析</h3>
                  <button 
                    onClick={() => {
                      // 切换全部展开/折叠
                      const allExpanded = Object.values(expandedCards).every(v => v);
                      const newState = !allExpanded;
                      const newExpandedCards: {[key: number]: boolean} = {};
                      result.result.answers.forEach((answer: any, i: number) => {
                        newExpandedCards[i] = newState;
                      });
                      setExpandedCards(newExpandedCards);
                    }}
                    className="text-blue-600 dark:text-blue-400 text-sm"
                  >
                    {Object.values(expandedCards).every(v => v) ? '全部收起' : '全部展开'}
                  </button>
                </div>
                {result.result.answers.map((answer: any, index: number) => 
                  renderResultCard(answer, index)
                )}
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  );
} 