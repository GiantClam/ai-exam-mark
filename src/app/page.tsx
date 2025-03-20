'use client';

import { useState, useEffect } from 'react';
import Image from "next/image";
import Link from "next/link";

export default function Home() {
  const [apiStatus, setApiStatus] = useState<string>("检查中...");
  
  useEffect(() => {
    // 检查API连接状态
    async function checkApiStatus() {
      try {
        const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';
        const response = await fetch(`${API_URL}/health`);
        if (response.ok) {
          setApiStatus("已连接");
        } else {
          setApiStatus("不可用");
        }
      } catch (error) {
        setApiStatus("连接失败");
        console.error("API连接错误:", error);
      }
    }
    
    checkApiStatus();
  }, []);

  return (
    <main className="container mx-auto px-4 py-8">
      <section className="text-center mb-16">
        <h1 className="text-4xl font-bold text-center mb-8 dark:text-white">智能作业批改系统</h1>
        <p className="text-center mb-12 text-lg dark:text-gray-200">使用AI助力教师批改学生作业，支持英语、语文、数学等多种科目</p>
        
        <div className="flex justify-center">
          <Link href="/homework" className="bg-blue-600 hover:bg-blue-700 text-white font-semibold py-3 px-8 rounded-lg shadow-md transition duration-300 transform hover:scale-105">
            开始批改
          </Link>
        </div>
      </section>
      
      <section className="max-w-5xl mx-auto">
        <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
          <div className="bg-white dark:bg-gray-800 rounded-lg shadow-md p-6 text-center">
            <div className="flex justify-center mb-4">
              <div className="bg-blue-100 dark:bg-blue-900 p-3 rounded-full">
                <svg xmlns="http://www.w3.org/2000/svg" className="h-8 w-8 text-blue-600 dark:text-blue-300" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 9a2 2 0 012-2h.93a2 2 0 001.664-.89l.812-1.22A2 2 0 0110.07 4h3.86a2 2 0 011.664.89l.812 1.22A2 2 0 0018.07 7H19a2 2 0 012 2v9a2 2 0 01-2 2H5a2 2 0 01-2-2V9z" />
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 13a3 3 0 11-6 0 3 3 0 016 0z" />
                </svg>
              </div>
            </div>
            <h2 className="text-xl font-semibold mb-2 dark:text-white">拍照上传</h2>
            <p className="text-gray-600 dark:text-gray-300 text-center">直接拍摄学生作业或上传已有图片，系统快速识别</p>
          </div>
          
          <div className="bg-white dark:bg-gray-800 rounded-lg shadow-md p-6 text-center">
            <div className="flex justify-center mb-4">
              <div className="bg-purple-100 dark:bg-purple-900 p-3 rounded-full">
                <svg xmlns="http://www.w3.org/2000/svg" className="h-8 w-8 text-purple-600 dark:text-purple-300" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z" />
                </svg>
              </div>
            </div>
            <h2 className="text-xl font-semibold mb-2 dark:text-white">AI分析</h2>
            <p className="text-gray-600 dark:text-gray-300 text-center">利用Google Gemini 2.0 Thinking模型提取手写答案，精准识别括号、下划线等位置内容</p>
          </div>
          
          <div className="bg-white dark:bg-gray-800 rounded-lg shadow-md p-6 text-center">
            <div className="flex justify-center mb-4">
              <div className="bg-green-100 dark:bg-green-900 p-3 rounded-full">
                <svg xmlns="http://www.w3.org/2000/svg" className="h-8 w-8 text-green-600 dark:text-green-300" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
              </div>
            </div>
            <h2 className="text-xl font-semibold mb-2 dark:text-white">结果展示</h2>
            <p className="text-gray-600 dark:text-gray-300 text-center">智能评估答案正确性，提供详细批改意见和建议，帮助教师提高工作效率</p>
          </div>
        </div>
      </section>
      
      <section className="mt-16 text-center">
        <Link href="/homework" className="text-blue-600 dark:text-blue-400 hover:text-blue-800 dark:hover:text-blue-300 font-semibold text-lg">
          立即体验 →
        </Link>
        <div className="mt-4 text-sm text-gray-500 dark:text-gray-400">
          API状态: <span className={`font-medium ${
            apiStatus === "已连接" ? "text-green-600 dark:text-green-400" : 
            apiStatus === "检查中..." ? "text-yellow-600 dark:text-yellow-400" : 
            "text-red-600 dark:text-red-400"
          }`}>{apiStatus}</span>
        </div>
      </section>
    </main>
  );
}
