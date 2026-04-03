// Web Worker Hook - 用于在 React 组件中使用 Web Worker
import { useEffect, useRef, useCallback } from 'react';

interface WorkerOptions {
  onMessage?: (event: MessageEvent) => void;
  onError?: (event: Event | string) => void;
}

function useWorker(workerUrl: string, options: WorkerOptions = {}) {
  const workerRef = useRef<Worker | null>(null);
  const initializedRef = useRef(false);

  useEffect(() => {
    // 只在客户端初始化
    if (typeof window === 'undefined') {
      return;
    }

    // 初始化 Worker
    if (!initializedRef.current && workerUrl) {
      try {
        workerRef.current = new Worker(workerUrl, { type: 'module' });
        initializedRef.current = true;

        // 设置消息处理
        if (options.onMessage) {
          workerRef.current.onmessage = options.onMessage;
        }

        // 设置错误处理
        if (options.onError) {
          workerRef.current.onerror = options.onError;
        }

        // 清理函数
        return () => {
          if (workerRef.current) {
            workerRef.current.terminate();
            initializedRef.current = false;
          }
        };
      } catch (error) {
        console.error('Failed to initialize worker:', error);
      }
    }

    return () => {
      if (workerRef.current) {
        workerRef.current.terminate();
        initializedRef.current = false;
      }
    };
  }, [workerUrl, options.onMessage, options.onError]);

  const postMessage = useCallback((message: any, transfer?: Transferable[]) => {
    if (workerRef.current) {
      workerRef.current.postMessage(message, transfer);
    }
  }, []);

  return { postMessage, worker: workerRef.current, isReady: initializedRef.current } as const;
}

export default useWorker;