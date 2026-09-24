import type { ChainState } from "../types";

interface StatusDotProps {
  state: ChainState;
  label?: string;
}

interface StateConfig {
  dot: string;
  text: string;
  label: string;
}

type StateConfigMap = {
  online: StateConfig;
  pending: StateConfig;
  offline: StateConfig;
};

const STATE_CONFIG: StateConfigMap = {
  online: {
    dot: "bg-status-online",
    text: "text-status-online",
    label: "Online",
  },
  pending: {
    dot: "bg-status-pending",
    text: "text-status-pending",
    label: "Beklemede",
  },
  offline: {
    dot: "bg-status-offline",
    text: "text-status-offline",
    label: "Offline",
  },
};

export default function StatusDot({ state, label }: StatusDotProps) {
  const config = STATE_CONFIG[state];
  return (
    <div className="flex items-center gap-2">
      <span className={`h-2 w-2 rounded-full ${config.dot}`} />
      <span className={`text-xs font-medium ${config.text}`}>
        {label ?? config.label}
      </span>
    </div>
  );
}