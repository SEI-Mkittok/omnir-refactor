'use strict';

const subscribeHook = async (z, bundle) => {
  const data = {
    url: bundle.targetUrl,
    events: ['ticket.created'],
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

const getTicket = async (z, bundle) => {
  const response = await z.request({
    url: `${bundle.authData.baseUrl}/api/v1/tickets`,
    params: { limit: 1 },
  });
  return response.data.data || response.data || [];
};

const getFallbackPayload = async (z, bundle) => {
  const tickets = await getTicket(z, bundle);
  return tickets.slice(0, 1);
};

module.exports = {
  key: 'new_ticket',
  noun: 'Ticket',

  display: {
    label: 'New Ticket',
    description: 'Triggers when a new support ticket is created in PraestOS.',
  },

  operation: {
    type: 'hook',
    performSubscribe: subscribeHook,
    performUnsubscribe: unsubscribeHook,
    perform: (z, bundle) => {
      const payload = bundle.cleanedRequest;
      return [payload];
    },
    performList: getFallbackPayload,

    sample: {
      id: 'a1b2c3d4-0000-0000-0000-000000000001',
      org_id: 'a1b2c3d4-0000-0000-0000-000000000000',
      subject: 'Cannot log in to my account',
      description: 'I get an error when I try to log in.',
      status: 'open',
      priority: 'medium',
      tags: [],
      created_at: '2026-03-01T10:00:00Z',
      updated_at: '2026-03-01T10:00:00Z',
    },

    outputFields: [
      { key: 'id', label: 'Ticket ID' },
      { key: 'subject', label: 'Subject' },
      { key: 'description', label: 'Description' },
      { key: 'status', label: 'Status' },
      { key: 'priority', label: 'Priority' },
      { key: 'contact_id', label: 'Contact ID' },
      { key: 'assignee_id', label: 'Assignee ID' },
      { key: 'created_at', label: 'Created At' },
    ],
  },
};
