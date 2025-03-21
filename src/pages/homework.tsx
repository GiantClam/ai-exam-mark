import { useState, useRef } from 'react';
import Image from 'next/image';

// 格式化JSON显示组件
const FormattedJSON = ({ data }: { data: any }) => {
  // 格式化显示JSON数据的函数
  const formatJSON = (json: any) => {
    if (!json) return null;
    
    try {
      // 如果是字符串但看起来像JSON，先尝试解析
      let jsonObj = json;
      if (typeof json === 'string' && (json.trim().startsWith('{') || json.trim().startsWith('['))) {
        try {
          jsonObj = JSON.parse(json);
        } catch (e) {
          // 如果解析失败，保持原样
          return <pre className="whitespace-pre-wrap break-words">{json}</pre>;
        }
      }
      
      // 渲染JSON对象
      if (typeof jsonObj === 'object' && jsonObj !== null) {
        if (Array.isArray(jsonObj)) {
          // 处理数组
          return (
            <div className="pl-4 border-l-2 border-blue-200">
              {jsonObj.map((item, index) => (
                <div key={index} className="mb-2">
                  <span className="text-gray-500">[{index}]:</span> 
                  {typeof item === 'object' && item !== null ? (
                    <div className="ml-2">{formatJSON(item)}</div>
                  ) : (
                    <span className="text-blue-700">{JSON.stringify(item)}</span>
                  )}
                </div>
              ))}
            </div>
          );
        } else {
          // 处理对象
          return (
            <div className="pl-4 border-l-2 border-green-200">
              {Object.entries(jsonObj).map(([key, value], index) => (
                <div key={index} className="mb-2">
                  <span className="font-medium text-green-700">{key}:</span> 
                  {typeof value === 'object' && value !== null ? (
                    <div className="ml-2">{formatJSON(value)}</div>
                  ) : (
                    <span className="text-blue-700">{JSON.stringify(value)}</span>
                  )}
                </div>
              ))}
            </div>
          );
        }
      }
      
      // 处理基本类型
      return <span className="text-blue-700">{JSON.stringify(jsonObj)}</span>;
    } catch (error) {
      console.error("格式化JSON时出错:", error);
      return <pre className="text-red-600 whitespace-pre-wrap break-words">{String(json)}</pre>;
    }
  };

  return (
    <div className="bg-gray-50 p-6 rounded-md border border-gray-100 overflow-auto max-h-[600px] text-sm">
      {formatJSON(data)}
    </div>
  );
};

// 用户友好的作业答案结果展示组件
const HomeworkResult = ({ data, originalResponse }: { data: any, originalResponse?: any }) => {
  // 解析数据
  let parsedData = data;
  if (typeof data === 'string') {
    try {
      // 清理字符串，移除Markdown代码块标记和其他可能导致JSON解析失败的字符
      let cleanedJson = data.trim();
      
      // 检测并移除Markdown代码块格式 (```json ... ```)
      if (cleanedJson.startsWith('```')) {
        console.log("检测到Markdown代码块格式，正在清理...");
        // 移除开头的 ```json 或 ``` 标记
        const startIdx = cleanedJson.indexOf('\n');
        if (startIdx > 0) {
          cleanedJson = cleanedJson.substring(startIdx + 1);
        }
        
        // 移除结尾的 ``` 标记
        const endIdx = cleanedJson.lastIndexOf('```');
        if (endIdx > 0) {
          cleanedJson = cleanedJson.substring(0, endIdx).trim();
        }
      }
      
      // 应用其他清理规则
      cleanedJson = cleanedJson
        .replace(/[\u0000-\u001F\u007F-\u009F]/g, "") // 移除控制字符
        .replace(/\\(\r\n|\n|\r)/gm, "\\n"); // 规范化换行

      console.log("尝试解析JSON:", cleanedJson.substring(0, 100) + "...");
      parsedData = JSON.parse(cleanedJson);
    } catch (e) {
      console.error("JSON解析错误:", e);
      
      // 尝试使用传入的原始响应
      if (originalResponse && originalResponse.result) {
        try {
          const result = originalResponse.result;
          if (typeof result === 'object') {
            parsedData = result;
            console.log("使用原始响应对象", parsedData);
          } else if (typeof result === 'string') {
            let cleanedResult = result.trim();
            
            // 检测并移除Markdown代码块格式
            if (cleanedResult.startsWith('```')) {
              const startIdx = cleanedResult.indexOf('\n');
              if (startIdx > 0) {
                cleanedResult = cleanedResult.substring(startIdx + 1);
              }
              
              const endIdx = cleanedResult.lastIndexOf('```');
              if (endIdx > 0) {
                cleanedResult = cleanedResult.substring(0, endIdx).trim();
              }
            }
            
            parsedData = JSON.parse(cleanedResult);
            console.log("解析原始响应的result字符串");
          }
        } catch (innerErr) {
          console.error("无法解析原始响应:", innerErr);
          
          return (
            <div className="bg-yellow-50 p-4 rounded-md border border-yellow-100 mb-4">
              <p className="text-yellow-700">无法解析作业结果数据</p>
              <details className="mt-2">
                <summary className="cursor-pointer text-sm text-blue-600">查看原始数据</summary>
                <pre className="mt-2 text-xs text-gray-800 whitespace-pre-wrap overflow-auto max-h-[300px]">{data}</pre>
              </details>
              
              {/* 手动解析已知格式的JSON字符串 */}
              <div className="mt-4 p-4 bg-white rounded border border-gray-200">
                <button 
                  onClick={() => {
                    // 尝试提取并直接使用清理后的数据
                    let cleanedData = data;
                    if (typeof data === 'string' && data.includes('```')) {
                      const startIdx = data.indexOf('```') + 3;
                      const languageEndIdx = data.indexOf('\n', startIdx);
                      const contentStartIdx = languageEndIdx + 1;
                      const contentEndIdx = data.lastIndexOf('```');
                      
                      if (contentStartIdx < contentEndIdx) {
                        cleanedData = data.substring(contentStartIdx, contentEndIdx).trim();
                      }
                    }
                    
                    // 复制清理后的数据到剪贴板
                    navigator.clipboard.writeText(cleanedData);
                    alert('已复制清理后的数据到剪贴板!');
                  }}
                  className="mb-2 px-2 py-1 bg-blue-50 text-blue-600 text-sm rounded"
                >
                  提取并复制JSON内容
                </button>
              </div>
            </div>
          );
        }
      } else {
        return (
          <div className="bg-yellow-50 p-4 rounded-md border border-yellow-100 mb-4">
            <p className="text-yellow-700">无法解析作业结果数据</p>
            <details className="mt-2">
              <summary className="cursor-pointer text-sm text-blue-600">查看原始数据</summary>
              <pre className="mt-2 text-xs text-gray-800 whitespace-pre-wrap overflow-auto max-h-[300px]">{data}</pre>
            </details>
          </div>
        );
      }
    }
  }

  // 检查是否有answers数组
  if (!parsedData || !parsedData.answers || !Array.isArray(parsedData.answers)) {
    return (
      <div className="bg-gray-50 p-6 rounded-md border border-gray-100">
        <p className="text-gray-800 font-medium mb-2">未检测到答案数据</p>
        <pre className="text-xs text-gray-600 whitespace-pre-wrap overflow-auto max-h-[400px]">
          {JSON.stringify(parsedData, null, 2)}
        </pre>
      </div>
    );
  }

  // 分组答案 - 按照题号分组
  const groupedAnswers: Record<string, any[]> = {};
  parsedData.answers.forEach((answer: any) => {
    const key = answer.questionNumber || '未知题号';
    if (!groupedAnswers[key]) {
      groupedAnswers[key] = [];
    }
    groupedAnswers[key].push(answer);
  });

  return (
    <div className="space-y-6">
      {/* 总体评价 */}
      {parsedData.feedback && (
        <div className="bg-blue-50 p-4 rounded-md border border-blue-100">
          <h3 className="text-lg font-medium text-blue-800 mb-2">总体评价</h3>
          <p className="text-blue-700">{parsedData.feedback}</p>
        </div>
      )}
      
      {/* 总分 */}
      {parsedData.overallScore && (
        <div className="bg-green-50 p-4 rounded-md border border-green-100">
          <h3 className="text-lg font-medium text-green-800">总分: <span className="text-2xl">{parsedData.overallScore}</span></h3>
        </div>
      )}

      {/* 题目和答案 */}
      <div className="bg-white rounded-lg border border-gray-200 overflow-hidden">
        <div className="grid grid-cols-12 bg-gray-100 p-3 font-medium text-gray-700 border-b border-gray-200">
          <div className="col-span-2">题号</div>
          <div className="col-span-4">答案</div>
          <div className="col-span-6">评价</div>
        </div>
        
        <div className="divide-y divide-gray-100">
          {Object.entries(groupedAnswers).map(([questionNumber, answers]) => 
            answers.map((answer, index) => (
              <div key={`${questionNumber}-${index}`} className="grid grid-cols-12 p-3 hover:bg-gray-50">
                <div className="col-span-2 font-medium text-gray-700">{answer.questionNumber}</div>
                <div className="col-span-4 text-gray-900">{answer.studentAnswer}</div>
                <div className="col-span-6">
                  {answer.isCorrect !== undefined && (
                    <span className={`inline-block px-2 py-1 text-xs font-medium rounded ${answer.isCorrect ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'} mr-2`}>
                      {answer.isCorrect ? '正确' : '错误'}
                    </span>
                  )}
                  {answer.evaluation && <span className="text-gray-700">{answer.evaluation}</span>}
                  {answer.correctAnswer && !answer.isCorrect && (
                    <div className="mt-1 text-red-600 text-sm">
                      正确答案: {answer.correctAnswer}
                    </div>
                  )}
                  {answer.explanation && (
                    <div className="mt-1 text-gray-600 text-sm">
                      {answer.explanation}
                    </div>
                  )}
                  {answer.suggestion && (
                    <div className="mt-1 text-blue-600 text-sm">
                      建议: {answer.suggestion}
                    </div>
                  )}
                </div>
              </div>
            ))
          )}
        </div>
      </div>
      
      {/* 调试按钮 - 查看原始JSON */}
      <div className="mt-4 flex justify-end">
        <details className="text-sm text-gray-500">
          <summary className="cursor-pointer hover:text-gray-700">查看原始数据</summary>
          <div className="mt-2 bg-gray-50 p-4 rounded border border-gray-200 overflow-auto max-h-[300px]">
            <pre className="text-xs">{JSON.stringify(parsedData, null, 2)}</pre>
          </div>
        </details>
      </div>
    </div>
  );
};

export default function HomeworkPage() {
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [previewUrl, setPreviewUrl] = useState<string | null>(null);
  const [isAnalyzing, setIsAnalyzing] = useState<boolean>(false);
  const [extractedText, setExtractedText] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [rawJsonResponse, setRawJsonResponse] = useState<any>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);
  
  // 作业类型和布局类型
  const [homeworkType, setHomeworkType] = useState<string>("general");
  const [layoutType, setLayoutType] = useState<string>("single");
  
  // 作业类型选项
  const homeworkTypes = [
    { value: "general", label: "通用" },
    { value: "english", label: "英语" },
    { value: "math", label: "数学" },
    { value: "chinese", label: "语文" }
  ];
  
  // 布局类型选项
  const layoutTypes = [
    { value: "single", label: "普通作业 (单栏)" },
    { value: "double", label: "试卷 (双栏，先左后右)" }
  ];
  
  // 获取API URL - 使用作业批改大模型API
  const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8280/api';
  
  // 处理文件选择
  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const files = e.target.files;
    if (files && files.length > 0) {
      const file = files[0];
      setSelectedFile(file);
      
      // 创建预览URL
      const objectUrl = URL.createObjectURL(file);
      setPreviewUrl(objectUrl);
      
      // 重置提取结果和错误
      setExtractedText(null);
      setError(null);
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
  
  // 调用大模型API直接提取答案
  const handleExtractText = async () => {
    if (!selectedFile) return;
    
    setIsAnalyzing(true);
    setError(null);
    setRawJsonResponse(null);
    
    // 创建FormData对象用于传输文件
    const formData = new FormData();
    formData.append('homework', selectedFile); // 使用'homework'作为后端期望的键名
    formData.append('type', homeworkType); // 传递选定的作业类型
    formData.append('layout', layoutType); // 传递选定的布局类型
    
    try {
      // 调用后端API
      const response = await fetch(`${API_URL}/homework/mark`, {
        method: 'POST',
        body: formData,
      });
      
      if (!response.ok) {
        throw new Error(`服务器响应错误: ${response.status}`);
      }
      
      const data = await response.json();
      
      // 调试输出，查看实际返回的数据格式
      console.log('API响应数据:', data);
      
      // 保存原始JSON响应
      setRawJsonResponse(data);
      
      // 处理不同的响应格式情况
      if (data.success) {
        // 标准成功响应
        if (data.result) {
          if (typeof data.result === 'string') {
            setExtractedText(data.result);
          } else {
            // 使用JSON.stringify保持格式但不再显示在文本区域
            setExtractedText(JSON.stringify(data.result, null, 2));
          }
        } else if (data.data) {
          // 兼容 {success: true, data: ...} 格式
          if (typeof data.data === 'string') {
            setExtractedText(data.data);
          } else {
            setExtractedText(JSON.stringify(data.data, null, 2));
          }
        } else if (data.text) {
          // 兼容 {success: true, text: ...} 格式
          setExtractedText(data.text);
        } else if (data.answer) {
          // 兼容 {success: true, answer: ...} 格式
          setExtractedText(data.answer);
        } else if (data.content) {
          // 兼容 {success: true, content: ...} 格式
          setExtractedText(data.content);
        } else {
          // 如果找不到已知的数据字段，尝试使用整个响应对象
          const possibleResultKeys = Object.keys(data).filter(key => 
            key !== 'success' && key !== 'error' && key !== 'message'
          );
          
          if (possibleResultKeys.length > 0) {
            const mainKey = possibleResultKeys[0];
            const mainValue = data[mainKey];
            
            if (typeof mainValue === 'string') {
              setExtractedText(mainValue);
            } else {
              setExtractedText(JSON.stringify(mainValue, null, 2));
            }
          } else {
            // API返回了成功但没有任何内容，显示一个友好的消息
            setExtractedText("提交成功，但服务器未返回答案内容。这可能是因为:\n1. 图片质量不足以识别\n2. 作业内容格式不受支持\n3. 后端处理逻辑异常\n\n请尝试上传更清晰的图片或联系管理员检查后端服务。");
          }
        }
      } else {
        throw new Error(data.error || data.message || '答案提取失败');
      }
    } catch (err) {
      console.error('提取答案时出错:', err);
      setError(err instanceof Error ? err.message : '提取答案时出错');
      setExtractedText(null);
    } finally {
      setIsAnalyzing(false);
    }
  };

  return (
    <div className="container mx-auto py-12 px-4">
      <div className="max-w-4xl mx-auto">
        <h1 className="text-3xl font-bold mb-8 text-gray-900">上传作业</h1>
        
        <div className="bg-white rounded-lg shadow-sm border border-gray-100 p-8 mb-8">
          <div className="mb-8">
            <h2 className="text-xl font-semibold mb-4 text-gray-900">选择作业图片</h2>
            <p className="text-gray-800 mb-6">
              上传清晰的作业照片，确保手写内容完整可见，以获得最佳的答案提取效果
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
          
          {/* 作业设置选项 */}
          {previewUrl && (
            <div className="mb-8">
              <h3 className="text-lg font-semibold mb-4 text-gray-800">作业设置</h3>
              
              <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                {/* 作业类型选择 */}
                <div>
                  <label htmlFor="homeworkType" className="block text-sm font-medium text-gray-700 mb-2">
                    作业类型
                  </label>
                  <select
                    id="homeworkType"
                    value={homeworkType}
                    onChange={(e) => setHomeworkType(e.target.value)}
                    className="w-full px-4 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                  >
                    {homeworkTypes.map((type) => (
                      <option key={type.value} value={type.value}>
                        {type.label}
                      </option>
                    ))}
                  </select>
                  <p className="mt-1 text-sm text-gray-500">
                    选择作业类型以获得更精确的批改结果
                  </p>
                </div>
                
                {/* 布局类型选择 */}
                <div>
                  <label htmlFor="layoutType" className="block text-sm font-medium text-gray-700 mb-2">
                    布局类型
                  </label>
                  <select
                    id="layoutType"
                    value={layoutType}
                    onChange={(e) => setLayoutType(e.target.value)}
                    className="w-full px-4 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                  >
                    {layoutTypes.map((type) => (
                      <option key={type.value} value={type.value}>
                        {type.label}
                      </option>
                    ))}
                  </select>
                  <p className="mt-1 text-sm text-gray-500">
                    {layoutType === 'single' ? '单栏布局适合普通作业' : '双栏布局适合试卷，会先处理左侧再处理右侧'}
                  </p>
                </div>
              </div>
              
              {/* 重置按钮 */}
              <div className="mt-6 flex justify-end">
                <button
                  onClick={() => {
                    setPreviewUrl(null);
                    setSelectedFile(null);
                    setHomeworkType("general");
                    setLayoutType("single");
                    setExtractedText(null);
                    setError(null);
                    setRawJsonResponse(null);
                  }}
                  className="px-4 py-2 text-sm text-gray-600 hover:text-gray-800 flex items-center"
                >
                  <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4 mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
                  </svg>
                  重置全部
                </button>
              </div>
            </div>
          )}
          
          {error && (
            <div className="mb-6 p-4 bg-red-50 border border-red-100 rounded-md text-red-700">
              {error}
            </div>
          )}
          
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
              ) : '提取答案'}
            </button>
          </div>
        </div>
        
        {extractedText && (
          <div className="bg-white rounded-lg shadow-sm border border-gray-100 p-8">
            <h2 className="text-xl font-semibold mb-4 text-gray-900">提取结果</h2>
            
            {/* 开发调试信息 */}
            <div className="mb-4 p-2 bg-gray-100 rounded text-xs">
              <p>结果类型: {typeof extractedText}</p>
              <p>是否以{'{'} 开头: {String(extractedText.trim().startsWith('{'))}</p>
              <p>是否以 [ 开头: {String(extractedText.trim().startsWith('['))}</p>
              <p>rawJsonResponse: {rawJsonResponse ? '有值' : '无值'}</p>
              {rawJsonResponse && <p>rawJsonResponse.result类型: {typeof rawJsonResponse.result}</p>}
            </div>
            
            {/* 尝试强制解析JSON并渲染HomeworkResult组件 */}
            <div className="mb-6">
              <HomeworkResult data={extractedText} originalResponse={rawJsonResponse} />
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