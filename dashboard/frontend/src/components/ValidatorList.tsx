import { useEffect, useState } from "react";
import type { Validator } from "../types";
import { fetchValidators, ApiRequestError } from "../api";

interface ValidatorListProps {
  chainName: string;
}

export default function ValidatorList({ chainName }: ValidatorListProps) {
  const [validators, setValidators] = useState<Validator[] | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    fetchValidators(chainName)
      .then((data) => {
        if (!cancelled) {
          setValidators(data);
          setError(null);
        }
      })
      .catch((err) => {
        if (!cancelled) {
          setError(
            err instanceof ApiRequestError
              ? err.message
              : "Beklenmeyen bir hata oluştu"
          );
        }
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, [chainName]);

  if (loading) {
    return (
      <div className="space-y-2">
        {[1, 2].map((i) => (
          <div key={i} className="h-16 animate-pulse rounded-md bg-bg-surface" />
        ))}
      </div>
    );
  }

  if (error) {
    return <p className="text-xs text-status-offline">{error}</p>;
  }

  if (!validators || validators.length === 0) {
    return (
      <p className="text-xs text-text-tertiary">Henüz validatör bulunamadı</p>
    );
  }

  return (
    <div className="space-y-3">
      {validators.map((v, idx) => {
        const isActive = v.status.toLowerCase().includes("active");
        return (
          <div
            key={v.validationID ?? `${v.nodeID}-${idx}`}
            className="rounded-md border border-border p-3 hover:bg-bg-surfaceHover"
          >
            <div className="flex items-center justify-between">
              <p className="break-all font-mono text-xs text-text-secondary">
                {v.nodeID}
              </p>
              <span
                className={`ml-3 inline-flex flex-shrink-0 items-center gap-1.5 text-xs ${
                  isActive ? "text-status-online" : "text-status-offline"
                }`}
              >
                <span
                  className={`h-1.5 w-1.5 rounded-full ${
                    isActive ? "bg-status-online" : "bg-status-offline"
                  }`}
                />
                {v.status}
              </span>
            </div>

            <div className="mt-2 grid grid-cols-2 gap-3 border-t border-border pt-2">
              <div>
                <p className="text-xs text-text-tertiary">Ağırlık</p>
                <p className="mt-0.5 font-mono text-xs text-text-secondary">
                  {v.weight}
                </p>
              </div>
              {v.uptime && (
                <div>
                  <p className="text-xs text-text-tertiary">Uptime</p>
                  <p className="mt-0.5 font-mono text-xs text-text-secondary">
                    {v.uptime}
                  </p>
                </div>
              )}
            </div>

            {v.validationID && (
              <div className="mt-2 border-t border-border pt-2">
                <p className="text-xs text-text-tertiary">Validation ID</p>
                <p className="mt-0.5 break-all font-mono text-xs text-text-secondary">
                  {v.validationID}
                </p>
              </div>
            )}
          </div>
        );
      })}
    </div>
  );
}