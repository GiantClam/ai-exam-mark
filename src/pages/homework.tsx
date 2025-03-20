import { useState, useRef } from 'react';
import Image from 'next/image';

export default function HomeworkPage() {
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [previewUrl, setPreviewUrl] = useState<string | null>(null);
  const [isAnalyzing, setIsAnalyzing] = useState<boolean>(false);
  const [extractedText, setExtractedText] = useState<string | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);
  
  // 处理文件选择
  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const files = e.target.files;
    if (files && files.length > 0) {
      const file = files[0];
      setSelectedFile(file);
      
      // 创建预览URL
      const objectUrl = URL.createObjectURL(file);
      setPreviewUrl(objectUrl);
      
      // 重置提取结果
      setExtractedText(null);
    }
  };
  
  // 触发文件选择对话框
  const handleSelectClick = () => {
    if (fileInputRef.current) {
      fileInputRef.current.click();
    }
  };
  
  // 处理拍照上传
  const handleCaptureClick = () => {
    if (fileInputRef.current) {
      fileInputRef.current.capture = 'environment';
      fileInputRef.current.click();
    }
  };
  
  // 模拟提取文字过程
  const handleExtractText = () => {
    if (!selectedFile) return;
    
    setIsAnalyzing(true);
    
    // 模拟API调用延迟
    setTimeout(() => {
      setExtractedText("这是从作业图片中提取的示例文字内容。\n\n在实际应用中，这里将显示AI从上传图片中识别出的手写文字内容。系统能够识别不同的字体和书写风格，并尽可能保持原始格式。\n\n提取后的文字可以被复制、编辑或保存，方便教师批改和学生复习使用。");
      setIsAnalyzing(false);
    }, 2000);
  };

  return (
    <div className="container mx-auto py-12 px-4">
      <div className="max-w-4xl mx-auto">
        <h1 className="text-3xl font-bold mb-8 text-gray-900">上传作业</h1>
        
        <div className="bg-white rounded-lg shadow-sm border border-gray-100 p-8 mb-8">
          <div className="mb-8">
            <h2 className="text-xl font-semibold mb-4 text-gray-900">选择作业图片</h2>
            <p className="text-gray-800 mb-6">
              上传清晰的作业照片，确保手写内容完整可见，以获得最佳的文字识别结果
            </p>
            
            <input
              type="file"
              ref={fileInputRef}
              onChange={handleFileChange}
              accept="image/*"
              className="hidden"
            />
            
            {!previewUrl ? (
              <div 
                onClick={handleSelectClick}
                className="border-2 border-dashed border-gray-200 rounded-lg p-8 text-center cursor-pointer hover:bg-gray-50 transition-colors"
              >
                <div className="flex flex-col items-center justify-center">
                  <svg xmlns="http://www.w3.org/2000/svg" className="h-16 w-16 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1} d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
                  </svg>
                  <p className="mt-4 text-gray-700">点击选择或拖放图片</p>
                </div>
              </div>
            ) : (
              <div className="mb-6">
                <div className="relative w-full h-[300px] rounded-lg overflow-hidden mb-4 border border-gray-100">
                  <Image
                    src={previewUrl}
                    alt="作业预览"
                    fill
                    className="object-contain"
                  />
                  <button
                    onClick={() => {
                      setPreviewUrl(null);
                      setSelectedFile(null);
                    }}
                    className="absolute top-2 right-2 bg-white p-2 rounded-full shadow-sm hover:bg-gray-100"
                  >
                    <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5 text-gray-600" viewBox="0 0 20 20" fill="currentColor">
                      <path fillRule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clipRule="evenodd" />
                    </svg>
                  </button>
                </div>
              </div>
            )}
            
            <div className="flex gap-4 mt-6 justify-center">
              <button
                onClick={handleSelectClick}
                className="px-5 py-2.5 bg-blue-50 text-blue-600 font-medium rounded-md hover:bg-blue-100 transition-colors"
              >
                选择图片
              </button>
              <button
                onClick={handleCaptureClick}
                className="px-5 py-2.5 bg-blue-600 text-white font-medium rounded-md hover:bg-blue-700 transition-colors"
              >
                拍照上传
              </button>
            </div>
          </div>
          
          <div className="flex justify-center">
            <button
              onClick={handleExtractText}
              disabled={!selectedFile || isAnalyzing}
              className={`${
                !selectedFile || isAnalyzing
                  ? 'bg-gray-200 text-gray-500 cursor-not-allowed'
                  : 'bg-blue-600 text-white hover:bg-blue-700'
              } font-medium px-8 py-3 rounded-md transition-colors w-full md:w-auto`}
            >
              {isAnalyzing ? (
                <span className="flex items-center justify-center">
                  <svg className="animate-spin -ml-1 mr-3 h-5 w-5" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                    <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                    <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                  </svg>
                  正在识别...
                </span>
              ) : '提取文字'}
            </button>
          </div>
        </div>
        
        {extractedText && (
          <div className="bg-white rounded-lg shadow-sm border border-gray-100 p-8">
            <h2 className="text-xl font-semibold mb-4 text-gray-900">提取结果</h2>
            
            <div className="bg-gray-50 p-6 rounded-md border border-gray-100 mb-6">
              <p className="text-gray-800 whitespace-pre-line">
                {extractedText}
              </p>
            </div>
            
            <div className="flex justify-end space-x-4">
              <button 
                onClick={() => navigator.clipboard.writeText(extractedText)}
                className="text-blue-600 hover:text-blue-800 font-medium flex items-center"
              >
                <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5 mr-1" viewBox="0 0 20 20" fill="currentColor">
                  <path d="M8 3a1 1 0 011-1h2a1 1 0 110 2H9a1 1 0 01-1-1z" />
                  <path d="M6 3a2 2 0 00-2 2v11a2 2 0 002 2h8a2 2 0 002-2V5a2 2 0 00-2-2 3 3 0 01-3 3H9a3 3 0 01-3-3z" />
                </svg>
                复制内容
              </button>
              
              <button className="text-blue-600 hover:text-blue-800 font-medium flex items-center">
                <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5 mr-1" viewBox="0 0 20 20" fill="currentColor">
                  <path fillRule="evenodd" d="M3 17a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1zm3.293-7.707a1 1 0 011.414 0L9 10.586V3a1 1 0 112 0v7.586l1.293-1.293a1 1 0 111.414 1.414l-3 3a1 1 0 01-1.414 0l-3-3a1 1 0 010-1.414z" clipRule="evenodd" />
                </svg>
                下载文本
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
} 