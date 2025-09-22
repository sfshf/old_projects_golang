import { storage } from '../mmkv';

/// system debug log

export interface Log {
  timestamp: number; // ms
  level: string; // info/error
  message: string; // log content
}

export const getLogs = (): Log[] => {
  const all: Log[] = [];
  all.push(...getInfoLogs().reverse());
  all.push(...getErrorLogs().reverse());
  return all.sort((a, b) => {
    if (a.timestamp < b.timestamp) {
      return 1;
    } else if (a.timestamp > b.timestamp) {
      return -1;
    }
    return 0;
  });
};

const infoLogs: Log[] = [];
export const clearInfoLogs = () => {
  infoLogs.splice(0, infoLogs.length);
};
export const insertInfoLog = (log: Log) => {
  if (log.level === 'info') {
    infoLogs.push(log);
    if (infoLogs.length === 101) {
      infoLogs.splice(0, 1);
    }
  }
};
export const getInfoLogs = () => {
  return infoLogs;
};

export interface ErrorLogsCache {
  logs: Log[];
}
export const debugErrorLogsKey = 'cache_debug.errorLogs';
export const currentDebugErrorLogs = (): null | ErrorLogsCache => {
  const curSetting = storage.getString(debugErrorLogsKey);
  if (curSetting) {
    return JSON.parse(curSetting);
  }
  return null;
};
export const updateDebugErrorLogs = (obj: null | ErrorLogsCache) => {
  if (!obj) {
    storage.delete(debugErrorLogsKey);
    return;
  }
  storage.set(debugErrorLogsKey, JSON.stringify(obj));
};
export const clearErrorLogs = () => {
  updateDebugErrorLogs(null);
};
export const insertErrorLog = (log: Log) => {
  if (log.level === 'error') {
    const cache = currentDebugErrorLogs();
    if (cache) {
      cache.logs.push(log);
      if (cache.logs.length === 21) {
        cache.logs.splice(0, 1);
      }
      updateDebugErrorLogs(cache);
    } else {
      updateDebugErrorLogs({ logs: [log] });
    }
  }
};
export const getErrorLogs = () => {
  const cache = currentDebugErrorLogs();
  if (cache) {
    return cache.logs;
  } else {
    return [];
  }
};
