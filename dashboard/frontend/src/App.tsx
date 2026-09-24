import { useEffect, useState } from "react";
import type { Deployment } from "./types";
import { fetchDeployments, ApiRequestError } from "./api";
import DeploymentCard from "./components/DeploymentCard";
import DeploymentDetail from "./components/DeploymentDetail";

function LoadingGrid() {
  return (
    <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
      {[1, 2, 3].map((i) => (
        <div
          key={i}
          className="h-40 animate-pulse rounded-lg border border-border bg-bg-surface"
        />
      ))}
    </div>
  );
}

function ErrorState({ message }: { message: string }) {
  return (
    <div className="rounded-lg border border-border bg-bg-surface p-8 text-center">
      <p className="text-sm font-medium text-status-offline">{message}</p>
      <p className="mt-1 text-xs text-text-tertiary">
        Backend'in çalıştığından emin olun (localhost:8080)
      </p>
    </div>
  );
}

function EmptyState() {
  return (
    <div className="rounded-lg border border-border bg-bg-surface p-8 text-center">
      <p className="text-sm text-text-secondary">Henüz deploy edilmiş bir zincir yok</p>
    </div>
  );
}

export default function App() {
  const [deployments, setDeployments] = useState<Deployment[] | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [selectedChain, setSelectedChain] = useState<string | null>(null);

  useEffect(() => {
    fetchDeployments()
      .then((data) => {
        setDeployments(data);
        setError(null);
      })
      .catch((err) => {
        setError(
          err instanceof ApiRequestError
            ? err.message
            : "Beklenmeyen bir hata oluştu"
        );
      })
      .finally(() => setLoading(false));
  }, []);

  return (
    <div className="min-h-screen bg-bg text-text-primary">
      <nav className="border-b border-border">
        <div className="mx-auto flex max-w-6xl items-center justify-between px-6 py-4">
          <div className="flex items-center gap-2">
            <span className="h-2 w-2 rounded-full bg-accent" />
            <h1 className="text-sm font-semibold tracking-wide">
              Zero to Secure L1
            </h1>
          </div>
          <div className="flex items-center gap-6 text-xs text-text-secondary">
            <span>Fuji Testnet</span>
          </div>
        </div>
      </nav>

      <main className="mx-auto max-w-6xl px-6 py-8">
        <div className="mb-6">
          <h2 className="text-2xl font-semibold text-text-primary">
            Deployments
          </h2>
          <p className="mt-1 text-sm text-text-secondary">
            Avalanche L1 dağıtımları ve canlı durumları
          </p>
        </div>

        {loading && <LoadingGrid />}
        {!loading && error && <ErrorState message={error} />}
        {!loading && !error && deployments && deployments.length === 0 && (
          <EmptyState />
        )}

        {!loading && !error && deployments && deployments.length > 0 && (
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
            {deployments.map((deployment) => {
              const isSelected = selectedChain === deployment.chainName;
              return (
                <div
                  key={deployment.chainName}
                  className="col-span-1 overflow-hidden rounded-lg"
                >
                  <DeploymentCard
                    deployment={deployment}
                    isSelected={isSelected}
                    onToggle={() =>
                      setSelectedChain(isSelected ? null : deployment.chainName)
                    }
                  />
                  {isSelected && <DeploymentDetail deployment={deployment} />}
                </div>
              );
            })}
          </div>
        )}
      </main>
    </div>
  );
}