import { useState } from 'react'
import { Key, Copy, Check, Plus, KeyRound } from 'lucide-react'
import { useApiKeys, useCreateApiKey, useRevokeApiKey } from '@/hooks/useApiKeys'
import { Table, type Column } from '@/components/ui/Table'
import { Badge } from '@/components/ui/Badge'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
  DialogClose,
} from '@/components/ui/Dialog'
import { cn, formatDate } from '@/lib/utils'
import type { APIKey, CreateAPIKeyResponse } from '@/api/types'

// ---- Create Key Dialog ----

interface CreateKeyDialogProps {
  open: boolean
  onClose: () => void
}

function CreateKeyDialog({ open, onClose }: CreateKeyDialogProps) {
  const { mutate: createKey, isPending } = useCreateApiKey()
  const [name, setName] = useState('')
  const [scope, setScope] = useState<'read' | 'write'>('read')
  const [error, setError] = useState('')
  const [created, setCreated] = useState<CreateAPIKeyResponse | null>(null)
  const [copied, setCopied] = useState(false)

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    if (!name.trim()) {
      setError('Name is required.')
      return
    }
    createKey(
      { name: name.trim(), scopes: [scope] },
      {
        onSuccess: (res) => setCreated(res),
        onError: () => setError('Failed to create API key.'),
      }
    )
  }

  const handleCopy = () => {
    if (!created) return
    navigator.clipboard.writeText(created.key).then(() => {
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    })
  }

  const handleClose = () => {
    setName('')
    setScope('read')
    setError('')
    setCreated(null)
    setCopied(false)
    onClose()
  }

  return (
    <Dialog open={open} onOpenChange={(o) => { if (!o) handleClose() }}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Create API Key</DialogTitle>
        </DialogHeader>

        {created ? (
          <div className="space-y-4 pt-2">
            <div className="rounded-lg border border-amber-200 bg-amber-50 p-3">
              <p className="text-sm font-medium text-amber-800">
                Copy this key now — it won't be shown again.
              </p>
            </div>
            <div className="space-y-1.5">
              <label className="text-sm font-medium text-slate-700">Your API Key</label>
              <div className="flex items-center gap-2">
                <Input
                  readOnly
                  value={created.key}
                  className="font-mono text-xs"
                />
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  onClick={handleCopy}
                  className="shrink-0"
                >
                  {copied ? <Check className="h-4 w-4 text-green-600" /> : <Copy className="h-4 w-4" />}
                </Button>
              </div>
            </div>
            <DialogFooter>
              <Button onClick={handleClose} size="sm">Done</Button>
            </DialogFooter>
          </div>
        ) : (
          <form onSubmit={handleSubmit} className="space-y-4 pt-2">
            <div className="space-y-1.5">
              <label className="text-sm font-medium text-slate-700">Name</label>
              <Input
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="e.g. ConnectWise integration"
                autoFocus
              />
            </div>
            <div className="space-y-1.5">
              <label className="text-sm font-medium text-slate-700">Scope</label>
              <div className="flex gap-3">
                {(['read', 'write'] as const).map((s) => (
                  <label
                    key={s}
                    className={cn(
                      'flex flex-1 cursor-pointer items-center justify-center gap-2 rounded-lg border px-3 py-2 text-sm font-medium transition-colors',
                      scope === s
                        ? 'border-indigo-600 bg-indigo-50 text-indigo-700'
                        : 'border-slate-200 text-slate-600 hover:border-slate-300 hover:bg-slate-50'
                    )}
                  >
                    <input
                      type="radio"
                      name="scope"
                      value={s}
                      checked={scope === s}
                      onChange={() => setScope(s)}
                      className="sr-only"
                    />
                    {s.charAt(0).toUpperCase() + s.slice(1)}
                  </label>
                ))}
              </div>
              <p className="text-xs text-slate-400">
                {scope === 'read' ? 'Read-only access — can list and fetch data.' : 'Full access — can create, update, and delete.'}
              </p>
            </div>
            {error && <p className="text-sm text-red-600">{error}</p>}
            <DialogFooter>
              <DialogClose asChild>
                <Button type="button" variant="outline" size="sm">Cancel</Button>
              </DialogClose>
              <Button type="submit" size="sm" disabled={isPending}>
                {isPending ? 'Creating…' : 'Create Key'}
              </Button>
            </DialogFooter>
          </form>
        )}
      </DialogContent>
    </Dialog>
  )
}

// ---- Revoke Confirm Dialog ----

interface RevokeDialogProps {
  apiKey: APIKey | null
  onClose: () => void
}

function RevokeDialog({ apiKey, onClose }: RevokeDialogProps) {
  const { mutate: revokeKey, isPending } = useRevokeApiKey()
  const [error, setError] = useState('')

  const handleRevoke = () => {
    if (!apiKey) return
    setError('')
    revokeKey(apiKey.id, {
      onSuccess: onClose,
      onError: () => setError('Failed to revoke key.'),
    })
  }

  return (
    <Dialog open={!!apiKey} onOpenChange={(o) => { if (!o) onClose() }}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Revoke API Key</DialogTitle>
        </DialogHeader>
        <div className="space-y-4 pt-2">
          <p className="text-sm text-slate-600">
            Revoke <span className="font-semibold">{apiKey?.name}</span> (
            <span className="font-mono">{apiKey?.key_prefix}…</span>)? Any
            integrations using this key will stop working immediately.
          </p>
          {error && <p className="text-sm text-red-600">{error}</p>}
          <DialogFooter>
            <DialogClose asChild>
              <Button type="button" variant="outline" size="sm">Cancel</Button>
            </DialogClose>
            <Button
              size="sm"
              variant="destructive"
              disabled={isPending}
              onClick={handleRevoke}
            >
              {isPending ? 'Revoking…' : 'Revoke Key'}
            </Button>
          </DialogFooter>
        </div>
      </DialogContent>
    </Dialog>
  )
}

// ---- Main page ----

export function APIKeysPage() {
  const [showCreate, setShowCreate] = useState(false)
  const [revokeKey, setRevokeKey] = useState<APIKey | null>(null)

  const { data: keys = [], isLoading } = useApiKeys()

  const columns: Column<APIKey>[] = [
    {
      key: 'name',
      header: 'Name',
      render: (k) => (
        <div className="flex items-center gap-2">
          <Key className="h-4 w-4 shrink-0 text-slate-400" />
          <span className="font-medium text-slate-900">{k.name}</span>
        </div>
      ),
    },
    {
      key: 'key_prefix',
      header: 'Key Prefix',
      render: (k) => (
        <span className="font-mono text-sm text-slate-600">{k.key_prefix}…</span>
      ),
    },
    {
      key: 'scopes',
      header: 'Scope',
      hideOnMobile: true,
      render: (k) => (
        <div className="flex gap-1">
          {(k.scopes ?? []).map((s) => (
            <Badge key={s} variant={s === 'write' ? 'yellow' : 'blue'}>{s}</Badge>
          ))}
        </div>
      ),
    },
    {
      key: 'revoked_at',
      header: 'Status',
      render: (k) =>
        k.revoked_at ? (
          <Badge variant="red">Revoked</Badge>
        ) : (
          <Badge variant="green">Active</Badge>
        ),
    },
    {
      key: 'last_used_at',
      header: 'Last Used',
      hideOnMobile: true,
      render: (k) => (
        <span className="text-sm text-slate-500">
          {k.last_used_at ? formatDate(k.last_used_at) : 'Never'}
        </span>
      ),
    },
    {
      key: 'created_at',
      header: 'Created',
      hideOnMobile: true,
      render: (k) => (
        <span className="text-sm text-slate-500">{formatDate(k.created_at)}</span>
      ),
    },
    {
      key: 'id',
      header: '',
      render: (k) =>
        !k.revoked_at ? (
          <div className="flex justify-end">
            <Button
              variant="outline"
              size="sm"
              className="text-red-600 hover:bg-red-50 hover:border-red-300"
              onClick={(e) => { e.stopPropagation(); setRevokeKey(k) }}
            >
              Revoke
            </Button>
          </div>
        ) : null,
    },
  ]

  return (
    <div className="space-y-4">
      {/* Header */}
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h1 className="text-2xl font-bold text-slate-900">API Keys</h1>
          <p className="mt-0.5 text-sm text-slate-500">
            {isLoading ? 'Loading…' : `${keys.length} key${keys.length !== 1 ? 's' : ''}`}
          </p>
        </div>
        <Button onClick={() => setShowCreate(true)}>
          <Plus className="h-4 w-4" />
          New Key
        </Button>
      </div>

      {/* Table */}
      <div className="rounded-xl border border-slate-200 bg-white overflow-hidden shadow-sm">
        <Table
          columns={columns}
          data={keys}
          isLoading={isLoading}
          emptyIcon={KeyRound}
          emptyTitle="No API keys"
          emptyDescription="Create a key to allow external integrations to connect."
          keyExtractor={(k) => k.id}
        />
      </div>

      {/* Dialogs */}
      <CreateKeyDialog open={showCreate} onClose={() => setShowCreate(false)} />
      <RevokeDialog apiKey={revokeKey} onClose={() => setRevokeKey(null)} />
    </div>
  )
}
