import { useEffect, useState } from 'react'

type Deployment = {
  chainName: string
  blockchainID: string
  rpcURL: string
  chainID: number
  network: string
  deployedAt: string
}

type ChainStatus = {
  chainName: string
  status: 'online' | 'offline'
  blockNumber?: string
  message?: string
  checkedAt: string
}

const apiBaseURL = import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8080'

function formatDate(value: string) {
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? 'Unknown' : date.toLocaleString()
}

function StatusDot({ status }: { status?: ChainStatus['status'] }) {
  const color = status === 'online' ? 'bg-emerald-400 shadow-[0_0_12px_#34d399]' : 'bg-slate-500'
  return <span className={`h-2.5 w-2.5 rounded-full ${color}`} />
}

export function App() {
  const [deployments, setDeployments] = useState<Deployment[]>([])
  const [selected, setSelected] = useState<Deployment | null>(null)
  const [status, setStatus] = useState<ChainStatus | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    fetch(`${apiBaseURL}/api/deployments`)
      .then((response) => response.ok ? response.json() : Promise.reject(new Error('Deployments could not be loaded.')))
      .then((items: Deployment[]) => {
        setDeployments(items)
        setSelected(items[0] ?? null)
      })
      .catch((reason: Error) => setError(reason.message))
      .finally(() => setLoading(false))
  }, [])

  useEffect(() => {
    if (!selected) return
    let active = true
    const refreshStatus = () => {
      fetch(`${apiBaseURL}/api/status/${encodeURIComponent(selected.chainName)}`)
        .then((response) => response.ok ? response.json() : Promise.reject(new Error('Status could not be loaded.')))
        .then((value: ChainStatus) => active && setStatus(value))
        .catch(() => active && setStatus({
          chainName: selected.chainName,
          status: 'offline',
          message: 'Dashboard backend is unavailable.',
          checkedAt: new Date().toISOString(),
        }))
    }
    setStatus(null)
    refreshStatus()
    const interval = window.setInterval(refreshStatus, 5_000)
    return () => { active = false; window.clearInterval(interval) }
  }, [selected])

  return (
    <main className="min-h-screen bg-slate-950 px-5 py-8 text-slate-100 sm:px-8 lg:px-12">
      <div className="mx-auto max-w-6xl">
        <header className="mb-10 flex flex-col gap-3 border-b border-slate-800 pb-7 sm:flex-row sm:items-end sm:justify-between">
          <div>
            <p className="mb-2 text-xs font-semibold uppercase tracking-[0.24em] text-cyan-400">Zero to Secure L1</p>
            <h1 className="text-3xl font-semibold tracking-tight">Network monitor</h1>
            <p className="mt-2 text-sm text-slate-400">Fuji deployment and RPC availability overview.</p>
          </div>
          <span className="rounded-full border border-slate-700 px-3 py-1 text-xs text-slate-400">Refreshes every 5 seconds</span>
        </header>

        {loading && <p className="text-slate-400">Loading deployments…</p>}
        {error && <p className="rounded-lg border border-rose-900 bg-rose-950/40 p-4 text-rose-200">{error}</p>}
        {!loading && !error && deployments.length === 0 && <p className="rounded-xl border border-dashed border-slate-700 p-6 text-slate-400">No deployment records were found.</p>}

        <section className="grid gap-4 md:grid-cols-2 lg:grid-cols-3" aria-label="Deployments">
          {deployments.map((deployment) => (
            <button key={deployment.chainName} type="button" onClick={() => setSelected(deployment)}
              className={`rounded-xl border p-5 text-left transition hover:border-cyan-500/70 hover:bg-slate-800/70 ${selected?.chainName === deployment.chainName ? 'border-cyan-500 bg-slate-800' : 'border-slate-800 bg-slate-900/70'}`}>
              <div className="flex items-center justify-between gap-3">
                <span className="text-lg font-semibold">{deployment.chainName}</span>
                <span className="rounded bg-slate-800 px-2 py-1 text-xs uppercase text-cyan-300">{deployment.network}</span>
              </div>
              <p className="mt-5 text-xs text-slate-500">Deployed</p>
              <p className="mt-1 text-sm text-slate-300">{formatDate(deployment.deployedAt)}</p>
            </button>
          ))}
        </section>

        {selected && <section className="mt-8 rounded-2xl border border-slate-800 bg-slate-900/70 p-6 sm:p-8" aria-live="polite">
          <div className="flex flex-col gap-5 sm:flex-row sm:items-start sm:justify-between">
            <div>
              <p className="text-xs font-semibold uppercase tracking-[0.2em] text-slate-500">Selected chain</p>
              <h2 className="mt-2 text-2xl font-semibold">{selected.chainName}</h2>
              <p className="mt-2 break-all text-sm text-slate-400">Chain ID: {selected.chainID} · {selected.blockchainID || 'Blockchain ID pending'}</p>
            </div>
            <div className="flex items-center gap-3 rounded-xl border border-slate-800 bg-slate-950 px-4 py-3">
              <StatusDot status={status?.status} />
              <div>
                <p className="text-sm font-medium capitalize">{status?.status ?? 'Checking'}</p>
                <p className="text-xs text-slate-500">{status?.checkedAt ? `Checked ${formatDate(status.checkedAt)}` : 'Contacting RPC…'}</p>
              </div>
            </div>
          </div>
          <div className="mt-8 grid gap-4 sm:grid-cols-2">
            <div className="rounded-xl bg-slate-950 p-5">
              <p className="text-xs uppercase tracking-wide text-slate-500">Latest block</p>
              <p className="mt-2 text-3xl font-semibold text-emerald-300">{status?.status === 'online' ? status.blockNumber : '—'}</p>
            </div>
            <div className="rounded-xl bg-slate-950 p-5">
              <p className="text-xs uppercase tracking-wide text-slate-500">RPC status</p>
              <p className="mt-2 text-sm leading-6 text-slate-300">{status?.message ?? 'Checking endpoint…'}</p>
            </div>
          </div>
        </section>}
      </div>
    </main>
  )
}

export default App
