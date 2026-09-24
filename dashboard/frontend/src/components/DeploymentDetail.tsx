import { useEffect, useRef, useState } from "react";
import type { ChainStatus, Deployment } from "../types";
import { fetchChainStatus, ApiRequestError } from "../api";
import StatusDot from "./StatusDot";
import ValidatorList from "./ValidatorList";
import TeleporterActivity from "./TeleporterActivity";

const POLL_INTERVAL_MS = 5000;

type Tab = "status" | "validators" | "teleporter";

const TABS: { id: Tab; label: string }[] = [
  { id: "status", label: "Durum" },
  { id: "validators", label: "Validatörler" },
  { id: "teleporter", label: "Teleporter" },
];

interface DeploymentDetailProps {
  deployment: Deployment;
}

async function copyToClipboard(text: string) {
  try {
    await navigator.clipboard.writeText(text);
  } catch {
    /* sessizce yok say */
  }
}

function CopyableId({ label, value }: { label: string; value: string }) {
  const [copied, setCopied] = useState(false);

  const handleCopy = async () => {
    await copyToClipboard(value);
    setCopied(true);
    setTimeout(() => setCopied(false), 1500);
  };

  return (
    <div>
      <p className="text-xs text-text-tertiary">{label}</p>
      <button
        onClick={handleCopy}
        className="mt-0.5 block break-all text-left font-mono text-sm text-text-secondary hover:text-text-primary"
        title="Kopyalamak için tıkla"
      >
        {copied ? "Kopyalandı" : value}
      </button>
    </div>
  );
}

function StatusTab({ deployment }: { deployment: Deployment }) {
  const [status, setStatus] = useState<ChainStatus | null>(null);
  const [fetchError, setFetchError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const intervalRef = useRef<ReturnType<typeof setInterval> | null>(null);

  useEffect(() => {
    let cancelled = false;

    const poll = async () => {
      try {
        const result = await fetchChainStatus(deployment.chainName);
        if (!cancelled) {
          setStatus(result);
          setFetchError(null);
        }
      } catch (err) {
        if (!cancelled) {
          setFetchError(
            err instanceof ApiRequestError
              ? err.message
              : "Beklenmeyen bir hata oluştu"
          );
          setStatus(null);
        }
      } finally {
        if (!cancelled) setLoading(false);
      }
    };

    poll();
    intervalRef.current = setInterval(poll, POLL_INTERVAL_MS);

    return () => {
      cancelled = true;
      if (intervalRef.current) clearInterval(intervalRef.current);
    };
  }, [deployment.chainName]);

  return (
    <div>
      <div className="flex items-center justify-between">
        <h4 className="text-sm font-semibold text-text-primary">
          Canlı Durum
        </h4>
        {loading ? (
          <span className="text-xs text-text-tertiary">Kontrol ediliyor...</span>
        ) : fetchError ? (
          <StatusDot state="offline" label="Offline" />
        ) : status ? (
          <StatusDot state={status.status} />
        ) : null}
      </div>

      {fetchError && (
        <p className="mt-2 text-xs text-status-offline">{fetchError}</p>
      )}

      {status && status.status === "online" && (
        <p className="mt-2 text-sm text-text-secondary">
          Son blok:{" "}
          <span className="font-mono text-text-primary">
            {status.blockNumber ?? "—"}
          </span>
        </p>
      )}

      {status && status.status !== "online" && !fetchError && (
        <p className="mt-2 text-xs text-text-tertiary">
          {status.message ?? "Validator henüz tam aktif değil"}
        </p>
      )}

      <div className="mt-4 grid grid-cols-1 gap-3 sm:grid-cols-2">
        <CopyableId label="Blockchain ID" value={deployment.blockchainID} />
        <CopyableId label="RPC URL" value={deployment.rpcURL} />
      </div>
    </div>
  );
}

export default function DeploymentDetail({
  deployment,
}: DeploymentDetailProps) {
  const [activeTab, setActiveTab] = useState<Tab>("status");

  return (
    <div className="border-t border-border bg-bg-elevated p-5">
      <div className="mb-4 flex gap-1 border-b border-border">
        {TABS.map((tab) => (
          <button
            key={tab.id}
            onClick={() => setActiveTab(tab.id)}
            className={`px-3 py-2 text-xs font-medium transition-colors ${
              activeTab === tab.id
                ? "border-b-2 border-accent text-text-primary"
                : "text-text-tertiary hover:text-text-secondary"
            }`}
          >
            {tab.label}
          </button>
        ))}
      </div>

      {activeTab === "status" && <StatusTab deployment={deployment} />}
      {activeTab === "validators" && (
        <ValidatorList chainName={deployment.chainName} />
      )}
      {activeTab === "teleporter" && (
        <TeleporterActivity chainName={deployment.chainName} />
      )}
    </div>
  );
}