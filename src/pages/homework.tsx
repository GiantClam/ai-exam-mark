import React, { useState, useEffect, useRef, useCallback, useMemo } from 'react';
import { Upload, Form, Select, InputNumber, Button, message, Card, Typography, Spin, Divider, Result, Tag, List, Collapse, Progress, Modal, Input, Switch } from 'antd';
import { InboxOutlined, CheckCircleOutlined, DownloadOutlined, EditOutlined, FilePdfOutlined, EyeOutlined } from '@ant-design/icons';
import type { UploadProps, UploadFile } from 'antd';
import axios from 'axios';
import * as XLSX from 'xlsx';
import { Document, Page, pdfjs } from 'react-pdf';
import 'react-pdf/dist/esm/Page/AnnotationLayer.css';
import 'react-pdf/dist/esm/Page/TextLayer.css';

// 使用 Worker 静态引入方式
pdfjs.GlobalWorkerOptions.workerSrc = '/pdf.worker.min.js';

// Log the versions to help with debugging
console.log(`[PDF] PDF.js version: ${pdfjs.version}`);

const { Dragger } = Upload;
const { Title, Paragraph, Text } = Typography;
const { Panel } = Collapse;

interface Answer {
  questionNumber?: string;
  studentAnswer: string;
  isCorrect?: boolean;
  correctAnswer?: string;
  correctSteps?: string;
  explanation?: string;
  evaluation?: string;
}

interface HomeworkResult {
  answers?: Answer[];
  feedback?: string;
  overallScore?: string;
}

interface FormValues {
  type: string;
  layout: string;
  pagesPerStudent: number;
}

interface AnswerDetail {
  questionNumber: string;
  studentAnswer: string;
  isCorrect: boolean;
  correctAnswer: string;
  explanation?: string;
}

interface StudentResult {
  name: string;
  class?: string;
  score: string;
  feedback: string;
  answers?: AnswerDetail[];
  totalQuestions: number;
  correctCount: number;
  pdfUrl?: string;
}

const Homework: React.FC = () => {
  const [form] = Form.useForm<FormValues>();
  const [fileList, setFileList] = useState<UploadFile[]>([]);
  const [loading, setLoading] = useState<boolean>(false);
  const [result, setResult] = useState<string | null>(null);
  const [displayResults, setDisplayResults] = useState<boolean>(false);
  const [activeKey, setActiveKey] = useState<string | string[]>([]);
  const [studentResults, setStudentResults] = useState<StudentResult[]>([]);
  
  // 添加任务相关状态
  const [taskId, setTaskId] = useState<string | null>(null);
  const [taskStatus, setTaskStatus] = useState<string | null>(null);
  const [processedCount, setProcessedCount] = useState(0);
  const [totalStudents, setTotalStudents] = useState(0);
  const [taskProgress, setTaskProgress] = useState(0);
  const [pollingInterval, setPollingInterval] = useState<NodeJS.Timeout | null>(null);

  // 添加最大轮询次数状态
  const [pollCount, setPollCount] = useState(0);
  const MAX_POLL_COUNT = 60; // 最多轮询60次，约2分钟

  // 添加错误计数状态
  const [errorCount, setErrorCount] = useState(0);

  // 使用useRef存储轮询定时器，确保引用稳定
  const pollingIntervalRef = useRef<NodeJS.Timeout | null>(null);

  // Add state for edit modal
  const [isEditModalVisible, setIsEditModalVisible] = useState(false);
  const [editingStudentIndex, setEditingStudentIndex] = useState<number | null>(null);
  const [editingAnswerIndex, setEditingAnswerIndex] = useState<number | null>(null);
  const [editForm] = Form.useForm();

  // Add state for student info edit modal
  const [isStudentEditModalVisible, setIsStudentEditModalVisible] = useState(false);
  const [studentEditForm] = Form.useForm();

  // Add state for master answers modal
  const [isMasterAnswersModalVisible, setIsMasterAnswersModalVisible] = useState(false);
  const [masterAnswersForm] = Form.useForm();

  // Add state for PDF viewer
  const [isPdfModalVisible, setIsPdfModalVisible] = useState(false);
  const [currentPdfUrl, setCurrentPdfUrl] = useState<string | null>(null);
  const [numPages, setNumPages] = useState<number | null>(null);
  const [currentPage, setCurrentPage] = useState(1);

  // Add scale state for PDF viewer
  const [pdfScale, setPdfScale] = useState<number>(1.0);

  // Add state for side-by-side view
  const [showSideBySide, setShowSideBySide] = useState<number | null>(null);

  // 使用useMemo记忆化PDF选项
  const pdfOptions = useMemo(() => ({
    cMapUrl: 'https://cdn.jsdelivr.net/npm/pdfjs-dist@3.11.174/cmaps/',
    cMapPacked: true,
    standardFontDataUrl: 'https://cdn.jsdelivr.net/npm/pdfjs-dist@3.11.174/standard_fonts/'
  }), []); // 空依赖数组表示这个对象只会在组件首次渲染时创建

  // 异步轮询任务状态
  useEffect(() => {
    console.log('组件初始化，设置清理函数');
    // 清理函数，确保组件卸载时停止轮询
    return () => {
      console.log('组件卸载，执行清理函数');
      if (pollingIntervalRef.current) {
        console.log(`组件卸载时清除定时器`);
        clearInterval(pollingIntervalRef.current);
        pollingIntervalRef.current = null;
      }
    };
  }, []);
  
  // 当taskId变化时开始轮询，添加更详细的日志
  useEffect(() => {
    if (taskId && taskId.trim() !== '') {
      console.log(`检测到有效的任务ID: ${taskId}，开始轮询`);
      setLoading(true); // 确保加载状态设置为true
      startPolling();
    } else if (taskId === null || taskId === '') {
      console.log('任务ID为空或无效，停止轮询');
      stopPolling();
    }
    
    // 不要在清理函数中停止轮询，避免在taskId变化时中断新设置的轮询
    return () => {
      console.log('taskId依赖项变化，但不停止轮询，由新的useEffect来管理');
      // 不调用stopPolling()，避免中断新开始的轮询
    };
  }, [taskId]);

  // 修改状态更新逻辑，在轮询过程中恢复时重置错误计数
  useEffect(() => {
    // 当成功获取一次任务状态后，重置错误计数
    if (taskStatus && errorCount > 0) {
      setErrorCount(0);
      console.log('成功恢复连接，重置错误计数');
    }
  }, [taskStatus]);

  // 在errorCount变化时添加提示
  useEffect(() => {
    if (errorCount === 1) {
      message.warning('查询状态出现问题，系统将自动重试', 3);
    } else if (errorCount > 1 && errorCount < 3) {
      message.warning(`连续 ${errorCount} 次查询失败，再失败 ${3-errorCount} 次将重置任务`, 3);
    }
  }, [errorCount]);

  // 在轮询开始时和轮询状态变化时，额外检查定时器状态
  // useEffect(() => {
  //   console.log(`轮询状态检查: 定时器${pollingInterval ? '正在运行' : '未运行'}, ref定时器${pollingIntervalRef.current ? '存在' : '不存在'}`);
    
  //   // 确保ref和state保持同步
  //   if (pollingInterval && !pollingIntervalRef.current) {
  //     pollingIntervalRef.current = pollingInterval;
  //   } else if (!pollingInterval && pollingIntervalRef.current) {
  //     // 如果state为null但ref不为null，修复这种不一致
  //     clearInterval(pollingIntervalRef.current);
  //     pollingIntervalRef.current = null;
  //     console.log('修复了定时器状态不一致');
  //   }
  // }, [pollingInterval, taskId]);

  // 开始轮询任务状态，添加更详细的日志
  const startPolling = () => {
    // 检查是否有有效的任务ID
    if (!taskId) {
      console.log('没有有效的任务ID，不开始轮询');
      return;
    }
    
    console.log(`准备开始轮询任务状态: ${taskId}`);
    
    // 确保先停止现有轮询
    if (pollingIntervalRef.current) {
      console.log('发现已存在的定时器，先清除');
      clearInterval(pollingIntervalRef.current);
      pollingIntervalRef.current = null;
      setPollingInterval(null);
    }
    
    // 重置轮询计数
    setPollCount(0);
    setErrorCount(0); // 重置错误计数
    
    // 立即执行第一次查询，不等待定时器触发
    console.log('立即执行第一次查询，不等待定时器');
    pollTaskStatus();
    
    console.log(`设置新的轮询间隔，每10秒查询一次任务: ${taskId}`);
    const interval = setInterval(() => {
      setPollCount(prev => {
        const newCount = prev + 1;
        console.log(`定时器触发，执行pollTaskStatus() - 任务ID: ${taskId}, 轮询次数: ${newCount}`);
        return newCount;
      });
      pollTaskStatus();
    }, 10000); // 每10秒轮询一次
    
    // 同时更新ref和state
    pollingIntervalRef.current = interval;
    setPollingInterval(interval);
    console.log(`设置轮询间隔成功，定时器ID已保存`);
  };

  // 停止轮询
  const stopPolling = () => {
    console.log(`停止轮询，当前定时器ID: ${pollingIntervalRef.current ? '存在' : '不存在'}`);
    if (pollingIntervalRef.current) {
      try {
        clearInterval(pollingIntervalRef.current);
        console.log('已清除轮询定时器');
      } catch (err) {
        console.error('清除定时器时出错:', err);
      }
      pollingIntervalRef.current = null;
      setPollingInterval(null);
      console.log('轮询已停止，但保留任务ID');
    } else {
      console.log('无需清除定时器，轮询定时器已为null');
    }
  };

  // 轮询任务状态
  const pollTaskStatus = useCallback(() => {
    // 检查当前taskId
    if (!taskId) {
      console.log('当前无任务ID，不执行轮询');
      return;
    }
    
    // 确保不重复创建定时器
    if (pollingIntervalRef.current === null) {
      console.log('轮询函数被直接调用，而不是通过定时器');
    }
    
    console.log('开始轮询任务状态:', taskId);
    
    const poll = async () => {
      try {
        // 再次检查taskId是否有效（可能在异步操作过程中被清空）
        if (!taskId) {
          console.log('任务ID已被清空，取消本次轮询');
          return;
        }
        
        console.log(`发送轮询请求: /api/tasks/${taskId}`);
        const response = await axios.get(`/api/tasks/${taskId}`);
        console.log('轮询结果:', response.data);
        
        // 检查API响应格式并适配处理
        if (response.data.success === undefined) {
          // 旧格式API响应处理
          const taskData = response.data;
          
          // 更新任务状态
          setTaskStatus(taskData.status);
          setProcessedCount(taskData.processed || 0);
          setTotalStudents(taskData.total_students || 0);
          
          // 计算进度
          if (taskData.total_students > 0) {
            const progress = (taskData.processed || 0) / taskData.total_students;
            setTaskProgress(progress);
          }
          
          // 检查任务是否完成
          if (taskData.status === 'completed') {
            console.log('任务完成:', taskData);
            stopPolling();
            setLoading(false);
            
            // 处理结果
            if (taskData.results) {
              console.log('处理任务结果');
              processResults(taskData.results);
            } else {
              console.error('任务完成但没有结果');
              message.warning('任务完成但没有返回结果');
              // 即使没有结果，也重置任务状态
              resetTaskState();
            }
            
            return;
          } else if (taskData.status === 'failed') {
            console.error('任务失败:', taskData.error);
            stopPolling();
            setLoading(false);
            message.error(`处理失败: ${taskData.error || '未知错误'}`);
            // 任务失败，重置任务状态
            resetTaskState();
            return;
          }
        } else if (!response.data.success) {
          // 新格式失败响应处理
          console.error('获取任务状态失败:', response.data.error);
          setErrorCount(prev => prev + 1);
          
          // 如果错误次数超过阈值，停止轮询
          if (errorCount >= MAX_POLL_COUNT - 1) {
            console.error('达到最大错误次数，停止轮询');
            stopPolling();
            setLoading(false);
            message.error('无法获取任务状态，请稍后重试');
            // 达到最大错误次数，重置任务状态
            resetTaskState();
            return;
          }
        } else {
          // 新格式成功响应处理
          // 重置错误计数
          setErrorCount(0);
          
          // 更新任务状态
          const taskData = response.data.data;
          setTaskStatus(taskData.status);
          setProcessedCount(taskData.processed_count || 0);
          setTotalStudents(taskData.total_students || 0);
          
          // 计算进度
          if (taskData.total_students > 0) {
            const progress = taskData.processed_count / taskData.total_students;
            setTaskProgress(progress);
          }
          
          // 检查任务是否完成
          if (taskData.status === 'completed') {
            console.log('任务完成:', taskData);
            stopPolling();
            setLoading(false);
            
            // 处理结果
            if (taskData.results) {
              console.log('处理任务结果');
              processResults(taskData.results);
            } else {
              console.error('任务完成但没有结果');
              message.warning('任务完成但没有返回结果');
              // 即使没有结果，也重置任务状态
              resetTaskState();
            }
            
            return;
          } else if (taskData.status === 'failed') {
            console.error('任务失败:', taskData.error);
            stopPolling();
            setLoading(false);
            message.error(`处理失败: ${taskData.error || '未知错误'}`);
            // 任务失败，重置任务状态
            resetTaskState();
            return;
          }
        }
      } catch (error) {
        console.error('轮询错误:', error);
        setErrorCount(prev => prev + 1);
        
        // 如果错误次数超过阈值，停止轮询
        if (errorCount >= MAX_POLL_COUNT - 1) {
          console.error('达到最大错误次数，停止轮询');
          stopPolling();
          setLoading(false);
          message.error('网络错误，无法获取任务状态');
          // 网络错误导致的停止轮询，也重置任务状态
          resetTaskState();
          return;
        }
      }
    };
    
    poll();
  }, [taskId, errorCount]);

  // 解析处理结果
  const processResults = (result: any): void => {
    try {
      if (!result) {
        console.error('处理结果为空');
        return;
      }
      
      console.log('开始处理结果类型:', typeof result);
      
      // 预处理：确保字符串格式的结果被解析为JSON
      let processedResult = result;
      if (typeof result === 'string') {
        // 去除可能的前后引号（处理双重引号问题）
        let cleanStr = result.trim();
        if ((cleanStr.startsWith('"') && cleanStr.endsWith('"')) || 
            (cleanStr.startsWith("'") && cleanStr.endsWith("'"))) {
          cleanStr = cleanStr.substring(1, cleanStr.length - 1);
          console.log('去除了外层引号');
        }
        
        // 尝试解析JSON
        try {
          processedResult = JSON.parse(cleanStr);
          console.log('成功将字符串解析为JSON对象');
        } catch (e) {
          console.error('JSON解析失败，尝试修复格式:', e);
          
          // 尝试修复常见问题
          try {
            // 替换不正确的引号和转义
            const fixedStr = cleanStr
              .replace(/\\"/g, '"')       // 替换转义的双引号
              .replace(/\\'/g, "'")       // 替换转义的单引号
              .replace(/\\\\/g, "\\");    // 替换双反斜杠
              
            processedResult = JSON.parse(fixedStr);
            console.log('修复后成功解析JSON');
          } catch (e2) {
            console.error('修复后仍然解析失败，保持原始格式:', e2);
            processedResult = cleanStr;
          }
        }
      }
      
      // 检查结果类型并规范化为数组
      let resultArray: any[] = [];
      
      // 如果结果已经是数组，直接使用
      if (Array.isArray(processedResult)) {
        console.log('结果是数组格式，长度:', processedResult.length);
        resultArray = processedResult;
        
        // 查看第一个元素，可能需要进一步解析
        if (processedResult.length > 0) {
          const firstItem = processedResult[0];
          if (typeof firstItem === 'string') {
            try {
              // 尝试解析第一个元素
              const parsed = JSON.parse(firstItem);
              if (Array.isArray(parsed) || (typeof parsed === 'object' && parsed !== null)) {
                console.log('第一个元素是可解析的JSON字符串，解析后类型:', typeof parsed, 
                  Array.isArray(parsed) ? `数组长度:${parsed.length}` : '');
                // 替换整个结果数组
                resultArray = [parsed];
              }
            } catch (e) {
              console.log('第一个元素不是有效的JSON字符串');
            }
          }
        }
      } 
      // 对象类型，尝试提取有用信息
      else if (typeof processedResult === 'object' && processedResult !== null) {
        if (processedResult.results || processedResult.data) {
          const dataToProcess = processedResult.results || processedResult.data;
          if (Array.isArray(dataToProcess)) {
            resultArray = dataToProcess;
          } else if (typeof dataToProcess === 'object') {
            resultArray = [dataToProcess];
          } else if (typeof dataToProcess === 'string') {
            // 递归处理嵌套的字符串结果
            processResults(dataToProcess);
            return;
          }
        } else if (processedResult.name || processedResult.answers || processedResult.score || processedResult.feedback) {
          // 看起来是单个学生结果
          resultArray = [processedResult];
        } else {
          // 将对象本身作为单个结果
          resultArray = [processedResult];
        }
      }
      // 其他类型直接显示
      else {
        console.log('结果类型无法处理:', typeof processedResult);
        setResult(JSON.stringify(processedResult));
        setDisplayResults(true);
        return;
      }
      
      // 处理解析后的结果数组
      if (resultArray.length > 0) {
        console.log('处理解析后的学生结果数组:', resultArray.length);
        
        // 调用parseStudentResults处理整个数组
        const parsedResults = parseStudentResults(resultArray);
        
        setStudentResults(parsedResults);
        setDisplayResults(true);
        
        // 任务已完成，重置任务相关状态
        resetTaskState();
      } else {
        console.error('解析后的结果数组为空');
        setResult(typeof processedResult === 'string' ? processedResult : JSON.stringify(processedResult));
        setDisplayResults(true);
        
        // 即使结果为空，也重置任务状态
        resetTaskState();
      }
    } catch (error) {
      console.error('处理结果总体失败:', error);
      setResult(typeof result === 'string' ? result : JSON.stringify(result));
      setDisplayResults(true);
      
      // 出错也重置任务状态
      resetTaskState();
    }
  };

  // 修复处理数组的逻辑
  const parseStudentResults = (dataArray: any[]): StudentResult[] => {
    console.log(`开始解析学生结果数据类型:`, typeof dataArray, Array.isArray(dataArray) ? `数组长度:${dataArray.length}` : '非数组');
    
    // 确保dataArray是一个可迭代的数组
    let processedArray = dataArray;
    
    // 修复数组嵌套问题
    if (Array.isArray(dataArray) && dataArray.length === 1) {
      const firstItem = dataArray[0];
      // 如果第一项也是数组，直接使用它
      if (Array.isArray(firstItem)) {
        processedArray = firstItem;
        console.log('使用第一个元素作为主数组，长度:', processedArray.length);
      } 
      // 如果第一项是字符串，尝试解析为数组
      else if (typeof firstItem === 'string') {
        try {
          const parsed = JSON.parse(firstItem);
          if (Array.isArray(parsed)) {
            processedArray = parsed;
            console.log('从字符串解析出数组，长度:', processedArray.length);
          } else {
            // 如果解析结果不是数组，将其包装为数组
            processedArray = [parsed];
            console.log('从字符串解析出非数组对象，将其包装为单元素数组');
          }
        } catch (e) {
          console.error('解析字符串为JSON失败:', e);
          // 保持原状，处理单个字符串元素
        }
      }
      // 如果第一项是非数组对象，但包含学生信息，将其包装
      else if (typeof firstItem === 'object' && firstItem !== null) {
        // 检查是否有answers属性，可能是单个学生对象
        if (firstItem.answers || firstItem.name) {
          processedArray = [firstItem];
          console.log('处理单个学生对象');
        }
      }
    }
    
    // 最终安全检查，确保processedArray可迭代
    if (!Array.isArray(processedArray)) {
      console.warn('处理后的数据仍然不是数组，创建空结果');
      return [];
    }
    
    const results: StudentResult[] = [];
    
    // 遍历处理每个学生数据
    for (let i = 0; i < processedArray.length; i++) {
      try {
        const studentData = processedArray[i];
        console.log(`处理第${i+1}个学生数据:`, typeof studentData);
        // Pass the index to parseStudentResult
        const result = parseStudentResult(studentData, i);
        results.push(result);
      } catch (e) {
        console.error(`解析第${i+1}个学生数据失败:`, e);
        // 添加一个错误占位结果
        results.push({
          name: `解析错误(学生${i+1})`,
          class: '',
          score: '0',
          feedback: '数据格式错误，无法解析',
          answers: [],
          totalQuestions: 0,
          correctCount: 0
        });
      }
    }
    
    console.log(`成功解析${results.length}个学生结果`);
    return results;
  };

  // 修改单个学生结果解析函数，确保只处理单个对象
  const parseStudentResult = (data: any, index: number = 0): StudentResult => {
    // 如果传入的是数组，调用数组处理函数
    if (Array.isArray(data)) {
      console.warn('parseStudentResult收到了数组数据，应该使用parseStudentResults函数');
      // 返回数组中的第一个元素，或创建一个空结果
      return data.length > 0 
        ? parseStudentResult(data[0]) 
        : {
            name: '未知学生',
            class: '',
            score: '0',
            feedback: '数据格式错误',
            answers: [],
            totalQuestions: 0,
            correctCount: 0
          };
    }
    
    console.log('解析单个学生结果:', data.name || '未知学生');
    
    try {
      // 基础信息
      const result: StudentResult = {
        name: data.name || '未知学生',
        class: data.class || '',
        answers: [],
        score: data.overallScore || data.score || '0',
        feedback: data.feedback || '',
        totalQuestions: 0,
        correctCount: 0
      };
      
      // 处理答案
      if (Array.isArray(data.answers)) {
        result.answers = data.answers.map((answer: any) => {
          const isCorrect = answer.isCorrect === true || answer.isCorrect === 'true';
          
          // 如果答案正确，增加正确计数
          if (isCorrect) {
            result.correctCount += 1;
          }
          
          // 增加总问题计数
          result.totalQuestions += 1;
          
          return {
            questionNumber: answer.questionNumber || '',
            studentAnswer: answer.studentAnswer || '',
            isCorrect: isCorrect,
            correctAnswer: answer.correctAnswer || '',
            explanation: answer.explanation || ''
          };
        });
      }
      
      // 如果没有提供分数但有答案数据，计算正确率作为分数
      if (!data.overallScore && !data.score && result.totalQuestions > 0) {
        const correctRate = (result.correctCount / result.totalQuestions) * 100;
        result.score = correctRate.toFixed(0);
      }
      
      console.log(`解析完成: ${result.name}, 题目数: ${result.totalQuestions}, 正确数: ${result.correctCount}, 得分: ${result.score}`);
      
      // 使用后台返回的 pdfUrl 字段，如果存在的话
      let pdfUrl = '';
      if (data.pdfUrl) {
        // 优先使用后台返回的 pdfUrl
        pdfUrl = data.pdfUrl;
        
        // 确保 URL 包含正确的 API 前缀
        if (!pdfUrl.startsWith('/api/files/') && !pdfUrl.startsWith('http')) {
          // 检查是否已经有 / 前缀
          if (!pdfUrl.startsWith('/')) {
            pdfUrl = `/${pdfUrl}`;
          }
          // 添加 API 路径前缀
          pdfUrl = `/api/files${pdfUrl}`;
          console.log(`[PDF] Added API prefix to URL for student ${index + 1}: ${pdfUrl}`);
        } else if (pdfUrl.startsWith('/api/file/')) {
          // 修复错误的 API 路径
          pdfUrl = pdfUrl.replace('/api/file/', '/api/files/');
          console.log(`[PDF] Fixed API URL path from /api/file/ to /api/files/ for student ${index + 1}`);
        }
        
        console.log(`[PDF] Using PDF URL from backend for student ${index + 1}: ${pdfUrl}`);
      } else if (taskId) {
        // 如果后台没有返回 pdfUrl，则使用本地逻辑构建 URL
        pdfUrl = `/api/files/split/student_${index + 1}.pdf`;
        console.log(`[PDF] Created local PDF URL for student ${index + 1}: ${pdfUrl} (TaskID: ${taskId})`);
      }
      
      return {
        ...result,
        pdfUrl: pdfUrl, // Add the PDF URL
      };
    } catch (error) {
      console.error('解析学生结果时出错:', error);
      return {
        name: '解析错误',
        class: '',
        answers: [],
        score: '0',
        feedback: '结果数据格式异常，无法解析',
        totalQuestions: 0,
        correctCount: 0
      };
    }
  };

  const uploadProps: UploadProps = {
    name: 'homework',
    multiple: false,
    fileList,
    beforeUpload: (file: File) => {
      const isImage = file.type.startsWith('image/');
      const isPDF = file.type === 'application/pdf';
      
      if (!isImage && !isPDF) {
        message.error('只支持图片或PDF文件！');
        return false;
      }

      const isLt100M = file.size / 1024 / 1024 < 100;
      if (!isLt100M) {
        message.error('文件大小不能超过100MB！');
        return false;
      }

      setFileList([file as unknown as UploadFile]);
      return false;
    },
    onChange(info: any) {
      setFileList(info.fileList.slice(-1));
    },
  };

  const handleSubmit = async (values: FormValues) => {
    try {
      // 重置状态
      setResult(null);
      setDisplayResults(false);
      setLoading(true);
      
      // 先清除之前的任务状态和轮询
      if (taskId) {
        console.log('清除之前的任务状态和轮询');
        stopPolling();
        setTaskId(null);
      }
      
      setTaskStatus(null);
      setProcessedCount(0);
      setTotalStudents(0);
      setTaskProgress(0);
      setStudentResults([]);
      
      console.log('开始上传作业，表单数据:', values);
      const formData = new FormData();
      const file = fileList[0]?.originFileObj;
      
      if (!file) {
        console.error('未选择文件');
        message.error('请选择要上传的文件！');
        setLoading(false);
        return;
      }

      console.log('准备上传文件:', {
        name: file.name,
        type: file.type,
        size: file.size
      });

      formData.append('homework', file);
      formData.append('type', values.type);
      formData.append('layout', values.layout);
      formData.append('pagesPerStudent', values.pagesPerStudent.toString());

      console.log('开始发送上传请求...');
      const response = await axios.post('/api/upload/homework', formData, {
        headers: {
          'Content-Type': 'multipart/form-data',
        },
      });

      console.log('API响应:', response.data);
      
      if (response.data.success) {
        let validTaskId = null;
        
        // 检查是否包含任务ID
        if (response.data.data && response.data.data.taskId) {
          // 新格式任务ID
          validTaskId = response.data.data.taskId.trim();
          if (!validTaskId) {
            console.warn('收到的任务ID为空字符串');
            message.warning('收到的任务ID无效，无法跟踪处理进度');
            setLoading(false);
            return;
          }
          
          console.log('获取到有效的任务ID:', validTaskId);
          setTotalStudents(response.data.data.totalStudents || 0);
          message.info('文件已接收，正在异步处理');
          
          // 确认任务ID有效后设置状态，这将触发轮询
          setTaskId(validTaskId);
          
        } else if (response.data.task_id) {
          // 旧格式任务ID
          validTaskId = response.data.task_id.trim();
          if (!validTaskId) {
            console.warn('收到的旧格式任务ID为空字符串');
            message.warning('收到的任务ID无效，无法跟踪处理进度');
            setLoading(false);
            return;
          }
          
          console.log('获取到有效的旧格式任务ID:', validTaskId);
          setTotalStudents(response.data.students || 0);
          message.info('文件已接收，正在异步处理');
          
          // 确认任务ID有效后设置状态，这将触发轮询
          setTaskId(validTaskId);
          
        } else if (response.data.status === 'success') {
          // 同步处理完成
          const responseData = response.data;
          console.log('同步处理完成，无需任务ID和轮询');
          console.log('Result type:', typeof responseData.results);
          console.log('Result length:', responseData.results ? responseData.results.length : 0);
          
          if (responseData.results) {
            setResult(JSON.stringify(responseData.results));
            processResults(responseData.results);
            setDisplayResults(true); // 确保显示结果
          } else if (responseData.result) {
            setResult(responseData.result);
            try {
              const parsedResult = JSON.parse(responseData.result);
              if (Array.isArray(parsedResult)) {
                processResults(parsedResult);
              } else {
                // 单个学生结果
                processResults([JSON.stringify(parsedResult)]);
              }
              setDisplayResults(true); // 确保显示结果
            } catch (e) {
              console.error('解析结果失败:', e);
              setDisplayResults(true); // 尝试显示原始结果
            }
          }
          
          setLoading(false);
          message.success('文件处理成功');
        } else {
          console.log('响应没有包含预期的字段，可能是格式变更');
          setLoading(false);
          message.warning('服务器返回了未预期的格式，请联系管理员');
        }
      } else {
        console.error('上传失败:', response.data.message || '未知错误');
        message.error(response.data.message || '上传失败，请重试');
        setLoading(false);
      }
    } catch (error: any) {
      console.error('发生错误:', {
        message: error.message,
        stack: error.stack,
        error: error
      });
      message.error(error.message || '操作失败，请重试！');
      setLoading(false);
    }
  };

  const resetForm = () => {
    form.resetFields();
    setFileList([]);
    setResult(null);
    setDisplayResults(false);
  };

  // 添加导出Excel功能
  const handleDownloadExcel = () => {
    if (!studentResults || studentResults.length === 0) {
      message.error('没有可下载的批改结果');
      return;
    }

    try {
      const fileName = '作业批改结果.xlsx';
      const workbook = XLSX.utils.book_new();
      
      // 创建总览表 - 按照参考图片格式调整
      
      // 先确定所有问题编号，以便创建列
      const allQuestionNumbers: Set<string> = new Set();
      studentResults.forEach(student => {
        if (student.answers && student.answers.length > 0) {
          student.answers.forEach(answer => {
            if (answer.questionNumber) {
              allQuestionNumbers.add(answer.questionNumber);
            }
          });
        }
      });
      
      // 对问题编号进行排序
      const sortedQuestionNumbers = Array.from(allQuestionNumbers).sort((a, b) => {
        // 尝试将问题编号转为数字进行排序
        const numA = parseInt(a);
        const numB = parseInt(b);
        if (!isNaN(numA) && !isNaN(numB)) {
          return numA - numB;
        }
        // 否则进行字符串排序
        return a.localeCompare(b);
      });
      
      // 为每个学生创建一行数据
      const overviewData = studentResults.map((student, index) => {
        // 获取学生所有题目答案状态
        const questionResults: Record<string, string> = {};
        
        // 初始化所有问题为空
        sortedQuestionNumbers.forEach(qNum => {
          questionResults[`题号${qNum}`] = "";
        });
        
        // 填充学生已回答的问题
        if (student.answers && student.answers.length > 0) {
          student.answers.forEach(answer => {
            if (answer.questionNumber) {
              questionResults[`题号${answer.questionNumber}`] = answer.isCorrect ? '正确' : '错误';
            }
          });
        }
        
        // 基础信息
        const baseData: Record<string, any> = {
          '序号': index + 1,
          '学生姓名': student.name || `学生${index + 1}`,
          '班级': student.class || 'N/A',
        };
        
        // 添加所有题目的答案状态
        sortedQuestionNumbers.forEach(qNum => {
          baseData[`题号${qNum}`] = questionResults[`题号${qNum}`];
        });
        
        // 添加总结信息
        return {
          ...baseData,
          '总分': student.score || 'N/A',
          '正确题数': student.correctCount || 0,
          '总题数': student.totalQuestions || 0,
          '正确率': student.totalQuestions ? `${((student.correctCount / student.totalQuestions) * 100).toFixed(1)}%` : 'N/A',
          '总体评价': student.feedback || 'N/A'
        };
      });
      
      // 创建工作表并应用样式
      const overviewSheet = XLSX.utils.json_to_sheet(overviewData);
      
      // 设置列宽
      const columnWidths: Array<{wch: number}> = [
        { wch: 8 },  // A 序号
        { wch: 12 }, // B 学生姓名
        { wch: 10 }, // C 班级
      ];
      
      // 为每个问题添加列宽
      sortedQuestionNumbers.forEach(() => {
        columnWidths.push({ wch: 10 }); // 题号列宽
      });
      
      // 添加后续列的宽度
      columnWidths.push(
        { wch: 10 }, // 总分
        { wch: 10 }, // 正确题数
        { wch: 10 }, // 总题数
        { wch: 10 }, // 正确率
        { wch: 40 }  // 总体评价
      );
      
      // 应用列宽
      overviewSheet['!cols'] = columnWidths;
      
      // 添加到工作簿
      XLSX.utils.book_append_sheet(workbook, overviewSheet, '总览');
      
      // 为每个学生创建详细表 - 保持原有逻辑
      studentResults.forEach((student, index) => {
        const studentName = student.name || `学生${index + 1}`;
        
        // 准备学生答案数据
        let answerData: Record<string, any>[] = [];
        if (student.answers && student.answers.length > 0) {
          answerData = student.answers.map(answer => {
            const baseData: Record<string, any> = {
              '题号': answer.questionNumber || 'N/A',
              '学生答案': answer.studentAnswer || 'N/A',
              '是否正确': answer.isCorrect ? '是' : '否'
            };
            
            // 如果有标准答案，添加进来
            if (!answer.isCorrect && answer.correctAnswer) {
              baseData['标准答案'] = answer.correctAnswer;
            }
            
            // 如果有解释，添加进来
            if (answer.explanation) {
              baseData['解释/评价'] = answer.explanation;
            }
            
            return baseData;
          });
        }
        
        // 如果有答案数据，添加到工作簿
        if (answerData.length > 0) {
          const answersSheet = XLSX.utils.json_to_sheet(answerData);
          // 设置详细表的列宽
          answersSheet['!cols'] = [
            { wch: 8 },  // 题号
            { wch: 40 }, // 学生答案
            { wch: 10 }, // 是否正确
            { wch: 40 }, // 标准答案
            { wch: 40 }  // 解释/评价
          ];
          XLSX.utils.book_append_sheet(workbook, answersSheet, studentName.substring(0, 31)); // Excel的工作表名称最长为31个字符
        }
      });
      
      // 导出Excel文件
      XLSX.writeFile(workbook, fileName);
      message.success('批改结果已成功导出为Excel文件');
    } catch (error) {
      console.error('导出Excel时出错:', error);
      message.error('导出Excel失败，请重试');
    }
  };

  const renderResults = () => {
    if (loading) {
      return (
        <div style={{ textAlign: 'center', marginTop: 24 }}>
          <Spin size="large" />
          <p>正在处理作业，请稍候...</p>
          {errorCount > 0 && (
            <div style={{ marginTop: 8 }}>
              <Tag color="warning">查询状态失败 ({errorCount}/3)，系统将自动重试</Tag>
            </div>
          )}
        </div>
      );
    }

    if (taskId && taskStatus && (taskStatus === 'processing' || taskStatus === 'pending')) {
      return renderTaskProgress();
    }

    if (studentResults.length === 0 && !result) {
      return (
        <Paragraph style={{ marginTop: 16 }}>
          未找到评估数据，请上传作业进行批改。
        </Paragraph>
      );
    }

    // 显示学生列表结果
    return (
      <div style={{ marginTop: 16 }}>
        {studentResults.length > 0 ? (
          <div>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
              <Paragraph style={{ margin: 0 }}>
                <Text strong>批改完成！共 {studentResults.length} 名学生的作业结果</Text>
              </Paragraph>
              <Button
                type="primary"
                icon={<DownloadOutlined />}
                onClick={handleDownloadExcel}
              >
                下载批改结果
              </Button>
            </div>
            
            {/* 使用可折叠面板展示学生结果 */}
            <Collapse defaultActiveKey={[]}>
              {studentResults.map((student, studentIndex) => (
                <Panel 
                  key={`student-${studentIndex}`}
                  header={
                    <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', width: '100%' }}>
                      <div>
                        <span style={{ fontWeight: 'bold' }}>{student.name}</span>
                        {student.class && <span style={{ marginLeft: 8, color: '#8c8c8c' }}>({student.class})</span>} 
                        <span style={{ marginLeft: 8 }}>- 得分: {student.score}</span>
                        <span style={{ marginLeft: 8, color: student.correctCount === student.totalQuestions ? '#52c41a' : '#1890ff' }}>
                          ({student.correctCount}/{student.totalQuestions})
                        </span>
                      </div>
                      <div onClick={e => e.stopPropagation()}>
                        <Button 
                          icon={<FilePdfOutlined />} 
                          size="small"
                          style={{ marginRight: 8 }}
                          onClick={() => handleViewPdf(studentIndex)}
                        >
                          查看PDF
                        </Button>
                        <Button 
                          icon={<EyeOutlined />} 
                          size="small"
                          style={{ marginRight: 8 }}
                          onClick={() => handleViewSideBySide(studentIndex)}
                        >
                          对照视图
                        </Button>
                        <Button 
                          icon={<EditOutlined />} 
                          size="small" 
                          onClick={() => handleEditStudent(studentIndex)}
                        >
                          编辑
                        </Button>
                      </div>
                    </div>
                  }
                >
                  <Card bordered={false} style={{ marginBottom: '0' }}>
                    <div>
                      <Text strong>评分: </Text>
                      <Text>{student.score || '未评分'}</Text>
                    </div>
                    {student.class && (
                      <div style={{ marginTop: 8 }}>
                        <Text strong>班级: </Text>
                        <Text>{student.class}</Text>
                      </div>
                    )}
                    <div style={{ marginTop: 8 }}>
                      <Text strong>反馈: </Text>
                      <div style={{ whiteSpace: 'pre-wrap' }}>
                        {student.feedback || '无反馈信息'}
                      </div>
                    </div>
                    
                    {student.answers && student.answers.length > 0 && (
                      <div style={{ marginTop: 16 }}>
                        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 8 }}>
                          <Text strong>答题详情: </Text>
                        </div>
                        <List
                          size="small"
                          bordered
                          dataSource={student.answers}
                          renderItem={(answer: AnswerDetail, answerIndex: number) => (
                            <List.Item>
                              <div style={{ width: '100%' }}>
                                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                                  <Text strong>题目 {answer.questionNumber}: </Text>
                                  <div>
                                    <Tag 
                                      color={answer.isCorrect ? 'success' : 'error'}
                                      onClick={() => toggleAnswerCorrectness(studentIndex, answerIndex)}
                                      style={{ cursor: 'pointer', marginRight: 8 }}
                                    >
                                      {answer.isCorrect ? '正确' : '错误'} (点击修改)
                                    </Tag>
                                    <Button 
                                      icon={<EditOutlined />} 
                                      size="small" 
                                      onClick={() => handleEditAnswer(studentIndex, answerIndex)}
                                    >
                                      编辑
                                    </Button>
                                  </div>
                                </div>
                                <div style={{ marginTop: 4 }}>
                                  <Text type="secondary">学生答案: </Text>
                                  <Text>{answer.studentAnswer}</Text>
                                </div>
                                {!answer.isCorrect && (
                                  <div style={{ marginTop: 4 }}>
                                    <Text type="secondary">正确答案: </Text>
                                    <Text type="success">{answer.correctAnswer}</Text>
                                  </div>
                                )}
                                {answer.explanation && (
                                  <div style={{ marginTop: 4 }}>
                                    <Text type="secondary">解释: </Text>
                                    <Text>{answer.explanation}</Text>
                                  </div>
                                )}
                              </div>
                            </List.Item>
                          )}
                        />
                      </div>
                    )}
                  </Card>
                </Panel>
              ))}
            </Collapse>
          </div>
        ) : (
          <div>
            <Paragraph>
              未能解析学生结果，原始数据:
              <pre style={{ maxHeight: '400px', overflow: 'auto' }}>{result}</pre>
            </Paragraph>
            <div style={{ marginTop: 16, display: 'flex', gap: '10px' }}>
              <Button type="primary" onClick={() => {resetForm(); setDisplayResults(false);}} >
                重新上传
              </Button>
              <Button
                type="default"
                icon={<DownloadOutlined />}
                onClick={handleDownloadExcel}
              >
                下载批改结果
              </Button>
              <Button
                type="default"
                onClick={handleUploadMasterAnswers}
              >
                设置标准答案
              </Button>
            </div>
          </div>
        )}
        
        {/* 对照视图模态窗口 - 移动至外层确保总是可见 */}
        <Modal
          title="答题详情与原文对照"
          open={showSideBySide !== null}
          onCancel={closeSideBySide}
          width={1400}
          style={{ top: 20 }}
          footer={[
            <Button 
              key="openPdf" 
              onClick={() => currentPdfUrl && openPdfDirectly(currentPdfUrl)}
              style={{ marginRight: 8 }}
            >
              在浏览器中打开PDF
            </Button>,
            <Button key="close" onClick={closeSideBySide}>
              关闭对照视图
            </Button>
          ]}
        >
          {showSideBySide !== null && (
            <div>
              <div style={{ marginBottom: 8, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <div>
                  <Text strong>当前PDF URL: </Text>
                  <Text code>{currentPdfUrl}</Text>
                </div>
                <div>
                  <Button 
                    type="default" 
                    icon={<FilePdfOutlined />}
                    onClick={() => currentPdfUrl && openPdfDirectly(currentPdfUrl)}
                  >
                    在浏览器中打开
                  </Button>
                </div>
              </div>
              <div style={{ display: 'flex', gap: '16px' }}>
                <div style={{ flex: '1 1 40%', maxHeight: '80vh', overflowY: 'auto' }}>
                  {showSideBySide !== null && studentResults[showSideBySide]?.answers && (
                    <List
                      size="small"
                      bordered
                      dataSource={studentResults[showSideBySide].answers}
                      renderItem={(answer: AnswerDetail, answerIndex: number) => (
                        <List.Item>
                          <div style={{ width: '100%' }}>
                            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                              <Text strong>题目 {answer.questionNumber}: </Text>
                              <div>
                                <Tag 
                                  color={answer.isCorrect ? 'success' : 'error'}
                                  onClick={() => toggleAnswerCorrectness(showSideBySide, answerIndex)}
                                  style={{ cursor: 'pointer', marginRight: 8 }}
                                >
                                  {answer.isCorrect ? '正确' : '错误'} (点击修改)
                                </Tag>
                                <Button 
                                  icon={<EditOutlined />} 
                                  size="small" 
                                  onClick={() => handleEditAnswer(showSideBySide, answerIndex)}
                                >
                                  编辑
                                </Button>
                              </div>
                            </div>
                            <div style={{ marginTop: 4 }}>
                              <Text type="secondary">学生答案: </Text>
                              <Text>{answer.studentAnswer}</Text>
                            </div>
                            {!answer.isCorrect && (
                              <div style={{ marginTop: 4 }}>
                                <Text type="secondary">正确答案: </Text>
                                <Text type="success">{answer.correctAnswer}</Text>
                              </div>
                            )}
                            {answer.explanation && (
                              <div style={{ marginTop: 4 }}>
                                <Text type="secondary">解释: </Text>
                                <Text>{answer.explanation}</Text>
                              </div>
                            )}
                          </div>
                        </List.Item>
                      )}
                    />
                  )}
                </div>
                <div style={{ flex: '1 1 60%', maxHeight: '80vh', overflowY: 'auto', display: 'flex', flexDirection: 'column' }}>
                  <div style={{ flex: 1, display: 'flex', justifyContent: 'center', overflow: 'auto' }}>
                    {currentPdfUrl ? (
                      <div key={currentPdfUrl} style={{ width: '100%', display: 'flex', justifyContent: 'center' }}>
                        <Document
                          file={currentPdfUrl}
                          onLoadSuccess={onDocumentLoadSuccess}
                          onLoadProgress={(progress) => console.log(`[PDF] 加载进度: ${JSON.stringify(progress)}`)}
                          loading={<Spin tip="PDF正在下载中，请等待..." />}
                          onLoadError={(error) => {
                            console.error('PDF加载错误:', error);
                            console.error('PDF URL:', currentPdfUrl);
                            // 更详细的错误信息
                            if (error.name === 'MissingPDFException') {
                              message.error(`PDF文件未找到。请确认服务器上是否存在该文件: ${currentPdfUrl}`);
                            } else if (error.name === 'InvalidPDFException') {
                              message.error(`无效的PDF文件格式: ${currentPdfUrl}`);
                            } else if (error.name === 'UnexpectedResponseException') {
                              message.error(`服务器响应异常: ${error.message}`);
                            } else {
                              message.error(`加载PDF文件失败 (${error.name}): ${error.message}`);
                            }
                            // 将错误信息记录到控制台，便于调试
                            console.log('完整错误对象:', JSON.stringify(error, null, 2));
                          }}
                          options={pdfOptions}
                          externalLinkTarget="_blank"
                          renderMode="canvas"
                        >
                          <Page 
                            pageNumber={currentPage} 
                            scale={pdfScale}
                            renderTextLayer={true}
                            renderAnnotationLayer={true}
                            error={<div>无法加载此页面，请尝试其他页面</div>}
                            noData={<div>正在加载PDF...</div>}
                            loading={<Spin tip="PDF正在加载中..." />}
                          />
                        </Document>
                      </div>
                    ) : (
                      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', height: '100%' }}>
                        <Spin tip="PDF加载中..." />
                      </div>
                    )}
                  </div>
                  <div style={{ marginTop: 16, display: 'flex', justifyContent: 'space-between', width: '100%' }}>
                    <div>
                      <Button onClick={previousPage} disabled={currentPage <= 1}>上一页</Button>
                      <Button onClick={nextPage} disabled={currentPage >= (numPages || 1)} style={{ marginLeft: 8 }}>下一页</Button>
                    </div>
                    <div>
                      <Text>{currentPage} / {numPages || 1}</Text>
                    </div>
                    <div>
                      <Button onClick={() => setPdfScale(prevScale => Math.max(0.5, prevScale - 0.1))}>缩小</Button>
                      <Button onClick={() => setPdfScale(1.0)} style={{ margin: '0 8px' }}>重置</Button>
                      <Button onClick={() => setPdfScale(prevScale => Math.min(2.0, prevScale + 0.1))}>放大</Button>
                      <Text style={{ marginLeft: 8 }}>{Math.round(pdfScale * 100)}%</Text>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          )}
        </Modal>
      </div>
    );
  };

  // 渲染任务处理进度
  const renderTaskProgress = () => {
    if (!taskId || !taskStatus) return null;
    
    // 如果任务状态是失败状态，显示错误信息而不是进度条
    if (taskStatus === 'failed' || taskStatus === 'error' || taskStatus === 'timeout') {
      const isTimeout = taskStatus === 'timeout';
      return (
        <Card title="处理状态" style={{ marginTop: 16 }}>
          <Result
            status="warning"
            title={isTimeout ? "处理超时" : "处理失败"}
            subTitle={isTimeout ? 
              "作业处理时间过长，可能是服务器繁忙或文件较复杂，您可以稍后查看结果或重新提交" : 
              "作业处理过程中发生错误，请尝试重新上传"}
            extra={
              <Button type="primary" onClick={resetForm}>
                重新上传
              </Button>
            }
          />
        </Card>
      );
    }
    
    return (
      <Card title="处理进度" style={{ marginTop: 16 }}>
        <Progress percent={taskProgress} status={taskStatus === 'processing' ? 'active' : 'normal'} />
        <div style={{ marginTop: 16 }}>
          <Text strong>已处理学生: </Text>
          <Text>{processedCount}</Text>
        </div>
        <div>
          <Text strong>总学生数: </Text>
          <Text>{totalStudents}</Text>
        </div>
        {errorCount > 0 && (
          <div style={{ marginTop: 8 }}>
            <Tag color="warning">查询状态失败 ({errorCount}/3)，系统将自动重试</Tag>
          </div>
        )}
      </Card>
    );
  };

  // 添加重置任务状态的函数
  const resetTaskState = () => {
    console.log('开始重置任务状态');
    
    // 先清除轮询
    console.log(`重置前检查：轮询定时器${pollingIntervalRef.current ? '存在' : '不存在'}`);
    if (pollingIntervalRef.current) {
      try {
        clearInterval(pollingIntervalRef.current);
        console.log('已清除轮询定时器');
      } catch (err) {
        console.error('清除定时器时出错:', err);
      }
      pollingIntervalRef.current = null;
      setPollingInterval(null);
    }
    
    // 设置标志位避免重复调用
    const previousTaskId = taskId;
    
    // 重置任务相关状态
    setTaskId(null);
    setTaskStatus(null);
    setProcessedCount(0);
    setTotalStudents(0);
    setTaskProgress(0);
    setPollCount(0);
    setErrorCount(0);
    
    console.log(`任务状态已重置，原任务ID: ${previousTaskId}`);
  };

  // 组件卸载时清理所有资源
  useEffect(() => {
    return () => {
      console.log('组件卸载，执行清理');
      if (pollingIntervalRef.current) {
        clearInterval(pollingIntervalRef.current);
        pollingIntervalRef.current = null;
      }
      if (pollingInterval) {
        clearInterval(pollingInterval);
      }
      console.log('组件卸载时已清理所有定时器');
    };
  }, []);

  // Handle editing answer details
  const handleEditAnswer = (studentIndex: number, answerIndex: number) => {
    setEditingStudentIndex(studentIndex);
    setEditingAnswerIndex(answerIndex);
    
    // Get the current answer data to populate the form
    const student = studentResults[studentIndex];
    const answer = student.answers![answerIndex];
    
    // Set form values
    editForm.setFieldsValue({
      studentAnswer: answer.studentAnswer,
      isCorrect: answer.isCorrect,
      correctAnswer: answer.correctAnswer,
      explanation: answer.explanation,
    });
    
    setIsEditModalVisible(true);
  };
  
  // Save edited answer
  const handleSaveEdit = () => {
    editForm.validateFields().then(values => {
      if (editingStudentIndex === null || editingAnswerIndex === null) return;
      
      setStudentResults(prevResults => {
        // Create a deep copy to avoid directly modifying state
        const newResults = [...prevResults];
        const student = newResults[editingStudentIndex];
        
        if (student && student.answers) {
          // Update the answer with form values
          student.answers[editingAnswerIndex] = {
            ...student.answers[editingAnswerIndex],
            studentAnswer: values.studentAnswer,
            isCorrect: values.isCorrect,
            correctAnswer: values.correctAnswer,
            explanation: values.explanation,
          };
          
          // Recalculate correctCount and score
          const correctAnswers = student.answers.filter(a => a.isCorrect).length;
          student.correctCount = correctAnswers;
          
          if (student.totalQuestions > 0) {
            const correctRate = correctAnswers / student.totalQuestions;
            const correctPercent = Math.round(correctRate * 100);
            student.score = `${correctPercent}`;
          }
        }
        
        return newResults;
      });
      
      // Close modal and reset editing state
      setIsEditModalVisible(false);
      setEditingStudentIndex(null);
      setEditingAnswerIndex(null);
      message.success('答案已更新');
    });
  };
  
  // Add this at the end of the component, before the return statement
  const handleCancelEdit = () => {
    setIsEditModalVisible(false);
    setEditingStudentIndex(null);
    setEditingAnswerIndex(null);
  };

  // 添加切换答案正确性的函数
  const toggleAnswerCorrectness = (studentIndex: number, answerIndex: number) => {
    setStudentResults(prevResults => {
      // 创建一个深拷贝，避免直接修改状态
      const newResults = [...prevResults];
      
      // 确保存在学生和答案
      if (newResults[studentIndex] && newResults[studentIndex].answers && newResults[studentIndex].answers![answerIndex]) {
        // 切换正确/错误状态
        const isCurrentlyCorrect = newResults[studentIndex].answers![answerIndex].isCorrect;
        newResults[studentIndex].answers![answerIndex].isCorrect = !isCurrentlyCorrect;
        
        // 更新该学生的correctCount
        const correctAnswers = newResults[studentIndex].answers!.filter(a => a.isCorrect).length;
        newResults[studentIndex].correctCount = correctAnswers;
        
        // 如果学生有总题数，更新正确率显示在评分中
        if (newResults[studentIndex].totalQuestions > 0) {
          const correctRate = correctAnswers / newResults[studentIndex].totalQuestions;
          const correctPercent = Math.round(correctRate * 100);
          newResults[studentIndex].score = `${correctPercent}`;
        }
        
        // 显示操作提示
        message.success(`已${!isCurrentlyCorrect ? '修正' : '标记'}为${!isCurrentlyCorrect ? '正确' : '错误'}`);
      }
      
      return newResults;
    });
  };

  // Handle editing student information
  const handleEditStudent = (studentIndex: number) => {
    setEditingStudentIndex(studentIndex);
    
    // Get the current student data to populate the form
    const student = studentResults[studentIndex];
    
    // Set form values
    studentEditForm.setFieldsValue({
      name: student.name,
      class: student.class,
      score: student.score,
      feedback: student.feedback,
    });
    
    setIsStudentEditModalVisible(true);
  };
  
  // Save edited student information
  const handleSaveStudentEdit = () => {
    studentEditForm.validateFields().then(values => {
      if (editingStudentIndex === null) return;
      
      setStudentResults(prevResults => {
        // Create a deep copy to avoid directly modifying state
        const newResults = [...prevResults];
        const student = newResults[editingStudentIndex];
        
        if (student) {
          // Update the student information with form values
          student.name = values.name;
          student.class = values.class;
          student.score = values.score;
          student.feedback = values.feedback;
        }
        
        return newResults;
      });
      
      // Close modal and reset editing state
      setIsStudentEditModalVisible(false);
      setEditingStudentIndex(null);
      message.success('学生信息已更新');
    });
  };
  
  // Cancel student info edit
  const handleCancelStudentEdit = () => {
    setIsStudentEditModalVisible(false);
    setEditingStudentIndex(null);
  };

  // Handle uploading master answers
  const handleUploadMasterAnswers = () => {
    // Check if we have any student results to work with
    if (studentResults.length === 0) {
      message.error('没有学生数据，无法设置答案');
      return;
    }
    
    // Get the first student to extract question numbers
    const firstStudent = studentResults[0];
    if (!firstStudent.answers || firstStudent.answers.length === 0) {
      message.error('第一个学生没有答题数据，无法设置答案');
      return;
    }
    
    // Prepare form initial values
    const initialValues: Record<string, string> = {};
    firstStudent.answers.forEach(answer => {
      const fieldName = `answer_${answer.questionNumber}`;
      // If we already have a correct answer for incorrect answers, use it
      initialValues[fieldName] = answer.isCorrect ? answer.studentAnswer : (answer.correctAnswer || '');
    });
    
    // Set the form values
    masterAnswersForm.setFieldsValue(initialValues);
    
    // Show the modal
    setIsMasterAnswersModalVisible(true);
  };
  
  // Save master answers and auto-grade
  const handleSaveMasterAnswers = () => {
    masterAnswersForm.validateFields().then(values => {
      // Create a mapping of question numbers to correct answers
      const correctAnswers: Record<string, string> = {};
      
      // Extract answers from form values
      Object.keys(values).forEach(key => {
        if (key.startsWith('answer_')) {
          const questionNumber = key.replace('answer_', '');
          correctAnswers[questionNumber] = values[key];
        }
      });
      
      // Update all students' answers
      setStudentResults(prevResults => {
        const newResults = [...prevResults];
        
        // For each student
        newResults.forEach(student => {
          if (!student.answers) return;
          
          let correctCount = 0;
          
          // For each of their answers
          student.answers.forEach(answer => {
            const correctAnswer = correctAnswers[answer.questionNumber];
            
            // Skip if we don't have a correct answer for this question
            if (!correctAnswer) return;
            
            // Compare and update
            const isCorrect = answer.studentAnswer === correctAnswer;
            answer.isCorrect = isCorrect;
            
            // If incorrect, update the correct answer
            if (!isCorrect) {
              answer.correctAnswer = correctAnswer;
            }
            
            // Count correct answers
            if (isCorrect) correctCount++;
          });
          
          // Update student's correct count and score
          student.correctCount = correctCount;
          if (student.totalQuestions > 0) {
            const correctRate = correctCount / student.totalQuestions;
            const correctPercent = Math.round(correctRate * 100);
            student.score = `${correctPercent}`;
          }
        });
        
        return newResults;
      });
      
      // Close the modal
      setIsMasterAnswersModalVisible(false);
      message.success('答案设置成功，已更新所有学生的评分');
    });
  };

  // Function to view student's PDF in a modal
  const handleViewPdf = (studentIndex: number) => {
    const student = studentResults[studentIndex];
    if (student && student.pdfUrl) {
      let pdfUrl = student.pdfUrl;
      
      // 确保 URL 包含正确的 API 前缀
      if (!pdfUrl.startsWith('/api/files/') && !pdfUrl.startsWith('http')) {
        // 检查是否已经有 / 前缀
        if (!pdfUrl.startsWith('/')) {
          pdfUrl = `/${pdfUrl}`;
        }
        // 添加 API 路径前缀
        pdfUrl = `/api/files${pdfUrl}`;
        console.log(`[PDF] Added API prefix to URL for modal view: ${pdfUrl}`);
      } else if (pdfUrl.startsWith('/api/file/')) {
        // 修复错误的 API 路径
        pdfUrl = pdfUrl.replace('/api/file/', '/api/files/');
        console.log(`[PDF] Fixed API URL path from /api/file/ to /api/files/ for modal view`);
      }
      
      console.log(`[PDF] Loading PDF in modal for student ${studentIndex + 1}: ${pdfUrl}`);
      
      // Set state for the PDF viewer
      setCurrentPdfUrl(pdfUrl);
      setCurrentPage(1);
      setIsPdfModalVisible(true);
    } else {
      message.error(`未找到学生${studentIndex + 1}的PDF文件`);
      console.error(`[PDF] Missing PDF URL for student ${studentIndex + 1}`);
    }
  };

  // Function to open PDF directly in a new tab
  const openPdfDirectly = (pdfUrl: string) => {
    // 确保 URL 包含正确的 API 前缀
    if (!pdfUrl.startsWith('/api/files/') && !pdfUrl.startsWith('http')) {
      // 检查是否已经有 / 前缀
      if (!pdfUrl.startsWith('/')) {
        pdfUrl = `/${pdfUrl}`;
      }
      // 添加 API 路径前缀
      pdfUrl = `/api/files${pdfUrl}`;
      console.log(`[PDF] Added API prefix to URL before opening in browser: ${pdfUrl}`);
    } else if (pdfUrl.startsWith('/api/file/')) {
      // 修复错误的 API 路径
      pdfUrl = pdfUrl.replace('/api/file/', '/api/files/');
      console.log(`[PDF] Fixed API URL path from /api/file/ to /api/files/ before opening in browser`);
    }
    
    console.log(`[PDF] Opening PDF directly in browser: ${pdfUrl}`);
    // Open PDF in a new tab
    window.open(pdfUrl, '_blank');
  };

  // Function to view student's PDF alongside answers
  const handleViewSideBySide = (studentIndex: number) => {
    const student = studentResults[studentIndex];
    console.log(`[DEBUG] 打开对照视图，学生索引: ${studentIndex}`, student);
    console.log(`[DEBUG] 学生数组长度: ${studentResults.length}`);
    console.log(`[DEBUG] studentResults 数据类型: ${typeof studentResults}`);
    
    if (student && student.pdfUrl) {
      let pdfUrl = student.pdfUrl;
      console.log(`[DEBUG] 原始 PDF URL: ${pdfUrl}`);
      
      // 确保 URL 包含正确的 API 前缀
      if (!pdfUrl.startsWith('/api/files/') && !pdfUrl.startsWith('http')) {
        // 检查是否已经有 / 前缀
        if (!pdfUrl.startsWith('/')) {
          pdfUrl = `/${pdfUrl}`;
        }
        // 添加 API 路径前缀
        pdfUrl = `/api/files${pdfUrl}`;
        console.log(`[PDF] Added API prefix to URL for side-by-side view: ${pdfUrl}`);
      } else if (pdfUrl.startsWith('/api/file/')) {
        // 修复错误的 API 路径
        pdfUrl = pdfUrl.replace('/api/file/', '/api/files/');
        console.log(`[PDF] Fixed API URL path from /api/file/ to /api/files/ for side-by-side view`);
      }
      
      console.log(`[PDF] Loading side-by-side PDF for student ${studentIndex + 1}: ${pdfUrl}`);
      
      // 检查Modal和PDF相关状态
      console.log(`[DEBUG] 当前Modal状态: showSideBySide=${showSideBySide}, currentPdfUrl=${currentPdfUrl}`);
      
      // 在设置状态前添加日志
      console.log(`[DEBUG] 设置状态: setCurrentPdfUrl(${pdfUrl}), setShowSideBySide(${studentIndex})`);
      
      // 直接设置状态，与handleViewPdf保持一致
      setCurrentPdfUrl(pdfUrl);
      setCurrentPage(1);
      setShowSideBySide(studentIndex);
      
      // 状态设置后添加确认日志
      console.log(`[DEBUG] 状态已设置，确认Modal应该显示 (showSideBySide=${studentIndex})`);
      
      // 尝试主动获取PDF以验证URL是否有效
      fetch(pdfUrl, { method: 'Get' })
        .then(response => {
          console.log(`[DEBUG] PDF URL 检查结果: ${response.status} ${response.statusText}`, response);
          if (!response.ok) {
            console.error(`[DEBUG] PDF URL 无效: ${pdfUrl}`);
            message.error(`PDF URL 无效 (${response.status}): ${pdfUrl}`);
          } else {
            console.log(`[DEBUG] PDF URL 有效: ${pdfUrl}`);
          }
        })
        .catch(error => {
          console.error(`[DEBUG] 检查PDF URL时出错:`, error);
          message.error(`检查PDF URL时出错: ${error.message}`);
        });
    } else {
      message.error(`未找到学生${studentIndex + 1}的PDF文件`);
      console.error(`[PDF] Missing PDF URL for student ${studentIndex + 1}`);
    }
  };

  // Close the side-by-side view
  const closeSideBySide = () => {
    setShowSideBySide(null);
    setCurrentPdfUrl(null);
  };

  // Handle document load success
  const onDocumentLoadSuccess = ({ numPages }: { numPages: number }) => {
    console.log(`[PDF] PDF loaded successfully with ${numPages} pages`);
    setNumPages(numPages);
  };

  // Handle page changes
  const changePage = (offset: number) => {
    setCurrentPage(prevPageNumber => {
      const newPage = (prevPageNumber || 1) + offset;
      return Math.max(1, Math.min(newPage, numPages || 1));
    });
  };

  const previousPage = () => changePage(-1);
  const nextPage = () => changePage(1);

  // 添加useEffect来确保表单实例与表单组件正确连接
  useEffect(() => {
    // 如果Edit Modal关闭了，重置相关状态
    if (!isEditModalVisible) {
      setEditingStudentIndex(null);
      setEditingAnswerIndex(null);
    }
  }, [isEditModalVisible]);

  // 同样处理学生编辑表单
  useEffect(() => {
    if (!isStudentEditModalVisible) {
      setEditingStudentIndex(null);
    }
  }, [isStudentEditModalVisible]);

  // 处理主答案表单
  useEffect(() => {
    if (!isMasterAnswersModalVisible) {
      // 表单关闭后可能需要的清理操作
    }
  }, [isMasterAnswersModalVisible]);

  return (
    <div style={{ 
      maxWidth: '95%', 
      width: '1400px', 
      margin: '0 auto', 
      padding: '24px' 
    }}>
      <Card>
        <Title level={2}>作业上传与批改</Title>
        <Paragraph>
          支持上传图片或PDF格式的作业文件。图片作业支持单栏和双栏布局，PDF作业支持按页数分割。
        </Paragraph>

        {!displayResults || (!result && studentResults.length === 0) ? (
          <Form
            form={form}
            layout="vertical"
            onFinish={handleSubmit}
            style={{ marginTop: 24 }}
          >
            <Form.Item
              label="上传文件"
              required
            >
              <Dragger {...uploadProps}>
                <p className="ant-upload-drag-icon">
                  <InboxOutlined />
                </p>
                <p className="ant-upload-text">点击或拖拽文件到此区域上传</p>
                <p className="ant-upload-hint">
                  支持单个图片或PDF文件，大小不超过100MB
                </p>
              </Dragger>
            </Form.Item>

            <Form.Item
              label="作业类型"
              name="type"
              initialValue="general"
              rules={[{ required: true, message: '请选择作业类型' }]}
            >
              <Select>
                <Select.Option value="general">通用作业</Select.Option>
                <Select.Option value="math">数学作业</Select.Option>
                <Select.Option value="english">英语作业</Select.Option>
                <Select.Option value="chinese">语文作业</Select.Option>
              </Select>
            </Form.Item>

            <Form.Item
              label="布局方式"
              name="layout"
              initialValue="single"
              rules={[{ required: true, message: '请选择布局方式' }]}
            >
              <Select>
                <Select.Option value="single">单栏布局</Select.Option>
                <Select.Option value="double">双栏布局</Select.Option>
              </Select>
            </Form.Item>

            <Form.Item
              label="每个学生作业页数"
              name="pagesPerStudent"
              initialValue={1}
              rules={[{ required: true, message: '请输入每个学生的作业页数' }]}
            >
              <InputNumber min={1} max={10} />
            </Form.Item>

            <Form.Item>
              <Button type="primary" htmlType="submit" loading={loading} disabled={loading}>
                {loading ? '批改中...' : '上传并开始批改'}
              </Button>
            </Form.Item>
          </Form>
        ) : (
          <>
            {renderResults()}
          </>
        )}

        {loading && (
          <div style={{ textAlign: 'center', marginTop: 24 }}>
            <Spin size="large" tip="正在批改作业中，请稍候..." />
            <Paragraph style={{ marginTop: 16 }}>
              作业批改可能需要几分钟时间，取决于文件大小和复杂度
            </Paragraph>
          </div>
        )}
      </Card>

      {/* Edit Answer Modal */}
      <Modal
        title="编辑答案详情"
        open={isEditModalVisible}
        onOk={handleSaveEdit}
        onCancel={handleCancelEdit}
        okText="保存"
        cancelText="取消"
      >
        <Form form={editForm} layout="vertical">
          <Form.Item
            name="studentAnswer"
            label="学生答案"
            rules={[{ required: true, message: '请输入学生答案' }]}
          >
            <Input.TextArea rows={3} />
          </Form.Item>
          
          <Form.Item name="isCorrect" label="答案是否正确" valuePropName="checked">
            <Switch checkedChildren="正确" unCheckedChildren="错误" />
          </Form.Item>
          
          <Form.Item
            name="correctAnswer"
            label="正确答案"
            rules={[{ required: editForm.getFieldValue('isCorrect') === false, message: '请输入正确答案' }]}
          >
            <Input.TextArea rows={3} />
          </Form.Item>
          
          <Form.Item name="explanation" label="解释/评价">
            <Input.TextArea rows={3} />
          </Form.Item>
        </Form>
      </Modal>

      {/* Student Info Edit Modal */}
      <Modal
        title="编辑学生信息"
        open={isStudentEditModalVisible}
        onOk={handleSaveStudentEdit}
        onCancel={handleCancelStudentEdit}
        okText="保存"
        cancelText="取消"
      >
        <Form form={studentEditForm} layout="vertical">
          <Form.Item
            name="name"
            label="学生姓名"
            rules={[{ required: true, message: '请输入学生姓名' }]}
          >
            <Input />
          </Form.Item>
          
          <Form.Item
            name="class"
            label="班级"
          >
            <Input placeholder="可选填写" />
          </Form.Item>
          
          <Form.Item
            name="score"
            label="分数"
            rules={[{ required: true, message: '请输入分数' }]}
          >
            <Input />
          </Form.Item>
          
          <Form.Item
            name="feedback"
            label="反馈"
            rules={[{ required: false }]}
          >
            <Input.TextArea rows={3} placeholder="输入对该学生的反馈意见" />
          </Form.Item>
        </Form>
      </Modal>

      {/* Master Answers Modal */}
      <Modal
        title="设置标准答案"
        open={isMasterAnswersModalVisible}
        onOk={handleSaveMasterAnswers}
        onCancel={() => setIsMasterAnswersModalVisible(false)}
        okText="保存并自动评分"
        cancelText="取消"
        width={600}
      >
        <Paragraph>
          请为每个题目设置标准答案。保存后，系统将自动与学生答案比对并更新评分。
        </Paragraph>
        
        <Form form={masterAnswersForm} layout="vertical">
          {studentResults.length > 0 && studentResults[0].answers && 
            studentResults[0].answers.map(answer => (
              <Form.Item
                key={`answer_${answer.questionNumber}`}
                name={`answer_${answer.questionNumber}`}
                label={`题目 ${answer.questionNumber} 的标准答案`}
                rules={[{ required: true, message: '请输入标准答案' }]}
              >
                <Input.TextArea rows={2} />
              </Form.Item>
            ))
          }
        </Form>
      </Modal>

      {/* PDF Viewer Modal */}
      <Modal
        title="查看作业"
        open={isPdfModalVisible}
        onCancel={() => setIsPdfModalVisible(false)}
        width={1200}
        style={{ top: 20 }}
        footer={[
          <Button key="direct" onClick={() => currentPdfUrl && openPdfDirectly(currentPdfUrl)} type="primary">
            在浏览器中直接打开
          </Button>,
          <Button key="close" onClick={() => setIsPdfModalVisible(false)}>
            关闭
          </Button>
        ]}
      >
        <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', height: '70vh' }}>
          <div style={{ overflow: 'auto', height: '100%', width: '100%', display: 'flex', justifyContent: 'center' }}>
            {currentPdfUrl ? (
              <div key={currentPdfUrl} style={{ width: '100%', display: 'flex', justifyContent: 'center' }}>
                <Document
                  file={currentPdfUrl}
                  onLoadSuccess={onDocumentLoadSuccess}
                  onLoadProgress={(progress) => console.log(`[PDF] 加载进度: ${JSON.stringify(progress)}`)}
                  loading={<Spin tip="PDF正在下载中，请等待..." />}
                  onLoadError={(error) => {
                    console.error('PDF加载错误:', error);
                    console.error('PDF URL:', currentPdfUrl);
                    // 更详细的错误信息
                    if (error.name === 'MissingPDFException') {
                      message.error(`PDF文件未找到。请确认服务器上是否存在该文件: ${currentPdfUrl}`);
                    } else if (error.name === 'InvalidPDFException') {
                      message.error(`无效的PDF文件格式: ${currentPdfUrl}`);
                    } else if (error.name === 'UnexpectedResponseException') {
                      message.error(`服务器响应异常: ${error.message}`);
                    } else {
                      message.error(`加载PDF文件失败 (${error.name}): ${error.message}`);
                    }
                    // 将错误信息记录到控制台，便于调试
                    console.log('完整错误对象:', JSON.stringify(error, null, 2));
                  }}
                  options={pdfOptions}
                  externalLinkTarget="_blank"
                  renderMode="canvas"
                >
                  <Page 
                    pageNumber={currentPage} 
                    scale={pdfScale}
                    renderTextLayer={true}
                    renderAnnotationLayer={true}
                    error={<div>无法加载此页面，请尝试其他页面</div>}
                    noData={<div>正在加载PDF...</div>}
                    loading={<Spin tip="PDF正在加载中..." />}
                  />
                </Document>
              </div>
            ) : (
              <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', height: '100%' }}>
                <Spin tip="PDF加载中..." />
              </div>
            )}
          </div>
          <div style={{ marginTop: 16, display: 'flex', justifyContent: 'space-between', width: '100%' }}>
            <div>
              <Button onClick={previousPage} disabled={currentPage <= 1}>上一页</Button>
              <Button onClick={nextPage} disabled={currentPage >= (numPages || 1)} style={{ marginLeft: 8 }}>下一页</Button>
            </div>
            <div>
              <Text>{currentPage} / {numPages || 1}</Text>
            </div>
            <div>
              <Button onClick={() => setPdfScale(prevScale => Math.max(0.5, prevScale - 0.1))}>缩小</Button>
              <Button onClick={() => setPdfScale(1.0)} style={{ margin: '0 8px' }}>重置</Button>
              <Button onClick={() => setPdfScale(prevScale => Math.min(2.0, prevScale + 0.1))}>放大</Button>
              <Text style={{ marginLeft: 8 }}>{Math.round(pdfScale * 100)}%</Text>
            </div>
          </div>
        </div>
      </Modal>
    </div>
  );
};

export default Homework; 