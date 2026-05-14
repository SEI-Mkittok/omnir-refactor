'use strict';

const subscribeHook = async (z, bundle) => {
  const data = {
    url: bundle.targetUrl,
    events: ['contact.created'],
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
    url: `${bundle.authData.baseUrl}/api/v1/contacts`,
    params: { limit: 1 },
  });
  const data = response.data.data || response.data || [];
  return data.slice(0, 1);
};

module.exports = {
  key: 'new_contact',
  noun: 'Contact',

  display: {
    label: 'New Contact',
    description: 'Triggers when a new contact is created in PraestOS.',
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
      id: 'a1b2c3d4-0000-0000-0000-000000000005',
      org_id: 'a1b2c3d4-0000-0000-0000-000000000000',
      first_name: 'Jane',
      last_name: 'Doe',
      email: 'jane.doe@example.com',
      phone: '+1-555-0100',
      stage: 'lead',
      lead_score: 0,
      tags: [],
      email_opt_out: false,
      created_at: '2026-03-01T10:00:00Z',
      updated_at: '2026-03-01T10:00:00Z',
    },

    outputFields: [
      { key: 'id', label: 'Contact ID' },
      { key: 'first_name', label: 'First Name' },
      { key: 'last_name', label: 'Last Name' },
      { key: 'email', label: 'Email' },
      { key: 'phone', label: 'Phone' },
      { key: 'stage', label: 'Stage' },
      { key: 'lead_score', label: 'Lead Score' },
      { key: 'account_id', label: 'Account ID' },
      { key: 'created_at', label: 'Created At' },
    ],
  },
};
