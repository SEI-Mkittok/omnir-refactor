'use strict';

// ticket.updated carries status change events. We filter in performList/perform
// to only surface updates where the status field changed.

const subscribeHook = async (z, bundle) => {
  const data = {
    url: bundle.targetUrl,
    events: ['ticket.updated'],
  };
  const response = await z.request({
    url: `${bundle.authData.baseUrl}/api/v1/webhooks`,
    method: 'POST',
    body: data,
  });
  return response.data;
};

const unsubscribeHook = async (z, bundle) => {
  const hookId = bundle.subscribeData.id;
  return z.request({
    url: `${bundle.authData.baseUrl}/api/v1/webhooks/${hookId}`,
    method: 'DELETE',
  });
};

const getFallbackPayload = async (z, bundle) => {
  const response = await z.request({
    url: `${bundle.authData.baseUrl}/api/v1/tickets`,
    params: { limit: 1 },
  });
  const data = response.data.data || response.data || [];
  return data.slice(0, 1);
};

module.exports = {
  key: 'ticket_status_changed',
  noun: 'Ticket',

  display: {
    label: 'Ticket Status Changed',
    description: 'Triggers when the status of an existing ticket changes in PraestOS.',
  },

  operation: {
    type: 'hook',
    performSubscribe: subscribeHook,
    performUnsubscribe: unsubscribeHook,
    perform: (z, bundle) => {
      const payload = bundle.cleanedRequest;
      // Only propagate if the event carries a status change
      if (payload.event === 'ticket.updated' && payload.data && payload.data.status) {
        return [payload.data];
      }
      return [];
    },
    performList: getFallbackPayload,

    sample: {
      id: 'a1b2c3d4-0000-0000-0000-000000000001',
      org_id: 'a1b2c3d4-0000-0000-0000-000000000000',
      subject: 'Cannot log in to my account',
      status: 'resolved',
      priority: 'medium',
      tags: [],
      created_at: '2026-03-01T10:00:00Z',
      updated_at: '2026-03-01T12:00:00Z',
    },

    outputFields: [
      { key: 'id', label: 'Ticket ID' },
      { key: 'subject', label: 'Subject' },
      { key: 'status', label: 'New Status' },
      { key: 'priority', label: 'Priority' },
      { key: 'contact_id', label: 'Contact ID' },
      { key: 'assignee_id', label: 'Assignee ID' },
      { key: 'updated_at', label: 'Updated At' },
    ],
  },
};
