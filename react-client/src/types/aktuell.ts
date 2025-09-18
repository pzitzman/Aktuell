// Aktuell Types for React Client

export interface ChangeEvent {
  id: string;
  operationType: 'insert' | 'update' | 'delete' | 'replace' | 'drop' | 'rename';
  database: string;
  collection: string;
  documentKey: Record<string, unknown>;
  fullDocument?: Record<string, unknown>;
  updatedFields?: Record<string, unknown>;
  removedFields?: string[];
  timestamp: string;
  clientTimestamp: string;
}

export interface SnapshotOptions {
  include_snapshot: boolean;
  snapshot_limit?: number;
  batch_size?: number;
  snapshot_filter?: Record<string, unknown>;
  snapshot_sort?: Record<string, unknown>;
}

export interface ClientMessage {
  type: 'subscribe' | 'unsubscribe' | 'ping';
  database?: string;
  collection?: string;
  requestId: string;
  subscriptionId?: string;
  snapshot_options?: SnapshotOptions;
}

export interface ServerMessage {
  type: 'change' | 'error' | 'pong' | 'snapshot' | 'snapshot_start' | 'snapshot_end';
  change?: ChangeEvent;
  error?: string;
  errorCode?: number;
  requestId?: string;
  success?: boolean;
  snapshot_data?: Record<string, unknown>[];
  snapshot_batch?: number;
  snapshot_total?: number;
  snapshot_remaining?: number;
}

export interface Subscription {
  id: string;
  clientId: string;
  database: string;
  collection: string;
  filter?: Record<string, unknown>;
  createdAt: string;
  snapshot_options?: SnapshotOptions;
}

export interface ConnectionStatus {
  connected: boolean;
  connecting: boolean;
  error?: string;
  lastConnected?: Date;
  reconnectAttempts: number;
}

export interface AktuellStreamStats {
  totalChanges: number;
  changesPerSecond: number;
  subscriptions: Subscription[];
  connectionUptime: number;
}