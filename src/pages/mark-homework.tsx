import React, { useState, useEffect } from 'react';
import { Card, Typography, Spin, message, Button } from 'antd';
import { useRouter } from 'next/router';
import { ReloadOutlined, DownloadOutlined } from '@ant-design/icons';
import * as XLSX from 'xlsx';

const { Title, Paragraph } = Typography;

// 定义接口类型
interface Answer {
  questionNumber: string;
  studentAnswer: string;
  isCorrect: boolean;
  explanation?: string;
  evaluation?: string;
  correctAnswer?: string;
  suggestion?: string;
  [key: string]: any; // 允许其他可能的字段
}

interface Student {
  name: string;
  answers: Answer[];
  overallScore: string;
  feedback: string;
  [key: string]: any; // 允许其他可能的字段
}

const MarkHomework: React.FC = () => {
  const router = useRouter();
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [parsedResult, setParsedResult] = useState<Student | Student[] | any>(null);

  useEffect(() => {
    console.log('MarkHomework 组件加载，路由参数:', router.query);
    
    if (router.query.result) {
      console.log('收到批改结果:', router.query.result);
      const resultString = router.query.result as string;
      setResult(resultString);
      setError(null);
      
      // 尝试解析JSON结果
      try {
        const parsed = JSON.parse(resultString);
        setParsedResult(parsed);
        console.log('成功解析批改结果为JSON:', parsed);
      } catch (err) {
        console.error('解析批改结果失败:', err);
        setParsedResult(null);
      }
    } else {
      console.warn('未收到批改结果');
      setError('未找到批改结果，请重新上传作业');
    }
  }, [router.query.result]);

  const handleRetry = () => {
    console.log('用户点击重试按钮');
    router.back();
  };

  const handleDownloadExcel = () => {
    if (!parsedResult) {
      message.error('没有可下载的批改结果');
      return;
    }

    try {
      // 确定数据结构并准备Excel数据
      let fileName = '作业批改结果.xlsx';
      
      // 处理数组类型结果（多个学生）
      if (Array.isArray(parsedResult)) {
        // 为每个学生创建一个工作表
        const workbook = XLSX.utils.book_new();
        
        // 创建总览表
        const overviewData = parsedResult.map((student, index) => {
          return {
            '序号': index + 1,
            '学生姓名': student.name || `学生${index + 1}`,
            '总分': student.overallScore || 'N/A',
            '总体评价': student.feedback || 'N/A'
          };
        });
        
        const overviewSheet = XLSX.utils.json_to_sheet(overviewData);
        XLSX.utils.book_append_sheet(workbook, overviewSheet, '总览');
        
        // 为每个学生创建详细表
        parsedResult.forEach((student, index) => {
          const studentName = student.name || `学生${index + 1}`;
          
          // 准备学生答案数据
          let studentData: Record<string, any>[] = [];
          if (Array.isArray(student.answers)) {
            studentData = student.answers.map((answer: Answer) => {
              const baseData: Record<string, any> = {
                '题号': answer.questionNumber || 'N/A',
                '学生答案': answer.studentAnswer || 'N/A',
                '是否正确': answer.isCorrect ? '是' : '否',
                '解释/评价': answer.explanation || answer.evaluation || 'N/A'
              };
              
              // 如果有标准答案，添加进来
              if (!answer.isCorrect && answer.correctAnswer) {
                baseData['标准答案'] = answer.correctAnswer;
              }
              
              // 如果有建议，添加进来
              if (answer.suggestion) {
                baseData['建议'] = answer.suggestion;
              }
              
              return baseData;
            });
          }
          
          const studentSheet = XLSX.utils.json_to_sheet(studentData);
          XLSX.utils.book_append_sheet(workbook, studentSheet, studentName.substring(0, 31)); // Excel的工作表名称最长为31个字符
        });
        
        // 导出Excel文件
        XLSX.writeFile(workbook, fileName);
        message.success('批改结果已成功导出为Excel文件');
      } 
      // 处理单个学生结果
      else if (parsedResult.name || parsedResult.answers) {
        const workbook = XLSX.utils.book_new();
        
        // 添加学生总览信息
        const overviewData = [{
          '学生姓名': parsedResult.name || '未知学生',
          '总分': parsedResult.overallScore || 'N/A',
          '总体评价': parsedResult.feedback || 'N/A'
        }];
        
        const overviewSheet = XLSX.utils.json_to_sheet(overviewData);
        XLSX.utils.book_append_sheet(workbook, overviewSheet, '总览');
        
        // 准备学生答案数据
        let answerData: Record<string, any>[] = [];
        if (Array.isArray(parsedResult.answers)) {
          answerData = parsedResult.answers.map((answer: Answer) => {
            const baseData: Record<string, any> = {
              '题号': answer.questionNumber || 'N/A',
              '学生答案': answer.studentAnswer || 'N/A',
              '是否正确': answer.isCorrect ? '是' : '否',
              '解释/评价': answer.explanation || answer.evaluation || 'N/A'
            };
            
            // 如果有标准答案，添加进来
            if (!answer.isCorrect && answer.correctAnswer) {
              baseData['标准答案'] = answer.correctAnswer;
            }
            
            // 如果有建议，添加进来
            if (answer.suggestion) {
              baseData['建议'] = answer.suggestion;
            }
            
            return baseData;
          });
        }
        
        const answersSheet = XLSX.utils.json_to_sheet(answerData);
        XLSX.utils.book_append_sheet(workbook, answersSheet, '详细答案');
        
        // 导出Excel文件
        XLSX.writeFile(workbook, fileName);
        message.success('批改结果已成功导出为Excel文件');
      }
      // 其他格式的结果，尝试直接导出
      else {
        // 将JSON对象转换为工作表
        const worksheet = XLSX.utils.json_to_sheet([parsedResult]);
        const workbook = XLSX.utils.book_new();
        XLSX.utils.book_append_sheet(workbook, worksheet, "批改结果");
        
        // 写入Excel文件并下载
        XLSX.writeFile(workbook, fileName);
        message.success('批改结果已成功导出为Excel文件');
      }
    } catch (error) {
      console.error('导出Excel时出错:', error);
      message.error('导出Excel失败，请重试');
    }
  };

  return (
    <div style={{ maxWidth: 800, margin: '0 auto', padding: '24px' }}>
      <Card>
        <Title level={2}>作业批改结果</Title>
        <Paragraph>
          系统已完成作业分析，以下是批改结果：
        </Paragraph>

        {loading && (
          <div style={{ textAlign: 'center', marginTop: 24 }}>
            <Spin size="large" tip="正在批改中..." />
          </div>
        )}

        {error && (
          <div style={{ marginTop: 24, textAlign: 'center' }}>
            <Paragraph type="danger">{error}</Paragraph>
            <Button 
              type="primary" 
              icon={<ReloadOutlined />} 
              onClick={handleRetry}
            >
              重试
            </Button>
          </div>
        )}

        {result && (
          <div style={{ marginTop: 24 }}>
            <Title level={3} style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
              <span>批改结果</span>
              <Button 
                type="primary"
                icon={<DownloadOutlined />}
                onClick={handleDownloadExcel}
                disabled={!parsedResult}
              >
                下载批改结果
              </Button>
            </Title>
            <Card>
              <pre style={{ whiteSpace: 'pre-wrap', wordWrap: 'break-word' }}>
                {result}
              </pre>
            </Card>
          </div>
        )}
      </Card>
    </div>
  );
};

export default MarkHomework; 