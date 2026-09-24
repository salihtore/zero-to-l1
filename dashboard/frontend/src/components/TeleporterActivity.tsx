import { useEffect, useRef, useState } from "react";
import type { TeleporterInfo } from "../types";
import { fetchTeleporterInfo, ApiRequestError } from "../api";

const POLL_INTERVAL_MS = 5000;

interface TeleporterActivityProps {
  chainName: string;
}

function formatRelativeTime(iso: string): string {
  const diffMs = Date.now() - new Date(iso).getTime();
  const diffMin = Math.round(diffMs / 60000);
  if (diffMin < 1) return "az önce";
  if (diffMin < 60) return `${diffMin} dk önce`;
  const diffHour = Math.round(diffMin / 60);
  return `${diffHour} sa önce`;
}

const STATUS_STYLES: Record<string, string> = {
  Delivered: "text-status-online",
  Verified: "text-status-online",
  "In-Flight": "text-status-pending",
};

export default function TeleporterActivity({
  chainName,
}: TeleporterActivityProps) {
  const [info, setInfo] = useState<TeleporterInfo | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const intervalRef = useRef<ReturnType<typeof setInterval> | null>(null);

  useEffect(() => {
    let cancelled = false;

    const poll = async () => {
      try {
        const result = await fetchTeleporterInfo(chainName);
        if (!cancelled) {
          setInfo(result);
          setError(null);
        }
      } catch (err) {
        if (!cancelled) {
          setError(
            err instanceof ApiRequestError
              ? err.message
              : "Beklenmeyen bir hata oluştu"
          );
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
  }, [chainName]);

  if (loading) {
    return (
      <div className="space-y-2">
        {[1, 2].map((i) => (
          <div key={i} className="h-20 animate-pulse rounded-md bg-bg-surface" />
        ))}
      </div>
    );
  }

  if (error) {
    return <p className="text-xs text-status-offline">{error}</p>;
  }

  if (!info) return null;

  return (
    <div>
      <div className="mb-4 grid grid-cols-1 gap-3">
        <div>
          <p className="text-xs text-text-tertiary">Messenger Adresi</p>
          <p className="mt-0.5 break-all font-mono text-xs text-text-secondary">
            {info.messengerAddress}
          </p>
        </div>
        <div>
          <p className="text-xs text-text-tertiary">Registry Adresi</p>
          <p className="mt-0.5 break-all font-mono text-xs text-text-secondary">
            {info.registryAddress}
          </p>
        </div>
        <div>
          <p className="text-xs text-text-tertiary">Relayer Durumu</p>
          <p className="mt-0.5 inline-flex items-center gap-1.5 text-xs text-status-online">
            <span className="h-1.5 w-1.5 rounded-full bg-status-online" />
            {info.relayerStatus}
          </p>
        </div>
      </div>

      {info.messages.length === 0 ? (
        <p className="text-xs text-text-tertiary">Henüz ICM mesajı yok</p>
      ) : (
        <div className="space-y-2">
          {info.messages.map((msg) => (
            <div
              key={msg.messageID}
              className="rounded-md border border-border p-3 hover:bg-bg-surfaceHover"
            >
              <div className="flex items-center justify-between">
                <span
                  className={`text-xs font-medium ${
                    STATUS_STYLES[msg.status] ?? "text-text-tertiary"
                  }`}
                >
                  {msg.status}
                </span>
                <span className="text-xs text-text-tertiary">
                  {formatRelativeTime(msg.timestamp)}
                </span>
              </div>

              <p className="mt-2 break-all font-mono text-xs text-text-tertiary">
                {msg.messageID}
              </p>

              <p className="mt-2 text-xs text-text-tertiary">
                {msg.sourceChain} → {msg.destChain}
              </p>

              <div className="mt-2 grid grid-cols-1 gap-2 border-t border-border pt-2 sm:grid-cols-2">
                <div>
                  <p className="text-xs text-text-tertiary">Gönderen</p>
                  <p className="mt-0.5 break-all font-mono text-xs text-text-secondary">
                    {msg.sender}
                  </p>
                </div>
                <div>
                  <p className="text-xs text-text-tertiary">Alıcı</p>
                  <p className="mt-0.5 break-all font-mono text-xs text-text-secondary">
                    {msg.receiver}
                  </p>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}