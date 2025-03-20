import Link from 'next/link';

export default function Results() {
  return (
    <div className="container mx-auto py-12 px-4">
      <div className="max-w-4xl mx-auto">
        <h1 className="text-3xl font-bold mb-8 text-gray-900">提取结果</h1>
        
        <div className="bg-white p-8 rounded-lg shadow-sm border border-gray-100 mb-8">
          <div className="flex items-center justify-between mb-6">
            <h2 className="text-xl font-semibold text-gray-900">作业1</h2>
            <span className="text-sm text-gray-700">2023-04-10 10:30</span>
          </div>
          
          <div className="mb-4 p-4 bg-gray-50 rounded border border-gray-100">
            <p className="text-gray-800 whitespace-pre-line">
              这是一个示例提取结果。在这里您可以看到从作业图片中提取的文字内容。
              系统会保留原始的段落格式，以便您更好地查看和使用提取的内容。
            </p>
          </div>
          
          <div className="flex justify-end space-x-3">
            <button className="text-blue-600 hover:text-blue-800 font-medium">
              复制内容
            </button>
            <button className="text-blue-600 hover:text-blue-800 font-medium">
              下载文本
            </button>
          </div>
        </div>
        
        <div className="text-center">
          <p className="text-gray-700 mb-4">还没有更多的提取结果</p>
          <Link 
            href="/homework"
            className="bg-blue-600 hover:bg-blue-700 text-white font-medium px-6 py-2 rounded-md transition-colors inline-block"
          >
            上传新作业
          </Link>
        </div>
      </div>
    </div>
  );
} 