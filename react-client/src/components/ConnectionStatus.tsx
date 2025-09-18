import type { ConnectionStatus as ConnectionStatusType } from '../types/aktuell';
import { CheckCircle, AlertCircle, XCircle, Loader2 } from 'lucide-react';

interface ConnectionStatusProps {
  status: ConnectionStatusType;
}

export function ConnectionStatus({ status }: ConnectionStatusProps) {
  const getStatusColor = () => {
    if (status.connected) return 'text-green-400';
    if (status.connecting) return 'text-yellow-400';
    if (status.error) return 'text-red-400';
    return 'text-slate-400';
  };

  const getStatusIcon = () => {
    if (status.connected) return <CheckCircle className="w-5 h-5" />;
    if (status.connecting) return <Loader2 className="w-5 h-5 animate-spin" />;
    if (status.error) return <XCircle className="w-5 h-5" />;
    return <AlertCircle className="w-5 h-5" />;
  };

  const getStatusText = () => {
    if (status.connected) return 'Connected';
    if (status.connecting) return 'Connecting...';
    if (status.error) return `Error: ${status.error}`;
    return 'Disconnected';
  };

  return (
    <div className="flex items-center space-x-3 p-4 bg-slate-800 rounded-lg border border-slate-700">
      <div className={`${getStatusColor()}`}>
        {getStatusIcon()}
      </div>
      <div className="flex-1">
        <div className={`font-medium ${getStatusColor()}`}>
          {getStatusText()}
        </div>
        {status.lastConnected && (
          <div className="text-xs text-slate-500 mt-1">
            Last Connected: {status.lastConnected.toLocaleString()}
          </div>
        )}
        {status.reconnectAttempts > 0 && (
          <div className="text-xs text-slate-500 mt-1">
            Reconnect attempts: {status.reconnectAttempts}
          </div>
        )}
      </div>
    </div>
  );
}