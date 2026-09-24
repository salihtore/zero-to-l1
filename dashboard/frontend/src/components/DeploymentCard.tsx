import type { Deployment } from "../types";

interface DeploymentCardProps {
  deployment: Deployment;
  isSelected: boolean;
  onToggle: () => void;
}

function formatDate(iso: string): string {
  try {
    return new Date(iso).toLocaleString("tr-TR", {
      dateStyle: "medium",
      timeStyle: "short",
    });
  } catch {
    return iso;
  }
}

export default function DeploymentCard({
  deployment,
  isSelected,
  onToggle,
}: DeploymentCardProps) {
  return (
    <button
      onClick={onToggle}
      className={`w-full text-left rounded-lg border bg-bg-surface p-5 transition-colors ${
        isSelected
          ? "border-accent"
          : "border-border hover:border-border-light hover:bg-bg-surfaceHover"
      }`}
    >
      <div className="flex items-start justify-between">
        <div>
          <p className="text-xs uppercase tracking-wide text-text-tertiary">
            {deployment.network}
          </p>
          <h3 className="mt-1 text-lg font-semibold text-text-primary">
            {deployment.chainName}
          </h3>
        </div>
        <span className="rounded-md bg-accent-muted px-2 py-1 text-xs font-medium text-accent">
          {deployment.tokenSymbol}
        </span>
      </div>

      <div className="mt-4 grid grid-cols-1 gap-3">
        <div>
          <p className="text-xs text-text-tertiary">Chain ID</p>
          <p className="mt-0.5 font-mono text-sm text-text-secondary">
            {deployment.chainID}
          </p>
        </div>
        <div>
          <p className="text-xs text-text-tertiary">Blockchain ID</p>
          <p className="mt-0.5 break-all font-mono text-sm text-text-secondary">
            {deployment.blockchainID}
          </p>
        </div>
      </div>

      <div className="mt-3 border-t border-border pt-3">
        <p className="text-xs text-text-tertiary">
          Deploy: {formatDate(deployment.deployedAt)}
        </p>
      </div>
    </button>
  );
}