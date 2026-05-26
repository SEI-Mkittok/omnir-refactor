import { useState } from 'react'
import { Megaphone, Plus, Trash2, Users } from 'lucide-react'
import { Button } from '@/components/ui/Button'
import { Badge } from '@/components/ui/Badge'
import {
  useAddCampaignMembers,
  useCampaignMembers,
  useCampaigns,
  useCreateCampaign,
  useDeleteCampaign,
  useUpdateCampaign,
} from '@/hooks/useCampaigns'
import type { Campaign, CampaignMemberInput, CampaignStatus } from '@/api/types'

function statusBadge(status: CampaignStatus) {
  if (status === 'active') return <Badge variant="green">Active</Badge>
  if (status === 'paused') return <Badge variant="yellow">Paused</Badge>
  if (status === 'completed') return <Badge variant="blue">Completed</Badge>
  if (status === 'archived') return <Badge variant="gray">Archived</Badge>
  return <Badge variant="default">Draft</Badge>
}

export function CampaignsPage() {
  const [draft, setDraft] = useState<Partial<Campaign>>({ name: '', type: 'marketing', status: 'draft', description: '' })
  const [selected, setSelected] = useState('')
  const [memberDraft, setMemberDraft] = useState<CampaignMemberInput>({ member_type: 'contact', member_id: '', source: 'manual' })
  const campaigns = useCampaigns()
  const createCampaign = useCreateCampaign()
  const updateCampaign = useUpdateCampaign()
  const deleteCampaign = useDeleteCampaign()
  const addMembers = useAddCampaignMembers()
  const rows = campaigns.data?.data ?? []
  const selectedCampaign = rows.find((campaign) => campaign.id === selected) ?? rows[0]
  const members = useCampaignMembers(selectedCampaign?.id ?? '')

  async function saveDraft() {
    await createCampaign.mutateAsync(draft)
    setDraft({ name: '', type: 'marketing', status: 'draft', description: '' })
  }

  async function addMember() {
    if (!selectedCampaign || !memberDraft.member_id) return
    await addMembers.mutateAsync({ id: selectedCampaign.id, members: [memberDraft] })
    setMemberDraft({ member_type: 'contact', member_id: '', source: 'manual' })
  }

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
        <div>
          <h1 className="text-2xl font-bold text-[#1A1D23]">Campaigns</h1>
          <p className="mt-1 text-sm text-[#6B7280]">Coordinate marketing lists, webform attribution, sequences, and automation handoff.</p>
        </div>
        <div className="rounded-md border border-slate-200 bg-white px-3 py-2 text-sm">
          <span className="block text-xs text-slate-500">Campaigns</span>
          <span className="font-semibold text-slate-900">{rows.length}</span>
        </div>
      </div>

      <section className="rounded-lg border border-slate-200 bg-white p-5 shadow-sm">
        <div className="mb-4 flex items-center gap-2">
          <Megaphone className="h-4 w-4 text-[#1B3A4B]" />
          <h2 className="text-base font-semibold text-slate-900">Create campaign</h2>
        </div>
        <div className="grid gap-3 md:grid-cols-[1fr_0.7fr_0.7fr_1fr_auto]">
          <input className="rounded-md border border-slate-200 px-3 py-2 text-sm" placeholder="Campaign name" value={draft.name ?? ''} onChange={(e) => setDraft({ ...draft, name: e.target.value })} />
          <input className="rounded-md border border-slate-200 px-3 py-2 text-sm" placeholder="Type" value={draft.type ?? ''} onChange={(e) => setDraft({ ...draft, type: e.target.value })} />
          <select className="rounded-md border border-slate-200 px-3 py-2 text-sm" value={draft.status} onChange={(e) => setDraft({ ...draft, status: e.target.value as CampaignStatus })}>
            <option value="draft">Draft</option>
            <option value="active">Active</option>
            <option value="paused">Paused</option>
            <option value="completed">Completed</option>
          </select>
          <input className="rounded-md border border-slate-200 px-3 py-2 text-sm" placeholder="Description" value={draft.description ?? ''} onChange={(e) => setDraft({ ...draft, description: e.target.value })} />
          <Button onClick={saveDraft} disabled={createCampaign.isPending || !draft.name}>Create</Button>
        </div>
      </section>

      <div className="grid gap-6 xl:grid-cols-[1.3fr_0.9fr]">
        <section className="overflow-hidden rounded-lg border border-slate-200 bg-white shadow-sm">
          <div className="border-b border-slate-100 px-5 py-4">
            <h2 className="text-base font-semibold text-slate-900">Campaign list</h2>
          </div>
          {rows.length === 0 ? (
            <div className="px-5 py-10 text-sm text-slate-500">No campaigns yet.</div>
          ) : (
            rows.map((campaign) => (
              <div
                key={campaign.id}
                onClick={() => setSelected(campaign.id)}
                role="button"
                tabIndex={0}
                className={`flex w-full cursor-pointer flex-col gap-3 border-b border-slate-100 px-5 py-4 text-left md:flex-row md:items-center md:justify-between ${selectedCampaign?.id === campaign.id ? 'bg-[#F4F7FA]' : 'hover:bg-slate-50'}`}
              >
                <div>
                  <div className="flex items-center gap-2">
                    <p className="font-medium text-slate-900">{campaign.name}</p>
                    {statusBadge(campaign.status)}
                  </div>
                  <p className="mt-1 text-xs text-slate-500">{campaign.type} · {campaign.description || 'No description'}</p>
                </div>
                <div className="flex gap-2">
                  <Button variant="outline" size="sm" onClick={(e) => { e.stopPropagation(); updateCampaign.mutate({ id: campaign.id, payload: { status: campaign.status === 'active' ? 'paused' : 'active' } }) }}>
                    {campaign.status === 'active' ? 'Pause' : 'Activate'}
                  </Button>
                  <Button variant="ghost" size="sm" onClick={(e) => { e.stopPropagation(); deleteCampaign.mutate(campaign.id) }}><Trash2 className="h-3.5 w-3.5" /></Button>
                </div>
              </div>
            ))
          )}
        </section>

        <section className="rounded-lg border border-slate-200 bg-white p-5 shadow-sm">
          <div className="mb-4 flex items-center gap-2">
            <Users className="h-4 w-4 text-[#1B3A4B]" />
            <h2 className="text-base font-semibold text-slate-900">Members</h2>
          </div>
          {selectedCampaign ? (
            <>
              <p className="text-sm font-medium text-slate-900">{selectedCampaign.name}</p>
              <div className="mt-4 grid gap-2">
                <select className="rounded-md border border-slate-200 px-3 py-2 text-sm" value={memberDraft.member_type} onChange={(e) => setMemberDraft({ ...memberDraft, member_type: e.target.value as CampaignMemberInput['member_type'] })}>
                  <option value="contact">Contact</option>
                  <option value="lead">Lead</option>
                  <option value="account">Account</option>
                </select>
                <input className="rounded-md border border-slate-200 px-3 py-2 text-sm" placeholder="Record ID" value={memberDraft.member_id} onChange={(e) => setMemberDraft({ ...memberDraft, member_id: e.target.value })} />
                <Button onClick={addMember} disabled={addMembers.isPending || !memberDraft.member_id}><Plus className="h-4 w-4" /> Add member</Button>
              </div>
              <div className="mt-5 divide-y divide-slate-100">
                {(members.data?.data ?? []).length === 0 ? (
                  <p className="py-3 text-sm text-slate-500">No members yet.</p>
                ) : (
                  members.data?.data.map((member) => (
                    <div key={member.id} className="py-3 text-sm">
                      <p className="font-medium text-slate-900">{member.member_type}</p>
                      <p className="text-xs text-slate-500">{member.member_id} · {member.source}</p>
                    </div>
                  ))
                )}
              </div>
            </>
          ) : (
            <p className="text-sm text-slate-500">Select a campaign to manage members.</p>
          )}
        </section>
      </div>
    </div>
  )
}
