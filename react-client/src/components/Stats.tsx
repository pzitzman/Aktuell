import type { ChangeEvent, ConnectionStatus } from '../types/aktuell';
import { Activity, Database, Clock, Wifi, WifiOff, AlertCircle } from 'lucide-react';

interface StatsProps {
  changesCount: number;
  subscriptionsCount: number;
  latestChange: ChangeEvent | null;
  connectionStatus: ConnectionStatus;
}

export function Stats({ 
  changesCount, 
  subscriptionsCount, 
  latestChange, 
  connectionStatus 
}: StatsProps) {
  const formatTimestamp = (timestamp: string) => {
    try {
      const date = new Date(timestamp);
      const now = new Date();
      const diffMs = now.getTime() - date.getTime();
      
      if (diffMs < 60000) { // Less than 1 minute
        return `${Math.floor(diffMs / 1000)}s ago`;
      } else if (diffMs < 3600000) { // Less than 1 hour
        return `${Math.floor(diffMs / 60000)}m ago`;
      } else {
        return date.toLocaleTimeString();
      }
    } catch {
      return 'Unknown';
    }
  };

  const getConnectionUptime = () => {
    if (!connectionStatus.connected || !connectionStatus.lastConnected) {
      return 'Not connected';
    }
    
    const uptimeMs = Date.now() - connectionStatus.lastConnected.getTime();
    const uptimeSeconds = Math.floor(uptimeMs / 1000);
    
    if (uptimeSeconds < 60) {
      return `${uptimeSeconds}s`;
    } else if (uptimeSeconds < 3600) {
      return `${Math.floor(uptimeSeconds / 60)}m ${uptimeSeconds % 60}s`;
    } else {
      const hours = Math.floor(uptimeSeconds / 3600);
      const minutes = Math.floor((uptimeSeconds % 3600) / 60);
      return `${hours}h ${minutes}m`;
    }
  };

  const statCards = [
    {
      title: 'Connection',
      value: connectionStatus.connected ? 'Connected' : 'Disconnected',
      subValue: connectionStatus.connected ? getConnectionUptime() : 
                connectionStatus.connecting ? 'Connecting...' :
                connectionStatus.error ? 'Error' : 'Idle',
      icon: connectionStatus.connected ? Wifi : 
            connectionStatus.error ? AlertCircle : WifiOff,
      color: connectionStatus.connected ? 'text-green-400' :
             connectionStatus.error ? 'text-red-400' : 'text-slate-400',
      bgColor: connectionStatus.connected ? 'bg-green-900/20' :
               connectionStatus.error ? 'bg-red-900/20' : 'bg-slate-900/20',
    },
    {
      title: 'Active Subscriptions',
      value: subscriptionsCount.toString(),
      subValue: subscriptionsCount === 1 ? 'subscription' : 'subscriptions',
      icon: Database,
      color: subscriptionsCount > 0 ? 'text-blue-400' : 'text-slate-400',
      bgColor: subscriptionsCount > 0 ? 'bg-blue-900/20' : 'bg-slate-900/20',
    },
    {
      title: 'Total Changes',
      value: changesCount.toString(),
      subValue: changesCount === 1 ? 'event received' : 'events received',
      icon: Activity,
      color: changesCount > 0 ? 'text-purple-400' : 'text-slate-400',
      bgColor: changesCount > 0 ? 'bg-purple-900/20' : 'bg-slate-900/20',
    },
    {
      title: 'Last Change',
      value: latestChange ? formatTimestamp(latestChange.clientTimestamp) : 'None',
      subValue: latestChange ? 
                `${latestChange.operationType} • ${latestChange.database}.${latestChange.collection}` : 
                'No events yet',
      icon: Clock,
      color: latestChange ? 'text-orange-400' : 'text-slate-400',
      bgColor: latestChange ? 'bg-orange-900/20' : 'bg-slate-900/20',
    },
  ];

  return (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
      {statCards.map((stat, index) => {
        const IconComponent = stat.icon;
        return (
          <div
            key={index}
            className={`p-4 bg-slate-800 rounded-lg border border-slate-700 ${stat.bgColor} transition-all duration-200 hover:scale-105`}
          >
            <div className="flex items-center justify-between mb-2">
              <h3 className="text-sm font-medium text-slate-400">{stat.title}</h3>
              <IconComponent className={`w-5 h-5 ${stat.color}`} />
            </div>
            
            <div className={`text-2xl font-bold mb-1 ${stat.color}`}>
              {stat.value}
            </div>
            
            <p className="text-xs text-slate-500 truncate" title={stat.subValue}>
              {stat.subValue}
            </p>

            {/* Connection specific details */}
            {index === 0 && connectionStatus.reconnectAttempts > 0 && (
              <div className="mt-2 text-xs text-yellow-400">
                Reconnect attempts: {connectionStatus.reconnectAttempts}
              </div>
            )}
          </div>
        );
      })}
    </div>
  );
}